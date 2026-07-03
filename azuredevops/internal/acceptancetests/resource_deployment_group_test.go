package acceptancetests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestAccDeploymentGroup_basic verifies that a deployment group can be created and imported
// against the standing fixture project (avoids the 1000-project limit).
func TestAccDeploymentGroup_basic(t *testing.T) {
	deploymentGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil); enableClassicPipelinesForFixtureProject(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDeploymentGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclDeploymentGroupBasicFixture(deploymentGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", deploymentGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "pool_id"),
					resource.TestCheckResourceAttr(tfNode, "description", ""),
					captureDeploymentGroupEvidence(tfNode),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDeploymentGroup_update verifies that a deployment group can be updated
// against the standing fixture project.
func TestAccDeploymentGroup_update(t *testing.T) {
	deploymentGroupNameFirst := testutils.GenerateResourceName()
	deploymentGroupNameSecond := testutils.GenerateResourceName()
	tfNode := "betterado_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil); enableClassicPipelinesForFixtureProject(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDeploymentGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclDeploymentGroupBasicFixture(deploymentGroupNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", deploymentGroupNameFirst),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclDeploymentGroupWithDescriptionFixture(deploymentGroupNameSecond, "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", deploymentGroupNameSecond),
					resource.TestCheckResourceAttr(tfNode, "description", "Updated description"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDeploymentGroup_withPoolId verifies that a deployment group can be created
// with an explicit pool_id referencing another deployment group's pool.
func TestAccDeploymentGroup_withPoolId(t *testing.T) {
	deploymentGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil); enableClassicPipelinesForFixtureProject(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDeploymentGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclDeploymentGroupWithPoolIdFixture(deploymentGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", deploymentGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "pool_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// checkDeploymentGroupDestroyedMux verifies that every deployment group referenced in the
// state is gone after destroy. Uses GetDirectClient since we're using ProtoV6ProviderFactories.
func checkDeploymentGroupDestroyedMux(s *terraform.State) error {
	clients, err := testutils.GetDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_deployment_group" {
			continue
		}

		deploymentGroupId, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Deployment group ID=%s cannot be parsed. Error=%v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		if _, err := readDeploymentGroup(clients, projectID, deploymentGroupId); err == nil {
			return fmt.Errorf("Deployment group ID %d should not exist", deploymentGroupId)
		}
	}

	return nil
}

// enableClassicPipelinesForFixtureProject enables classic pipeline creation for the
// fixture project. Deployment groups require classic deployment pipelines to be
// allowed at the project level.
//
// Strategy (applied in order):
//  1. SDK-native PATCH via BuildClient.UpdateBuildGeneralSettings using the combined
//     DisableClassicPipelineCreation flag (project-level, proper SDK routing).
//  2. Raw HTTP PATCH with all known flags to the project-level endpoint (belt-and-
//     suspenders: covers newer API fields the SDK struct does not expose).
//  3. Read-back to verify the setting took effect; if still blocked, log a warning
//     and skip the test (cannot proceed when org-level policy blocks creation).
//
// The org-level `_apis/build/generalsettings` endpoint does not exist (404); the
// org-level "disable classic pipelines" policy is managed through Organisation
// Settings → Pipelines → Settings in the ADO UI and cannot be overridden at the
// project level when set there. If that org-level policy is active the test skips.
func enableClassicPipelinesForFixtureProject(t *testing.T) {
	t.Helper()
	orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if orgURL == "" || pat == "" {
		t.Fatal("enableClassicPipelinesForFixtureProject: AZDO_ORG_SERVICE_URL / AZDO_PERSONAL_ACCESS_TOKEN must be set")
	}

	ctx := context.Background()

	// ── 1. SDK-native PATCH (project level, combined flag) ────────────────────
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	azdoClients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("enableClassicPipelinesForFixtureProject: GetAzdoClient: %v", err)
	}

	falseVal := false
	_, err = azdoClients.BuildClient.UpdateBuildGeneralSettings(ctx, build.UpdateBuildGeneralSettingsArgs{
		Project: converter.String(SharedFixtureProjectName),
		NewSettings: &build.PipelineGeneralSettings{
			DisableClassicPipelineCreation: &falseVal,
		},
	})
	if err != nil {
		t.Logf("enableClassicPipelinesForFixtureProject: SDK UpdateBuildGeneralSettings: %v (continuing with raw PATCH)", err)
	}

	// ── 2. Raw HTTP PATCH with all newer flags (belt-and-suspenders) ──────────
	// The SDK struct only has the combined flag; the newer per-type flags
	// (disableClassicBuildPipelineCreation / disableClassicDeploymentPipelineCreation)
	// must be sent via raw HTTP to cover ADO versions that expose them separately.
	rawBody := `{"disableClassicPipelineCreation":false,"disableClassicBuildPipelineCreation":false,"disableClassicDeploymentPipelineCreation":false}`
	projectURL := orgURL + "/" + SharedFixtureProjectName + "/_apis/build/generalsettings?api-version=7.1-preview.1"

	req, err := http.NewRequest(http.MethodPatch, projectURL, strings.NewReader(rawBody))
	if err != nil {
		t.Fatalf("enableClassicPipelinesForFixtureProject: build request: %v", err)
	}
	req.SetBasicAuth("", pat)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("enableClassicPipelinesForFixtureProject: PATCH %s: %v", projectURL, err)
	}
	respBodyBytes, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		t.Logf("enableClassicPipelinesForFixtureProject: reading PATCH response body: %v", readErr)
	}
	t.Logf("enableClassicPipelinesForFixtureProject: PATCH %s returned %d: %s", projectURL, resp.StatusCode, string(respBodyBytes))

	// Brief pause to allow ADO to propagate the setting change before read-back / canary.
	time.Sleep(2 * time.Second)

	// ── 3. Read-back to verify ────────────────────────────────────────────────
	getURL := projectURL // same URL, GET
	getReq, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		t.Logf("enableClassicPipelinesForFixtureProject: build GET request: %v (skipping read-back)", err)
		return
	}
	getReq.SetBasicAuth("", pat)
	getReq.Header.Set("Content-Type", "application/json")

	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		t.Logf("enableClassicPipelinesForFixtureProject: GET %s: %v (skipping read-back)", getURL, err)
		return
	}
	getBodyBytes, getReadErr := io.ReadAll(getResp.Body)
	_ = getResp.Body.Close()
	if getReadErr != nil {
		t.Logf("enableClassicPipelinesForFixtureProject: reading GET response body: %v", getReadErr)
	}
	t.Logf("enableClassicPipelinesForFixtureProject: GET %s returned %d: %s", getURL, getResp.StatusCode, string(getBodyBytes))

	// Parse the read-back to log the current settings for diagnostics.
	var settings map[string]interface{}
	if jsonErr := json.Unmarshal(getBodyBytes, &settings); jsonErr == nil {
		t.Logf("enableClassicPipelinesForFixtureProject: read-back settings: disableClassicPipelineCreation=%v disableClassicBuildPipelineCreation=%v disableClassicDeploymentPipelineCreation=%v",
			settings["disableClassicPipelineCreation"],
			settings["disableClassicBuildPipelineCreation"],
			settings["disableClassicDeploymentPipelineCreation"],
		)
	}

	// ── 4. Canary check: try creating a minimal deployment group via the ADO SDK.
	// This is the definitive test — the project-level read-back can show "false"
	// even when an org-level policy blocks creation. If the canary fails with
	// "classic pipelines are disabled", the setting is blocked at org level and
	// we skip rather than fail with a cryptic error.
	canaryName := "canary-dg-" + testutils.GenerateResourceName()
	canaryDG, canaryErr := azdoClients.TaskAgentClient.AddDeploymentGroup(ctx, taskagent.AddDeploymentGroupArgs{
		Project: converter.String(SharedFixtureProjectName),
		DeploymentGroup: &taskagent.DeploymentGroupCreateParameter{
			Name: &canaryName,
		},
	})
	if canaryErr != nil {
		errStr := canaryErr.Error()
		t.Logf("enableClassicPipelinesForFixtureProject: canary deployment group creation failed: %v", canaryErr)
		if strings.Contains(strings.ToLower(errStr), "classic pipelines are disabled") {
			t.Skip("enableClassicPipelinesForFixtureProject: org-level policy blocks classic deployment pipeline creation; skipping test")
		}
		// Other errors: fatal so it's visible
		t.Fatalf("enableClassicPipelinesForFixtureProject: unexpected canary failure: %v", canaryErr)
	}
	// Canary succeeded — delete it immediately before the test runs.
	if canaryDG != nil && canaryDG.Id != nil {
		if delErr := azdoClients.TaskAgentClient.DeleteDeploymentGroup(ctx, taskagent.DeleteDeploymentGroupArgs{
			Project:           converter.String(SharedFixtureProjectName),
			DeploymentGroupId: canaryDG.Id,
		}); delErr != nil {
			t.Logf("enableClassicPipelinesForFixtureProject: delete canary deployment group: %v (non-fatal)", delErr)
		}
	}
}

// readDeploymentGroup looks up a DeploymentGroup by project and ID.
func readDeploymentGroup(clients *client.AggregatedClient, projectID string, deploymentGroupId int) (*taskagent.DeploymentGroup, error) {
	return clients.TaskAgentClient.GetDeploymentGroup(clients.Ctx, taskagent.GetDeploymentGroupArgs{
		Project:           &projectID,
		DeploymentGroupId: &deploymentGroupId,
	})
}

// captureDeploymentGroupEvidence performs a live API GET of the deployment group and writes
// forge demo live-evidence to .forge/live-evidence/acceptance-resource-deployment-group.json (AC3).
// Best-effort: a capture failure never fails the test.
func captureDeploymentGroupEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		dgIDStr := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		dgID, parseErr := strconv.Atoi(dgIDStr)
		if parseErr != nil {
			return nil //nolint:nilerr // best-effort: parse failure must not fail the test
		}

		clients, clientErr := testutils.GetDirectClient()
		if clientErr != nil {
			return nil //nolint:nilerr // best-effort: client failure must not fail the test
		}

		dg, readErr := readDeploymentGroup(clients, projectID, dgID)
		if readErr != nil || dg == nil {
			return nil //nolint:nilerr // best-effort: API failure must not fail the test
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		url := fmt.Sprintf(
			"%s/%s/_apis/distributedtask/deploymentgroups/%s?api-version=7.1-preview.1",
			orgURL, projectID, dgIDStr,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-deployment-group", url, dg)
		return nil
	}
}

// hclDeploymentGroupBasicFixture creates a deployment group using the standing fixture project.
func hclDeploymentGroupBasicFixture(deploymentGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_deployment_group" "test" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
}
`, deploymentGroupName, SharedFixtureProjectName)
}

// hclDeploymentGroupWithDescriptionFixture creates a deployment group with a description
// using the standing fixture project.
func hclDeploymentGroupWithDescriptionFixture(deploymentGroupName string, description string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[3]q
}

resource "betterado_deployment_group" "test" {
  project_id  = data.betterado_project.fixture.id
  name        = %[1]q
  description = %[2]q
}
`, deploymentGroupName, description, SharedFixtureProjectName)
}

// hclDeploymentGroupWithPoolIdFixture creates a source deployment group and a second one
// that references the source's pool_id, all within the standing fixture project.
func hclDeploymentGroupWithPoolIdFixture(deploymentGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_deployment_group" "pool_source" {
  project_id = data.betterado_project.fixture.id
  name       = "%[1]s-source"
}

resource "betterado_deployment_group" "test" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
  pool_id    = betterado_deployment_group.pool_source.pool_id
}
`, deploymentGroupName, SharedFixtureProjectName)
}
