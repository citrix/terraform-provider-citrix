// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("citrix_quickdeploy_catalog", &resource.Sweeper{
		Name: "citrix_quickdeploy_catalog",
		F: func(hypervisor string) error {
			ctx := context.Background()
			client := sharedClientForSweepers(ctx)

			var errs *multierror.Error
			// MCS AD machine catalog sweep
			quickdeployCatalogName := os.Getenv("TEST_QUICKDEPLOY_CATALOG_NAME")
			err := quickdeployCatalogSweeper(ctx, quickdeployCatalogName, client)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			return errs.ErrorOrNil()
		},
	})
}

func TestQuickDeployCatalogPreCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	if v := os.Getenv("TEST_QUICKDEPLOY_CATALOG_NAME"); v == "" {
		t.Fatal("TEST_QUICKDEPLOY_CATALOG_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_QUICKDEPLOY_IMAGE_NAME"); v == "" {
		t.Fatal("TEST_QUICKDEPLOY_IMAGE_NAME must be set for acceptance tests")
	}
	// Need to support image update in quickdeploy catalog first before enabling this check
	// if v := os.Getenv("TEST_QUICKDEPLOY_IMAGE_NAME_UPDATED"); v == "" {
	// 	t.Fatal("TEST_QUICKDEPLOY_IMAGE_NAME_UPDATED must be set for acceptance tests")
	// }
}

func TestQuickDeployCatalogResource(t *testing.T) {
	name := os.Getenv("TEST_QUICKDEPLOY_CATALOG_NAME")
	imageName := os.Getenv("TEST_QUICKDEPLOY_IMAGE_NAME")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestQuickDeployCatalogPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: func() string {
					catalogResource := BuildQuickDeployCatalogResource(t, quickdeploycatalog_testResources, name)
					return composeTestResourceTf(
						catalogResource,
						BuildQuickDeployTemplateImageDataSource(t, quickdeploy_template_image_test_data_source_using_name, imageName),
					)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "name", name),
					// Verify catalog type
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "catalog_type", "MultiSession"),
					// Verify number of users
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "number_of_users", "2"),
					// Verify subscription name
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "subscription_name", "Citrix Managed"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_quickdeploy_catalog.test-catalog",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"machine_naming_scheme"},
			},
			//Update description, master image and add machine test
			{
				Config: func() string {
					catalogResource := BuildQuickDeployCatalogResource(t, quickdeploycatalog_testResources_updated, name)
					return composeTestResourceTf(
						catalogResource,
						BuildQuickDeployTemplateImageDataSource(t, quickdeploy_template_image_test_data_source_using_name, imageName),
					)
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of catalog
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "name", name),
					// Verify catalog type
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "catalog_type", "MultiSession"),
					// Verify number of users
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "number_of_users", "10"),
					// Verify subscription name
					resource.TestCheckResourceAttr("citrix_quickdeploy_catalog.test-catalog", "subscription_name", "Citrix Managed"),
				),
			},
			//Delete testing automatically occurs in TestCase
		},
	})
}

var (
	quickdeploycatalog_testResources = `
	resource citrix_quickdeploy_catalog test-catalog {
		name = "%s"
		catalog_type = "MultiSession"
		region = "eastus"
		subscription_name = "Citrix Managed"
		template_image_id = data.citrix_quickdeploy_template_image.test_image.id
		machine_size = "d8asv5"
		storage_type = "StandardSSD_LRS"
		number_of_users = 2
		max_users_per_vm = 8
		machine_naming_scheme = {
			naming_scheme = "gotest-###"
			naming_scheme_type = "Numeric"
		}
		power_schedule = {
			peak_min_instances = 1
			off_peak_min_instances = 1
			weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
			peak_start_time = 9
			peak_end_time = 18
			peak_time_zone_id = "Pacific Standard Time"
			peak_off_delay = 15
		}
	}
	`
	quickdeploycatalog_testResources_updated = `
	resource citrix_quickdeploy_catalog test-catalog {
		name = "%s"
		catalog_type = "MultiSession"
		region = "eastus"
		subscription_name = "Citrix Managed"
		template_image_id = data.citrix_quickdeploy_template_image.test_image.id
		machine_size = "d8asv5"
		storage_type = "StandardSSD_LRS"
		number_of_users = 10
		max_users_per_vm = 8
		machine_naming_scheme = {
			naming_scheme = "gotest-###"
			naming_scheme_type = "Numeric"
		}
		power_schedule = {
			peak_min_instances = 1
			off_peak_min_instances = 1
			weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
			peak_start_time = 9
			peak_end_time = 18
			peak_time_zone_id = "Pacific Standard Time"
			peak_off_delay = 15
		}
	}
	`
)

func BuildQuickDeployCatalogResource(t *testing.T, quickDeployCatalogResource, catalogName string) string {
	return fmt.Sprintf(quickDeployCatalogResource, catalogName)
}

func quickdeployCatalogSweeper(ctx context.Context, catalogName string, client *citrixclient.CitrixDaasClient) error {
	getManagedCatalogsRequest := client.QuickDeployClient.CatalogCMD.GetCustomerManagedCatalogs(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId)
	catalogs, httpResp, err := citrixclient.ExecuteWithRetry[*citrixquickdeploy.CustomerManagedCatalogOverviewsModel](getManagedCatalogsRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Resource does not exist in remote, no need to delete
			return nil
		}
		return fmt.Errorf("Error getting machine catalog: %w", err)
	}

	// Get Catalog ID from name
	catalogId := ""
	for _, catalog := range catalogs.GetItems() {
		if catalog.GetName() == catalogName {
			catalogId = catalog.GetId()
			break
		}
	}
	if catalogId == "" {
		// Resource does not exist in remote, no need to delete
		return nil
	}

	deleteManagedCatalogRequest := client.QuickDeployClient.CatalogCMD.DeleteCustomerCatalog(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId, catalogId)
	httpResp, err = citrixclient.AddRequestData(deleteManagedCatalogRequest, client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		log.Printf("Error destroying %s during sweep: %s", catalogName, err)
	}
	return nil
}
