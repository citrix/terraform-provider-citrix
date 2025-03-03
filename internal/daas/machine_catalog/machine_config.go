// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureMachineConfigModel struct {
	ServiceOffering types.String `tfsdk:"service_offering"`
	/** Azure Hypervisor **/
	AzureMasterImage         types.Object `tfsdk:"azure_master_image"`
	AzurePvsConfiguration    types.Object `tfsdk:"azure_pvs_config"`
	AzurePreparedImage       types.Object `tfsdk:"prepared_image"` // PreparedImageConfigModel
	MasterImageNote          types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions types.Object `tfsdk:"image_update_reboot_options"`
	VdaResourceGroup         types.String `tfsdk:"vda_resource_group"`
	StorageType              types.String `tfsdk:"storage_type"`
	UseAzureComputeGallery   types.Object `tfsdk:"use_azure_compute_gallery"`
	LicenseType              types.String `tfsdk:"license_type"`
	UseManagedDisks          types.Bool   `tfsdk:"use_managed_disks"`
	MachineProfile           types.Object `tfsdk:"machine_profile"`
	WritebackCache           types.Object `tfsdk:"writeback_cache"`
	DiskEncryptionSet        types.Object `tfsdk:"disk_encryption_set"`
	EnrollInIntune           types.Bool   `tfsdk:"enroll_in_intune"`
}

func (AzureMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration For Azure MCS and PVS Streaming catalogs.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The Azure VM Sku to use when creating machines.",
				Required:    true,
			},
			"azure_pvs_config":   AzurePvsConfigurationModel{}.GetSchema(),
			"azure_master_image": AzureMasterImageModel{}.GetSchema(),
			"prepared_image":     PreparedImageConfigModel{}.GetSchema(),
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("prepared_image")),
				},
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.StandardLRS,
						util.StandardSSDLRS,
						util.Premium_LRS,
						util.AzureEphemeralOSDisk,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = req.StateValue.ValueString() == util.AzureEphemeralOSDisk || req.PlanValue.ValueString() == util.AzureEphemeralOSDisk
						},
						"Updating storage_type is not allowed when using Azure Ephemeral OS Disk.",
						"Updating storage_type is not allowed when using Azure Ephemeral OS Disk.",
					),
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueString() != req.PlanValue.ValueString()
						},
						"Updating storage_type is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating storage_type is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"use_azure_compute_gallery": AzureComputeGallerySettings{}.GetSchema(),
			"license_type": schema.StringAttribute{
				Description: "Windows license type used to provision virtual machines in Azure at the base compute rate. License types include: `Windows_Client` and `Windows_Server`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.WindowsClientLicenseType,
						util.WindowsServerLicenseType,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueString() != req.PlanValue.ValueString()
						},
						"Updating license_type is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating license_type is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"enroll_in_intune": schema.BoolAttribute{
				Description: "Specify whether to enroll machines in Microsoft Intune. Use this property only when `identity_type` is set to `AzureAD`.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"disk_encryption_set": util.AzureDiskEncryptionSetModel{}.GetSchema(),
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
			"machine_profile": util.AzureMachineProfileModel{}.GetSchema(),
			"writeback_cache": AzureWritebackCacheModel{}.GetSchema(),
		},
	}
}

func (AzureMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return AzureMachineConfigModel{}.GetSchema().Attributes
}

type AwsMachineConfigModel struct {
	ServiceOffering          types.String `tfsdk:"service_offering"`
	MasterImage              types.String `tfsdk:"master_image"`
	MasterImageNote          types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions types.Object `tfsdk:"image_update_reboot_options"`
	/** AWS Hypervisor **/
	ImageAmi       types.String `tfsdk:"image_ami"`
	SecurityGroups types.List   `tfsdk:"security_groups"` // List[String]
	TenancyType    types.String `tfsdk:"tenancy_type"`
}

func (AwsMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration For AWS EC2 MCS catalog.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The AWS VM Sku to use when creating machines.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(util.AwsEc2InstanceTypeRegex),
						"must follow AWS EC2 instance type naming convention in lower case. Eg: t2.micro, m5.large, etc.",
					),
				},
			},
			"master_image": schema.StringAttribute{
				Description: "The name of the virtual machine image that will be used.",
				Required:    true,
			},
			"image_ami": schema.StringAttribute{
				Description: "AMI of the AWS image to be used as the template image for the machine catalog.",
				Required:    true,
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"security_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Security groups to associate with the machine. When omitted, the default security group of the VPC will be used by default.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"tenancy_type": schema.StringAttribute{
				Description: "Tenancy type of the machine. Choose between `Shared`, `Instance` and `Host`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Shared",
						"Instance",
						"Host",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (AwsMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return AwsMachineConfigModel{}.GetSchema().Attributes
}

type GcpMachineConfigModel struct {
	MasterImage              types.String `tfsdk:"master_image"`
	MasterImageNote          types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions types.Object `tfsdk:"image_update_reboot_options"`
	/** GCP Hypervisor **/
	MachineProfile  types.String `tfsdk:"machine_profile"`
	MachineSnapshot types.String `tfsdk:"machine_snapshot"`
	StorageType     types.String `tfsdk:"storage_type"`
	WritebackCache  types.Object `tfsdk:"writeback_cache"` // GcpWritebackCacheModel
}

func (GcpMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration For GCP MCS catalog.",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"master_image": schema.StringAttribute{
				Description: "The name of the virtual machine snapshot or VM template that will be used. This identifies the hard disk to be used and the default values for the memory and processors.",
				Required:    true,
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
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
			"writeback_cache": GcpWritebackCacheModel{}.GetSchema(),
		},
	}
}

func (GcpMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return GcpMachineConfigModel{}.GetSchema().Attributes
}

type VsphereMachineConfigModel struct {
	/** vSphere Hypervisor **/
	MasterImageVm                types.String `tfsdk:"master_image_vm"`
	VspherePreparedImage         types.Object `tfsdk:"prepared_image"` // PreparedImageConfigModel
	ResourcePoolPath             types.String `tfsdk:"resource_pool_path"`
	ImageSnapshot                types.String `tfsdk:"image_snapshot"`
	MasterImageNote              types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions     types.Object `tfsdk:"image_update_reboot_options"`
	CpuCount                     types.Int64  `tfsdk:"cpu_count"`
	MemoryMB                     types.Int64  `tfsdk:"memory_mb"`
	WritebackCache               types.Object `tfsdk:"writeback_cache"` // VsphereAndSCVMMWritebackCacheModel
	MachineProfile               types.String `tfsdk:"machine_profile"`
	UseFullDiskCloneProvisioning types.Bool   `tfsdk:"use_full_disk_clone_provisioning"`
}

func (VsphereMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration for vSphere MCS catalog.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"master_image_vm": schema.StringAttribute{
				Description: "The name of the virtual machine that will be used as master image. This property is case sensitive.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(util.NoPathRegex),
						"must not contain any path.",
					),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("prepared_image")),
				},
			},
			"resource_pool_path": schema.StringAttribute{
				Description: "The Resource Pool path under which the `master_image_vm` is located. This property is case sensitive.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_snapshot": schema.StringAttribute{
				Description: "The Snapshot of the virtual machine specified in `master_image_vm`. Specify the relative path of the snapshot. Eg: snaphost-1/snapshot-2/snapshot-3. This property is case sensitive.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("prepared_image")),
				},
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"prepared_image":              PreparedImageConfigModel{}.GetSchema(),
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"cpu_count": schema.Int64Attribute{
				Description: "The number of processors that virtual machines created from the provisioning scheme should use.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"memory_mb": schema.Int64Attribute{
				Description: "The maximum amount of memory that virtual machines created from the provisioning scheme should use.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(4),
				},
			},
			"use_full_disk_clone_provisioning": schema.BoolAttribute{
				Description: "Specify if virtual machines created from the provisioning scheme should be created using the dedicated full disk clone feature. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"writeback_cache": VsphereAndSCVMMWritebackCacheModel{}.GetSchema(),
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
			},
		},
	}
}

func (VsphereMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return VsphereMachineConfigModel{}.GetSchema().Attributes
}

type XenserverMachineConfigModel struct {
	/** XenServer Hypervisor **/
	MasterImageVm                types.String `tfsdk:"master_image_vm"`
	ImageSnapshot                types.String `tfsdk:"image_snapshot"`
	MasterImageNote              types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions     types.Object `tfsdk:"image_update_reboot_options"`
	CpuCount                     types.Int64  `tfsdk:"cpu_count"`
	MemoryMB                     types.Int64  `tfsdk:"memory_mb"`
	WritebackCache               types.Object `tfsdk:"writeback_cache"` // XenserverWritebackCacheModel
	UseFullDiskCloneProvisioning types.Bool   `tfsdk:"use_full_disk_clone_provisioning"`
}

func (XenserverMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration For XenServer MCS catalog.",
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
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"cpu_count": schema.Int64Attribute{
				Description: "Number of CPU cores for the VDA VMs.",
				Required:    true,
			},
			"memory_mb": schema.Int64Attribute{
				Description: "Size of the memory in MB for the VDA VMs.",
				Required:    true,
			},
			"writeback_cache": XenserverWritebackCacheModel{}.GetSchema(),
			"use_full_disk_clone_provisioning": schema.BoolAttribute{
				Description: "Specify if virtual machines created from the provisioning scheme should be created using the dedicated full disk clone feature. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (XenserverMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return XenserverMachineConfigModel{}.GetSchema().Attributes
}

type NutanixMachineConfigModel struct {
	Container                types.String `tfsdk:"container"`
	MasterImage              types.String `tfsdk:"master_image"`
	MasterImageNote          types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions types.Object `tfsdk:"image_update_reboot_options"`
	CpuCount                 types.Int64  `tfsdk:"cpu_count"`
	CoresPerCpuCount         types.Int64  `tfsdk:"cores_per_cpu_count"`
	MemoryMB                 types.Int64  `tfsdk:"memory_mb"`
}

func (NutanixMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration For Nutanix MCS catalog.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"container": schema.StringAttribute{
				Description: "The name of the container where the virtual machines' identity disks will be placed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"master_image": schema.StringAttribute{
				Description: "The name of the master image that will be the template for all virtual machines in this catalog.",
				Required:    true,
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"cpu_count": schema.Int64Attribute{
				Description: "The number of processors that virtual machines created from the provisioning scheme should use.",
				Required:    true,
			},
			"cores_per_cpu_count": schema.Int64Attribute{
				Description: "The number of cores per processor that virtual machines created from the provisioning scheme should use.",
				Required:    true,
			},
			"memory_mb": schema.Int64Attribute{
				Description: "The maximum amount of memory that virtual machines created from the provisioning scheme should use.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (NutanixMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return NutanixMachineConfigModel{}.GetSchema().Attributes
}

type SCVMMMachineConfigModel struct {
	MasterImage                  types.String `tfsdk:"master_image"`
	ImageSnapshot                types.String `tfsdk:"image_snapshot"`
	MasterImageNote              types.String `tfsdk:"master_image_note"`
	ImageUpdateRebootOptions     types.Object `tfsdk:"image_update_reboot_options"`
	CpuCount                     types.Int64  `tfsdk:"cpu_count"`
	MemoryMB                     types.Int64  `tfsdk:"memory_mb"`
	UseFullDiskCloneProvisioning types.Bool   `tfsdk:"use_full_disk_clone_provisioning"`
	WritebackCache               types.Object `tfsdk:"writeback_cache"` // VsphereAndSCVMMWritebackCacheModel
}

func (SCVMMMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine Configuration for SCVMM MCS catalog.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"master_image": schema.StringAttribute{
				Description: "The name of the virtual machine that will be used as master image.",
				Required:    true,
			},
			"image_snapshot": schema.StringAttribute{
				Description: "The Snapshot of the virtual machine specified in `master_image`. Specify the relative path of the snapshot. Eg: snaphost-1/snapshot-2/snapshot-3. This property is case sensitive.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"cpu_count": schema.Int64Attribute{
				Description: "The number of processors that virtual machines created from the provisioning scheme should use.",
				Required:    true,
			},
			"memory_mb": schema.Int64Attribute{
				Description: "The maximum amount of memory that virtual machines created from the provisioning scheme should use.",
				Required:    true,
			},
			"use_full_disk_clone_provisioning": schema.BoolAttribute{
				Description: "Specify if virtual machines created from the provisioning scheme should be created using the dedicated full disk clone feature. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"writeback_cache": VsphereAndSCVMMWritebackCacheModel{}.GetSchema(),
		},
	}
}

func (SCVMMMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return SCVMMMachineConfigModel{}.GetSchema().Attributes
}

type AzureMasterImageModel struct {
	ResourceGroup      types.String `tfsdk:"resource_group"`
	SharedSubscription types.String `tfsdk:"shared_subscription"`
	MasterImage        types.String `tfsdk:"master_image"`
	StorageAccount     types.String `tfsdk:"storage_account"`
	Container          types.String `tfsdk:"container"`
	GalleryImage       types.Object `tfsdk:"gallery_image"`
}

func (AzureMasterImageModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Details of the Azure Image to use for creating machines.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"resource_group": schema.StringAttribute{
				Description: "The Azure Resource Group where the image VHD / managed disk / snapshot for creating machines is located.",
				Required:    true,
			},
			"shared_subscription": schema.StringAttribute{
				Description: "The Azure Subscription ID where the image VHD / managed disk / snapshot for creating machines is located. Only required if the image is not in the same subscription of the hypervisor.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
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
			"gallery_image": util.GalleryImageModel{}.GetSchema(),
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = !req.StateValue.IsUnknown() && !req.StateValue.IsNull() && req.PlanValue.IsNull()
				},
				"Changing machine catalog image type requires replacing the machine catalog resource.",
				"Changing machine catalog image type requires replacing the machine catalog resource.",
			),
		},
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("prepared_image")),
		},
	}
}

func (AzureMasterImageModel) GetAttributes() map[string]schema.Attribute {
	return AzureMasterImageModel{}.GetSchema().Attributes
}

type AzurePvsConfigurationModel struct {
	PvsSiteId  types.String `tfsdk:"pvs_site_id"`
	PvsVdiskId types.String `tfsdk:"pvs_vdisk_id"`
}

func (AzurePvsConfigurationModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "PVS Configuration to create machine catalog using PVSStreaming.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"pvs_site_id": schema.StringAttribute{
				Description: "The id of the PVS site to use for creating machines.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pvs_vdisk_id": schema.StringAttribute{
				Description: "The id of the PVS vDisk to use for creating machines.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Validators: []validator.Object{
			objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("azure_master_image"), path.MatchRelative().AtParent().AtName("prepared_image")),
		},
	}
}

func (AzurePvsConfigurationModel) GetAttributes() map[string]schema.Attribute {
	return AzurePvsConfigurationModel{}.GetSchema().Attributes
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

func (AzureWritebackCacheModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Write-back Cache config. Leave this empty to disable Write-back Cache. Write-back Cache requires Machine image with Write-back Cache plugin installed.",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"persist_wbc": schema.BoolAttribute{
				Description: "Persist Write-back Cache",
				Optional:    true,
			},
			"wbc_disk_storage_type": schema.StringAttribute{
				Description: "Type of the storage for Write-back Cache disk. Choose between `Standard_LRS`, `StandardSSD_LRS`, and `Premium_LRS`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.StandardSSDLRS,
						util.StandardLRS,
						util.Premium_LRS,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueString() != req.PlanValue.ValueString()
						},
						"Updating wbc_disk_storage_type is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating wbc_disk_storage_type is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"persist_os_disk": schema.BoolAttribute{
				Description: "Persist the OS disk when power cycling the non-persistent provisioned virtual machine.",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueBool() != req.PlanValue.ValueBool()
						},
						"Updating persist_os_disk is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating persist_os_disk is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"persist_vm": schema.BoolAttribute{
				Description: "Persist the non-persistent provisioned virtual machine in Azure environments when power cycling. This property only applies when the PersistOsDisk property is set to True.",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueBool() != req.PlanValue.ValueBool()
						},
						"Updating persist_vm is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating persist_vm is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"storage_cost_saving": schema.BoolAttribute{
				Description: "Save storage cost by downgrading the storage type of the disk to Standard HDD when VM shut down.",
				Optional:    true,
			},
			"writeback_cache_disk_size_gb": schema.Int64Attribute{
				Description: "The size in GB of any temporary storage disk used by the write back cache.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !checkIfCatalogAttributeCanBeUpdated(ctx, req.State) && req.StateValue.ValueInt64() != req.PlanValue.ValueInt64()
						},
						"Updating writeback_cache_disk_size_gb is not allowed for catalogs with PVS Streaming provisioning type.",
						"Updating writeback_cache_disk_size_gb is not allowed for catalogs with PVS Streaming provisioning type.",
					),
				},
			},
			"writeback_cache_memory_size_mb": schema.Int64Attribute{
				Description: "The size of the in-memory write back cache in MB.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (AzureWritebackCacheModel) GetAttributes() map[string]schema.Attribute {
	return AzureWritebackCacheModel{}.GetSchema().Attributes
}

type AzureComputeGallerySettings struct {
	ReplicaRatio   types.Int64 `tfsdk:"replica_ratio"`
	ReplicaMaximum types.Int64 `tfsdk:"replica_maximum"`
}

func (AzureComputeGallerySettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Use this to place prepared image in Azure Compute Gallery. Required when `storage_type = Azure_Ephemeral_OS_Disk`." +
			"\n\n~> **Please Note** `use_azure_compute_gallery` cannot be specified when the prepared image is using a shared image gallery. The machine catalog will inherit the azure compute gallery settings of the prepared image.",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"replica_ratio": schema.Int64Attribute{
				Description: "The ratio of virtual machines to image replicas that you want Azure to keep.",
				Required:    true,
			},
			"replica_maximum": schema.Int64Attribute{
				Description: "The maximum number of image replicas that you want Azure to keep.",
				Required:    true,
			},
		},
	}
}

func (AzureComputeGallerySettings) GetAttributes() map[string]schema.Attribute {
	return AzureComputeGallerySettings{}.GetSchema().Attributes
}

type GcpWritebackCacheModel struct {
	PersistWBC                 types.Bool   `tfsdk:"persist_wbc"`
	WBCDiskStorageType         types.String `tfsdk:"wbc_disk_storage_type"`
	PersistOsDisk              types.Bool   `tfsdk:"persist_os_disk"`
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
}

func (GcpWritebackCacheModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
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
			},
			"persist_os_disk": schema.BoolAttribute{
				Description: "Persist the OS disk when power cycling the non-persistent provisioned virtual machine.",
				Required:    true,
			},
			"writeback_cache_disk_size_gb": schema.Int64Attribute{
				Description: "The size in GB of any temporary storage disk used by the write back cache.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"writeback_cache_memory_size_mb": schema.Int64Attribute{
				Description: "The size of the in-memory write back cache in MB.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (GcpWritebackCacheModel) GetAttributes() map[string]schema.Attribute {
	return GcpWritebackCacheModel{}.GetSchema().Attributes
}

type XenserverWritebackCacheModel struct {
	WriteBackCacheDiskSizeGB   types.Int64 `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64 `tfsdk:"writeback_cache_memory_size_mb"`
}

func (XenserverWritebackCacheModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Write-back Cache config. Leave this empty to disable Write-back Cache.",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"writeback_cache_disk_size_gb": schema.Int64Attribute{
				Description: "The size in GB of any temporary storage disk used by the write back cache.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"writeback_cache_memory_size_mb": schema.Int64Attribute{
				Description: "The size of the in-memory write back cache in MB.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (XenserverWritebackCacheModel) GetAttributes() map[string]schema.Attribute {
	return XenserverWritebackCacheModel{}.GetSchema().Attributes
}

type VsphereAndSCVMMWritebackCacheModel struct {
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
	WriteBackCacheDriveLetter  types.String `tfsdk:"writeback_cache_drive_letter"`
}

func (VsphereAndSCVMMWritebackCacheModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Write-back Cache config. Leave this empty to disable Write-back Cache.",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"writeback_cache_disk_size_gb": schema.Int64Attribute{
				Description: "The size in GB of any temporary storage disk used by the write back cache.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"writeback_cache_memory_size_mb": schema.Int64Attribute{
				Description: "The size of the in-memory write back cache in MB.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"writeback_cache_drive_letter": schema.StringAttribute{
				Description: "The drive letter assigned for write back cache disk.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1),
				},
			},
		},
	}
}

func (VsphereAndSCVMMWritebackCacheModel) GetAttributes() map[string]schema.Attribute {
	return VsphereAndSCVMMWritebackCacheModel{}.GetSchema().Attributes
}

type ImageUpdateRebootOptionsModel struct {
	RebootDuration        types.Int64  `tfsdk:"reboot_duration"`
	WarningDuration       types.Int64  `tfsdk:"warning_duration"`
	WarningMessage        types.String `tfsdk:"warning_message"`
	WarningRepeatInterval types.Int64  `tfsdk:"warning_repeat_interval"`
}

func (ImageUpdateRebootOptionsModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The options for how rebooting is performed for image update. When omitted, image update on the VDAs will be performed on next shutdown.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"reboot_duration": schema.Int64Attribute{
				Description: "Approximate maximum duration over which the reboot cycle runs, in minutes. " +
					"-> **Note** Set to `-1` to skip reboot, and perform image update on the VDAs on next shutdown. " +
					"Set to `0` to reboot all machines immediately.",
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(-1),
				},
			},
			"warning_duration": schema.Int64Attribute{
				Description: "Time in minutes prior to a machine reboot at which a warning message is displayed in all user sessions on that machine. When omitted, no warning about reboot will be displayed in user session." +
					"-> **Note** When `reboot_duration` is set to `-1`, if a warning message should be displayed, `warning_duration` has to be set to `-1` to show the warning message immediately." +
					"-> **Note** When `reboot_duration` is not set to `-1`, `warning_duration` cannot be set to `-1`.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(-1),
					int64validator.NoneOf(0),
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("warning_message"),
					}...),
				},
			},
			"warning_message": schema.StringAttribute{
				Description: "Warning message displayed in user sessions on a machine scheduled for a reboot. The optional pattern '%m%' is replaced by the number of minutes until the reboot.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"warning_repeat_interval": schema.Int64Attribute{
				Description: "Number of minutes to wait before showing the reboot warning message again.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("warning_duration"),
					}...),
				},
			},
		},
	}
}

func (ImageUpdateRebootOptionsModel) GetAttributes() map[string]schema.Attribute {
	return ImageUpdateRebootOptionsModel{}.GetSchema().Attributes
}

func (rebootOptions ImageUpdateRebootOptionsModel) ValidateConfig(diagnostics *diag.Diagnostics) {
	rebootDuration := int32(rebootOptions.RebootDuration.ValueInt64())
	warningDuration := int32(rebootOptions.WarningDuration.ValueInt64())
	if rebootDuration == -1 && warningDuration > 0 {
		diagnostics.AddAttributeError(
			path.Root("warning_duration"),
			"Invalid Reboot Warning Duration",
			"warning_duration can only be set to -1 or 0 when reboot_duration is set to -1.",
		)
	}
	if rebootDuration != -1 && warningDuration == -1 {
		diagnostics.AddAttributeError(
			path.Root("warning_duration"),
			"Invalid Reboot Warning Duration",
			"warning_duration cannot be set to -1 when reboot_duration is not set to -1.",
		)
	}
	if !rebootOptions.WarningRepeatInterval.IsNull() && rebootOptions.WarningRepeatInterval.ValueInt64() >= rebootOptions.WarningDuration.ValueInt64() {
		diagnostics.AddAttributeError(
			path.Root("warning_repeat_interval"),
			"Invalid Reboot Warning Repeat Interval",
			"warning_repeat_interval must be shorter than warning_duration.",
		)
	}
}

func (mc *AzureMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel, provisioningType *citrixorchestration.ProvisioningType) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	if provScheme.GetServiceOffering() != "" {
		mc.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	}

	// Refresh Master Image for non PVS catalogs
	if *provisioningType != citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING {
		if provScheme.CurrentImageVersion != nil {
			currentImage := provScheme.GetCurrentImageVersion()
			imageVersion := currentImage.GetImageVersion()
			imageDefinition := imageVersion.GetImageDefinition()
			preparedImageConfig := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, mc.AzurePreparedImage)
			preparedImageConfig.ImageDefinition = types.StringValue(imageDefinition.GetId())
			preparedImageConfig.ImageVersion = types.StringValue(imageVersion.GetId())
			mc.AzurePreparedImage = util.TypedObjectToObjectValue(ctx, diagnostics, preparedImageConfig)

			if attributesMap, err := util.ResourceAttributeMapFromObject(AzureMasterImageModel{}); err == nil {
				mc.AzureMasterImage = types.ObjectNull(attributesMap)
			} else {
				diagnostics.AddWarning("Error when creating null AzureMasterImageModel", err.Error())
			}
		} else {
			masterImage := provScheme.GetMasterImage()
			azureMasterImage := util.ObjectValueToTypedObject[AzureMasterImageModel](ctx, diagnostics, mc.AzureMasterImage)
			masterImageXdPath := masterImage.GetXDPath()
			if masterImageXdPath != "" {
				segments := strings.Split(masterImage.GetXDPath(), "\\")
				lastIndex := len(segments)
				resourceTag := strings.Split(segments[lastIndex-1], ".")
				resourceType := resourceTag[len(resourceTag)-1]

				if strings.EqualFold(resourceType, util.ImageVersionResourceType) {
					azureMasterImage.GalleryImage,
						azureMasterImage.ResourceGroup,
						azureMasterImage.SharedSubscription =
						util.ParseMasterImageToUpdateGalleryImageModel(ctx, diagnostics, azureMasterImage.GalleryImage, masterImage, segments, lastIndex)

					// Clear other master image details
					azureMasterImage.MasterImage = types.StringNull()
					azureMasterImage.StorageAccount = types.StringNull()
					azureMasterImage.Container = types.StringNull()
				} else {
					azureMasterImage.MasterImage,
						azureMasterImage.ResourceGroup,
						azureMasterImage.SharedSubscription,
						azureMasterImage.GalleryImage,
						azureMasterImage.StorageAccount,
						azureMasterImage.Container =
						util.ParseMasterImageToUpdateAzureImageSpecs(ctx, diagnostics, resourceType, masterImage, segments, lastIndex)
				}
			}

			mc.AzureMasterImage = util.TypedObjectToObjectValue(ctx, diagnostics, azureMasterImage)
			// Refresh Master Image Note
			currentDiskImage := provScheme.GetCurrentDiskImage()
			mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

			if attributesMap, err := util.ResourceAttributeMapFromObject(PreparedImageConfigModel{}); err == nil {
				mc.AzurePreparedImage = types.ObjectNull(attributesMap)
			} else {
				diagnostics.AddWarning("Error when creating null PreparedImageConfigModel", err.Error())
			}
		}
	} else {
		azurePvsConfiguration := util.ObjectValueToTypedObject[AzurePvsConfigurationModel](ctx, diagnostics, mc.AzurePvsConfiguration)
		// Set values for PVS Streaming catalogs
		if provScheme.HasPVSSite() {
			azurePvsConfiguration.PvsSiteId = types.StringValue(provScheme.GetPVSSite())
		}

		if provScheme.HasPVSVDisk() {
			azurePvsConfiguration.PvsVdiskId = types.StringValue(provScheme.GetPVSVDisk())
		}
		mc.AzurePvsConfiguration = util.TypedObjectToObjectValue(ctx, diagnostics, azurePvsConfiguration)
		// Set Master Image Note as empty for PVS Streaming catalogs
		mc.MasterImageNote = types.StringValue("")
	}

	// Refresh Machine Profile
	if provScheme.MachineProfile != nil {
		machineProfile := provScheme.GetMachineProfile()
		machineProfileModel := util.ParseAzureMachineProfileResponseToModel(machineProfile)
		mc.MachineProfile = util.TypedObjectToObjectValue(ctx, diagnostics, machineProfileModel)
	} else {
		if attributesMap, err := util.ResourceAttributeMapFromObject(util.AzureMachineProfileModel{}); err == nil {
			mc.MachineProfile = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null AzureMachineProfileModel", err.Error())
		}
	}

	// Refresh Writeback Cache
	azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, diagnostics, mc.WritebackCache)
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		azureWbcModel.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			azureWbcModel.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		// default bool values to false because Orchestration won't return them in the custom properties
		azureWbcModel.PersistOsDisk = types.BoolValue(false)
		azureWbcModel.PersistVm = types.BoolValue(false)
		azureWbcModel.PersistWBC = types.BoolValue(false)
		azureWbcModel.StorageCostSaving = types.BoolValue(false)
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
	isUseEphemeralOsDiskSet := false
	for _, stringPair := range customProperties {
		switch stringPair.GetName() {
		case "StorageType":
			if !isUseEphemeralOsDiskSet {
				mc.StorageType = types.StringValue(stringPair.GetValue())
			}
		case "StorageAccountType":
			if !isUseEphemeralOsDiskSet {
				mc.StorageType = types.StringValue(stringPair.GetValue())
			}
		case "UseManagedDisks":
			mc.UseManagedDisks = util.StringToTypeBool(stringPair.GetValue())
		case "ResourceGroups":
			mc.VdaResourceGroup = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			azureWbcModel.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
		case "PersistWBC":
			azureWbcModel.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
		case "PersistOsDisk":
			azureWbcModel.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
		case "PersistVm":
			azureWbcModel.PersistVm = util.StringToTypeBool(stringPair.GetValue())
		case "StorageTypeAtShutdown":
			azureWbcModel.StorageCostSaving = types.BoolValue(true)
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
			diskEncryptionSetModel := util.ObjectValueToTypedObject[util.AzureDiskEncryptionSetModel](ctx, diagnostics, mc.DiskEncryptionSet)
			diskEncryptionSetModel = util.RefreshDiskEncryptionSetModel(diskEncryptionSetModel, desId)

			mc.DiskEncryptionSet = util.TypedObjectToObjectValue(ctx, diagnostics, diskEncryptionSetModel)

			isDesSet = true
		case "SharedImageGalleryReplicaRatio":
			if stringPair.GetValue() != "" {
				isUseSharedImageGallerySet = true
				azureComputeGallerySettingsModel := util.ObjectValueToTypedObject[AzureComputeGallerySettings](ctx, diagnostics, mc.UseAzureComputeGallery)

				replicaRatio, _ := strconv.Atoi(stringPair.GetValue())
				azureComputeGallerySettingsModel.ReplicaRatio = types.Int64Value(int64(replicaRatio))
				mc.UseAzureComputeGallery = util.TypedObjectToObjectValue(ctx, diagnostics, azureComputeGallerySettingsModel)
			}
		case "SharedImageGalleryReplicaMaximum":
			if stringPair.GetValue() != "" {
				isUseSharedImageGallerySet = true
				azureComputeGallerySettingsModel := util.ObjectValueToTypedObject[AzureComputeGallerySettings](ctx, diagnostics, mc.UseAzureComputeGallery)
				replicaMaximum, _ := strconv.Atoi(stringPair.GetValue())
				azureComputeGallerySettingsModel.ReplicaMaximum = types.Int64Value(int64(replicaMaximum))
				mc.UseAzureComputeGallery = util.TypedObjectToObjectValue(ctx, diagnostics, azureComputeGallerySettingsModel)
			}
		case "UseEphemeralOsDisk":
			if strings.EqualFold(stringPair.GetValue(), "true") {
				mc.StorageType = types.StringValue(util.AzureEphemeralOSDisk)
				isUseEphemeralOsDiskSet = true
			}
		default:
		}
	}

	if wbcDiskSize != 0 {
		// Finish refresh Writeback Cache
		mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, azureWbcModel)
	}

	if !isLicenseTypeSet && !mc.LicenseType.IsNull() {
		mc.LicenseType = types.StringNull()
	}

	if !isDesSet && !mc.DiskEncryptionSet.IsNull() {
		if attributesMap, err := util.ResourceAttributeMapFromObject(util.AzureDiskEncryptionSetModel{}); err == nil {
			mc.DiskEncryptionSet = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null AzureDiskEcryptionSetModel", err.Error())
		}
	}

	if !isUseSharedImageGallerySet && !mc.UseAzureComputeGallery.IsNull() {
		if attributesMap, err := util.ResourceAttributeMapFromObject(AzureComputeGallerySettings{}); err == nil {
			mc.UseAzureComputeGallery = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null AzureComputeGallerySettings", err.Error())
		}
	}
}

func (mc *AwsMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel) {
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
	mc.ImageAmi = types.StringValue(strings.TrimSuffix((strings.Split(masterImage.GetName(), " (")[1]), ")"))

	// Refresh Master Image Note
	currentDiskImage := provScheme.GetCurrentDiskImage()
	mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

	// Refresh Security Group
	securityGroups := provScheme.GetSecurityGroups()
	mc.SecurityGroups = util.StringArrayToStringList(ctx, diagnostics, securityGroups)

	// Refresh Tenancy Type
	tenancyType := provScheme.GetTenancyType()
	mc.TenancyType = types.StringValue(tenancyType)
}

func (mc *GcpMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel) {
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
			mc.MasterImage = types.StringValue(strings.TrimSuffix(segments[lastIndex-2], ".snapshot"))
		} else {
			// If path slices equals to 4, that means a VM was used for the catalog
			mc.MasterImage = types.StringValue(masterImage.GetName())
		}
	}

	// Refresh Master Image Note
	currentDiskImage := provScheme.GetCurrentDiskImage()
	mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

	// Refresh Machine Profile
	machineProfile := provScheme.GetMachineProfile()
	if machineProfileName := machineProfile.GetName(); machineProfileName != "" {
		mc.MachineProfile = types.StringValue(machineProfileName)
	}

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	writebackCache := util.ObjectValueToTypedObject[GcpWritebackCacheModel](ctx, diagnostics, mc.WritebackCache)

	if wbcDiskSize != 0 {
		writebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			writebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		// default bool values to false because Orchestration won't return them in the custom properties
		writebackCache.PersistOsDisk = types.BoolValue(false)
		writebackCache.PersistWBC = types.BoolValue(false)
	}

	mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, writebackCache)
	writebackCache = util.ObjectValueToTypedObject[GcpWritebackCacheModel](ctx, diagnostics, mc.WritebackCache)
	//Refresh custom properties
	customProperties := provScheme.GetCustomProperties()
	for _, stringPair := range customProperties {
		switch stringPair.GetName() {
		case "StorageType":
			mc.StorageType = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			writebackCache.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
		case "PersistWBC":
			writebackCache.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
		case "PersistOsDisk":
			writebackCache.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
		default:
		}
	}
	mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, writebackCache)
}

func (mc *VsphereMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	if provScheme.CurrentImageVersion != nil {
		// refresh image version
		currentImage := provScheme.GetCurrentImageVersion()
		imageVersion := currentImage.GetImageVersion()
		imageDefinition := imageVersion.GetImageDefinition()
		preparedImageConfig := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, mc.VspherePreparedImage)
		preparedImageConfig.ImageDefinition = types.StringValue(imageDefinition.GetId())
		preparedImageConfig.ImageVersion = types.StringValue(imageVersion.GetId())
		mc.VspherePreparedImage = util.TypedObjectToObjectValue(ctx, diagnostics, preparedImageConfig)
		// Set default/null values for other properties
		mc.MasterImageVm = types.StringNull()
		mc.ResourcePoolPath = types.StringValue("")
		mc.ImageSnapshot = types.StringNull()
	} else {
		// Refresh Master Image
		masterImage, imageSnapshot, resourcePoolPath := parseOnPremImagePath(catalog)
		mc.MasterImageVm = types.StringValue(masterImage)
		mc.ImageSnapshot = types.StringValue(imageSnapshot)
		mc.ResourcePoolPath = types.StringValue(resourcePoolPath)

		// Refresh Master Image Note
		currentDiskImage := provScheme.GetCurrentDiskImage()
		mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())
		if attributesMap, err := util.ResourceAttributeMapFromObject(PreparedImageConfigModel{}); err == nil {
			mc.VspherePreparedImage = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null PreparedImageConfigModel", err.Error())
		}
	}

	// Refresh Memory
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		writebackCache := VsphereAndSCVMMWritebackCacheModel{}
		writebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			writebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		if provScheme.GetWriteBackCacheDriveLetter() != "" {
			writebackCache.WriteBackCacheDriveLetter = types.StringValue(provScheme.GetWriteBackCacheDriveLetter())
		}
		mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, writebackCache)
	}

	machineProfile := provScheme.GetMachineProfile()

	if machineProfileXdPath := machineProfile.GetXDPath(); machineProfileXdPath != "" {
		machineProfileParts := strings.Split(machineProfileXdPath, "\\")
		machineProfileName := machineProfileParts[len(machineProfileParts)-1]
		machineProfileTemplateName := strings.TrimSuffix(machineProfileName, ".template")
		mc.MachineProfile = types.StringValue(machineProfileTemplateName)
	}

	mc.UseFullDiskCloneProvisioning = types.BoolValue(provScheme.GetUseFullDiskCloneProvisioning())
}

func (mc *XenserverMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	// Refresh Service Offering
	provScheme := catalog.GetProvisioningScheme()
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))

	masterImage, imageSnapshot, _ := parseOnPremImagePath(catalog)
	mc.MasterImageVm = types.StringValue(masterImage)
	mc.ImageSnapshot = types.StringValue(imageSnapshot)

	// Refresh Master Image Note
	currentDiskImage := provScheme.GetCurrentDiskImage()
	mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	writebackCache := util.ObjectValueToTypedObject[XenserverWritebackCacheModel](ctx, diagnostics, mc.WritebackCache)
	if wbcDiskSize != 0 {
		writebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			writebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, writebackCache)
	}

	mc.UseFullDiskCloneProvisioning = types.BoolValue(provScheme.GetUseFullDiskCloneProvisioning())
}

func (mc *NutanixMachineConfigModel) RefreshProperties(catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	// Refresh Master Image
	masterImage := provScheme.GetMasterImage()
	mc.MasterImage = types.StringValue(masterImage.GetName())

	// Refresh Master Image Note
	currentDiskImage := provScheme.GetCurrentDiskImage()
	mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

	// Refresh Memory
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))
	mc.CoresPerCpuCount = types.Int64Value(int64(provScheme.GetCoresPerCpuCount()))
	mc.Container = types.StringValue(provScheme.GetNutanixContainer())
}

func (mc *SCVMMMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel) {
	provScheme := catalog.GetProvisioningScheme()

	// Refresh Master Image
	masterImage, imageSnapshot, _ := parseOnPremImagePath(catalog)
	mc.MasterImage = types.StringValue(masterImage)
	mc.ImageSnapshot = types.StringValue(imageSnapshot)

	// Refresh Master Image Note
	currentDiskImage := provScheme.GetCurrentDiskImage()
	mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())

	// Refresh Memory
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))

	// Refresh Writeback Cache
	wbcDiskSize := provScheme.GetWriteBackCacheDiskSizeGB()
	wbcMemorySize := provScheme.GetWriteBackCacheMemorySizeMB()
	if wbcDiskSize != 0 {
		writebackCache := VsphereAndSCVMMWritebackCacheModel{}
		writebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if wbcMemorySize != 0 {
			writebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
		if provScheme.GetWriteBackCacheDriveLetter() != "" {
			writebackCache.WriteBackCacheDriveLetter = types.StringValue(provScheme.GetWriteBackCacheDriveLetter())
		}
		mc.WritebackCache = util.TypedObjectToObjectValue(ctx, diagnostics, writebackCache)
	}

	mc.UseFullDiskCloneProvisioning = types.BoolValue(provScheme.GetUseFullDiskCloneProvisioning())
}

func parseOnPremImagePath(catalog citrixorchestration.MachineCatalogDetailResponseModel) (masterImage, imageSnapshot string, resourcePoolPath string) {
	provScheme := catalog.GetProvisioningScheme()
	currentDiskImage := provScheme.GetCurrentDiskImage()
	currentImage := currentDiskImage.GetImage()
	relativePath := currentImage.GetRelativePath()

	// Refresh Master Image
	/*
			* For On-Premise snapshot image, the RelativePath looks like:
			* {VM name}.vm/{VM snapshot name}.snapshot(/{VM snapshot name}.snapshot)*
			* A new snapshot will be created if it was not specified. There will always be at least one snapshot in the path.

			* For Vsphere image, the RelativePath can also include resource pool and look like:
		    * {Resource Pool Name}.resourcepool/{Resource Pool Name}.resourcepool/{VM name}.vm/{VM snapshot name}.snapshot(/{VM snapshot name}.snapshot)*
	*/

	// Find the last index of ".resourcepool"
	lastResourcePoolIndex := strings.LastIndex(relativePath, ".resourcepool")
	if lastResourcePoolIndex != -1 {
		// Extract the resource pool path
		resourcePoolVal := relativePath[:lastResourcePoolIndex+len(".resourcepool")]
		// Remove all occurrences of ".resourcepool" from the extracted path
		resourcePoolPath = strings.ReplaceAll(resourcePoolVal, ".resourcepool", "")
		// Trim the resource pool path from the relative path
		relativePath = relativePath[lastResourcePoolIndex+len(".resourcepool/"):]
	}

	// Find the index of ".vm"
	vmIndex := strings.Index(relativePath, ".vm")
	if vmIndex == -1 {
		return "", "", ""
	}
	// Extract the master image name and trim the ".vm"
	masterImage = relativePath[:vmIndex]

	// Extract the snapshot part of the path
	snapshotPath := relativePath[vmIndex+len(".vm/"):]
	imageSnapshot = strings.ReplaceAll(snapshotPath, ".snapshot", "")

	return masterImage, imageSnapshot, resourcePoolPath
}

type PreparedImageConfigModel struct {
	ImageDefinition types.String `tfsdk:"image_definition"`
	ImageVersion    types.String `tfsdk:"image_version"`
}

func (PreparedImageConfigModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Specifying the prepared master image to be used for machine catalog.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"image_definition": schema.StringAttribute{
				Description: "ID of the image definition.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"image_version": schema.StringAttribute{
				Description: "ID of the image version.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = !req.StateValue.IsUnknown() && !req.StateValue.IsNull() && req.PlanValue.IsNull()
				},
				"Changing machine catalog image type requires replacing the machine catalog resource.",
				"Changing machine catalog image type requires replacing the machine catalog resource.",
			),
		},
	}
}

func (PreparedImageConfigModel) GetAttributes() map[string]schema.Attribute {
	return PreparedImageConfigModel{}.GetSchema().Attributes
}
