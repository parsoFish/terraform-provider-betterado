package testutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops"
)

// GetMuxedProviderFactories returns a ProtoV6ProviderFactories map that serves
// both the SDKv2 resources (via protocol-upgrade) and the framework resources
// (via the framework provider stub) through a single mux server.
//
// Use this in resource.TestCase.ProtoV6ProviderFactories when the resource
// under test is registered on the framework provider (e.g. betterado_task_group).
func GetMuxedProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"betterado": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			// Upgrade the SDKv2 provider to protocol 6.
			sdkv2Provider := azuredevops.Provider()
			upgradedSdkv2, err := tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer {
				return schema.NewGRPCProviderServer(sdkv2Provider)
			})
			if err != nil {
				return nil, err
			}

			// Build the mux that combines upgraded SDKv2 + framework provider.
			mux, err := tf6muxserver.NewMuxServer(ctx,
				func() tfprotov6.ProviderServer { return upgradedSdkv2 },
				providerserver.NewProtocol6(azuredevops.NewFrameworkProvider()),
			)
			if err != nil {
				return nil, err
			}
			return mux.ProviderServer(), nil
		},
	}
}
