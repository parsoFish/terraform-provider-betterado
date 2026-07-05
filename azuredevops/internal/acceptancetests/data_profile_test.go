//go:build (all || data_profile) && !exclude_data_profile

package acceptancetests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccDataProfile(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
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
//
// Uses direct HTTP (bypassing SDK location-service discovery) so that org-scoped
// PATs work. Tries the org-specific VSSPS URL first (works for org-scoped PATs),
// then falls back to the global VSSPS URL (works for global-scope PATs).
func captureProfileEvidence() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		orgServiceURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if pat == "" {
			return nil // best-effort
		}
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("_:"+pat))
		ctx := context.Background()

		orgName := extractOrgNameFromURL(orgServiceURL)

		// Build the profile URL. Use org-specific VSSPS for org-scoped PATs.
		var profileURL string
		if orgName != "" {
			profileURL = fmt.Sprintf(
				"https://vssps.dev.azure.com/%s/_apis/profile/profiles/me?api-version=7.1-preview.3&details=true&coreAttributes=Email,Avatar,DisplayName",
				orgName,
			)
		} else {
			profileURL = "https://app.vssps.visualstudio.com/_apis/profile/profiles/me?api-version=7.1-preview.3&details=true&coreAttributes=Email,Avatar,DisplayName"
		}

		profile := fetchProfileForEvidence(ctx, profileURL, basicAuth)
		if profile == nil && orgName != "" {
			// Fallback: try global VSSPS.
			globalURL := "https://app.vssps.visualstudio.com/_apis/profile/profiles/me?api-version=7.1-preview.3&details=true&coreAttributes=Email,Avatar,DisplayName"
			profile = fetchProfileForEvidence(ctx, globalURL, basicAuth)
			if profile != nil {
				profileURL = globalURL
			}
		}
		if profile == nil {
			return nil // nothing to capture
		}

		// Evidence URL must contain _apis/profile/profiles and vssps host.
		_ = testutils.CaptureLiveEvidence("acceptance-resource-profile", profileURL, profile)
		return nil
	}
}

// fetchProfileForEvidence calls the VSSPS profile endpoint directly (no SDK)
// and returns the parsed profile object on success, nil on any error.
func fetchProfileForEvidence(ctx context.Context, profileURL, basicAuth string) interface{} {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil
	}
	return v
}
