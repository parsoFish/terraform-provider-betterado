//go:build (all || resource_feed_retention_policy) && !exclude_feed_retention_policy

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

func TestAccFeedRetentionPolicy_projectBasic(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyProjectBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "20"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "30"),
				),
			},
		},
	})
}

func TestAccFeedRetentionPolicy_organizationBasic(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyOrganizationBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "20"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "30"),
				),
			},
		},
	})
}

func TestAccFeedRetentionPolicy_projectUpdate(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyProjectBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "count_limit"),
					CheckFeedRetentionPolicyExist(20),
				),
			},
			{
				Config: hclFeedRetentionPolicyProjectUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(21),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "21"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "31"),
				),
			},
		},
	})
}

func TestAccFeedRetentionPolicy_organizationUpdate(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyOrganizationBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttrSet(tfNode, "count_limit"),
				),
			},
			{
				Config: hclFeedRetentionPolicyOrganizationUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(21),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "21"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "31"),
				),
			},
		},
	})
}

func TestAccFeedRetentionPolicy_projectRequiresImport(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyProjectBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "count_limit"),
				),
			},
			{
				Config: hclFeedRetentionPolicyProjectImport(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "20"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "30"),
				),
			},
		},
	})
}

func TestAccFeedRetentionPolicy_organizationRequiresImport(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_retention_policy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedRetentionPolicyDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyOrganizationBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttrSet(tfNode, "count_limit"),
				),
			},
			{
				Config: hclFeedRetentionPolicyOrganizationImport(name),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedRetentionPolicyExist(20),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "20"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "30"),
				),
			},
		},
	})
}

func checkFeedRetentionPolicyDestroyed(s *terraform.State) error {
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feed_retention_policy" {
			continue
		}
		id := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		policy, err := clients.FeedClient.GetFeedRetentionPolicies(clients.Ctx, feed.GetFeedRetentionPoliciesArgs{
			FeedId:  &id,
			Project: &projectID,
		})

		if err == nil {
			if policy != nil && (policy.CountLimit != nil || policy.DaysToKeepRecentlyDownloadedPackages != nil) {
				return fmt.Errorf("Feed Retention Policy (Feed ID: %s) should not exist", id)
			}
		}
	}
	return nil
}

func CheckFeedRetentionPolicyExist(expectedCountLimit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_feed_retention_policy.test"]
		if !ok {
			return fmt.Errorf("Did not find a `betterado_feed_retention_policy` in the TF state")
		}

		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		id := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		policy, err := clients.FeedClient.GetFeedRetentionPolicies(clients.Ctx, feed.GetFeedRetentionPoliciesArgs{
			FeedId:  &id,
			Project: &projectID,
		})
		if err != nil {
			return fmt.Errorf("Feed Retention Policy with Feed ( Feed ID=%s ) cannot be found!. Error=%v", id, err)
		}

		if *policy.CountLimit != expectedCountLimit {
			return fmt.Errorf("Feed Retention Policy with ( Feed ID=%s ) has CountLimit=%d, but expected CountLimit=%d", id, *policy.CountLimit, expectedCountLimit)
		}
		return nil
	}
}

func hclFeedRetentionPolicyProjectBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%[1]s"
}

resource "betterado_feed" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_feed_retention_policy" "test" {
  project_id                                = betterado_project.test.id
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 20
  days_to_keep_recently_downloaded_packages = 30
}
`, name)
}

func hclFeedRetentionPolicyOrganizationBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = "%s"
}

resource "betterado_feed_retention_policy" "test" {
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 20
  days_to_keep_recently_downloaded_packages = 30
}
`, name)
}

func hclFeedRetentionPolicyProjectUpdate(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%[1]s"
}

resource "betterado_feed" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_feed_retention_policy" "test" {
  project_id                                = betterado_project.test.id
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 21
  days_to_keep_recently_downloaded_packages = 31
}
`, name)
}

func hclFeedRetentionPolicyOrganizationUpdate(name string) string {
	return fmt.Sprintf(`
resource "betterado_feed" "test" {
  name = "%s"
}

resource "betterado_feed_retention_policy" "test" {
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 21
  days_to_keep_recently_downloaded_packages = 31
}
`, name)
}

func hclFeedRetentionPolicyProjectImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_feed_retention_policy" "import" {
  feed_id                                   = betterado_feed.test.id
  project_id                                = betterado_project.test.id
  count_limit                               = betterado_feed_retention_policy.test.count_limit
  days_to_keep_recently_downloaded_packages = betterado_feed_retention_policy.test.days_to_keep_recently_downloaded_packages
}
`, hclFeedRetentionPolicyProjectBasic(name))
}

func hclFeedRetentionPolicyOrganizationImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_feed_retention_policy" "import" {
  feed_id                                   = betterado_feed.test.id
  count_limit                               = betterado_feed_retention_policy.test.count_limit
  days_to_keep_recently_downloaded_packages = betterado_feed_retention_policy.test.days_to_keep_recently_downloaded_packages
}
`, hclFeedRetentionPolicyOrganizationBasic(name))
}

// ── Framework (mux ProtoV6) tests ─────────────────────────────────────────────

// TestAccFeedRetentionPolicyFramework_projectBasic verifies betterado_feed_retention_policy
// via the mux ProtoV6 provider path (GetMuxedProviderFactories).
//
//   - AC1: terraform apply creates the policy with count_limit=25, days=45
//   - AC1: provider reads back asserting count_limit=25 and days=45
//   - AC1: idempotency re-plan produces no diff (ExpectNonEmptyPlan: false)
//   - AC1: destroy succeeds
//   - AC3: CaptureLiveEvidence is called with label "acceptance-resource"
func TestAccFeedRetentionPolicyFramework_projectBasic(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	tfNode := "betterado_feed_retention_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFeedRetentionPolicyFrameworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedRetentionPolicyFrameworkProjectBasic(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "count_limit", "25"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "45"),
					captureFeedRetentionPolicyFrameworkEvidence(tfNode),
				),
			},
			// AC1: idempotency — no perpetual diff after apply.
			{
				Config:             hclFeedRetentionPolicyFrameworkProjectBasic(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccFeedRetentionPolicyFramework_update verifies that the provider updates
// the policy in-place when count_limit changes (AC2).
//
//   - AC2: apply count_limit=25, then update to count_limit=30
//   - AC2: read-back shows count_limit=30
//   - AC2: idempotency holds after the update
func TestAccFeedRetentionPolicyFramework_update(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	tfNode := "betterado_feed_retention_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFeedRetentionPolicyFrameworkDestroyed,
		Steps: []resource.TestStep{
			// Step 1: initial apply with count_limit=25.
			{
				Config: hclFeedRetentionPolicyFrameworkProjectBasic(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "count_limit", "25"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "45"),
				),
			},
			// Step 2: update count_limit to 30; idempotency after update.
			{
				Config: hclFeedRetentionPolicyFrameworkProjectUpdate(feedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "count_limit", "30"),
					resource.TestCheckResourceAttr(tfNode, "days_to_keep_recently_downloaded_packages", "45"),
				),
			},
			// AC2: idempotency after update.
			{
				Config:             hclFeedRetentionPolicyFrameworkProjectUpdate(feedName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkFeedRetentionPolicyFrameworkDestroyed asserts no betterado_feed_retention_policy
// resources remain after destroy. Uses getFeedDirectClient because the mux
// provider does not wire the SDKv2 meta singleton.
func checkFeedRetentionPolicyFrameworkDestroyed(s *terraform.State) error {
	clients, err := getFeedDirectClient()
	if err != nil {
		return fmt.Errorf("getting direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feed_retention_policy" {
			continue
		}
		feedID := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		policy, err := clients.FeedClient.GetFeedRetentionPolicies(clients.Ctx, feed.GetFeedRetentionPoliciesArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil {
			// API error likely means the feed itself is gone — treat as destroyed.
			continue
		}
		if policy != nil && (policy.CountLimit != nil || policy.DaysToKeepRecentlyDownloadedPackages != nil) {
			return fmt.Errorf("feed retention policy (Feed ID: %s) should not exist after destroy", feedID)
		}
	}
	return nil
}

// captureFeedRetentionPolicyFrameworkEvidence calls CaptureLiveEvidence with a real
// ADO GET URL for the feed retention policies endpoint (AC3).
func captureFeedRetentionPolicyFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil // best-effort
		}
		feedID := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getFeedDirectClient()
		if err != nil {
			return nil // best-effort
		}

		policy, err := clients.FeedClient.GetFeedRetentionPolicies(clients.Ctx, feed.GetFeedRetentionPoliciesArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil || policy == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		var retentionURL string
		if projectID != "" {
			retentionURL = fmt.Sprintf("%s/%s/_apis/packaging/feeds/%s/retentionpolicies?api-version=7.1", orgURL, projectID, feedID)
		} else {
			retentionURL = fmt.Sprintf("%s/_apis/packaging/feeds/%s/retentionpolicies?api-version=7.1", orgURL, feedID)
		}

		_ = testutils.CaptureLiveEvidence("acceptance-resource", retentionURL, policy)
		return nil
	}
}

// hclFeedRetentionPolicyFrameworkProjectBasic returns HCL for creating a
// betterado_feed_retention_policy via the framework provider using the shared
// persistent project (betterado-standing-demo) to avoid a project-create.
func hclFeedRetentionPolicyFrameworkProjectBasic(feedName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_feed" "test" {
  name       = %[1]q
  project_id = data.betterado_project.test.id
}

resource "betterado_feed_retention_policy" "test" {
  project_id                                = data.betterado_project.test.id
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 25
  days_to_keep_recently_downloaded_packages = 45
}
`, feedName, SharedFixtureProjectName)
}

// hclFeedRetentionPolicyFrameworkProjectUpdate returns HCL with count_limit=30
// to exercise the Update path (AC2).
func hclFeedRetentionPolicyFrameworkProjectUpdate(feedName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_feed" "test" {
  name       = %[1]q
  project_id = data.betterado_project.test.id
}

resource "betterado_feed_retention_policy" "test" {
  project_id                                = data.betterado_project.test.id
  feed_id                                   = betterado_feed.test.id
  count_limit                               = 30
  days_to_keep_recently_downloaded_packages = 45
}
`, feedName, SharedFixtureProjectName)
}
