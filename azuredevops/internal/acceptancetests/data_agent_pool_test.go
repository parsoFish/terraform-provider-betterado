//go:build (all || data_source_agent_pool) && !exclude_data_source_agent_pool

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccAgentPoolDataSource_basic(t *testing.T) {
	agentPoolName := testutils.GenerateResourceName()
	tfNode := "data.betterado_agent_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceAgentPoolBasic(agentPoolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "name", agentPoolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
				),
			},
			// Idempotency — no perpetual diff
			{
				Config:             hclDataSourceAgentPoolBasic(agentPoolName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclDataSourceAgentPoolBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_agent_pool" "test" {
  name           = "%s"
  auto_provision = false
  auto_update    = false
  pool_type      = "automation"
}

data "betterado_agent_pool" "test" {
  name = betterado_agent_pool.test.name
}

`, name)
}
