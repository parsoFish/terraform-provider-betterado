package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// getGroupDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used because ProtoV6ProviderFactories does not configure the SDKv2 provider singleton,
// so testutils.GetProvider().Meta() would be nil when using GetMuxedProviderFactories().
func getGroupDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureGroupEvidence performs a real live API GET of the created group and persists
// the response as forge demo live-evidence (before the resource is destroyed).
// Best-effort: a capture failure never fails the test.
func captureGroupEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		processId := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_reference_name"]
		groupId := res.Primary.ID

		clients, err := getGroupDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort: client build failure does not fail the test
		}

		expand := workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout
		workItemType, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(context.Background(),
			workitemtrackingprocess.GetProcessWorkItemTypeArgs{
				ProcessId:  converter.UUID(processId),
				WitRefName: &witRefName,
				Expand:     &expand,
			})
		if err != nil || workItemType == nil || workItemType.Layout == nil {
			return nil //nolint:nilerr // best-effort: API failure does not fail the test
		}

		// Find the group in the layout
		var foundGroup *workitemtrackingprocess.Group
		if workItemType.Layout.Pages != nil {
			for _, page := range *workItemType.Layout.Pages {
				if page.Sections == nil {
					continue
				}
				for _, section := range *page.Sections {
					if section.Groups == nil {
						continue
					}
					for _, g := range *section.Groups {
						if g.Id != nil && *g.Id == groupId {
							gCopy := g
							foundGroup = &gCopy
							break
						}
					}
					if foundGroup != nil {
						break
					}
				}
				if foundGroup != nil {
					break
				}
			}
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/work/processdefinitions/%s/workItemTypes/%s/layout?api-version=7.1", orgURL, processId, witRefName)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-group", url, foundGroup)
		return nil
	}
}

func TestAccWorkitemtrackingprocessGroup_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicGroup(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Group"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					captureGroupEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessGroup_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicGroup(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Group"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: updatedGroup(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Updated Group"),
					resource.TestCheckResourceAttr(tfNode, "visible", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessGroup_Move(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_group.test"

	var originalSectionId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicGroup(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Group"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					resource.TestCheckResourceAttrWith(tfNode, "section_id", func(value string) error {
						originalSectionId = value
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: movedGroup(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Group"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					resource.TestCheckResourceAttrWith(tfNode, "section_id", func(value string) error {
						if value == originalSectionId {
							return fmt.Errorf("section_id should have changed, but is still %s", originalSectionId)
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// NOTE: This test might be flakey during refresh and/or import due to eventual consistent read
func TestAccWorkitemtrackingprocessGroup_WithMultipleControlTypes(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config:       groupWithMultipleControlTypes(workItemTypeName, processName),
				RefreshState: false,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "page_id"),
					resource.TestCheckResourceAttrSet(tfNode, "section_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "All Control Types Group"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "control.#", "6"),
					// HtmlFieldControl - for rich text HTML fields
					resource.TestCheckResourceAttr(tfNode, "control.0.id", "System.Description"),
					resource.TestCheckResourceAttr(tfNode, "control.0.label", "Description"),
					resource.TestCheckResourceAttr(tfNode, "control.0.control_type", "HtmlFieldControl"),
					resource.TestCheckResourceAttr(tfNode, "control.0.visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.0.read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.0.order", "0"),
					resource.TestCheckResourceAttr(tfNode, "control.0.inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.0.overridden", "false"),
					// FieldControl - for plain text fields
					resource.TestCheckResourceAttr(tfNode, "control.1.id", "System.Title"),
					resource.TestCheckResourceAttr(tfNode, "control.1.label", "Title"),
					resource.TestCheckResourceAttr(tfNode, "control.1.control_type", "FieldControl"),
					resource.TestCheckResourceAttr(tfNode, "control.1.visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.1.read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.1.order", "1"),
					resource.TestCheckResourceAttr(tfNode, "control.1.inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.1.overridden", "false"),
					// FieldControl - for identity fields
					resource.TestCheckResourceAttr(tfNode, "control.2.id", "System.AssignedTo"),
					resource.TestCheckResourceAttr(tfNode, "control.2.label", "Assigned To"),
					resource.TestCheckResourceAttr(tfNode, "control.2.control_type", "FieldControl"),
					resource.TestCheckResourceAttr(tfNode, "control.2.visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.2.read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.2.order", "2"),
					resource.TestCheckResourceAttr(tfNode, "control.2.inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.2.overridden", "false"),
					// DateTimeControl - for date/time fields
					resource.TestCheckResourceAttr(tfNode, "control.3.id", "System.CreatedDate"),
					resource.TestCheckResourceAttr(tfNode, "control.3.label", "Created Date"),
					resource.TestCheckResourceAttr(tfNode, "control.3.control_type", "DateTimeControl"),
					resource.TestCheckResourceAttr(tfNode, "control.3.visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.3.read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.3.order", "3"),
					resource.TestCheckResourceAttr(tfNode, "control.3.inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.3.overridden", "false"),
					// WorkItemClassificationControl - for tree path fields
					resource.TestCheckResourceAttr(tfNode, "control.4.id", "System.AreaPath"),
					resource.TestCheckResourceAttr(tfNode, "control.4.label", "Area Path"),
					resource.TestCheckResourceAttr(tfNode, "control.4.control_type", "WorkItemClassificationControl"),
					resource.TestCheckResourceAttr(tfNode, "control.4.visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.4.read_only", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.4.order", "4"),
					resource.TestCheckResourceAttr(tfNode, "control.4.inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "control.4.overridden", "false"),
					// Contribution control (extension)
					resource.TestCheckResourceAttr(tfNode, "control.5.id", "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"),
					resource.TestCheckResourceAttr(tfNode, "control.5.order", "5"),
					resource.TestCheckResourceAttr(tfNode, "control.5.is_contribution", "true"),
					resource.TestCheckResourceAttr(tfNode, "control.5.contribution.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "control.5.contribution.0.contribution_id", "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"),
					resource.TestCheckResourceAttr(tfNode, "control.5.contribution.0.height", "50"),
					resource.TestCheckResourceAttr(tfNode, "control.5.contribution.0.inputs.FieldName", "System.Tags"),
					resource.TestCheckResourceAttr(tfNode, "control.5.contribution.0.inputs.Values", "Option1;Option2;Option3"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: groupImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func basicGroup(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_group" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].id
  label                         = "Test Group"
}
`, workItemType)
}

func updatedGroup(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_group" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].id
  label                         = "Updated Group"
  visible                       = false
}
`, workItemType)
}

func movedGroup(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_group" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[1].id
  label                         = "Test Group"
}
`, workItemType)
}

func groupWithMultipleControlTypes(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_extension" "test" {
  publisher_id = "ms-devlabs"
  extension_id = "vsts-extensions-multivalue-control"
}

resource "betterado_workitemtrackingprocess_group" "test" {
  depends_on                    = [betterado_extension.test]
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].id
  label                         = "All Control Types Group"

  # HtmlFieldControl - for rich text HTML fields
  control {
    id    = "System.Description"
    label = "Description"
  }

  # FieldControl - for plain text fields
  control {
    id    = "System.Title"
    label = "Title"
  }

  # FieldControl - for identity fields
  control {
    id    = "System.AssignedTo"
    label = "Assigned To"
  }

  # DateTimeControl - for date/time fields (System.CreatedDate is a system field available in all work item types)
  control {
    id    = "System.CreatedDate"
    label = "Created Date"
  }

  # WorkItemClassificationControl - for tree path fields
  control {
    id    = "System.AreaPath"
    label = "Area Path"
  }

  # Contribution control (extension)
  control {
    id              = "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"
    is_contribution = true
    contribution {
      contribution_id = "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"
      height          = 50
      inputs = {
        FieldName = "System.Tags"
        Values    = "Option1;Option2;Option3"
      }
    }
  }
}
`, workItemType)
}

func groupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		processId := rs.Primary.Attributes["process_id"]
		witRefName := rs.Primary.Attributes["work_item_type_reference_name"]
		pageId := rs.Primary.Attributes["page_id"]
		sectionId := rs.Primary.Attributes["section_id"]
		groupId := rs.Primary.ID
		return fmt.Sprintf("%s/%s/%s/%s/%s", processId, witRefName, pageId, sectionId, groupId), nil
	}
}
