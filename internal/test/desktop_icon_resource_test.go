package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestDesktopIconPreCheck validates the necessary env variable exist
// in the testing environment
func TestDesktopIconPreCheck(t *testing.T) {
	if v := os.Getenv("TEST_DESKTOP_ICON_RAW_DATA"); v == "" {
		t.Fatal("TEST_DESKTOP_ICON_RAW_DATA must be set for acceptance tests")
	}
}

func TestDesktopIconResource(t *testing.T) {
	desktopIconRawData := os.Getenv("TEST_DESKTOP_ICON_RAW_DATA")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestDesktopIconPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: BuildDesktopIconResource(t, testDesktopIconResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify raw data of desktop icon
					resource.TestCheckResourceAttr("citrix_desktop_icon.testDesktopIcon", "raw_data", desktopIconRawData),
				),
			},
			// ImportState testing
			{
				ResourceName:      "citrix_desktop_icon.testDesktopIcon",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

var (
	testDesktopIconResource = `
resource "citrix_desktop_icon" "testDesktopIcon" {
	raw_data = "%s"
}
`
)

func BuildDesktopIconResource(t *testing.T, desktopIcon string) string {
	desktopIconRawData := os.Getenv("TEST_DESKTOP_ICON_RAW_DATA")
	return fmt.Sprintf(desktopIcon, desktopIconRawData)
}
