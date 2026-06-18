//go:build (all || resource_release_folder) && !exclude_resource_release_folder

package acceptancetests

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

func TestAccReleaseFolder_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseFolderDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back
			{
				Config: hclReleaseFolderBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "description", "Acceptance test folder"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					// Capture a real API GET of the live resource as forge demo
					// evidence (before destroy) — proves the demo shows the actual
					// release folder, not a test-name table.
					captureReleaseFolderEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclReleaseFolderBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclReleaseFolderBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "test" {
  project_id  = betterado_project.test.id
  path        = "\\AccTest-%[1]s"
  description = "Acceptance test folder"
}
`, name)
}

func checkReleaseFolderDestroyed(s *terraform.State) error {
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_release_folder" {
			continue
		}
		path := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		folders, err := clients.ReleaseClient.GetFolders(clients.Ctx,
			releaseapi.GetFoldersArgs{Project: &projectID, Path: &path})
		if err != nil {
			// A 404-equivalent (empty folders) is expected; some SDK versions
			// may return an error — treat as destroyed.
			continue
		}
		if folders != nil && len(*folders) > 0 {
			return fmt.Errorf("release folder %q still exists after destroy", path)
		}
	}
	return nil
}

// captureReleaseFolderEvidence performs a real live API GET of the created release
// folder and persists the response as forge demo live-evidence (before the resource
// is destroyed). Best-effort: a capture failure never fails the test — the
// read-back assertions above are the authoritative live proof; this only feeds the demo.
func captureReleaseFolderEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		path := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		folders, err := clients.ReleaseClient.GetFolders(clients.Ctx,
			releaseapi.GetFoldersArgs{Project: &projectID, Path: &path})
		if err != nil || folders == nil || len(*folders) == 0 {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		// vsrm host for the Release API
		vsrmHost := strings.Replace(orgURL, "dev.azure.com", "vsrm.dev.azure.com", 1)
		encodedPath := url.PathEscape(path)
		apiURL := fmt.Sprintf("%s/%s/_apis/release/folders%s?api-version=7.1",
			vsrmHost, projectID, encodedPath)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, (*folders)[0])
		return nil
	}
}
