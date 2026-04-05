package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccStorageKeyDatasource(t *testing.T) {
	name := testutils.GenerateResourceName() + "@contoso.com"
	tfNode := "data.betterado_storage_key.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testutils.PreCheck(t, nil) },
		ProviderFactories:         testutils.GetProviderFactories(),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: hclStorageKeyDataSource(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
				),
			},
		},
	})
}

func hclStorageKeyDataSource(name string) string {
	return fmt.Sprintf(`
resource "betterado_user_entitlement" "test" {
  principal_name       = "%s"
  account_license_type = "express"
}

data "betterado_storage_key" "test" {
  descriptor = betterado_user_entitlement.test.descriptor
}`, name)
}
