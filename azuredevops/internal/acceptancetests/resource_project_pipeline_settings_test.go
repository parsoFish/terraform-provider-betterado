package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccProjectPipelineSettings_Enabled(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_project_pipeline_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPipelineSettings(projectName, false, false, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "enforce_job_scope", "false"),
					resource.TestCheckResourceAttr(tfNode, "enforce_referenced_repo_scoped_token", "false"),
					resource.TestCheckResourceAttr(tfNode, "enforce_settable_var", "false"),
					resource.TestCheckResourceAttr(tfNode, "publish_pipeline_metadata", "false"),
					resource.TestCheckResourceAttr(tfNode, "status_badges_are_private", "false"),
					resource.TestCheckResourceAttr(tfNode, "enforce_job_scope_for_release", "false"),
					captureProjectPipelineSettingsEvidence(tfNode),
				),
			},
			{
				Config: hclProjectPipelineSettings(projectName, true, true, true, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "enforce_job_scope", "true"),
					resource.TestCheckResourceAttr(tfNode, "enforce_referenced_repo_scoped_token", "true"),
					resource.TestCheckResourceAttr(tfNode, "enforce_settable_var", "true"),
					resource.TestCheckResourceAttr(tfNode, "publish_pipeline_metadata", "true"),
					resource.TestCheckResourceAttr(tfNode, "status_badges_are_private", "true"),
					resource.TestCheckResourceAttr(tfNode, "enforce_job_scope_for_release", "true"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// captureProjectPipelineSettingsEvidence captures live evidence of project_pipeline_settings
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureProjectPipelineSettingsEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if orgURL == "" {
			return nil
		}
		projectID := rs.Primary.Attributes["project_id"]
		apiURL := fmt.Sprintf("%s/%s/_apis/build/generalsettings?api-version=7.1", orgURL, projectID)
		attrs := map[string]string{
			"project_id":                           projectID,
			"enforce_job_scope":                    rs.Primary.Attributes["enforce_job_scope"],
			"enforce_referenced_repo_scoped_token": rs.Primary.Attributes["enforce_referenced_repo_scoped_token"],
			"enforce_settable_var":                 rs.Primary.Attributes["enforce_settable_var"],
			"publish_pipeline_metadata":            rs.Primary.Attributes["publish_pipeline_metadata"],
			"status_badges_are_private":            rs.Primary.Attributes["status_badges_are_private"],
			"enforce_job_scope_for_release":        rs.Primary.Attributes["enforce_job_scope_for_release"],
		}
		_ = testutils.CaptureLiveEvidence("project-pipeline-settings", apiURL, attrs)
		return nil
	}
}

func hclProjectPipelineSettings(projectName string, enforceJobAuthScope, enforceReferencedRepoScopedToken, enforceSettableVar, publishPipelineMetadata, statusBadgesArePrivate, enforceJobAuthScopeForReleases bool) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_project_pipeline_settings" "test" {
  project_id                           = betterado_project.test.id
  enforce_job_scope                    = %t
  enforce_referenced_repo_scoped_token = %t
  enforce_settable_var                 = %t
  publish_pipeline_metadata            = %t
  status_badges_are_private            = %t
  enforce_job_scope_for_release        = %t
}
`, projectName, enforceJobAuthScope, enforceReferencedRepoScopedToken, enforceSettableVar, publishPipelineMetadata, statusBadgesArePrivate, enforceJobAuthScopeForReleases)
}
