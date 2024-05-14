// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

type AzureMachineConfigModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	/** Azure Hypervisor **/
	AzureMasterImage       *AzureMasterImageModel       `tfsdk:"azure_master_image"`
	VdaResourceGroup       types.String                 `tfsdk:"vda_resource_group"`
	StorageType            types.String                 `tfsdk:"storage_type"`
	UseAzureComputeGallery *AzureComputeGallerySettings `tfsdk:"use_azure_compute_gallery"`
	LicenseType            types.String                 `tfsdk:"license_type"`
	UseManagedDisks        types.Bool                   `tfsdk:"use_managed_disks"`
	MachineProfile         *AzureMachineProfileModel    `tfsdk:"machine_profile"`
	WritebackCache         *AzureWritebackCacheModel    `tfsdk:"writeback_cache"`
	DiskEncryptionSet      *AzureDiskEncryptionSetModel `tfsdk:"disk_encryption_set"`
	EnrollInIntune         types.Bool                   `tfsdk:"enroll_in_intune"`
}

type AwsMachineConfigModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	MasterImage     types.String `tfsdk:"master_image"`
	/** AWS Hypervisor **/
	ImageAmi       types.String   `tfsdk:"image_ami"`
	SecurityGroups []types.String `tfsdk:"security_groups"`
	TenancyType    types.String   `tfsdk:"tenancy_type"`
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

type NutanixMachineConfigModel struct {
	Container        types.String `tfsdk:"container"`
	MasterImage      types.String `tfsdk:"master_image"`
	CpuCount         types.Int64  `tfsdk:"cpu_count"`
	CoresPerCpuCount types.Int64  `tfsdk:"cores_per_cpu_count"`
	MemoryMB         types.Int64  `tfsdk:"memory_mb"`
}

type GalleryImageModel struct {
	Gallery    types.String `tfsdk:"gallery"`
	Definition types.String `tfsdk:"definition"`
	Version    types.String `tfsdk:"version"`
}

type AzureMasterImageModel struct {
	ResourceGroup      types.String       `tfsdk:"resource_group"`
	SharedSubscription types.String       `tfsdk:"shared_subscription"`
	MasterImage        types.String       `tfsdk:"master_image"`
	StorageAccount     types.String       `tfsdk:"storage_account"`
	Container          types.String       `tfsdk:"container"`
	GalleryImage       *GalleryImageModel `tfsdk:"gallery_image"`
}

type AzureMachineProfileModel struct {
	MachineProfileVmName        types.String `tfsdk:"machine_profile_vm_name"`
	MachineProfileResourceGroup types.String `tfsdk:"machine_profile_resource_group"`
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

type AzureComputeGallerySettings struct {
	ReplicaRatio   types.Int64 `tfsdk:"replica_ratio"`
	ReplicaMaximum types.Int64 `tfsdk:"replica_maximum"`
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

type AzureDiskEncryptionSetModel struct {
	DiskEncryptionSetName          types.String `tfsdk:"disk_encryption_set_name"`
	DiskEncryptionSetResourceGroup types.String `tfsdk:"disk_encryption_set_resource_group"`
}

func (mc *AzureMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	if provScheme.GetServiceOffering() != "" {
		mc.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	}

	// Refresh Master Image
	masterImage := provScheme.GetMasterImage()
	if mc.AzureMasterImage == nil {
		mc.AzureMasterImage = &AzureMasterImageModel{}
	}

	if mc.AzureMasterImage.GalleryImage != nil {
		/* For Azure Image Gallery image, the XDPath looks like:
		 * XDHyp:\\HostingUnits\\{resource pool}\\image.folder\\{resource group}.resourcegroup\\{gallery name}.gallery\\{image name}.imagedefinition\\{image version}.imageversion
		 * The Name property in MasterImage will be image version instead of image definition (name of the image)
		 */
		mc.AzureMasterImage.GalleryImage.Version = types.StringValue(masterImage.GetName())
	} else {
		mc.AzureMasterImage.MasterImage = types.StringValue(masterImage.GetName())
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
				mc.AzureMasterImage.Container = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				mc.AzureMasterImage.StorageAccount = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
			} else if strings.EqualFold(resourceType, util.ImageVersionResourceType) {
				// Gallery image
				if mc.AzureMasterImage.GalleryImage == nil {
					mc.AzureMasterImage.GalleryImage = &GalleryImageModel{}
				}
				mc.AzureMasterImage.GalleryImage.Definition = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				mc.AzureMasterImage.GalleryImage.Gallery = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
			}
			mc.AzureMasterImage.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-4], ".")[0])
		} else {
			// Snapshot or Managed Disk
			mc.AzureMasterImage.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
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

	if provScheme.GetDeviceManagementType() == citrixorchestration.DEVICEMANAGEMENTTYPE_INTUNE {
		mc.EnrollInIntune = types.BoolValue(true)
	} else if mc.EnrollInIntune.ValueBool() {
		mc.EnrollInIntune = types.BoolValue(false)
	}

	//Refresh custom properties
	customProperties := provScheme.GetCustomProperties()
	isLicenseTypeSet := false
	isDesSet := false
	isUseSharedImageGallerySet := false
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
		case "LicenseType":
			licenseType := stringPair.GetValue()
			if licenseType == "" {
				mc.LicenseType = types.StringNull()
			} else {
				mc.LicenseType = types.StringValue(licenseType)
			}
			isLicenseTypeSet = true
		case "DiskEncryptionSetId":
			desId := stringPair.GetValue()
			desArray := strings.Split(desId, "/")
			desName := desArray[len(desArray)-1]
			resourceGroupsIndex := slices.Index(desArray, "resourceGroups")
			resourceGroupName := desArray[resourceGroupsIndex+1]
			if mc.DiskEncryptionSet == nil {
				mc.DiskEncryptionSet = &AzureDiskEncryptionSetModel{}
			}
			if !strings.EqualFold(mc.DiskEncryptionSet.DiskEncryptionSetName.ValueString(), desName) {
				mc.DiskEncryptionSet.DiskEncryptionSetName = types.StringValue(desName)
			}
			if !strings.EqualFold(mc.DiskEncryptionSet.DiskEncryptionSetResourceGroup.ValueString(), resourceGroupName) {
				mc.DiskEncryptionSet.DiskEncryptionSetResourceGroup = types.StringValue(resourceGroupName)
			}
			isDesSet = true
		case "SharedImageGalleryReplicaRatio":
			if stringPair.GetValue() != "" {
				isUseSharedImageGallerySet = true
				if mc.UseAzureComputeGallery == nil {
					mc.UseAzureComputeGallery = &AzureComputeGallerySettings{}
				}

				replicaRatio, _ := strconv.Atoi(stringPair.GetValue())
				mc.UseAzureComputeGallery.ReplicaRatio = types.Int64Value(int64(replicaRatio))
			}
		case "SharedImageGalleryReplicaMaximum":
			if stringPair.GetValue() != "" {
				isUseSharedImageGallerySet = true
				if mc.UseAzureComputeGallery == nil {
					mc.UseAzureComputeGallery = &AzureComputeGallerySettings{}
				}

				replicaMaximum, _ := strconv.Atoi(stringPair.GetValue())
				mc.UseAzureComputeGallery.ReplicaMaximum = types.Int64Value(int64(replicaMaximum))
			}
		default:
		}
	}

	if !isLicenseTypeSet && !mc.LicenseType.IsNull() {
		mc.LicenseType = types.StringNull()
	}

	if !isDesSet && mc.DiskEncryptionSet != nil {
		mc.DiskEncryptionSet = nil
	}

	if !isUseSharedImageGallerySet && mc.UseAzureComputeGallery != nil {
		mc.UseAzureComputeGallery = nil
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

	// Refresh Security Group
	securityGroups := provScheme.GetSecurityGroups()
	mc.SecurityGroups = util.ConvertPrimitiveStringArrayToBaseStringArray(securityGroups)

	// Refresh Tenancy Type
	tenancyType := provScheme.GetTenancyType()
	mc.TenancyType = types.StringValue(tenancyType)
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

func (mc *NutanixMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	// Refresh Master Image
	masterImage := provScheme.GetMasterImage()
	mc.MasterImage = types.StringValue(masterImage.GetName())

	// Refresh Memory
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))
	mc.CoresPerCpuCount = types.Int64Value(int64(provScheme.GetCoresPerCpuCount()))
	mc.Container = types.StringValue(provScheme.GetNutanixContainer())
}

func parseAzureMachineProfileResponseToModel(machineProfileResponse citrixorchestration.HypervisorResourceRefResponseModel) *AzureMachineProfileModel {
	machineProfileModel := AzureMachineProfileModel{}
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
