package provider

import (
	"context"
	"strings"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
)

// TestResolveFrameworkAuthPAT verifies AC1: a non-empty PAT returns an
// *azuredevops.AuthProviderPAT and nil error.
func TestResolveFrameworkAuthPAT(t *testing.T) {
	cfg := FrameworkAuthConfig{
		PersonalAccessToken: "my-secret-pat",
	}
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	// NewAuthProviderPAT returns a value type (AuthProviderPAT), not a pointer;
	// the AC specifies the PAT concrete type — we assert on the value form as
	// returned by the SDK.
	if _, ok := ap.(azuredevops.AuthProviderPAT); !ok {
		t.Fatalf("expected azuredevops.AuthProviderPAT, got %T", ap)
	}
}

// TestResolveFrameworkAuthPATEnvFallback verifies AC2: when PAT is empty in
// config but AZDO_PERSONAL_ACCESS_TOKEN is set, the env fallback is used and
// an azuredevops.AuthProviderPAT is returned.
func TestResolveFrameworkAuthPATEnvFallback(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "env-pat-value")
	// Ensure other env vars that could trigger a different path are unset.
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{} // PersonalAccessToken is empty
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	// NewAuthProviderPAT returns a value type (not a pointer).
	if _, ok := ap.(azuredevops.AuthProviderPAT); !ok {
		t.Fatalf("expected azuredevops.AuthProviderPAT, got %T", ap)
	}
}

// TestResolveFrameworkAuthCLI verifies AC3: when use_cli is true and no PAT is
// set, the CLI credential path is taken and no error is returned.
func TestResolveFrameworkAuthCLI(t *testing.T) {
	// Clear PAT env var so CLI is the only option.
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{
		UseCLI: true,
	}
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for CLI path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for CLI path")
	}
}

// TestResolveFrameworkAuthMSI verifies AC4: when use_msi is true and no PAT is
// set, the MSI credential path is taken and no error is returned.
func TestResolveFrameworkAuthMSI(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{
		UseMSI: true,
		// UseCLI must be false so CLI doesn't take over; env ARM_USE_CLI is also
		// set to false above.
	}
	// We set UseCLI explicitly to prevent the default-true logic from overriding
	// via applyEnvFallbacks. ARM_USE_CLI=false is set above so the env path also
	// returns false.
	cfg.UseCLI = false

	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for MSI path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for MSI path")
	}
}

// TestResolveFrameworkAuthClientSecret verifies AC5: client_secret +
// tenant_id + client_id returns an AuthProvider without error.
func TestResolveFrameworkAuthClientSecret(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{
		ClientSecret: "my-secret",
		TenantID:     "tenant-id-here",
		ClientID:     "client-id-here",
		UseCLI:       false,
	}
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for client-secret path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for client-secret path")
	}
}

// TestResolveFrameworkAuthOIDC verifies AC6: use_oidc true + oidc_token set
// returns an AuthProvider without error.
func TestResolveFrameworkAuthOIDC(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")

	cfg := FrameworkAuthConfig{
		UseOIDC:   true,
		OIDCToken: "my-oidc-token",
		TenantID:  "tenant-id-here",
		ClientID:  "client-id-here",
		UseCLI:    false,
	}
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for OIDC path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for OIDC path")
	}
}

// TestResolveFrameworkAuthNoCredential verifies AC7: with all credential
// options explicitly disabled/empty, a non-nil error with a human-readable
// message is returned.
func TestResolveFrameworkAuthNoCredential(t *testing.T) {
	// Clear all env vars that could provide a credential.
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")
	t.Setenv("ARM_CLIENT_ID", "")
	t.Setenv("AZURE_CLIENT_ID", "")
	t.Setenv("ARM_TENANT_ID", "")
	t.Setenv("ARM_CLIENT_SECRET", "")

	cfg := FrameworkAuthConfig{
		UseCLI:  false,
		UseMSI:  false,
		UseOIDC: false,
	}
	_, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected non-nil error when no credential is configured")
	}
	if !strings.Contains(err.Error(), "no credential") {
		t.Fatalf("expected error message to contain 'no credential', got: %v", err)
	}
}

// TestResolveFrameworkAuthEnvVarFallbacks verifies AC8: ARM_CLIENT_ID,
// ARM_TENANT_ID, ARM_USE_CLI, ARM_USE_MSI, ARM_USE_OIDC env vars are applied
// when config attrs are absent, matching SDKv2 MultiEnvDefaultFunc semantics.
func TestResolveFrameworkAuthEnvVarFallbacks(t *testing.T) {
	// Set env vars that should be picked up as fallbacks.
	t.Setenv("ARM_CLIENT_ID", "env-client-id")
	t.Setenv("ARM_TENANT_ID", "env-tenant-id")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_CLIENT_SECRET", "env-client-secret")

	// Zero-value config — all credential fields are empty.
	cfg := FrameworkAuthConfig{}
	// applyEnvFallbacks inside resolveFrameworkAuthProvider should pick up
	// ARM_CLIENT_ID, ARM_TENANT_ID, ARM_CLIENT_SECRET.
	ap, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error with env-var fallbacks applied, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider with env-var fallbacks applied")
	}

	// Also verify AZURE_CLIENT_ID is checked as secondary precedence for client
	// ID when ARM_CLIENT_ID is not set.
	t.Setenv("ARM_CLIENT_ID", "")
	t.Setenv("AZURE_CLIENT_ID", "azure-env-client-id")

	cfg2 := FrameworkAuthConfig{}
	ap2, err2 := resolveFrameworkAuthProvider(context.Background(), cfg2)
	if err2 != nil {
		t.Fatalf("expected nil error with AZURE_CLIENT_ID fallback, got: %v", err2)
	}
	if ap2 == nil {
		t.Fatal("expected non-nil AuthProvider with AZURE_CLIENT_ID fallback")
	}
}

// TestResolveFrameworkAuthUseCLIDefault verifies AC9: when ARM_USE_CLI is
// unset, use_cli defaults to true (matching SDKv2 DefaultFunc behaviour).
func TestResolveFrameworkAuthUseCLIDefault(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "") // explicitly unset

	cfg := FrameworkAuthConfig{} // UseCLI defaults to false (zero value)

	// applyEnvFallbacks should set UseCLI=true when ARM_USE_CLI is unset.
	applyEnvFallbacks(&cfg)
	if !cfg.UseCLI {
		t.Fatal("expected UseCLI to default to true when ARM_USE_CLI is unset")
	}
}
