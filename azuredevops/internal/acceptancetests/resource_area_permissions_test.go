package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/datahelper"
)

func TestAccAreaPermissions_SetPermissions(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	config := hclAreaPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Deny",
			"WORK_ITEM_WRITE": "Deny",
		},
	})
	tfNodeRoot := "betterado_area_permissions.root-permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckNoResourceAttr(tfNodeRoot, "path"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.WORK_ITEM_WRITE", "deny"),
				),
			},
		},
	})
}

func TestAccAreaPermissions_UpdatePermissions(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	config1 := hclAreaPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Deny",
			"WORK_ITEM_WRITE": "Deny",
		},
	})
	config2 := hclAreaPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "Allow",
			"DELETE":          "Deny",
			"WORK_ITEM_WRITE": "Deny",
		},
	})
	tfNodeRoot := "betterado_area_permissions.root-permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckNoResourceAttr(tfNodeRoot, "path"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckNoResourceAttr(tfNodeRoot, "path"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.WORK_ITEM_WRITE", "deny"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckNoResourceAttr(tfNodeRoot, "path"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.WORK_ITEM_WRITE", "deny"),
				),
			},
		},
	})
}

func hclAreaPermissions(projectName string, permissions map[string]map[string]string) string {
	rootPermissions := datahelper.JoinMap(permissions["root"], "=", "\n")

	return fmt.Sprintf(`
%s

data "betterado_group" "tf-project-readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_area_permissions" "root-permissions" {
  project_id = betterado_project.project.id
  principal  = data.betterado_group.tf-project-readers.id
  permissions = {
		%s
  }
}


`, testutils.HclProjectResource(projectName), rootPermissions)
}
