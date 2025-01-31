// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureImageSpecsDataSourceModel struct {
	// Required Attributes
	ServiceOffering types.String `tfsdk:"service_offering"`
	LicenseType     types.String `tfsdk:"license_type"`
	StorageType     types.String `tfsdk:"storage_type"`

	// Optional Attributes
	MachineProfile    types.Object `tfsdk:"machine_profile"`
	DiskEncryptionSet types.Object `tfsdk:"disk_encryption_set"`
}

func (AzureImageSpecsDataSourceModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Image configuration for Azure image version.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The Azure VM Sku to use when creating machines.",
				Computed:    true,
			},
			"license_type": schema.StringAttribute{
				Description: "Windows license type used to provision virtual machines in Azure at the base compute rate. License types include: `Windows_Client` and `Windows_Server`.",
				Computed:    true,
			},
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Computed:    true,
			},
			"machine_profile":     util.AzureMachineProfileModel{}.GetDataSourceSchema(),
			"disk_encryption_set": util.AzureDiskEncryptionSetModel{}.GetDataSourceSchema(),
		},
	}
}

func (AzureImageSpecsDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AzureImageSpecsDataSourceModel{}.GetDataSourceSchema().Attributes
}

type VsphereImageSpecsDataSourceModel struct {
	MasterImageVm  types.String `tfsdk:"master_image_vm"`
	ImageSnapshot  types.String `tfsdk:"image_snapshot"`
	CpuCount       types.Int32  `tfsdk:"cpu_count"`
	MemoryMB       types.Int32  `tfsdk:"memory_mb"`
	MachineProfile types.String `tfsdk:"machine_profile"`
	DiskSize       types.Int32  `tfsdk:"disk_size"`
}

func (VsphereImageSpecsDataSourceModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Image configuration for vSphere image version.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"master_image_vm": schema.StringAttribute{
				Description: "The name of the virtual machine that will be used as master image. This property is case sensitive.",
				Computed:    true,
			},
			"image_snapshot": schema.StringAttribute{
				Description: "The Snapshot of the virtual machine specified in `master_image_vm`. Specify the relative path of the snapshot. Eg: snaphost-1/snapshot-2/snapshot-3. This property is case sensitive.",
				Computed:    true,
			},
			"cpu_count": schema.Int32Attribute{
				Description: "The number of processors that virtual machines created from the provisioning scheme should use.",
				Computed:    true,
			},
			"memory_mb": schema.Int32Attribute{
				Description: "The maximum amount of memory that virtual machines created from the provisioning scheme should use.",
				Computed:    true,
			},
			"machine_profile": schema.StringAttribute{
				Description: "The name of the virtual machine template that will be used to identify the default value for the tags, virtual machine size, boot diagnostics and host cache property of OS disk.",
				Computed:    true,
			},
			"disk_size": schema.Int32Attribute{
				Description: "The size of the disk in GB.",
				Computed:    true,
			},
		},
	}
}

func (VsphereImageSpecsDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return VsphereImageSpecsDataSourceModel{}.GetDataSourceSchema().Attributes
}

func (ImageVersionModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source an image version. **Note that this feature is in Tech Preview.**",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the image version.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("version_number")),
				},
			},
			"image_definition": schema.StringAttribute{
				Description: "Id of the image definition to associate this image version with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"version_number": schema.Int32Attribute{
				Description: "The version number of the image version.",
				Optional:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor to use for creating this image version.",
				Computed:    true,
			},
			"hypervisor_resource_pool": schema.StringAttribute{
				Description: "Id of the hypervisor resource pool to use for creating this image version.",
				Computed:    true,
			},
			"network_mapping": schema.ListNestedAttribute{
				Description:  "Specifies how the attached NICs are mapped to networks.",
				Computed:     true,
				NestedObject: util.NetworkMappingModel{}.GetDataSourceSchema(),
			},
			"description": schema.StringAttribute{
				Description: "Description of the image version.",
				Computed:    true,
			},
			"azure_image_specs":   AzureImageSpecsDataSourceModel{}.GetDataSourceSchema(),
			"vsphere_image_specs": VsphereImageSpecsDataSourceModel{}.GetDataSourceSchema(),
			"session_support": schema.StringAttribute{
				Description: "Session support for the image version.",
				Computed:    true,
			},
			"os_type": schema.StringAttribute{
				Description: "The OS type of the image version.",
				Computed:    true,
			},
		},
	}
}

func (ImageVersionModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ImageVersionModel{}.GetDataSourceSchema().Attributes
}

func (r ImageVersionModel) RefreshDataSourcePropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, imageVersion *citrixorchestration.ImageVersionResponseModel) ImageVersionModel {
	r, imageSpecs, specConfigured := r.RefreshImageVersionBaseProperties(ctx, diagnostics, imageVersion)
	if specConfigured {
		return r
	}

	imageContext := imageSpecs.GetContext()
	masterImage := imageSpecs.GetMasterImage()
	imageScheme := imageContext.GetImageScheme()
	// Refresh NetworkMapping
	networkMaps := imageScheme.GetNetworkMaps()
	if len(networkMaps) > 0 && !r.NetworkMapping.IsNull() {
		r.NetworkMapping = util.RefreshListValueProperties[util.NetworkMappingModel, citrixorchestration.NetworkMapResponseModel](ctx, diagnostics, r.NetworkMapping, networkMaps, util.GetOrchestrationNetworkMappingKey)
	} else {
		r.NetworkMapping = util.TypedArrayToObjectList[util.NetworkMappingModel](ctx, diagnostics, nil)
	}

	switch imageContext.GetPluginFactoryName() {
	case util.AZURERM_FACTORY_NAME:
		azureImageSpecs := AzureImageSpecsDataSourceModel{}

		azureImageSpecs.ServiceOffering = parseAzureImageVersionServiceOffering(imageScheme.GetServiceOffering())

		licenseType, storageType, des, err := parseAzureImageCustomProperties(ctx, diagnostics, false, imageScheme.GetCustomProperties(), azureImageSpecs.DiskEncryptionSet)
		if err != nil {
			return r
		}
		azureImageSpecs.LicenseType = licenseType
		azureImageSpecs.StorageType = storageType
		azureImageSpecs.DiskEncryptionSet = des

		updatedMachineProfile, err := refreshAzureImageVersionMachineProfile(ctx, diagnostics, false, imageScheme)
		if err == nil {
			azureImageSpecs.MachineProfile = updatedMachineProfile
		}

		r.AzureImageSpecs = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, azureImageSpecs)
	case util.VMWARE_FACTORY_NAME:
		vsphereImageSpecs := VsphereImageSpecsDataSourceModel{}

		masterImageXdPath := masterImage.GetXDPath()
		masterImageVm, imageSnapshot := parseVsphereImageXdPath(masterImageXdPath)
		vsphereImageSpecs.MasterImageVm = types.StringValue(masterImageVm)
		vsphereImageSpecs.ImageSnapshot = types.StringValue(imageSnapshot)
		vsphereImageSpecs.CpuCount = types.Int32Value(imageScheme.GetCpuCount())
		vsphereImageSpecs.MemoryMB = types.Int32Value(imageScheme.GetMemoryMB())

		if imageScheme.MachineProfile == nil {
			vsphereImageSpecs.MachineProfile = types.StringNull()
		} else {
			machineProfile := imageScheme.GetMachineProfile()
			vsphereImageSpecs.MachineProfile = types.StringValue(machineProfile.GetName())
		}

		if imageSpecs.DiskSize != nil {
			vsphereImageSpecs.DiskSize = types.Int32Value(imageSpecs.GetDiskSize())
		} else {
			vsphereImageSpecs.DiskSize = types.Int32Null()
		}

		r.VsphereImageSpecs = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, vsphereImageSpecs)
	default:
		diagnostics.AddError(
			"Error refreshing Image Version data source",
			fmt.Sprintf("Hypervisor connection type %s is not supported", imageContext.GetPluginFactoryName()),
		)
	}

	return r
}
