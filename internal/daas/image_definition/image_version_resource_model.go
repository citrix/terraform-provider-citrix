// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VsphereImageSpecsModel struct {
	MasterImageVm  types.String `tfsdk:"master_image_vm"`
	ImageSnapshot  types.String `tfsdk:"image_snapshot"`
	CpuCount       types.Int32  `tfsdk:"cpu_count"`
	MemoryMB       types.Int32  `tfsdk:"memory_mb"`
	MachineProfile types.String `tfsdk:"machine_profile"`
	DiskSize       types.Int32  `tfsdk:"disk_size"`
}

func (VsphereImageSpecsModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Image configuration for vSphere image version.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"master_image_vm": schema.StringAttribute{
				Description: "The name of the virtual machine that will be used as master image. This property is case sensitive.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(util.NoPathRegex),
						"must not contain any path.",
					),
				},
			},
			"image_snapshot": schema.StringAttribute{
				Description: "The Snapshot of the virtual machine specified in `master_image_vm`. Specify the relative path of the snapshot. Eg: snaphost-1/snapshot-2/snapshot-3. This property is case sensitive.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu_count": schema.Int32Attribute{
				Description: "The number of processors that virtual machines created from the provisioning scheme should use.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"memory_mb": schema.Int32Attribute{
				Description: "The maximum amount of memory that virtual machines created from the provisioning scheme should use.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(4),
				},
			},
			"machine_profile": schema.StringAttribute{
				Description: "The name of the virtual machine template that will be used to identify the default value for the tags, virtual machine size, boot diagnostics and host cache property of OS disk.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = req.StateValue.IsNull() != req.ConfigValue.IsNull()
						},
						"Force replace when machine_profile is added or removed. Update is allowed only if previously set.",
						"Force replace when machine_profile is added or removed. Update is allowed only if previously set.",
					),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"disk_size": schema.Int32Attribute{
				Description: "The size of the disk in GB.",
				Computed:    true,
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}
}

func (VsphereImageSpecsModel) GetAttributes() map[string]schema.Attribute {
	return VsphereImageSpecsModel{}.GetSchema().Attributes
}

type AmazonWorkspacesCoreImageSpecsModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	MasterImage     types.String `tfsdk:"master_image"`
	ImageAmi        types.String `tfsdk:"image_ami"`
	MachineProfile  types.Object `tfsdk:"machine_profile"`
}

func (AmazonWorkspacesCoreImageSpecsModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Image configuration for Amazon Workspaces Core image version.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The AWS VM Sku to use when creating machines.",
				Required:    true,
			},
			"master_image": schema.StringAttribute{
				Description: "The name of the virtual machine image that will be used.",
				Required:    true,
			},
			"image_ami": schema.StringAttribute{
				Description: "AMI of the AWS image to be used as the template image for the machine catalog.",
				Required:    true,
			},
			"machine_profile": util.AmazonWorkspacesCoreMachineProfileModel{}.GetSchema(),
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}
}

func (AmazonWorkspacesCoreImageSpecsModel) GetAttributes() map[string]schema.Attribute {
	return AmazonWorkspacesCoreImageSpecsModel{}.GetSchema().Attributes
}

type ImageVersionModel struct {
	// Computed Attributes
	Id            types.String `tfsdk:"id"`
	VersionNumber types.Int32  `tfsdk:"version_number"`

	// Required Attributes
	ImageDefinition types.String `tfsdk:"image_definition"`
	Hypervisor      types.String `tfsdk:"hypervisor"`
	ResourcePool    types.String `tfsdk:"hypervisor_resource_pool"`
	NetworkMapping  types.List   `tfsdk:"network_mapping"` // List[NetworkMappingModel]

	// Optional Attributes
	Description                    types.String `tfsdk:"description"`
	AzureImageSpecs                types.Object `tfsdk:"azure_image_specs"`
	VsphereImageSpecs              types.Object `tfsdk:"vsphere_image_specs"`
	AmazonWorkspacesCoreImageSpecs types.Object `tfsdk:"amazon_workspaces_core_image_specs"`
	SessionSupport                 types.String `tfsdk:"session_support"`
	OsType                         types.String `tfsdk:"os_type"`
	Timeout                        types.Object `tfsdk:"timeout"`
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
			"network_mapping": schema.ListNestedAttribute{
				Description:  "Specifies how the attached NICs are mapped to networks.",
				Optional:     true,
				NestedObject: util.NetworkMappingModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the image version.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"azure_image_specs":                  util.AzureImageSpecsModel{}.GetSchema(),
			"vsphere_image_specs":                VsphereImageSpecsModel{}.GetSchema(),
			"amazon_workspaces_core_image_specs": AmazonWorkspacesCoreImageSpecsModel{}.GetSchema(),
			"session_support": schema.StringAttribute{
				Description: "Session support for the image version.",
				Computed:    true,
			},
			"os_type": schema.StringAttribute{
				Description: "The OS type of the image version.",
				Computed:    true,
			},
			"timeout": ImageVersionTimeout{}.GetSchema(),
		},
	}
}

func (ImageVersionModel) GetAttributes() map[string]schema.Attribute {
	return ImageVersionModel{}.GetSchema().Attributes
}

func (ImageVersionModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

type ImageVersionTimeout struct {
	Create types.Int32 `tfsdk:"create"`
	Delete types.Int32 `tfsdk:"delete"`
}

func getImageVersionTimeoutConfigs() util.TimeoutConfigs {
	return util.TimeoutConfigs{
		Create:        true,
		CreateDefault: 30,
		CreateMin:     5,

		Delete:        true,
		DeleteDefault: 10,
		DeleteMin:     5,
	}
}

func (ImageVersionTimeout) GetSchema() schema.SingleNestedAttribute {
	return util.GetTimeoutSchema("image version", getImageVersionTimeoutConfigs())
}

func (ImageVersionTimeout) GetAttributes() map[string]schema.Attribute {
	return ImageVersionTimeout{}.GetSchema().Attributes
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
		azureImageSpecs := util.ObjectValueToTypedObject[util.AzureImageSpecsModel](ctx, diagnostics, r.AzureImageSpecs)

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

		azureImageSpecs = util.ParseMasterImageToAzureImageModel(ctx, diagnostics, azureImageSpecs, masterImage)
		r.AzureImageSpecs = util.TypedObjectToObjectValue(ctx, diagnostics, azureImageSpecs)
	case util.VMWARE_FACTORY_NAME:
		imageScheme := imageContext.GetImageScheme()
		masterImageXdPath := masterImage.GetXDPath()
		vsphereImageSpecs := util.ObjectValueToTypedObject[VsphereImageSpecsModel](ctx, diagnostics, r.VsphereImageSpecs)
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

		r.VsphereImageSpecs = util.TypedObjectToObjectValue(ctx, diagnostics, vsphereImageSpecs)
	case util.AMAZON_WORKSPACES_CORE_FACTORY_NAME:
		amazonWorkspacesCoreImageSpecs := util.ObjectValueToTypedObject[AmazonWorkspacesCoreImageSpecsModel](ctx, diagnostics, r.AmazonWorkspacesCoreImageSpecs)
		imageScheme := imageContext.GetImageScheme()
		/* For AWS master image, the returned master image name looks like:
		* {Image Name} (ami-000123456789abcde)
		* The Name property in MasterImage will be image name without ami id appended
		 */
		amazonWorkspacesCoreImageSpecs.MasterImage = types.StringValue(strings.Split(masterImage.GetName(), " (ami-")[0])
		amazonWorkspacesCoreImageSpecs.ImageAmi = types.StringValue(strings.TrimSuffix((strings.Split(masterImage.GetName(), " (")[1]), ")"))

		if imageScheme.MachineProfile != nil {
			machineProfile := imageScheme.GetMachineProfile()
			machineProfileModel := util.ParseAwsMachineProfileResponseToModel(machineProfile)
			amazonWorkspacesCoreImageSpecs.MachineProfile = util.TypedObjectToObjectValue(ctx, diagnostics, machineProfileModel)
		} else {
			if attributesMap, err := util.ResourceAttributeMapFromObject(util.AwsMachineProfileModel{}); err == nil {
				amazonWorkspacesCoreImageSpecs.MachineProfile = types.ObjectNull(attributesMap)
			} else {
				diagnostics.AddWarning("Error when creating null AmazonWorkspacesCoreMachineProfileModel", err.Error())
			}
		}

		if imageScheme.GetServiceOffering() != "" {
			amazonWorkspacesCoreImageSpecs.ServiceOffering = types.StringValue(strings.TrimSuffix(imageScheme.GetServiceOffering(), ".serviceoffering"))
		}
		r.AmazonWorkspacesCoreImageSpecs = util.TypedObjectToObjectValue(ctx, diagnostics, amazonWorkspacesCoreImageSpecs)
	default:
		diagnostics.AddError(
			"Error refreshing Image Version",
			fmt.Sprintf("Hypervisor connection type %s is not supported", imageContext.GetPluginFactoryName()),
		)
	}

	return r
}

func parseVsphereImageXdPath(masterImageXdPath string) (string, string) {
	vmIndex := strings.Index(masterImageXdPath, ".vm")
	if vmIndex == -1 {
		return "", ""
	}
	// Extract the master image name and trim the ".vm"
	masterImageVmPath := masterImageXdPath[:vmIndex]
	masterImageVmArr := strings.Split(masterImageVmPath, "\\")
	masterImageVm := masterImageVmArr[len(masterImageVmArr)-1]

	// Extract the snapshot part of the path
	snapshotPath := masterImageXdPath[vmIndex+len(".vm/"):]
	imageSnapshot := strings.ReplaceAll(snapshotPath, ".snapshot", "")
	imageSnapshot = strings.ReplaceAll(imageSnapshot, "\\", "/")

	return masterImageVm, imageSnapshot
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

func refreshAmazonWSCImageVersionMachineProfile(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, imageScheme citrixorchestration.ImageSchemeResponseModel) (types.Object, error) {
	var machineProfileToReturn types.Object
	if imageScheme.MachineProfile != nil {
		machineProfile := imageScheme.GetMachineProfile()
		machineProfileModel := util.ParseAmazonWorkspacesCoreMachineProfileResponseToModel(machineProfile)
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
			attributesMap, err = util.ResourceAttributeMapFromObject(util.AmazonWorkspacesCoreMachineProfileModel{})
		} else {
			attributesMap, err = util.DataSourceAttributeMapFromObject(util.AmazonWorkspacesCoreMachineProfileModel{})
		}
		if err != nil {
			diagnostics.AddWarning("Error when creating null AmazonWorkspacesCoreMachineProfileModel", err.Error())
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
