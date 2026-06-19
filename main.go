package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops"
)

// version is set by the GoReleaser build via -ldflags "-X main.version=...".
// It defaults to "dev" for local builds.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	ctx := context.Background()

	// Wrap the existing SDKv2 provider at protocol 6 so it can be muxed.
	upgradedSdkv2, err := tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer {
		p := azuredevops.Provider()
		p.UserAgent("terraform-provider-betterado", version)
		return schema.NewGRPCProviderServer(p)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Combine the upgraded SDKv2 provider with the framework provider stub.
	muxServer, err := tf6muxserver.NewMuxServer(ctx,
		func() tfprotov6.ProviderServer { return upgradedSdkv2 },
		providerserver.NewProtocol6(azuredevops.NewFrameworkProvider()),
	)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/parsoFish/betterado",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
