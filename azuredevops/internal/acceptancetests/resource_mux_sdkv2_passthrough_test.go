//go:build all

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

// TestAccMuxSdkv2Passthrough verifies that existing SDKv2 resources continue to
// work correctly when served through the mux entrypoint introduced in
// INIT-2026-06-19-framework-mux-entrypoint. It uses betterado_release_folder as a
// representative SDKv2 resource (lightweight, no required project sub-resources).
//
// Evidence is captured in .forge/live-evidence/acceptance-resource.json via
// CaptureLiveEvidence, satisfying the forge demo live-evidence contract.
func TestAccMuxSdkv2Passthrough(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	folderPath := `\MuxSmokeTest`
	tfNode := "betterado_release_folder.smoke"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkMuxSmokeFolderDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclMuxSmokeFolder(projectName, folderPath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						tfNode, "path", folderPath,
					),
					captureMuxPassthroughEvidence(tfNode),
				),
			},
		},
	})
}

func hclMuxSmokeFolder(projectName, folderPath string) string {
	return fmt.Sprintf(`
resource "betterado_project" "smoke" {
  name               = %q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "smoke" {
  project_id = betterado_project.smoke.id
  path       = %q
}
`, projectName, folderPath)
}

func checkMuxSmokeFolderDestroyed(s *terraform.State) error {
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
			continue
		}
		if folders != nil && len(*folders) > 0 {
			return fmt.Errorf("mux smoke release folder %q still exists after destroy", path)
		}
	}
	return nil
}

// captureMuxPassthroughEvidence performs a real live API GET of the created release
// folder via the Release API and persists the response as forge demo live-evidence
// (before the resource is destroyed). The label "acceptance-resource" matches the
// forge unifier's checkpoint. Best-effort: a capture failure never fails the test.
func captureMuxPassthroughEvidence(tfNode string) resource.TestCheckFunc {
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
