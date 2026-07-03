package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccVariableGroupDataSource_Basic(t *testing.T) {
	variableGroupName := testutils.GenerateResourceName()

	tfNode := "data.betterado_variable_group.vg"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupDataSourceFixture(variableGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", variableGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "variable.#"),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "3"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVariableGroupDataSource_KeyVault(t *testing.T) {
	t.Skip("Skipping test TestAccVariableGroup_DataSourceKeyVault: azure key vault not provisioned on test infrastructure")
	variableGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_variable_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: variableGroupKeyVaultFixture(variableGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", variableGroupName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "variable.#"),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "2"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclVariableGroupDataSourceFixture creates a VG on the standing fixture project and reads it back via data source.
func hclVariableGroupDataSourceFixture(variableGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_variable_group" "vg" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "A sample variable group."
  allow_access = true
  variable = [
    {
      name         = "key1"
      secret_value = "value1"
      is_secret    = true
    },
    {
      name  = "key2"
      value = "value2"
    },
    {
      name = "key3"
    },
  ]
}

data "betterado_variable_group" "vg" {
  project_id = data.betterado_project.fixture.id
  name       = betterado_variable_group.vg.name
}
`, variableGroupName, SharedFixtureProjectName)
}

func variableGroupKeyVaultFixture(vgName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_serviceendpoint_azurerm" "test" {
  project_id            = data.betterado_project.fixture.id
  service_endpoint_name = "Sample AzureRM"
  description           = "Managed by Terraform"
  credentials {
    serviceprincipalid  = "00000000-0000-0000-0000-000000000000"
    serviceprincipalkey = "0000000000000000000000000000000000000"
  }
  azurerm_spn_tenantid      = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_id   = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_name = "Test Sub Name"
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "Test Variable Group Description"
  allow_access = true

  key_vault = [{
    name                = "MY-KV"
    service_endpoint_id = betterado_serviceendpoint_azurerm.test.id
  }]

  variable = [
    {
      name = "var01"
    },
    {
      name = "var02"
    },
  ]
}`, vgName, SharedFixtureProjectName)
}
