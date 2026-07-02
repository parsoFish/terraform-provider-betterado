package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyStatusCheck_basic(t *testing.T) {
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceBasic(projectName, repoName, "update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "update")),
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
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceComplete(projectName, repoName),
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
	statusCheckTfNode := "betterado_branch_policy_status_check.p"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBranchPolicyStatusCheckResourceBasic(projectName, repoName, "update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusCheckTfNode, "settings.0.name", "update"),
				),
			}, {
				Config: hclBranchPolicyStatusCheckResourceUpdate(projectName, repoName, "releaseCheck", true, "conditional", "updateName"),
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

func hclBranchPolicyStatusCheckResourceTemplate(projectName string, repoName string) string {
	// projectName is no longer used: the shared fixture project is always used instead.
	return fmt.Sprintf(`
data "betterado_project" "p" {
  name = %[2]q
}

resource "betterado_git_repository" "r" {
  project_id = data.betterado_project.p.id
  name       = "%[1]s"
  initialization {
    init_type = "Clean"
  }
}
`, repoName, SharedFixtureProjectName)
}

func hclBranchPolicyStatusCheckResourceBasic(projectName string, repoName string, statusName string) string {
	projectAndRepo := hclBranchPolicyStatusCheckResourceTemplate(projectName, repoName)
	statusCheck := fmt.Sprintf(`
resource "betterado_branch_policy_status_check" "p" {
  project_id = data.betterado_project.p.id

  enabled  = true
  blocking = true

  settings {
    name = "%s"
    scope {
      repository_id  = betterado_git_repository.r.id
      repository_ref = betterado_git_repository.r.default_branch
      match_type     = "Exact"
    }
  }
}
`, statusName)

	return fmt.Sprintf(`%s %s`, projectAndRepo, statusCheck)
}

func hclBranchPolicyStatusCheckResourceComplete(projectName string, repoName string) string {
	return fmt.Sprintf(
		`%s %s`,
		hclBranchPolicyStatusCheckResourceTemplate(projectName, repoName), `


resource "betterado_user_entitlement" "user" {
  principal_name       = "mail@email.com"
  account_license_type = "basic"
}

resource "betterado_branch_policy_status_check" "p" {
  project_id = data.betterado_project.p.id

  enabled  = true
  blocking = true

  settings {
    name                 = "Release"
    author_id            = betterado_user_entitlement.user.id
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
`)
}

func hclBranchPolicyStatusCheckResourceUpdate(projectName string, repoName string,
	statusName string, invalid bool, applicability string, displayName string,
) string {
	statusCheck := fmt.Sprintf(`
data "betterado_group" "group" {
  project_id = data.betterado_project.p.id
  name       = "Project Administrators"
}

resource "betterado_branch_policy_status_check" "p" {
  project_id = data.betterado_project.p.id

  enabled  = true
  blocking = true

  settings {
    name                 = "%s"
    author_id            = data.betterado_group.group.origin_id
    invalidate_on_update = %t
    applicability        = "%s"
    display_name         = "%s"
    filename_patterns    = ["*.go", "**.ts"]

    scope {
      repository_id  = betterado_git_repository.r.id
      repository_ref = betterado_git_repository.r.default_branch
      match_type     = "Exact"
    }
  }
}
`, statusName, invalid, applicability, displayName)

	return fmt.Sprintf(
		`%s %s`,
		hclBranchPolicyStatusCheckResourceTemplate(projectName, repoName),
		statusCheck)
}
