// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureImageSpecsModel struct {
	// Required Attributes
	ServiceOffering types.String `tfsdk:"service_offering"`
	LicenseType     types.String `tfsdk:"license_type"`
	StorageType     types.String `tfsdk:"storage_type"`

	// Optional Attributes
	MachineProfile    types.Object `tfsdk:"machine_profile"`
	DiskEncryptionSet types.Object `tfsdk:"disk_encryption_set"`
	NetworkMapping    types.List   `tfsdk:"network_mapping"` // List[NetworkMappingModel]

	// Master Image Attributes
	ResourceGroup      types.String `tfsdk:"resource_group"`
	SharedSubscription types.String `tfsdk:"shared_subscription"`
	MasterImage        types.String `tfsdk:"master_image"`
	GalleryImage       types.Object `tfsdk:"gallery_image"`
}

func (AzureImageSpecsModel) GetSchema() schema.SingleNestedAttribute {
	galleryImageSchema := util.GalleryImageModel{}.GetSchema()
	galleryImageSchema.Validators = []validator.Object{
		objectvalidator.AlsoRequires(path.Expressions{
			path.MatchRelative().AtParent().AtName("resource_group"),
		}...),
	}
	return schema.SingleNestedAttribute{
		Description: "Image configuration for Azure image version.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The Azure VM Sku to use when creating machines.",
				Required:    true,
			},
			"license_type": schema.StringAttribute{
				Description: "Windows license type used to provision virtual machines in Azure at the base compute rate. License types include: `Windows_Client` and `Windows_Server`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.WindowsClientLicenseType,
						util.WindowsServerLicenseType,
					),
				},
			},
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.StandardLRS,
						util.StandardSSDLRS,
						util.Premium_LRS,
					),
				},
			},
			"machine_profile":     util.AzureMachineProfileModel{}.GetSchema(),
			"disk_encryption_set": util.AzureDiskEncryptionSetModel{}.GetSchema(),
			"network_mapping": schema.ListNestedAttribute{
				Description:  "Specifies how the attached NICs are mapped to networks.",
				Optional:     true,
				NestedObject: util.NetworkMappingModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"resource_group": schema.StringAttribute{
				Description: "The Azure Resource Group where the managed disk / snapshot for creating machines is located.",
				Required:    true,
			},
			"shared_subscription": schema.StringAttribute{
				Description: "The Azure Subscription ID where the managed disk / snapshot for creating machines is located. Only required if the image is not in the same subscription of the hypervisor.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"master_image": schema.StringAttribute{
				Description: "The name of the virtual machine snapshot or VM template that will be used. This identifies the hard disk to be used and the default values for the memory and processors. Omit this field if you want to use gallery_image.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("gallery_image")),
				},
			},
			"gallery_image": galleryImageSchema,
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}
}

func (AzureImageSpecsModel) GetAttributes() map[string]schema.Attribute {
	return AzureImageSpecsModel{}.GetSchema().Attributes
}

type ImageVersionModel struct {
	// Computed Attributes
	Id            types.String `tfsdk:"id"`
	VersionNumber types.Int32  `tfsdk:"version_number"`

	// Required Attributes
	ImageDefinition types.String `tfsdk:"image_definition"`
	Hypervisor      types.String `tfsdk:"hypervisor"`
	ResourcePool    types.String `tfsdk:"hypervisor_resource_pool"`

	// Optional Attributes
	Description     types.String `tfsdk:"description"`
	AzureImageSpecs types.Object `tfsdk:"azure_image_specs"`
}

func (ImageVersionModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an image version. **Note that this feature is in Tech Preview.**",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the image version.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_number": schema.Int32Attribute{
				Description: "The version number of the image version.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"image_definition": schema.StringAttribute{
				Description: "Id of the image definition to associate this image version with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor to use for creating this image version.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hypervisor_resource_pool": schema.StringAttribute{
				Description: "Id of the hypervisor resource pool to use for creating this image version.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the image version.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"azure_image_specs": AzureImageSpecsModel{}.GetSchema(),
		},
	}
}

func (ImageVersionModel) GetAttributes() map[string]schema.Attribute {
	return ImageVersionModel{}.GetSchema().Attributes
}

func (r ImageVersionModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, imageVersion *citrixorchestration.ImageVersionResponseModel) ImageVersionModel {
	r.Id = types.StringValue(imageVersion.GetId())
	r.VersionNumber = types.Int32Value(imageVersion.GetNumber())
	imageDefinition := imageVersion.GetImageDefinition()
	r.ImageDefinition = types.StringValue(imageDefinition.GetId())
	r.Description = types.StringValue(imageVersion.GetDescription())

	var imageContext citrixorchestration.ImageVersionSpecContextResponseModel
	var masterImage citrixorchestration.HypervisorResourceRefResponseModel
	imageContextConfigured := false
	for _, spec := range imageVersion.GetImageVersionSpecs() {
		if spec.Context != nil {
			context := spec.GetContext()
			if context.ImageScheme == nil {
				continue
			}
			masterImage = spec.GetMasterImage()
			imageContext = spec.GetContext()
			resourcePool := spec.GetResourcePool()
			hypervisor := resourcePool.GetHypervisor()
			r.Hypervisor = types.StringValue(hypervisor.GetId())
			r.ResourcePool = types.StringValue(resourcePool.GetId())
			imageContextConfigured = true
			break
		}
	}
	if !imageContextConfigured {
		diagnostics.AddError(
			"Error refreshing Image Version",
			"Image Version does not have image context configured",
		)
	}

	switch imageContext.GetPluginFactoryName() {
	case util.AZURERM_FACTORY_NAME:
		imageScheme := imageContext.GetImageScheme()
		azureImageSpecs := util.ObjectValueToTypedObject[AzureImageSpecsModel](ctx, diagnostics, r.AzureImageSpecs)

		serviceOfferingXdPath := imageScheme.GetServiceOffering()
		serviceOfferingSegments := strings.Split(serviceOfferingXdPath, "\\")
		serviceOfferingLastIndex := len(serviceOfferingSegments)
		serviceOffering := strings.TrimSuffix(serviceOfferingSegments[serviceOfferingLastIndex-1], ".serviceoffering")
		azureImageSpecs.ServiceOffering = types.StringValue(serviceOffering)

		// Set initial values before refreshing the custom properties
		azureImageSpecs.LicenseType = types.StringNull()
		azureImageSpecs.StorageType = types.StringNull()
		attributeMap, err := util.ResourceAttributeMapFromObject(util.AzureDiskEncryptionSetModel{})
		if err != nil {
			diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
			return r
		}
		azureImageSpecs.DiskEncryptionSet = types.ObjectNull(attributeMap)

		for _, customerProperty := range imageScheme.GetCustomProperties() {
			if strings.EqualFold(customerProperty.GetName(), "LicenseType") && customerProperty.GetValue() != "" {
				azureImageSpecs.LicenseType = types.StringValue(customerProperty.GetValue())
			} else if strings.EqualFold(customerProperty.GetName(), "StorageType") && customerProperty.GetValue() != "" {
				azureImageSpecs.StorageType = types.StringValue(customerProperty.GetValue())
			} else if strings.EqualFold(customerProperty.GetName(), "DiskEncryptionSetId") && customerProperty.GetValue() != "" {
				diskEncryptionSetModel := util.ObjectValueToTypedObject[util.AzureDiskEncryptionSetModel](ctx, diagnostics, azureImageSpecs.DiskEncryptionSet)
				diskEncryptionSetModel = util.RefreshDiskEncryptionSetModel(diskEncryptionSetModel, customerProperty.GetValue())
				diskEncryptionSetModel.DiskEncryptionSetResourceGroup = types.StringValue(strings.ToLower(diskEncryptionSetModel.DiskEncryptionSetResourceGroup.ValueString()))
				diskEncryptionSetModel.DiskEncryptionSetName = types.StringValue(strings.ToLower(diskEncryptionSetModel.DiskEncryptionSetName.ValueString()))
				azureImageSpecs.DiskEncryptionSet = util.TypedObjectToObjectValue(ctx, diagnostics, diskEncryptionSetModel)
			}
		}

		if imageScheme.MachineProfile != nil {
			// Refresh machine profile
			machineProfile := imageScheme.GetMachineProfile()
			machineProfileModel := util.ParseAzureMachineProfileResponseToModel(machineProfile)
			azureImageSpecs.MachineProfile = util.TypedObjectToObjectValue(ctx, diagnostics, machineProfileModel)
		} else {
			if attributesMap, err := util.ResourceAttributeMapFromObject(util.AzureMachineProfileModel{}); err == nil {
				azureImageSpecs.MachineProfile = types.ObjectNull(attributesMap)
			} else {
				diagnostics.AddWarning("Error when creating null AzureMachineProfileModel", err.Error())
			}
		}

		// Refresh NetworkMapping
		networkMaps := imageScheme.GetNetworkMaps()
		if len(networkMaps) > 0 && !azureImageSpecs.NetworkMapping.IsNull() {
			azureImageSpecs.NetworkMapping = util.RefreshListValueProperties[util.NetworkMappingModel, citrixorchestration.NetworkMapResponseModel](ctx, diagnostics, azureImageSpecs.NetworkMapping, networkMaps, util.GetOrchestrationNetworkMappingKey)
		} else {
			azureImageSpecs.NetworkMapping = util.TypedArrayToObjectList[util.NetworkMappingModel](ctx, diagnostics, nil)
		}

		azureImageSpecs = ParseMasterImageToAzureImageModel(ctx, diagnostics, azureImageSpecs, masterImage)
		r.AzureImageSpecs = util.TypedObjectToObjectValue(ctx, diagnostics, azureImageSpecs)
	default:
		diagnostics.AddError(
			"Error refreshing Image Version",
			fmt.Sprintf("Hypervisor connection type %s is not supported", imageContext.GetPluginFactoryName()),
		)
	}

	return r
}

func ParseMasterImageToAzureImageModel(ctx context.Context, diagnostics *diag.Diagnostics, azureImageSpecs AzureImageSpecsModel, masterImage citrixorchestration.HypervisorResourceRefResponseModel) AzureImageSpecsModel {
	masterImageXdPath := masterImage.GetXDPath()
	masterImageSegments := strings.Split(masterImageXdPath, "\\")
	masterImageLastIndex := len(masterImageSegments)
	masterImageResourceTag := strings.Split(masterImageSegments[masterImageLastIndex-1], ".")
	masterImageResourceType := masterImageResourceTag[len(masterImageResourceTag)-1]
	if strings.EqualFold(masterImageResourceType, util.ImageVersionResourceType) {
		azureImageSpecs.GalleryImage,
			azureImageSpecs.ResourceGroup,
			azureImageSpecs.SharedSubscription =
			util.ParseMasterImageToUpdateGalleryImageModel(ctx, diagnostics, azureImageSpecs.GalleryImage, masterImage, masterImageSegments, masterImageLastIndex)

		// Clear other master image details
		azureImageSpecs.MasterImage = types.StringNull()
	} else {
		// Snapshot or Managed Disk
		azureImageSpecs.MasterImage,
			azureImageSpecs.ResourceGroup,
			azureImageSpecs.SharedSubscription,
			azureImageSpecs.GalleryImage,
			_,
			_ =
			util.ParseMasterImageToUpdateAzureImageSpecs(ctx, diagnostics, masterImageResourceType, masterImage, masterImageSegments, masterImageLastIndex)
	}
	return azureImageSpecs
}
