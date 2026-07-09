package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccDescriptorDatasource_user(t *testing.T) {
	name := testutils.GenerateResourceName() + "@contoso.com"
	tfNode := "data.betterado_descriptor.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProviderFactories:         testutils.GetSDKv2ProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDescriptorDataSourceUser(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
		},
	})
}

func TestAccDescriptorDatasource_project(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "data.betterado_descriptor.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProviderFactories:         testutils.GetSDKv2ProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDescriptorDataSourceProject(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
		},
	})
}

func TestAccDescriptorDatasource_group(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	tfNode := "data.betterado_descriptor.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProviderFactories:         testutils.GetSDKv2ProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclDescriptorDataSourceGroup(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
		},
	})
}

func hclDescriptorDataSourceUser(name string) string {
	return fmt.Sprintf(`
resource "betterado_user_entitlement" "test" {
  principal_name       = "%s"
  account_license_type = "express"
}

data "betterado_descriptor" "test" {
  storage_key = betterado_user_entitlement.test.id
}`, name)
}

func hclDescriptorDataSourceProject(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_descriptor" "test" {
  storage_key = betterado_project.test.id
}`, projectName)
}

func hclDescriptorDataSourceGroup(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%s"
}

data "betterado_descriptor" "test" {
  storage_key = betterado_project.test.id
}`, projectName, groupName)
}
