// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSchemaForMachineCatalogResource() schema.Schema {
	return schema.Schema{
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
			"is_power_managed": schema.BoolAttribute{
				Description: "Specify if the machines in the machine catalog will be power managed.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"is_remote_pc": schema.BoolAttribute{
				Description: "Specify if this catalog is for Remote PC access.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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
				Description: "Session support type. Choose between `SingleSession` and `MultiSession`. Session support should be SingleSession when `is_remote_pc = true`",
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
				Description: "Type of Vda Upgrade. Choose between LTSR and CR. When omitted, Vda Upgrade is disabled.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"LTSR",
						"CR",
					),
				},
			},
			"provisioning_type": schema.StringAttribute{
				Description: "Specifies how the machines are provisioned in the catalog.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.PROVISIONINGTYPE_MCS),
						string(citrixorchestration.PROVISIONINGTYPE_MANUAL),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_accounts": schema.ListNestedAttribute{
				Description: "List of machine accounts to add to the catalog. Only to be used when using `provisioning_type = MANUAL`",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hypervisor": schema.StringAttribute{
							Description: "The Id of the hypervisor in which the machines reside. Required only if `is_power_managed = true`",
							Optional:    true,
						},
						"machines": schema.ListNestedAttribute{
							Description: "List of machines",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"machine_account": schema.StringAttribute{
										Description: "The Computer AD Account for the machine. Must be in the format DOMAIN\\MACHINE.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexp.MustCompile(util.SamRegex), "must be in the format DOMAIN\\MACHINE"),
										},
									},
									"machine_name": schema.StringAttribute{
										Description: "The name of the machine. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"region": schema.StringAttribute{
										Description: "**[Azure, GCP: Required]** The region in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"resource_group_name": schema.StringAttribute{
										Description: "**[Azure: Required]** The resource group in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"project_name": schema.StringAttribute{
										Description: "**[GCP: Required]** The project name in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"availability_zone": schema.StringAttribute{
										Description: "**[AWS: Required]** The availability zone in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"datacenter": schema.StringAttribute{
										Description: "**[VSphere: Required]** The datacenter in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
									"cluster": schema.StringAttribute{
										Description: "**[VSphere: Optional]** The cluster in which the machine resides. To be used only if `is_power_managed = true`",
										Optional:    true,
									},
									"host": schema.StringAttribute{
										Description: "**[VSphere: Required]** The IP address or FQDN of the host in which the machine resides. Required only if `is_power_managed = true`",
										Optional:    true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"remote_pc_ous": schema.ListNestedAttribute{
				Description: "Organizational Units to be included in the Remote PC machine catalog. Only to be used when `is_remote_pc = true`. For adding machines, use `machine_accounts`.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"include_subfolders": schema.BoolAttribute{
							Description: "Specify if subfolders should be included.",
							Required:    true,
						},
						"ou_name": schema.StringAttribute{
							Description: "Name of the OU.",
							Required:    true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"provisioning_scheme": schema.SingleNestedAttribute{
				Description: "Machine catalog provisioning scheme. Required when `provisioning_type = MCS`",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"hypervisor": schema.StringAttribute{
						Description: "Id of the hypervisor for creating the machines. Required only if using power managed machines.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						},
					},
					"hypervisor_resource_pool": schema.StringAttribute{
						Description: "Id of the hypervisor resource pool that will be used for provisioning operations.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						},
					},
					"azure_machine_config": schema.SingleNestedAttribute{
						Description: "Machine Configuration For Azure MCS catalog.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"service_offering": schema.StringAttribute{
								Description: "The Azure VM Sku to use when creating machines.",
								Required:    true,
							},
							"resource_group": schema.StringAttribute{
								Description: "The Azure Resource Group where the image VHD / managed disk / snapshot for creating machines is located.",
								Required:    true,
							},
							"master_image": schema.StringAttribute{
								Description: "The name of the virtual machine snapshot or VM template that will be used. This identifies the hard disk to be used and the default values for the memory and processors. Omit this field if you want to use gallery_image.",
								Optional:    true,
							},
							"storage_account": schema.StringAttribute{
								Description: "The Azure Storage Account where the image VHD for creating machines is located. Only applicable to Azure VHD image blob.",
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
								Description: "The Azure Storage Account Container where the image VHD for creating machines is located. Only applicable to Azure VHD image blob.",
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
							},
							"storage_type": schema.StringAttribute{
								Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"Standard_LRS",
										"StandardSSD_LRS",
										"Premium_LRS",
									),
									stringvalidator.ExactlyOneOf(path.Expressions{
										path.MatchRelative().AtParent().AtName("use_ephemeral_os_disk"),
									}...),
								},
							},
							"vda_resource_group": schema.StringAttribute{
								Description: "Designated resource group where the VDA VMs will be located on Azure.",
								Optional:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"use_managed_disks": schema.BoolAttribute{
								Description: "Indicate whether to use Azure managed disks for the provisioned virtual machine.",
								Optional:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"use_ephemeral_os_disk": schema.BoolAttribute{
								Description: "Indicate whether to use ephemeral OS disks on local virtual machine storage. They incur no cost, but the VM size used must support them and the temporary disk associated must be greater than or equal to the size of the published image.",
								Optional:    true,
								Validators: []validator.Bool{
									boolvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("use_managed_disks"),
									}...),
									boolvalidator.AlsoRequires(path.Expressions{
										path.MatchRelative().AtParent().AtName("place_image_in_gallery"),
									}...),
								},
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"license_type": schema.StringAttribute{
								Description: "Windows license type used to provision virtual machines in Azure at the base compute rate. License types include: `Windows_Client` and `Windows_Server`.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"Windows_Client",
										"Windows_Server",
									),
								},
							},
							"place_image_in_gallery": schema.SingleNestedAttribute{
								Description: "Indicate whether to use the Azure Compute Gallery to store the published image.",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"replica_ratio": schema.Int64Attribute{
										Description: "Defines the ratio of machines to gallery image version replicas.",
										Optional:    true,
										Computed:    true,
										Default:     int64default.StaticInt64(40),
										Validators: []validator.Int64{
											int64validator.Between(1, 1000),
										},
									},
									"replica_maximum": schema.Int64Attribute{
										Description: "Defines the maximum number of replicas for each gallery image version. Azure currently supports up to 10 replicas for a gallery image single version.",
										Optional:    true,
										Computed:    true,
										Default:     int64default.StaticInt64(10),
										Validators: []validator.Int64{
											int64validator.Between(1, 10),
										},
									},
								},
							},
							"disk_encryption_set_id": schema.StringAttribute{
								Description: "The Resource ID of the Disk Encryption Set (DES) representing the customer-managed key (CMK) for server-side encryption (SSE) of managed disks.",
								Optional:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"machine_profile": schema.SingleNestedAttribute{
								Description: "The name of the virtual machine template that will be used to identify the default value for the tags, virtual machine size, boot diagnostics, host cache property of OS disk, accelerated networking and availability zone." + "<br />" +
									"Required when identity_type is set to `AzureAD`",
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"machine_profile_vm_name": schema.StringAttribute{
										Description: "The name of the machine profile virtual machine.",
										Required:    true,
									},
									"machine_profile_resource_group": schema.StringAttribute{
										Description: "The resource group name where machine profile VM is located in.",
										Required:    true,
									},
								},
							},
							"writeback_cache": schema.SingleNestedAttribute{
								Description: "Write-back Cache config. Leave this empty to disable Write-back Cache. Write-back Cache requires Machine image with Write-back Cache plugin installed.",
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
					"aws_machine_config": schema.SingleNestedAttribute{
						Description: "Machine Configuration For AWS EC2 MCS catalog.",
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
						},
					},
					"gcp_machine_config": schema.SingleNestedAttribute{
						Description: "Machine Configuration For GCP MCS catalog.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"master_image": schema.StringAttribute{
								Description: "The name of the virtual machine snapshot or VM template that will be used. This identifies the hard disk to be used and the default values for the memory and processors.",
								Required:    true,
							},
							"machine_profile": schema.StringAttribute{
								Description: "The name of the virtual machine template that will be used to identify the default value for the tags, virtual machine size, boot diagnostics, host cache property of OS disk, accelerated networking and availability zone. If not specified, the VM specified in master_image will be used as template.",
								Optional:    true,
							},
							"machine_snapshot": schema.StringAttribute{
								Description: "The name of the virtual machine snapshot of a GCP VM that will be used as master image.",
								Optional:    true,
							},
							"storage_type": schema.StringAttribute{
								Description: "Storage type used for provisioned virtual machine disks on GCP. Storage types include: `pd-standar`, `pd-balanced`, `pd-ssd` and `pd-extreme`.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.OneOf(
										"pd-standard",
										"pd-balanced",
										"pd-ssd",
										"pd-extreme",
									),
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
												"pd-standard",
												"pd-balanced",
												"pd-ssd",
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
									"persist_vm": schema.BoolAttribute{
										Description: "Not supported for GCP.",
										Computed:    true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
									"storage_cost_saving": schema.BoolAttribute{
										Description: "Not supported for GCP.",
										Computed:    true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
					"machine_domain_identity": schema.SingleNestedAttribute{
						Description: "The domain identity for machines in the machine catalog." + "<br />" +
							"Required when identity_type is set to `ActiveDirectory`",
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"domain": schema.StringAttribute{
								Description: "The AD domain name for the pool. Specify this in FQDN format; for example, MyDomain.com.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(regexp.MustCompile(util.DomainFqdnRegex), "must be in FQDN format"),
								},
							},
							"domain_ou": schema.StringAttribute{
								Description: "The organization unit that computer accounts will be created into.",
								Optional:    true,
							},
							"service_account": schema.StringAttribute{
								Description: "Service account for the domain. Only the username is required; do not include the domain name.",
								Required:    true,
							},
							"service_account_password": schema.StringAttribute{
								Description: "Service account password for the domain.",
								Required:    true,
								Sensitive:   true,
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
						Description: "Specifies how the attached NICs are mapped to networks. If this parameter is omitted, provisioned VMs are created with a single NIC, which is mapped to the default network in the hypervisor resource pool.  If this parameter is supplied, machines are created with the number of NICs specified in the map, and each NIC is attached to the specified network." + "<br />" +
							"Required when `provisioning_scheme.identity_type` is `AzureAD`.",
						Optional: true,
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
					"availability_zones": schema.StringAttribute{
						Description: "The Availability Zones for provisioning virtual machines. Use a comma as a delimiter for multiple availability_zones.",
						Optional:    true,
					},
					"identity_type": schema.StringAttribute{
						Description: "The identity type of the machines to be created. Supported values are`ActiveDirectory`, `AzureAD`, and `HybridAzureAD`.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY),
								string(citrixorchestration.IDENTITYTYPE_AZURE_AD),
								string(citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD),
							),
							validators.AlsoRequiresOnValues(
								[]string{
									string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY),
								},
								path.MatchRelative().AtParent().AtName("machine_domain_identity"),
							),
							validators.AlsoRequiresOnValues(
								[]string{
									string(citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD),
								},
								path.MatchRelative().AtParent().AtName("machine_domain_identity"),
							),
							validators.AlsoRequiresOnValues(
								[]string{
									string(citrixorchestration.IDENTITYTYPE_AZURE_AD),
								},
								path.MatchRelative().AtParent().AtName("azure_machine_config"),
								path.MatchRelative().AtParent().AtName("azure_machine_config").AtName("machine_profile"),
								path.MatchRelative().AtParent().AtName("network_mapping"),
							),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
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
						},
					},
				},
			},
		},
	}
}
