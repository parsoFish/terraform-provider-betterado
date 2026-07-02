//go:build (all || resource_security_permissions) && !exclude_resource_security_permissions

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

// TestAccSecurityPermissionsFramework exercises the terraform-plugin-framework
// implementation of betterado_security_permissions via the mux provider path:
//
//  1. apply — sets permissions on the Project namespace for the Readers group
//  2. read-back checkpoint — asserts each permission value + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// Uses the Project namespace (52d39943-cb85-4d7f-8fa8-c6baac873819) and a
// freshly created ADO project so no manual setup is required beyond env vars.
func TestAccSecurityPermissionsFramework(t *testing.T) {
	tfNodePermissions := "betterado_security_permissions.fw_permissions"
	projectName := testutils.GenerateResourceName()

	config := hclSecurityPermissionsFramework(projectName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodePermissions, "namespace_id"),
					resource.TestCheckResourceAttrSet(tfNodePermissions, "token"),
					resource.TestCheckResourceAttrSet(tfNodePermissions, "principal"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.%", "3"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.GENERIC_READ", "allow"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.GENERIC_WRITE", "deny"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.DELETE", "deny"),
					captureSecurityPermissionsFrameworkEvidence(tfNodePermissions),
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

// hclSecurityPermissionsFramework builds HCL that:
// 1. Creates an ADO project (so we have a known project_id)
// 2. Uses the framework betterado_security_namespace data source to get the Project namespace ID
// 3. Uses the framework betterado_security_namespace_token data source to build the token
// 4. Uses data.betterado_identity_group to find the Readers group
// 5. Creates betterado_security_permissions with the framework resource
func hclSecurityPermissionsFramework(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "project" {
  name = %[1]q
}

data "betterado_security_namespace" "project_ns" {
  name = "Project"
}

data "betterado_security_namespace_token" "project_token" {
  namespace_name = "Project"
  identifiers = {
    project_id = betterado_project.project.id
  }
}

data "betterado_identity_group" "readers" {
  project_id = betterado_project.project.id
  name       = "[${betterado_project.project.name}]\\Readers"
}

resource "betterado_security_permissions" "fw_permissions" {
  namespace_id = data.betterado_security_namespace.project_ns.id
  token        = data.betterado_security_namespace_token.project_token.token
  principal    = data.betterado_identity_group.readers.subject_descriptor

  permissions = {
    GENERIC_READ  = "allow"
    GENERIC_WRITE = "deny"
    DELETE        = "deny"
  }
}
`, projectName)
}

// captureSecurityPermissionsFrameworkEvidence writes
// .forge/live-evidence/acceptance-resource.json with a real ADO Security REST API
// ACL GET URL. Best-effort: a failure never fails the test.
func captureSecurityPermissionsFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		namespaceID := res.Primary.Attributes["namespace_id"]
		token := res.Primary.Attributes["token"]
		if namespaceID == "" || token == "" {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf(
			"%s/_apis/accesscontrollists/%s?token=%s&api-version=7.1",
			orgURL,
			namespaceID,
			token,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, nil)
		return nil
	}
}
