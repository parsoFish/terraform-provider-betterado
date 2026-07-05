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

// getControlDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used because ProtoV6ProviderFactories does not configure the SDKv2 provider singleton.
func getControlDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureControlEvidence performs a real live API GET of the created control
// and persists the response as forge demo live-evidence (before resource destroy).
// Best-effort: a capture failure never fails the test.
func captureControlEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		processId := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_reference_name"]

		clients, err := getControlDirectClient()
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
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-control", url, workItemType.Layout)
		return nil
	}
}

func TestAccWorkitemtrackingprocessControl_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Control"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "read_only", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					captureControlEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessControl_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Control"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttr(tfNode, "read_only", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: updatedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Updated Control"),
					resource.TestCheckResourceAttr(tfNode, "visible", "false"),
					resource.TestCheckResourceAttr(tfNode, "read_only", "true"),
					resource.TestCheckResourceAttr(tfNode, "order", "0"),
					resource.TestCheckResourceAttr(tfNode, "metadata", "test metadata"),
					resource.TestCheckResourceAttr(tfNode, "watermark", "Enter a title"),
					resource.TestCheckResourceAttr(tfNode, "control_type", "FieldControl"),
					resource.TestCheckResourceAttr(tfNode, "inherited", "false"),
					resource.TestCheckResourceAttr(tfNode, "overridden", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessControl_Move(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_control.test"

	var originalGroupId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Control"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					resource.TestCheckResourceAttrWith(tfNode, "group_id", func(value string) error {
						originalGroupId = value
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: movedControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "label", "Test Control"),
					resource.TestCheckResourceAttr(tfNode, "visible", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					resource.TestCheckResourceAttrWith(tfNode, "group_id", func(value string) error {
						if value == originalGroupId {
							return fmt.Errorf("group_id should have changed, but is still %s", originalGroupId)
						}
						return nil
					}),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessControl_Contribution(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_control.test"

	const multivaluePublisher = "ms-devlabs"
	const multivalueExtension = "vsts-extensions-multivalue-control"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
			// Ensure the multivalue-control extension is installed before the test
			// runs. Managing it as a Terraform resource causes TF1590010 flakiness
			// when a previous test run left it installed. We install it directly via
			// the API and clean up in CheckDestroy instead.
			testutils.EnsureExtensionInstalled(t, multivaluePublisher, multivalueExtension)
		},
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			testutils.EnsureExtensionUninstalled(t, multivaluePublisher, multivalueExtension)
			return testutils.CheckProcessDestroyed(s)
		},
		Steps: []resource.TestStep{
			{
				Config: contributionControl(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttrSet(tfNode, "work_item_type_reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "is_contribution", "true"),
					resource.TestCheckResourceAttr(tfNode, "contribution.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "contribution.0.contribution_id", "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"),
					resource.TestCheckResourceAttr(tfNode, "contribution.0.height", "50"),
					resource.TestCheckResourceAttr(tfNode, "contribution.0.inputs.FieldName", "System.Tags"),
					resource.TestCheckResourceAttr(tfNode, "contribution.0.inputs.Values", "Option1;Option2;Option3"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: controlImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func basicControl(workItemTypeName string, processName string) string {
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

resource "betterado_workitemtrackingprocess_control" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id                      = betterado_workitemtrackingprocess_group.test.id
  control_id                    = "System.Title"
  label                         = "Test Control"
}
`, workItemType)
}

func updatedControl(workItemTypeName string, processName string) string {
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

resource "betterado_workitemtrackingprocess_control" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  control_id                    = "System.Title"
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id                      = betterado_workitemtrackingprocess_group.test.id
  label                         = "Updated Control"
  visible                       = false
  read_only                     = true
  order                         = 0
  metadata                      = "test metadata"
  watermark                     = "Enter a title"
}
`, workItemType)
}

func movedControl(workItemTypeName string, processName string) string {
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

resource "betterado_workitemtrackingprocess_group" "test2" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].id
  label                         = "Test Group 2"
}

resource "betterado_workitemtrackingprocess_control" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id                      = betterado_workitemtrackingprocess_group.test2.id
  control_id                    = "System.Title"
  label                         = "Test Control"
}
`, workItemType)
}

func contributionControl(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

# Note: the ms-devlabs.vsts-extensions-multivalue-control extension is installed
# by the test PreCheck via EnsureExtensionInstalled (not managed as a Terraform
# resource) to avoid TF1590010 "already installed" failures in parallel/retry runs.

resource "betterado_workitemtrackingprocess_group" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  page_id                       = betterado_workitemtrackingprocess_workitemtype.test.pages[0].id
  section_id                    = betterado_workitemtrackingprocess_workitemtype.test.pages[0].sections[0].id
  label                         = "Test Group"
}

resource "betterado_workitemtrackingprocess_control" "test" {
  process_id                    = betterado_workitemtrackingprocess_process.test.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  group_id                      = betterado_workitemtrackingprocess_group.test.id
  control_id                    = "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"
  is_contribution               = true
  contribution {
    contribution_id = "ms-devlabs.vsts-extensions-multivalue-control.multivalue-form-control"
    height          = 50
    inputs = {
      FieldName = "System.Tags"
      Values    = "Option1;Option2;Option3"
    }
  }
}
`, workItemType)
}

func controlImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		processId := rs.Primary.Attributes["process_id"]
		witRefName := rs.Primary.Attributes["work_item_type_reference_name"]
		groupId := rs.Primary.Attributes["group_id"]
		controlId := rs.Primary.ID
		return fmt.Sprintf("%s/%s/%s/%s", processId, witRefName, groupId, controlId), nil
	}
}
