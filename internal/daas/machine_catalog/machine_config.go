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
	ResourceGroup    types.String              `tfsdk:"resource_group"`
	StorageAccount   types.String              `tfsdk:"storage_account"`
	Container        types.String              `tfsdk:"container"`
	GalleryImage     *GalleryImageModel        `tfsdk:"gallery_image"`
	VdaResourceGroup types.String              `tfsdk:"vda_resource_group"`
	StorageType      types.String              `tfsdk:"storage_type"`
	UseManagedDisks  types.Bool                `tfsdk:"use_managed_disks"`
	MachineProfile   *MachineProfileModel      `tfsdk:"machine_profile"`
	WritebackCache   *AzureWritebackCacheModel `tfsdk:"writeback_cache"`
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
	MachineProfile  types.String            `tfsdk:"machine_profile"`
	MachineSnapshot types.String            `tfsdk:"machine_snapshot"`
	StorageType     types.String            `tfsdk:"storage_type"`
	WritebackCache  *GcpWritebackCacheModel `tfsdk:"writeback_cache"`
}

type VsphereMachineConfigModel struct {
	/** Vsphere Hypervisor **/
	MasterImageVm  types.String                `tfsdk:"master_image_vm"`
	ImageSnapshot  types.String                `tfsdk:"image_snapshot"`
	CpuCount       types.Int64                 `tfsdk:"cpu_count"`
	MemoryMB       types.Int64                 `tfsdk:"memory_mb"`
	WritebackCache *VsphereWritebackCacheModel `tfsdk:"writeback_cache"`
}

type XenserverMachineConfigModel struct {
	/** XenServer Hypervisor **/
	MasterImageVm  types.String                  `tfsdk:"master_image_vm"`
	ImageSnapshot  types.String                  `tfsdk:"image_snapshot"`
	CpuCount       types.Int64                   `tfsdk:"cpu_count"`
	MemoryMB       types.Int64                   `tfsdk:"memory_mb"`
	WritebackCache *XenserverWritebackCacheModel `tfsdk:"writeback_cache"`
}

// WritebackCacheModel maps the write back cacheconfiguration schema data.
type AzureWritebackCacheModel struct {
	PersistWBC                 types.Bool   `tfsdk:"persist_wbc"`
	WBCDiskStorageType         types.String `tfsdk:"wbc_disk_storage_type"`
	PersistOsDisk              types.Bool   `tfsdk:"persist_os_disk"`
	PersistVm                  types.Bool   `tfsdk:"persist_vm"`
	StorageCostSaving          types.Bool   `tfsdk:"storage_cost_saving"`
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
}

type GcpWritebackCacheModel struct {
	PersistWBC                 types.Bool   `tfsdk:"persist_wbc"`
	WBCDiskStorageType         types.String `tfsdk:"wbc_disk_storage_type"`
	PersistOsDisk              types.Bool   `tfsdk:"persist_os_disk"`
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
}

type XenserverWritebackCacheModel struct {
	WriteBackCacheDiskSizeGB   types.Int64 `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64 `tfsdk:"writeback_cache_memory_size_mb"`
}

type VsphereWritebackCacheModel struct {
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
	WriteBackCacheDriveLetter  types.String `tfsdk:"writeback_cache_drive_letter"`
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

			if strings.EqualFold(resourceType, util.VhdResourceType) {
				// VHD image
				mc.Container = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				mc.StorageAccount = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
			} else if strings.EqualFold(resourceType, util.ImageVersionResourceType) {
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
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		if mc.WritebackCache == nil {
			mc.WritebackCache = &AzureWritebackCacheModel{}
		}
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
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
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		if mc.WritebackCache == nil {
			mc.WritebackCache = &GcpWritebackCacheModel{}
		}
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
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

func (mc *VsphereMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	// Refresh Master Image
	masterImage, imageSnapshot := parseOnPremImagePath(catalog)
	mc.MasterImageVm = types.StringValue(masterImage)
	mc.ImageSnapshot = types.StringValue(imageSnapshot)

	// Refresh Memory
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		if mc.WritebackCache == nil {
			mc.WritebackCache = &VsphereWritebackCacheModel{}
		}
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			mc.WritebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		if provScheme.GetWriteBackCacheDriveLetter() != "" {
			mc.WritebackCache.WriteBackCacheDriveLetter = types.StringValue(provScheme.GetWriteBackCacheDriveLetter())
		}
	}
}

func (mc *XenserverMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))

	masterImage, imageSnapshot := parseOnPremImagePath(catalog)
	mc.MasterImageVm = types.StringValue(masterImage)
	mc.ImageSnapshot = types.StringValue(imageSnapshot)

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		if mc.WritebackCache == nil {
			mc.WritebackCache = &XenserverWritebackCacheModel{}
		}
		mc.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			mc.WritebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
	}
}

func parseAzureMachineProfileResponseToModel(machineProfileResponse citrixorchestration.HypervisorResourceRefResponseModel) *MachineProfileModel {
	machineProfileModel := MachineProfileModel{}
	if machineProfileName := machineProfileResponse.GetName(); machineProfileName != "" {
		machineProfileSegments := strings.Split(machineProfileResponse.GetXDPath(), "\\")
		lastIndex := len(machineProfileSegments) - 1
		machineProfile := machineProfileSegments[lastIndex]
		machineProfileParent := machineProfileSegments[lastIndex-1]
		if strings.HasSuffix(machineProfile, ".templatespecversion") {
			templateSpecName := strings.Split(machineProfileParent, ".")[0]
			machineProfileModel.AzureTemplateSpecName = types.StringValue(templateSpecName)
			machineProfileModel.AzureTemplateSpecVersion = types.StringValue(machineProfileName)
			machineProfileParent = machineProfileSegments[lastIndex-2]
		} else {
			machineProfileModel.MachineProfileVmName = types.StringValue(machineProfileName)
		}
		machineProfileParentType := strings.Split(machineProfileParent, ".")[1]
		if machineProfileParentType == "resourcegroup" {
			machineProfileModel.MachineProfileResourceGroup = types.StringValue(strings.Split(machineProfileParent, ".")[0])
		}
	} else {
		machineProfileModel.MachineProfileVmName = types.StringNull()
		machineProfileModel.MachineProfileResourceGroup = types.StringNull()
		machineProfileModel.AzureTemplateSpecName = types.StringNull()
		machineProfileModel.AzureTemplateSpecVersion = types.StringNull()
	}
	return &machineProfileModel
}

func parseOnPremImagePath(catalog citrixorchestration.MachineCatalogDetailResponseModel) (masterImage, imageSnapshot string) {
	provScheme := catalog.GetProvisioningScheme()
	currentDiskImage := provScheme.GetCurrentDiskImage()
	currentImage := currentDiskImage.GetImage()
	relativePath := currentImage.GetRelativePath()

	// Refresh Master Image
	/*
	 * For On-Premise snapshot image, the RelativePath looks like:
	 * {VM name}.vm/{VM snapshot name}.snapshot(/{VM snapshot name}.snapshot)*
	 * A new snapshot will be created if it was not specified. There will always be at least one snapshot in the path.
	 */
	imageSegments := strings.Split(relativePath, "/")
	masterImage = strings.Split(imageSegments[0], ".")[0]

	snapshot := strings.Split(imageSegments[1], ".")[0]
	for i := 2; i < len(imageSegments); i++ {
		snapshot = snapshot + "/" + strings.Split(imageSegments[i], ".")[0]
	}

	return masterImage, snapshot
}
