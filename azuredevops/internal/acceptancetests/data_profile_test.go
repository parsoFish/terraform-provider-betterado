//go:build (all || data_profile) && !exclude_data_profile

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	profileapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/profile"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccDataProfile(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "betterado_profile" "me" { id = "me" }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betterado_profile.me", "display_name"),
					resource.TestCheckResourceAttrSet("data.betterado_profile.me", "email_address"),
					captureProfileEvidence(),
				),
			},
			{
				Config:             `data "betterado_profile" "me" { id = "me" }`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// captureProfileEvidence performs a real live API GET of the authenticated user's
// ADO profile and persists the response as forge demo live-evidence. Best-effort:
// a capture failure never fails the test.
func captureProfileEvidence() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort
		}

		id := "me"
		details := true
		coreAttrs := "Email,Avatar,DisplayName,ContactWithOffers"
		profile, err := clients.ProfileClient.GetProfile(clients.Ctx, profileapi.GetProfileArgs{
			Id:             &id,
			Details:        &details,
			CoreAttributes: &coreAttrs,
		})
		if err != nil || profile == nil {
			return nil
		}

		url := fmt.Sprintf(
			"https://app.vssps.visualstudio.com/_apis/profile/profiles/%s?details=true&api-version=3.0",
			id,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-profile", url, profile)
		return nil
	}
}
