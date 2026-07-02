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
func TestAccProjectFeatures_roundtrip(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_project_features.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: disable testplans and artifacts.
			{
				Config: hclProjectFeatureBasic(projectName, "disabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "features.testplans", "disabled"),
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "disabled"),
					captureProjectFeaturesEvidence(tfNode),
				),
				// Idempotency: re-plan after apply must produce no diff.
				ExpectNonEmptyPlan: false,
			},
			// Step 2: re-enable testplans, keep artifacts disabled.
			{
				Config: hclProjectFeatureBasic(projectName, "enabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "features.testplans", "enabled"),
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "disabled"),
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

func hclProjectFeatureBasic(name, testPlanState, artifactState string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_project_features" "test" {
  project_id = betterado_project.test.id
  features = {
    "testplans" = "%[2]s"
    "artifacts" = "%[3]s"
  }
}`, name, testPlanState, artifactState)
}
