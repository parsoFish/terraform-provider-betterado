//go:build (all || resource_feed) && !exclude_feed

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	feedapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccFeedFramework_basic verifies the betterado_feed resource (framework
// path) for an org-scoped feed: create → read-back with name and id set →
// idempotency re-plan → destroy.
func TestAccFeedFramework_basic(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	tfNode := "betterado_feed.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFeedFrameworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedFrameworkBasic(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", feedName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureFeedFrameworkEvidence(tfNode),
				),
			},
			// Idempotency — no perpetual diff.
			{
				Config:             hclFeedFrameworkBasic(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccFeedFramework_withProject verifies the betterado_feed resource
// (framework path) for a project-scoped feed: create → read-back with name,
// id and project_id set → idempotency re-plan → destroy.
//
// Uses SharedFixtureProjectName (betterado-standing-demo) via a data source
// so no new ADO project is created — the org is at the 1000-project cap, so
// any project-create attempt would fail immediately.
func TestAccFeedFramework_withProject(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	tfNode := "betterado_feed.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFeedFrameworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedFrameworkWithProject(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", feedName),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
				),
			},
			// Idempotency — no perpetual diff.
			{
				Config:             hclFeedFrameworkWithProject(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkFeedFrameworkDestroyed asserts that no betterado_feed resource remains
// after destroy. Uses getDirectClient() (same as task group tests) because the
// mux provider does not wire the SDKv2 meta singleton.
//
// ADO's DeleteFeed is a soft delete: the feed moves to a recycle bin but
// GetFeed by GUID still returns it (with DeletedDate set). We treat a feed
// with a non-nil DeletedDate as "destroyed" — it is no longer an active feed
// and Terraform destroy has done its job.
func checkFeedFrameworkDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("getting direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feed" {
			continue
		}
		id := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		feedDetail, err := clients.FeedClient.GetFeed(clients.Ctx, feedapi.GetFeedArgs{
			FeedId:  &id,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil {
			// 404 or any error → feed is gone, which is what we want.
			continue
		}
		// Feed still returned by ADO — check whether it is in the soft-deleted
		// (recycle-bin) state. DeletedDate is set when DeleteFeed has been called.
		if feedDetail != nil && feedDetail.DeletedDate != nil {
			// Soft-deleted: Terraform destroy succeeded; recycle-bin cleanup is
			// handled by ADO automatically (or by permanent_delete=true).
			continue
		}
		return fmt.Errorf("feed (ID: %s) should not exist after destroy", id)
	}
	return nil
}

// captureFeedFrameworkEvidence performs a real live API GET of the created feed
// and persists the response as forge demo live-evidence. Best-effort: a capture
// failure never fails the test.
func captureFeedFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		feedID := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getFeedDirectClient()
		if err != nil {
			return nil // best-effort
		}

		feedDetail, err := clients.FeedClient.GetFeed(clients.Ctx, feedapi.GetFeedArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil || feedDetail == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		var url string
		if projectID != "" {
			url = fmt.Sprintf("%s/%s/_apis/packaging/feeds/%s?api-version=7.1", orgURL, projectID, feedID)
		} else {
			url = fmt.Sprintf("%s/_apis/packaging/feeds/%s?api-version=7.1", orgURL, feedID)
		}

		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, feedDetail)
		return nil
	}
}

// getFeedDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used because ProtoV6ProviderFactories does not wire the SDKv2 singleton Meta.
func getFeedDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

func hclFeedFrameworkBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = %q
}
`, name)
}

func hclFeedFrameworkWithProject(feedName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_feed" "test" {
  name       = %[1]q
  project_id = data.betterado_project.test.id
}
`, feedName, SharedFixtureProjectName)
}

// nilIfEmptyStr returns nil when s is empty, otherwise &s.
// Prevents passing "" as an ADO project parameter (org-scoped feeds
// should receive nil, not empty string, in the feed REST API).
func nilIfEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
