//go:build all || datasource_build_definition_framework

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccBuildDefinition_Framework_DataSource exercises the framework data source
// data.betterado_build_definition via the mux provider:
//
//  1. apply — creates a betterado_build_definition (framework resource) + reads it back
//     via a data source block.
//  2. asserts name and id are set on the data source.
//  3. idempotency — re-plan produces no diff.
//
// Live evidence is captured via CaptureLiveEvidence with label "acceptance-resource"
// and the real REST GET URL for the build definition.
//
// Uses SharedReleaseFixture so no new ADO project is created — the org is at
// the 1000-project cap.
func TestAccBuildDefinition_Framework_DataSource(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_build_definition.fw_ds"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: apply + read-back via data source + capture live evidence.
			{
				Config: hclBuildDefinitionFrameworkDataSource(name, fixture.ProjectID, fixture.RepoID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
					resource.TestCheckResourceAttr(tfDataNode, "name", name),
					captureBuildDefinitionFrameworkDSEvidence(tfDataNode),
				),
			},
			// Step 2: idempotency — no perpetual diff.
			{
				Config:             hclBuildDefinitionFrameworkDataSource(name, fixture.ProjectID, fixture.RepoID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclBuildDefinitionFrameworkDataSource returns Terraform HCL that creates a
// betterado_build_definition resource and reads it back with a data source.
func hclBuildDefinitionFrameworkDataSource(name, projectID, repoID string) string {
	return fmt.Sprintf(`
resource "betterado_build_definition" "fw_ds_res" {
  project_id = %[2]q
  name       = %[1]q
  repository {
    repo_type   = "TfsGit"
    repo_id     = %[3]q
    branch_name = "master"
    yml_path    = "azure-pipelines.yml"
  }
}

data "betterado_build_definition" "fw_ds" {
  project_id = %[2]q
  name       = betterado_build_definition.fw_ds_res.name
  depends_on = [betterado_build_definition.fw_ds_res]
}
`, name, projectID, repoID)
}

// captureBuildDefinitionFrameworkDSEvidence performs a real live API GET of the
// build definition and persists the response as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func captureBuildDefinitionFrameworkDSEvidence(tfDataNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfDataNode]
		if !ok {
			return nil
		}
		idStr := res.Primary.Attributes["id"]
		projectID := res.Primary.Attributes["project_id"]
		if idStr == "" || projectID == "" {
			return nil
		}

		// Parse the definition ID from the state attribute.
		defID := 0
		if _, err := fmt.Sscanf(idStr, "%d", &defID); err != nil || defID == 0 {
			return nil
		}

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort; never fail the test on evidence capture
		}

		// Perform a live GET for the definition.
		defResp, err := clients.BuildClient.GetDefinition(clients.Ctx, build.GetDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
		if err != nil || defResp == nil {
			return nil
		}

		// Build the REST GET URL.
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		for len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		apiURL := fmt.Sprintf(
			"%s/%s/_apis/build/definitions/%d?api-version=7.1",
			orgURL, projectID, defID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, defResp)
		return nil
	}
}
