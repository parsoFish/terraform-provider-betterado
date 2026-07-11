package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/entrauth/aztfauth"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
)

// FrameworkAuthConfig holds all 19 credential attributes decoded from the
// Terraform provider configuration block. Zero-value fields trigger env-var
// fallback resolution inside resolveFrameworkAuthProvider.
//
// UseCLI is a pointer so that the explicit-false case (use_cli = false in HCL)
// is distinguished from the null/absent case (attribute not set): null means
// "apply env-var fallback then default to true"; explicit false means "disable
// CLI auth regardless of environment".
type FrameworkAuthConfig struct {
	OrgServiceURL                string
	PersonalAccessToken          string
	ClientID                     string
	ClientIDFilePath             string
	TenantID                     string
	AuxiliaryTenantIDs           []string
	ClientCertificatePath        string
	ClientCertificate            string
	ClientCertificatePassword    string
	ClientSecret                 string
	ClientSecretPath             string
	OIDCRequestToken             string
	OIDCRequestURL               string
	OIDCToken                    string
	OIDCTokenFilePath            string
	OIDCAzureServiceConnectionID string
	UseOIDC                      bool
	// UseCLI is a *bool so null (absent) is distinguishable from explicit false.
	// nil  → not set in HCL; applyEnvFallbacks will check ARM_USE_CLI, then default true.
	// &false → explicitly set to false in HCL; CLI auth is disabled unconditionally.
	// &true  → explicitly set to true in HCL; CLI auth is enabled.
	UseCLI *bool
	UseMSI bool
}

// multiEnvLookup returns the value of the first env var in vars that is
// non-empty, mimicking SDKv2 MultiEnvDefaultFunc semantics.
func multiEnvLookup(vars ...string) string {
	for _, v := range vars {
		if val := os.Getenv(v); val != "" {
			return val
		}
	}
	return ""
}

// parseBoolEnv parses an environment variable as a boolean.
//   - Returns (false, false, "") when the env var is unset or empty.
//   - Returns (val, true, "") when set and valid.
//   - Returns (false, false, warning) when set but not a valid boolean; the
//     warning names the variable and value so callers can surface a diagnostic.
func parseBoolEnv(name string) (val bool, set bool, warning string) {
	raw := os.Getenv(name)
	if raw == "" {
		return false, false, ""
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false, fmt.Sprintf(
			"ignoring malformed boolean env var %s=%q: expected true/false/1/0 (env var treated as unset)",
			name, raw,
		)
	}
	return b, true, ""
}

// applyEnvFallbacks fills in zero-value fields on cfg using the canonical
// environment variables, matching SDKv2 MultiEnvDefaultFunc / EnvDefaultFunc
// semantics (ARM_* / AZURE_* precedence).
//
// It returns a slice of human-readable warning strings for any env vars that
// were set but could not be parsed (e.g. ARM_USE_CLI=flase). Callers should
// surface these as provider diagnostics.
//
// UseCLI behaviour:
//   - If cfg.UseCLI is non-nil (explicit HCL value), it is left unchanged.
//   - If cfg.UseCLI is nil, ARM_USE_CLI is checked; if unset, defaults to true
//     (matching SDKv2's EnvDefaultFunc("ARM_USE_CLI", true) behaviour).
func applyEnvFallbacks(cfg *FrameworkAuthConfig) []string {
	var warnings []string

	if cfg.PersonalAccessToken == "" {
		cfg.PersonalAccessToken = multiEnvLookup("AZDO_PERSONAL_ACCESS_TOKEN")
	}
	if cfg.ClientID == "" {
		cfg.ClientID = multiEnvLookup("ARM_CLIENT_ID", "AZURE_CLIENT_ID")
	}
	if cfg.ClientIDFilePath == "" {
		cfg.ClientIDFilePath = multiEnvLookup("ARM_CLIENT_ID_FILE_PATH")
	}
	if cfg.TenantID == "" {
		cfg.TenantID = multiEnvLookup("ARM_TENANT_ID")
	}
	if len(cfg.AuxiliaryTenantIDs) == 0 {
		if raw := multiEnvLookup("ARM_AUXILIARY_TENANT_IDS"); raw != "" {
			parts := strings.Split(raw, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}
			cfg.AuxiliaryTenantIDs = parts
		}
	}
	if cfg.ClientCertificatePath == "" {
		cfg.ClientCertificatePath = multiEnvLookup("ARM_CLIENT_CERTIFICATE_PATH")
	}
	if cfg.ClientCertificate == "" {
		cfg.ClientCertificate = multiEnvLookup("ARM_CLIENT_CERTIFICATE")
	}
	if cfg.ClientCertificatePassword == "" {
		cfg.ClientCertificatePassword = multiEnvLookup("ARM_CLIENT_CERTIFICATE_PASSWORD")
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret = multiEnvLookup("ARM_CLIENT_SECRET")
	}
	if cfg.ClientSecretPath == "" {
		cfg.ClientSecretPath = multiEnvLookup("ARM_CLIENT_SECRET_PATH", "ARM_CLIENT_SECRET_FILE_PATH")
	}
	if cfg.OIDCRequestToken == "" {
		cfg.OIDCRequestToken = multiEnvLookup("ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN")
	}
	if cfg.OIDCRequestURL == "" {
		cfg.OIDCRequestURL = multiEnvLookup("ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL", "SYSTEM_OIDCREQUESTURI")
	}
	if cfg.OIDCToken == "" {
		cfg.OIDCToken = multiEnvLookup("ARM_OIDC_TOKEN")
	}
	if cfg.OIDCTokenFilePath == "" {
		cfg.OIDCTokenFilePath = multiEnvLookup("ARM_OIDC_TOKEN_FILE_PATH", "AZURE_FEDERATED_TOKEN_FILE")
	}
	if cfg.OIDCAzureServiceConnectionID == "" {
		cfg.OIDCAzureServiceConnectionID = multiEnvLookup(
			"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID",
			"ARM_OIDC_AZURE_SERVICE_CONNECTION_ID",
			"AZURESUBSCRIPTION_SERVICE_CONNECTION_ID",
		)
	}

	// Boolean env-var fallbacks. The caller may have already set these from the
	// HCL config; only override if the env var is explicitly set.
	if !cfg.UseOIDC {
		if v, set, w := parseBoolEnv("ARM_USE_OIDC"); set {
			cfg.UseOIDC = v
		} else if w != "" {
			warnings = append(warnings, w)
		}
	}
	if !cfg.UseMSI {
		if v, set, w := parseBoolEnv("ARM_USE_MSI"); set {
			cfg.UseMSI = v
		} else if w != "" {
			warnings = append(warnings, w)
		}
	}

	// use_cli: only apply env-var / default logic when the caller did NOT supply
	// an explicit HCL value (cfg.UseCLI == nil). An explicit false means "user
	// intentionally disabled CLI auth" and must be honoured even when
	// ARM_USE_CLI is unset (security-relevant: prevents ambient az identity).
	if cfg.UseCLI == nil {
		v, set, w := parseBoolEnv("ARM_USE_CLI")
		if w != "" {
			warnings = append(warnings, w)
		}
		if set {
			cfg.UseCLI = &v
		} else {
			// Default to true when neither HCL nor env var supplies a value,
			// matching SDKv2 EnvDefaultFunc("ARM_USE_CLI", true) behaviour.
			t := true
			cfg.UseCLI = &t
		}
	}

	return warnings
}

// useCLIValue returns the effective bool value of cfg.UseCLI, defaulting to
// false if the pointer is nil (should not occur after applyEnvFallbacks).
func useCLIValue(cfg *FrameworkAuthConfig) bool {
	if cfg.UseCLI == nil {
		return false
	}
	return *cfg.UseCLI
}

// resolveFrameworkAuthProvider accepts a FrameworkAuthConfig (decoded from the
// HCL provider block) and returns an azuredevops.AuthProvider. It mirrors the
// credential-priority logic of GetAuthProvider in the SDKv2 provider:
//
//  1. PAT — fastest path; no AAD round-trip.
//  2. AAD (client-secret / cert / OIDC / MSI / CLI) via aztfauth.NewCredential.
//
// Env-var fallbacks are applied before credential resolution so callers only
// need to set the struct fields from the HCL config; the function handles the
// ambient environment.
//
// Any diagnostic warnings from env-var parsing (e.g. malformed booleans) are
// logged via the standard logger. Callers that want framework diagnostics should
// call applyEnvFallbacks themselves and inspect the returned warnings.
func resolveFrameworkAuthProvider(_ context.Context, cfg FrameworkAuthConfig) (azuredevops.AuthProvider, []string, error) {
	warnings := applyEnvFallbacks(&cfg)

	// 1. Personal Access Token
	if cfg.PersonalAccessToken != "" {
		return azuredevops.NewAuthProviderPAT(cfg.PersonalAccessToken), warnings, nil
	}

	// 2. AAD / CLI / MSI / OIDC via aztfauth
	cred, err := aztfauth.NewCredential(aztfauth.Option{
		Logger:                     log.New(log.Default().Writer(), "[DEBUG] ", log.LstdFlags|log.Lmsgprefix),
		TenantId:                   cfg.TenantID,
		ClientId:                   cfg.ClientID,
		ClientIdFile:               cfg.ClientIDFilePath,
		UseClientSecret:            true,
		ClientSecret:               cfg.ClientSecret,
		ClientSecretFile:           cfg.ClientSecretPath,
		UseClientCert:              true,
		ClientCertBase64:           cfg.ClientCertificate,
		ClientCertPfxFile:          cfg.ClientCertificatePath,
		ClientCertPassword:         []byte(cfg.ClientCertificatePassword),
		UseOIDCToken:               cfg.UseOIDC,
		OIDCToken:                  cfg.OIDCToken,
		UseOIDCTokenFile:           cfg.UseOIDC,
		OIDCTokenFile:              cfg.OIDCTokenFilePath,
		UseOIDCTokenRequest:        cfg.UseOIDC,
		OIDCRequestToken:           cfg.OIDCRequestToken,
		OIDCRequestURL:             cfg.OIDCRequestURL,
		ADOServiceConnectionId:     cfg.OIDCAzureServiceConnectionID,
		UseMSI:                     cfg.UseMSI,
		UseAzureCLI:                useCLIValue(&cfg),
		AdditionallyAllowedTenants: cfg.AuxiliaryTenantIDs,
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("no credential method resolved: set personal_access_token / use_cli / use_msi / use_oidc or client_secret+tenant_id+client_id: %w", err)
	}

	const AzureDevOpsAppDefaultScope = "499b84ac-1321-427f-aa17-267ca6975798/.default"
	ap := azuredevops.NewAuthProviderAAD(cred, policy.TokenRequestOptions{
		Scopes: []string{AzureDevOpsAppDefaultScope},
	})
	return ap, warnings, nil
}
