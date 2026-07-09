package testutils

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	internalprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
)

// GetProviderFactories returns a ProtoV6ProviderFactories map that serves the
// pure framework provider for use in acceptance tests. This replaces the former
// GetMuxedProviderFactories which used tf5to6server and tf6muxserver.
func GetProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"betterado": providerserver.NewProtocol6WithError(internalprovider.NewFrameworkProvider("test")),
	}
}
