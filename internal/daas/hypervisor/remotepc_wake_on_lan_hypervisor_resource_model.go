// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type RemotePCWakeOnLANHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Zone     types.String `tfsdk:"zone"`
	Scopes   types.Set    `tfsdk:"scopes"`   // Set[string]
	Metadata types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	/**** RemotePC Connection Details ****/
	MaxAbsoluteActiveActions            types.Int64 `tfsdk:"max_absolute_active_actions"`
	MaxAbsoluteNewActionsPerMinute      types.Int64 `tfsdk:"max_absolute_new_actions_per_minute"`
	MaxPowerActionsPercentageOfMachines types.Int64 `tfsdk:"max_power_actions_percentage_of_machines"`
}

func (RemotePCWakeOnLANHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a RemotePC Wake On LAN Connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the hypervisor.",
				Required:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the hypervisor is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the hypervisor to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"metadata": util.GetMetadataListSchema("Hypervisor"),
			"max_absolute_active_actions": schema.Int64Attribute{
				Description: "Maximum number of actions that can execute in parallel on the hypervisor. Default is 20.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_absolute_new_actions_per_minute": schema.Int64Attribute{
				Description: "Maximum number of actions that can be started on the hypervisor per-minute. Default is 20.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_power_actions_percentage_of_machines": schema.Int64Attribute{
				Description: "Maximum percentage of machines on the hypervisor which can have their power state changed simultaneously. Default is 40.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(40),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
	}
}

func (RemotePCWakeOnLANHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return RemotePCWakeOnLANHypervisorResourceModel{}.GetSchema().Attributes
}

func (RemotePCWakeOnLANHypervisorResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r RemotePCWakeOnLANHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) RemotePCWakeOnLANHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())

	scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	r.MaxAbsoluteActiveActions = types.Int64Value(int64(hypervisor.GetMaxAbsoluteActiveActions()))
	r.MaxAbsoluteNewActionsPerMinute = types.Int64Value(int64(hypervisor.GetMaxAbsoluteNewActionsPerMinute()))
	r.MaxPowerActionsPercentageOfMachines = types.Int64Value(int64(hypervisor.GetMaxPowerActionsPercentageOfMachines()))

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), hypervisor.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())

	return r
}
