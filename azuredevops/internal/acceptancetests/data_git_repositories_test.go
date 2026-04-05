package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTfsGitRepositories_DataSource_Basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	tfNode := "data.betterado_git_repositories.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		Providers:                 testutils.GetProviders(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hckGitRepositoriesDatSourceBasic(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", repoName),
					resource.TestCheckResourceAttr(tfNode, "repositories.0.name", repoName),
					resource.TestCheckResourceAttr(tfNode, "repositories.0.default_branch", "refs/heads/master"),
					resource.TestCheckResourceAttrSet(tfNode, "repositories.0.disabled"),
				),
			},
		},
	})
}

func TestAccTfsGitRepositories_DataSource_all(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	tfNode := "data.betterado_git_repositories.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		Providers:                 testutils.GetProviders(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hckGitRepositoriesDatSourceAll(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "repositories.#", "2"),
				),
			},
		},
	})
}

func hckGitRepositoriesDatSourceBasic(projectName, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

data "betterado_git_repositories" "test" {
  project_id = betterado_project.test.id
  name       = "%[2]s"
  depends_on = [betterado_git_repository.test]
}
`, projectName, repoName)
}

func hckGitRepositoriesDatSourceAll(projectName, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

data "betterado_git_repositories" "test" {
  project_id = betterado_project.test.id
  depends_on = [betterado_git_repository.test]
}
`, projectName, repoName)
}
