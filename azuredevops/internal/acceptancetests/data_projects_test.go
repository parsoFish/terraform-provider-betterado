package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccProjects_dataSource exercises the framework data.betterado_projects data source
// by listing projects filtered by the betterado-standing-demo name. No new project is
// created — the org is at the 1000-project cap.
func TestAccProjects_dataSource(t *testing.T) {
	tfNode := "data.betterado_projects.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectsDataSourceByName(SharedFixtureProjectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "projects.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "projects.0.name", SharedFixtureProjectName),
				),
			},
		},
	})
}

// TestAccProjects_DataSource_SingleProject verifies that data.betterado_projects returns
// exactly 1 project when filtered by an existing project name. Uses the standing-demo
// project so no new project is created.
func TestAccProjects_DataSource_SingleProject(t *testing.T) {
	tfNode := "data.betterado_projects.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectsDataSourceByName(SharedFixtureProjectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "projects.#", "1"),
				),
			},
		},
	})
}

// TestAccProjects_DataSource_EmptyResult verifies that data.betterado_projects returns
// 0 projects for a name that doesn't exist.
func TestAccProjects_DataSource_EmptyResult(t *testing.T) {
	tfNode := "data.betterado_projects.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceProjectsEmptyResult(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "projects.#", "0"),
				),
			},
		},
	})
}

// hclProjectsDataSourceByName lists projects filtered by name using the framework
// data source.
func hclProjectsDataSourceByName(name string) string {
	return fmt.Sprintf(`
data "betterado_projects" "test" {
  name = %q
}
`, name)
}

func hclDataSourceProjectsEmptyResult() string {
	return `
data "betterado_projects" "test" {
  name  = "invalid_name_that_does_not_exist_xyz"
  state = "wellFormed"
}
`
}
