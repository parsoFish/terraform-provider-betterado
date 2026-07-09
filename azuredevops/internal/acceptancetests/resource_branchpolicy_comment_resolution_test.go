package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyCommentResolution_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	resourceNode := "betterado_branch_policy_comment_resolution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclCommentResolutionBasic(projectID, name, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNode, "enabled", "true"),
				),
			}, {
				Config: hclCommentResolutionBasic(projectID, name, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNode, "enabled", "false"),
				),
			}, {
				ResourceName:      resourceNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(resourceNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclCommentResolutionBasic(projectID, name string, enabled, blocking bool) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "test" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_branch_policy_comment_resolution" "test" {
  project_id = %[1]q
  enabled    = %[3]t
  blocking   = %[4]t
  settings {
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}`, projectID, name, enabled, blocking)
}
