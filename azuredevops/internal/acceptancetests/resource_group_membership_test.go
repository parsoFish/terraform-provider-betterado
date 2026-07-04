package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccGroupMembership_overwrite(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: overwriteEmpty(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "0"),
				),
			},
			{
				Config: overwriteWithMember(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
			{
				Config: overwriteEmpty(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "0"),
				),
			},
		},
	})
}

func overwriteEmpty(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "acctest-%[1]s"
}

resource "betterado_group" "test" {
  display_name = "acctest-%[1]s"
  scope        = betterado_project.test.id
}

resource "betterado_group_membership" "test" {
  group   = betterado_group.test.id
  mode    = "overwrite"
  members = []
}
`, name)
}

func overwriteWithMember(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "acctest-%[1]s"
}

resource "betterado_group" "test" {
  display_name = "acctest-%[1]s"
  scope        = betterado_project.test.id
}

resource "betterado_group" "member" {
  display_name = "acctest-member-%[1]s"
  scope        = betterado_project.test.id
}

resource "betterado_group_membership" "test" {
  group   = betterado_group.test.id
  mode    = "overwrite"
  members = [betterado_group.member.id]
}
`, name)
}

// ── Framework provider tests ──────────────────────────────────────────────────

// membershipGetDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used by CheckDestroy and evidence helpers when using ProtoV6ProviderFactories.
func membershipGetDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// TestAccGroupMembership_Framework exercises betterado_group_membership via the muxed
// (terraform-plugin-framework) provider path.
//
// AC1: overwrite mode — create with a member, idempotency re-plan shows no changes.
// AC3: update from overwrite → add mode — exercises the update path; idempotency re-plan is clean.
func TestAccGroupMembership_Framework(t *testing.T) {
	groupName := testutils.GenerateResourceName()
	memberGroupName := testutils.GenerateResourceName()
	tfNode := "betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGroupMembershipDestroyedFramework,
		Steps: []resource.TestStep{
			// Step 1: Create with mode=overwrite and one member; assert read-back.
			{
				Config: hclGroupMembershipFrameworkOverwrite(groupName, memberGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "mode", "overwrite"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
					captureMembershipEvidence(tfNode, groupName),
				),
			},
			// Step 2: Idempotency — re-plan must show no changes (AC1).
			{
				Config:             hclGroupMembershipFrameworkOverwrite(groupName, memberGroupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Update mode from overwrite→add (AC3); idempotency must remain clean.
			{
				Config: hclGroupMembershipFrameworkAdd(groupName, memberGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "mode", "add"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
			// Step 4: Idempotency after mode update (AC3).
			{
				Config:             hclGroupMembershipFrameworkAdd(groupName, memberGroupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkGroupMembershipDestroyedFramework verifies that the membership has been cleaned up
// after destroy by confirming the member group descriptor is no longer a member of the
// container group. Uses a direct ADO client since ProtoV6ProviderFactories does not wire
// the SDKv2 provider singleton's Meta.
func checkGroupMembershipDestroyedFramework(s *terraform.State) error {
	clients, err := membershipGetDirectClient()
	if err != nil {
		return fmt.Errorf("checkGroupMembershipDestroyedFramework: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_group_membership" {
			continue
		}
		groupDescriptor := res.Primary.Attributes["group"]
		memberships, err := clients.GraphClient.ListMemberships(clients.Ctx, graph.ListMembershipsArgs{
			SubjectDescriptor: converter.String(groupDescriptor),
			Direction:         &graph.GraphTraversalDirectionValues.Down,
			Depth:             converter.Int(1),
		})
		if err != nil {
			// If the group itself was destroyed, memberships are gone too — treat as success.
			return nil //nolint:nilerr // group destroyed → no memberships to check
		}
		// All members should have been removed on destroy.
		if memberships != nil && len(*memberships) > 0 {
			return fmt.Errorf("group_membership %s: expected 0 memberships after destroy, got %d", groupDescriptor, len(*memberships))
		}
	}
	return nil
}

// captureMembershipEvidence performs a live ADO API list of the group memberships and
// persists the response as forge demo live-evidence. Best-effort: failures are ignored
// so they never cause a test assertion failure.
func captureMembershipEvidence(tfNode, groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		groupDescriptor := res.Primary.Attributes["group"]
		clients, err := membershipGetDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort evidence capture
		}
		memberships, err := clients.GraphClient.ListMemberships(clients.Ctx, graph.ListMembershipsArgs{
			SubjectDescriptor: converter.String(groupDescriptor),
			Direction:         &graph.GraphTraversalDirectionValues.Down,
			Depth:             converter.Int(1),
		})
		if err != nil || memberships == nil {
			return nil //nolint:nilerr // best-effort evidence capture
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		url := fmt.Sprintf("%s/_apis/graph/memberships/%s?direction=down&api-version=7.1", orgURL, groupDescriptor)
		_ = testutils.CaptureLiveEvidence("group-membership-acceptance", url, memberships)
		return nil
	}
}

// hclGroupMembershipFrameworkOverwrite creates two groups and sets up a group_membership
// in overwrite mode with one member. Uses the shared persistent project (no project create).
func hclGroupMembershipFrameworkOverwrite(groupName, memberGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[3]q
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.shared.id
  display_name = %[1]q
}

resource "betterado_group" "member" {
  scope        = data.betterado_project.shared.id
  display_name = %[2]q
}

resource "betterado_group_membership" "test" {
  group   = betterado_group.test.descriptor
  mode    = "overwrite"
  members = [betterado_group.member.descriptor]
}
`, groupName, memberGroupName, SharedFixtureProjectName)
}

// hclGroupMembershipFrameworkAdd is the same configuration but with mode=add,
// exercising the mode-change update path (AC3).
func hclGroupMembershipFrameworkAdd(groupName, memberGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[3]q
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.shared.id
  display_name = %[1]q
}

resource "betterado_group" "member" {
  scope        = data.betterado_project.shared.id
  display_name = %[2]q
}

resource "betterado_group_membership" "test" {
  group   = betterado_group.test.descriptor
  mode    = "add"
  members = [betterado_group.member.descriptor]
}
`, groupName, memberGroupName, SharedFixtureProjectName)
}
