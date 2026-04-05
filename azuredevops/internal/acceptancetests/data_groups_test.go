package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// Validates that a configuration containing a project group lookup is able to read the resource correctly.
// Because this is a data source, there are no resources to inspect in AzDO
func TestAccGroupsDataSource_Read_Project(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_groups.groups"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupsDataSourceBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource_Read_NoProject(t *testing.T) {
	tfNode := "data.betterado_groups.groups"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupsDataSourceAllGroups(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource_ProjectID_FiltersOutCollectionGroups(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupsDataProjectScopedConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_group.collection_readers", "descriptor"),
					resource.TestCheckResourceAttrSet("data.betterado_groups.project_groups", "groups.#"),
					testAccCheckCollectionGroupNotInProjectGroups(
						"data.betterado_groups.project_groups",
						"betterado_group.collection_readers",
					),
				),
			},
		},
	})
}

func hclGroupsDataSourceBasic(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "project" {
  name               = "%[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_groups" "groups" {
  project_id = betterado_project.project.id
}
`, projectName)
}

func hclGroupsDataSourceAllGroups() string {
	return `data "betterado_groups" "groups" {}`
}

func hclGroupsDataProjectScopedConfig(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_group" "collection_readers" {
  display_name = "Readers"
}

resource "betterado_project" "test" {
  name       = "%s"
  depends_on = [betterado_group.collection_readers]
}

data "betterado_group" "project_admins" {
  name       = "Project Administrators"
  project_id = betterado_project.test.id
}

resource "betterado_group_membership" "make_collection_visible" {
  mode  = "add"
  group = data.betterado_group.project_admins.descriptor
  members = [
    betterado_group.collection_readers.descriptor
  ]
}

data "betterado_groups" "project_groups" {
  project_id = betterado_project.test.id
  depends_on = [betterado_group_membership.make_collection_visible]
}
`, projectName)
}

func testAccCheckCollectionGroupNotInProjectGroups(projectGroupsNode string, collectionGroupNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pg, ok := s.RootModule().Resources[projectGroupsNode]
		if !ok {
			return fmt.Errorf("not found: %s", projectGroupsNode)
		}

		cg, ok := s.RootModule().Resources[collectionGroupNode]
		if !ok {
			return fmt.Errorf("not found: %s", collectionGroupNode)
		}

		collectionDesc := cg.Primary.Attributes["descriptor"]
		if collectionDesc == "" {
			return fmt.Errorf("collection descriptor missing")
		}

		for _, v := range pg.Primary.Attributes {
			if v == collectionDesc {
				return fmt.Errorf(
					"collection-level group %q present in project-scoped group list",
					collectionDesc,
				)
			}
		}

		return nil
	}
}
