package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/datahelper"
)

func TestAccIterationPermissions_SetPermissions(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	config := hclIterationPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Deny",
		},
		"iteration": {
			"CREATE_CHILDREN": "Allow",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Allow",
		},
	})
	tfNodeRoot := "betterado_iteration_permissions.root-permissions"
	tfNodeIteration := "betterado_iteration_permissions.iteration-permissions"

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
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "principal"),
					resource.TestCheckResourceAttr(tfNodeIteration, "path", "Iteration 1"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.CREATE_CHILDREN", "allow"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.DELETE", "allow"),
				),
			},
		},
	})
}

func TestAccIterationPermissions_UpdatePermissions(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	config1 := hclIterationPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Deny",
		},
		"iteration": {
			"CREATE_CHILDREN": "Allow",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Allow",
		},
	})
	config2 := hclIterationPermissions(projectName, map[string]map[string]string{
		"root": {
			"CREATE_CHILDREN": "Allow",
			"GENERIC_READ":    "NotSet",
			"DELETE":          "Deny",
		},
		"iteration": {
			"CREATE_CHILDREN": "Deny",
			"GENERIC_READ":    "Allow",
			"DELETE":          "NotSet",
		},
	})
	tfNodeRoot := "betterado_iteration_permissions.root-permissions"
	tfNodeIteration := "betterado_iteration_permissions.iteration-permissions"

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
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "principal"),
					resource.TestCheckResourceAttr(tfNodeIteration, "path", "Iteration 1"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.CREATE_CHILDREN", "allow"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.DELETE", "allow"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckNoResourceAttr(tfNodeRoot, "path"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CREATE_CHILDREN", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.GENERIC_READ", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeIteration, "principal"),
					resource.TestCheckResourceAttr(tfNodeIteration, "path", "Iteration 1"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.CREATE_CHILDREN", "deny"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.GENERIC_READ", "allow"),
					resource.TestCheckResourceAttr(tfNodeIteration, "permissions.DELETE", "notset"),
				),
			},
		},
	})
}

func hclIterationPermissions(projectName string, permissions map[string]map[string]string) string {
	rootPermissions := datahelper.JoinMap(permissions["root"], "=", "\n")
	iterationPermissions := datahelper.JoinMap(permissions["iteration"], "=", "\n")

	return fmt.Sprintf(`
%s

data "betterado_group" "tf-project-readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_iteration_permissions" "root-permissions" {
  project_id = betterado_project.project.id
  principal  = data.betterado_group.tf-project-readers.id
  permissions = {
		%s
  }
}

resource "betterado_iteration_permissions" "iteration-permissions" {
  project_id = betterado_project.project.id
  principal  = data.betterado_group.tf-project-readers.id
  path       = "Iteration 1"
  permissions = {
		%s
  }
}


`, testutils.HclProjectResource(projectName), rootPermissions, iterationPermissions)
}
