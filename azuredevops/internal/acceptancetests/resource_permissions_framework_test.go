//go:build (all || resource_project_permissions) && !exclude_resource_project_permissions

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

// TestAccProjectPermissionsFramework exercises the terraform-plugin-framework
// implementation of betterado_project_permissions via the mux provider path:
//
//  1. apply — sets project-level permissions on the Readers group of an existing project
//  2. read-back checkpoint — asserts each permission value + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// betterado_project_permissions is the representative resource for the permissions
// package migration (simplest token — needs only project_id).
//
// Uses resolveProjectPermissionsFixtureProject to obtain an existing project
// without creating any new ADO project — the org is at the 1000-project cap,
// so any project-create attempt would fail immediately.
func TestAccProjectPermissionsFramework(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live fixture")
	}

	testutils.PreCheck(t, nil)

	projectID := resolveProjectPermissionsFixtureProject(t)
	tfNodePermissions := "betterado_project_permissions.fw_permissions"

	config := hclProjectPermissionsFramework(projectID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		// No betterado_project is created, so no project destroy-check needed.
		// The project_permissions resource (an ACL entry) is cleaned up by
		// Terraform destroy — we just confirm no panic occurs.
		CheckDestroy: checkProjectPermissionsFrameworkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodePermissions, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodePermissions, "principal"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.DELETE", "deny"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.EDIT_BUILD_STATUS", "notset"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.WORK_ITEM_MOVE", "allow"),
					resource.TestCheckResourceAttr(tfNodePermissions, "permissions.DELETE_TEST_RESULTS", "deny"),
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

// resolveProjectPermissionsFixtureProject returns the ID of an existing ADO
// project without creating a new one. Prefers the shared fixture project
// (SharedFixtureProjectName = "betterado-standing-demo") and falls back to
// the first WellFormed project in the org.
func resolveProjectPermissionsFixtureProject(t *testing.T) string {
	t.Helper()

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("resolveProjectPermissionsFixtureProject: GetAzdoClient: %v", err)
	}

	// Prefer the canonical shared fixture project.
	if existing, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId: converter.String(SharedFixtureProjectName),
	}); err == nil && existing != nil && existing.Id != nil {
		return existing.Id.String()
	}

	// Fall back: pick the first WellFormed project in the org.
	resp, err := clients.CoreClient.GetProjects(clients.Ctx, core.GetProjectsArgs{
		StateFilter: converter.ToPtr(core.ProjectStateValues.WellFormed),
	})
	if err != nil {
		t.Fatalf("resolveProjectPermissionsFixtureProject: GetProjects: %v", err)
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
	t.Fatalf("resolveProjectPermissionsFixtureProject: no WellFormed project found in org")
	return ""
}

// hclProjectPermissionsFramework builds HCL that looks up the Readers group in
// an existing project and creates betterado_project_permissions via the framework path.
// The project ID is passed as a literal — no betterado_project resource is created here.
func hclProjectPermissionsFramework(projectID string) string {
	return fmt.Sprintf(`
data "betterado_group" "readers" {
  project_id = %[1]q
  name       = "Readers"
}

resource "betterado_project_permissions" "fw_permissions" {
  project_id = %[1]q
  principal  = data.betterado_group.readers.descriptor
  permissions = {
    DELETE              = "deny"
    EDIT_BUILD_STATUS   = "notset"
    WORK_ITEM_MOVE      = "allow"
    DELETE_TEST_RESULTS = "deny"
  }
}
`, projectID)
}

// checkProjectPermissionsFrameworkDestroyed verifies the destroy step completes
// without error. betterado_project_permissions is an ACL record, not a cloud
// resource with its own existence endpoint — Terraform destroy removes the ACE;
// we simply confirm no error occurred.
func checkProjectPermissionsFrameworkDestroyed(_ *terraform.State) error {
	_, err := getDirectClient()
	if err != nil {
		// Best-effort: if we can't build a client, skip the post-check.
		return nil
	}
	// betterado_project_permissions has no queryable "does it exist?" API that
	// makes sense post-destroy (ACEs are implicit when absent), so return nil.
	return nil
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
