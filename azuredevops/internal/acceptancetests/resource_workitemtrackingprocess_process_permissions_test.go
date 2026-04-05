package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessProcessPermissions_SetPermissions_InheritedProcess(t *testing.T) {
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_process_permissions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclInheritedProcessPermissions(processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "3"),
				),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessProcessPermissions_SetPermissions_SystemProcess(t *testing.T) {
	tfNode := "betterado_workitemtrackingprocess_process_permissions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclSystemProcessPermissions(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "process_id", agileSystemProcessTypeId),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "1"),
				),
			},
		},
	})
}

func hclInheritedProcessPermissions(processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

data "betterado_group" "project-collection-administrators" {
  name = "Project Collection Administrators"
}

resource "betterado_workitemtrackingprocess_process_permissions" "test" {
  process_id = betterado_workitemtrackingprocess_process.test.id
  principal  = data.betterado_group.project-collection-administrators.id
  permissions = {
    "Edit"                         = "Allow"
    "Delete"                       = "Deny"
    "AdministerProcessPermissions" = "Allow"
  }
}
`, processName, agileSystemProcessTypeId)
}

func hclSystemProcessPermissions() string {
	return fmt.Sprintf(`
data "betterado_group" "project-collection-administrators" {
  name = "Project Collection Administrators"
}

resource "betterado_workitemtrackingprocess_process_permissions" "test" {
  process_id = "%s"
  principal  = data.betterado_group.project-collection-administrators.id
  permissions = {
    "Create" = "Allow"
  }
}
`, agileSystemProcessTypeId)
}
