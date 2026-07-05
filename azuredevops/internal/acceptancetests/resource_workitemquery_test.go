//go:build (all || resource_workitemquery) && !exclude_resource_workitemquery

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// getQueryDirectClient builds an AggregatedClient directly from AZDO env vars for
// use in CheckDestroy and evidence helpers (ProtoV6ProviderFactories does not expose Meta).
func getQueryDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// checkQueryDestroyed verifies all betterado_workitemquery resources in state have
// been deleted from ADO after a destroy step.
func checkQueryDestroyed(s *terraform.State) error {
	clients, err := getQueryDirectClient()
	if err != nil {
		return fmt.Errorf("checkQueryDestroyed: failed to build ADO client: %v", err)
	}
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_workitemquery" {
			continue
		}
		id := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		_, err := clients.WorkItemTrackingClient.GetQuery(clients.Ctx, workitemtracking.GetQueryArgs{
			Project: &projectID,
			Query:   &id,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				continue // expected
			}
			return fmt.Errorf("error reading query %s after destroy: %v", id, err)
		}
		return fmt.Errorf("query %s still exists after destroy", id)
	}
	return nil
}

// captureQueryEvidence performs a live API GET of the created query and persists
// the response as forge demo live-evidence. Best-effort: failure never fails the test.
func captureQueryEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		id := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getQueryDirectClient()
		if err != nil {
			return nil //nolint:nilerr // best-effort
		}
		q, err := clients.WorkItemTrackingClient.GetQuery(clients.Ctx, workitemtracking.GetQueryArgs{
			Project: &projectID,
			Query:   &id,
		})
		if err != nil || q == nil {
			return nil //nolint:nilerr // best-effort
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		url := fmt.Sprintf("%s/%s/_apis/wit/queries/%s?api-version=7.1", orgURL, projectID, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemquery", url, q)
		return nil
	}
}

// TestAccWorkItemQuery_UnderArea creates a work item query under "My Queries" in the
// shared fixture project, performs a full read-back, captures live evidence,
// verifies idempotency, then destroys.
func TestAccWorkItemQuery_UnderArea(t *testing.T) {
	queryName := testutils.GenerateResourceName()
	wiql := "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project"
	tfNode := "betterado_workitemquery.query"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkQueryDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclWorkItemQueryShared(queryName, wiql),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "name", queryName),
					resource.TestCheckResourceAttr(tfNode, "wiql", wiql),
					resource.TestCheckResourceAttr(tfNode, "area", "My Queries"),
					captureQueryEvidence(tfNode),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclWorkItemQueryShared(queryName, wiql),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclWorkItemQueryShared creates a work item query in the PERSISTENT shared fixture project
// (SharedFixtureProjectName) to avoid the 1000-project ADO org cap.
func hclWorkItemQueryShared(queryName, wiql string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[1]q
}

resource "betterado_workitemquery" "query" {
  project_id = data.betterado_project.test.id
  name       = %[2]q
  area       = "My Queries"
  wiql       = %[3]q
}
`, SharedFixtureProjectName, queryName, wiql)
}
