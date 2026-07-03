package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccBranchPolicyBuildValidation_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	name := testutils.GenerateResourceName()
	buildValidationTfNode := "betterado_branch_policy_build_validation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclBuildValidationBasic(projectID, name, true, true, "build validation", 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(buildValidationTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.filename_patterns.#", "3"),
				),
			}, {
				Config: hclBuildValidationBasic(projectID, name, false, false, "build validation rename", 720),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(buildValidationTfNode, "enabled", "false"),
					resource.TestCheckResourceAttr(buildValidationTfNode, "settings.0.filename_patterns.#", "3"),
				),
			}, {
				ResourceName:      buildValidationTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(buildValidationTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclBuildValidationBasic(projectID, name string, enabled, blocking bool, displayName string, validDuration int) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "test" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_build_definition" "test" {
  project_id      = %[1]q
  name            = "%[2]s-build"
  agent_pool_name = "Azure Pipelines"
  path            = "\\"

  repository {
    repo_type   = "TfsGit"
    repo_id     = betterado_git_repository.test.id
    branch_name = "main"
    yml_path    = "path/to/yaml"
  }
}

resource "betterado_branch_policy_build_validation" "test" {
  project_id = %[1]q
  enabled    = %[3]t
  blocking   = %[4]t
  settings {
    display_name        = "%[5]s"
    valid_duration      = %[6]d
    build_definition_id = betterado_build_definition.test.id
    filename_patterns = [
      "/WebApp/*",
      "!/WebApp/Tests/*",
      "*.cs"
    ]
    scope {
      repository_id  = betterado_git_repository.test.id
      repository_ref = "refs/heads/release"
      match_type     = "Exact"
    }
  }
}`, projectID, name, enabled, blocking, displayName, validDuration)
}
