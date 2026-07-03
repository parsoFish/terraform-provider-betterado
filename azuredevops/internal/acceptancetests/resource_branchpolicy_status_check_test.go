package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyStatusCheck_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceBasic(projectID, repoName, "update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "update"),
				),
			}, {
				ResourceName:      statusCheckTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(statusCheckTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBranchPolicyStatusCheck_complete(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceComplete(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(statusCheckTfNode, "settings.0.author_id"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "Release"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.invalidate_on_update", "true"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.applicability", "conditional"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.display_name", "PreCheck"),
				),
			}, {
				ResourceName:      statusCheckTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(statusCheckTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBranchPolicyStatusCheckUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceBasic(projectID, repoName, "update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "update"),
				),
			}, {
				Config: hclBranchPolicyStatusCheckResourceUpdate(projectID, repoName, "releaseCheck", true, "conditional", "updateName"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(statusCheckTfNode, "settings.0.author_id"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "releaseCheck"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.invalidate_on_update", "true"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.applicability", "conditional"),
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.display_name", "updateName"),
				),
			}, {
				ResourceName:      statusCheckTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(statusCheckTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclBranchPolicyStatusCheckResourceTemplate(projectID string, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "r" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}
`, projectID, repoName)
}

func hclBranchPolicyStatusCheckResourceBasic(projectID string, repoName string, statusName string) string {
	projectAndRepo := hclBranchPolicyStatusCheckResourceTemplate(projectID, repoName)
	statusCheck := fmt.Sprintf(`
resource "betterado_branch_policy_status_check" "p" {
  project_id = %[2]q

  enabled  = true
  blocking = true

  settings {
    name = "%[1]s"
    scope {
      repository_id  = betterado_git_repository.r.id
      repository_ref = betterado_git_repository.r.default_branch
      match_type     = "Exact"
    }
  }
}
`, statusName, projectID)

	return fmt.Sprintf(`%s %s`, projectAndRepo, statusCheck)
}

func hclBranchPolicyStatusCheckResourceComplete(projectID string, repoName string) string {
	return fmt.Sprintf(
		`%s %s`,
		hclBranchPolicyStatusCheckResourceTemplate(projectID, repoName),
		fmt.Sprintf(`
data "betterado_group" "author" {
  project_id = %[1]q
  name       = "Project Administrators"
}

resource "betterado_branch_policy_status_check" "p" {
  project_id = %[1]q

  enabled  = true
  blocking = true

  settings {
    name                 = "Release"
    author_id            = data.betterado_group.author.origin_id
    invalidate_on_update = true
    applicability        = "conditional"
    display_name         = "PreCheck"
    filename_patterns    = ["*.go", "**.ts"]

    scope {
      repository_id  = betterado_git_repository.r.id
      repository_ref = betterado_git_repository.r.default_branch
      match_type     = "Exact"
    }
  }
}
`, projectID),
	)
}

func hclBranchPolicyStatusCheckResourceUpdate(projectID string, repoName string,
	statusName string, invalid bool, applicability string, displayName string,
) string {
	statusCheck := fmt.Sprintf(`
data "betterado_group" "group" {
  project_id = %[6]q
  name       = "Project Administrators"
}

resource "betterado_branch_policy_status_check" "p" {
  project_id = %[6]q

  enabled  = true
  blocking = true

  settings {
    name                 = "%[1]s"
    author_id            = data.betterado_group.group.origin_id
    invalidate_on_update = %[2]t
    applicability        = "%[3]s"
    display_name         = "%[4]s"
    filename_patterns    = ["*.go", "**.ts"]

    scope {
      repository_id  = betterado_git_repository.r.id
      repository_ref = betterado_git_repository.r.default_branch
      match_type     = "Exact"
    }
  }
}
`, statusName, invalid, applicability, displayName, repoName, projectID)

	return fmt.Sprintf(
		`%s %s`,
		hclBranchPolicyStatusCheckResourceTemplate(projectID, repoName),
		statusCheck,
	)
}
