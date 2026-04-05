package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccVariableGroupDataSource_Basic(t *testing.T) {
	variableGroupName := testutils.GenerateResourceName()
	createAndGetVariableGroupData := fmt.Sprintf("%s\n%s\n%s",
		testutils.HclProjectResource(testutils.GenerateResourceName()),
		testutils.HclVariableGroupResource(variableGroupName, true),
		testutils.HclVariableGroupDataSource())

	tfNode := "data.betterado_variable_group.vg"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: createAndGetVariableGroupData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", variableGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "variable.#"),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "3"),
				),
			},
		},
	})
}

func TestAccVariableGroupDataSource_KeyVault(t *testing.T) {
	t.Skip("Skipping test TestAccVariableGroup_DataSourceKeyVault: azure key vault not provisioned on test infrastructure")
	projectName := testutils.GenerateResourceName()
	variableGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_variable_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: variableGroupKeyVault(projectName, variableGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", variableGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "variable.#"),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "2"),
				),
			},
		},
	})
}

func variableGroupKeyVault(projectName, vgName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_azurerm" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "Sample AzureRM"
  description           = "Managed by Terraform"
  credentials { # TODO
    serviceprincipalid  = "00000000-0000-0000-0000-000000000000"
    serviceprincipalkey = "0000000000000000000000000000000000000"
  } # TODO
  azurerm_spn_tenantid      = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_id   = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_name = "Test Sub Name"
}

resource "betterado_variable_group" "test" {
  project_id   = betterado_project.test.id
  name         = "%s"
  description  = "Test Variable Group Description"
  allow_access = true

  key_vault {
    name                = "MY-KV"
    service_endpoint_id = betterado_serviceendpoint_azurerm.test.id
  }

  variable {
    name = "var01"
  }

  variable {
    name = "var02"
  }
}`, projectName, vgName)
}
