//go:build (all || data_release_folder) && !exclude_data_release_folder

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataReleaseFolder_Basic creates a project and a release folder resource,
// then reads it back via the betterado_release_folder data source.
// It verifies that the description attribute matches the resource and that
// a re-plan produces no diff (idempotency).
func TestAccDataReleaseFolder_Basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfResNode := "betterado_release_folder.test"
	tfDataNode := "data.betterado_release_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseFolderBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "description",
						tfResNode, "description"),
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclDataReleaseFolderBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDataReleaseFolderBasic builds a Terraform config that creates a project +
// release folder resource and then reads it back via the data source.
func hclDataReleaseFolderBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "test" {
  project_id  = betterado_project.test.id
  path        = "\\DataSourceTest-%[1]s"
  description = "Created by acceptance test"
}

data "betterado_release_folder" "test" {
  project_id = betterado_release_folder.test.project_id
  path       = betterado_release_folder.test.path
}
`, name)
}
