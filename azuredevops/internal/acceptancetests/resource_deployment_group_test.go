package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccDeploymentGroup_basic verifies that a deployment group can be created and imported
// against the standing fixture project (avoids the 1000-project limit).
func TestAccDeploymentGroup_basic(t *testing.T) {
	deploymentGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
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
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
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
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
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
