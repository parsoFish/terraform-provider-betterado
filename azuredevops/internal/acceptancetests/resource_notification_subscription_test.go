//go:build (all || resource_notification_subscription) && !exclude_resource_notification_subscription

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// TestAccNotificationSubscription_basic creates an ADO notification subscription,
// verifies it is read back correctly (idempotency check), captures live API evidence,
// and confirms destroy removes it.
//
// Requires: TF_ACC=1, AZDO_ORG_SERVICE_URL, AZDO_PERSONAL_ACCESS_TOKEN, AZDO_TEST_AAD_USER_EMAIL
func TestAccNotificationSubscription_basic(t *testing.T) {
	tfNode := "betterado_notification_subscription.test"
	email := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkNotificationSubscriptionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back
			{
				Config: hclNotificationSubscriptionBasic(email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "subscription_type", "ms.vss-work.workitem-changed-event"),
					resource.TestCheckResourceAttr(tfNode, "channel_type", "EmailHtml"),
					resource.TestCheckResourceAttr(tfNode, "channel_address", email),
					captureNotificationSubscriptionEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclNotificationSubscriptionBasic(email),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclNotificationSubscriptionBasic returns HCL that creates a notification subscription
// scoped to the shared fixture project. It uses betterado_identity_user to resolve
// the subscriber_id from AZDO_TEST_AAD_USER_EMAIL.
func hclNotificationSubscriptionBasic(email string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

data "betterado_identity_user" "subscriber" {
  mail = %[2]q
}

resource "betterado_notification_subscription" "test" {
  project_id        = data.betterado_project.test.id
  subscription_type = "ms.vss-work.workitem-changed-event"
  subscriber_id     = data.betterado_identity_user.subscriber.id
  channel_type      = "EmailHtml"
  channel_address   = %[2]q
}
`, SharedFixtureProjectName, email)
}

// checkNotificationSubscriptionDestroyed verifies that all notification subscriptions
// in the Terraform state have been deleted from ADO.
func checkNotificationSubscriptionDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkNotificationSubscriptionDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_notification_subscription" {
			continue
		}

		subID := res.Primary.ID
		sub, err := clients.NotificationClient.GetSubscription(clients.Ctx, notificationapi.GetSubscriptionArgs{
			SubscriptionId: &subID,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				// Subscription is gone — expected.
				continue
			}
			return fmt.Errorf("error reading notification subscription %s after destroy: %v", subID, err)
		}
		if sub != nil && sub.Id != nil {
			return fmt.Errorf("notification subscription %s still exists after destroy", subID)
		}
	}

	return nil
}

// captureNotificationSubscriptionEvidence performs a real live API GET of the created
// subscription and persists the response as forge demo live-evidence (before the
// resource is destroyed). Best-effort: a capture failure never fails the test.
func captureNotificationSubscriptionEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		subID := res.Primary.ID
		if subID == "" {
			return nil
		}

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}

		sub, err := clients.NotificationClient.GetSubscription(clients.Ctx, notificationapi.GetSubscriptionArgs{
			SubscriptionId: &subID,
		})
		if err != nil || sub == nil {
			return nil // best-effort
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/notification/subscriptions/%s?api-version=7.1", orgURL, subID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, sub)
		return nil
	}
}
