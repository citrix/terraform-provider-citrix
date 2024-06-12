// Copyright © 2024. Citrix Systems, Inc.

package main

import (
	"context"
	"flag"
	"log"

	"github.com/citrix/terraform-provider-citrix/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name citrix --examples-dir internal/examples

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/citrix/citrix",
		Debug:   debug,
	})

	if err != nil {
		log.Fatal(err.Error())
	}
}
