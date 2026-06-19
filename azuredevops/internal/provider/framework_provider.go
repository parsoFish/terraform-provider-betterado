package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// BetteradoFrameworkProvider is the terraform-plugin-framework provider stub.
// It is multiplexed with the existing SDKv2 provider in main.go so that
// framework resources can be registered here without touching main.go.
type BetteradoFrameworkProvider struct{}

// NewFrameworkProvider returns a provider.Provider for use in the mux setup.
func NewFrameworkProvider() provider.Provider {
	return &BetteradoFrameworkProvider{}
}

func (p *BetteradoFrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azuredevops"
}

func (p *BetteradoFrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *BetteradoFrameworkProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

// FRAMEWORK EXTENSION POINT
//
// To register a new terraform-plugin-framework resource, add it to the slice
// returned by Resources() below:
//
//	func (p *BetteradoFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
//	    return []func() resource.Resource{
//	        newMyFrameworkResource,   // ← add here
//	    }
//	}
//
// To register a new framework data source, add it to DataSources():
//
//	func (p *BetteradoFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
//	    return []func() datasource.DataSource{
//	        newMyFrameworkDataSource, // ← add here
//	    }
//	}
//
// No changes to main.go are needed after the mux is wired in this file.

func (p *BetteradoFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *BetteradoFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
