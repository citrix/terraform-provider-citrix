// Copyright © 2024. Citrix Systems, Inc.
package util

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vSphere Additional Data Constants
const CPU_COUNT_PROPERTY_NAME string = "CpuCount"

type GalleryImageModel struct {
	Gallery    types.String `tfsdk:"gallery"`
	Definition types.String `tfsdk:"definition"`
	Version    types.String `tfsdk:"version"`
}

func (GalleryImageModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Details of the Azure Image Gallery image to use for creating machines. Only Applicable to Azure Image Gallery image.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"gallery": schema.StringAttribute{
				Description: "The Azure Image Gallery where the image for creating machines is located. Only applicable to Azure Image Gallery image.",
				Required:    true,
			},
			"definition": schema.StringAttribute{
				Description: "The image definition for the image to be used in the Azure Image Gallery. Only applicable to Azure Image Gallery image.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "The image version for the image to be used in the Azure Image Gallery. Only applicable to Azure Image Gallery image.",
				Required:    true,
			},
		},
		Validators: []validator.Object{
			objectvalidator.AlsoRequires(path.Expressions{
				path.MatchRelative().AtParent().AtName("resource_group"),
			}...),
			objectvalidator.ConflictsWith(path.Expressions{
				path.MatchRelative().AtParent().AtName("storage_account"),
			}...),
			objectvalidator.ConflictsWith(path.Expressions{
				path.MatchRelative().AtParent().AtName("container"),
			}...),
			objectvalidator.ConflictsWith(path.Expressions{
				path.MatchRelative().AtParent().AtName("master_image"),
			}...),
		},
	}
}

func (GalleryImageModel) GetAttributes() map[string]schema.Attribute {
	return GalleryImageModel{}.GetSchema().Attributes
}

type AzureDiskEncryptionSetModel struct {
	DiskEncryptionSetName          types.String `tfsdk:"disk_encryption_set_name"`
	DiskEncryptionSetResourceGroup types.String `tfsdk:"disk_encryption_set_resource_group"`
}

func (AzureDiskEncryptionSetModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The configuration for Disk Encryption Set (DES). The DES must be in the same subscription and region as your resources. If your master image is encrypted with a DES, use the same DES when creating this machine catalog. When using a DES, if you later disable the key with which the corresponding DES is associated in Azure, you can no longer power on the machines in this catalog or add machines to it.",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"disk_encryption_set_name": schema.StringAttribute{
				Description: "The name of the disk encryption set.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(LowerCaseRegex),
						"must be all in lowercase",
					),
				},
			},
			"disk_encryption_set_resource_group": schema.StringAttribute{
				Description: "The name of the resource group in which the disk encryption set resides.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(LowerCaseRegex),
						"must be all in lowercase",
					),
				},
			},
		},
	}
}

func (AzureDiskEncryptionSetModel) GetDataSourceSchema() dataSourceSchema.SingleNestedAttribute {
	return dataSourceSchema.SingleNestedAttribute{
		Description: "The configuration for Disk Encryption Set (DES). The DES must be in the same subscription and region as your resources. If your master image is encrypted with a DES, use the same DES when creating this machine catalog. When using a DES, if you later disable the key with which the corresponding DES is associated in Azure, you can no longer power on the machines in this catalog or add machines to it.",
		Computed:    true,
		Attributes: map[string]dataSourceSchema.Attribute{
			"disk_encryption_set_name": dataSourceSchema.StringAttribute{
				Description: "The name of the disk encryption set.",
				Computed:    true,
			},
			"disk_encryption_set_resource_group": dataSourceSchema.StringAttribute{
				Description: "The name of the resource group in which the disk encryption set resides.",
				Computed:    true,
			},
		},
	}
}

func (AzureDiskEncryptionSetModel) GetAttributes() map[string]schema.Attribute {
	return AzureDiskEncryptionSetModel{}.GetSchema().Attributes
}

func (AzureDiskEncryptionSetModel) GetDataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return AzureDiskEncryptionSetModel{}.GetDataSourceSchema().Attributes
}

func ParseDiskEncryptionSetIdToNameAndResourceGroup(desId string) (string, string) {
	desArray := strings.Split(desId, "/")
	desName := desArray[len(desArray)-1]
	resourceGroupsIndex := slices.Index(desArray, "resourceGroups")
	resourceGroupName := desArray[resourceGroupsIndex+1]
	return strings.ToLower(desName), strings.ToLower(resourceGroupName)
}

func RefreshDiskEncryptionSetModel(diskEncryptionSetModelToRefresh AzureDiskEncryptionSetModel, desId string) AzureDiskEncryptionSetModel {
	desName, resourceGroupName := ParseDiskEncryptionSetIdToNameAndResourceGroup(desId)
	if !strings.EqualFold(diskEncryptionSetModelToRefresh.DiskEncryptionSetName.ValueString(), desName) {
		diskEncryptionSetModelToRefresh.DiskEncryptionSetName = types.StringValue(desName)
	}
	if !strings.EqualFold(diskEncryptionSetModelToRefresh.DiskEncryptionSetResourceGroup.ValueString(), resourceGroupName) {
		diskEncryptionSetModelToRefresh.DiskEncryptionSetResourceGroup = types.StringValue(resourceGroupName)
	}
	return diskEncryptionSetModelToRefresh
}

// ensure NetworkMappingModel implements RefreshableListItemWithAttributes
var _ RefreshableListItemWithAttributes[citrixorchestration.NetworkMapResponseModel] = NetworkMappingModel{}

// NetworkMappingModel maps the nested network mapping resource schema data.
type NetworkMappingModel struct {
	NetworkDevice types.String `tfsdk:"network_device"`
	Network       types.String `tfsdk:"network"`
}

func (n NetworkMappingModel) GetKey() string {
	return n.NetworkDevice.ValueString()
}

func (NetworkMappingModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"network_device": schema.StringAttribute{
				Description: "Name or Id of the network device.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network"),
					}...),
				},
			},
			"network": schema.StringAttribute{
				Description: "The name of the virtual network that the device should be attached to. This must be a subnet within a Virtual Private Cloud item in the resource pool to which the Machine Catalog is associated." + "<br />" +
					"For AWS, please specify the network mask of the network you want to use within the VPC.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network_device"),
					}...),
				},
			},
		},
	}
}

func (NetworkMappingModel) GetDataSourceSchema() dataSourceSchema.NestedAttributeObject {
	return dataSourceSchema.NestedAttributeObject{
		Attributes: map[string]dataSourceSchema.Attribute{
			"network_device": dataSourceSchema.StringAttribute{
				Description: "Name or Id of the network device.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network"),
					}...),
				},
			},
			"network": dataSourceSchema.StringAttribute{
				Description: "The name of the virtual network that the device should be attached to. This must be a subnet within a Virtual Private Cloud item in the resource pool to which the Machine Catalog is associated." + "<br />" +
					"For AWS, please specify the network mask of the network you want to use within the VPC.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network_device"),
					}...),
				},
			},
		},
	}
}

func (NetworkMappingModel) GetAttributes() map[string]schema.Attribute {
	return NetworkMappingModel{}.GetSchema().Attributes
}

func (NetworkMappingModel) GetDataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return NetworkMappingModel{}.GetDataSourceSchema().Attributes
}

func (networkMapping NetworkMappingModel) RefreshListItem(_ context.Context, _ *diag.Diagnostics, nic citrixorchestration.NetworkMapResponseModel) ResourceModelWithAttributes {
	networkMapping.NetworkDevice = types.StringValue(nic.GetDeviceId())
	network := nic.GetNetwork()
	segments := strings.Split(network.GetXDPath(), "\\")
	lastIndex := len(segments)

	networkName := (strings.TrimSuffix(segments[lastIndex-1], ".network"))
	matchAws := regexp.MustCompile(AwsNetworkNameRegex)
	if matchAws.MatchString(networkName) {
		/* For AWS Network, the XDPath looks like:
		* XDHyp:\\HostingUnits\\{resource pool}\\{availability zone}.availabilityzone\\{network ip}`/{prefix length} (vpc-{vpc-id}).network
		* The Network property should be set to {network ip}/{prefix length}
		 */
		networkName = strings.ReplaceAll(strings.Split((networkName), " ")[0], "`/", "/")
	}
	networkMapping.Network = types.StringValue(networkName)
	return networkMapping
}

// Azure Machine Profile Model
type AzureMachineProfileModel struct {
	MachineProfileVmName              types.String `tfsdk:"machine_profile_vm_name"`
	MachineProfileTemplateSpecName    types.String `tfsdk:"machine_profile_template_spec_name"`
	MachineProfileTemplateSpecVersion types.String `tfsdk:"machine_profile_template_spec_version"`
	MachineProfileResourceGroup       types.String `tfsdk:"machine_profile_resource_group"`
}

func (AzureMachineProfileModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The name of the virtual machine or template spec that will be used to identify the default value for the tags, virtual machine size, boot diagnostics, host cache property of OS disk, accelerated networking and availability zone." + "<br />" +
			"Required when provisioning_type is set to PVSStreaming or when identity_type is set to `AzureAD`",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"machine_profile_vm_name": schema.StringAttribute{
				Description: "The name of the machine profile virtual machine.",
				Optional:    true,
			},
			"machine_profile_template_spec_name": schema.StringAttribute{
				Description: "The name of the machine profile template spec.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("machine_profile_template_spec_version"),
					}...),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRelative().AtParent().AtName("machine_profile_vm_name"),
					}...),
				},
			},
			"machine_profile_template_spec_version": schema.StringAttribute{
				Description: "The version of the machine profile template spec.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("machine_profile_template_spec_name"),
					}...),
				},
			},
			"machine_profile_resource_group": schema.StringAttribute{
				Description: "The name of the resource group where the machine profile VM or template spec is located.",
				Required:    true,
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = req.ConfigValue.IsNull() != req.StateValue.IsNull()
				},
				"Force replace when machine_profile is added or removed. Update is allowed only if previously set.",
				"Force replace when machine_profile is added or removed. Update is allowed only if previously set.",
			),
		},
	}
}

func (AzureMachineProfileModel) GetDataSourceSchema() dataSourceSchema.SingleNestedAttribute {
	return dataSourceSchema.SingleNestedAttribute{
		Description: "The name of the virtual machine or template spec that will be used to identify the default value for the tags, virtual machine size, boot diagnostics, host cache property of OS disk, accelerated networking and availability zone.",
		Computed:    true,
		Attributes: map[string]dataSourceSchema.Attribute{
			"machine_profile_vm_name": dataSourceSchema.StringAttribute{
				Description: "The name of the machine profile virtual machine.",
				Computed:    true,
			},
			"machine_profile_template_spec_name": dataSourceSchema.StringAttribute{
				Description: "The name of the machine profile template spec.",
				Computed:    true,
			},
			"machine_profile_template_spec_version": dataSourceSchema.StringAttribute{
				Description: "The version of the machine profile template spec.",
				Computed:    true,
			},
			"machine_profile_resource_group": dataSourceSchema.StringAttribute{
				Description: "The name of the resource group where the machine profile VM or template spec is located.",
				Computed:    true,
			},
		},
	}
}

func (AzureMachineProfileModel) GetAttributes() map[string]schema.Attribute {
	return AzureMachineProfileModel{}.GetSchema().Attributes
}

func (AzureMachineProfileModel) GetDataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return AzureMachineProfileModel{}.GetDataSourceSchema().Attributes
}

type AzureImageSpecsModel struct {
	// Required Attributes
	ServiceOffering types.String `tfsdk:"service_offering"`
	LicenseType     types.String `tfsdk:"license_type"`
	StorageType     types.String `tfsdk:"storage_type"`

	// Optional Attributes
	MachineProfile    types.Object `tfsdk:"machine_profile"`
	DiskEncryptionSet types.Object `tfsdk:"disk_encryption_set"`

	// Master Image Attributes
	ResourceGroup      types.String `tfsdk:"resource_group"`
	SharedSubscription types.String `tfsdk:"shared_subscription"`
	MasterImage        types.String `tfsdk:"master_image"`
	GalleryImage       types.Object `tfsdk:"gallery_image"`
}

func (AzureImageSpecsModel) GetSchema() schema.SingleNestedAttribute {
	galleryImageSchema := GalleryImageModel{}.GetSchema()
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
						WindowsClientLicenseType,
						WindowsServerLicenseType,
					),
				},
			},
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						StandardLRS,
						StandardSSDLRS,
						Premium_LRS,
					),
				},
			},
			"machine_profile":     AzureMachineProfileModel{}.GetSchema(),
			"disk_encryption_set": AzureDiskEncryptionSetModel{}.GetSchema(),
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
		Validators: []validator.Object{
			objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("vsphere_image_specs")),
		},
	}
}

func (AzureImageSpecsModel) GetAttributes() map[string]schema.Attribute {
	return AzureImageSpecsModel{}.GetSchema().Attributes
}

func HandleMachineProfileForAzureMcsPvsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, hypervisorName string, resourcePoolName string, machineProfile AzureMachineProfileModel, errorTitle string) (string, error) {
	machineProfileResourceGroup := machineProfile.MachineProfileResourceGroup.ValueString()
	machineProfileVmOrTemplateSpecVersion := machineProfile.MachineProfileVmName.ValueString()
	resourceType := VirtualMachineResourceType
	queryPath := fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
	errorMessage := fmt.Sprintf("Failed to locate machine profile vm %s on Azure", machineProfile.MachineProfileVmName.ValueString())
	isUsingTemplateSpec := false
	if machineProfile.MachineProfileVmName.IsNull() {
		isUsingTemplateSpec = true
		machineProfileVmOrTemplateSpecVersion = machineProfile.MachineProfileTemplateSpecVersion.ValueString()
		queryPath = fmt.Sprintf("%s\\%s.templatespec", queryPath, machineProfile.MachineProfileTemplateSpecName.ValueString())
		resourceType = ""
		errorMessage = fmt.Sprintf("Failed to locate machine profile template spec %s with version %s on Azure", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfile.MachineProfileTemplateSpecVersion.ValueString())
	}
	machineProfileResource, httpResp, err := GetSingleResourceFromHypervisorWithNoCacheRetry(ctx, client, diag, hypervisorName, resourcePoolName, queryPath, machineProfileVmOrTemplateSpecVersion, resourceType, "")
	if err != nil {
		diag.AddError(
			errorTitle,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\n%s, error: %s", errorMessage, err.Error()),
		)
		return "", err
	}
	if isUsingTemplateSpec {
		// validate the template spec
		isValid, errorMsg := ValidateHypervisorResource(ctx, client, hypervisorName, resourcePoolName, machineProfileResource.GetRelativePath())
		if !isValid {
			diag.AddError(
				errorTitle,
				fmt.Sprintf("Failed to validate template spec %s with version %s, %s", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfileVmOrTemplateSpecVersion, errorMsg),
			)
			return "", fmt.Errorf("failed to validate template spec %s with version %s, %s", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfileVmOrTemplateSpecVersion, errorMsg)
		}
	}

	return machineProfileResource.GetXDPath(), nil
}

func ParseAzureMachineProfileResponseToModel(machineProfileResponse citrixorchestration.HypervisorResourceRefResponseModel) *AzureMachineProfileModel {
	machineProfileModel := AzureMachineProfileModel{}
	if machineProfileName := machineProfileResponse.GetName(); machineProfileName != "" {
		machineProfileSegments := strings.Split(machineProfileResponse.GetXDPath(), "\\")
		lastIndex := len(machineProfileSegments) - 1
		if strings.HasSuffix(machineProfileSegments[lastIndex], "templatespecversion") {
			machineProfileModel.MachineProfileTemplateSpecVersion = types.StringValue(machineProfileName)

			templateSpecIndex := slices.IndexFunc(machineProfileSegments, func(machineProfileSegment string) bool {
				return strings.Contains(machineProfileSegment, ".templatespec")
			})

			if templateSpecIndex != -1 {
				templateSpec := strings.TrimSuffix(machineProfileSegments[templateSpecIndex], ".templatespec")
				machineProfileModel.MachineProfileTemplateSpecName = types.StringValue(templateSpec)
			}
		} else {
			machineProfileModel.MachineProfileVmName = types.StringValue(machineProfileName)
		}

		resourceGroupIndex := slices.IndexFunc(machineProfileSegments, func(machineProfileSegment string) bool {
			return strings.Contains(machineProfileSegment, ".resourcegroup")
		})

		if resourceGroupIndex != -1 {
			resourceGroup := strings.TrimSuffix(machineProfileSegments[resourceGroupIndex], ".resourcegroup")
			machineProfileModel.MachineProfileResourceGroup = types.StringValue(resourceGroup)
		}
	} else {
		machineProfileModel.MachineProfileVmName = types.StringNull()
		machineProfileModel.MachineProfileResourceGroup = types.StringNull()
	}
	return &machineProfileModel
}

func ParseNetworkMappingToClientModel(networkMappings []NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel, hypervisorPluginId string) ([]citrixorchestration.NetworkMapRequestModel, error) {
	var networks []citrixorchestration.HypervisorResourceRefResponseModel
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		networks = resourcePool.Subnets
	} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == NUTANIX_PLUGIN_ID {
		networks = resourcePool.Networks
	}

	var res = []citrixorchestration.NetworkMapRequestModel{}
	for _, networkMapping := range networkMappings {
		var networkName string
		if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == NUTANIX_PLUGIN_ID {
			networkName = networkMapping.Network.ValueString()
		} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
			networkName = fmt.Sprintf("%s (%s)", networkMapping.Network.ValueString(), resourcePool.GetResourcePoolRootId())
		}
		network := slices.IndexFunc(networks, func(c citrixorchestration.HypervisorResourceRefResponseModel) bool {
			return strings.EqualFold(c.GetName(), networkName)
		})
		if network == -1 {
			return res, fmt.Errorf("network %s not found", networkName)
		}

		networkMapRequestModel := citrixorchestration.NetworkMapRequestModel{
			NetworkDeviceNameOrId: *citrixorchestration.NewNullableString(networkMapping.NetworkDevice.ValueStringPointer()),
			NetworkPath:           networks[network].GetXDPath(),
		}

		if resourcePool.GetConnectionType() == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER || (resourcePool.GetConnectionType() == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == NUTANIX_PLUGIN_ID) {
			networkMapRequestModel.SetDeviceNameOrId(networks[network].GetId())
		}

		res = append(res, networkMapRequestModel)
	}

	return res, nil
}

func BuildAzureMasterImagePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, galleryImage types.Object, sharedSubscription string, resourceGroup string, storageAccount string, storageContainer string, masterImage string, hypervisor string, hypervisorResourcePool string, errorTitle string) (string, error) {
	imagePath := ""
	imageBasePath := "image.folder"
	queryPath := ""
	err := error(nil)
	var httpResp *http.Response
	if sharedSubscription != "" {
		imageBasePath = fmt.Sprintf("image.folder\\%s.sharedsubscription", sharedSubscription)
	}
	if masterImage != "" {
		if storageAccount != "" && storageContainer != "" {
			queryPath = fmt.Sprintf(
				"%s\\%s.resourcegroup\\%s.storageaccount\\%s.container",
				imageBasePath,
				resourceGroup,
				storageAccount,
				storageContainer)
			imagePath, httpResp, err = GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor, hypervisorResourcePool, queryPath, masterImage, "", "")
			if err != nil {
				diagnostics.AddError(
					errorTitle,
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to resolve master image VHD %s in container %s of storage account %s, error: %s", masterImage, storageContainer, storageAccount, err.Error()),
				)
				return imagePath, err
			}
		} else {
			queryPath = fmt.Sprintf(
				"%s\\%s.resourcegroup",
				imageBasePath,
				resourceGroup)
			imagePath, httpResp, err = GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor, hypervisorResourcePool, queryPath, masterImage, "", "")
			if err != nil {
				diagnostics.AddError(
					errorTitle,
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to resolve master image Managed Disk or Snapshot %s, error: %s", masterImage, err.Error()),
				)
				return imagePath, err
			}
		}
	} else if !galleryImage.IsNull() {
		azureGalleryImage := ObjectValueToTypedObject[GalleryImageModel](ctx, diagnostics, galleryImage)
		gallery := azureGalleryImage.Gallery.ValueString()
		definition := azureGalleryImage.Definition.ValueString()
		version := azureGalleryImage.Version.ValueString()
		if gallery != "" && definition != "" {
			queryPath = fmt.Sprintf(
				"%s\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
				imageBasePath,
				resourceGroup,
				gallery,
				definition)
			imagePath, httpResp, err = GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor, hypervisorResourcePool, queryPath, version, "", "")
			if err != nil {
				diagnostics.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", definition, version, gallery, err.Error()),
				)
				return imagePath, err
			}
		}
	}
	return imagePath, nil
}

func ParseMasterImageToUpdateGalleryImageModel(ctx context.Context, diagnostics *diag.Diagnostics, galleryImage types.Object, masterImage citrixorchestration.HypervisorResourceRefResponseModel, masterImageSegments []string, masterImageLastIndex int) (types.Object, types.String, types.String) {
	/* For Azure Image Gallery image, the XDPath looks like:
	* XDHyp:\\HostingUnits\\{resource pool}\\image.folder\\{resource group}.resourcegroup\\{gallery name}.gallery\\{image name}.imagedefinition\\{image version}.imageversion
	* The Name property in MasterImage will be image version instead of image definition (name of the image)
	 */
	azureGalleryImageModel := ObjectValueToTypedObject[GalleryImageModel](ctx, diagnostics, galleryImage)
	azureGalleryImageModel.Version = types.StringValue(masterImage.GetName())
	// Extract {image name} from {image name}.imagedefinition
	azureGalleryImageModel.Definition = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-2], ".imagedefinition"))
	// Extract {gallery name} from {gallery name}.gallery
	azureGalleryImageModel.Gallery = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-3], ".gallery"))

	galleryImageToReturn := TypedObjectToObjectValue(ctx, diagnostics, azureGalleryImageModel)
	resourceGroupToReturn := types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-4], ".resourcegroup"))
	segment := strings.Split(masterImageSegments[masterImageLastIndex-5], ".")
	resourceType := segment[len(segment)-1]
	var sharedSubscriptionToReturn types.String
	if strings.EqualFold(resourceType, SharedSubscriptionResourceType) {
		sharedSubscriptionToReturn = types.StringValue(segment[0])
	} else {
		sharedSubscriptionToReturn = types.StringNull()
	}

	return galleryImageToReturn, resourceGroupToReturn, sharedSubscriptionToReturn
}

func ParseMasterImageToUpdateAzureImageSpecs(ctx context.Context, diagnostics *diag.Diagnostics, resourceType string, masterImage citrixorchestration.HypervisorResourceRefResponseModel, masterImageSegments []string, masterImageLastIndex int) (types.String, types.String, types.String, types.Object, types.String, types.String) {
	var masterImageToReturn types.String
	var storageAccountToReturn types.String
	var containerToReturn types.String
	var resourceGroupToReturn types.String
	var galleryImageToReturn types.Object
	var sharedSubscriptionToReturn types.String
	if strings.EqualFold(resourceType, VhdResourceType) {
		// VHD image
		masterImageToReturn = types.StringValue(masterImage.GetName())
		containerToReturn = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-2], ".container"))
		storageAccountToReturn = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-3], ".storageaccount"))
		resourceGroupToReturn = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-4], ".resourcegroup"))

		segment := strings.Split(masterImageSegments[masterImageLastIndex-5], ".")
		resourceType := segment[len(segment)-1]
		if strings.EqualFold(resourceType, SharedSubscriptionResourceType) {
			sharedSubscriptionToReturn = types.StringValue(segment[0])
		} else {
			sharedSubscriptionToReturn = types.StringNull()
		}
	} else {
		// Snapshot or Managed Disk
		masterImageToReturn = types.StringValue(masterImage.GetName())
		resourceGroupToReturn = types.StringValue(strings.TrimSuffix(masterImageSegments[masterImageLastIndex-2], ".resourcegroup"))
		segment := strings.Split(masterImageSegments[masterImageLastIndex-3], ".")
		resourceType := segment[len(segment)-1]
		if strings.EqualFold(resourceType, SharedSubscriptionResourceType) {
			sharedSubscriptionToReturn = types.StringValue(segment[0])
		} else {
			sharedSubscriptionToReturn = types.StringNull()
		}

		// Clear VHD image details
		storageAccountToReturn = types.StringNull()
		containerToReturn = types.StringNull()
	}

	// Clear gallery image details
	attributeMap, err := ResourceAttributeMapFromObject(GalleryImageModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
	} else {
		galleryImageToReturn = types.ObjectNull(attributeMap)
	}

	return masterImageToReturn, resourceGroupToReturn, sharedSubscriptionToReturn, galleryImageToReturn, storageAccountToReturn, containerToReturn
}

func ParseMasterImageToAzureImageModel(ctx context.Context, diagnostics *diag.Diagnostics, azureImageSpecs AzureImageSpecsModel, masterImage citrixorchestration.HypervisorResourceRefResponseModel) AzureImageSpecsModel {
	masterImageXdPath := masterImage.GetXDPath()
	masterImageSegments := strings.Split(masterImageXdPath, "\\")
	masterImageLastIndex := len(masterImageSegments)
	masterImageResourceTag := strings.Split(masterImageSegments[masterImageLastIndex-1], ".")
	masterImageResourceType := masterImageResourceTag[len(masterImageResourceTag)-1]
	if strings.EqualFold(masterImageResourceType, ImageVersionResourceType) {
		azureImageSpecs.GalleryImage,
			azureImageSpecs.ResourceGroup,
			azureImageSpecs.SharedSubscription =
			ParseMasterImageToUpdateGalleryImageModel(ctx, diagnostics, azureImageSpecs.GalleryImage, masterImage, masterImageSegments, masterImageLastIndex)

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
			ParseMasterImageToUpdateAzureImageSpecs(ctx, diagnostics, masterImageResourceType, masterImage, masterImageSegments, masterImageLastIndex)
	}
	return azureImageSpecs
}
