// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCitrixManagedAzureImageResourcePreCheck(t *testing.T) {
	if v := os.Getenv("TEST_TEMPLATE_IMAGE_NAME"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_NOTES"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_NOTES must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_SUBSCRIPTION_NAME"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_SUBSCRIPTION_NAME must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_VHD_URI"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_VHD_URI must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_GUEST_DISK_URI"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_GUEST_DISK_URI must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_NAME_UPDATED"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_NAME_UPDATED must be set for acceptance tests")
	}

	if v := os.Getenv("TEST_TEMPLATE_IMAGE_NOTES_UPDATED"); v == "" {
		t.Fatal("TEST_TEMPLATE_IMAGE_NOTES_UPDATED must be set for acceptance tests")
	}
}

func TestCitrixManagedAzureImageResource(t *testing.T) {
	name := os.Getenv("TEST_TEMPLATE_IMAGE_NAME")
	notes := os.Getenv("TEST_TEMPLATE_IMAGE_NOTES")
	subscriptionName := os.Getenv("TEST_TEMPLATE_IMAGE_SUBSCRIPTION_NAME")
	vhdUri := os.Getenv("TEST_TEMPLATE_IMAGE_VHD_URI")
	guestDiskUri := os.Getenv("TEST_TEMPLATE_IMAGE_GUEST_DISK_URI")

	name_updated := os.Getenv("TEST_TEMPLATE_IMAGE_NAME_UPDATED")
	notes_updated := os.Getenv("TEST_TEMPLATE_IMAGE_NOTES_UPDATED")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestCitrixManagedAzureImageResourcePreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing for QCS AWS Workspaces Image resource
			{
				Config: composeTestResourceTf(
					BuildCitrixManagedAzureImageResource(t, name, notes, subscriptionName, vhdUri, guestDiskUri),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// AWS Workspaces Image Tests
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "name", name),
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "notes", notes),
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "region", "East US"),
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "vhd_uri", vhdUri),
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "guest_disk_uri", guestDiskUri),
				),
			},

			// Update testing for QCS AWS Workspaces Image resource
			{
				Config: composeTestResourceTf(
					BuildCitrixManagedAzureImageResource(t, name_updated, notes_updated, subscriptionName, vhdUri, guestDiskUri),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "name", name_updated),
					resource.TestCheckResourceAttr("citrix_quickdeploy_template_image.test_template_image", "notes", notes_updated),
				),
			},
		},
	})
}

func BuildCitrixManagedAzureImageResource(t *testing.T, name, note, subscriptionName, vhdUri, guestDiskUri string) string {
	return fmt.Sprintf(testCitrixManagedImageResource, name, note, subscriptionName, vhdUri, guestDiskUri)
}

var (
	testCitrixManagedImageResource = `
	resource "citrix_quickdeploy_template_image" "test_template_image" {
		name = "%s"
		notes = "%s"
		subscription_name = "%s"
		region = "East US"
		vhd_uri = "%s"
		machine_generation = "V2"
		os_platform = "Windows"
		vtpm_enabled = true
		secure_boot_enabled = true
		guest_disk_uri = "%s"
	}
	`
)
