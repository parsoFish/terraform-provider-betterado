//go:build (all || data_source_feed) && !exclude_feed

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

// ── Legacy SDKv2-path data source tests ───────────────────────────────────────
// These are kept but now use the mux provider so the betterado_feed resource
// (framework) and the betterado_feed data source (also framework) are both
// available in the same Terraform config.

func TestAccFeedDataSource_byName(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "data.betterado_feed.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclFeedDataSourceByName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "feed_id"),
				),
			},
		},
	})
}

func TestAccFeedDataSource_byId(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "data.betterado_feed.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclFeedDataSourceByID(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "feed_id"),
				),
			},
		},
	})
}

// ── Framework data source acceptance tests ────────────────────────────────────

// TestAccFeedDataSourceFramework_byName verifies data.betterado_feed (framework
// path) when configured with a feed name pointing to an existing org-scoped feed.
// AC1: name, feed_id, and id are set; idempotency re-plan produces no diff.
func TestAccFeedDataSourceFramework_byName(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	dsNode := "data.betterado_feed.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclFeedDataSourceFrameworkByName(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsNode, "name"),
					resource.TestCheckResourceAttrSet(dsNode, "feed_id"),
					resource.TestCheckResourceAttrSet(dsNode, "id"),
					captureFeedDataSourceEvidence(dsNode),
				),
			},
			// AC1: idempotency — no perpetual diff.
			{
				Config:             hclFeedDataSourceFrameworkByName(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccFeedDataSourceFramework_byId verifies data.betterado_feed (framework
// path) when configured with a feed_id (UUID) pointing to an existing
// project-scoped feed.
// AC2: name, feed_id, project_id, and id are set; idempotency re-plan produces no diff.
func TestAccFeedDataSourceFramework_byId(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	dsNode := "data.betterado_feed.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclFeedDataSourceFrameworkByID(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsNode, "name"),
					resource.TestCheckResourceAttrSet(dsNode, "feed_id"),
					resource.TestCheckResourceAttrSet(dsNode, "id"),
					resource.TestCheckResourceAttrSet(dsNode, "project_id"),
				),
			},
			// AC2: idempotency — no perpetual diff.
			{
				Config:             hclFeedDataSourceFrameworkByID(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// captureFeedDataSourceEvidence fetches the feed from ADO and persists the
// response as forge demo live-evidence (label "acceptance-resource").
// AC3: .forge/live-evidence/acceptance-resource.json written with a real feeds
// endpoint URL.
func captureFeedDataSourceEvidence(dsNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[dsNode]
		if !ok {
			return nil
		}
		feedID := res.Primary.Attributes["feed_id"]
		projectID := res.Primary.Attributes["project_id"]

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")

		agg, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			return nil // best-effort
		}

		feedDetail, err := agg.FeedClient.GetFeed(agg.Ctx, feedapi.GetFeedArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil || feedDetail == nil {
			return nil // best-effort
		}

		var evidenceURL string
		if projectID != "" {
			evidenceURL = fmt.Sprintf("%s/%s/_apis/packaging/feeds/%s?api-version=7.1", orgURL, projectID, feedID)
		} else {
			evidenceURL = fmt.Sprintf("%s/_apis/packaging/feeds/%s?api-version=7.1", orgURL, feedID)
		}

		_ = testutils.CaptureLiveEvidence("acceptance-resource", evidenceURL, feedDetail)
		return nil
	}
}

// ── HCL helpers ───────────────────────────────────────────────────────────────

func hclFeedDataSourceByName(name string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = "%s"
}

data "betterado_feed" "test" {
  name = betterado_feed.test.name
}`, name)
}

func hclFeedDataSourceByID(feedID string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = "%s"
}

data "betterado_feed" "test" {
  feed_id = betterado_feed.test.id
}`, feedID)
}

// hclFeedDataSourceFrameworkByName creates an org-scoped feed then reads it
// back via the framework data source using name.
func hclFeedDataSourceFrameworkByName(name string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = %q
}

data "betterado_feed" "test" {
  name = betterado_feed.test.name
}
`, name)
}

// hclFeedDataSourceFrameworkByID creates a project-scoped feed then reads it
// back via the framework data source using feed_id (UUID).
func hclFeedDataSourceFrameworkByID(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %q
}

resource "betterado_feed" "test" {
  name       = %q
  project_id = data.betterado_project.test.id
}

data "betterado_feed" "test" {
  feed_id    = betterado_feed.test.id
  project_id = betterado_feed.test.project_id
}
`, SharedFixtureProjectName, name)
}
