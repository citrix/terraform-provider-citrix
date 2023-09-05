package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestZoneResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { TestOnPremProviderPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "citrix_daas_zone" "test" {
	name        = "second zone"
    description = "description for go test zone"
    metadata    = [
        {
            name = "key1",
            value = "value1"
        },
        {
            name = "key2",
            value = "value2"
        },
        {
            name = "key3",
            value = "value3"
        }
    ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", "second zone"),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", "description for go test zone"),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "3"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.name", "key1"),
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.0.value", "value1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_daas_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the Orchestration
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: `
resource "citrix_daas_zone" "test" {
	name        = "second zone - updated"
    description = "updated description for go test zone"
    metadata    = [
        {
            name = "key1",
            value = "value1"
        },
        {
            name = "key2",
            value = "value2"
        },
        {
            name = "key3",
            value = "value3"
        },
		{
            name = "key4",
            value = "value4"
        },
    ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify name of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "name", "second zone - updated"),
					// Verify description of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "description", "updated description for go test zone"),
					// Verify number of meta data of zone
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.#", "4"),
					// Verify first meta data value
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.name", "key4"),
					resource.TestCheckResourceAttr("citrix_daas_zone.test", "metadata.3.value", "value4"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
