package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// getWorkItemTypeDirectClient builds an AggregatedClient directly from env vars.
// Used by CheckDestroy and evidence helpers because ProtoV6ProviderFactories
// does not wire the SDKv2 provider singleton's Meta.
func getWorkItemTypeDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// checkWorkItemTypeDestroyed verifies that all processes and work item types
// referenced in the state are destroyed.
func checkWorkItemTypeDestroyed(s *terraform.State) error {
	clients, err := getWorkItemTypeDirectClient()
	if err != nil {
		return fmt.Errorf("checkWorkItemTypeDestroyed: failed to build ADO client: %v", err)
	}

	timeout := 10 * time.Second

	for _, res := range s.RootModule().Resources {
		switch res.Type {
		case "betterado_workitemtrackingprocess_process":
			processID, err2 := uuid.Parse(res.Primary.ID)
			if err2 != nil {
				return err2
			}
			err3 := retry.RetryContext(clients.Ctx, timeout, func() *retry.RetryError {
				_, err4 := clients.CoreClient.GetProcessById(clients.Ctx, core.GetProcessByIdArgs{
					ProcessId: &processID,
				})
				if err4 == nil {
					return retry.RetryableError(fmt.Errorf("process %s should not exist", processID))
				}
				if utils.ResponseWasNotFound(err4) {
					return nil
				}
				return retry.NonRetryableError(err4)
			})
			if err3 != nil {
				return err3
			}

		case "betterado_workitemtrackingprocess_workitemtype":
			referenceName := res.Primary.ID
			processID := res.Primary.Attributes["process_id"]

			_, err2 := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(clients.Ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
				ProcessId:  converter.UUID(processID),
				WitRefName: &referenceName,
				Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.None,
			})
			if err2 != nil {
				if utils.ResponseWasNotFound(err2) {
					continue
				}
				return fmt.Errorf("error reading work item type %s after destroy: %v", referenceName, err2)
			}
			return fmt.Errorf("work item type %s still exists after destroy", referenceName)
		}
	}

	return nil
}

// captureWorkItemTypeEvidence performs a live API GET of the created work item type
// and persists the response as forge demo live-evidence (before destroy).
// Best-effort: a capture failure never fails the test.
func captureWorkItemTypeEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		referenceName := res.Primary.ID
		processID := res.Primary.Attributes["process_id"]

		clients, err := getWorkItemTypeDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		wit, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(clients.Ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
			ProcessId:  converter.UUID(processID),
			WitRefName: &referenceName,
			Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
		})
		if err != nil || wit == nil {
			return nil //nolint:nilerr // best-effort
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/work/processes/%s/workitemtypes/%s?api-version=7.1", orgURL, processID, referenceName)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-workitemtype", url, wit)
		return nil
	}
}

func TestAccWorkitemtrackingprocessWorkItemType_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_workitemtype.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemTypeDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicWorkItemType(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", workItemTypeName),
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttrSet(tfNode, "color"),
					resource.TestCheckResourceAttrSet(tfNode, "icon"),
					resource.TestCheckResourceAttrSet(tfNode, "reference_name"),
					captureWorkItemTypeEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getWorkItemTypeStateIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessWorkItemType_CreateAndUpdate(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()

	tfNode := "betterado_workitemtrackingprocess_workitemtype.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemTypeDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicWorkItemType(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", workItemTypeName),
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttrSet(tfNode, "color"),
					resource.TestCheckResourceAttrSet(tfNode, "icon"),
					resource.TestCheckResourceAttrSet(tfNode, "reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "pages.#"),
					captureWorkItemTypeEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getWorkItemTypeStateIdFunc(tfNode),
			},
			{
				Config: disabledWorkItemType(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", workItemTypeName),
					resource.TestCheckResourceAttrSet(tfNode, "process_id"),
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					resource.TestCheckResourceAttrSet(tfNode, "color"),
					resource.TestCheckResourceAttrSet(tfNode, "icon"),
					resource.TestCheckResourceAttrSet(tfNode, "reference_name"),
					resource.TestCheckResourceAttrSet(tfNode, "pages.#"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getWorkItemTypeStateIdFunc(tfNode),
			},
		},
	})
}

func basicWorkItemType(name string, processName string) string {
	proc := process(processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
}
`, proc, name)
}

func disabledWorkItemType(name string, processName string) string {
	proc := process(processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name       = "%s"
  process_id = betterado_workitemtrackingprocess_process.test.id
  is_enabled = false
}
`, proc, name)
}

func getWorkItemTypeStateIdFunc(tfNode string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		res := state.RootModule().Resources[tfNode]
		id := res.Primary.Attributes["id"]
		processId := res.Primary.Attributes["process_id"]
		return fmt.Sprintf("%s/%s", processId, id), nil
	}
}
