// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	Metadata   types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	VmTagging  types.Bool   `tfsdk:"vm_tagging"`
	/**** Resource Pool Details ****/
	Region         types.String `tfsdk:"region"`
	VirtualNetwork types.String `tfsdk:"virtual_network"`
	Subnets        types.List   `tfsdk:"subnets"` // List[string]
	/** Azure Resource Pool **/
	VirtualNetworkResourceGroup types.String `tfsdk:"virtual_network_resource_group"`
}

func (AzureHypervisorResourcePoolResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an Azure hypervisor resource pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the resource pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource pool. Name should be unique across all hypervisors.",
				Required:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor for which the resource pool needs to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"virtual_network_resource_group": schema.StringAttribute{
				Description: "The name of the resource group where the vnet resides.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vm_tagging": schema.BoolAttribute{
				Description: "Indicates whether VMs created by provisioning operations should be tagged. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"virtual_network": schema.StringAttribute{
				Description: "Name of the cloud virtual network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Subnets to allocate VDAs within the virtual network.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Description: "Cloud Region where the virtual network sits in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !req.ConfigValue.IsNull() && !req.StateValue.IsNull() &&
								(location.Normalize(req.ConfigValue.ValueString()) != location.Normalize(req.StateValue.ValueString()))
						},
						"Force replacement when region changes, unless changing between Azure region name (East US) and Id (eastus)",
						"Force replacement when region changes, unless changing between Azure region name (East US) and Id (eastus)",
					),
				},
			},
			"metadata": util.GetMetadataListSchema("Hypervisor Resource Pool"),
		},
	}
}

func (AzureHypervisorResourcePoolResourceModel) GetAttributes() map[string]schema.Attribute {
	return AzureHypervisorResourcePoolResourceModel{}.GetSchema().Attributes
}

func (AzureHypervisorResourcePoolResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r AzureHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) AzureHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	region := resourcePool.GetRegion()
	if r.shouldSetRegion(region) {
		r.Region = types.StringValue(region.GetName())
	}
	virtualNetwork := resourcePool.GetVirtualNetwork()
	resourceGroupName := getResourceGroupNameFromVnetId(virtualNetwork.GetId())
	r.VirtualNetworkResourceGroup = types.StringValue(resourceGroupName)
	r.VirtualNetwork = types.StringValue(virtualNetwork.GetName())
	r.VmTagging = types.BoolValue(resourcePool.GetVMTaggingEnabled())
	var res []string
	for _, model := range resourcePool.GetSubnets() {
		res = append(res, model.GetName())
	}
	r.Subnets = util.RefreshListValues(ctx, diagnostics, r.Subnets, res)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), resourcePool.GetMetadata())
	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	return r
}

func (r AzureHypervisorResourcePoolResourceModel) shouldSetRegion(region citrixorchestration.HypervisorResourceRefResponseModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.IsNull() || r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}

func getResourceGroupNameFromVnetId(vnetId string) string {
	resourceGroupAndVnetName := strings.Split(vnetId, "/")
	return resourceGroupAndVnetName[0]
}
