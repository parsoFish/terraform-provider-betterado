package acceptancetests

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccProviderAuth(t *testing.T) {
	if ok := os.Getenv("AZDO_ORG_SERVICE_URL"); ok == "" {
		t.Skip("Skipping as `AZDO_ORG_SERVICE_URL` is not specified")
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: providerAuthConfig(),
			},
		},
	})
}

func providerAuthConfig() string {
	return `
data "betterado_projects" "test" {
  name  = "Test Project"
  state = "wellFormed"
}`
}
