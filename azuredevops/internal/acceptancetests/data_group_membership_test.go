package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccGroupMembershipData_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	tfNode := "data.betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataMemberShipBasic(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
		},
	})
}

func hclDataMemberShipBasic(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "admin" {
  project_id = betterado_project.test.id
  name       = "Build Administrators"
}

data "betterado_group" "contributors" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%s"

  members = [
    data.betterado_group.admin.descriptor,
    data.betterado_group.contributors.descriptor
  ]
}
data "betterado_group_membership" "test" {
  group_descriptor = betterado_group.test.descriptor
}`, projectName, groupName)
}
