package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// getInheritedControlDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used because ProtoV6ProviderFactories does not configure the SDKv2 provider singleton,
// so testutils.GetProvider().Meta() would be nil when using GetMuxedProviderFactories().
func getInheritedControlDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureInheritedControlEvidence performs a real live API GET of the inherited
// control's work item type layout and persists it as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func captureInheritedControlEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		processId := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_reference_name"]

		clients, err := getInheritedControlDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		expand := workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout
		workItemType, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(context.Background(),
			workitemtrackingprocess.GetProcessWorkItemTypeArgs{
				ProcessId:  converter.UUID(processId),
				WitRefName: &witRefName,
				Expand:     &expand,
			})
		if err != nil || workItemType == nil {
			return nil //nolint:nilerr // best-effort
		}

		url := fmt.Sprintf("https://dev.azure.com/davidgparsonson/_apis/work/processdefinitions/%s/workItemTypes/%s/layout?api-version=7.1", processId, witRefName)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-inherited-control", url, workItemType.Layout)
		return nil
	}
}

func TestAccWorkitemtrackingprocessInheritedControl_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_inherited_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicInheritedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureInheritedControlEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getInheritedControlImportIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessInheritedControl_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_inherited_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicInheritedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getInheritedControlImportIdFunc(tfNode),
			},
			{
				Config: updatedInheritedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getInheritedControlImportIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessInheritedControl_Revert(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_inherited_control.test"

	var processId, witRefName, groupId, controlId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicInheritedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrWith(tfNode, "process_id", func(value string) error {
						processId = value
						return nil
					}),
					resource.TestCheckResourceAttrWith(tfNode, "work_item_type_id", func(value string) error {
						witRefName = value
						return nil
					}),
					resource.TestCheckResourceAttrWith(tfNode, "group_id", func(value string) error {
						groupId = value
						return nil
					}),
					resource.TestCheckResourceAttrWith(tfNode, "id", func(value string) error {
						controlId = value
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getInheritedControlImportIdFunc(tfNode),
			},
			{
				Config: inheritedControlRevertConfig(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					checkInheritedControlRevertedFunc(&processId, &witRefName, &groupId, &controlId),
				),
			},
		},
	})
}

func basicInheritedControl(workItemTypeName string, processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}

resource "betterado_workitemtrackingprocess_inherited_control" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id          = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].groups[0].id
  control_id        = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].groups[0].controls[0].id
  visible           = false
}
`, processName, agileSystemProcessTypeId, workItemTypeName)
}

func updatedInheritedControl(workItemTypeName string, processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}

resource "betterado_workitemtrackingprocess_inherited_control" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id          = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].groups[0].id
  control_id        = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].groups[0].controls[0].id
  visible           = true
  label             = "Custom Label"
}
`, processName, agileSystemProcessTypeId, workItemTypeName)
}

func inheritedControlRevertConfig(workItemTypeName string, processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}
`, processName, agileSystemProcessTypeId, workItemTypeName)
}

func getInheritedControlImportIdFunc(tfNode string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		res := state.RootModule().Resources[tfNode]
		id := res.Primary.Attributes["id"]
		processId := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_id"]
		groupId := res.Primary.Attributes["group_id"]
		return fmt.Sprintf("%s/%s/%s/%s", processId, witRefName, groupId, id), nil
	}
}

func findGroupById(layout *workitemtrackingprocess.FormLayout, groupId string) *workitemtrackingprocess.Group {
	if layout == nil || layout.Pages == nil {
		return nil
	}
	for _, page := range *layout.Pages {
		if page.Sections == nil {
			continue
		}
		for _, section := range *page.Sections {
			if section.Groups == nil {
				continue
			}
			for _, group := range *section.Groups {
				if group.Id != nil && *group.Id == groupId {
					return &group
				}
			}
		}
	}
	return nil
}

func findControlInGroup(group *workitemtrackingprocess.Group, controlId string) *workitemtrackingprocess.Control {
	if group == nil || group.Controls == nil {
		return nil
	}
	for _, control := range *group.Controls {
		if control.Id != nil && *control.Id == controlId {
			return &control
		}
	}
	return nil
}

func checkInheritedControlRevertedFunc(processId, witRefName, groupId, controlId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getInheritedControlDirectClient()
		if err != nil {
			return fmt.Errorf("building client: %w", err)
		}

		// Get the work item type layout to verify the control still exists and is no longer overridden
		args := workitemtrackingprocess.GetProcessWorkItemTypeArgs{
			ProcessId:  converter.UUID(*processId),
			WitRefName: witRefName,
			Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
		}
		workItemType, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(context.Background(), args)
		if err != nil {
			return fmt.Errorf("error getting work item type: %+v", err)
		}

		if workItemType == nil || workItemType.Layout == nil {
			return fmt.Errorf("work item type or layout is nil")
		}

		// Find the group - it must still exist
		group := findGroupById(workItemType.Layout, *groupId)
		if group == nil {
			return fmt.Errorf("group %s was removed, but inherited groups should not be removed", *groupId)
		}

		// Find the control - it must still exist (revert should not remove the control)
		control := findControlInGroup(group, *controlId)
		if control == nil {
			return fmt.Errorf("control %s was removed, but inherited controls should be reverted not removed", *controlId)
		}

		// The control should be marked as inherited and not overridden
		if control.Inherited == nil || !*control.Inherited {
			return fmt.Errorf("control %s should be marked as inherited after revert", *controlId)
		}

		if control.Overridden != nil && *control.Overridden {
			return fmt.Errorf("control %s should not be overridden after revert", *controlId)
		}

		return nil
	}
}
