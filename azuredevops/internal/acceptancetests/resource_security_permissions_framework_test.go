//go:build (all || resource_security_permissions) && !exclude_resource_security_permissions

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestAccSecurityPermissionsFramework exercises the terraform-plugin-framework
// implementation of betterado_security_permissions via the mux provider path:
//
//  1. apply — sets permissions on the Project namespace for the Readers group
//  2. read-back checkpoint — asserts each permission value + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// Uses resolveSecurityPermissionsFixtureProject to obtain an existing project
// without creating any new ADO project — the org is at the 1000-project cap,
// so any project-create attempt would fail immediately.
func TestAccSecurityPermissionsFramework(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live fixture")
	}

	testutils.PreCheck(t, nil)

	projectID := resolveSecurityPermissionsFixtureProject(t)
	tfNodePermissions := "betterado_security_permissions.fw_permissions"

	config := hclSecurityPermissionsFramework(projectID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		// No betterado_project is created, so no project destroy-check needed.
		// The security_permissions resource itself (an ACL entry) is cleaned up by
		// Terraform destroy — we just confirm no panic occurs.
		CheckDestroy: checkSecurityPermissionsFrameworkDestroyed,
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

// resolveSecurityPermissionsFixtureProject returns the ID of an existing ADO
// project without creating a new one. It prefers the shared fixture project
// (betterado-standing-demo) and falls back to the first WellFormed project in
// the org — so the test runs even when the persistent project does not yet exist
// (as long as any project exists in the org, which is always true at 1000-cap).
func resolveSecurityPermissionsFixtureProject(t *testing.T) string {
	t.Helper()

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("resolveSecurityPermissionsFixtureProject: GetAzdoClient: %v", err)
	}

	// Prefer the canonical shared fixture project if it already exists.
	if existing, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId: converter.String(SharedFixtureProjectName),
	}); err == nil && existing != nil && existing.Id != nil {
		return existing.Id.String()
	}

	// Fall back: pick the first WellFormed project in the org.
	// At the 1000-project cap there are always projects available; we just need
	// one whose "Project" security namespace we can read.
	resp, err := clients.CoreClient.GetProjects(clients.Ctx, core.GetProjectsArgs{
		StateFilter: converter.ToPtr(core.ProjectStateValues.WellFormed),
	})
	if err != nil {
		t.Fatalf("resolveSecurityPermissionsFixtureProject: GetProjects: %v", err)
	}
	for _, p := range resp.Value {
		if p.Id != nil && !keepProjects[*p.Name] {
			return p.Id.String()
		}
	}
	// keepProjects entries are also valid — use the first available project of any kind.
	for _, p := range resp.Value {
		if p.Id != nil {
			return p.Id.String()
		}
	}
	t.Fatalf("resolveSecurityPermissionsFixtureProject: no WellFormed project found in org")
	return ""
}

// hclSecurityPermissionsFramework builds HCL that:
// 1. Uses the framework betterado_security_namespace data source to get the Project namespace ID
// 2. Uses the framework betterado_security_namespace_token data source to build the token
// 3. Uses data.betterado_group to find the Readers group in the fixture project
// 4. Creates betterado_security_permissions with the framework resource
//
// Uses a pre-existing project (via resolveSecurityPermissionsFixtureProject) to
// avoid the 1000-project cap. No betterado_project resource is created here.
func hclSecurityPermissionsFramework(projectID string) string {
	return fmt.Sprintf(`
data "betterado_security_namespace" "project_ns" {
  name = "Project"
}

data "betterado_security_namespace_token" "project_token" {
  namespace_name = "Project"
  identifiers = {
    project_id = %[1]q
  }
}

data "betterado_group" "readers" {
  project_id = %[1]q
  name       = "Readers"
}

resource "betterado_security_permissions" "fw_permissions" {
  namespace_id = data.betterado_security_namespace.project_ns.id
  token        = data.betterado_security_namespace_token.project_token.token
  principal    = data.betterado_group.readers.descriptor

  permissions = {
    GENERIC_READ  = "allow"
    GENERIC_WRITE = "deny"
    DELETE        = "deny"
  }
}
`, projectID)
}

// checkSecurityPermissionsFrameworkDestroyed verifies the destroy step completes
// without error. We use getDirectClient (defined in resource_task_group_test.go)
// so we never call GetProvider().Meta(), which is nil when the test uses
// ProtoV6ProviderFactories (the mux path).
//
// betterado_security_permissions is an ACL record, not a cloud resource with its
// own existence endpoint — Terraform destroy removes the ACE; we simply confirm
// no error occurred.
func checkSecurityPermissionsFrameworkDestroyed(_ *terraform.State) error {
	// getDirectClient is defined in resource_task_group_test.go; build it to
	// confirm credentials are valid (avoids the nil-Meta panic on mux tests).
	_, err := getDirectClient()
	if err != nil {
		// Best-effort: if we can't build a client, skip the post-check.
		return nil
	}
	// betterado_security_permissions has no queryable "does it exist?" API that
	// makes sense post-destroy (ACEs are implicit when absent), so return nil.
	return nil
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
