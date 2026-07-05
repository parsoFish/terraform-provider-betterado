//go:build all || provider_framework

package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	frameworkprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestFrameworkProvider_HasReleaseFolderResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()
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
	p := frameworkprovider.NewFrameworkProvider()

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
	p := frameworkprovider.NewFrameworkProvider()

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
