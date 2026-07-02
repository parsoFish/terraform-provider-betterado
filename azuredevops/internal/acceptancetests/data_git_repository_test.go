package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccGitRepository_DataSource(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "data.betterado_git_repository.repository"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDataRepository(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "disabled"),
				),
			},
		},
	})
}

func TestAccGitRepository_DataSource_notExist(t *testing.T) {
	name := testutils.GenerateResourceName()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      hclDataRepositoryNotExist(name),
				ExpectError: regexp.MustCompile(`Repository with name notExist does not exist`),
			},
		},
	})
}

func hclDataRepository(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%[1]s"
}

data "betterado_git_repository" "repository" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}
`, projectName)
}

func hclDataRepositoryNotExist(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%[1]s"
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "notExist"
}
`, name)
}
