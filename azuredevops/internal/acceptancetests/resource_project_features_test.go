//go:build all

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/core"
)

// TestAccProjectFeatures_roundtrip exercises the full lifecycle of
// betterado_project_features on the framework provider:
//
//	apply (disable features) → read-back → idempotency re-plan → update (re-enable) → destroy
//
// It uses GetMuxedProviderFactories so both betterado_project (framework) and
// betterado_project_features (framework) are available in the same test.
//
// The test resolves an existing ADO project via smokeResolveProject — which checks
// AZDO_TEST_EXISTING_PROJECT first, then auto-discovers the first wellFormed project
// via GetProjects. This avoids creating a project (which fails when the org is at
// the 1000-project limit). The project UUID is injected directly into HCL.
//
// NOTE: testplans is intentionally excluded from this test because it requires a
// Basic+TestPlans or Visual Studio subscription — the ADO API returns HTTP 200 but
// silently leaves the state unchanged when the license is absent. artifacts and
// boards are license-free and reliably toggle on any project type.
func TestAccProjectFeatures_roundtrip(t *testing.T) {
	// smokeResolveProject resolves an existing ADO project without creating one.
	// Calls testutils.PreCheck which skips immediately if TF_ACC is not set.
	projectID, _ := smokeResolveProject(t)

	tfNode := "betterado_project_features.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: disable artifacts and boards (license-free features).
			{
				Config: hclProjectFeatureBasic(projectID, "disabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "disabled"),
					resource.TestCheckResourceAttr(tfNode, "features.boards", "disabled"),
					captureProjectFeaturesEvidence(tfNode),
				),
				// Idempotency: re-plan after apply must produce no diff.
				ExpectNonEmptyPlan: false,
			},
			// Step 2: re-enable artifacts, keep boards disabled.
			{
				Config: hclProjectFeatureBasic(projectID, "enabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "enabled"),
					resource.TestCheckResourceAttr(tfNode, "features.boards", "disabled"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// captureProjectFeaturesEvidence reads back the feature states via the ADO REST
// API and records live evidence for the forge demo pipeline. Best-effort: never
// fails the test check.
func captureProjectFeaturesEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("resource %s not found in state", tfNode)
		}
		projectID := rs.Primary.ID

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		if orgURL == "" || pat == "" {
			return nil // not in a live environment — skip evidence capture
		}

		// Build client. Best-effort: ignore errors so evidence capture never
		// fails an assertion.
		clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err == nil {
			states, sErr := core.GetProjectFeatureStatesForEvidence(clients, projectID)
			if sErr == nil {
				apiURL := fmt.Sprintf(
					"%s/_apis/FeatureManagement/FeatureStatesForScope/host/project/%s?api-version=7.1",
					orgURL, projectID,
				)
				_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, states)
			}
		}

		return nil
	}
}

// hclProjectFeatureBasic returns Terraform HCL that manages two license-free
// features (artifacts, boards) on an existing project identified by its UUID.
// testplans is intentionally excluded because it requires a paid license and
// the ADO API silently ignores enable requests when that license is absent,
// which would produce a "Provider produced inconsistent result after apply" error.
//
// artifactState is applied to "artifacts"; boardState is applied to "boards".
func hclProjectFeatureBasic(projectID, artifactState, boardState string) string {
	return fmt.Sprintf(`
resource "betterado_project_features" "test" {
  project_id = %[1]q
  features = {
    "artifacts" = %[2]q
    "boards"    = %[3]q
  }
}`, projectID, artifactState, boardState)
}
