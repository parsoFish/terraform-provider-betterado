//go:build (all || data_agent_queue) && !exclude_data_agent_queue

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccAgentQueueDataSource_Basic reads the default "Azure Pipelines" hosted
// queue from the standing fixture project — no project creation required.
func TestAccAgentQueueDataSource_Basic(t *testing.T) {
	agentQueueName := "Azure Pipelines"
	tfNode := "data.betterado_agent_queue.queue"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclAgentQueueDataSourceBasic(agentQueueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "name", agentQueueName),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "agent_pool_id"),
					captureAgentQueueEvidence(tfNode),
				),
			},
			// Idempotency — no perpetual diff.
			{
				Config:             hclAgentQueueDataSourceBasic(agentQueueName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclAgentQueueDataSourceBasic uses the standing fixture project so no project is created.
func hclAgentQueueDataSourceBasic(queueName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

data "betterado_agent_queue" "queue" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
}
`, queueName, SharedFixtureProjectName)
}

// captureAgentQueueEvidence performs a real live API GET and writes forge demo
// live-evidence to .forge/live-evidence/acceptance-resource-agent-queue.json (AC3).
// Best-effort: never fails the test.
func captureAgentQueueEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		queueIDStr := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		queueID, err := strconv.Atoi(queueIDStr)
		if err != nil {
			return nil
		}
		clients, err := getDirectClient()
		if err != nil {
			return nil
		}
		queue, err := clients.TaskAgentClient.GetAgentQueue(clients.Ctx, taskagent.GetAgentQueueArgs{
			QueueId: &queueID,
			Project: &projectID,
		})
		if err != nil || queue == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/distributedtask/queues/%s?api-version=7.1", orgURL, projectID, queueIDStr)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-agent-queue", url, queue)
		return nil
	}
}
