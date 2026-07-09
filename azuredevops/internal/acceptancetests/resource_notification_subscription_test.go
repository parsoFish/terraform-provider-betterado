//go:build (all || resource_notification_subscription) && !exclude_resource_notification_subscription

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// TestAccNotificationSubscription_basic creates an ADO notification subscription,
// verifies it is read back correctly (idempotency check), captures live API evidence,
// and confirms destroy removes it.
//
// Requires: TF_ACC=1, AZDO_ORG_SERVICE_URL, AZDO_PERSONAL_ACCESS_TOKEN
// AZDO_TEST_AAD_USER_EMAIL is optional — if absent, a real user is auto-discovered.
func TestAccNotificationSubscription_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live acceptance test")
	}

	email, subscriberID := resolveNotificationSubscriberDirect(t)

	tfNode := "betterado_notification_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkNotificationSubscriptionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back
			{
				Config: hclNotificationSubscriptionBasic(email, subscriberID),
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
				Config:             hclNotificationSubscriptionBasic(email, subscriberID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// resolveNotificationSubscriberDirect resolves a real ADO user for the notification
// subscription test without requiring AZDO_TEST_AAD_USER_EMAIL.
//
// Precedence:
//  1. AZDO_TEST_AAD_USER_EMAIL is set → use that email, look up the identity UUID.
//  2. Fall back: page through GraphClient.ListUsers and pick the first user with
//     a non-empty MailAddress; look up that user's identity UUID.
func resolveNotificationSubscriberDirect(t *testing.T) (email, identityUUID string) {
	t.Helper()

	clients, err := getDirectClient()
	if err != nil {
		t.Fatalf("resolveNotificationSubscriberDirect: getDirectClient: %v", err)
	}

	// ── 1. Prefer an explicitly-set email env var ─────────────────────────────
	if envEmail := os.Getenv("AZDO_TEST_AAD_USER_EMAIL"); envEmail != "" {
		uuid, resolveErr := resolveIdentityUUIDByEmailDirect(t, clients, envEmail)
		if resolveErr != nil || uuid == "" {
			t.Fatalf("resolveNotificationSubscriberDirect: could not resolve identity for AZDO_TEST_AAD_USER_EMAIL=%q: %v", envEmail, resolveErr)
		}
		return envEmail, uuid
	}

	// ── 2. Auto-discover via GraphClient.ListUsers ───────────────────────────
	// ContinuationToken in ListUsersArgs is *string; in the response it is *[]string.
	subjectTypes := []string{"aad", "msa"}
	var contToken string
	for {
		args := graph.ListUsersArgs{
			SubjectTypes: &subjectTypes,
		}
		if contToken != "" {
			args.ContinuationToken = &contToken
		}
		page, listErr := clients.GraphClient.ListUsers(clients.Ctx, args)
		if listErr != nil {
			t.Fatalf("resolveNotificationSubscriberDirect: GraphClient.ListUsers: %v", listErr)
		}
		if page == nil || page.GraphUsers == nil {
			break
		}
		for _, u := range *page.GraphUsers {
			if u.MailAddress == nil || *u.MailAddress == "" {
				continue
			}
			mail := *u.MailAddress
			uuid, resolveErr := resolveIdentityUUIDByEmailDirect(t, clients, mail)
			if resolveErr != nil || uuid == "" {
				continue // try next user
			}
			t.Logf("resolveNotificationSubscriberDirect: auto-discovered subscriber email=%q uuid=%q", mail, uuid)
			return mail, uuid
		}
		if page.ContinuationToken == nil || len(*page.ContinuationToken) == 0 || (*page.ContinuationToken)[0] == "" {
			break
		}
		contToken = (*page.ContinuationToken)[0]
	}

	t.Fatalf("resolveNotificationSubscriberDirect: no ADO user with a valid mail address and resolvable identity found")
	return "", ""
}

// resolveIdentityUUIDByEmailDirect returns the identity UUID (GUID) for a given email
// address by calling IdentityClient.ReadIdentities with SearchFilter="MailAddress".
// Returns ("", nil) if no identity is found (not an error — caller may skip).
func resolveIdentityUUIDByEmailDirect(t *testing.T, clients *client.AggregatedClient, email string) (string, error) {
	t.Helper()
	filter := "MailAddress"
	identities, err := clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
		SearchFilter: &filter,
		FilterValue:  &email,
	})
	if err != nil {
		return "", fmt.Errorf("IdentityClient.ReadIdentities(MailAddress=%q): %w", email, err)
	}
	if identities == nil || len(*identities) == 0 {
		return "", nil
	}
	for _, id := range *identities {
		if id.Id != nil {
			return id.Id.String(), nil
		}
	}
	return "", nil
}

// hclNotificationSubscriptionBasic returns HCL that creates a notification subscription.
// subscriber_id is passed as a literal UUID string so no identity data source lookup is needed.
func hclNotificationSubscriptionBasic(email, subscriberID string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_notification_subscription" "test" {
  project_id        = data.betterado_project.test.id
  subscription_type = "ms.vss-work.workitem-changed-event"
  subscriber_id     = %[2]q
  channel_type      = "EmailHtml"
  channel_address   = %[3]q
}
`, SharedFixtureProjectName, subscriberID, email)
}

// checkNotificationSubscriptionDestroyed verifies that all notification subscriptions
// in the Terraform state have been deleted from ADO.
//
// ADO's notification subscription DELETE is asynchronous: after the API call
// succeeds (HTTP 204) the subscription transitions to status "pendingDeletion"
// before it is fully removed. A subsequent GET still returns the subscription
// with that status. We therefore treat both a 404 response AND a subscription
// whose status is "pendingDeletion" as "effectively deleted".
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
			// ADO soft-deletes subscriptions: after DELETE the status transitions
			// to "pendingDeletion" before the record is removed. Treat that as
			// destroyed so the test does not report a false failure.
			if sub.Status != nil && *sub.Status == notificationapi.SubscriptionStatusValues.PendingDeletion {
				continue
			}
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
