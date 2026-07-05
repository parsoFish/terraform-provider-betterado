package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccClientConfig_LoadsCorrectProperties(t *testing.T) {
	tfNode := "data.betterado_client_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "betterado_client_config" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "organization_url"),
					captureClientConfigEvidence(t, tfNode),
				),
			},
		},
	})
}

// captureClientConfigEvidence captures live evidence of the client_config data source
// for the forge demo pipeline. Best-effort: never fails the test check.
//
// The fixture_project_id is resolved dynamically from the betterado-standing-demo
// project so that it appears in the capture honestly — not as a hardcoded constant.
func captureClientConfigEvidence(t *testing.T, tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil // resource not found — skip evidence capture
		}
		orgURL := rs.Primary.Attributes["organization_url"]
		if orgURL == "" {
			orgURL = os.Getenv("AZDO_ORG_SERVICE_URL")
		}
		if orgURL == "" {
			return nil
		}
		// Resolve the fixture project ID dynamically so the capture is tied
		// to the actual standing-demo project in the running org.
		fixtureProjectID := ResolveFixtureProjectID(t)
		apiURL := fmt.Sprintf("%s/_apis/connectionData?api-version=7.1", orgURL)
		attrs := map[string]string{
			"name":               rs.Primary.Attributes["name"],
			"status":             rs.Primary.Attributes["status"],
			"tenant_id":          rs.Primary.Attributes["tenant_id"],
			"owner_id":           rs.Primary.Attributes["owner_id"],
			"organization_url":   rs.Primary.Attributes["organization_url"],
			"fixture_project_id": fixtureProjectID,
		}
		_ = testutils.CaptureLiveEvidence("data-client-config", apiURL, attrs)
		return nil
	}
}
