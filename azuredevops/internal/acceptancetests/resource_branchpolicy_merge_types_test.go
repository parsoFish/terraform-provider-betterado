package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyMergeTypes_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	buildValidationTfNode := "betterado_branch_policy_merge_types.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMergeTypesBasic(projectID, name, true, true, true, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(buildValidationTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_squash", "true"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_rebase_and_fast_forward", "true"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_basic_no_fast_forward", "true"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_rebase_with_merge", "true"),
				),
			}, {
				Config: hclMergeTypesBasic(projectID, name, false, false, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(buildValidationTfNode, "enabled", "false"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_squash", "false"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_rebase_and_fast_forward", "false"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_basic_no_fast_forward", "false"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.allow_rebase_with_merge", "false"),
				),
			}, {
				ResourceName:      buildValidationTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(buildValidationTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclMergeTypesBasic(projectID, name string, enabled, blocking, allowSquash, allowRebase, allowNoFastForward, allowRebaseMerge bool) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "test" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_branch_policy_merge_types" "test" {
  project_id = %[1]q
  enabled    = %[3]t
  blocking   = %[4]t
  settings {
    allow_squash                  = %[5]t
    allow_rebase_and_fast_forward = %[6]t
    allow_basic_no_fast_forward   = %[7]t
    allow_rebase_with_merge       = %[8]t
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}`, projectID, name, enabled, blocking, allowSquash, allowRebase, allowNoFastForward, allowRebaseMerge)
}
