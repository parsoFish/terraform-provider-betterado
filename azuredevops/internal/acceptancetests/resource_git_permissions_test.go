package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccGitPermissions_projectGroup(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tfNode := "betterado_git_permissions.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitPermissionsProjectGroup(projectName),
				Check: resource.ComposeTestCheckFunc(
					CheckGitPermissionProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "3"),
				),
			},
		},
	})
}

func TestAccGitPermissions_organizationGroup(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tfNode := "betterado_git_permissions.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitPermissionsOrganizationGroup(projectName),
				Check: resource.ComposeTestCheckFunc(
					CheckGitPermissionProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "3"),
				),
			},
		},
	})
}

func TestAccGitPermissions_user(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	userName := testutils.GenerateResourceName()

	tfNode := "betterado_git_permissions.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitPermissionsCustomUser(projectName, userName),
				Check: resource.ComposeTestCheckFunc(
					CheckGitPermissionProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "3"),
				),
			},
		},
	})
}

func TestAccGitPermissions_builtinUser(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tfNode := "betterado_git_permissions.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGitPermissionsBuiltinUser(projectName),
				Check: resource.ComposeTestCheckFunc(
					CheckGitPermissionProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal"),
					resource.TestCheckResourceAttr(tfNode, "permissions.%", "3"),
				),
			},
		},
	})
}

func CheckGitPermissionProjectExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources["betterado_project.test"]
		if !ok {
			return fmt.Errorf("Did not find a project in the TF state")
		}

		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		id := resource.Primary.ID
		project, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
			ProjectId:           &id,
			IncludeCapabilities: converter.Bool(true),
			IncludeHistory:      converter.Bool(false),
		})
		if err != nil {
			return fmt.Errorf("Project with ID=%s cannot be found!. Error=%v", id, err)
		}

		if *project.Name != expectedName {
			return fmt.Errorf("Project with ID=%s has Name=%s, but expected Name=%s", id, *project.Name, expectedName)
		}

		return nil
	}
}

func hclGitPermissionsProjectGroup(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

resource "betterado_git_permissions" "test" {
  project_id = betterado_project.test.id
  principal  = data.betterado_group.test.id
  permissions = {
    CreateRepository = "Deny"
    DeleteRepository = "Deny"
    RenameRepository = "NotSet"
  }
}`, projectName)
}

func hclGitPermissionsOrganizationGroup(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  name = "Project Collection Build Service Accounts"
}

resource "betterado_git_permissions" "test" {
  project_id = betterado_project.test.id
  principal  = data.betterado_group.test.id
  permissions = {
    CreateRepository = "Deny"
    DeleteRepository = "Deny"
    RenameRepository = "NotSet"
  }
}`, projectName)
}

func hclGitPermissionsCustomUser(projectName, userName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_user_entitlement" "test" {
  principal_name = "%s@adtest.com"
}

resource "betterado_git_permissions" "test" {
  project_id = betterado_project.test.id
  principal  = betterado_user_entitlement.test.descriptor
  permissions = {
    CreateRepository = "Deny"
    DeleteRepository = "Deny"
    RenameRepository = "NotSet"
  }
}`, projectName, userName)
}

func hclGitPermissionsBuiltinUser(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_client_config" "test" {}

data "betterado_identity_user" "test" {
  name = "${betterado_project.test.name} Build Service (${compact(split("/", data.betterado_client_config.test.organization_url))[2]})"
}

resource "betterado_git_permissions" "test" {
  project_id = betterado_project.test.id
  principal  = data.betterado_identity_user.test.subject_descriptor
  permissions = {
    CreateRepository = "Deny"
    DeleteRepository = "Allow"
    RenameRepository = "NotSet"
  }
}`, projectName)
}
