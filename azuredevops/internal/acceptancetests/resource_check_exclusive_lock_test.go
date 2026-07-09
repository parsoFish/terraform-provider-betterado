package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccCheckExclusiveLock_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceEndpointName := testutils.GenerateResourceName()
	timeout := 43200
	newTimeout := 21600

	resourceType := "betterado_check_exclusive_lock"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckExclusiveLockResourceBasic(projectID, serviceEndpointName, timeout),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfCheckNode, "target_resource_id"),
					resource.TestCheckResourceAttrSet(tfCheckNode, "target_resource_type"),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", fmt.Sprintf("%d", timeout)),
				),
			},
			{
				Config: hclCheckExclusiveLockResourceBasic(projectID, serviceEndpointName, newTimeout),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfCheckNode, "target_resource_id"),
					resource.TestCheckResourceAttrSet(tfCheckNode, "target_resource_type"),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", fmt.Sprintf("%d", newTimeout)),
				),
			},
		},
	})
}

func hclCheckExclusiveLockResourceBasic(projectID string, serviceEndpointName string, timeout int) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_check_exclusive_lock" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"
  timeout              = %d
}`, projectID, serviceEndpointName, projectID, timeout)
}
