//go:build (all || resource_release_definition_permissions) && !exclude_resource_release_definition_permissions

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/datahelper"
)

// TestAccReleaseDefinitionPermissionsFramework exercises the terraform-plugin-framework
// implementation of betterado_release_definition_permissions via the mux provider path:
//
//  1. apply — sets all 13 writable ReleaseManagement ACL bits via the framework resource
//  2. read-back checkpoint — asserts each permission value + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — cleans up cleanly
//
// Uses SharedReleaseFixture to obtain a pre-existing persistent project
// (betterado-standing-demo) so no new ADO project is created.
func TestAccReleaseDefinitionPermissionsFramework(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	tfNodeRoot := "betterado_release_definition_permissions.fw_permissions"

	// All writable ReleaseManagement2 ACL bits (mirrors SDKv2 test coverage).
	allPerms := map[string]string{
		"ViewReleases":                 "Allow",
		"EditReleaseEnvironment":       "Deny",
		"DeleteReleases":               "Deny",
		"ManageReleaseSettings":        "Allow",
		"ViewReleaseDefinition":        "Allow",
		"EditReleaseDefinition":        "Allow",
		"DeleteReleaseDefinition":      "Deny",
		"ManageReleaseApprovers":       "NotSet",
		"CreateReleases":               "Allow",
		"ManageReleases":               "Allow",
		"AdministerReleasePermissions": "Deny",
		"ManageDeployments":            "Allow",
		"DeleteReleaseEnvironment":     "Deny",
	}
	config := hclReleaseDefinitionPermissionsFramework(fixture, allPerms)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionFrameworkPermissionsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "release_definition_id"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "13"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ViewReleases", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.EditReleaseEnvironment", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleases", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ManageReleaseSettings", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ViewReleaseDefinition", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.EditReleaseDefinition", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleaseDefinition", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ManageReleaseApprovers", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CreateReleases", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ManageReleases", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.AdministerReleasePermissions", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ManageDeployments", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleaseEnvironment", "deny"),
					captureReleaseDefinitionPermissionsFrameworkEvidence(tfNodeRoot),
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

// hclReleaseDefinitionPermissionsFramework builds HCL that creates a release definition
// and attaches permissions to it via betterado_release_definition_permissions.
// Uses the shared fixture project and the Readers group as the principal.
func hclReleaseDefinitionPermissionsFramework(fixture SharedFixtureResult, perms map[string]string) string {
	rootPermissions := datahelper.JoinMap(perms, "=", "\n")
	releaseName := testutils.GenerateResourceName()

	return fmt.Sprintf(`
resource "betterado_release_definition" "fw_release" {
  project_id = %[1]q
  name       = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}

data "betterado_group" "fw-readers" {
  project_id = %[1]q
  name       = "Readers"
}

resource "betterado_release_definition_permissions" "fw_permissions" {
  project_id            = %[1]q
  principal             = data.betterado_group.fw-readers.id
  release_definition_id = betterado_release_definition.fw_release.id

  permissions = {
    %[3]s
  }
}
`, fixture.ProjectID, releaseName, rootPermissions)
}

// captureReleaseDefinitionPermissionsFrameworkEvidence writes
// .forge/live-evidence/acceptance-resource.json with a real vsrm.dev.azure.com
// ACL REST GET URL. Best-effort: a failure never fails the test.
func captureReleaseDefinitionPermissionsFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		releaseDefinitionID := res.Primary.Attributes["release_definition_id"]
		if projectID == "" || releaseDefinitionID == "" {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		vsrmURL := strings.Replace(orgURL, "https://dev.azure.com", "https://vsrm.dev.azure.com", 1)
		url := fmt.Sprintf(
			"%s/_apis/accesscontrollists/c788c23e-1b46-4162-8f5e-d7585343b5de?token=%s%%2F%s&api-version=7.1",
			vsrmURL,
			projectID, releaseDefinitionID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, nil)
		return nil
	}
}

// checkReleaseDefinitionFrameworkPermissionsDestroyed verifies that the release
// definition created during the test has been cleaned up (uses getDirectClient
// so it works with the mux provider path).
func checkReleaseDefinitionFrameworkPermissionsDestroyed(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type == "betterado_release_definition" {
			return checkReleaseDefinitionDestroyed(s)
		}
	}
	return nil
}
