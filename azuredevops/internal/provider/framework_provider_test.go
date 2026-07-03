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

// TestFrameworkProvider asserts that all seven testplan types — four resources
// (NewTestPlanResource, NewTestSuiteResource, NewTestConfigurationResource,
// NewTestVariableResource) and three data sources (NewTestPlanDataSource,
// NewTestRunDataSource, NewTestResultDataSource) — are registered in the
// framework provider.  Removing any single registration causes this test to fail.
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
