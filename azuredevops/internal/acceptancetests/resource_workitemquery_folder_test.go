package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// Folder under an area (Shared Queries)
func TestAccWorkItemQueryFolder_UnderArea(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	folderName := "tfacc-folder-area"

	config := hclWorkItemQueryFolderUnderArea(projectName, folderName)

	res := "betterado_workitemquery_folder.folder"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeTestCheckFunc(
				testutils.CheckProjectExists(projectName),
				resource.TestCheckResourceAttr(res, "name", folderName),
				resource.TestCheckResourceAttrSet(res, "project_id"),
			),
		}},
	})
}

// Folder under a folder (child folder has parent_id referencing parent folder)
func TestAccWorkItemQueryFolder_UnderFolder(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	parentFolderName := "tfacc-folder-parent"
	childFolderName := "tfacc-folder-child"

	config := hclWorkItemQueryFolderUnderFolder(projectName, parentFolderName, childFolderName)

	parentRes := "betterado_workitemquery_folder.parent"
	childRes := "betterado_workitemquery_folder.child"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeTestCheckFunc(
				testutils.CheckProjectExists(projectName),
				resource.TestCheckResourceAttr(parentRes, "name", parentFolderName),
				resource.TestCheckResourceAttr(childRes, "name", childFolderName),
				resource.TestCheckResourceAttrSet(childRes, "parent_id"),
			),
		}},
	})
}

func hclWorkItemQueryFolderUnderArea(projectName, folderName string) string {
	return testutils.HclProjectResource(projectName) + fmt.Sprintf(`
resource "betterado_workitemquery_folder" "folder" {
  project_id = betterado_project.project.id
  name       = "%s"
  area       = "My Queries"
}
`, folderName)
}

func hclWorkItemQueryFolderUnderFolder(projectName, parentFolderName, childFolderName string) string {
	return testutils.HclProjectResource(projectName) + fmt.Sprintf(`
resource "betterado_workitemquery_folder" "parent" {
  project_id = betterado_project.project.id
  name       = "%s"
  area       = "My Queries"
}

resource "betterado_workitemquery_folder" "child" {
  project_id = betterado_project.project.id
  name       = "%s"
  parent_id  = betterado_workitemquery_folder.parent.id
}
`, parentFolderName, childFolderName)
}
