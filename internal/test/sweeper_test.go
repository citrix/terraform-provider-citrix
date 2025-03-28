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

	if environment == "" {
		environment = "Production" // default to production
	}

	if customerId == "" {
		customerId = "CitrixOnPremises"
	}

	onPremises := false
	if customerId == "CitrixOnPremises" {
		onPremises = true
	}

	apiGateway := true
	ccUrl := ""
	if !onPremises {
		if environment == "Production" {
			ccUrl = "api.cloud.com"
		} else if environment == "Staging" {
			ccUrl = "api.cloudburrito.com"
		} else if environment == "Japan" {
			ccUrl = "api.citrixcloud.jp"
		} else if environment == "JapanStaging" {
			ccUrl = "api.citrixcloudstaging.jp"
		} else if environment == "Gov" {
			ccUrl = fmt.Sprintf("registry.citrixworkspacesapi.us/%s", customerId)
		} else if environment == "GovStaging" {
			ccUrl = fmt.Sprintf("registry.ctxwsstgapi.us/%s", customerId)
		}
		if hostname == "" {
			if environment == "Gov" {
				hostname = fmt.Sprintf("%s.xendesktop.us", customerId)
				apiGateway = false
			} else if environment == "GovStaging" {
				hostname = fmt.Sprintf("%s.xdstaging.us", customerId)
				apiGateway = false
			} else {
				hostname = ccUrl
			}
		} else if !strings.HasPrefix(hostname, "api.") {
			// When a cloud customer sets explicit hostname to the cloud DDC, bypass API Gateway
			apiGateway = false
		}
	}

	authUrl := ""
	isGov := false
	if onPremises {
		authUrl = fmt.Sprintf("https://%s/citrix/orchestration/api/tokens", hostname)
	} else {
		if environment == "Production" {
			authUrl = fmt.Sprintf("https://api.cloud.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Staging" {
			authUrl = fmt.Sprintf("https://api.cloudburrito.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Japan" {
			authUrl = fmt.Sprintf("https://api.citrixcloud.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "JapanStaging" {
			authUrl = fmt.Sprintf("https://api.citrixcloudstaging.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Gov" {
			authUrl = fmt.Sprintf("https://trust.citrixworkspacesapi.us/%s/tokens/clients", customerId)
			isGov = true
		} else if environment == "GovStaging" {
			authUrl = fmt.Sprintf("https://trust.ctxwsstgapi.us/%s/tokens/clients", customerId)
			isGov = true
		} else {
			authUrl = fmt.Sprintf("https://%s/cctrustoauth2/%s/tokens/clients", hostname, customerId)
		}
	}

	userAgent := "citrix-terraform-provider/" + "gotester" + " (https://github.com/citrix/terraform-provider-citrix)"

	// Initialize CVAD client
	client := &citrixclient.CitrixDaasClient{}
	token, _, _ := client.SetupCitrixClientsContext(ctx, authUrl, ccUrl, hostname, customerId, clientId, clientSecret, onPremises, apiGateway, isGov, disableSslVerification, &userAgent, environment, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	if !onPremises {
		client.InitializeCitrixCloudClients(ctx, ccUrl, hostname, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	}
	client.InitializeCitrixDaasClient(ctx, customerId, token, onPremises, apiGateway, disableSslVerification, &userAgent)

	return client
}
