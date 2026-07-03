package acceptancetests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccGitRepositoryFile_DataSource(t *testing.T) {
	tfNode := "data.betterado_git_repository_file.test"

	repoName := testutils.GenerateResourceName()
	branch := "refs/heads/master"
	file := "foo.txt"
	content := "bar"
	commitMessage := "first commit"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDataRepositoryFile(repoName, branch, file, content, commitMessage, file),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "content", content),
					resource.TestCheckResourceAttr(tfNode, "last_commit_message", commitMessage),
				),
			},
		},
	})
}

func TestAccGitRepositoryFile_DataSource_notExist(t *testing.T) {
	repoName := testutils.GenerateResourceName()
	branch := "refs/heads/master"
	file := "foo.txt"
	content := "bar"
	commitMessage := "first commit"
	not_exists_file := "not_exists.txt"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      hclDataRepositoryFile(repoName, branch, file, content, commitMessage, not_exists_file),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Error: Item not found, repositoryID: [A-Za-z0-9-]+, branch: %s, file: %s", regexp.QuoteMeta(strings.Split(branch, "/")[2]), regexp.QuoteMeta(not_exists_file))),
			},
		},
	})
}

// hclDataRepositoryFile creates a git repo and file in the shared fixture project.
// Uses SharedFixtureProjectName so no new ADO project is created (org at 1000-project cap).
func hclDataRepositoryFile(repoName, branch, rfile, content, commitMessage, dfile string) string {
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

resource "betterado_git_repository_file" "test" {
  repository_id  = betterado_git_repository.test.id
  branch         = %[3]q
  file           = %[4]q
  content        = %[5]q
  commit_message = %[6]q
}

data "betterado_git_repository_file" "test" {
  repository_id = betterado_git_repository.test.id
  branch        = %[3]q
  file          = %[7]q
  depends_on    = [betterado_git_repository_file.test]
}

`, SharedFixtureProjectName, repoName, branch, rfile, content, commitMessage, dfile)
}
