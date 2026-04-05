package acceptancetests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccAgentQueueDataSource_Basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	agentQueueName := "Azure Pipelines"
	agentQueueData := testutils.HclAgentQueueDataSource(projectName, agentQueueName)

	tfNode := "data.betterado_agent_queue.queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: agentQueueData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "name", agentQueueName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "agent_pool_id"),
				),
			},
		},
	})
}
