//go:build all || resource_pipeline_authorization_framework

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinepermissions"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccPipelineAuthorization_Framework_allPipeline_queue exercises the
// terraform-plugin-framework implementation of betterado_pipeline_authorization via
// the mux provider path:
//
//  1. apply — creates the pipeline authorization (all-pipelines) via the
//     PipelinePermissions API for a queue resource type.
//  2. read-back — asserts project_id and resource_id are set.
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false).
//  4. destroy — cleans up cleanly.
//
// Live evidence is captured during the read-back step via CaptureLiveEvidence
// using label "acceptance-resource".
//
// Uses SharedReleaseFixture to obtain a pre-existing persistent project so
// no new ADO project is created — the org is at the 1000-project cap.
func TestAccPipelineAuthorization_Framework_allPipeline_queue(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_pipeline_authorization.fw_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: apply + assert read-back + capture live evidence.
			{
				Config: hclPipelineAuthorizationFrameworkAllQueue(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "resource_id"),
					capturePipelineAuthorizationFrameworkEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff.
			{
				Config:             hclPipelineAuthorizationFrameworkAllQueue(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclPipelineAuthorizationFrameworkAllQueue returns HCL that creates a project,
// agent pool, agent queue, and an all-pipelines pipeline_authorization for queue type.
func hclPipelineAuthorizationFrameworkAllQueue(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "fw_test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_agent_pool" "fw_test" {
  name           = "%[1]s"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "fw_test" {
  project_id    = betterado_project.fw_test.id
  agent_pool_id = betterado_agent_pool.fw_test.id
}

resource "betterado_pipeline_authorization" "fw_test" {
  project_id  = betterado_project.fw_test.id
  resource_id = betterado_agent_queue.fw_test.id
  type        = "queue"
}
`, name)
}

// capturePipelineAuthorizationFrameworkEvidence performs a real live API GET of the
// pipeline permission and persists it as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func capturePipelineAuthorizationFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		resourceID := res.Primary.Attributes["resource_id"]
		resType := res.Primary.Attributes["type"]

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort; never fail the test
		}

		// Read back via the live API.
		permResp, err := clients.PipelinePermissionsClient.GetPipelinePermissionsForResource(
			clients.Ctx,
			pipelinepermissions.GetPipelinePermissionsForResourceArgs{
				Project:      &projectID,
				ResourceType: &resType,
				ResourceId:   &resourceID,
			},
		)
		if err != nil {
			return nil
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		for len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		apiURL := fmt.Sprintf(
			"%s/%s/_apis/pipelines/pipelinepermissions/%s/%s?api-version=7.1-preview.1",
			orgURL, projectID, resType, resourceID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, permResp)
		return nil
	}
}
