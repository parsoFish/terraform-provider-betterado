package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops"
)

// version is set by the GoReleaser build via -ldflags "-X main.version=...".
// It defaults to "dev" for local builds.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		Debug:        debug,
		ProviderAddr: "registry.terraform.io/parsoFish/betterado",
		ProviderFunc: func() *schema.Provider {
			p := azuredevops.Provider()
			p.UserAgent("terraform-provider-betterado", version)
			return p
		},
	})
}
