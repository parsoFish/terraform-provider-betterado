//go:build (all || data_accounts) && !exclude_data_accounts

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	accountsapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts"
	profileapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/profile"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccDataAccounts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "betterado_accounts" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.betterado_accounts.test", "accounts.#"),
					captureAccountsEvidence(),
				),
			},
			{
				Config:             `data "betterado_accounts" "test" {}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// captureAccountsEvidence performs a real live API GET of the accounts list
// and persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test.
func captureAccountsEvidence() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}

		// Fetch the authenticated user's profile to get their UUID (memberId).
		id := "me"
		details := true
		coreAttrs := "Email,DisplayName"
		profile, err := clients.ProfileClient.GetProfile(clients.Ctx, profileapi.GetProfileArgs{
			Id:             &id,
			Details:        &details,
			CoreAttributes: &coreAttrs,
		})
		if err != nil || profile == nil || profile.Id == nil {
			// Fall back to listing all accounts without a memberId filter.
			allAccounts, err2 := clients.AccountsClient.GetAccounts(clients.Ctx, accountsapi.GetAccountsArgs{})
			if err2 != nil || allAccounts == nil {
				return nil
			}
			url := "https://app.vssps.visualstudio.com/_apis/accounts?api-version=7.1"
			_ = testutils.CaptureLiveEvidence("acceptance-resource", url, allAccounts)
			return nil
		}

		memberID := *profile.Id
		accounts, err := clients.AccountsClient.GetAccounts(clients.Ctx, accountsapi.GetAccountsArgs{
			MemberId: &memberID,
		})
		if err != nil || accounts == nil {
			return nil
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		_ = orgURL // used for context; the accounts endpoint is vssps
		url := fmt.Sprintf(
			"https://app.vssps.visualstudio.com/_apis/accounts?memberId=%s&api-version=7.1",
			memberID.String(),
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, accounts)
		return nil
	}
}
