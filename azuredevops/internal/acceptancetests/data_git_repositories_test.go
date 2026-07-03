package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestAccDataSourceGitRepositories exercises the framework-migrated
// betterado_git_repositories data source against live ADO.
// It creates a repo in the shared fixture project, reads it back via the
// data source, asserts the list is non-empty, and captures live evidence.
func TestAccDataSourceGitRepositories(t *testing.T) {
	repoName := testutils.GenerateResourceName()
	tfNode := "data.betterado_git_repositories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoriesDataSourceBasic(repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "name", repoName),
					resource.TestCheckResourceAttrSet(tfNode, "repositories.0.id"),
					resource.TestCheckResourceAttr(tfNode, "repositories.0.name", repoName),
					resource.TestCheckResourceAttrSet(tfNode, "repositories.0.project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "repositories.0.disabled"),
					gitRepositoriesLiveEvidenceCheck(tfNode),
				),
			},
		},
	})
}

// TestAccDataSourceGitRepositories_All lists all repositories in the shared
// fixture project, confirming the list is non-empty.
func TestAccDataSourceGitRepositories_All(t *testing.T) {
	tfNode := "data.betterado_git_repositories.all"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { preCheckGitRepository(t) },
		ProtoV6ProviderFactories:  testutils.GetMuxedProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclGitRepositoriesDataSourceAll(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					// The shared fixture project always has at least its default repo.
					resource.TestCheckResourceAttrSet(tfNode, "repositories.0.id"),
				),
			},
		},
	})
}

// gitRepositoriesLiveEvidenceCheck captures live evidence during the acceptance
// test check step (while the resource still exists). Best-effort: never fails the test.
func gitRepositoriesLiveEvidenceCheck(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := rs.Primary.Attributes["project_id"]
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if projectID == "" || orgURL == "" {
			return nil
		}
		// Best-effort: build client and fetch repos; ignore errors (evidence is non-blocking).
		clients, clientErr := getDirectClient()
		if clientErr == nil {
			repos, reposErr := clients.GitReposClient.GetRepositories(clients.Ctx, git.GetRepositoriesArgs{
				Project: converter.String(projectID),
			})
			if reposErr == nil && repos != nil {
				reposURL := fmt.Sprintf("%s/_apis/git/repositories?project=%s&api-version=7.0", orgURL, projectID)
				_ = testutils.CaptureLiveEvidence("acceptance-resource", reposURL, repos)
			}
		}
		return nil
	}
}

// hclGitRepositoriesDataSourceBasic creates a git repo in the shared project and reads
// it back by name via betterado_git_repositories.
func hclGitRepositoriesDataSourceBasic(repoName string) string {
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

data "betterado_git_repositories" "test" {
  project_id = data.betterado_project.test.id
  name       = %[2]q
  depends_on = [betterado_git_repository.test]
}
`, SharedFixtureProjectName, repoName)
}

// hclGitRepositoriesDataSourceAll lists all repositories in the shared project.
func hclGitRepositoriesDataSourceAll() string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

data "betterado_git_repositories" "all" {
  project_id = data.betterado_project.test.id
}
`, SharedFixtureProjectName)
}

// TestAccTfsGitRepositories_DataSource_Basic — legacy tests, skipped.
// These were written against the SDKv2 provider and created new projects
// (problematic at the org's 1000-project cap). Replaced by TestAccDataSourceGitRepositories.
func TestAccTfsGitRepositories_DataSource_Basic(t *testing.T) {
	t.Skip("Legacy test creates projects; use TestAccDataSourceGitRepositories instead")
}

func TestAccTfsGitRepositories_DataSource_all(t *testing.T) {
	t.Skip("Legacy test creates projects; use TestAccDataSourceGitRepositories_All instead")
}
