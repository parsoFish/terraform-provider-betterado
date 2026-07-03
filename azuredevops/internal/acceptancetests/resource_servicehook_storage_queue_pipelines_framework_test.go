//go:build (all || resource_servicehook_storage_queue_pipelines) && !exclude_servicehooks

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServicehookStorageQueuePipelinesFramework_basic(t *testing.T) {
	fixture := SharedReleaseFixture(t) // reuses betterado-standing-demo — no new project
	queueName := "acctest-queue"
	accountKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	stateFilter := "Completed"
	resultFilter := "Succeeded"
	tfNode := "betterado_servicehook_storage_queue_pipelines.fw_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkServicehookStorageQueuePipelinesFrameworkDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back + capture evidence
			{
				Config: hclServicehookStorageQueuePipelinesFramework(fixture.ProjectID, accountKey, queueName, stateFilter, resultFilter),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "queue_name", queueName),
					resource.TestCheckResourceAttr(tfNode, "account_key", accountKey),
					resource.TestCheckResourceAttr(tfNode, "stage_state_changed_event.0.stage_state_filter", stateFilter),
					resource.TestCheckResourceAttr(tfNode, "stage_state_changed_event.0.stage_result_filter", resultFilter),
					captureServicehookStorageQueueEvidence(tfNode),
				),
			},
			// Step 2: idempotency
			{
				Config:             hclServicehookStorageQueuePipelinesFramework(fixture.ProjectID, accountKey, queueName, stateFilter, resultFilter),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclServicehookStorageQueuePipelinesFramework(projectID, accountKey, queueName, stateFilter, resultFilter string) string {
	return fmt.Sprintf(`
resource "betterado_servicehook_storage_queue_pipelines" "fw_test" {
  project_id  = %[1]q
  account_key = %[2]q
  queue_name  = %[3]q

  stage_state_changed_event {
    stage_state_filter  = %[4]q
    stage_result_filter = %[5]q
  }
}
`, projectID, accountKey, queueName, stateFilter, resultFilter)
}

// checkServicehookStorageQueuePipelinesFrameworkDestroyed verifies the subscription
// is gone after destroy. Uses getDirectClient (defined in resource_task_group_test.go)
// because ProtoV6ProviderFactories does not wire the SDKv2 singleton's Meta.
func checkServicehookStorageQueuePipelinesFrameworkDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkServicehookStorageQueuePipelinesFrameworkDestroyed: build client: %w", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_servicehook_storage_queue_pipelines" {
			continue
		}
		subID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid subscription ID %q: %w", res.Primary.ID, err)
		}
		_, err = clients.ServiceHooksClient.GetSubscription(clients.Ctx, servicehooks.GetSubscriptionArgs{
			SubscriptionId: &subID,
		})
		if err == nil {
			return fmt.Errorf("servicehook subscription %s still exists after destroy", res.Primary.ID)
		}
	}
	return nil
}

// captureServicehookStorageQueueEvidence performs a live ADO REST GET of the
// created subscription and persists it as forge demo live-evidence (before destroy).
// Best-effort: a capture failure never fails the acceptance test.
func captureServicehookStorageQueueEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		subID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil
		}
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort; do not fail the test
		}
		subscription, err := clients.ServiceHooksClient.GetSubscription(clients.Ctx, servicehooks.GetSubscriptionArgs{
			SubscriptionId: &subID,
		})
		if err != nil || subscription == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		apiURL := fmt.Sprintf("%s/_apis/hooks/subscriptions/%s?api-version=7.1", orgURL, subID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, subscription)
		return nil
	}
}
