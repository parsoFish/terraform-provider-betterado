package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyWorkItemLinking_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	resourceNode := "betterado_branch_policy_work_item_linking.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclWorkItemLinkingBasic(name, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNode, "enabled", "true"),
				),
			}, {
				Config: hclWorkItemLinkingBasic(name, false, false),
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

func hclWorkItemLinkingBasic(name string, enabled, blocking bool) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name        = "%[1]s"
  description = "description"
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_branch_policy_work_item_linking" "test" {
  project_id = betterado_project.test.id
  enabled    = %[2]t
  blocking   = %[3]t
  settings {
    scope {
      repository_id  = data.betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}`, name, enabled, blocking)
}
