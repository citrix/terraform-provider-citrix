// Copyright © 2025. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/middleware"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// **Important Note**: Please make sure this function is updated when provider.Configure() is updated
func sharedClientForSweepers(ctx context.Context) *citrixclient.CitrixDaasClient {
	clientId := os.Getenv("CITRIX_CLIENT_ID")
	clientSecret := os.Getenv("CITRIX_CLIENT_SECRET")
	hostname := os.Getenv("CITRIX_HOSTNAME")
	environment := os.Getenv("CITRIX_ENVIRONMENT")
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	disableSslVerification := strings.EqualFold(os.Getenv("CITRIX_DISABLE_SSL_VERIFICATION"), "true")
	catalog_service_host_name := os.Getenv("CITRIX_QUICK_DEPLOY_HOST_NAME")

	if environment == "" {
		environment = "Production" // default to production
	}

	if customerId == "" {
		customerId = "CitrixOnPremises"
	}

	onPremises := customerId == "CitrixOnPremises"

	apiGateway := true
	ccUrl := ""
	if !onPremises {
		switch environment {
		case "Production":
			ccUrl = "api.cloud.com"
		case "Staging":
			ccUrl = "api.cloudburrito.com"
		case "Japan":
			ccUrl = "api.citrixcloud.jp"
		case "JapanStaging":
			ccUrl = "api.citrixcloudstaging.jp"
		case "Gov":
			ccUrl = fmt.Sprintf("registry.citrixworkspacesapi.us/%s", customerId)
		case "GovStaging":
			ccUrl = fmt.Sprintf("registry.ctxwsstgapi.us/%s", customerId)
		}
		if hostname == "" {
			switch environment {
			case "Gov":
				hostname = fmt.Sprintf("%s.xendesktop.us", customerId)
				apiGateway = false
			case "GovStaging":
				hostname = fmt.Sprintf("%s.xdstaging.us", customerId)
				apiGateway = false
			default:
				hostname = ccUrl
			}
		} else if !strings.HasPrefix(hostname, "api.") {
			// When a cloud customer sets explicit hostname to the cloud DDC, bypass API Gateway
			apiGateway = false
		}
	}

	var authUrl string
	isGov := false
	if onPremises {
		authUrl = fmt.Sprintf("https://%s/citrix/orchestration/api/tokens", hostname)
	} else {
		switch environment {
		case "Production":
			authUrl = fmt.Sprintf("https://api.cloud.com/cctrustoauth2/%s/tokens/clients", customerId)
		case "Staging":
			authUrl = fmt.Sprintf("https://api.cloudburrito.com/cctrustoauth2/%s/tokens/clients", customerId)
		case "Japan":
			authUrl = fmt.Sprintf("https://api.citrixcloud.jp/cctrustoauth2/%s/tokens/clients", customerId)
		case "JapanStaging":
			authUrl = fmt.Sprintf("https://api.citrixcloudstaging.jp/cctrustoauth2/%s/tokens/clients", customerId)
		case "Gov":
			authUrl = fmt.Sprintf("https://trust.citrixworkspacesapi.us/%s/tokens/clients", customerId)
			isGov = true
		case "GovStaging":
			authUrl = fmt.Sprintf("https://trust.ctxwsstgapi.us/%s/tokens/clients", customerId)
			isGov = true
		default:
			authUrl = fmt.Sprintf("https://%s/cctrustoauth2/%s/tokens/clients", hostname, customerId)
		}
	}

	catalogServiceHostname := ""
	if catalog_service_host_name != "" {
		// If customer specified a quick create host name, use it
		catalogServiceHostname = catalog_service_host_name
	} else {
		switch environment {
		case "Production":
			catalogServiceHostname = "api.cloud.com/catalogservice"
		case "Staging":
			catalogServiceHostname = "api.cloudburrito.com/catalogservice"
		case "Japan":
			catalogServiceHostname = "api.citrixcloud.jp/catalogservice"
		case "JapanStaging":
			catalogServiceHostname = "api.citrixcloudstaging.jp/catalogservice"
		case "Gov":
			catalogServiceHostname = "api.cloud.us/catalogservice"
		case "GovStaging":
			catalogServiceHostname = "api.cloudstaging.us/catalogservice"
		}
	}

	userAgent := "citrix-terraform-provider/" + "gotester" + " (https://github.com/citrix/terraform-provider-citrix)"

	// Initialize CVAD client
	client := &citrixclient.CitrixDaasClient{}
	//nolint:errcheck // Test setup, errors not critical
	token, _, _ := client.SetupCitrixClientsContext(ctx, authUrl, ccUrl, hostname, customerId, clientId, clientSecret, onPremises, apiGateway, isGov, disableSslVerification, &userAgent, environment, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	if !onPremises {
		client.InitializeCitrixCloudClients(ctx, ccUrl, hostname, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	}
	//nolint:errcheck // Test setup, errors not critical
	_, _ = client.InitializeCitrixDaasClient(ctx, customerId, token, onPremises, apiGateway, disableSslVerification, &userAgent)

	// Set Quick Deploy Client
	if catalogServiceHostname != "" {
		client.InitializeQuickDeployClient(ctx, catalogServiceHostname, middleware.MiddlewareAuthFunc)
	}

	return client
}
