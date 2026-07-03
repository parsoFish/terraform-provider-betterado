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
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: `data "betterado_client_config" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "status"),
					resource.TestCheckResourceAttrSet(tfNode, "tenant_id"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
					resource.TestCheckResourceAttr(tfNode, "organization_url", os.Getenv("AZDO_ORG_SERVICE_URL")),
					captureClientConfigEvidence(tfNode),
				),
			},
		},
	})
}

// captureClientConfigEvidence captures live evidence of the client_config data source
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureClientConfigEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil // resource not found — skip evidence capture
		}
		orgURL := rs.Primary.Attributes["organization_url"]
		if orgURL == "" {
			orgURL = os.Getenv("AZDO_ORG_SERVICE_URL")
		}
		apiURL := fmt.Sprintf("%s/_apis/connectionData?api-version=7.1", orgURL)
		attrs := map[string]string{
			"name":             rs.Primary.Attributes["name"],
			"status":           rs.Primary.Attributes["status"],
			"tenant_id":        rs.Primary.Attributes["tenant_id"],
			"owner_id":         rs.Primary.Attributes["owner_id"],
			"organization_url": rs.Primary.Attributes["organization_url"],
		}
		_ = testutils.CaptureLiveEvidence("data-client-config", apiURL, attrs)
		return nil
	}
}
