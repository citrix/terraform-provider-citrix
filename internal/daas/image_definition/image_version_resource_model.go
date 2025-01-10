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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	SessionSupport  types.String `tfsdk:"session_support"`
	OsType          types.String `tfsdk:"os_type"`
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

func (ImageVersionModel) GetAttributes() map[string]schema.Attribute {
	return ImageVersionModel{}.GetSchema().Attributes
}

func (r ImageVersionModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, imageVersion *citrixorchestration.ImageVersionResponseModel) ImageVersionModel {
	r, imageSpecs, specConfigured := r.RefreshImageVersionBaseProperties(ctx, diagnostics, imageVersion)
	if specConfigured {
		return r
	}
	imageContext := imageSpecs.GetContext()

	resourcePool := imageSpecs.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	r.Hypervisor = types.StringValue(hypervisor.GetId())
	r.ResourcePool = types.StringValue(resourcePool.GetId())

	imageRuntimeEnvironment := imageSpecs.GetImageRuntimeEnvironment()
	r.SessionSupport = types.StringValue(imageRuntimeEnvironment.GetVDASessionSupport())
	vdaOS := imageRuntimeEnvironment.GetOperatingSystem()
	r.OsType = types.StringValue(vdaOS.GetType())

	masterImage := imageSpecs.GetMasterImage()

	switch imageContext.GetPluginFactoryName() {
	case util.AZURERM_FACTORY_NAME:
		imageScheme := imageContext.GetImageScheme()
		azureImageSpecs := util.ObjectValueToTypedObject[AzureImageSpecsModel](ctx, diagnostics, r.AzureImageSpecs)

		azureImageSpecs.ServiceOffering = parseAzureImageVersionServiceOffering(imageScheme.GetServiceOffering())

		licenseType, storageType, des, err := parseAzureImageCustomProperties(ctx, diagnostics, true, imageScheme.GetCustomProperties(), azureImageSpecs.DiskEncryptionSet)
		if err != nil {
			return r
		}
		azureImageSpecs.LicenseType = licenseType
		azureImageSpecs.StorageType = storageType
		azureImageSpecs.DiskEncryptionSet = des

		updatedMachineProfile, err := refreshAzureImageVersionMachineProfile(ctx, diagnostics, true, imageScheme)
		if err == nil {
			azureImageSpecs.MachineProfile = updatedMachineProfile
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

func (r ImageVersionModel) RefreshImageVersionBaseProperties(ctx context.Context, diagnostics *diag.Diagnostics, imageVersion *citrixorchestration.ImageVersionResponseModel) (ImageVersionModel, citrixorchestration.ImageVersionSpecResponseModel, bool) {
	r.Id = types.StringValue(imageVersion.GetId())
	r.VersionNumber = types.Int32Value(imageVersion.GetNumber())
	imageDefinition := imageVersion.GetImageDefinition()
	r.ImageDefinition = types.StringValue(imageDefinition.GetId())
	r.Description = types.StringValue(imageVersion.GetDescription())

	imageSpecs, specConfigured := identifyImageVersionSpec(diagnostics, imageVersion.GetImageVersionSpecs())
	if !specConfigured {
		return r, imageSpecs, false
	}

	resourcePool := imageSpecs.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	r.Hypervisor = types.StringValue(hypervisor.GetId())
	r.ResourcePool = types.StringValue(resourcePool.GetId())

	imageRuntimeEnvironment := imageSpecs.GetImageRuntimeEnvironment()
	r.SessionSupport = types.StringValue(imageRuntimeEnvironment.GetVDASessionSupport())
	vdaOS := imageRuntimeEnvironment.GetOperatingSystem()
	r.OsType = types.StringValue(vdaOS.GetType())
	return r, imageSpecs, false
}

func refreshAzureImageVersionMachineProfile(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, imageScheme citrixorchestration.ImageSchemeResponseModel) (types.Object, error) {
	var machineProfileToReturn types.Object

	if imageScheme.MachineProfile != nil {
		// Refresh machine profile
		machineProfile := imageScheme.GetMachineProfile()
		machineProfileModel := util.ParseAzureMachineProfileResponseToModel(machineProfile)
		if isResource {
			machineProfileToReturn = util.TypedObjectToObjectValue(ctx, diagnostics, machineProfileModel)
		} else {
			machineProfileToReturn = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, machineProfileModel)
		}
		return machineProfileToReturn, nil
	} else {
		var attributesMap map[string]attr.Type
		var err error
		if isResource {
			attributesMap, err = util.ResourceAttributeMapFromObject(util.AzureMachineProfileModel{})
		} else {
			attributesMap, err = util.DataSourceAttributeMapFromObject(util.AzureMachineProfileModel{})
		}
		if err != nil {
			diagnostics.AddWarning("Error when creating null AzureMachineProfileModel", err.Error())
			return machineProfileToReturn, err
		}
		machineProfileToReturn = types.ObjectNull(attributesMap)
		return machineProfileToReturn, err
	}
}

func identifyImageVersionSpec(diagnostics *diag.Diagnostics, imageVersionSpecs []citrixorchestration.ImageVersionSpecResponseModel) (citrixorchestration.ImageVersionSpecResponseModel, bool) {
	for _, spec := range imageVersionSpecs {
		if spec.Context != nil {
			context := spec.GetContext()
			if context.ImageScheme == nil {
				continue
			}
			return spec, true
		}
	}
	diagnostics.AddError(
		"Error refreshing Image Version",
		"Image Version does not have image context configured",
	)
	return citrixorchestration.ImageVersionSpecResponseModel{}, false
}

func parseAzureImageVersionServiceOffering(serviceOfferingXdPath string) types.String {
	serviceOfferingSegments := strings.Split(serviceOfferingXdPath, "\\")
	serviceOfferingLastIndex := len(serviceOfferingSegments)
	serviceOffering := strings.TrimSuffix(serviceOfferingSegments[serviceOfferingLastIndex-1], ".serviceoffering")
	return types.StringValue(serviceOffering)
}

func parseAzureImageCustomProperties(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, customProperties []citrixorchestration.NameValueStringPairModel, des types.Object) (types.String, types.String, types.Object, error) {
	// Set initial values before refreshing the custom properties\
	licenseType := types.StringNull()
	storageType := types.StringNull()
	var attributeMap map[string]attr.Type
	var err error
	if isResource {
		attributeMap, err = util.ResourceAttributeMapFromObject(util.AzureDiskEncryptionSetModel{})
	} else {
		attributeMap, err = util.DataSourceAttributeMapFromObject(util.AzureDiskEncryptionSetModel{})
	}
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return licenseType, storageType, des, err
	}
	des = types.ObjectNull(attributeMap)
	for _, customerProperty := range customProperties {
		if strings.EqualFold(customerProperty.GetName(), "LicenseType") && customerProperty.GetValue() != "" {
			licenseType = types.StringValue(customerProperty.GetValue())
		} else if strings.EqualFold(customerProperty.GetName(), "StorageType") && customerProperty.GetValue() != "" {
			storageType = types.StringValue(customerProperty.GetValue())
		} else if strings.EqualFold(customerProperty.GetName(), "DiskEncryptionSetId") && customerProperty.GetValue() != "" {
			diskEncryptionSetModel := util.AzureDiskEncryptionSetModel{}
			desName, desResourceGroup := util.ParseDiskEncryptionSetIdToNameAndResourceGroup(customerProperty.GetValue())
			diskEncryptionSetModel.DiskEncryptionSetResourceGroup = types.StringValue(strings.ToLower(desResourceGroup))
			diskEncryptionSetModel.DiskEncryptionSetName = types.StringValue(strings.ToLower(desName))
			if isResource {
				des = util.TypedObjectToObjectValue(ctx, diagnostics, diskEncryptionSetModel)
			} else {
				des = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, diskEncryptionSetModel)
			}
		}
	}
	return licenseType, storageType, des, nil
}
