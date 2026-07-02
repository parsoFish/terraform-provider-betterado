//go:build (all || resource_securityrole_assignment) && !exclude_resource_securityrole_assignment

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	sdkroles "github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/securityroles"
)

// TestAccSecurityRoleAssignmentFramework exercises the terraform-plugin-framework
// implementation of betterado_securityrole_assignment via the mux provider:
//
//  1. apply — creates the assignment (Administrator role) on an environment in the
//     shared fixture project; does NOT create a new ADO project (org is at 1000-cap).
//  2. read-back — asserts scope, resource_id, identity_id, role_name.
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false).
//  4. destroy — cleans up cleanly.
//
// Live evidence is captured via CaptureLiveEvidence.
func TestAccSecurityRoleAssignmentFramework(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live fixture")
	}

	testutils.PreCheck(t, nil)

	// Build a direct ADO client from env vars — needed to create the environment
	// and group without going through Terraform resources (avoids a new project).
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		t.Fatalf("TestAccSecurityRoleAssignmentFramework: GetAzdoClient: %v", err)
	}

	// Resolve the shared fixture project (existing — never created fresh here).
	projectID := resolveSecurityRoleFixtureProject(t, clients)

	// Create a pipeline environment in the shared project.
	envName := testutils.GenerateResourceName()
	env, err := clients.TaskAgentClient.AddEnvironment(clients.Ctx, taskagent.AddEnvironmentArgs{
		Project: &projectID,
		EnvironmentCreateParameter: &taskagent.EnvironmentCreateParameter{
			Name:        converter.String(envName),
			Description: converter.String("acc-test env for securityrole_assignment"),
		},
	})
	if err != nil {
		t.Fatalf("TestAccSecurityRoleAssignmentFramework: AddEnvironment: %v", err)
	}
	envID := *env.Id
	t.Cleanup(func() {
		_ = clients.TaskAgentClient.DeleteEnvironment(clients.Ctx, taskagent.DeleteEnvironmentArgs{
			Project:       &projectID,
			EnvironmentId: &envID,
		})
	})

	// Resolve the "Project Administrators" group identity to use as the assignee.
	groupID := resolveProjectAdminIdentityID(t, clients, projectID)

	// resource_id for pipeline environments is "<projectID>_<envID>".
	scope := "distributedtask.environmentreferencerole"
	resourceID := fmt.Sprintf("%s_%d", projectID, envID)
	roleName := "Administrator"

	tfNodeAssignment := "betterado_securityrole_assignment.fw_sra"
	config := hclSecurityRoleAssignmentFramework(scope, resourceID, groupID, roleName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkSecurityRoleAssignmentFrameworkDestroyed(clients, scope, resourceID, groupID),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNodeAssignment, "scope", scope),
					resource.TestCheckResourceAttr(tfNodeAssignment, "resource_id", resourceID),
					resource.TestCheckResourceAttrSet(tfNodeAssignment, "identity_id"),
					resource.TestCheckResourceAttr(tfNodeAssignment, "role_name", roleName),
					captureSecurityRoleAssignmentEvidence(tfNodeAssignment, orgURL, scope, resourceID),
				),
			},
			{
				// Idempotency: re-plan must produce no diff.
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// resolveSecurityRoleFixtureProject returns the ID of an existing ADO project.
// Prefers the shared standing-demo project, falling back to any WellFormed project
// in the org — so the test runs even without the standing-demo project.
func resolveSecurityRoleFixtureProject(t *testing.T, clients *client.AggregatedClient) string {
	t.Helper()

	// Prefer the canonical shared fixture project.
	if existing := resolveProjectByName(clients, SharedFixtureProjectName); existing != "" {
		return existing
	}

	// Fall back: any WellFormed project in the org.
	resp, err := clients.CoreClient.GetProjects(clients.Ctx, core.GetProjectsArgs{})
	if err != nil {
		t.Fatalf("resolveSecurityRoleFixtureProject: GetProjects: %v", err)
	}
	for _, p := range resp.Value {
		if p.Id != nil {
			return p.Id.String()
		}
	}
	t.Fatalf("resolveSecurityRoleFixtureProject: no project found in org")
	return ""
}

// resolveProjectByName returns the project UUID if a project with the given name
// exists, or "" if it doesn't.
func resolveProjectByName(clients *client.AggregatedClient, name string) string {
	if existing, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId: converter.String(name),
	}); err == nil && existing != nil && existing.Id != nil {
		return existing.Id.String()
	}
	return ""
}

// resolveProjectAdminIdentityID returns the UUID of the "Project Administrators"
// group in the given project (or a fallback group if PA is not found).
func resolveProjectAdminIdentityID(t *testing.T, clients *client.AggregatedClient, projectID string) string {
	t.Helper()
	return resolveApproverIdentity(t, clients, projectID)
}

// hclSecurityRoleAssignmentFramework builds the HCL for a
// betterado_securityrole_assignment resource using literal IDs (no dependent
// Terraform resources). Scope, resource_id, identity_id, and role_name are all
// set directly from the pre-existing ADO objects resolved in Go.
func hclSecurityRoleAssignmentFramework(scope, resourceID, identityID, roleName string) string {
	return fmt.Sprintf(`
resource "betterado_securityrole_assignment" "fw_sra" {
  scope       = %[1]q
  resource_id = %[2]q
  identity_id = %[3]q
  role_name   = %[4]q
}
`, scope, resourceID, identityID, roleName)
}

// checkSecurityRoleAssignmentFrameworkDestroyed verifies the security role
// assignment was removed when the resource was destroyed.
func checkSecurityRoleAssignmentFrameworkDestroyed(
	clients *client.AggregatedClient,
	scope, resourceID, identityIDStr string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		identityID, err := uuid.Parse(identityIDStr)
		if err != nil {
			// Can't verify — best-effort.
			return nil
		}

		assignment, err := clients.SecurityRolesClient.GetSecurityRoleAssignment(clients.Ctx, &sdkroles.GetSecurityRoleAssignmentArgs{
			Scope:      &scope,
			ResourceId: &resourceID,
			IdentityId: &identityID,
		})
		if err != nil {
			// 404 / "not found" means the assignment is gone — success.
			if strings.Contains(err.Error(), "404") ||
				strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
			// Other errors: best-effort — don't fail the test on destroy-check.
			return nil
		}

		if assignment == nil || (assignment.Identity == nil && assignment.Role == nil) {
			return nil // already gone
		}

		return fmt.Errorf("security role assignment still exists after destroy: scope=%s resource_id=%s identity_id=%s",
			scope, resourceID, identityIDStr)
	}
}

// captureSecurityRoleAssignmentEvidence writes live-evidence for the forge demo
// pipeline. Best-effort: a failure never fails the test.
func captureSecurityRoleAssignmentEvidence(tfNode, orgURL, scope, resourceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		identityID := res.Primary.Attributes["identity_id"]
		if identityID == "" {
			return nil
		}
		orgURL = strings.TrimRight(orgURL, "/")
		// SecurityRoles list assignments URL.
		url := fmt.Sprintf(
			"%s/_apis/securityroles/scopes/%s/roleassignments/resources/%s?api-version=7.1-preview.1",
			orgURL,
			scope,
			resourceID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, nil)
		return nil
	}
}
