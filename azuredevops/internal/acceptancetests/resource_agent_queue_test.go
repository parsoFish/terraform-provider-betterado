package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccResourceAgentQueue_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAgentQueueBasic(projectName, poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			}, {
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAgentQueue_basedOnPool(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAgentQueueBasedObnPool(projectName, poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "agent_pool_id"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			}, {
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclAgentQueueBasic(projectName, queueName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_agent_queue" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
}
`, projectName, queueName)
}

func hclAgentQueueBasedObnPool(projectName, poolName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_agent_pool" "test" {
  name           = "%s"
  auto_provision = false
  auto_update    = false
  pool_type      = "automation"
}

resource "betterado_agent_queue" "test" {
  project_id    = betterado_project.test.id
  agent_pool_id = betterado_agent_pool.test.id
}
`, projectName, poolName)
}
