//go:build (all || data_release_folder) && !exclude_data_release_folder

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataReleaseFolder_Basic creates a release folder resource in the
// pre-existing shared project (SharedReleaseFixture), then reads it back via
// the betterado_release_folder data source.
// It verifies that the description attribute matches the resource and that
// a re-plan produces no diff (idempotency).
//
// Uses SharedReleaseFixture so no new ADO project is created — the org is at
// the 1000-project cap, so any project-create attempt would fail immediately.
func TestAccDataReleaseFolder_Basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfResNode := "betterado_release_folder.test"
	tfDataNode := "data.betterado_release_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseFolderBasic(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "description",
						tfResNode, "description"),
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclDataReleaseFolderBasic(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDataReleaseFolderBasic builds a Terraform config that creates a release
// folder resource within the shared project and reads it back via the data source.
// No betterado_project resource is created — the shared project already exists.
func hclDataReleaseFolderBasic(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_folder" "test" {
  project_id  = %[2]q
  path        = "\\DataSourceTest-%[1]s"
  description = "Created by acceptance test"
}

data "betterado_release_folder" "test" {
  project_id = betterado_release_folder.test.project_id
  path       = betterado_release_folder.test.path
}
`, name, fixture.ProjectID)
}
