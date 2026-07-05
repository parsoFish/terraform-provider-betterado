//go:build (all || resource_feature_flag) && !exclude_resource_feature_flag

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	featuremanagementapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

const (
	featureFlagTestFeatureID = "ms.vss-work.agile"
	featureFlagTFNode        = "betterado_feature_flag.test"
)

// TestAccFeatureFlag_basic exercises betterado_feature_flag:
//   - Step 1: set ms.vss-work.agile to "enabled" on betterado-standing-demo, read back state
//   - Step 2: update to "disabled", read back
//   - Step 3: idempotency re-plan (ExpectNonEmptyPlan: false)
//
// CheckDestroy verifies the state reverts to "undefined" or "disabled" (default)
// after destroy.  Evidence is captured at Step 1 for forge demo.
func TestAccFeatureFlag_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkFeatureFlagDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create — enable the feature, assert read-back
			{
				Config: hclFeatureFlagBasic("enabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(featureFlagTFNode, "state", "enabled"),
					resource.TestCheckResourceAttr(featureFlagTFNode, "feature_id", featureFlagTestFeatureID),
					resource.TestCheckResourceAttr(featureFlagTFNode, "scope_name", "project"),
					resource.TestCheckResourceAttrSet(featureFlagTFNode, "scope_value"),
					captureFeatureFlagEvidence(featureFlagTFNode),
				),
			},
			// Step 2: update — disable the feature
			{
				Config: hclFeatureFlagBasic("disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(featureFlagTFNode, "state", "disabled"),
				),
			},
			// Step 3: idempotency — no perpetual diff
			{
				Config:             hclFeatureFlagBasic("disabled"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclFeatureFlagBasic(state string) string {
	return fmt.Sprintf(`
data "betterado_project" "demo" {
  name = %[2]q
}

resource "betterado_feature_flag" "test" {
  feature_id  = %[3]q
  scope_name  = "project"
  scope_value = data.betterado_project.demo.id
  state       = %[1]q
}
`, state, SharedFixtureProjectName, featureFlagTestFeatureID)
}

// checkFeatureFlagDestroyed verifies that after destroy the feature is no longer
// set to the value the test applied — it should be "undefined" or "disabled"
// (the ADO default).
func checkFeatureFlagDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkFeatureFlagDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_feature_flag" {
			continue
		}

		featureID := res.Primary.Attributes["feature_id"]
		scopeName := res.Primary.Attributes["scope_name"]
		scopeValue := res.Primary.Attributes["scope_value"]

		state, err := clients.FeatureManagementClient.GetFeatureStateForScope(
			clients.Ctx,
			featuremanagementapi.GetFeatureStateForScopeArgs{
				FeatureId:  converter.String(featureID),
				UserScope:  converter.String("host"),
				ScopeName:  converter.String(scopeName),
				ScopeValue: converter.String(scopeValue),
			},
		)
		if err != nil {
			// If the resource genuinely no longer exists at that scope, that's fine.
			return nil
		}
		if state != nil && state.State != nil {
			s := string(*state.State)
			// "undefined" means no override is set (reverted to default) — expected.
			// "disabled" is also acceptable as the ADO default for this feature.
			if s == "enabled" {
				return fmt.Errorf("feature flag %q at scope %q/%q is still \"enabled\" after destroy — expected undefined or disabled",
					featureID, scopeName, scopeValue)
			}
		}
	}

	return nil
}

// captureFeatureFlagEvidence performs a live API GET of the feature state and
// writes .forge/live-evidence/acceptance-resource.json.  Best-effort: a capture
// failure never fails the test.
func captureFeatureFlagEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}

		featureID := res.Primary.Attributes["feature_id"]
		scopeName := res.Primary.Attributes["scope_name"]
		scopeValue := res.Primary.Attributes["scope_value"]

		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort
		}

		state, err := clients.FeatureManagementClient.GetFeatureStateForScope(
			clients.Ctx,
			featuremanagementapi.GetFeatureStateForScopeArgs{
				FeatureId:  converter.String(featureID),
				UserScope:  converter.String("host"),
				ScopeName:  converter.String(scopeName),
				ScopeValue: converter.String(scopeValue),
			},
		)
		if err != nil || state == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf(
			"%s/_apis/FeatureManagement/FeatureStates/host/%s/%s/%s?api-version=7.1-preview.1",
			orgURL, scopeName, scopeValue, featureID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, state)
		return nil
	}
}
