//go:build (all || resource_agent_queue) && !exclude_resource_agent_queue

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccAgentQueue_basedOnPool creates an agent pool and wires it to a queue
// in the standing fixture project (no project creation).
func TestAccAgentQueue_basedOnPool(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclAgentQueueBasedOnPool(poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "agent_pool_id"),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			// Idempotency — no perpetual diff.
			{
				Config:             hclAgentQueueBasedOnPool(poolName),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// hclAgentQueueBasedOnPool uses the standing fixture project so no project is created.
func hclAgentQueueBasedOnPool(poolName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_agent_pool" "test" {
  name           = %[1]q
  auto_provision = false
  auto_update    = false
  pool_type      = "automation"
}

resource "betterado_agent_queue" "test" {
  project_id    = data.betterado_project.fixture.id
  agent_pool_id = betterado_agent_pool.test.id
}
`, poolName, SharedFixtureProjectName)
}
