// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type OpenShiftHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Zone     types.String `tfsdk:"zone"`
	Scopes   types.Set    `tfsdk:"scopes"`   // Set[string]
	Metadata types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Tenants  types.Set    `tfsdk:"tenants"`  // Set[string]

	/** OpenShift Connection **/
	Addresses                           types.List   `tfsdk:"addresses"`
	ServiceAccountToken                 types.String `tfsdk:"service_account_token"`
	SslThumbprints                      types.List   `tfsdk:"ssl_thumbprints"`
	MaxAbsoluteActiveActions            types.Int64  `tfsdk:"max_absolute_active_actions"`
	MaxAbsoluteNewActionsPerMinute      types.Int64  `tfsdk:"max_absolute_new_actions_per_minute"`
	MaxPowerActionsPercentageOfMachines types.Int64  `tfsdk:"max_power_actions_percentage_of_machines"`
}

func (OpenShiftHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an OpenShift hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
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
			"addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Hypervisor address(es). At least one is required.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"service_account_token": schema.StringAttribute{
				Description: "The Application Secret of the service principal used to access the OpenShift APIs.",
				Required:    true,
				Sensitive:   true,
			},
			"ssl_thumbprints": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "SSL certificate thumbprints to consider acceptable for this connection. If not specified, and the hypervisor uses SSL for its connection, the SSL certificate's root certification authority and any intermediate certificates must be trusted.",
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.SslThumbprintRegex), "must be specified with SSL thumbprint without colons"),
					),
				},
			},
			"max_absolute_active_actions": schema.Int64Attribute{
				Description: "Maximum number of actions that can execute in parallel on the hypervisor. Default is 100.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_absolute_new_actions_per_minute": schema.Int64Attribute{
				Description: "Maximum number of actions that can be started on the hypervisor per-minute. Default is 10.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_power_actions_percentage_of_machines": schema.Int64Attribute{
				Description: "Maximum percentage of machines on the hypervisor which can have their power state changed simultaneously. Default is 20.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
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
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the hypervisor connection.",
				Computed:    true,
			},
		},
	}
}

func (OpenShiftHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return OpenShiftHypervisorResourceModel{}.GetSchema().Attributes
}

func (OpenShiftHypervisorResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r OpenShiftHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) OpenShiftHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	r.Addresses = util.RefreshListValues(ctx, diagnostics, r.Addresses, hypervisor.GetAddresses())
	r.MaxAbsoluteActiveActions = types.Int64Value(int64(hypervisor.GetMaxAbsoluteActiveActions()))
	r.MaxAbsoluteNewActionsPerMinute = types.Int64Value(int64(hypervisor.GetMaxAbsoluteNewActionsPerMinute()))
	r.MaxPowerActionsPercentageOfMachines = types.Int64Value(int64(hypervisor.GetMaxPowerActionsPercentageOfMachines()))
	sslThumbprints := util.RefreshListValues(ctx, diagnostics, r.SslThumbprints, hypervisor.GetSslThumbprints())
	r.SslThumbprints = sslThumbprints

	scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), hypervisor.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, hypervisor.GetTenants())

	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	return r
}
