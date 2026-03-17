// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/hashicorp/terraform-provider-dns/internal/provider"
)

func main() {
	ctx := context.Background()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	sdkv2Server, err := tf5to6server.UpgradeServer(ctx, provider.New().GRPCProvider)
	if err != nil {
		panic(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(provider.NewFrameworkProvider()),
		func() tfprotov6.ProviderServer { return sdkv2Server },
		func() tfprotov6.ProviderServer { return provider.CodeMigrationServer{} },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/hashicorp/dns",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
