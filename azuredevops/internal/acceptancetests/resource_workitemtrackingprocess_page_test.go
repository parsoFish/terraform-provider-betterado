package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessPage_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_page.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicPage(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					testutils.TestCheckAttrGreaterThan(tfNode, "sections.#", 0),
					capturePageEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: pageImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessPage_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_page.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicPage(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					testutils.TestCheckAttrGreaterThan(tfNode, "sections.#", 0),
				),
			},
			{
				Config: updatedPage(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					testutils.TestCheckAttrGreaterThan(tfNode, "sections.#", 0),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: pageImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func basicPage(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_page" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.id
  label             = "Test Page"
}
`, workItemType)
}

func updatedPage(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_page" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.id
  label             = "Updated Page"
  visible           = false
  order             = 4
}
`, workItemType)
}

func pageImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		processId := rs.Primary.Attributes["process_id"]
		witRefName := rs.Primary.Attributes["work_item_type_id"]
		pageId := rs.Primary.ID
		return fmt.Sprintf("%s/%s/%s", processId, witRefName, pageId), nil
	}
}

// capturePageEvidence performs a real live API GET of the created page and
// persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test.
func capturePageEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		pageID := res.Primary.ID
		processID := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_id"]

		pageUUID, err := uuid.Parse(pageID)
		if err != nil {
			return nil //nolint:nilerr // best-effort: bad ID does not fail the test
		}
		_ = pageUUID

		// Build client directly from env vars (not GetProvider().Meta() — mux path).
		clients, err := testutils.GetDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort: client build failure does not fail the test
		}

		processUUID, err := uuid.Parse(processID)
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		workItemType, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemType(clients.Ctx,
			workitemtrackingprocess.GetProcessWorkItemTypeArgs{
				ProcessId:  &processUUID,
				WitRefName: &witRefName,
				Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
			})
		if err != nil || workItemType == nil || workItemType.Layout == nil || workItemType.Layout.Pages == nil {
			return nil //nolint:nilerr // best-effort
		}

		var apiResponse *workitemtrackingprocess.Page
		for i := range *workItemType.Layout.Pages {
			p := &(*workItemType.Layout.Pages)[i]
			if p.Id != nil && *p.Id == pageID {
				apiResponse = p
				break
			}
		}
		if apiResponse == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/work/processes/%s/workItemTypes/%s/layout/pages/%s?api-version=7.1",
			orgURL, processID, witRefName, pageID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-page", url, apiResponse)
		return nil
	}
}
