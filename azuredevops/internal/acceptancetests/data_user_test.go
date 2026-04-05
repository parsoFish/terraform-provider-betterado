package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccUser_dataSource(t *testing.T) {
	userName := "foo@email.com"
	tfNode := "data.betterado_user.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataUserBasic(userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "subject_kind"),
					resource.TestCheckResourceAttrSet(tfNode, "principal_name"),
					resource.TestCheckResourceAttrSet(tfNode, "mail_address"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
					resource.TestCheckResourceAttrSet(tfNode, "display_name"),
					resource.TestCheckResourceAttrSet(tfNode, "domain"),
				),
			},
		},
	})
}

func hclDataUserBasic(uname string) string {
	return fmt.Sprintf(`
resource "betterado_user_entitlement" "test" {
  principal_name       = "%[1]s"
  account_license_type = "basic"
}

data "betterado_user" "test" {
  descriptor = betterado_user_entitlement.test.descriptor
  depends_on = [betterado_user_entitlement.test]
}`, uname)
}
