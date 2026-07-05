package acceptancetests

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccGitRepoFile_basic(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	tfRepoFileNode := "betterado_git_repository_file.test"

	branch := "refs/heads/master"
	file := "foo.txt"
	contentFirst := "bar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryFileBasic(gitRepoName, branch, file, contentFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoFileNode, "file", file),
					resource.TestCheckResourceAttr(tfRepoFileNode, "content", contentFirst),
					resource.TestCheckResourceAttr(tfRepoFileNode, "branch", branch),
					resource.TestCheckResourceAttrSet(tfRepoFileNode, "commit_message"),
					checkGitRepoFileContent(contentFirst),
				),
			},
			{
				ResourceName:            tfRepoFileNode,
				ImportStateIdFunc:       repositoryFileIdFunc(tfRepoFileNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite_on_create"},
			},
		},
	})
}

func TestAccGitRepoFile_complete(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	tfRepoFileNode := "betterado_git_repository_file.test"

	branch := "refs/heads/master"
	file := "foo.txt"
	contentFirst := "bar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryFileComplete(gitRepoName, branch, file, contentFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoFileNode, "file", file),
					resource.TestCheckResourceAttr(tfRepoFileNode, "content", contentFirst),
					resource.TestCheckResourceAttr(tfRepoFileNode, "branch", branch),
					resource.TestCheckResourceAttrSet(tfRepoFileNode, "commit_message"),
					checkGitRepoFileContent(contentFirst),
				),
			},
			{
				ResourceName:            tfRepoFileNode,
				ImportStateIdFunc:       repositoryFileIdFunc(tfRepoFileNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite_on_create"},
			},
		},
	})
}

func TestAccGitRepoFile_authorEmailPolicy(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	tfRepoFileNode := "betterado_git_repository_file.test"

	branch := "refs/heads/master"
	file := "foo.txt"
	contentFirst := "bar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryFileAuthorEmailPolicy(gitRepoName, branch, file, contentFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoFileNode, "file", file),
					resource.TestCheckResourceAttr(tfRepoFileNode, "content", contentFirst),
					resource.TestCheckResourceAttr(tfRepoFileNode, "branch", branch),
					resource.TestCheckResourceAttrSet(tfRepoFileNode, "commit_message"),
					checkGitRepoFileContent(contentFirst),
				),
			},
			{
				ResourceName:            tfRepoFileNode,
				ImportStateIdFunc:       repositoryFileIdFunc(tfRepoFileNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite_on_create"},
			},
		},
	})
}

func TestAccGitRepoFile_update(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	tfRepoFileNode := "betterado_git_repository_file.test"

	branch := "refs/heads/master"
	file := "foo.txt"
	contentFirst := "bar"
	contentSecond := "baz"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoryFileBasic(gitRepoName, branch, file, contentFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoFileNode, "file", file),
					resource.TestCheckResourceAttr(tfRepoFileNode, "content", contentFirst),
					resource.TestCheckResourceAttr(tfRepoFileNode, "branch", branch),
					resource.TestCheckResourceAttrSet(tfRepoFileNode, "commit_message"),
					checkGitRepoFileContent(contentFirst),
				),
			},
			{
				ResourceName:            tfRepoFileNode,
				ImportStateIdFunc:       repositoryFileIdFunc(tfRepoFileNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite_on_create"},
			},
			{
				Config: hclGitRepositoryFileBasic(gitRepoName, branch, file, contentSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfRepoFileNode, "file", file),
					resource.TestCheckResourceAttr(tfRepoFileNode, "content", contentSecond),
					resource.TestCheckResourceAttr(tfRepoFileNode, "branch", branch),
					resource.TestCheckResourceAttrSet(tfRepoFileNode, "commit_message"),
					checkGitRepoFileContent(contentSecond),
				),
			},
			{
				ResourceName:            tfRepoFileNode,
				ImportStateIdFunc:       repositoryFileIdFunc(tfRepoFileNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite_on_create"},
			},
			{
				Config: hclGitRepositoryFileWithoutFile(gitRepoName),
				Check: resource.ComposeTestCheckFunc(
					checkGitRepoFileNotExists(file),
				),
			},
		},
	})
}

func TestAccGitRepoFile_incorrectBranch(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      hclGitRepositoryFileBasic(gitRepoName, "foobar", "foo", "bar"),
				ExpectError: regexp.MustCompile(`Branch not found`),
			},
		},
	})
}

func repositoryFileIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Resource node not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["repository_id"], rs.Primary.Attributes["file"]), nil
	}
}

func checkGitRepoFileNotExists(fileName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkGitRepoFileNotExists: build client: %v", err)
		}

		repo, ok := s.RootModule().Resources["betterado_git_repository.test"]
		if !ok {
			return fmt.Errorf("Did not find a repo definition in the TF state")
		}

		ctx := context.Background()
		_, err = clients.GitReposClient.GetItem(ctx, git.GetItemArgs{
			RepositoryId: &repo.Primary.ID,
			Path:         &fileName,
		})
		if err != nil && !strings.Contains(err.Error(), "could not be found in the repository") {
			return err
		}

		return nil
	}
}

func checkGitRepoFileContent(expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkGitRepoFileContent: build client: %v", err)
		}

		gitFile, ok := s.RootModule().Resources["betterado_git_repository_file.test"]
		if !ok {
			return fmt.Errorf("Did not find a repo definition in the TF state")
		}

		fileID := gitFile.Primary.ID
		comps := strings.Split(fileID, "/")
		repoID := comps[0]
		file := comps[1]

		ctx := context.Background()
		r, err := clients.GitReposClient.GetItemContent(ctx, git.GetItemContentArgs{
			RepositoryId: &repoID,
			Path:         &file,
		})
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(r); err != nil {
			return err
		}

		if buf.String() != expectedContent {
			return fmt.Errorf("Unexpected git file content: %v", buf.String())
		}

		return nil
	}
}

// hclGitRepositoryFileBasic creates a git repository file in the shared fixture project.
func hclGitRepositoryFileBasic(repoName, branch, file, content string) string {
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
  repository_id = betterado_git_repository.test.id
  branch        = %[3]q
  file          = %[4]q
  content       = %[5]q
}
`, SharedFixtureProjectName, repoName, branch, file, content)
}

func hclGitRepositoryFileComplete(repoName, branch, file, content string) string {
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
  repository_id   = betterado_git_repository.test.id
  branch          = %[3]q
  file            = %[4]q
  content         = %[5]q
  author_name     = "author"
  author_email    = "auhtor@test.com"
  committer_name  = "comitter"
  committer_email = "committer@test.com"
}
`, SharedFixtureProjectName, repoName, branch, file, content)
}

func hclGitRepositoryFileAuthorEmailPolicy(repoName, branch, file, content string) string {
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

resource "betterado_repository_policy_author_email_pattern" "test" {
  project_id            = data.betterado_project.test.id
  enabled               = true
  blocking              = true
  author_email_patterns = ["auhtor@test.com"]
  repository_ids        = [betterado_git_repository.test.id]
}

resource "betterado_git_repository_file" "test" {
  repository_id   = betterado_git_repository.test.id
  branch          = %[3]q
  file            = %[4]q
  content         = %[5]q
  author_name     = "author"
  author_email    = "auhtor@test.com"
  committer_name  = "comitter"
  committer_email = "committer@test.com"
  depends_on      = [betterado_repository_policy_author_email_pattern.test]
}
`, SharedFixtureProjectName, repoName, branch, file, content)
}

func hclGitRepositoryFileWithoutFile(repoName string) string {
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
}`, SharedFixtureProjectName, repoName)
}
