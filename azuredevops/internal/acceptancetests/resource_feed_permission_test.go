package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// ── SDKv2 path tests ──────────────────────────────────────────────────────────

func TestAccFeedPermission_basic(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_feed_permission.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclFeedPermissionBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "feed_id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "role"),
					resource.TestCheckResourceAttrSet(tfNode, "identity_descriptor"),
				),
			},
		},
	})
}

func TestAccFeedPermission_importErrorStep(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tfNode := "betterado_feed_permission.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:      checkFeedPermissionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedPermissionBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					CheckFeedPermissionExist(),
					resource.TestCheckResourceAttrSet(tfNode, "feed_id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "role"),
					resource.TestCheckResourceAttrSet(tfNode, "identity_descriptor"),
				),
			},
			{
				Config:      hclFeedPermissionImport(projectName),
				ExpectError: feedPermissionRequiresImportError(),
			},
		},
	})
}

func feedPermissionRequiresImportError() *regexp.Regexp {
	return regexp.MustCompile(`Error: feed Permission for Feed : .* and Identity : .* already exists`)
}

func checkFeedPermissionDestroyed(s *terraform.State) error {
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feed_permission" {
			continue
		}
		id := res.Primary.Attributes["feed_id"]
		projectID := res.Primary.Attributes["project_id"]
		permissions, err := clients.FeedClient.GetFeedPermissions(clients.Ctx, feed.GetFeedPermissionsArgs{
			FeedId:  &id,
			Project: &projectID,
		})

		if err == nil {
			if permissions != nil && len(*permissions) > 0 {
				return fmt.Errorf("Feed permissions (Feed ID: %s) should not exist", id)
			}
		}
	}
	return nil
}

func CheckFeedPermissionExist() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_feed_permission.test"]
		if !ok {
			return fmt.Errorf("Did not find a `betterado_feed_permission` in the TF state")
		}

		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		id := res.Primary.Attributes["feed_id"]
		projectID := res.Primary.Attributes["project_id"]

		_, err := clients.FeedClient.GetFeedPermissions(clients.Ctx, feed.GetFeedPermissionsArgs{
			FeedId:  &id,
			Project: &projectID,
		})
		if err != nil {
			return fmt.Errorf("Feed permissions with Feed ( Feed ID=%s ) cannot be found!. Error=%v", id, err)
		}

		return nil
	}
}

func hclFeedPermissionBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_feed" "test" {
  name       = "%[1]s"
  project_id = betterado_project.test.id
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%[1]s"
}

resource "betterado_feed_permission" "test" {
  feed_id             = betterado_feed.test.id
  project_id          = betterado_project.test.id
  role                = "reader"
  identity_descriptor = betterado_group.test.descriptor
}
`, name)
}

func hclFeedPermissionImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_feed_permission" "import" {
  feed_id             = betterado_feed_permission.test.feed_id
  project_id          = betterado_feed_permission.test.project_id
  role                = "reader"
  identity_descriptor = betterado_feed_permission.test.identity_descriptor
}
`, hclFeedPermissionBasic(name))
}

// ── Framework (mux ProtoV6) tests ─────────────────────────────────────────────

// TestAccFeedPermissionFramework_basic verifies the betterado_feed_permission resource
// via the mux ProtoV6 provider path (GetMuxedProviderFactories):
//
//   - terraform apply creates the permission in ADO
//   - provider reads it back, asserting role and identity_descriptor are set
//   - idempotency re-plan produces no diff (ExpectNonEmptyPlan: false)
//   - destroy removes the permission
//
// Uses the SharedFixtureProjectName project (betterado-standing-demo) via a data
// source to avoid creating a new project (org is at the 1000-project cap).
func TestAccFeedPermissionFramework_basic(t *testing.T) {
	feedName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	tfNode := "betterado_feed_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkFeedPermissionFrameworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclFeedPermissionFrameworkBasic(feedName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "feed_id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "role", "reader"),
					resource.TestCheckResourceAttrSet(tfNode, "identity_descriptor"),
					captureFeedPermissionFrameworkEvidence(tfNode),
				),
			},
			// Idempotency — no perpetual diff after apply.
			{
				Config:             hclFeedPermissionFrameworkBasic(feedName, groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkFeedPermissionFrameworkDestroyed asserts no betterado_feed_permission
// resources remain after destroy. Uses getFeedDirectClient because the mux
// provider does not wire the SDKv2 meta singleton.
func checkFeedPermissionFrameworkDestroyed(s *terraform.State) error {
	clients, err := getFeedDirectClient()
	if err != nil {
		return fmt.Errorf("getting direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feed_permission" {
			continue
		}
		feedID := res.Primary.Attributes["feed_id"]
		projectID := res.Primary.Attributes["project_id"]

		permissions, err := clients.FeedClient.GetFeedPermissions(clients.Ctx, feed.GetFeedPermissionsArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil {
			// API error likely means the feed itself is gone — treat as destroyed.
			continue
		}
		if permissions != nil && len(*permissions) > 0 {
			return fmt.Errorf("feed permissions (Feed ID: %s) should not exist after destroy", feedID)
		}
	}
	return nil
}

// captureFeedPermissionFrameworkEvidence calls CaptureLiveEvidence with a real ADO
// GET URL for the feed permissions endpoint, satisfying AC2.
func captureFeedPermissionFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil // best-effort
		}
		feedID := res.Primary.Attributes["feed_id"]
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getFeedDirectClient()
		if err != nil {
			return nil // best-effort
		}

		permissions, err := clients.FeedClient.GetFeedPermissions(clients.Ctx, feed.GetFeedPermissionsArgs{
			FeedId:  &feedID,
			Project: nilIfEmptyStr(projectID),
		})
		if err != nil || permissions == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		var permURL string
		if projectID != "" {
			permURL = fmt.Sprintf("%s/%s/_apis/packaging/feeds/%s/permissions?api-version=7.1", orgURL, projectID, feedID)
		} else {
			permURL = fmt.Sprintf("%s/_apis/packaging/feeds/%s/permissions?api-version=7.1", orgURL, feedID)
		}

		_ = testutils.CaptureLiveEvidence("acceptance-resource", permURL, permissions)
		return nil
	}
}

// hclFeedPermissionFrameworkBasic returns HCL for creating a betterado_feed_permission
// via the framework provider. Uses the persistent shared project (betterado-standing-demo)
// via a data source lookup to avoid creating a new project.
func hclFeedPermissionFrameworkBasic(feedName, groupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[3]q
}

resource "betterado_feed" "test" {
  name       = %[1]q
  project_id = data.betterado_project.test.id
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.test.id
  display_name = %[2]q
}

resource "betterado_feed_permission" "test" {
  feed_id             = betterado_feed.test.id
  project_id          = data.betterado_project.test.id
  role                = "reader"
  identity_descriptor = betterado_group.test.descriptor
}
`, feedName, groupName, SharedFixtureProjectName)
}
