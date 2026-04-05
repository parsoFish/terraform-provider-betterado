package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointSonarCloud_dataSource(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "data.betterado_serviceendpoint_sonarcloud.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclServiceEndpointSonarCloudDataSource(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "service_endpoint_name", name),
					resource.TestCheckResourceAttrSet(tfNode, "service_endpoint_name"),
				),
			},
		},
	})
}

func hclServiceEndpointSonarCloudDataSource(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_serviceendpoint_sonarcloud" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%[1]s"
  token                 = "0000000000000000000000000000000000000000"
  description           = "Managed by Terraform"
}

data "betterado_serviceendpoint_sonarcloud" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = betterado_serviceendpoint_sonarcloud.test.service_endpoint_name
}
`, name)
}
