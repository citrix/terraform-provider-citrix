// Copyright © 2023. Citrix Systems, Inc.

package daas

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &machineCatalogResource{}
	_ resource.ResourceWithConfigure   = &machineCatalogResource{}
	_ resource.ResourceWithImportState = &machineCatalogResource{}
)

// NewMachineCatalogResource is a helper function to simplify the provider implementation.
func NewMachineCatalogResource() resource.Resource {
	return &machineCatalogResource{}
}

// machineCatalogResource is the resource implementation.
type machineCatalogResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *machineCatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_machine_catalog"
}

// Schema defines the schema for the data source.
func (r *machineCatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a machine catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the machine catalog.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine catalog.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the machine catalog.",
				Optional:    true,
			},
			"service_account": schema.StringAttribute{
				Description: "Service account for the domain.",
				Required:    true,
			},
			"service_account_password": schema.StringAttribute{
				Description: "Service account password for the domain.",
				Required:    true,
				Sensitive:   true,
			},
			"allocation_type": schema.StringAttribute{
				Description: "Denotes how the machines in the catalog are allocated to a user. Choose between `Static` and `Random`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedAllocationTypeEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"session_support": schema.StringAttribute{
				Description: "Session support type. Choose between `SingleSession` and `MultiSession`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionSupportEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the machine catalog is associated with.",
				Required:    true,
			},
			"vda_upgrade_type": schema.StringAttribute{
				Description: "Type of Vda Upgrade. Choose between LTSR and CR",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"LTSR",
						"CR",
					),
				},
			},
			"provisioning_scheme": schema.SingleNestedAttribute{
				Description: "Machine catalog provisioning scheme.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"machine_config": schema.SingleNestedAttribute{
						Description: "Machine Configuration",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"hypervisor": schema.StringAttribute{
								Description: "Id of the hypervisor for creating the machines.",
								Required:    true,
							},
							"hypervisor_resource_pool": schema.StringAttribute{
								Description: "Id of the hypervisor resource pool that will be used for provisioning operations.",
								Required:    true,
							},
							"service_offering": schema.StringAttribute{
								Description: "**[Azure, AWS: Required]** The VM Sku of a Cloud service offering to use when creating machines.",
								Optional:    true,
							},
							"master_image": schema.StringAttribute{
								Description: "The name of the virtual machine snapshot or VM template that will be used. This identifies the hard disk to be used and the default values for the memory and processors.",
								Required:    true,
							},
							"resource_group": schema.StringAttribute{
								Description: "**[Azure: Required]** The Azure Resource Group where the image VHD for creating machines is located.",
								Optional:    true,
							},
							"storage_account": schema.StringAttribute{
								Description: "**[Azure: Optional]** The Azure Storage Account where the image VHD for creating machines is located. Only applicable to Azure VHD image blob.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("container"),
									}...),
									stringvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("resource_group"),
									}...),
								},
							},
							"container": schema.StringAttribute{
								Description: "**[Azure: Optional]** The Azure Storage Account Container where the image VHD for creating machines is located. Only applicable to Azure VHD image blob.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("storage_account"),
									}...),
									stringvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("resource_group"),
									}...),
								},
							},
							"gallery_image": schema.SingleNestedAttribute{
								Description: "**[Azure: Optional]** Details of the Azure Image Gallery image to use for creating machines. Only Applicable to Azure Image Gallery image.",
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
							},
							"image_ami": schema.StringAttribute{
								Description: "**[AWS: Required]** AMI of the AWS image to be used as the template image for the machine catalog.",
								Optional:    true,
							},
							"machine_profile": schema.StringAttribute{
								Description: "**[GCP: Optional]** The name of the virtual machine template that will be used to identify the default value for the tags, virtual machine size, boot diagnostics, host cache property of OS disk, accelerated networking and availability zone. If not specified, the VM specified in master_image will be used as template.",
								Optional:    true,
							},
							"machine_snapshot": schema.StringAttribute{
								Description: "**[GCP: Optional]** The name of the virtual machine snapshot of a GCP VM that will be used as master image.",
								Optional:    true,
							},
						},
					},
					"number_of_total_machines": schema.Int64Attribute{
						Description: "Number of VDA machines allocated in the catalog.",
						Required:    true,
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"network_mapping": schema.SingleNestedAttribute{
						Description: "Specifies how the attached NICs are mapped to networks.  If this parameter is omitted, provisioned VMs are created with a single NIC, which is mapped to the default network in the hypervisor resource pool.  If this parameter is supplied, machines are created with the number of NICs specified in the map, and each NIC is attached to the specified network.",
						Optional:    true,
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
					},
					"machine_account_creation_rules": schema.SingleNestedAttribute{
						Description: "Rules specifying how Active Directory machine accounts should be created when machines are provisioned.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"naming_scheme": schema.StringAttribute{
								Description: "Defines the template name for AD accounts created in the identity pool.",
								Required:    true,
							},
							"naming_scheme_type": schema.StringAttribute{
								Description: "Type of naming scheme. This defines the format of the variable part of the AD account names that will be created. Choose between `Numeric`, `Alphabetic` and `Unicode`.",
								Required:    true,
								Validators: []validator.String{
									util.GetValidatorFromEnum(citrixorchestration.AllowedAccountNamingSchemeTypeEnumValues),
								},
							},
							"domain": schema.StringAttribute{
								Description: "The AD domain name for the pool. Specify this in FQDN format; for example, MyDomain.com.",
								Required:    true,
							},
							"domain_ou": schema.StringAttribute{
								Description: "The organization unit that computer accounts will be created into.",
								Optional:    true,
							},
						},
					},
					"availability_zones": schema.StringAttribute{
						Description: "The Azure Availability Zones containing provisioned virtual machines. Use a comma as a delimiter for multiple availability_zones.",
						Optional:    true,
					},
					"storage_type": schema.StringAttribute{
						Description: "**[Azure, GCP: Required]** Storage account type used for provisioned virtual machine disks on Azure / GCP." + "<br />" +
							"Azure storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`." + "<br />" +
							"GCP storage types include: `pd-standar`, `pd-balanced`, `pd-ssd` and `pd-extreme`.",
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"Standard_LRS",
								"StandardSSD_LRS",
								"Premium_LRS",
								"pd-standard",
								"pd-balanced",
								"pd-ssd",
								"pd-extreme",
							),
						},
					},
					"vda_resource_group": schema.StringAttribute{
						Description: "**[Azure: Optional]** Designated resource group where the VDA VMs will be located on Azure.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"use_managed_disks": schema.BoolAttribute{
						Description: "**[Azure: Optional]** Indicate whether to use Azure managed disks for the provisioned virtual machine.",
						Optional:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"writeback_cache": schema.SingleNestedAttribute{
						Description: "Write-back Cache config. Leave this empty to disable Write-back Cache.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"persist_wbc": schema.BoolAttribute{
								Description: "Persist Write-back Cache",
								Required:    true,
							},
							"wbc_disk_storage_type": schema.StringAttribute{
								Description: "Type of naming scheme. Choose between Numeric and Alphabetic.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"StandardSSD_LRS",
										"Standard_LRS",
										"Premium_LRS",
									),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"persist_os_disk": schema.BoolAttribute{
								Description: "Persist the OS disk when power cycling the non-persistent provisioned virtual machine.",
								Required:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"persist_vm": schema.BoolAttribute{
								Description: "Persist the non-persistent provisioned virtual machine in Azure environments when power cycling. This property only applies when the PersistOsDisk property is set to True.",
								Required:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"storage_cost_saving": schema.BoolAttribute{
								Description: "Save storage cost by downgrading the storage type of the disk to Standard HDD when VM shut down.",
								Required:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"writeback_cache_disk_size_gb": schema.Int64Attribute{
								Description: "The size in GB of any temporary storage disk used by the write back cache.",
								Required:    true,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
							"writeback_cache_memory_size_mb": schema.Int64Attribute{
								Description: "The size of the in-memory write back cache in MB.",
								Optional:    true,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
								PlanModifiers: []planmodifier.Int64{ // TO DO - Allow updating master image
									int64planmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *machineCatalogResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *machineCatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan models.MachineCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var machineAccountCreationRules citrixorchestration.MachineAccountCreationRulesRequestModel
	machineAccountCreationRules.SetNamingScheme(plan.ProvisioningScheme.MachineAccountCreationRules.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(plan.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating machine catalog",
			"Unsupported machine account naming scheme type.",
		)
		return
	}
	machineAccountCreationRules.SetNamingSchemeType(*namingScheme)
	machineAccountCreationRules.SetDomain(plan.ProvisioningScheme.MachineAccountCreationRules.Domain.ValueString())
	machineAccountCreationRules.SetOU(plan.ProvisioningScheme.MachineAccountCreationRules.Ou.ValueString())

	var provisioningScheme citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel
	provisioningScheme.SetNumTotalMachines(int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()))
	provisioningScheme.SetIdentityType(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY) // Non-Managed setup does not support non-domain joined
	provisioningScheme.SetWorkGroupMachines(false)                                        // Non-Managed setup does not support non-domain joined
	provisioningScheme.SetMachineAccountCreationRules(machineAccountCreationRules)
	provisioningScheme.SetResourcePool(plan.ProvisioningScheme.MachineConfig.HypervisorResourcePool.ValueString())

	// Resolve resource path for service offering and master image
	hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, plan.ProvisioningScheme.MachineConfig.Hypervisor.ValueString())
	if err != nil {
		return
	}

	hypervisorResourcePool, err := GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.ProvisioningScheme.MachineConfig.Hypervisor.ValueString(), plan.ProvisioningScheme.MachineConfig.HypervisorResourcePool.ValueString())
	if err != nil {
		return
	}

	serviceOffering := plan.ProvisioningScheme.MachineConfig.ServiceOffering.ValueString()
	masterImage := plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString()

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		queryPath := "serviceoffering.folder"
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, "serviceoffering", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return
		}
		provisioningScheme.SetServiceOfferingPath(serviceOfferingPath)

		resourceGroup := plan.ProvisioningScheme.MachineConfig.ResourceGroup.ValueString()
		imagePath := ""
		if masterImage != "" {
			storageAccount := plan.ProvisioningScheme.MachineConfig.StorageAccount.ValueString()
			container := plan.ProvisioningScheme.MachineConfig.Container.ValueString()
			if storageAccount != "" && container != "" {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.storageaccount\\%s.container",
					resourceGroup,
					storageAccount,
					container)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, masterImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image VHD %s in container %s of storage account %s, error: %s", masterImage, container, storageAccount, err.Error()),
					)
					return
				}
			} else {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup",
					resourceGroup)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, masterImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image Managed Disk or Snapshot %s, error: %s", masterImage, err.Error()),
					)
					return
				}
			}
		} else if plan.ProvisioningScheme.MachineConfig.GalleryImage != nil {
			gallery := plan.ProvisioningScheme.MachineConfig.GalleryImage.Gallery.ValueString()
			definition := plan.ProvisioningScheme.MachineConfig.GalleryImage.Definition.ValueString()
			version := plan.ProvisioningScheme.MachineConfig.GalleryImage.Version.ValueString()
			if gallery != "" && definition != "" {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
					resourceGroup,
					gallery,
					definition)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", masterImage, version, gallery, err.Error()),
					)
					return
				}
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", serviceOffering, "serviceoffering", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return
		}
		provisioningScheme.SetServiceOfferingPath(serviceOfferingPath)

		imageId := fmt.Sprintf("%s (%s)", masterImage, plan.ProvisioningScheme.MachineConfig.ImageAmi.ValueString())
		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, "template", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", masterImage, plan.ProvisioningScheme.MachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return
		}
		provisioningScheme.SetMasterImagePath(imagePath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if serviceOffering != "" {
			resp.Diagnostics.AddError(
				"Error creating Machine Catalog",
				"GCP machine catalog does not support service_offering. Please use master_image (and optionally with machine_snapshot) or machine_profile to specify the GCP VM you want to use as a template for the VM SKU.",
			)
			return
		}
		imagePath := ""
		snapshot := plan.ProvisioningScheme.MachineConfig.MachineSnapshot.ValueString()
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString())
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, plan.ProvisioningScheme.MachineConfig.MachineSnapshot.ValueString(), "snapshot", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate master image snapshot %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString(), "vm", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		machineProfile := plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfile, "vm", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping := models.ParseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool)
		provisioningScheme.SetNetworkMapping(networkMapping)
	}

	if plan.ProvisioningScheme.WritebackCache != nil {
		provisioningScheme.SetUseWriteBackCache(true)
		provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
		if !plan.ProvisioningScheme.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
			provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
		}
		if plan.ProvisioningScheme.WritebackCache.PersistVm.ValueBool() && !plan.ProvisioningScheme.WritebackCache.PersistOsDisk.ValueBool() {
			resp.Diagnostics.AddError(
				"Error creating machine catalog",
				"Could not set persist_vm attribute, which can only be set when persist_os_disk = true",
			)
			return
		}

	}

	customProperties := models.ParseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType)

	provisioningScheme.SetCustomProperties(*customProperties)

	var body citrixorchestration.CreateMachineCatalogRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetProvisioningType(citrixorchestration.PROVISIONINGTYPE_MCS)        // Block PVS and manual. Only support MCS
	body.SetMinimumFunctionalLevel(citrixorchestration.FUNCTIONALLEVEL_L7_20) // Hard-coding VDA feature level to be same as QCS
	body.SetIsRemotePC(false)                                                 // Block RemotePC
	allocationType, err := citrixorchestration.NewAllocationTypeFromValue(plan.AllocationType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported allocation type.",
		)
		return
	}
	body.SetAllocationType(*allocationType)
	sessionSupport, err := citrixorchestration.NewSessionSupportFromValue(plan.SessionSupport.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported allocation type.",
		)
		return
	}
	body.SetSessionSupport(*sessionSupport)
	persistChanges := citrixorchestration.PERSISTCHANGES_DISCARD
	if *sessionSupport == citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION && *allocationType == citrixorchestration.ALLOCATIONTYPE_STATIC {
		persistChanges = citrixorchestration.PERSISTCHANGES_ON_LOCAL
	}
	body.SetPersistUserChanges(persistChanges)
	body.SetZone(plan.Zone.ValueString())
	body.SetProvisioningScheme(provisioningScheme)
	if !plan.VdaUpgradeType.IsNull() {
		body.SetVdaUpgradeType(citrixorchestration.VdaUpgradeType(plan.VdaUpgradeType.ValueString()))
	} else {
		body.SetVdaUpgradeType(citrixorchestration.VDAUPGRADETYPE_NOT_SET)
	}

	createMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsCreateMachineCatalog(ctx)

	// Add domain credential header
	header := generateAdminCredentialHeader(plan)
	createMachineCatalogRequest = createMachineCatalogRequest.XAdminCredential(header)

	// Add request body
	createMachineCatalogRequest = createMachineCatalogRequest.CreateMachineCatalogRequestModel(body)

	// Make request async
	createMachineCatalogRequest = createMachineCatalogRequest.Async(true)

	// Create new machine catalog
	_, httpResp, err := citrixdaasclient.AddRequestData(createMachineCatalogRequest, r.client).Execute()
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	jobId := citrixdaasclient.GetJobIdFromHttpResponse(*httpResp)
	jobResponseModel, err := r.client.WaitForJob(ctx, jobId, 120)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"TransactionId: "+txId+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorDetail := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			errorDetail,
		)
	}

	// Get the new catalog
	catalog, err := GetMachineCatalog(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString(), true)

	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, catalog, hypervisor.GetConnectionType())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *machineCatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state models.MachineCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalog state from Orchestration
	catalogId := state.Id.ValueString()

	catalog, _, err := readMachineCatalog(ctx, r.client, resp, catalogId)
	if err != nil {
		return
	}

	// Resolve resource path for service offering and master image
	provScheme := catalog.GetProvisioningScheme()
	resourcePool := provScheme.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	hypervisorName := hypervisor.GetName()
	if hypervisorName != "" {
		hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorName)
		if err != nil {
			return
		}

		// Overwrite items with refreshed state
		state = state.RefreshPropertyValues(ctx, catalog, hypervisor.GetConnectionType())
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *machineCatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan models.MachineCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalogs from Orchestration
	catalogId := plan.Id.ValueString()
	catalogName := plan.Name.ValueString()
	_, err := GetMachineCatalog(ctx, r.client, &resp.Diagnostics, catalogId, true)

	if err != nil {
		return
	}

	// Generate API request body from plan
	var body citrixorchestration.UpdateMachineCatalogRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetZone(plan.Zone.ValueString())
	if !plan.VdaUpgradeType.IsNull() {
		body.SetVdaUpgradeType(citrixorchestration.VdaUpgradeType(plan.VdaUpgradeType.ValueString()))
	} else {
		body.SetVdaUpgradeType(citrixorchestration.VDAUPGRADETYPE_NOT_SET)
	}

	// Resolve resource path for service offering and master image
	hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, plan.ProvisioningScheme.MachineConfig.Hypervisor.ValueString())
	if err != nil {
		return
	}

	hypervisorResourcePool, err := GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.ProvisioningScheme.MachineConfig.Hypervisor.ValueString(), plan.ProvisioningScheme.MachineConfig.HypervisorResourcePool.ValueString())
	if err != nil {
		return
	}

	serviceOffering := plan.ProvisioningScheme.MachineConfig.ServiceOffering.ValueString()

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		queryPath := "serviceoffering.folder"
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, "serviceoffering", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", serviceOffering, "serviceoffering", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if serviceOffering != "" {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				"GCP machine catalog does not support service_offering. Please use master_image (and optionally with machine_snapshot) or machine_profile to specify the GCP VM you want to use as a template for the VM SKU.",
			)
			return
		}
		machineProfile := plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, r.client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), "vm", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return
			}
			body.SetMachineProfilePath(machineProfilePath)
		}
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping := models.ParseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool)
		body.SetNetworkMapping(networkMapping)
	} else {
		var state models.MachineCatalogResourceModel
		req.State.Get(ctx, &state)
		if state.ProvisioningScheme.NetworkMapping != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog "+catalogName,
				"Machine catalog created with explicit Network Mapping in Provisioning Scheme must be updated with explicit Network Mapping",
			)
			return
		}
	}

	customProperties := models.ParseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType)
	body.SetCustomProperties(*customProperties)

	updateMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalog(ctx, catalogId)
	updateMachineCatalogRequest = updateMachineCatalogRequest.UpdateMachineCatalogRequestModel(body)
	catalog, httpResp, err := citrixdaasclient.AddRequestData(updateMachineCatalogRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = updateCatalogImage(ctx, r.client, resp, catalog, hypervisor, hypervisorResourcePool, plan)

	if err != nil {
		return
	}

	if catalog.GetTotalCount() > int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) {
		// delete machines from machine catalog
		err = deleteMachinesFromMachineCatalog(ctx, r.client, resp, catalog, plan)
		if err != nil {
			return
		}
	}

	if catalog.GetTotalCount() < int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) {
		// add machines to machine catalog
		err = addMachinesToMachineCatalog(ctx, r.client, resp, catalog, plan)
		if err != nil {
			return
		}
	}

	// Fetch updated machine catalog from GetMachineCatalog.
	catalog, err = GetMachineCatalog(ctx, r.client, &resp.Diagnostics, catalogId, true)
	if err != nil {
		return
	}

	// Update resource state with updated items and timestamp
	plan = plan.RefreshPropertyValues(ctx, catalog, hypervisor.GetConnectionType())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *machineCatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state models.MachineCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := state.Id.ValueString()

	_, httpResp, err := readMachineCatalog(ctx, r.client, nil, catalogId)

	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Machine Catalog "+catalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)

		return
	}

	// Delete existing order
	catalogName := state.Name.ValueString()
	deleteMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsDeleteMachineCatalog(ctx, catalogId)

	// Add domain credential header
	header := generateAdminCredentialHeader(state)
	deleteMachineCatalogRequest = deleteMachineCatalogRequest.XAdminCredential(header).DeleteVm(true).DeleteAccount(citrixorchestration.MACHINEACCOUNTDELETEOPTION_DELETE).Async(true)
	httpResp, err = citrixdaasclient.AddRequestData(deleteMachineCatalogRequest, r.client).Execute()
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	jobId := citrixdaasclient.GetJobIdFromHttpResponse(*httpResp)
	jobResponseModel, err := r.client.WaitForJob(ctx, jobId, 60)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorDetail := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		resp.Diagnostics.AddError(
			"Error deleting Machine Catalog "+catalogName,
			errorDetail,
		)
	}
}

func (r *machineCatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func generateAdminCredentialHeader(plan models.MachineCatalogResourceModel) string {
	credential := fmt.Sprintf("%s\\%s:%s", plan.ProvisioningScheme.MachineAccountCreationRules.Domain.ValueString(), plan.ServiceAccount.ValueString(), plan.ServiceAccountPassword.ValueString())
	encodedData := base64.StdEncoding.EncodeToString([]byte(credential))
	header := fmt.Sprintf("Basic %s", encodedData)

	return header
}

func generateBatchApiHeaders(client *citrixdaasclient.CitrixDaasClient, plan models.MachineCatalogResourceModel, generateCredentialHeader bool) ([]citrixorchestration.NameValueStringPairModel, *http.Response, error) {
	headers := []citrixorchestration.NameValueStringPairModel{}

	cwsAuthToken, httpResp, err := client.SignIn()
	var token string
	if err != nil {
		return headers, httpResp, err
	}

	if cwsAuthToken != "" {
		token = strings.Split(cwsAuthToken, "=")[1]
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("Authorization")
		header.SetValue("Bearer " + token)
		headers = append(headers, header)
	}

	if generateCredentialHeader {
		adminCredentialHeader := generateAdminCredentialHeader(plan)
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("X-AdminCredential")
		header.SetValue(adminCredentialHeader)
		headers = append(headers, header)
	}

	return headers, httpResp, err
}

func GetMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogId string, addErrorToDiagnostics bool) (*citrixorchestration.MachineCatalogDetailResponseModel, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogId)
	catalog, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, client)
	if err != nil && addErrorToDiagnostics {
		diagnostics.AddError(
			"Error reading Machine Catalog "+machineCatalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return catalog, err
}

func readMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, machineCatalogId string) (*citrixorchestration.MachineCatalogDetailResponseModel, *http.Response, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogId)
	catalog, httpResp, err := util.ReadResource[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, ctx, client, resp, "Machine Catalog", machineCatalogId)
	return catalog, httpResp, err
}

func updateCatalogImage(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel, plan models.MachineCatalogResourceModel) error {

	catalogName := catalog.GetName()
	catalogId := catalog.GetId()

	provScheme := catalog.GetProvisioningScheme()
	masterImage := provScheme.GetMasterImage()

	// Check if XDPath has changed for the image
	imagePath := ""
	var err error
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		newImage := plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString()
		resourceGroup := plan.ProvisioningScheme.MachineConfig.ResourceGroup.ValueString()
		if newImage != "" {
			storageAccount := plan.ProvisioningScheme.MachineConfig.StorageAccount.ValueString()
			container := plan.ProvisioningScheme.MachineConfig.Container.ValueString()
			if storageAccount != "" && container != "" {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.storageaccount\\%s.container",
					resourceGroup,
					storageAccount,
					container)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image VHD %s in container %s of storage account %s, error: %s", newImage, container, storageAccount, err.Error()),
					)
					return err
				}
			} else {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup",
					resourceGroup)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image Managed Disk or Snapshot %s, error: %s", newImage, err.Error()),
					)
					return err
				}
			}
		} else if plan.ProvisioningScheme.MachineConfig.GalleryImage != nil {
			gallery := plan.ProvisioningScheme.MachineConfig.GalleryImage.Gallery.ValueString()
			definition := plan.ProvisioningScheme.MachineConfig.GalleryImage.Definition.ValueString()
			version := plan.ProvisioningScheme.MachineConfig.GalleryImage.Version.ValueString()
			if gallery != "" && definition != "" {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
					resourceGroup,
					gallery,
					definition)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", newImage, version, gallery, err.Error()),
					)
					return err
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		imageId := fmt.Sprintf("%s (%s)", plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString(), plan.ProvisioningScheme.MachineConfig.ImageAmi.ValueString())
		imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, "template", "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString(), plan.ProvisioningScheme.MachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		newImage := plan.ProvisioningScheme.MachineConfig.MasterImage.ValueString()
		snapshot := plan.ProvisioningScheme.MachineConfig.MachineSnapshot.ValueString()
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", newImage)
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, plan.ProvisioningScheme.MachineConfig.MachineSnapshot.ValueString(), "snapshot", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image snapshot %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", newImage, "vm", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", plan.ProvisioningScheme.MachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		}
	}

	if masterImage.GetXDPath() == imagePath {
		return nil
	}

	// Update Master Image for Machine Catalog
	var updateProvisioningSchemeModel citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel
	var rebootOption citrixorchestration.RebootMachinesRequestModel

	// Update the image immediately
	rebootOption.SetRebootDuration(60)
	rebootOption.SetWarningDuration(15)
	rebootOption.SetWarningMessage("Warning: An important update is about to be installed. To ensure that no loss of data occurs, save any outstanding work and close all applications.")
	updateProvisioningSchemeModel.SetRebootOptions(rebootOption)
	updateProvisioningSchemeModel.SetMasterImagePath(imagePath)
	updateProvisioningSchemeModel.SetStoreOldImage(true)
	updateProvisioningSchemeModel.SetMinimumFunctionalLevel("L7_20")
	updateMasterImageRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalogProvisioningScheme(ctx, catalogId)
	updateMasterImageRequest = updateMasterImageRequest.UpdateMachineCatalogProvisioningSchemeRequestModel(updateProvisioningSchemeModel)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateMasterImageRequest, client).Async(true).Execute()
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Image for Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	jobId := citrixdaasclient.GetJobIdFromHttpResponse(*httpResp)
	jobResponseModel, err := client.WaitForJob(ctx, jobId, 60)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Image for Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return err
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorDetail := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		resp.Diagnostics.AddError(
			"Error updating image for Machine Catalog "+catalogName,
			errorDetail,
		)
		return err
	}

	return nil
}

func deleteMachinesFromMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan models.MachineCatalogResourceModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	if catalog.GetAllocationType() != citrixorchestration.ALLOCATIONTYPE_RANDOM {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"Deleting machine(s) is supported for machine catalogs with Random allocation type only.",
		)
		return fmt.Errorf("deleting machine(s) is supported for machine catalogs with Random allocation type only")
	}

	getMachinesRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalogMachines(ctx, catalogId)
	getMachinesResponse, httpResp, err := citrixdaasclient.AddRequestData(getMachinesRequest, client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting machine(s) from Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not retrieve machines for machine catalog",
		)
		return err
	}

	machineDeleteRequestCount := int(catalog.GetTotalCount()) - int(plan.ProvisioningScheme.NumTotalMachines.ValueInt64())
	machinesToDelete := []citrixorchestration.MachineResponseModel{}

	for _, machine := range getMachinesResponse.GetItems() {
		if !machine.GetDeliveryGroup().Id.IsSet() || machine.GetSessionCount() == 0 {
			machinesToDelete = append(machinesToDelete, machine)
		}

		if len(machinesToDelete) == machineDeleteRequestCount {
			break
		}
	}

	machinesToDeleteCount := len(machinesToDelete)

	if machineDeleteRequestCount > machinesToDeleteCount {
		errorString := fmt.Sprintf("%d machine(s) requested to be deleted. %d machine(s) qualify for deletion.", machineDeleteRequestCount, machinesToDeleteCount)

		resp.Diagnostics.AddError(
			"Error deleting machine(s) from Machine Catalog "+catalogName,
			errorString+" Ensure machine that needs to be deleted has no active sessions.",
		)

		return err
	}

	// first put machine in maintenance mode and then perform delete
	// use batch API for both

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client, plan, false)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nCould not put machine(s) into maintenance mode before deleting them, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}

	for index, machineToDelete := range machinesToDelete {
		isMachineInMaintenanceMode := machineToDelete.GetInMaintenanceMode()

		if !isMachineInMaintenanceMode {
			// machine is not in maintenance mode. Put machine in maintenance mode first before deleting
			var updateMachineModel citrixorchestration.UpdateMachineRequestModel
			updateMachineModel.SetInMaintenanceMode(true)
			updateMachineStringBody, err := util.ConvertToString(updateMachineModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error removing Machine(s) from Machine Catalog "+catalogName,
					"An unexpected error occurred: "+err.Error(),
				)
				return err
			}

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(strconv.Itoa(index))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(fmt.Sprintf("/Machines/%s", machineToDelete.GetId()))
			batchRequestItem.SetBody(updateMachineStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
	}

	if len(batchRequestItems) > 0 {
		// If there are any machines that need to be put in maintenance mode
		var batchRequestModel citrixorchestration.BatchRequestModel
		batchRequestModel.SetItems(batchRequestItems)
		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting machine(s) from Machine Catalog "+catalogName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return err
		}

		if successfulJobs < len(batchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while putting machine(s) into maintenance mode before deleting them. %d of %d machines were put in the maintenance mode.", successfulJobs, len(batchRequestItems))
			err = fmt.Errorf(errMsg)
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog "+catalogName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return err
		}
	}

	batchApiHeaders, httpResp, err = generateBatchApiHeaders(client, plan, true)
	txId = citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nCould not delete machine(s) from machine catalog, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	batchRequestItems = []citrixorchestration.BatchRequestItemModel{}
	for index, machineToDelete := range machinesToDelete {
		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(strconv.Itoa(index))
		batchRequestItem.SetMethod(http.MethodDelete)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetRelativeUrl(fmt.Sprintf("/Machines/%s?deleteVm=true&deleteAccount=Delete", machineToDelete.GetId()))
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	batchRequestModel := citrixorchestration.BatchRequestModel{}
	batchRequestModel.SetItems(batchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting machine(s) from Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(machinesToDelete) {
		errMsg := fmt.Sprintf("An error occurred while deleting machine(s) from Machine Catalog. %d of %d machines were deleted from the Machine Catalog.", successfulJobs, len(batchRequestItems))
		err = fmt.Errorf(errMsg)
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}

	return nil
}

func addMachinesToMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan models.MachineCatalogResourceModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	addMachinesCount := int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) - catalog.GetTotalCount()

	var updateMachineAccountCreationRule citrixorchestration.UpdateMachineAccountCreationRulesRequestModel
	updateMachineAccountCreationRule.SetNamingScheme(plan.ProvisioningScheme.MachineAccountCreationRules.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(plan.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"Unsupported machine account naming scheme type.",
		)
		return err
	}
	updateMachineAccountCreationRule.SetNamingSchemeType(*namingScheme)
	updateMachineAccountCreationRule.SetDomain(plan.ProvisioningScheme.MachineAccountCreationRules.Domain.ValueString())
	updateMachineAccountCreationRule.SetOU(plan.ProvisioningScheme.MachineAccountCreationRules.Ou.ValueString())
	var addMachineRequestBody citrixorchestration.AddMachineToMachineCatalogDetailRequestModel
	addMachineRequestBody.SetMachineAccountCreationRules(updateMachineAccountCreationRule)

	addMachineRequestStringBody, err := util.ConvertToString(addMachineRequestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"An unexpected error occurred: "+err.Error(),
		)
		return err
	}

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client, plan, true)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nCould not add machine to Machine Catalog, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	for i := 0; i < int(addMachinesCount); i++ {
		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetReference(strconv.Itoa(i))
		batchRequestItem.SetRelativeUrl(fmt.Sprintf("/MachineCatalogs/%s/Machines", catalogId))
		batchRequestItem.SetBody(addMachineRequestStringBody)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}
	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(batchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding machine(s) to Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < int(addMachinesCount) {
		errMsg := fmt.Sprintf("An error occurred while adding machine(s) to the Machine Catalog. %d of %d machines were added to the Machine Catalog.", successfulJobs, addMachinesCount)
		err = fmt.Errorf(errMsg)
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}

	return nil
}
