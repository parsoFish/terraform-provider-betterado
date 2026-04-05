package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointDockerRegistry_data_withName(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "data.betterado_serviceendpoint_dockerregistry"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclDataServiceConnectionDockerRegistryWithName(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
		},
	})
}

func TestAccServiceEndpointDockerRegistry_data_withID(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "data.betterado_serviceendpoint_dockerregistry"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclDataServiceConnectionDockerRegistryWithID(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "service_endpoint_id"),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
		},
	})
}

func hclDataServiceConnectionDockerRegistryWithName(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_dockerregistry" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  docker_email          = "test@email.com"
  docker_username       = "testuser"
  docker_password       = "secret"
}

data "betterado_serviceendpoint_dockerregistry" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = betterado_serviceendpoint_dockerregistry.test.service_endpoint_name
}
`, projectName, serviceEndpointName)
}

func hclDataServiceConnectionDockerRegistryWithID(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_dockerregistry" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  docker_email          = "test@email.com"
  docker_username       = "testuser"
  docker_password       = "secret"
}

data "betterado_serviceendpoint_dockerregistry" "test" {
  project_id          = betterado_project.test.id
  service_endpoint_id = betterado_serviceendpoint_dockerregistry.test.id
}
`, projectName, serviceEndpointName)
}
