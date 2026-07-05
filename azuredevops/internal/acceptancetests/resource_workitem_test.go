//go:build (all || resource_workitem) && !exclude_resource_workitem

package acceptancetests

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// getWIDirectClient builds an AggregatedClient directly from AZDO env vars for use
// in CheckDestroy and evidence helpers (ProtoV6ProviderFactories does not expose Meta).
func getWIDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// checkWorkItemDestroyed verifies all betterado_workitem resources in state have
// been deleted from ADO after a destroy step.
func checkWorkItemDestroyed(s *terraform.State) error {
	clients, err := getWIDirectClient()
	if err != nil {
		return fmt.Errorf("checkWorkItemDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_workitem" {
			continue
		}
		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid work item ID %q: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		_, err = clients.WorkItemTrackingClient.GetWorkItem(clients.Ctx, workitemtracking.GetWorkItemArgs{
			Project: &projectID,
			Id:      &id,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				continue // expected: resource gone
			}
			return fmt.Errorf("error reading work item %d after destroy: %v", id, err)
		}
		return fmt.Errorf("work item %d still exists after destroy", id)
	}
	return nil
}

// captureWorkItemEvidence performs a real live API GET of the created work item
// and persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test.
func captureWorkItemEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		idStr := res.Primary.ID
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getWIDirectClient()
		if err != nil {
			return nil // best-effort
		}
		workItem, err := clients.WorkItemTrackingClient.GetWorkItem(clients.Ctx, workitemtracking.GetWorkItemArgs{
			Project: &projectID,
			Id:      &id,
		})
		if err != nil || workItem == nil {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		url := fmt.Sprintf("%s/%s/_apis/wit/workItems/%d?api-version=7.1", orgURL, projectID, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitem", url, workItem)
		return nil
	}
}

// TestAccWorkItem_basic creates a work item in the shared fixture project,
// asserts a full read-back, captures live evidence, verifies idempotency, then destroys.
// Uses the muxed provider so the framework resource is reachable.
func TestAccWorkItem_basic(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemBasicShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
					captureWorkItemEvidence(tfNode),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             workItemBasicShared(workItemTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccWorkItem_titleUpdate(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	workItemTitleUpdated := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemBasicShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
			{
				Config: workItemBasicShared(workItemTitleUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitleUpdated),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
		},
	})
}

func TestAccWorkItem_tagUpdate(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemBasicShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
			{
				Config: workItemTagUpdateShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
			{
				Config: workItemBasicShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
		},
	})
}

func TestAccWorkItem_parent(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemParentShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
		},
	})
}

func TestAccWorkItem_parentUpdate(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemParentShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
			{
				Config: workItemParentUpdateShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
		},
	})
}

func TestAccWorkItem_parentDelete(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemParentShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
			{
				Config: workItemParentDeleteShared(workItemTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "Issue"),
					resource.TestCheckResourceAttr(tfNode, "state", "Active"),
				),
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportStateVerifyIgnore: []string{"parent_id"},
			},
		},
	})
}

func TestAccWorkItem_additionalFieldsJson(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"
	storyPoints := 5.00
	risk := "3 - Low"
	acceptanceCriteria := testutils.GenerateResourceName()
	acceptanceCriteriaUpdate := testutils.GenerateResourceName()
	jsonString := fmt.Sprintf("{\"Microsoft.VSTS.Scheduling.StoryPoints\": %f,\"Microsoft.VSTS.Common.AcceptanceCriteria\" =\"%s\"}", storyPoints, acceptanceCriteria)
	jsonStringUpdateAddRemove := fmt.Sprintf("{\"Microsoft.VSTS.Common.Risk\": \"%s\",\"Microsoft.VSTS.Common.AcceptanceCriteria\" =\"%s\"}", risk, acceptanceCriteriaUpdate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemAdditionalFieldsShared(workItemTitle, jsonString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "User Story"),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
					resource.TestCheckResourceAttrWith(tfNode, "additional_fields_json", func(value string) error {
						var m map[string]interface{}
						if err := json.Unmarshal([]byte(value), &m); err != nil {
							return err
						}
						if m["Microsoft.VSTS.Scheduling.StoryPoints"] != storyPoints {
							return fmt.Errorf("expected Microsoft.VSTS.Scheduling.StoryPoints %f got %f", storyPoints, m["Microsoft.VSTS.Scheduling.StoryPoints"])
						}
						if m["Microsoft.VSTS.Common.AcceptanceCriteria"] != acceptanceCriteria {
							return fmt.Errorf("expected Microsoft.VSTS.Common.AcceptanceCriteria %s, got %s", acceptanceCriteria, m["Microsoft.VSTS.Common.AcceptanceCriteria"])
						}
						return nil
					}),
				),
			},
			{
				Config: workItemAdditionalFieldsShared(workItemTitle, jsonStringUpdateAddRemove),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", "User Story"),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
					resource.TestCheckResourceAttrWith(tfNode, "additional_fields_json", func(value string) error {
						var m map[string]interface{}
						if err := json.Unmarshal([]byte(value), &m); err != nil {
							return err
						}
						if v, ok := m["Microsoft.VSTS.Scheduling.StoryPoints"]; ok {
							return fmt.Errorf("expected Microsoft.VSTS.Scheduling.StoryPoints field to have been removed, but it was returned by the api, got %f", v)
						}
						if m["Microsoft.VSTS.Common.AcceptanceCriteria"] != acceptanceCriteriaUpdate {
							return fmt.Errorf("expected Microsoft.VSTS.Common.AcceptanceCriteria %s, got %s", acceptanceCriteriaUpdate, m["Microsoft.VSTS.Common.AcceptanceCriteria"])
						}
						if m["Microsoft.VSTS.Common.Risk"] != risk {
							return fmt.Errorf("expected Microsoft.VSTS.Common.Risk %s, got %s", risk, m["Microsoft.VSTS.Common.Risk"])
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportStateVerifyIgnore: []string{
					"additional_fields_json",
				},
			},
		},
	})
}

func TestAccWorkItem_description(t *testing.T) {
	workItemTitle := testutils.GenerateResourceName()
	tfNode := "betterado_workitem.test"
	description := testutils.GenerateResourceName()
	descriptionUpdate := testutils.GenerateResourceName()
	descriptionEmpty := ""
	itemType := "User Story"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemDestroyed,
		Steps: []resource.TestStep{
			{
				Config: workItemDescriptionNoneShared(workItemTitle, itemType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttr(tfNode, "type", itemType),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
				),
			},
			{
				Config: workItemDescriptionShared(workItemTitle, itemType, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", itemType),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
					resource.TestCheckResourceAttr(tfNode, "description", description),
				),
			},
			{
				Config: workItemDescriptionShared(workItemTitle, itemType, descriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "type", itemType),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
					resource.TestCheckResourceAttr(tfNode, "description", descriptionUpdate),
				),
			},
			{
				Config: workItemDescriptionShared(workItemTitle, itemType, descriptionEmpty),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttr(tfNode, "title", workItemTitle),
					resource.TestCheckResourceAttr(tfNode, "type", itemType),
					resource.TestCheckResourceAttr(tfNode, "state", "New"),
					// TODO: Update when Computed is removed from description
					resource.TestCheckResourceAttr(tfNode, "description", descriptionUpdate),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
			},
		},
	})
}

// ── Shared-fixture Terraform config helpers ────────────────────────────────────
// All helpers below use the persistent SharedFixtureProjectName project to avoid
// the 1000-project org cap. No ad-hoc betterado_project resources are created.

// sharedProjectDataSource returns the data source block for the shared fixture project.
func sharedProjectDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %q
}
`, SharedFixtureProjectName)
}

// workItemBasicShared creates a work item in the PERSISTENT shared fixture project
// (SharedFixtureProjectName) to avoid the 1000-project ADO org cap.
func workItemBasicShared(title string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "test" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
}
`, SharedFixtureProjectName, title)
}

func workItemTagUpdateShared(title string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "test" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
  state      = "Active"
  tags       = ["tag1", "tag2"]
}
`, SharedFixtureProjectName, title)
}

func workItemParentShared(title string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "parent" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
}

resource "betterado_workitem" "test" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
  parent_id  = betterado_workitem.parent.id
}
`, SharedFixtureProjectName, title)
}

func workItemParentDeleteShared(title string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "parent" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
}

resource "betterado_workitem" "test" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = "Issue"
}
`, SharedFixtureProjectName, title)
}

func workItemParentUpdateShared(title string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "parent" {
  title      = "%[2]s Parent"
  project_id = data.betterado_project.test.id
  type       = "Issue"
}

resource "betterado_workitem" "parent2" {
  title      = "%[2]s Parent2"
  project_id = data.betterado_project.test.id
  type       = "Issue"
}

resource "betterado_workitem" "test" {
  project_id = data.betterado_project.test.id
  title      = %[2]q
  type       = "Issue"
  parent_id  = betterado_workitem.parent2.id
}
`, SharedFixtureProjectName, title)
}

func workItemDescriptionShared(title string, itemType string, description string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "test" {
  title       = %[2]q
  project_id  = data.betterado_project.test.id
  type        = %[3]q
  description = %[4]q
}
`, SharedFixtureProjectName, title, itemType, description)
}

func workItemDescriptionNoneShared(title string, itemType string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "test" {
  title      = %[2]q
  project_id = data.betterado_project.test.id
  type       = %[3]q
}
`, SharedFixtureProjectName, title, itemType)
}

func workItemAdditionalFieldsShared(title string, jsonString string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitem" "test" {
  title                  = %[2]q
  project_id             = data.betterado_project.test.id
  type                   = "User Story"
  additional_fields_json = jsonencode(%[3]s)
}
`, SharedFixtureProjectName, title, jsonString)
}
