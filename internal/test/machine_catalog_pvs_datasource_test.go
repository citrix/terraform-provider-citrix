// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPvsPreCheck_Azure(t *testing.T) {
	if v := os.Getenv("TEST_PVS_FARM_NAME"); v == "" {
		t.Fatal("TEST_PVS_FARM_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_SITE_NAME"); v == "" {
		t.Fatal("TEST_PVS_SITE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_STORE_NAME"); v == "" {
		t.Fatal("TEST_PVS_STORE_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_VDISK_NAME"); v == "" {
		t.Fatal("TEST_PVS_VDISK_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_SITE_ID"); v == "" {
		t.Fatal("TEST_PVS_SITE_ID must be set for acceptance tests")
	}
	if v := os.Getenv("TEST_PVS_VDISK_ID"); v == "" {
		t.Fatal("TEST_PVS_VDISK_ID must be set for acceptance tests")
	}
}

func TestPvsDataSource(t *testing.T) {
	pvs_site_id := os.Getenv("TEST_PVS_SITE_ID")
	pvs_vdisk_id := os.Getenv("TEST_PVS_VDISK_ID")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestPvsPreCheck_Azure(t)
		},
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: BuildPvsResource(t, pvs_test_data_source),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the ID of the admin scope
					resource.TestCheckResourceAttr("data.citrix_pvs.test_pvs_config", "pvs_site_id", pvs_site_id),
					// Verify the description of the admin scope
					resource.TestCheckResourceAttr("data.citrix_pvs.test_pvs_config", "pvs_vdisk_id", pvs_vdisk_id),
				),
			},
		},
	})
}

var (
	pvs_test_data_source = `
	data "citrix_pvs" "test_pvs_config" {
		pvs_farm_name = "%s"
		pvs_site_name = "%s"
		pvs_store_name = "%s"
		pvs_vdisk_name = "%s"
	}
	`
)

func BuildPvsResource(t *testing.T, pvsConfig string) string {
	farmName := os.Getenv("TEST_PVS_FARM_NAME")
	siteName := os.Getenv("TEST_PVS_SITE_NAME")
	storeName := os.Getenv("TEST_PVS_STORE_NAME")
	vdiskName := os.Getenv("TEST_PVS_VDISK_NAME")
	return fmt.Sprintf(pvsConfig, farmName, siteName, storeName, vdiskName)
}
