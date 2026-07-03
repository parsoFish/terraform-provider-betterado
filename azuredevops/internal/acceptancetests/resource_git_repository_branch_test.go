package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/tfhelper"
)

func TestAccGitRepositoryBranch_fromBranch(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	branchName := testutils.GenerateResourceName()
	resNode := "betterado_git_repository_branch.test"

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { preCheckGitRepository(t) },
			ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
			CheckDestroy:             checkGitRepoDestroyed,
			Steps: []resource.TestStep{
				{
					Config: hclGitRepoBranchesFromBranch(gitRepoName, branchName),
					Check: resource.ComposeTestCheckFunc(
						checkRepositoryBranchExist(branchName),
						resource.TestCheckResourceAttr(resNode, "name", branchName),
						resource.TestCheckResourceAttr(resNode, "ref_branch", "master"),
						resource.TestCheckResourceAttrSet(resNode, "last_commit_id"),
					),
				},
				{
					ResourceName:            resNode,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateIdFunc:       hclRepositoryBranchID,
					ImportStateVerifyIgnore: []string{"ref_branch"},
				},
			},
		},
	)
}

func TestAccGitRepositoryBranch_fromCommit(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	branchName := testutils.GenerateResourceName()
	resNode := "betterado_git_repository_branch.test"

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { preCheckGitRepository(t) },
			ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
			CheckDestroy:             checkGitRepoDestroyed,
			Steps: []resource.TestStep{
				{
					Config: hclGitRepoBranchesFromCommit(gitRepoName, branchName),
					Check: resource.ComposeTestCheckFunc(
						checkRepositoryBranchExist(branchName),
						resource.TestCheckResourceAttr(resNode, "name", branchName),
						resource.TestCheckResourceAttrSet(resNode, "ref_commit_id"),
						resource.TestCheckResourceAttrSet(resNode, "last_commit_id"),
					),
				},
				{
					ResourceName:            resNode,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateIdFunc:       hclRepositoryBranchID,
					ImportStateVerifyIgnore: []string{"ref_commit_id"},
				},
			},
		},
	)
}

func TestAccGitRepositoryBranch_invalidRef(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	branchName := testutils.GenerateResourceName()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { preCheckGitRepository(t) },
			ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
			CheckDestroy:             checkGitRepoDestroyed,
			Steps: []resource.TestStep{
				{
					Config:      hclGitRepoBranchInvalidRef(gitRepoName, branchName),
					ExpectError: regexp.MustCompile(`No refs found that match ref "refs/tags/0.0.0"`),
				},
			},
		},
	)
}

func TestAccGitRepositoryBranch_requireImportError(t *testing.T) {
	gitRepoName := testutils.GenerateResourceName()
	branchName := testutils.GenerateResourceName()

	resource.Test(
		t, resource.TestCase{
			PreCheck:                 func() { preCheckGitRepository(t) },
			CheckDestroy:             checkGitRepoDestroyed,
			ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config:      hclGitRepoBranchesImportError(gitRepoName, branchName),
					ExpectError: regexp.MustCompile(`Update refs failed. Update Status: staleOldObjectId`),
				},
			},
		},
	)
}

func hclRepositoryBranchID(state *terraform.State) (string, error) {
	res := state.RootModule().Resources["betterado_git_repository_branch.test"]
	repositoryName := res.Primary.Attributes["repository_id"]
	name := res.Primary.Attributes["name"]
	return fmt.Sprintf("%s:%s", repositoryName, name), nil
}

func checkRepositoryBranchExist(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_git_repository_branch.test"]
		if !ok {
			return fmt.Errorf("Did not find `betterado_git_repository_branch` in the TF state")
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("building direct client: %+v", err)
		}
		repoId, branchName, err := tfhelper.ParseGitRepoBranchID(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parse resource IDs: %w", err)
		}

		branch, err := clients.GitReposClient.GetBranch(clients.Ctx, git.GetBranchArgs{
			RepositoryId: &repoId,
			Name:         &branchName,
		})
		if err != nil {
			return fmt.Errorf("Repositroy: %s, Branch: %s cannot be found. Error=%v", repoId, branchName, err)
		}

		if *branch.Name != expectedName {
			return fmt.Errorf("Branch Name=%s, but expected Name=%s", *branch.Name, expectedName)
		}
		return nil
	}
}

// hclGitRepoBranchesFromBranch returns HCL using the shared fixture project so
// no new project is created (the org is at its 1000-project limit).
func hclGitRepoBranchesFromBranch(gitRepoName, branchName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = "%[1]s"
}

resource "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_git_repository_branch" "test" {
  repository_id = betterado_git_repository.test.id
  name          = "%[3]s"
  ref_branch    = "master"
}`, SharedFixtureProjectName, gitRepoName, branchName)
}

// hclGitRepoBranchesFromCommit returns HCL using the shared fixture project.
func hclGitRepoBranchesFromCommit(gitRepoName, branchName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = "%[1]s"
}

resource "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_git_repository_branch" "from_master" {
  repository_id = betterado_git_repository.test.id
  name          = "testbranch-%[3]s"
  ref_branch    = "master"
}

resource "betterado_git_repository_branch" "test" {
  repository_id = betterado_git_repository.test.id
  name          = "%[3]s"
  ref_commit_id = betterado_git_repository_branch.from_master.last_commit_id
}`, SharedFixtureProjectName, gitRepoName, branchName)
}

// hclGitRepoBranchInvalidRef returns HCL using the shared fixture project.
func hclGitRepoBranchInvalidRef(gitRepoName, branchName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = "%[1]s"
}

resource "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_git_repository_branch" "from_master" {
  repository_id = betterado_git_repository.test.id
  name          = "testbranch-%[3]s"
  ref_branch    = "master"
}

resource "betterado_git_repository_branch" "from_commit_id" {
  repository_id = betterado_git_repository.test.id
  name          = "testbranch2-%[3]s"
  ref_commit_id = betterado_git_repository_branch.from_master.last_commit_id
}

resource "betterado_git_repository_branch" "from_nonexistent_tag" {
  repository_id = betterado_git_repository.test.id
  name          = "testbranch-non-existent-tag"
  ref_tag       = "0.0.0"
}`, SharedFixtureProjectName, gitRepoName, branchName)
}

// hclGitRepoBranchesImportError returns HCL using the shared fixture project.
func hclGitRepoBranchesImportError(gitRepoName, branchName string) string {
	return fmt.Sprintf(`
%s

resource "betterado_git_repository_branch" "import" {
  repository_id = betterado_git_repository_branch.test.repository_id
  name          = betterado_git_repository_branch.test.name
  ref_branch    = betterado_git_repository_branch.test.ref_branch
}`, hclGitRepoBranchesFromBranch(gitRepoName, branchName))
}
