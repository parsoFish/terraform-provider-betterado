//go:build (all || resource_servicehook_webhook_tfs) && !exclude_servicehooks

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

func TestAccServicehookWebhookTfsFramework_basic(t *testing.T) {
	fixture := SharedReleaseFixture(t) // reuses betterado-standing-demo — no new project
	url := "https://example.com/webhook-fw"
	tfNode := "betterado_servicehook_webhook_tfs.fw_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkServicehookWebhookTfsFrameworkDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + read-back + capture evidence
			{
				Config: hclServicehookWebhookTfsFramework(fixture.ProjectID, url),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "url", url),
					resource.TestCheckResourceAttr(tfNode, "git_push.#", "1"),
					captureServicehookWebhookTfsEvidence(tfNode),
				),
			},
			// Step 2: idempotency
			{
				Config:             hclServicehookWebhookTfsFramework(fixture.ProjectID, url),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclServicehookWebhookTfsFramework(projectID, url string) string {
	return fmt.Sprintf(`
resource "betterado_servicehook_webhook_tfs" "fw_test" {
  project_id = %[1]q
  url        = %[2]q
  git_push {}
}
`, projectID, url)
}

// checkServicehookWebhookTfsFrameworkDestroyed verifies the subscription
// is gone after destroy. Uses getDirectClient (defined in resource_task_group_test.go)
// because ProtoV6ProviderFactories does not wire the SDKv2 singleton's Meta.
func checkServicehookWebhookTfsFrameworkDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkServicehookWebhookTfsFrameworkDestroyed: build client: %w", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_servicehook_webhook_tfs" {
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

// captureServicehookWebhookTfsEvidence performs a live ADO REST GET of the
// created subscription and persists it as forge demo live-evidence (before destroy).
// Best-effort: a capture failure never fails the acceptance test.
func captureServicehookWebhookTfsEvidence(tfNode string) resource.TestCheckFunc {
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
