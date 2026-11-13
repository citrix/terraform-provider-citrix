// Copyright © 2026. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"strconv"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NutanixPCMachineConfigModel struct {
	Cluster                  types.String `tfsdk:"cluster"`
	MasterImage              types.String `tfsdk:"master_image"`
	MasterImageVersion       types.String `tfsdk:"master_image_version"`
	MasterImageNote          types.String `tfsdk:"master_image_note"`
	PreparedImage            types.Object `tfsdk:"prepared_image"`
	ImageUpdateRebootOptions types.Object `tfsdk:"image_update_reboot_options"`
	MachineProfile           types.String `tfsdk:"machine_profile"`
	MachineProfileVersion    types.String `tfsdk:"machine_profile_version"`
	CpuCount                 types.Int64  `tfsdk:"cpu_count"`
	CoresPerCpuCount         types.Int64  `tfsdk:"cores_per_cpu_count"`
	MemoryMB                 types.Int64  `tfsdk:"memory_mb"`
}

func (NutanixPCMachineConfigModel) GetSchema() schema.SingleNestedAttribute {
	preparedImageSchema := PreparedImageConfigModel{}.GetSchema()
	preparedImageSchema.Description = "The prepared image (image definition and image version) to use for the catalog. The image version must belong to the catalog's resource pool, because Nutanix Prism Central image versions can't be shared across resource pools. Requires `machine_profile` and `machine_profile_version`."
	preparedImageSchema.Validators = append(preparedImageSchema.Validators,
		objectvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("machine_profile")),
	)

	return schema.SingleNestedAttribute{
		Description: "Machine Configuration for Nutanix AHV Prism Central MCS catalog. Requires `network_mapping`, whose networks are looked up under `cluster`.",
		Optional:    true,
		Validators: []validator.Object{
			objectvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("network_mapping")),
		},
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Name of the Prism Central-managed cluster to provision machines in. The cluster can't be changed after the catalog is created; changing it replaces the catalog.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"master_image": schema.StringAttribute{
				Description: "Name of the Prism Central template used as the master image. Exactly one of `master_image` and `prepared_image` must be specified.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("prepared_image")),
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("master_image_version")),
				},
			},
			"master_image_version": schema.StringAttribute{
				Description: "Name of the template version used as the master image. Changing it publishes the new image to the catalog.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("master_image")),
				},
			},
			"master_image_note": schema.StringAttribute{
				Description: "The note for the master image. Can't be set with `prepared_image`. Defaults to an empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("prepared_image")),
				},
			},
			"prepared_image":              preparedImageSchema,
			"image_update_reboot_options": ImageUpdateRebootOptionsModel{}.GetSchema(),
			"machine_profile": schema.StringAttribute{
				Description: "Name of the Prism Central template used as the machine profile. Changing the machine profile between templates updates the catalog; adding or removing it replaces the catalog. Required with `prepared_image`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("machine_profile_version")),
				},
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
			"machine_profile_version": schema.StringAttribute{
				Description: "Name of the template version used as the machine profile.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("machine_profile")),
				},
			},
			"cpu_count": schema.Int64Attribute{
				Description: "Number of CPU sockets of each machine.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"cores_per_cpu_count": schema.Int64Attribute{
				Description: "Number of cores per CPU socket of each machine.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"memory_mb": schema.Int64Attribute{
				Description: "Memory of each machine in MB.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
	}
}

func (NutanixPCMachineConfigModel) GetAttributes() map[string]schema.Attribute {
	return NutanixPCMachineConfigModel{}.GetSchema().Attributes
}

func (mc *NutanixPCMachineConfigModel) RefreshProperties(ctx context.Context, diagnostics *diag.Diagnostics, catalog citrixorchestration.MachineCatalogDetailResponseModel, clusterName string) {
	provScheme := catalog.GetProvisioningScheme()

	if provScheme.CurrentImageVersion != nil {
		currentImage := provScheme.GetCurrentImageVersion()
		imageVersion := currentImage.GetImageVersion()
		imageDefinition := imageVersion.GetImageDefinition()
		preparedImageConfig := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, mc.PreparedImage)
		preparedImageConfig.ImageDefinition = types.StringValue(imageDefinition.GetId())
		preparedImageConfig.ImageVersion = types.StringValue(imageVersion.GetId())
		mc.PreparedImage = util.TypedObjectToObjectValue(ctx, diagnostics, preparedImageConfig)
		mc.MasterImage = types.StringNull()
		mc.MasterImageVersion = types.StringNull()
		mc.MasterImageNote = types.StringValue("")
	} else {
		masterImage := provScheme.GetMasterImage()
		if template, version, ok := util.ParseNutanixPCTemplateVersionPath(masterImage.GetXDPath()); ok {
			mc.MasterImage = types.StringValue(template)
			mc.MasterImageVersion = types.StringValue(version)
		}
		currentDiskImage := provScheme.GetCurrentDiskImage()
		mc.MasterImageNote = types.StringValue(currentDiskImage.GetMasterImageNote())
		if attributesMap, err := util.ResourceAttributeMapFromObject(PreparedImageConfigModel{}); err == nil {
			mc.PreparedImage = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null PreparedImageConfigModel", err.Error())
		}
	}

	machineProfile := provScheme.GetMachineProfile()
	if template, version, ok := util.ParseNutanixPCTemplateVersionPath(machineProfile.GetXDPath()); ok {
		mc.MachineProfile = types.StringValue(template)
		mc.MachineProfileVersion = types.StringValue(version)
	} else {
		mc.MachineProfile = types.StringNull()
		mc.MachineProfileVersion = types.StringNull()
	}

	mc.CpuCount = types.Int64Value(int64(provScheme.GetCpuCount()))
	mc.MemoryMB = types.Int64Value(int64(provScheme.GetMemoryMB()))
	if cores, ok := util.GetNutanixPCCoresPerCpu(provScheme.GetCustomProperties()); ok {
		mc.CoresPerCpuCount = types.Int64Value(cores)
	}

	if clusterName != "" {
		mc.Cluster = types.StringValue(clusterName)
	}
}

// getMachineProfilePath resolves the machine profile template version, or returns "" when no machine profile is set.
func (mc NutanixPCMachineConfigModel) getMachineProfilePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, errorTitle string) (string, error) {
	if mc.MachineProfile.ValueString() == "" {
		return "", nil
	}
	machineProfilePath, httpResp, err := util.GetNutanixPCTemplateVersionPath(ctx, client, diagnostics, hypervisorName, poolName, mc.MachineProfile.ValueString(), mc.MachineProfileVersion.ValueString())
	if err != nil {
		diagnostics.AddError(
			errorTitle,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\nFailed to locate machine profile template %s version %s on Nutanix Prism Central, error: %s", mc.MachineProfile.ValueString(), mc.MachineProfileVersion.ValueString(), err.Error()),
		)
		return "", err
	}
	return machineProfilePath, nil
}

func nutanixPCMachineConfigMismatchMessage(hypervisor *citrixorchestration.HypervisorDetailResponseModel) string {
	return fmt.Sprintf("nutanix_pc_machine_config must be set if and only if the resource pool belongs to a Nutanix AHV Prism Central hypervisor. Hypervisor %s has connection type %s.", hypervisor.GetName(), hypervisor.GetConnectionType())
}

// ensureNutanixPCCoresPerCpu re-sends CPUCores when an edit let the Prism Central plugin replace it with the
// template's cores per socket, which it does whenever a Set-ProvScheme carries no custom property change.
func ensureNutanixPCCoresPerCpu(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, plan MachineCatalogResourceModel, catalogId string, timeout int32) error {
	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
	if provSchemeModel.NutanixPCMachineConfig.IsNull() {
		return nil
	}
	nutanixPCMachineConfig := util.ObjectValueToTypedObject[NutanixPCMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.NutanixPCMachineConfig)

	catalog, err := util.GetMachineCatalog(ctx, client, &resp.Diagnostics, catalogId, true)
	if err != nil {
		return err
	}
	provScheme := catalog.GetProvisioningScheme()
	if cores, ok := util.GetNutanixPCCoresPerCpu(provScheme.GetCustomProperties()); ok && cores == nutanixPCMachineConfig.CoresPerCpuCount.ValueInt64() {
		return nil
	}

	var body citrixorchestration.UpdateMachineCatalogRequestModel
	customProperties := []citrixorchestration.NameValueStringPairModel{}
	util.AppendNameValueStringPair(&customProperties, util.NutanixPCCpuCoresKey, strconv.FormatInt(nutanixPCMachineConfig.CoresPerCpuCount.ValueInt64(), 10))
	body.SetCustomProperties(customProperties)

	updateRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalog(ctx, catalogId).UpdateMachineCatalogRequestModel(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateRequest, client).Async(true).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalog.GetName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nFailed to restore cores_per_cpu_count: "+util.ReadClientError(err),
		)
		return err
	}
	return util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating Machine Catalog "+catalog.GetName(), &resp.Diagnostics, timeout)
}
