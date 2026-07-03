package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/datahelper"
)

// TestAccVariableGroupPermissions_SetPermissions verifies that permissions can be set
// on a variable group. Uses the standing fixture project to avoid the 1000-project limit.
func TestAccVariableGroupPermissions_SetPermissions(t *testing.T) {
	variableGroupName := testutils.GenerateResourceName()
	config := hclVariableGroupPermissions(variableGroupName, map[string]string{
		"View":        "allow",
		"Administer":  "allow",
		"Create":      "allow",
		"ViewSecrets": "notset",
		"Use":         "allow",
		"Owner":       "allow",
	})
	tfNode := "betterado_variable_group_permissions.permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttrSet(tfNode, "variable_group_id"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "6"),
					resource.TestCheckResourceAttr(tfNode, "permissions.View", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Administer", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Create", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.ViewSecrets", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Use", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Owner", "allow"),
				),
			},
		},
	})
}

// TestAccVariableGroupPermissions_UpdatePermissions verifies that permissions can be
// updated on a variable group. Uses the standing fixture project.
func TestAccVariableGroupPermissions_UpdatePermissions(t *testing.T) {
	variableGroupName := testutils.GenerateResourceName()
	config1 := hclVariableGroupPermissions(variableGroupName, map[string]string{
		"View":        "allow",
		"Administer":  "allow",
		"Create":      "allow",
		"ViewSecrets": "notset",
		"Use":         "allow",
		"Owner":       "allow",
	})
	config2 := hclVariableGroupPermissions(variableGroupName, map[string]string{
		"View":        "allow",
		"Administer":  "notset",
		"Create":      "notset",
		"ViewSecrets": "notset",
		"Use":         "notset",
		"Owner":       "notset",
	})
	tfNode := "betterado_variable_group_permissions.permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttrSet(tfNode, "variable_group_id"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "6"),
					resource.TestCheckResourceAttr(tfNode, "permissions.View", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Administer", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Create", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.ViewSecrets", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Use", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Owner", "allow"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttrSet(tfNode, "variable_group_id"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "6"),
					resource.TestCheckResourceAttr(tfNode, "permissions.View", "allow"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Administer", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Create", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.ViewSecrets", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Use", "notset"),
					resource.TestCheckResourceAttr(tfNode, "permissions.Owner", "notset"),
				),
			},
		},
	})
}

// hclVariableGroupPermissions generates HCL using the standing fixture project
// (avoids the 1000-project limit). The variable group is created in the fixture
// project; only the VG and its permissions are created/destroyed per test run.
func hclVariableGroupPermissions(variableGroupName string, permissions map[string]string) string {
	variableGroupPermissions := datahelper.JoinMap(permissions, "=", "\n")

	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[3]q
}

resource "betterado_variable_group" "example" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "Test Description"
  allow_access = true

  variable = [{
    name  = "key1"
    value = "val1"
  }]
}

data "betterado_group" "tf-project-readers" {
  project_id = data.betterado_project.fixture.id
  name       = "Readers"
}

resource "betterado_variable_group_permissions" "permissions" {
  project_id        = data.betterado_project.fixture.id
  variable_group_id = betterado_variable_group.example.id
  principal         = data.betterado_group.tf-project-readers.id
  permissions = {
	%[2]s
  }
}


`, variableGroupName,
		variableGroupPermissions,
		SharedFixtureProjectName,
	)
}
