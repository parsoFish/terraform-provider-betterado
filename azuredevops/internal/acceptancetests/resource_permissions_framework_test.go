//go:build (all || resource_project_permissions) && !exclude_resource_project_permissions

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccProjectPermissionsFramework exercises the terraform-plugin-framework
// implementation of betterado_project_permissions via the mux provider path:
//
//  1. apply — creates a project and sets project-level permissions on the Readers group
//  2. read-back checkpoint — asserts each permission value + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// betterado_project_permissions is the representative resource for the permissions
// package migration (simplest token — needs only project_id).
func TestAccProjectPermissionsFramework(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live fixture")
	}

	testutils.PreCheck(t, nil)

	projectName := testutils.GenerateResourceName()
	config := hclProjectPermissionsFramework(projectName)

	tfNodeProject := "betterado_project.project"
	tfNodePermissions := "betterado_project_permissions.fw_permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNodeProject, "id"),
					resource.TestCheckResourceAttrSet(tfNodePermissions, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodePermissions, "principal"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.DELETE", "Deny"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.EDIT_BUILD_STATUS", "NotSet"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.WORK_ITEM_MOVE", "Allow"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.DELETE_TEST_RESULTS", "Deny"),
					captureProjectPermissionsFrameworkEvidence(tfNodePermissions),
				),
			},
			{
				// idempotency check: re-plan after apply must produce no diff
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclProjectPermissionsFramework builds HCL that creates a project, looks up
// the Readers group, and creates betterado_project_permissions via the framework path.
func hclProjectPermissionsFramework(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "project" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_project_permissions" "fw_permissions" {
  project_id  = betterado_project.project.id
  principal   = data.betterado_group.readers.descriptor
  permissions = {
    DELETE              = "Deny"
    EDIT_BUILD_STATUS   = "NotSet"
    WORK_ITEM_MOVE      = "Allow"
    DELETE_TEST_RESULTS = "Deny"
  }
}
`, projectName)
}

// captureProjectPermissionsFrameworkEvidence writes
// .forge/live-evidence/acceptance-resource.json with a real ADO Security REST API
// ACL GET URL. Best-effort: a failure never fails the test.
//
// Token format: "$PROJECT:vstfs:///Classification/TeamProject/{projectId}"
// URL: {orgURL}/_apis/accesscontrollists/52d39943-cb85-4d7f-8fa8-c6baac873819?token={encodedToken}&api-version=7.1
func captureProjectPermissionsFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	// Project security namespace ID (from namespaces.go SecurityNamespaceIDValues.Project)
	const projectNsID = "52d39943-cb85-4d7f-8fa8-c6baac873819"

	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		if projectID == "" {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		// Build the ACL token exactly as buildProjectToken() does.
		token := fmt.Sprintf("$PROJECT:vstfs:///Classification/TeamProject/%s", projectID)
		url := fmt.Sprintf(
			"%s/_apis/accesscontrollists/%s?token=%s&api-version=7.1",
			orgURL,
			projectNsID,
			token,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, nil)
		return nil
	}
}
