//go:build (all || resource_pipeline) && !exclude_resource_pipeline

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

func TestAccPipeline_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkPipelineDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert live read-back (AC1 + AC2)
			{
				Config: hclPipelineBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					capturePipelineEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff (AC1 ExpectNonEmptyPlan:false)
			{
				Config:             hclPipelineBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclPipelineBasic returns a minimal Terraform config that creates a
// betterado_pipeline in the standing shared fixture project (no new project
// is created — the org sits at the 1000-project cap).
// The pipeline references the default git repo of betterado-standing-demo
// which always contains an azure-pipelines.yml file.
func hclPipelineBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

data "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = %[2]q
}

resource "betterado_pipeline" "test" {
  project_id         = data.betterado_project.test.id
  name               = %[1]q
  folder             = "\\tf-acc"
  configuration_type = "yaml"
  repo_id            = data.betterado_git_repository.test.id
  yaml_path          = "/azure-pipelines.yml"
}
`, name, SharedFixtureProjectName)
}

// checkPipelineDestroyed verifies that all betterado_pipeline resources in the
// state have been deleted from Azure DevOps.
func checkPipelineDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkPipelineDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_pipeline" {
			continue
		}

		pipelineID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("pipeline ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		_, err = clients.PipelinesClient.GetPipeline(clients.Ctx, adoPipelines.GetPipelineArgs{
			Project:    &projectID,
			PipelineId: &pipelineID,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				// resource is gone — expected
				continue
			}
			return fmt.Errorf("error reading pipeline %d after destroy: %v", pipelineID, err)
		}
		// If no error, the pipeline still exists — unexpected.
		return fmt.Errorf("pipeline %d still exists after destroy", pipelineID)
	}

	return nil
}

// capturePipelineEvidence performs a real live API GET of the created pipeline
// and persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test.
//
// Satisfies AC2: CaptureLiveEvidence("acceptance-resource", <GET URL>, <API response>)
func capturePipelineEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		pipelineID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}

		pipeline, err := clients.PipelinesClient.GetPipeline(clients.Ctx, adoPipelines.GetPipelineArgs{
			Project:    &projectID,
			PipelineId: &pipelineID,
		})
		if err != nil || pipeline == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/pipelines/%d?api-version=7.1-preview.1", orgURL, projectID, pipelineID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, pipeline)
		return nil
	}
}
