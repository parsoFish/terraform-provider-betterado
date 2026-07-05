//go:build (all || data_source_security_namespace) && !exclude_data_source_security_namespace

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataSecurityNamespaceFramework exercises the framework implementation
// of betterado_security_namespace via the mux provider path.
func TestAccDataSecurityNamespaceFramework(t *testing.T) {
	tfNodeNamespace := "data.betterado_security_namespace.project"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataSecurityNamespaceFramework("Project"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeNamespace, "id"),
					resource.TestCheckResourceAttr(tfNodeNamespace, "name", "Project"),
					resource.TestCheckResourceAttrSet(tfNodeNamespace, "display_name"),
					resource.TestCheckResourceAttrSet(tfNodeNamespace, "actions.#"),
				),
			},
		},
	})
}

func hclDataSecurityNamespaceFramework(name string) string {
	return fmt.Sprintf(`
data "betterado_security_namespace" "project" {
  name = %[1]q
}
`, name)
}
