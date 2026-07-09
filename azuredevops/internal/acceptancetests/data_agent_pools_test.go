//go:build (all || data_source_agent_pools) && !exclude_data_source_agent_pools

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// Verifies that the following sequence of events occurs without error:
//
//	(1) TF can create agent pools (framework resource)
//	(2) A data source is added to the configuration, and that data source can find the created pools
func TestAccAgentPoolsDataSource_Basic(t *testing.T) {
	agentPoolName := testutils.GenerateResourceName()
	agentPool1Name := agentPoolName + "_1"
	agentPool2Name := agentPoolName + "_2"

	createAgent1 := testutils.HclAgentPoolResourceAppendPoolNameToResourceName(agentPool1Name)
	createAgent2 := testutils.HclAgentPoolResourceAppendPoolNameToResourceName(agentPool2Name)
	agentPoolsData := testutils.HclAgentPoolsDataSource()
	createAgentPools := fmt.Sprintf("%s\n%s", createAgent1, createAgent2)

	tfNode := "data.betterado_agent_pools.pools"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: createAgentPools,
			},
			{
				Config: agentPoolsData,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckNestedKeyExistsWithValue(tfNode, "name", agentPool1Name),
					testutils.CheckNestedKeyExistsWithValue(tfNode, "name", agentPool2Name),
				),
			},
			// Idempotency — no perpetual diff
			{
				Config:             agentPoolsData,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
