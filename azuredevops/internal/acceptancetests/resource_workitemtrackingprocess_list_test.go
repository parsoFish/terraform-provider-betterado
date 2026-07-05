package acceptancetests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

func TestAccWorkitemtrackingprocessList_Basic(t *testing.T) {
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicList(listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "url"),
					captureListEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessList_Update(t *testing.T) {
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: basicList(listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: updatedList(listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: basicList(listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessList_Integer(t *testing.T) {
	listName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkListDestroyed,
		Steps: []resource.TestStep{
			{
				Config: integerList(listName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func basicList(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name  = "%s"
  items = ["Red", "Green", "Blue"]
}
`, name)
}

func updatedList(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name         = "%s"
  items        = ["Red", "Green", "Blue", "Yellow"]
  is_suggested = true
}
`, name)
}

func integerList(name string) string {
	return fmt.Sprintf(`
resource "betterado_workitemtrackingprocess_list" "test" {
  name  = "%s"
  type  = "integer"
  items = ["1", "2", "3"]
}
`, name)
}

// captureListEvidence reads back the list via the Azure DevOps API and calls
// testutils.CaptureLiveEvidence to satisfy AC4.
func captureListEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}

		listID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		agg, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			return nil //nolint:nilerr // best-effort: client build failure does not fail the test
		}

		apiResponse, err := agg.WorkItemTrackingProcessClient.GetList(agg.Ctx, workitemtrackingprocess.GetListArgs{
			ListId: &listID,
		})
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}

		url := fmt.Sprintf("%s/_apis/work/processes/lists/%s?api-version=7.1",
			orgURL, listID.String())
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-list", url, apiResponse)
		return nil
	}
}

// checkListDestroyed verifies that all lists referenced in the state are destroyed.
// Builds the client directly from environment variables so this works regardless
// of whether the test uses SDKv2 or ProtoV6ProviderFactories (mux).
func checkListDestroyed(s *terraform.State) error {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		return fmt.Errorf("checkListDestroyed: building client: %w", err)
	}
	timeout := 10 * time.Second

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_workitemtrackingprocess_list" {
			continue
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return err
		}

		err = retry.RetryContext(clients.Ctx, timeout, func() *retry.RetryError {
			_, err := clients.WorkItemTrackingProcessClient.GetList(clients.Ctx, workitemtrackingprocess.GetListArgs{
				ListId: &id,
			})
			if err == nil {
				return retry.RetryableError(fmt.Errorf("list with ID %s should not exist", id.String()))
			}
			if utils.ResponseWasNotFound(err) {
				return nil
			}

			return retry.NonRetryableError(err)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
