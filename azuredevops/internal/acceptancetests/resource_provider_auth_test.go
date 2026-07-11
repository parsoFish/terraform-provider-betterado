//go:build all || resource_provider_auth

package acceptancetests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/Azure/entrauth/aztfauth"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccAuthParity_CLIPath verifies that the framework Configure() path can
// authenticate using az CLI (non-PAT) credentials and successfully perform a
// live ADO data-source read. Requires TF_ACC=1 and az CLI logged in with a
// principal that can mint ADO-audience tokens.
//
// If az CLI cannot mint ADO-audience tokens (wrong tenant, not logged in, etc.)
// the test skips immediately and documents the fallback path.
func TestAccAuthParity_CLIPath(t *testing.T) {
	// AC1: skip without TF_ACC — hollow gate pattern.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live auth-parity acceptance test")
	}

	// Probe az CLI availability: can it mint an ADO-audience token?
	// Resource UUID 499b84ac-1321-427f-aa17-267ca6975798 is the Azure DevOps app.
	//nolint:gosec // intentional CLI probe — arguments are constant
	probeCmd := exec.Command("az", "account", "get-access-token",
		"--resource", "499b84ac-1321-427f-aa17-267ca6975798")
	if err := probeCmd.Run(); err != nil {
		t.Skipf("az CLI cannot mint ADO-audience tokens (AADSTS9002313 or not logged in) — "+
			"taking credential-construction fallback path; run TestAccAuthParity_CredentialConstruction instead. "+
			"probe error: %v", err)
	}

	// AC5: use t.Setenv instead of os.Unsetenv so AZDO_PERSONAL_ACCESS_TOKEN is
	// automatically restored after this test, preventing PAT-based tests running
	// in the same process from seeing a missing token.
	// MUST happen BEFORE provider init so the CLI auth path is exercised.
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")

	orgServiceURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	if orgServiceURL == "" {
		t.Skip("AZDO_ORG_SERVICE_URL not set; skipping live auth-parity acceptance test")
	}

	tfNode := "data.betterado_project.auth_parity"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclAuthParityCLIConfig(orgServiceURL, SharedFixtureProjectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", SharedFixtureProjectName),
					captureAuthParityEvidence(tfNode, orgServiceURL),
				),
			},
		},
	})
}

// TestAccAuthParity_CredentialConstruction is a unit-style credential
// construction proof that does NOT require live credentials or TF_ACC.
// It verifies that aztfauth.NewCredential can construct a credential object
// for each auth method variant without error, confirming that all credential
// paths are wired correctly in the provider.
//
// Coverage:
//   - ClientSecret: service principal with client secret
//   - OIDCToken: OIDC static token
//   - OIDCTokenFile: OIDC token read from a file path
//   - OIDCTokenRequest: OIDC token via Actions/ADO request URL
//   - CLI: Azure CLI credential
//   - MSI: Managed Service Identity
//
// Note: ClientCertificate is NOT tested here because aztfauth.NewCredential
// requires actual PEM/PFX certificate data at construction time — it calls
// getClientCert() eagerly, which returns "no client certificate available"
// for an empty ClientCertBase64/ClientCertPfxFile. An empty cert leaves the
// credential chain empty, causing entrauth.NewCredential to fail with
// "sources must contain at least one TokenCredential". Certificate auth
// construction is covered by auth_test.go (TestResolveFrameworkAuth_*)
// via a stub path instead.
func TestAccAuthParity_CredentialConstruction(t *testing.T) {
	tests := []struct {
		name string
		opts aztfauth.Option
	}{
		{
			name: "ClientSecret",
			opts: aztfauth.Option{
				TenantId:        "00000000-0000-0000-0000-000000000000",
				ClientId:        "00000000-0000-0000-0000-000000000001",
				ClientSecret:    "fake-secret",
				UseClientSecret: true,
			},
		},
		{
			name: "OIDCToken",
			opts: aztfauth.Option{
				TenantId:     "00000000-0000-0000-0000-000000000000",
				ClientId:     "00000000-0000-0000-0000-000000000001",
				UseOIDCToken: true,
				OIDCToken:    "fake-oidc-token",
			},
		},
		{
			name: "OIDCTokenFile",
			opts: aztfauth.Option{
				TenantId:         "00000000-0000-0000-0000-000000000000",
				ClientId:         "00000000-0000-0000-0000-000000000001",
				UseOIDCTokenFile: true,
				OIDCTokenFile:    "/dev/null",
			},
		},
		{
			name: "OIDCTokenRequest",
			opts: aztfauth.Option{
				TenantId:            "00000000-0000-0000-0000-000000000000",
				ClientId:            "00000000-0000-0000-0000-000000000001",
				UseOIDCTokenRequest: true,
				OIDCRequestToken:    "fake-request-token",
				OIDCRequestURL:      "https://example.com/oidc/token",
			},
		},
		{
			name: "CLI",
			opts: aztfauth.Option{
				UseAzureCLI: true,
			},
		},
		{
			name: "MSI",
			opts: aztfauth.Option{
				UseMSI: true,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := aztfauth.NewCredential(tc.opts)
			if err != nil {
				t.Errorf("aztfauth.NewCredential(%s): unexpected error: %v", tc.name, err)
			}
		})
	}
}

// hclAuthParityCLIConfig returns HCL that configures the provider with CLI auth
// (no personal_access_token) and reads the shared fixture project via data source.
func hclAuthParityCLIConfig(orgURL, projectName string) string {
	return fmt.Sprintf(`
provider "betterado" {
  org_service_url = %[1]q
  use_cli         = true
}

data "betterado_project" "auth_parity" {
  name = %[2]q
}
`, orgURL, projectName)
}

// captureAuthParityEvidence is a TestCheckFunc that captures live evidence after
// the CLI-auth data source read succeeds. Best-effort: never fails the test.
func captureAuthParityEvidence(tfNode, orgURL string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["id"]
		if projectID == "" {
			return nil
		}

		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		projectURL := fmt.Sprintf("%s/_apis/projects/%s?api-version=7.1", orgURL, projectID)

		_ = testutils.CaptureLiveEvidence("acceptance-auth-parity", projectURL, map[string]string{
			"project":     SharedFixtureProjectName,
			"auth_method": "cli",
			"status":      "read-back-ok",
		})
		return nil
	}
}
