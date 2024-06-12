// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"os"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment

func TestProviderPreCheck(t *testing.T) {
	if v := os.Getenv("CITRIX_CLIENT_ID"); v == "" {
		t.Fatal("CITRIX_CLIENT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("CITRIX_CLIENT_SECRET"); v == "" {
		t.Fatal("CITRIX_CLIENT_SECRET must be set for acceptance tests")
	}

	if v := os.Getenv("CITRIX_CUSTOMER_ID"); v == "" || v == "CitrixOnPremises" {
		testOnPremProviderPreCheck(t)
	}
}

func testOnPremProviderPreCheck(t *testing.T) {
	if v := os.Getenv("CITRIX_HOSTNAME"); v == "" {
		t.Fatal("CITRIX_HOSTNAME must be set for acceptance tests")
	}
}

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"citrix": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)
