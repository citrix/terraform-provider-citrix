// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureMachineConfigModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	MasterImage     types.String `tfsdk:"master_image"`
	/** Azure Hypervisor **/
	ResourceGroup       types.String              `tfsdk:"resource_group"`
	StorageAccount      types.String              `tfsdk:"storage_account"`
	Container           types.String              `tfsdk:"container"`
	GalleryImage        *GalleryImageModel        `tfsdk:"gallery_image"`
	VdaResourceGroup    types.String              `tfsdk:"vda_resource_group"`
	StorageType         types.String              `tfsdk:"storage_type"`
	UseManagedDisks     types.Bool                `tfsdk:"use_managed_disks"`
	UseEphemeralOsDisk  types.Bool                `tfsdk:"use_ephemeral_os_disk"`
	LicenseType         types.String              `tfsdk:"license_type"`
	PlaceImageInGallery *PlaceImageInGalleryModel `tfsdk:"place_image_in_gallery"`
	DiskEncryptionSetId types.String              `tfsdk:"disk_encryption_set_id"`
	MachineProfile      *MachineProfileModel      `tfsdk:"machine_profile"`
	WritebackCache      *WritebackCacheModel      `tfsdk:"writeback_cache"`
}

type AwsMachineConfigModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	MasterImage     types.String `tfsdk:"master_image"`
	/** AWS Hypervisor **/
	ImageAmi types.String `tfsdk:"image_ami"`
}

type GcpMachineConfigModel struct {
	MasterImage types.String `tfsdk:"master_image"`
	/** GCP Hypervisor **/
	MachineProfile  types.String         `tfsdk:"machine_profile"`
	MachineSnapshot types.String         `tfsdk:"machine_snapshot"`
	StorageType     types.String         `tfsdk:"storage_type"`
	WritebackCache  *WritebackCacheModel `tfsdk:"writeback_cache"`
}

func (mc *AzureMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	if provScheme.GetServiceOffering() != "" {
		mc.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	}

	// Refresh Master Image
	masterImage := provScheme.GetMasterImage()
	if mc.GalleryImage != nil {
		/* For Azure Image Gallery image, the XDPath looks like:
		 * XDHyp:\\HostingUnits\\{resource pool}\\image.folder\\{resource group}.resourcegroup\\{gallery name}.gallery\\{image name}.imagedefinition\\{image version}.imageversion
		 * The Name property in MasterImage will be image version instead of image definition (name of the image)
		 */
		mc.GalleryImage.Version = types.StringValue(masterImage.GetName())
	} else {
		mc.MasterImage = types.StringValue(masterImage.GetName())
	}

	masterImageXdPath := masterImage.GetXDPath()
	if masterImageXdPath != "" {
		segments := strings.Split(masterImage.GetXDPath(), "\\")
		lastIndex := len(segments)
		if lastIndex == 8 {
			resourceTag := strings.Split(segments[lastIndex-1], ".")
			resourceType := resourceTag[len(resourceTag)-1]

			if resourceType == "vhd" {
				// VHD image
				mc.Container = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				mc.StorageAccount = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
			} else if resourceType == "imageversion" {
				// Gallery image
				mc.GalleryImage.Definition = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				mc.GalleryImage.Gallery = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
			}
			mc.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-4], ".")[0])
		} else {
			// Snapshot or Managed Disk
			mc.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
		}
	}

	// Refresh Machine Profile
	if provScheme.MachineProfile != nil {
		machineProfile := provScheme.GetMachineProfile()
		machineProfileModel := parseAzureMachineProfileResponseToModel(machineProfile)
		mc.MachineProfile = machineProfileModel
	} else {
		mc.MachineProfile = nil
	}

	// Refresh Writeback Cache
	if mc.WritebackCache != nil {
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if !mc.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
			mc.WritebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
	}

	//Refresh custom properties
	customProperties := provScheme.GetCustomProperties()
	for _, stringPair := range customProperties {
		switch stringPair.GetName() {
		case "StorageType":
			mc.StorageType = types.StringValue(stringPair.GetValue())
		case "UseManagedDisks":
			mc.UseManagedDisks = util.StringToTypeBool(stringPair.GetValue())
		case "UseEphemeralOsDisk":
			mc.UseEphemeralOsDisk = util.StringToTypeBool(stringPair.GetValue())
		case "LicenseType":
			if !mc.LicenseType.IsNull() || stringPair.GetValue() != "" {
				mc.LicenseType = types.StringValue(stringPair.GetValue())
			}
		case "SharedImageGalleryReplicaRatio":
			if mc.PlaceImageInGallery != nil {
				mc.PlaceImageInGallery.ReplicaRatio = util.StringToTypeInt64(stringPair.GetValue())
			}
		case "SharedImageGalleryReplicaMaximum":
			if mc.PlaceImageInGallery != nil {
				mc.PlaceImageInGallery.ReplicaMaximum = util.StringToTypeInt64(stringPair.GetValue())
			}
		case "DiskEncryptionSetId":
			mc.DiskEncryptionSetId = types.StringValue(stringPair.GetValue())
		case "ResourceGroups":
			mc.VdaResourceGroup = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			if mc.WritebackCache != nil {
				mc.WritebackCache.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
			}
		case "PersistWBC":
			if mc.WritebackCache != nil {
				mc.WritebackCache.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistOsDisk":
			if mc.WritebackCache != nil {
				mc.WritebackCache.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistVm":
			if mc.WritebackCache != nil {
				mc.WritebackCache.PersistVm = util.StringToTypeBool(stringPair.GetValue())
			}
		case "StorageTypeAtShutdown":
			if mc.WritebackCache != nil {
				mc.WritebackCache.StorageCostSaving = types.BoolValue(true)
			}
		default:
		}
	}
}

func (mc *AwsMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	if provScheme.GetServiceOffering() != "" {
		mc.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	}

	// Refresh Master Image
	masterImage := provScheme.GetMasterImage()
	/* For AWS master image, the returned master image name looks like:
	 * {Image Name} (ami-000123456789abcde)
	 * The Name property in MasterImage will be image name without ami id appended
	 */
	mc.MasterImage = types.StringValue(strings.Split(masterImage.GetName(), " (ami-")[0])
}

func (mc *GcpMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	// Refresh Master Image
	/* For GCP snapshot image, the XDPath looks like:
	 * XDHyp:\\HostingUnits\\{resource pool}\\{VM name}.vm\\{VM snapshot name}.snapshot
	 * The Name property in MasterImage will be VM snapshot name instead of VM name
	 */
	masterImage := provScheme.GetMasterImage()
	masterImageXdPath := masterImage.GetXDPath()
	if masterImageXdPath != "" {
		segments := strings.Split(masterImage.GetXDPath(), "\\")
		lastIndex := len(segments)
		// Snapshot
		if lastIndex > 4 {
			// If path slices are more than 4, that means a snapshot was used for the catalog
			mc.MachineSnapshot = types.StringValue(masterImage.GetName())
			mc.MasterImage = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
		} else {
			// If path slices equals to 4, that means a VM was used for the catalog
			mc.MasterImage = types.StringValue(masterImage.GetName())
		}
	}

	// Refresh Machine Profile
	machineProfile := provScheme.GetMachineProfile()
	if machineProfileName := machineProfile.GetName(); machineProfileName != "" {
		mc.MachineProfile = types.StringValue(machineProfileName)
	}

	// Refresh Writeback Cache
	if mc.WritebackCache != nil {
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if !mc.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
			mc.WritebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
	}

	//Refresh custom properties
	customProperties := provScheme.GetCustomProperties()
	for _, stringPair := range customProperties {
		switch stringPair.GetName() {
		case "StorageType":
			mc.StorageType = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			if mc.WritebackCache != nil {
				mc.WritebackCache.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
			}
		case "PersistWBC":
			if mc.WritebackCache != nil {
				mc.WritebackCache.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistOsDisk":
			if mc.WritebackCache != nil {
				mc.WritebackCache.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
			}
		default:
		}
	}
}

func parseAzureMachineProfileResponseToModel(machineProfileResponse citrixorchestration.HypervisorResourceRefResponseModel) *MachineProfileModel {
	machineProfileModel := MachineProfileModel{}
	if machineProfileName := machineProfileResponse.GetName(); machineProfileName != "" {
		machineProfileModel.MachineProfileVmName = types.StringValue(machineProfileName)
		machineProfileSegments := strings.Split(machineProfileResponse.GetXDPath(), "\\")
		lastIndex := len(machineProfileSegments) - 1
		machineProfileParent := machineProfileSegments[lastIndex-1]
		machineProfileParentType := strings.Split(machineProfileParent, ".")[1]
		if machineProfileParentType == "resourcegroup" {
			machineProfileModel.MachineProfileResourceGroup = types.StringValue(strings.Split(machineProfileParent, ".")[0])
		}
	} else {
		machineProfileModel.MachineProfileVmName = types.StringNull()
		machineProfileModel.MachineProfileResourceGroup = types.StringNull()
	}
	return &machineProfileModel
}
