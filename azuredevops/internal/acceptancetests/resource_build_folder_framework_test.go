//go:build (all || resource_build_folder_framework) && !exclude_resource_build_folder_framework

package acceptancetests

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccBuildFolder_Framework_basic exercises the terraform-plugin-framework implementation
// of betterado_build_folder via the mux provider path:
//
//  1. apply — creates the folder via the Build API with a non-default description
//  2. read-back — asserts folder attributes match the configuration
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// Live evidence is captured during the read-back step via CaptureLiveEvidence.
//
// Uses SharedReleaseFixture to obtain a pre-existing persistent project
// (betterado-standing-demo) so no new ADO project is created — the org is at
// the 1000-project cap, so any project-create attempt would fail immediately.
func TestAccBuildFolder_Framework_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfNode := "betterado_build_folder.fw_test"
	path := `\AccTestFW-` + name

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkBuildFolderFrameworkDestroyed(path, fixture.ProjectID),
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back + capture live evidence
			{
				Config: hclBuildFolderFramework(name, fixture.ProjectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "path", path),
					resource.TestCheckResourceAttr(tfNode, "description", "Acceptance test framework build folder"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureBuildFolderFrameworkEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclBuildFolderFramework(name, fixture.ProjectID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclBuildFolderFramework(name, projectID string) string {
	return fmt.Sprintf(`
resource "betterado_build_folder" "fw_test" {
  project_id  = %[2]q
  path        = "\\AccTestFW-%[1]s"
  description = "Acceptance test framework build folder"
}
`, name, projectID)
}

// checkBuildFolderFrameworkDestroyed returns a CheckDestroyFunc that verifies
// the build folder has been deleted after destroy.
func checkBuildFolderFrameworkDestroyed(path, projectID string) resource.TestCheckDestroyFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkBuildFolderFrameworkDestroyed: build client: %w", err)
		}
		folders, err := clients.BuildClient.GetFolders(clients.Ctx, build.GetFoldersArgs{
			Project: &projectID,
			Path:    &path,
		})
		if err != nil {
			// GetFolders returns an error (including 404-equivalent) — treat as destroyed.
			return nil
		}
		if folders != nil && len(*folders) > 0 {
			return fmt.Errorf("build folder %q still exists after destroy", path)
		}
		return nil
	}
}

// captureBuildFolderFrameworkEvidence performs a real live API GET of the
// created build folder and persists the response as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func captureBuildFolderFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		path := res.Primary.Attributes["path"]
		projectID := res.Primary.Attributes["project_id"]

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort; do not fail the test on evidence capture
		}
		folders, err := clients.BuildClient.GetFolders(clients.Ctx, build.GetFoldersArgs{
			Project: &projectID,
			Path:    &path,
		})
		if err != nil || folders == nil || len(*folders) == 0 {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		// Remove trailing slash if present.
		for len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		encodedPath := url.PathEscape(path)
		apiURL := fmt.Sprintf("%s/%s/_apis/build/folders%s?api-version=7.1-preview.2",
			orgURL, projectID, encodedPath)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, (*folders)[0])
		return nil
	}
}
