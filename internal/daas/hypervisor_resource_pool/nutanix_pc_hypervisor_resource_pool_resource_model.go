// Copyright © 2026. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NutanixPCHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	Metadata   types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Timeout    types.Object `tfsdk:"timeout"`
	VmTagging  types.Bool   `tfsdk:"vm_tagging"`
}

func (NutanixPCHypervisorResourcePoolResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a Nutanix AHV Prism Central hypervisor resource pool. The cluster and networks are selected per machine catalog.",
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
				Description: "Id of the Nutanix AHV Prism Central hypervisor for which the resource pool needs to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"timeout":  ResourcePoolTimeout{}.GetSchema(),
			"metadata": util.GetMetadataListSchema("Hypervisor Resource Pool"),
			"vm_tagging": schema.BoolAttribute{
				Description: "Indicates whether VMs created by provisioning operations should be tagged. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (NutanixPCHypervisorResourcePoolResourceModel) GetAttributes() map[string]schema.Attribute {
	return NutanixPCHypervisorResourcePoolResourceModel{}.GetSchema().Attributes
}

func (NutanixPCHypervisorResourcePoolResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r NutanixPCHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) NutanixPCHypervisorResourcePoolResourceModel {
	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())
	r.VmTagging = types.BoolValue(resourcePool.GetVMTaggingEnabled())

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), resourcePool.GetMetadata())
	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	return r
}
