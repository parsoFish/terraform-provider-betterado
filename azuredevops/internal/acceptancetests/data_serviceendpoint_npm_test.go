package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointNpm_dataSource(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "data.betterado_serviceendpoint_npm.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclServiceEndpointNpmDataSource(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "service_endpoint_name", name),
				),
			},
		},
	})
}

func hclServiceEndpointNpmDataSource(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_serviceendpoint_npm" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%[1]s"
  access_token          = "redacted"
  url                   = "http://url.com/"
}

data "betterado_serviceendpoint_npm" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = betterado_serviceendpoint_npm.test.service_endpoint_name
}
`, name)
}
