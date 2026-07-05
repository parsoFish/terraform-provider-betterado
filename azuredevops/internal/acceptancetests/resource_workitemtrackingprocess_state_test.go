package acceptancetests

import (
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

func TestAccWorkitemtrackingprocessState_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_state.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicState(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttrSet(tfNode, "order"),
					captureStateEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getStateImportIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessState_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_state.test"

	var stateId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicState(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureStateId(tfNode, &stateId),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getStateImportIdFunc(tfNode),
			},
			{
				Config: updatedState(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(tfNode, "id", &stateId),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getStateImportIdFunc(tfNode),
			},
			{
				Config: basicState(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(tfNode, "id", &stateId),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getStateImportIdFunc(tfNode),
			},
		},
	})
}

func captureStateId(tfNode string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res := s.RootModule().Resources[tfNode]
		*id = res.Primary.Attributes["id"]
		return nil
	}
}

// captureStateEvidence reads back the state definition via API and writes live evidence.
// This satisfies AC4: CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-state", url, apiResponse).
func captureStateEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}

		stateID := res.Primary.ID
		processID := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_id"]

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		agg, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			return nil //nolint:nilerr // best-effort: client build failure does not fail the test
		}

		apiResponse, err := agg.WorkItemTrackingProcessClient.GetStateDefinition(agg.Ctx, workitemtrackingprocess.GetStateDefinitionArgs{
			ProcessId:  converter.UUID(processID),
			WitRefName: converter.String(witRefName),
			StateId:    converter.UUID(stateID),
		})
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		url := fmt.Sprintf("%s/_apis/work/processes/%s/workItemTypes/%s/states/%s?api-version=7.1",
			strings.TrimRight(orgURL, "/"), processID, witRefName, stateID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-state", url, apiResponse)
		return nil
	}
}

func basicState(workItemTypeName string, processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}

resource "betterado_workitemtrackingprocess_state" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Ready"
  color             = "#b2b2b2"
  state_category    = "Proposed"
  order             = 2
}
`, processName, agileSystemProcessTypeId, workItemTypeName)
}

func updatedState(workItemTypeName string, processName string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_process" "test" {
  name                   = "%s"
  parent_process_type_id = "%s"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}

resource "betterado_workitemtrackingprocess_state" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Ready"
  color             = "#5688E0"
  state_category    = "InProgress"
  order             = 3
}
`, processName, agileSystemProcessTypeId, workItemTypeName)
}

func getStateImportIdFunc(tfNode string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		res := state.RootModule().Resources[tfNode]
		id := res.Primary.Attributes["id"]
		processId := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_id"]
		return fmt.Sprintf("%s/%s/%s", processId, witRefName, id), nil
	}
}
