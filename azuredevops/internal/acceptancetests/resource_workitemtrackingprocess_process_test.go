//go:build (all || resource_workitemtrackingprocess_process) && !exclude_resource_workitemtrackingprocess_process

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

func TestAccWorkitemtrackingprocessProcess_Basic(t *testing.T) {
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_process.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: process(processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", processName),
					resource.TestCheckResourceAttrSet(tfNode, "reference_name"),
					resource.TestCheckResourceAttr(tfNode, "is_default", "false"),
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "true"),
					resource.TestCheckResourceAttrSet(tfNode, "customization_type"),
					resource.TestCheckResourceAttr(tfNode, "parent_process_type_id", "adcc42ab-9882-485e-a3ed-7678f01f66bc"),
					captureProcessEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getProcessStateIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessProcess_CreateDisabled(t *testing.T) {
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_process.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: disabledProcess(processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", processName),
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "false"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getProcessStateIdFunc(tfNode),
			},
		},
	})
}

func TestAccWorkitemtrackingprocessProcess_CreateAndUpdate(t *testing.T) {
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_process.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: process(processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", processName),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getProcessStateIdFunc(tfNode),
			},
			{
				Config: disabledProcess(processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "is_enabled", "false"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getProcessStateIdFunc(tfNode),
			},
		},
	})
}

// process() and disabledProcess() are defined in workitemtrackingprocess_shared_test.go
// (no build tag) so they are available to all workitemtrackingprocess acceptance tests.

func getProcessStateIdFunc(tfNode string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		res := state.RootModule().Resources[tfNode]
		return res.Primary.Attributes["id"], nil
	}
}

// captureProcessEvidence reads back the process via API and writes live evidence.
func captureProcessEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		processID := res.Primary.ID

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		agg, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}

		process, err := agg.WorkItemTrackingProcessClient.GetProcessByItsId(agg.Ctx, workitemtrackingprocess.GetProcessByItsIdArgs{
			ProcessTypeId: converter.UUID(processID),
			Expand:        &workitemtrackingprocess.GetProcessExpandLevelValues.None,
		})
		if err != nil {
			return nil // best-effort
		}

		url := fmt.Sprintf("%s/_apis/work/processes/%s?api-version=7.1",
			strings.TrimRight(orgURL, "/"), processID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-process", url, process)
		return nil
	}
}

// agileSystemProcessTypeId is defined in workitemtrackingprocess_shared_test.go
// so it is available to all workitemtrackingprocess acceptance tests regardless of build tags.
