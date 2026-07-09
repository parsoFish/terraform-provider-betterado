//go:build (all || resource_environment || data_environment) && !(exclude_resource_environment && exclude_data_environment)

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Verifies that the following sequence of events occurs without error:
//
//	(1) TF apply creates environment in the standing fixture project
//	(2) TF state values are set
//	(3) Environment can be queried by ID and has expected name
//	(4) TF apply updates environment with new name
//	(5) Environment can be queried by ID and has expected name (live evidence captured)
//	(6) Idempotency re-plan produces no diff
//	(7) Import round-trip succeeds
//	(8) TF destroy deletes environment
//	(9) Environment can no longer be queried by ID
func TestAccEnvironment_CreateAndUpdate(t *testing.T) {
	environmentNameFirst := testutils.GenerateResourceName()
	environmentNameSecond := testutils.GenerateResourceName()
	tfNode := "betterado_environment.environment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkEnvironmentDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclEnvironmentResource(environmentNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", environmentNameFirst),
					resource.TestCheckResourceAttr(tfNode, "description", ""),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					checkEnvironmentExists(environmentNameFirst),
				),
			},
			{
				Config: hclEnvironmentResource(environmentNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", environmentNameSecond),
					resource.TestCheckResourceAttr(tfNode, "description", ""),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					checkEnvironmentExists(environmentNameSecond),
					captureEnvironmentEvidence(tfNode),
				),
			},
			// Idempotency — no perpetual diff
			{
				Config:             hclEnvironmentResource(environmentNameSecond),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				// Resource Acceptance Testing https://www.terraform.io/docs/extend/resources/import.html#resource-acceptance-testing-implementation
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// hclEnvironmentResource creates an environment in the standing fixture project.
// Uses an existing project instead of creating a new one to avoid the 1000-project org limit.
func hclEnvironmentResource(environmentName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_environment" "environment" {
  project_id  = data.betterado_project.fixture.id
  name        = %[1]q
  description = ""
}
`, environmentName, SharedFixtureProjectName)
}

// checkEnvironmentExists verifies the environment exists in AzDO with the expected name.
func checkEnvironmentExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_environment.environment"]
		if !ok {
			return fmt.Errorf("Did not find an environment in the TF state")
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("building direct client: %v", err)
		}
		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("parse ID error, ID: %v. Error= %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		environment, err := clients.TaskAgentClient.GetEnvironmentById(clients.Ctx, taskagent.GetEnvironmentByIdArgs{
			Project:       converter.String(projectID),
			EnvironmentId: &id,
		})
		if err != nil {
			return fmt.Errorf("Environment with ID=%d cannot be found!. Error=%v", id, err)
		}

		if *environment.Name != expectedName {
			return fmt.Errorf("Environment with ID=%d has Name=%s, but expected Name=%s", id, *environment.Name, expectedName)
		}

		return nil
	}
}

// checkEnvironmentDestroyed verifies that environments referenced in the state are gone.
func checkEnvironmentDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_environment" {
			continue
		}

		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Environment ID=%s cannot be parsed. Error=%v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		// indicates the environment still exists - this should fail the test
		if _, err := clients.TaskAgentClient.GetEnvironmentById(clients.Ctx, taskagent.GetEnvironmentByIdArgs{
			Project:       converter.String(projectID),
			EnvironmentId: &id,
		}); err == nil {
			return fmt.Errorf("Environment ID %d should not exist", id)
		}
	}

	return nil
}

// captureEnvironmentEvidence performs a real live API GET and writes forge demo
// live-evidence to .forge/live-evidence/acceptance-resource-environment.json (AC3).
// Best-effort: never fails the test.
func captureEnvironmentEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		envIDStr := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		envID, err := strconv.Atoi(envIDStr)
		if err != nil {
			return nil
		}
		clients, err := getDirectClient()
		if err != nil {
			return nil
		}
		env, err := clients.TaskAgentClient.GetEnvironmentById(clients.Ctx, taskagent.GetEnvironmentByIdArgs{
			EnvironmentId: &envID,
			Project:       &projectID,
		})
		if err != nil || env == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/distributedtask/environments/%s?api-version=7.1", orgURL, projectID, envIDStr)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-environment", url, env)
		return nil
	}
}
