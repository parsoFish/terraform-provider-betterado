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
// The test reuses the persistent standing-demo project (betterado-standing-demo)
// via SharedReleaseFixture — which resolves or creates the project via ADO API
// and returns its UUID. The project UUID is then injected directly into HCL so
// the test never relies on a data source lookup (which would fail if the project
// doesn't exist yet). This pattern matches TestAccMuxSdkv2Passthrough.
func TestAccProjectFeatures_roundtrip(t *testing.T) {
	// SharedReleaseFixture resolves or creates betterado-standing-demo and
	// returns its UUID. If TF_ACC is not set it calls t.Skip immediately.
	fixture := SharedReleaseFixture(t)
	projectID := fixture.ProjectID

	tfNode := "betterado_project_features.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			// Step 1: disable testplans and artifacts.
			{
				Config: hclProjectFeatureBasic(projectID, "disabled", "disabled"),
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
				Config: hclProjectFeatureBasic(projectID, "enabled", "disabled"),
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

// hclProjectFeatureBasic returns Terraform HCL that manages features on an
// existing project identified by its UUID. No data source or resource create
// is needed — the project ID is resolved by SharedReleaseFixture before the
// Terraform test lifecycle starts.
func hclProjectFeatureBasic(projectID, testPlanState, artifactState string) string {
	return fmt.Sprintf(`
resource "betterado_project_features" "test" {
  project_id = %[1]q
  features = {
    "testplans" = %[2]q
    "artifacts" = %[3]q
  }
}`, projectID, testPlanState, artifactState)
}
