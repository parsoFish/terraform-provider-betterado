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

func TestAccReleaseDefinitionPermissions_SetPermissions(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	config := hclReleaseDefinitionPermissions(fixture, map[string]string{
		"ViewReleases":           "Allow",
		"EditReleaseEnvironment": "NotSet",
		"DeleteReleases":         "Deny",
		"CreateReleases":         "Deny",
	})
	tfNodeRoot := "betterado_release_definition_permissions.permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "release_definition_id"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ViewReleases", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.EditReleaseEnvironment", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleases", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CreateReleases", "deny"),
				),
			},
			{
				// idempotency check: plan after apply must be empty
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccReleaseDefinitionPermissions_UpdatePermissions(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	config1 := hclReleaseDefinitionPermissions(fixture, map[string]string{
		"ViewReleases":           "Deny",
		"EditReleaseEnvironment": "NotSet",
		"DeleteReleases":         "Deny",
		"CreateReleases":         "Deny",
	})
	config2 := hclReleaseDefinitionPermissions(fixture, map[string]string{
		"ViewReleases":           "Allow",
		"EditReleaseEnvironment": "Allow",
		"DeleteReleases":         "Deny",
		"CreateReleases":         "NotSet",
	})
	tfNodeRoot := "betterado_release_definition_permissions.permissions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "release_definition_id"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ViewReleases", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.EditReleaseEnvironment", "notset"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleases", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CreateReleases", "deny"),
				),
			},
			{
				// idempotency check after first apply
				Config:             config1,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "release_definition_id"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "4"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.ViewReleases", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.EditReleaseEnvironment", "allow"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.DeleteReleases", "deny"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.CreateReleases", "notset"),
				),
			},
			{
				// idempotency check after update apply
				Config:             config2,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccReleaseDefinitionPermissions_AllWritablePermissions(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	tfNodeRoot := "betterado_release_definition_permissions.permissions"

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
	}
	config := hclReleaseDefinitionPermissions(fixture, allPerms)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNodeRoot, "project_id"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "principal"),
					resource.TestCheckResourceAttrSet(tfNodeRoot, "release_definition_id"),
					resource.TestCheckResourceAttr(tfNodeRoot, "permissions.%", "12"),
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
					captureReleaseDefinitionPermissionsEvidence(tfNodeRoot),
				),
			},
			{
				// idempotency check: plan after apply must be empty
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// captureReleaseDefinitionPermissionsEvidence writes .forge/live-evidence/acceptance-resource.json
// with a liveEvidence.url pointing to the real vsrm.dev.azure.com ACL REST endpoint.
// Best-effort: a capture failure never fails the test.
func captureReleaseDefinitionPermissionsEvidence(tfNode string) resource.TestCheckFunc {
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
		// Release definition ACL lives on the vsrm host, not dev.azure.com.
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

// hclReleaseDefinitionPermissions builds HCL for testing betterado_release_definition_permissions.
// It uses the shared fixture project (no inline betterado_project block), creates a minimal
// release definition with attribute syntax, looks up the Readers group, and applies the given
// permissions map to that group on the release definition.
func hclReleaseDefinitionPermissions(fixture SharedFixtureResult, permissions map[string]string) string {
	rootPermissions := datahelper.JoinMap(permissions, "=", "\n")
	releaseName := testutils.GenerateResourceName()

	return fmt.Sprintf(`
resource "betterado_release_definition" "release" {
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

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage (VS402877) — see resource_release_definition_test.go.
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

data "betterado_group" "tf-project-readers" {
  project_id = %[1]q
  name       = "Readers"
}

resource "betterado_release_definition_permissions" "permissions" {
  project_id            = %[1]q
  principal             = data.betterado_group.tf-project-readers.id
  release_definition_id = betterado_release_definition.release.id

  permissions = {
    %[3]s
  }
}
`, fixture.ProjectID, releaseName, rootPermissions)
}
