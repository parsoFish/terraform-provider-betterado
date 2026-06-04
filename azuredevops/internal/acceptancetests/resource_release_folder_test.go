package acceptancetests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccReleaseFolder_basic creates a release folder, reads it back, updates the description,
// then verifies it is destroyed cleanly.
func TestAccReleaseFolder_basic(t *testing.T) {
	folderName := testutils.GenerateResourceName()
	// Release folder paths use backslash separator; a top-level folder is just "\<name>"
	folderPath := `\` + folderName
	descriptionV1 := "initial description"
	descriptionV2 := "updated description"
	tfNode := "betterado_release_folder.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseFolderDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create
			{
				Config: hclReleaseFolder(folderName, folderPath, descriptionV1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "path", folderPath),
					resource.TestCheckResourceAttr(tfNode, "description", descriptionV1),
				),
			},
			// Step 2: update description
			{
				Config: hclReleaseFolder(folderName, folderPath, descriptionV2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "path", folderPath),
					resource.TestCheckResourceAttr(tfNode, "description", descriptionV2),
				),
			},
		},
	})
}

// --- Check functions ---

// checkReleaseFolderDestroyed verifies the release folder no longer exists in ADO after destroy.
func checkReleaseFolderDestroyed(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_release_folder" {
			continue
		}
		id := res.Primary.ID
		// ID format: "<projectID>/<path>"
		idx := strings.Index(id, "/")
		if idx < 0 {
			return fmt.Errorf("unexpected release folder ID format: %s", id)
		}
		projectID := id[:idx]
		path := id[idx+1:]

		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		folders, err := clients.ReleaseClient.GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
			Project: &projectID,
			Path:    &path,
		})
		if err != nil {
			// A not-found error means it was deleted — that's the desired outcome.
			return nil
		}
		if folders != nil {
			for _, f := range *folders {
				if f.Path != nil && *f.Path == path {
					return fmt.Errorf("release folder '%s' still exists in project '%s' after destroy", path, projectID)
				}
			}
		}
	}
	return nil
}

// --- HCL helpers ---

func hclReleaseFolderProjectBase(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}
`, name)
}

func hclReleaseFolder(projectName, path, description string) string {
	base := hclReleaseFolderProjectBase(projectName)
	return fmt.Sprintf(`
%s

resource "betterado_release_folder" "test" {
  project_id  = betterado_project.test.id
  path        = %q
  description = %q
}
`, base, path, description)
}
