//go:build (all || data_pipeline) && !exclude_data_pipeline

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
)

func TestAccDataPipeline_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfResNode := "betterado_pipeline.test"
	tfDataNode := "data.betterado_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkPipelineDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create resource + read back via data source; verify matching attrs.
			{
				Config: hclDataPipelineBasic(name),
				Check: resource.ComposeTestCheckFunc(
					// Data source name matches resource name (AC3)
					resource.TestCheckResourceAttrPair(tfDataNode, "name", tfResNode, "name"),
					// Data source folder matches resource folder (AC3)
					resource.TestCheckResourceAttrPair(tfDataNode, "folder", tfResNode, "folder"),
					// Data source revision matches resource revision (AC3)
					resource.TestCheckResourceAttrPair(tfDataNode, "revision", tfResNode, "revision"),
					// Data source ID is set
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
					// Capture live evidence for the data-source GET (AC1)
					captureDataPipelineEvidence(tfDataNode),
				),
			},
			// Step 2: idempotency — no perpetual diff (AC3 ExpectNonEmptyPlan:false)
			{
				Config:             hclDataPipelineBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDataPipelineBasic returns a config that creates a betterado_pipeline resource
// and then reads it back via the betterado_pipeline data source (by project_id + id).
// Uses the standing shared fixture project — no new project is created.
// The pipeline references the default git repo of betterado-standing-demo
// which always contains an azure-pipelines.yml file.
func hclDataPipelineBasic(name string) string {
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

data "betterado_pipeline" "test" {
  project_id  = betterado_pipeline.test.project_id
  pipeline_id = tonumber(betterado_pipeline.test.id)
}
`, name, SharedFixtureProjectName)
}

// captureDataPipelineEvidence performs a real live API GET of the pipeline via
// the data source read-back and persists the response as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
//
// Satisfies AC1: CaptureLiveEvidence("pipeline-datasource-read", <GET URL>, <API response>)
func captureDataPipelineEvidence(tfDataNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfDataNode]
		if !ok {
			return nil
		}
		idStr := res.Primary.Attributes["id"]
		pipelineID, err := strconv.Atoi(idStr)
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
		_ = testutils.CaptureLiveEvidence("pipeline-datasource-read", url, pipeline)
		return nil
	}
}
