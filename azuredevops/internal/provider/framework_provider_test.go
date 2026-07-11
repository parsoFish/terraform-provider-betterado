//go:build all || provider_framework

package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	frameworkprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestFrameworkProvider_HasReleaseFolderResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_release_folder" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_release_folder")
}

func TestFrameworkProvider_HasUserEntitlementResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_user_entitlement" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_user_entitlement")
}

func TestFrameworkProvider_HasTaskGroupResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_task_group" || metaResp.TypeName == "betterado_betterado_task_group" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_task_group")
}

func TestFrameworkProvider_HasGroupEntitlementResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_group_entitlement" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_group_entitlement")
}

func TestFrameworkProvider_HasServicePrincipalEntitlementResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_service_principal_entitlement" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_service_principal_entitlement")
}

func TestFrameworkProvider_HasNotificationSubscriptionResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_notification_subscription" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_notification_subscription resource")
}

func TestFrameworkProvider_HasNotificationSubscriptionDataSource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	factories := provWithDataSources.DataSources(context.Background())
	found := false
	for _, factory := range factories {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_notification_subscription" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_notification_subscription data source")
}

func TestFrameworkProvider_HasPipelineResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_pipeline" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_pipeline resource")
}

func TestFrameworkProvider_HasPipelineDataSource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	factories := provWithDataSources.DataSources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one data source factory")

	foundPipeline := false
	foundPipelineRun := false
	for _, factory := range factories {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		switch metaResp.TypeName {
		case "betterado_pipeline":
			foundPipeline = true
		case "betterado_pipeline_run":
			foundPipelineRun = true
		}
	}
	require.True(t, foundPipeline, "framework provider must register betterado_pipeline data source")
	require.True(t, foundPipelineRun, "framework provider must register betterado_pipeline_run data source")
}

func TestFrameworkProvider_HasAccountsDataSource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	factories := provWithDataSources.DataSources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one data source factory")

	found := false
	for _, factory := range factories {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_accounts" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_accounts data source")
}

func TestFrameworkProvider_HasProfileDataSource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	factories := provWithDataSources.DataSources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one data source factory")

	found := false
	for _, factory := range factories {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_profile" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_profile data source")
}

func TestFrameworkProvider(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")

	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	// Collect registered resource type names.
	resourceTypes := map[string]bool{}
	for _, factory := range provWithResources.Resources(context.Background()) {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		resourceTypes[metaResp.TypeName] = true
	}

	// Collect registered data-source type names.
	dataSourceTypes := map[string]bool{}
	for _, factory := range provWithDataSources.DataSources(context.Background()) {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		dataSourceTypes[metaResp.TypeName] = true
	}

	// Assert each of the four testplan resource types is present.
	// NewTestPlanResource → betterado_test_plan (resource)
	require.True(t, resourceTypes["betterado_test_plan"],
		"NewTestPlanResource: betterado_test_plan must be registered in Resources()")

	// NewTestSuiteResource → betterado_test_suite (resource)
	require.True(t, resourceTypes["betterado_test_suite"],
		"NewTestSuiteResource: betterado_test_suite must be registered in Resources()")

	// NewTestConfigurationResource → betterado_test_configuration (resource)
	require.True(t, resourceTypes["betterado_test_configuration"],
		"NewTestConfigurationResource: betterado_test_configuration must be registered in Resources()")

	// NewTestVariableResource → betterado_test_variable (resource)
	require.True(t, resourceTypes["betterado_test_variable"],
		"NewTestVariableResource: betterado_test_variable must be registered in Resources()")

	// Assert each of the three testplan data-source types is present.
	// NewTestPlanDataSource → betterado_test_plan (data source)
	require.True(t, dataSourceTypes["betterado_test_plan"],
		"NewTestPlanDataSource: betterado_test_plan must be registered in DataSources()")

	// NewTestRunDataSource → betterado_test_run (data source)
	require.True(t, dataSourceTypes["betterado_test_run"],
		"NewTestRunDataSource: betterado_test_run must be registered in DataSources()")

	// NewTestResultDataSource → betterado_test_result (data source)
	require.True(t, dataSourceTypes["betterado_test_result"],
		"NewTestResultDataSource: betterado_test_result must be registered in DataSources()")
}

func TestFrameworkProvider_HasPipelineApprovalResources(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")

	// Check resource
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	foundResource := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_pipeline_approval" {
			foundResource = true
			break
		}
	}
	require.True(t, foundResource, "framework provider must register betterado_pipeline_approval resource")

	// Check data source
	provWithDataSources, ok := p.(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})
	require.True(t, ok, "framework provider must implement DataSources()")

	dsFactories := provWithDataSources.DataSources(context.Background())
	foundDataSource := false
	for _, factory := range dsFactories {
		ds := factory()
		var metaResp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_pipeline_approvals" {
			foundDataSource = true
			break
		}
	}
	require.True(t, foundDataSource, "framework provider must register betterado_pipeline_approvals data source")
}

func TestFrameworkProvider_HasExtensionInstallResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider("test")
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_extension_install" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_extension_install")
}

// TestFrameworkProvider_MuxFree verifies that the pure-framework provider can be
// instantiated and serves resources without requiring SDKv2 mux. This is the
// post-mux-cutover sanity check for WI-4.
func TestFrameworkProvider_MuxFree(t *testing.T) {
	// Verify the framework provider can be instantiated and serves resources
	// without requiring SDKv2 mux. This is the post-mux-cutover sanity check.
	p := frameworkprovider.NewFrameworkProvider("test")
	require.NotNil(t, p, "NewFrameworkProvider must return a non-nil provider")

	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must expose resources")

	// Spot-check: betterado_project must be registered (representative resource).
	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)
		if metaResp.TypeName == "betterado_project" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_project")
}

// buildProviderSchema returns the provider schema and its tftypes.Type so
// tests can construct a well-typed tfsdk.Config without going through the
// full plugin server stack.
func buildProviderSchema(t *testing.T) (schema.Schema, tftypes.Type) {
	t.Helper()
	p := frameworkprovider.NewFrameworkProvider("test")
	var schemaResp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "provider.Schema must not return errors")
	schemaType := schemaResp.Schema.Type().TerraformType(context.Background())
	return schemaResp.Schema, schemaType
}

// nullStringAttrs returns a map of tftypes.Value for every string attribute in
// the provider schema, with all values set to null.
func nullProviderConfig(schemaType tftypes.Type) map[string]tftypes.Value {
	objType, ok := schemaType.(tftypes.Object)
	if !ok {
		panic("provider schema type must be tftypes.Object")
	}
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil) // null
	}
	return vals
}

// TestFrameworkConfigure_NoCredential verifies that calling Configure() with a
// config where all credential fields are empty/false and all relevant env vars
// are unset produces exactly one error diagnostic naming the available auth
// methods.
func TestFrameworkConfigure_NoCredential(t *testing.T) {
	// Unset all credential env vars so the resolver cannot fall back to them.
	credEnvVars := []string{
		"AZDO_PERSONAL_ACCESS_TOKEN",
		"ARM_CLIENT_ID", "AZURE_CLIENT_ID", "ARM_CLIENT_ID_FILE_PATH",
		"ARM_TENANT_ID",
		"ARM_CLIENT_SECRET", "ARM_CLIENT_SECRET_PATH", "ARM_CLIENT_SECRET_FILE_PATH",
		"ARM_CLIENT_CERTIFICATE_PATH", "ARM_CLIENT_CERTIFICATE", "ARM_CLIENT_CERTIFICATE_PASSWORD",
		"ARM_OIDC_TOKEN", "ARM_OIDC_TOKEN_FILE_PATH", "AZURE_FEDERATED_TOKEN_FILE",
		"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN",
		"ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL", "SYSTEM_OIDCREQUESTURI",
		"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID",
		"AZURESUBSCRIPTION_SERVICE_CONNECTION_ID",
		"ARM_USE_OIDC", "ARM_USE_MSI",
	}
	for _, ev := range credEnvVars {
		t.Setenv(ev, "")
	}
	// Explicitly disable CLI fallback: applyEnvFallbacks defaults use_cli=true
	// when ARM_USE_CLI is unset; set it to "false" so the resolver does not
	// fall back to Azure CLI authentication.
	t.Setenv("ARM_USE_CLI", "false")
	// Also unset org_service_url env.
	t.Setenv("AZDO_ORG_SERVICE_URL", "https://dev.azure.com/fakeorg")

	provSchema, schemaType := buildProviderSchema(t)

	// Build a null-filled config plus the org URL; all credentials empty.
	configVals := nullProviderConfig(schemaType)
	objType := schemaType.(tftypes.Object)

	// Set org_service_url to a non-empty value so the resolver reaches the
	// credential check (not blocked by missing org URL).
	configVals["org_service_url"] = tftypes.NewValue(tftypes.String, "https://dev.azure.com/fakeorg")

	// Explicitly set use_cli=false (not null) so the resolver sees it disabled.
	configVals["use_cli"] = tftypes.NewValue(objType.AttributeTypes["use_cli"], false)
	configVals["use_msi"] = tftypes.NewValue(objType.AttributeTypes["use_msi"], false)
	configVals["use_oidc"] = tftypes.NewValue(objType.AttributeTypes["use_oidc"], false)

	rawConfig := tftypes.NewValue(schemaType, configVals)

	p := frameworkprovider.NewFrameworkProvider("test")
	var configResp provider.ConfigureResponse
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    rawConfig,
			Schema: provSchema,
		},
	}, &configResp)

	require.True(t, configResp.Diagnostics.HasError(),
		"Configure() must report an error when no credential is available")

	// AC7: exactly one error; its detail must reference the available auth methods.
	var errs []string
	for _, d := range configResp.Diagnostics.Errors() {
		errs = append(errs, d.Summary())
	}
	require.Len(t, errs, 1, "Configure() must produce exactly one error diagnostic, got: %v", errs)
	require.Contains(t, errs[0], "credential",
		"error summary must reference available auth methods (got: %q)", errs[0])
}
