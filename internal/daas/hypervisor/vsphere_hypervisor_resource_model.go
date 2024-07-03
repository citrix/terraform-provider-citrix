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
type VsphereHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Zone   types.String `tfsdk:"zone"`
	Scopes types.Set    `tfsdk:"scopes"` // Set[string]
	/** vSphere Connection **/
	Username                            types.String `tfsdk:"username"`
	Password                            types.String `tfsdk:"password"`
	PasswordFormat                      types.String `tfsdk:"password_format"`
	Addresses                           types.List   `tfsdk:"addresses"`       // List[string]
	SslThumbprints                      types.List   `tfsdk:"ssl_thumbprints"` // List[string]
	MaxAbsoluteActiveActions            types.Int64  `tfsdk:"max_absolute_active_actions"`
	MaxAbsoluteNewActionsPerMinute      types.Int64  `tfsdk:"max_absolute_new_actions_per_minute"`
	MaxPowerActionsPercentageOfMachines types.Int64  `tfsdk:"max_power_actions_percentage_of_machines"`
}

func (VsphereHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a VMware vSphere hypervisor.",
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
			"username": schema.StringAttribute{
				Description: "Username of the hypervisor.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password of the hypervisor.",
				Required:    true,
				Sensitive:   true,
			},
			"password_format": schema.StringAttribute{
				Description: "Password format of the hypervisor. Choose between Base64 and PlainText.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_BASE64),
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_PLAIN_TEXT),
					),
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
			"ssl_thumbprints": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "SSL certificate thumbprints to consider acceptable for this connection.  If not specified, and the hypervisor uses SSL for its connection, the SSL certificate's root certification authority and any intermediate certificates must be trusted.",
				Optional:    true,
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
				Description: "Maximum number of actions that can execute in parallel on the hypervisor. Default is 40.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(40),
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
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
		},
	}
}

func (VsphereHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return VsphereHypervisorResourceModel{}.GetSchema().Attributes
}

func (r VsphereHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) VsphereHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	r.Username = types.StringValue(hypervisor.GetUserName())
	r.Addresses = util.RefreshListValues(ctx, diagnostics, r.Addresses, hypervisor.GetAddresses())
	r.MaxAbsoluteActiveActions = types.Int64Value(int64(hypervisor.GetMaxAbsoluteActiveActions()))
	r.MaxAbsoluteNewActionsPerMinute = types.Int64Value(int64(hypervisor.GetMaxAbsoluteNewActionsPerMinute()))
	r.MaxPowerActionsPercentageOfMachines = types.Int64Value(int64(hypervisor.GetMaxPowerActionsPercentageOfMachines()))

	sslThumbprints := util.RefreshListValues(ctx, diagnostics, r.SslThumbprints, hypervisor.GetSslThumbprints())
	if !sslThumbprints.IsNull() {
		r.SslThumbprints = sslThumbprints
	} else {
		r.SslThumbprints = types.ListNull(types.StringType)
	}

	scopeIds := util.GetIdsForScopeObjects(hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	return r
}
