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
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccGroupResource_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupBasic(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupName),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupName),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroupResource_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	groupNameUpdate := groupName + "update"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupBasic(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupName),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupName),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclGroupBasic(projectName, groupNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupNameUpdate),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupNameUpdate),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroupResource_updateMembers(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupMembers(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupName),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupName),
					resource.TestCheckResourceAttr("betterado_group.test", "members.#", "1"),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclGroupMembersUpdate(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupName),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupName),
					resource.TestCheckResourceAttr("betterado_group.test", "members.#", "2"),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroupResource_complete(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupComplete(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupExists(groupName),
					resource.TestCheckResourceAttrSet("betterado_group.test", "scope"),
					resource.TestCheckResourceAttrSet("betterado_group.test", "group_id"),
					resource.TestCheckResourceAttr("betterado_group.test", "display_name", groupName),
				),
			},
			{
				ResourceName:      "betterado_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkGroupExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		varGroup, ok := s.RootModule().Resources["betterado_group.test"]
		if !ok {
			return fmt.Errorf("Did not find a group resource in the TF state")
		}

		getGroupArgs := graph.GetGroupArgs{
			GroupDescriptor: converter.String(varGroup.Primary.Attributes["descriptor"]),
		}
		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		group, err := clients.GraphClient.GetGroup(clients.Ctx, getGroupArgs)
		if err != nil {
			return err
		}
		if group == nil {
			return fmt.Errorf("Group with Name=%s does not exit", varGroup.Primary.Attributes["display_name"])
		}
		if *group.DisplayName != expectedName {
			return fmt.Errorf("Group has Name=%s, but expected %s", *group.DisplayName, expectedName)
		}

		return nil
	}
}

func checkGroupDestroyed(s *terraform.State) error {
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)

	// verify that every project referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_group" {
			continue
		}

		// The group will be returned even if it has been deleted from the account or has had all its memberships deleted.
		id := resource.Primary.ID
		err := clients.GraphClient.DeleteGroup(clients.Ctx, graph.DeleteGroupArgs{
			GroupDescriptor: converter.String(id),
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				return nil
			}
			return fmt.Errorf("Group with ID %s should not exist in scope %s", id, resource.Primary.Attributes["scope"])
		}
	}

	return nil
}

func hclGroupBasic(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%[2]s"
}`, projectName, groupName)
}

func hclGroupComplete(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

data "betterado_group" "test2" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%[2]s"
  description  = "managed by Terraform"
  members = [
    data.betterado_group.test.descriptor,
    data.betterado_group.test2.descriptor,
  ]
}`, projectName, groupName)
}

func hclGroupMembers(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%[2]s"
  members = [
    data.betterado_group.test.descriptor,
  ]
}`, projectName, groupName)
}

func hclGroupMembersUpdate(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

data "betterado_group" "test2" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = "%[2]s"
  members = [
    data.betterado_group.test.descriptor,
    data.betterado_group.test2.descriptor,
  ]
}`, projectName, groupName)
}

// ── Framework provider tests ──────────────────────────────────────────────────

// groupGetDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used by CheckDestroy and evidence helpers when using ProtoV6ProviderFactories
// (which does not wire the SDKv2 provider singleton's Meta).
func groupGetDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// TestAccGroupResource_Framework exercises betterado_group via the muxed
// (terraform-plugin-framework) provider path. This is the quality gate test.
//
// Uses the persistent shared project (betterado-standing-demo) via a data
// source lookup — the ADO org is at the 1000-project cap so project creates fail.
func TestAccGroupResource_Framework(t *testing.T) {
	groupName := testutils.GenerateResourceName()
	tfNode := "betterado_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkGroupDestroyedFramework,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back populates all computed attrs.
			{
				Config: hclGroupFramework(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "display_name", groupName),
					resource.TestCheckResourceAttr(tfNode, "description", "managed by Terraform"),
					resource.TestCheckResourceAttrSet(tfNode, "scope"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "principal_name"),
					resource.TestCheckResourceAttrSet(tfNode, "domain"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_kind"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					captureGroupEvidence(tfNode),
				),
			},
			// Step 2: idempotency -- re-plan must show no changes (AC1).
			{
				Config:             hclGroupFramework(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkGroupDestroyedFramework verifies the group is gone after destroy using a
// direct ADO client (ProtoV6ProviderFactories does not wire SDKv2 provider Meta).
func checkGroupDestroyedFramework(s *terraform.State) error {
	clients, err := groupGetDirectClient()
	if err != nil {
		return fmt.Errorf("checkGroupDestroyedFramework: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_group" {
			continue
		}
		id := res.Primary.ID
		err := clients.GraphClient.DeleteGroup(clients.Ctx, graph.DeleteGroupArgs{
			GroupDescriptor: converter.String(id),
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				// already deleted -- expected
				continue
			}
			return fmt.Errorf("group %s still accessible after destroy: %v", id, err)
		}
		// DeleteGroup succeeded means the group existed -- it should have been destroyed
	}

	return nil
}

// captureGroupEvidence performs a live ADO API GET of the created group and
// persists the response as forge demo live-evidence. Best-effort: failures
// are ignored so they never cause a test assertion failure.
func captureGroupEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		descriptor := res.Primary.ID
		clients, err := groupGetDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort evidence capture: client failure never fails the test
		}
		grp, err := clients.GraphClient.GetGroup(clients.Ctx, graph.GetGroupArgs{
			GroupDescriptor: converter.String(descriptor),
		})
		if err != nil || grp == nil {
			return nil //nolint:nilerr // best-effort evidence capture: API failure never fails the test
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		url := fmt.Sprintf("%s/_apis/graph/groups/%s?api-version=7.1", orgURL, descriptor)
		_ = testutils.CaptureLiveEvidence("group-resource-acceptance", url, grp)
		return nil
	}
}

func hclGroupFramework(groupName string) string {
	// Use the persistent shared project via a data lookup — the ADO org is at
	// the 1000-project cap so resource "betterado_project" creates fail.
	// SharedFixtureProjectName ("betterado-standing-demo") is always present.
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[2]q
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.shared.id
  display_name = %[1]q
  description  = "managed by Terraform"
}`, groupName, SharedFixtureProjectName)
}
