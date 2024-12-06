// Copyright © 2024. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAzureImageVersionPreCheck(t *testing.T) {
	checkTestEnvironmentVariables(t, imageVersionTestVariables)
}

func TestAzureImageVersionResource(t *testing.T) {
	AzureImageVersionResourceTestHelper(t, false)
}

func TestAzureImageVersionResourcePre121(t *testing.T) {
	AzureImageVersionResourceTestHelper(t, true)
}

func AzureImageVersionResourceTestHelper(t *testing.T, pre121 bool) {
	hypervisor := os.Getenv("TEST_AZURE_IMAGE_VERSION_HYPERVISOR")
	resourcePool := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL")
	description := os.Getenv("TEST_AZURE_IMAGE_VERSION_DESCRIPTION")
	descriptionUpdated := description + "-updated"
	serviceOffering := os.Getenv("TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING")
	resourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP")
	masterImage := os.Getenv("TEST_AZURE_IMAGE_VERSION_MASTER_IMAGE")
	gallery := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY")
	galleryDefinition := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_DEFINITION")
	galleryVersion := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_VERSION")
	mpResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_RESOURCE_GROUP")
	mpName := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_VM_NAME")
	desName := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_NAME")
	desResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_RESOURCE_GROUP")

	importStateVerifyIgnore := []string{"azure_image_specs.network_mapping"}

	var imageDefinitionResource string
	if pre121 {
		imageDefinitionResource = BuildImageDefinitionTestResourcePre121(t)
	} else {
		imageDefinitionResource = BuildAzureImageDefinitionTestResource(t)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			TestProviderPreCheck(t)
			TestAzureImageDefinitionResourcePreCheck(t)
			TestAzureImageVersionPreCheck(t)
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionBasicMasterImage(t, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", description),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image", masterImage),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image"),
					// Verify machine_profile is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile"),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_image_version.test_azure_image_version",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
				ImportStateIdFunc:       generateAzureImageVersionImportStateId,
			},
			// Update and Read testing
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionBasicMasterImage(t, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", descriptionUpdated),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image", masterImage),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image"),
					// Verify machine_profile is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile"),
					// Verify disk_encryption_set is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set"),
				),
			},
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionFullMasterImage(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", description),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image", masterImage),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image"),
					// Verify machine_profile.machine_profile_resource_group
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile.machine_profile_resource_group", mpResourceGroup),
					// Verify machine_profile.machine_profile_vm_name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile.machine_profile_vm_name", mpName),
					// Verify disk_encryption_set.disk_encryption_set_name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set.disk_encryption_set_name", desName),
					// Verify disk_encryption_set.disk_encryption_set_resource_group
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set.disk_encryption_set_resource_group", desResourceGroup),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_image_version.test_azure_image_version",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
				ImportStateIdFunc:       generateAzureImageVersionImportStateId,
			},
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionBasicGalleryImage(t, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", description),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image"),
					// Verify the image gallery name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.gallery", gallery),
					// Verify the image gallery definition
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.definition", galleryDefinition),
					// Verify the image gallery version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.version", galleryVersion),
					// Verify machine_profile is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile"),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "citrix_image_version.test_azure_image_version",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
				ImportStateIdFunc:       generateAzureImageVersionImportStateId,
			},
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionBasicGalleryImage(t, true),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", descriptionUpdated),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image"),
					// Verify the image gallery name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.gallery", gallery),
					// Verify the image gallery definition
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.definition", galleryDefinition),
					// Verify the image gallery version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.version", galleryVersion),
					// Verify machine_profile is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile"),
					// Verify gallery_image is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set"),
				),
			},
			{
				Config: composeTestResourceTf(
					imageDefinitionResource,
					BuildAzureImageVersionFullGalleryImage(t),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the hypervisor of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor", hypervisor),
					// Verify the resourcePool of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "hypervisor_resource_pool", resourcePool),
					// Verify the description of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "description", description),
					// Verify the service_offering of the image version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.service_offering", serviceOffering),
					// Verify the resource group of base Azure image
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.resource_group", resourceGroup),
					// Verify the master image of the image version is not set
					resource.TestCheckNoResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.master_image"),
					// Verify the image gallery name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.gallery", gallery),
					// Verify the image gallery definition
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.definition", galleryDefinition),
					// Verify the image gallery version
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.gallery_image.version", galleryVersion),
					// Verify machine_profile.machine_profile_resource_group
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile.machine_profile_resource_group", mpResourceGroup),
					// Verify machine_profile.machine_profile_vm_name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.machine_profile.machine_profile_vm_name", mpName),
					// Verify disk_encryption_set.disk_encryption_set_name
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set.disk_encryption_set_name", desName),
					// Verify disk_encryption_set.disk_encryption_set_resource_group
					resource.TestCheckResourceAttr("citrix_image_version.test_azure_image_version", "azure_image_specs.disk_encryption_set.disk_encryption_set_resource_group", desResourceGroup),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func generateAzureImageVersionImportStateId(state *terraform.State) (string, error) {
	resourceName := "citrix_image_version.test_azure_image_version"
	var rawState map[string]string
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[resourceName]; ok {
				rawState = v.Primary.Attributes
			}
		}
	}

	return fmt.Sprintf("%s,%s", rawState["image_definition"], rawState["id"]), nil
}

func BuildAzureImageVersionBasicMasterImage(t *testing.T, updated bool) string {
	hypervisor := os.Getenv("TEST_AZURE_IMAGE_VERSION_HYPERVISOR")
	resourcePool := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL")
	description := os.Getenv("TEST_AZURE_IMAGE_VERSION_DESCRIPTION")
	if updated {
		description += "-updated"
	}
	serviceOffering := os.Getenv("TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING")
	network := os.Getenv("TEST_AZURE_IMAGE_VERSION_SUBNET")
	resourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP")
	masterImage := os.Getenv("TEST_AZURE_IMAGE_VERSION_MASTER_IMAGE")

	return fmt.Sprintf(azureImageVersionTestResource_basicMasterImage, hypervisor, resourcePool, description, serviceOffering, network, resourceGroup, masterImage)
}

func BuildAzureImageVersionFullMasterImage(t *testing.T) string {
	hypervisor := os.Getenv("TEST_AZURE_IMAGE_VERSION_HYPERVISOR")
	resourcePool := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL")
	description := os.Getenv("TEST_AZURE_IMAGE_VERSION_DESCRIPTION")
	serviceOffering := os.Getenv("TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING")
	network := os.Getenv("TEST_AZURE_IMAGE_VERSION_SUBNET")
	resourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP")
	masterImage := os.Getenv("TEST_AZURE_IMAGE_VERSION_MASTER_IMAGE")

	machineProfileResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_RESOURCE_GROUP")
	machineProfileVmName := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_VM_NAME")

	desName := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_NAME")
	desResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_RESOURCE_GROUP")

	return fmt.Sprintf(azureImageVersionTestResource_fullMasterImage, hypervisor, resourcePool, description, serviceOffering, network, resourceGroup, masterImage, machineProfileResourceGroup, machineProfileVmName, desName, desResourceGroup)
}

func BuildAzureImageVersionBasicGalleryImage(t *testing.T, updated bool) string {
	hypervisor := os.Getenv("TEST_AZURE_IMAGE_VERSION_HYPERVISOR")
	resourcePool := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL")
	description := os.Getenv("TEST_AZURE_IMAGE_VERSION_DESCRIPTION")
	if updated {
		description += "-updated"
	}
	serviceOffering := os.Getenv("TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING")
	network := os.Getenv("TEST_AZURE_IMAGE_VERSION_SUBNET")
	resourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP")
	gallery := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY")
	galleryDefinition := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_DEFINITION")
	galleryVersion := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_VERSION")

	return fmt.Sprintf(azureImageVersionTestResource_basicGalleryImage, hypervisor, resourcePool, description, serviceOffering, network, resourceGroup, gallery, galleryDefinition, galleryVersion)
}

func BuildAzureImageVersionFullGalleryImage(t *testing.T) string {
	hypervisor := os.Getenv("TEST_AZURE_IMAGE_VERSION_HYPERVISOR")
	resourcePool := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL")
	description := os.Getenv("TEST_AZURE_IMAGE_VERSION_DESCRIPTION")
	serviceOffering := os.Getenv("TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING")
	network := os.Getenv("TEST_AZURE_IMAGE_VERSION_SUBNET")
	resourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP")
	gallery := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY")
	galleryDefinition := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_DEFINITION")
	galleryVersion := os.Getenv("TEST_AZURE_IMAGE_VERSION_GALLERY_VERSION")

	machineProfileResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_RESOURCE_GROUP")
	machineProfileVmName := os.Getenv("TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_VM_NAME")

	desName := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_NAME")
	desResourceGroup := os.Getenv("TEST_AZURE_IMAGE_VERSION_DES_RESOURCE_GROUP")

	return fmt.Sprintf(azureImageVersionTestResource_fullGalleryImage, hypervisor, resourcePool, description, serviceOffering, network, resourceGroup, gallery, galleryDefinition, galleryVersion, machineProfileResourceGroup, machineProfileVmName, desName, desResourceGroup)
}

var (
	imageVersionTestVariables = []string{
		"TEST_AZURE_IMAGE_VERSION_HYPERVISOR",
		"TEST_AZURE_IMAGE_VERSION_RESOURCE_POOL",
		"TEST_AZURE_IMAGE_VERSION_DESCRIPTION",
		"TEST_AZURE_IMAGE_VERSION_SERVICE_OFFERING",
		"TEST_AZURE_IMAGE_VERSION_SUBNET",
		"TEST_AZURE_IMAGE_VERSION_RESOURCE_GROUP",
		"TEST_AZURE_IMAGE_VERSION_MASTER_IMAGE",
		"TEST_AZURE_IMAGE_VERSION_GALLERY",
		"TEST_AZURE_IMAGE_VERSION_GALLERY_DEFINITION",
		"TEST_AZURE_IMAGE_VERSION_GALLERY_VERSION",
		"TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_RESOURCE_GROUP",
		"TEST_AZURE_IMAGE_VERSION_MACHINE_PROFILE_VM_NAME",
		"TEST_AZURE_IMAGE_VERSION_DES_NAME",
		"TEST_AZURE_IMAGE_VERSION_DES_RESOURCE_GROUP",
	}

	azureImageVersionTestResource_basicMasterImage = `
resource "citrix_image_version" "test_azure_image_version" {
    image_definition = citrix_image_definition.test_image_definition.id
	hypervisor = "%s"
	hypervisor_resource_pool = "%s" 
	description = "%s"
	azure_image_specs = {
		service_offering = "%s"
		storage_type = "StandardSSD_LRS"
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		resource_group = "%s"
		master_image = "%s"
	}
}
`

	azureImageVersionTestResource_fullMasterImage = `
resource "citrix_image_version" "test_azure_image_version" {
    image_definition = citrix_image_definition.test_image_definition.id
	hypervisor = "%s"
	hypervisor_resource_pool = "%s" 
	description = "%s"
	azure_image_specs = {
		service_offering = "%s"
		storage_type = "StandardSSD_LRS"
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		resource_group = "%s"
		master_image = "%s"
		machine_profile = {
            machine_profile_resource_group = "%s"
            machine_profile_vm_name = "%s"
        }
        disk_encryption_set = {
            disk_encryption_set_name           = "%s"
            disk_encryption_set_resource_group = "%s"
        }
	}
}
`

	azureImageVersionTestResource_basicGalleryImage = `
resource "citrix_image_version" "test_azure_image_version" {
    image_definition = citrix_image_definition.test_image_definition.id
	hypervisor = "%s"
	hypervisor_resource_pool = "%s" 
	description = "%s"
	azure_image_specs = {
		service_offering = "%s"
		storage_type = "StandardSSD_LRS"
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		resource_group = "%s"
		gallery_image = {
        	gallery    = "%s"
            definition = "%s"
            version    = "%s"
        }
	}
}
`

	azureImageVersionTestResource_fullGalleryImage = `
resource "citrix_image_version" "test_azure_image_version" {
    image_definition = citrix_image_definition.test_image_definition.id
	hypervisor = "%s"
	hypervisor_resource_pool = "%s" 
	description = "%s"
	azure_image_specs = {
		service_offering = "%s"
		storage_type = "StandardSSD_LRS"
		network_mapping = [
			{
				network_device = "0"
				network 	   = "%s"
			}
		]
		resource_group = "%s"
		gallery_image = {
        	gallery    = "%s"
            definition = "%s"
            version    = "%s"
        }
		machine_profile = {
            machine_profile_resource_group = "%s"
            machine_profile_vm_name = "%s"
        }
		disk_encryption_set = {
            disk_encryption_set_name           = "%s"
            disk_encryption_set_resource_group = "%s"
        }
	}
}
`
)
