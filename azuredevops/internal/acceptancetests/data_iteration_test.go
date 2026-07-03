//go:build (all || resource_iteration) && !exclude_resource_iteration

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
)

// getIterationDirectClient builds an AggregatedClient from AZDO env vars.
// ProtoV6ProviderFactories does not expose Meta(), so evidence helpers build their own.
func getIterationDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureIterationEvidence performs a live API GET of the iteration classification node
// and persists the response as forge demo live-evidence. Best-effort: never fails the test.
func captureIterationEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		path := res.Primary.Attributes["path"]

		clients, err := getIterationDirectClient()
		if err != nil {
			return nil //nolint:nilerr
		}
		structureType := workitemtracking.TreeStructureGroupValues.Iterations
		node, err := clients.WorkItemTrackingClient.GetClassificationNode(clients.Ctx, workitemtracking.GetClassificationNodeArgs{
			Project:        &projectID,
			StructureGroup: &structureType,
			Path:           &path,
		})
		if err != nil || node == nil {
			return nil //nolint:nilerr
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		url := fmt.Sprintf("%s/%s/_apis/wit/classificationnodes/iterations/%s?api-version=7.1",
			orgURL, SharedFixtureProjectName, path)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-iteration", url, node)
		return nil
	}
}

// hclIterationDataSource returns HCL that looks up the shared fixture project and reads
// its root iteration classification node. Uses a data source lookup to avoid creating a
// new project (the ADO org is at the 1000-project cap).
func hclIterationDataSource(fetchChildren bool) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_iteration" "root-iteration" {
  project_id     = data.betterado_project.shared.id
  fetch_children = %[2]t
}
`, SharedFixtureProjectName, fetchChildren)
}

func TestAccIterationDataSource_Read(t *testing.T) {
	tfNode := "data.betterado_iteration.root-iteration"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclIterationDataSource(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "path"),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "has_children"),
					captureIterationEvidence(tfNode),
				),
			},
		},
	})
}

func TestAccIterationDataSource_ReadNoChildren(t *testing.T) {
	tfNode := "data.betterado_iteration.root-iteration"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclIterationDataSource(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "path"),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "has_children"),
				),
			},
		},
	})
}
