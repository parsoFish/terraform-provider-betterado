package provider

import (
	"context"
	"strings"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
)

// boolPtr is a test helper that returns a pointer to the provided bool value.
func boolPtr(b bool) *bool { return &b }

// TestResolveFrameworkAuthPAT verifies that a non-empty PAT returns an
// *azuredevops.AuthProviderPAT and nil error.
func TestResolveFrameworkAuthPAT(t *testing.T) {
	cfg := FrameworkAuthConfig{
		PersonalAccessToken: "my-secret-pat",
	}
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
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

// TestResolveFrameworkAuthPATEnvFallback verifies that when PAT is empty in
// config but AZDO_PERSONAL_ACCESS_TOKEN is set, the env fallback is used and
// an azuredevops.AuthProviderPAT is returned.
func TestResolveFrameworkAuthPATEnvFallback(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "env-pat-value")
	// Ensure other env vars that could trigger a different path are unset.
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{} // PersonalAccessToken is empty
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	// NewAuthProviderPAT returns a value type (not a pointer).
	if _, ok := ap.(azuredevops.AuthProviderPAT); !ok {
		t.Fatalf("expected azuredevops.AuthProviderPAT, got %T", ap)
	}
}

// TestResolveFrameworkAuthCLI verifies that when use_cli is true and no PAT is
// set, the CLI credential path is taken and no error is returned.
func TestResolveFrameworkAuthCLI(t *testing.T) {
	// Clear PAT env var so CLI is the only option.
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_MSI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{
		UseCLI: boolPtr(true),
	}
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for CLI path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for CLI path")
	}
}

// TestResolveFrameworkAuthMSI verifies that when use_msi is true and no PAT is
// set, the MSI credential path is taken and no error is returned.
func TestResolveFrameworkAuthMSI(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "false")
	t.Setenv("ARM_USE_OIDC", "false")

	cfg := FrameworkAuthConfig{
		UseMSI: true,
		UseCLI: boolPtr(false),
	}

	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for MSI path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for MSI path")
	}
}

// TestResolveFrameworkAuthClientSecret verifies that client_secret +
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
		UseCLI:       boolPtr(false),
	}
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for client-secret path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for client-secret path")
	}
}

// TestResolveFrameworkAuthOIDC verifies that use_oidc true + oidc_token set
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
		UseCLI:    boolPtr(false),
	}
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected nil error for OIDC path, got: %v", err)
	}
	if ap == nil {
		t.Fatal("expected non-nil AuthProvider for OIDC path")
	}
}

// TestResolveFrameworkAuthNoCredential verifies that with all credential
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
		UseCLI:  boolPtr(false),
		UseMSI:  false,
		UseOIDC: false,
	}
	_, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected non-nil error when no credential is configured")
	}
	if !strings.Contains(err.Error(), "no credential") {
		t.Fatalf("expected error message to contain 'no credential', got: %v", err)
	}
}

// TestResolveFrameworkAuthEnvVarFallbacks verifies that ARM_CLIENT_ID,
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
	ap, _, err := resolveFrameworkAuthProvider(context.Background(), cfg)
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
	ap2, _, err2 := resolveFrameworkAuthProvider(context.Background(), cfg2)
	if err2 != nil {
		t.Fatalf("expected nil error with AZURE_CLIENT_ID fallback, got: %v", err2)
	}
	if ap2 == nil {
		t.Fatal("expected non-nil AuthProvider with AZURE_CLIENT_ID fallback")
	}
}

// TestResolveFrameworkAuthUseCLIDefault verifies AC9 / that when ARM_USE_CLI
// is unset, use_cli defaults to true (matching SDKv2 DefaultFunc behaviour).
func TestResolveFrameworkAuthUseCLIDefault(t *testing.T) {
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")
	t.Setenv("ARM_USE_CLI", "") // explicitly unset

	cfg := FrameworkAuthConfig{} // UseCLI is nil (not set)

	// applyEnvFallbacks should set UseCLI=true when ARM_USE_CLI is unset.
	applyEnvFallbacks(&cfg)
	if cfg.UseCLI == nil || !*cfg.UseCLI {
		t.Fatal("expected UseCLI to default to true when ARM_USE_CLI is unset")
	}
}

// TestApplyEnvFallbacks_ExplicitFalseUseCLI covers AC1: when use_cli is
// explicitly set to false in the config (non-nil *bool = false) AND ARM_USE_CLI
// is unset in the environment, CLI auth must remain disabled.
// The old bool implementation collapsed null → false → re-defaulted to true,
// breaking this security-relevant invariant.
func TestApplyEnvFallbacks_ExplicitFalseUseCLI(t *testing.T) {
	// Ensure ARM_USE_CLI is absent from the environment.
	t.Setenv("ARM_USE_CLI", "")

	cfg := FrameworkAuthConfig{
		UseCLI: boolPtr(false), // explicit false — user wrote use_cli = false in HCL
	}

	applyEnvFallbacks(&cfg)

	if cfg.UseCLI == nil {
		t.Fatal("UseCLI must not be nil after applyEnvFallbacks")
	}
	if *cfg.UseCLI {
		t.Fatal("AC1 violated: explicit use_cli=false was overridden to true by applyEnvFallbacks; " +
			"CLI auth must remain disabled when use_cli is explicitly false in HCL and ARM_USE_CLI is unset")
	}
}

// TestApplyEnvFallbacks_AuxiliaryTenantIDsTrimSpace covers AC2: when
// ARM_AUXILIARY_TENANT_IDS contains whitespace-padded elements, each element
// must be trimmed before storing.
func TestApplyEnvFallbacks_AuxiliaryTenantIDsTrimSpace(t *testing.T) {
	t.Setenv("ARM_AUXILIARY_TENANT_IDS", "tenant-a, tenant-b")

	cfg := FrameworkAuthConfig{}
	applyEnvFallbacks(&cfg)

	if len(cfg.AuxiliaryTenantIDs) != 2 {
		t.Fatalf("expected 2 auxiliary tenant IDs, got %d: %v", len(cfg.AuxiliaryTenantIDs), cfg.AuxiliaryTenantIDs)
	}
	if cfg.AuxiliaryTenantIDs[0] != "tenant-a" {
		t.Errorf("expected first element 'tenant-a', got %q", cfg.AuxiliaryTenantIDs[0])
	}
	if cfg.AuxiliaryTenantIDs[1] != "tenant-b" {
		t.Errorf("expected second element 'tenant-b', got %q", cfg.AuxiliaryTenantIDs[1])
	}
}

// TestParseBoolEnv_MalformedDiagnostic covers AC3: a malformed boolean env var
// (e.g. ARM_USE_CLI=flase) must produce a non-empty warning string naming the
// variable and value, instead of silently treating it as unset.
func TestParseBoolEnv_MalformedDiagnostic(t *testing.T) {
	t.Setenv("ARM_USE_CLI", "flase") // intentional typo

	val, set, warning := parseBoolEnv("ARM_USE_CLI")

	if set {
		t.Errorf("expected set=false for malformed value, got set=true (val=%v)", val)
	}
	if warning == "" {
		t.Fatal("AC3 violated: parseBoolEnv must return a non-empty warning for malformed boolean values")
	}
	if !strings.Contains(warning, "ARM_USE_CLI") {
		t.Errorf("warning must name the env var 'ARM_USE_CLI', got: %q", warning)
	}
	if !strings.Contains(warning, "flase") {
		t.Errorf("warning must include the bad value 'flase', got: %q", warning)
	}
}

// TestApplyEnvFallbacks_MalformedBoolWarning covers AC3 at the applyEnvFallbacks
// level: a malformed ARM_USE_CLI value must propagate a warning in the returned
// slice so callers (Configure()) can surface it as a provider diagnostic.
func TestApplyEnvFallbacks_MalformedBoolWarning(t *testing.T) {
	t.Setenv("ARM_USE_CLI", "flase") // intentional typo
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")

	cfg := FrameworkAuthConfig{}
	warnings := applyEnvFallbacks(&cfg)

	found := false
	for _, w := range warnings {
		if strings.Contains(w, "ARM_USE_CLI") && strings.Contains(w, "flase") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AC3 violated: applyEnvFallbacks must return a warning for ARM_USE_CLI=flase, got warnings: %v", warnings)
	}
}
