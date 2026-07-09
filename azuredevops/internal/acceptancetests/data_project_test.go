package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccProject_dataSource_withID exercises the framework data.betterado_project data
// source via lookup by project ID. Uses the existing betterado-standing-demo project
// so no new project is created (the ADO org is at the 1000-project cap).
func TestAccProject_dataSource_withID(t *testing.T) {
	tfNode := "data.betterado_project.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		// No CheckDestroy: standing-demo is never deleted.
		Steps: []resource.TestStep{
			{
				Config: hclProjectDataSourceWithIDStandingDemo(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", SharedFixtureProjectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
				),
			},
		},
	})
}

// TestAccProject_dataSource_withName exercises the framework data.betterado_project data
// source via lookup by project name. Uses the existing betterado-standing-demo project
// so no new project is created (the ADO org is at the 1000-project cap).
func TestAccProject_dataSource_withName(t *testing.T) {
	tfNode := "data.betterado_project.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		// No CheckDestroy: standing-demo is never deleted.
		Steps: []resource.TestStep{
			{
				Config: hclProjectDataSourceWithNameStandingDemo(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", SharedFixtureProjectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
				),
			},
		},
	})
}

// hclProjectDataSourceWithIDStandingDemo looks up betterado-standing-demo by project ID
// (resolved via a data source lookup by name).
func hclProjectDataSourceWithIDStandingDemo() string {
	return fmt.Sprintf(`
# Look up the standing-demo project by name so we get its UUID.
data "betterado_project" "by_name" {
  name = %q
}

# Now look it up by the ID we resolved above.
data "betterado_project" "test" {
  project_id = data.betterado_project.by_name.id
}
`, SharedFixtureProjectName)
}

// hclProjectDataSourceWithNameStandingDemo looks up betterado-standing-demo by name.
func hclProjectDataSourceWithNameStandingDemo() string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %q
}
`, SharedFixtureProjectName)
}
