package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccGitRepository_DataSource(t *testing.T) {
	repoName := testutils.GenerateResourceName()
	tfNode := "data.betterado_git_repository.repository"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDataRepository(repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "name", repoName),
					resource.TestCheckResourceAttrSet(tfNode, "disabled"),
				),
			},
		},
	})
}

func TestAccGitRepository_DataSource_notExist(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      hclDataRepositoryNotExist(),
				ExpectError: regexp.MustCompile(`Repository with name notExist does not exist`),
			},
		},
	})
}

// hclDataRepository creates a git repo in the shared project and reads it back
// via the data source. Uses shared fixture project to avoid creating new ADO projects
// (org is at the 1000-project cap).
func hclDataRepository(repoName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = %[2]q
  initialization {
    init_type = "Clean"
  }
}

data "betterado_git_repository" "repository" {
  project_id = data.betterado_project.test.id
  name       = betterado_git_repository.test.name
}
`, SharedFixtureProjectName, repoName)
}

// hclDataRepositoryNotExist queries a data source for a non-existent repo in the
// shared project. No project creation needed.
func hclDataRepositoryNotExist() string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

data "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = "notExist"
}
`, SharedFixtureProjectName)
}
