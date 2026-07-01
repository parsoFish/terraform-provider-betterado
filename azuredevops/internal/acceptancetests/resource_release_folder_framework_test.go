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

// TestAccReleaseFolderFramework exercises the terraform-plugin-framework implementation
// of betterado_release_folder via the mux provider path:
//
//  1. apply — creates the folder via the Release Management API
//  2. read-back — asserts the folder attributes match
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// Live evidence is captured during the read-back step via CaptureLiveEvidence.
func TestAccReleaseFolderFramework(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_folder.fw_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseFolderFrameworkDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back + capture evidence
			{
				Config: hclReleaseFolderFramework(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "path", `\AccTestFW-`+name),
					resource.TestCheckResourceAttr(tfNode, "description", "Acceptance test framework folder"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureReleaseFolderFrameworkEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclReleaseFolderFramework(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclReleaseFolderFramework(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "fw_test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "fw_test" {
  project_id  = betterado_project.fw_test.id
  path        = "\\AccTestFW-%[1]s"
  description = "Acceptance test framework folder"
}
`, name)
}

func checkReleaseFolderFrameworkDestroyed(s *terraform.State) error {
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
			// A 404-equivalent (empty folders) is expected; treat as destroyed.
			continue
		}
		if folders != nil && len(*folders) > 0 {
			return fmt.Errorf("release folder %q still exists after destroy", path)
		}
	}
	return nil
}

// captureReleaseFolderFrameworkEvidence performs a real live API GET of the
// created release folder and persists the response as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func captureReleaseFolderFrameworkEvidence(tfNode string) resource.TestCheckFunc {
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
		vsrmHost := strings.Replace(orgURL, "dev.azure.com", "vsrm.dev.azure.com", 1)
		encodedPath := url.PathEscape(path)
		apiURL := fmt.Sprintf("%s/%s/_apis/release/folders%s?api-version=7.1",
			vsrmHost, projectID, encodedPath)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, (*folders)[0])
		return nil
	}
}
