// Copyright © 2025. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_service_account", &resource.Sweeper{
		Name: "citrix_service_account",
		F: func(serviceAccount string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			var errs *multierror.Error

			serviceAccountName := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME")
			err := serviceAccountSweeper(ctx, client, serviceAccountName)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			// test updated role
			serviceAccountNameUpdated := serviceAccountName + "-updated"
			err = serviceAccountSweeper(ctx, client, serviceAccountNameUpdated)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			return errs.ErrorOrNil()
		},
		Dependencies: []string{"citrix_machine_catalog"},
	})
}

func TestServiceAccountPreCheck_AD(t *testing.T) {
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_DISPLAY_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AD_DOMAIN_NAME"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AD_DOMAIN_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID_UPDATED"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID_UPDATED must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET_UPDATED"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET_UPDATED must be set for acceptance tests")
	}
}

func TestServiceAccountPreCheck_AAD(t *testing.T) {
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_DISPLAY_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_TENANT_ID"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AAD_TENANT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_APP_ID"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AAD_APP_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_APP_SECRET"); v == "" {
		t.Fatal("TEST_SERVICE_ACCOUNT_AAD_APP_SECRET must be set for acceptance tests")
	}
}

func TestServiceAccountAD(t *testing.T) {

	displayName := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME")
	adDomainName := os.Getenv("TEST_SERVICE_ACCOUNT_AD_DOMAIN_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestServiceAccountPreCheck_AD(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: composeTestResourceTf(BuildServiceAccountResourceAD(t, testServiceAccountResourceAD)),

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAD", "display_name", displayName),
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAD", "identity_provider_identifier", adDomainName),
				),
			},

			// ImportState testing
			{
				ResourceName:      "citrix_service_account.testServiceAccountAD",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"account_secret", "account_secret_format"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(BuildServiceAccountResourceAD(t, testServiceAccountResourceAD_updated)),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of hypervisor
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAD", "display_name", fmt.Sprintf("%s-updated", displayName)),
				),
			},
		},
	})
}

func TestServiceAccountAAD(t *testing.T) {

	displayName := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME")
	tenantId := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_TENANT_ID")
	appId := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_APP_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestServiceAccountPreCheck_AAD(t)
		},
		Steps: []resource.TestStep{

			// Create and Read testing
			{
				Config: composeTestResourceTf(BuildServiceAccountResourceAAD(t, testServiceAccountResourceAAD)),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAAD", "display_name", displayName),
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAAD", "identity_provider_identifier", tenantId),
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAAD", "account_id", appId),
				),
			},

			// ImportState testing
			{
				ResourceName:            "citrix_service_account.testServiceAccountAAD",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"account_secret", "account_secret_format"},
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(BuildServiceAccountResourceAAD(t, testServiceAccountResourceAAD_updated)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_service_account.testServiceAccountAAD", "display_name", fmt.Sprintf("%s-updated", displayName)),
				),
			},
		},
	})
}

func BuildServiceAccountResourceAD(t *testing.T, serviceAccount string) string {
	displayName := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME")
	adDomainName := os.Getenv("TEST_SERVICE_ACCOUNT_AD_DOMAIN_NAME")
	accountId := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID")
	accountSecret := os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET")

	if serviceAccount == testServiceAccountResourceAD_updated {
		accountId = os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_ID_UPDATED")
		accountSecret = os.Getenv("TEST_SERVICE_ACCOUNT_AD_ACCOUNT_SECRET_UPDATED")
	}

	return fmt.Sprintf(serviceAccount, displayName, adDomainName, accountId, accountSecret)
}

func BuildServiceAccountResourceAAD(t *testing.T, serviceAccount string) string {
	displayName := os.Getenv("TEST_SERVICE_ACCOUNT_DISPLAY_NAME")
	tenantId := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_TENANT_ID")
	appId := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_APP_ID")
	appSecret := os.Getenv("TEST_SERVICE_ACCOUNT_AAD_APP_SECRET")

	return fmt.Sprintf(serviceAccount, displayName, tenantId, appId, appSecret)
}

var (
	testServiceAccountResourceAD = `
	resource citrix_service_account testServiceAccountAD {
		display_name = "%s"
    	identity_provider_type = "ActiveDirectory"
    	identity_provider_identifier = "%s"
    	account_id = "%s"
    	account_secret = "%s"
    	account_secret_format = "PlainText"
	}
	`

	testServiceAccountResourceAD_updated = `
	resource citrix_service_account testServiceAccountAD {
		display_name = "%s-updated"
		identity_provider_type = "ActiveDirectory"
		identity_provider_identifier = "%s"
		account_id = "%s"
		account_secret = "%s"
		account_secret_format = "PlainText"
	}
	`
	testServiceAccountResourceAAD = `
	resource citrix_service_account testServiceAccountAAD {
		display_name = "%s"
    	identity_provider_type = "AzureAD"
    	identity_provider_identifier = "%s"
    	account_id = "%s"
    	account_secret = "%s"
    	account_secret_format = "PlainText"
	}`

	testServiceAccountResourceAAD_updated = `
	resource citrix_service_account testServiceAccountAAD {
		display_name = "%s-updated"
    	identity_provider_type = "AzureAD"
    	identity_provider_identifier = "%s"
    	account_id = "%s"
    	account_secret = "%s"
    	account_secret_format = "PlainText"
	}`
)

func serviceAccountSweeper(ctx context.Context, client *citrixclient.CitrixDaasClient, serviceAccountName string) error {
	getServiceAccountsRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetServiceAccounts(ctx)
	serviceAccountsResponse, _, err := citrixclient.ExecuteWithRetry[*citrixorchestration.ServiceAccountResponseModelCollection](getServiceAccountsRequest, client)
	if err != nil {
		return fmt.Errorf("Error getting service accounts: %s", err)
	}

	serviceAccounts := serviceAccountsResponse.GetItems()
	for _, serviceAccount := range serviceAccounts {
		if strings.EqualFold(serviceAccount.GetDisplayName(), serviceAccountName) {
			serviceAccountId := serviceAccount.GetServiceAccountUid()
			deleteServiceAccountRequest := client.ApiClient.IdentityAPIsDAAS.IdentityDeleteServiceAccount(ctx, serviceAccountId)
			httpResp, err := citrixclient.AddRequestData(deleteServiceAccountRequest, client).Execute()
			if err != nil && httpResp.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error deleting Service Account " + serviceAccountName + " during sweep",
				)
			}

			// service account deleted successfully
			return nil
		}
	}

	// service account does not exist and was probably deleted

	return nil
}
