package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent"
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

// Configure creates the AggregatedClient from environment variables and stores
// it as resp.ResourceData so that framework resources (e.g. TaskGroupResource)
// can retrieve it via resource.ConfigureRequest.ProviderData.
//
// The mux multiplexes the provider configuration call across both the SDKv2
// provider and this framework stub. The SDKv2 provider parses the HCL provider
// block; this stub reads the same values from the canonical env vars so that
// it can build its own client independently (the two providers share no runtime
// state).
func (p *BetteradoFrameworkProvider) Configure(ctx context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")

	// If neither credential source is available we leave resp.ResourceData nil
	// and let individual resource Configure calls report the error when they
	// actually need the client. This avoids breaking the provider configure
	// step when running plan/apply with non-PAT auth (e.g. AAD) — the SDKv2
	// provider handles those cases; the framework stub only needs the client
	// when its own resources are used.
	if orgURL == "" || pat == "" {
		return
	}

	authProvider := azuredevops.NewAuthProviderPAT(pat)
	agg, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure Azure DevOps client",
			err.Error(),
		)
		return
	}

	resp.ResourceData = agg
	resp.DataSourceData = agg
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
	return []func() resource.Resource{
		taskagent.NewTaskGroupResource,
	}
}

func (p *BetteradoFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
