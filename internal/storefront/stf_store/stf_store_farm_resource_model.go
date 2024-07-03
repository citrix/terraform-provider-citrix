// Copyright © 2024. Citrix Systems, Inc.

package stf_store

import (
	"context"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// STFStoreFarmResourceModel maps the resource schema data.
type STFStoreFarmResourceModel struct {
	StoreService               types.String `tfsdk:"store_virtual_path"`             // The virtual path of the StoreService.
	FarmName                   types.String `tfsdk:"farm_name"`                      // The name of the Farm.
	FarmType                   types.String `tfsdk:"farm_type"`                      // The type of the Farm.
	Servers                    types.List   `tfsdk:"servers"`                        // List[string] The list of servers in the Farm.
	Port                       types.Int64  `tfsdk:"port"`                           // Service communication port.
	SSLRelayPort               types.Int64  `tfsdk:"ssl_relay_port"`                 // The SSL Relay port
	TransportType              types.Int64  `tfsdk:"transport_type"`                 // Type of transport to use. Http, Https, SSL for example
	LoadBalance                types.Bool   `tfsdk:"load_balance"`                   // Round robin load balance the xml service servers.
	XMLValidationEnabled       types.Bool   `tfsdk:"xml_validation_enabled"`         // Enable XML service endpoint validation
	XMLValidationSecret        types.String `tfsdk:"xml_validation_secret"`          // XML service endpoint validation shared secret
	ServiceUrls                types.List   `tfsdk:"server_urls"`                    // List[string] The url to the service location used to provide web and SaaS apps via this farm.
	AllFailedBypassDuration    types.Int64  `tfsdk:"all_failed_bypass_duration"`     // Period of time to skip all xml service requests should all servers fail to respond.
	BypassDuration             types.Int64  `tfsdk:"bypass_duration"`                // Period of time to skip a server when is fails to respond.
	TicketTimeToLive           types.Int64  `tfsdk:"ticket_time_to_live"`            // Period of time an ICA launch ticket is valid once requested on pre 7.0 XenApp and XenDesktop farms.
	RadeTicketTimeToLive       types.Int64  `tfsdk:"rade_ticket_time_to_live"`       // Period of time a RADE launch ticket is valid once requested on pre 7.0 XenApp and XenDesktop farms.
	MaxFailedServersPerRequest types.Int64  `tfsdk:"max_failed_servers_per_request"` // Maximum number of servers within a single farm that can fail before aborting a request.
	Zones                      types.List   `tfsdk:"zones"`                          // List[string] The list of Zone names associated with the farm.
	Product                    types.String `tfsdk:"product"`                        // Cloud deployments only otherwise ignored. The product name of the farm configured.
	RestrictPoPs               types.String `tfsdk:"restrict_pops"`                  // Cloud deployments only otherwise ignored. Restricts GWaaS traffic to the specified POP.
	FarmGuid                   types.String `tfsdk:"farm_guid"`                      // Cloud deployments only otherwise ignored. A tag indicating the scope of the farm.
}

func (*stfStoreFarmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "StoreFront Store Farm Config.",
		Attributes: map[string]schema.Attribute{
			"store_virtual_path": schema.StringAttribute{
				Description: "The Virtual Path of the StoreFront Store Service linked to the Farm.",
				Required:    true,
			},
			"farm_name": schema.StringAttribute{
				Description: "The name of the Farm.",
				Required:    true,
			},
			"farm_type": schema.StringAttribute{
				Description: "The type of the Farm. Can be XenApp, XenDesktop, AppController, VDIinaBox, Store or SPA.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"XenApp", "XenDesktop", "AppController", "VDIinaBox", "Store", "SPA",
					),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Service communication port. Default is 443",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(443),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"ssl_relay_port": schema.Int64Attribute{
				Description: "The SSL Relay port. Default is 443",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(443),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"transport_type": schema.Int64Attribute{
				Description: "Type of transport to use. Http, Https, SSL for example. ",
				Optional:    true,
				Computed:    true,
			},
			"load_balance": schema.BoolAttribute{
				Description: "Round robin load balance the xml service servers. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"xml_validation_enabled": schema.BoolAttribute{
				Description: "Enable XML service endpoint validation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"xml_validation_secret": schema.StringAttribute{
				Description: "XML service endpoint validation shared secret.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"all_failed_bypass_duration": schema.Int64Attribute{
				Description: "Period of time to skip all xml service requests should all servers fail to respond.",
				Optional:    true,
				Computed:    true,
			},
			"bypass_duration": schema.Int64Attribute{
				Description: "Period of time to skip a server when is fails to respond. Defaults to 60.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"ticket_time_to_live": schema.Int64Attribute{
				Description: "Period of time an ICA launch ticket is valid once requested on pre 7.0 XenApp and XenDesktop farms. Defaults to 0",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(200),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"rade_ticket_time_to_live": schema.Int64Attribute{
				Description: "Period of time a RADE launch ticket is valid once requested on pre 7.0 XenApp and XenDesktop farms. Defaults to 100.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_failed_servers_per_request": schema.Int64Attribute{
				Description: "Maximum number of servers within a single farm that can fail before aborting a request.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"product": schema.StringAttribute{
				Description: "Cloud deployments only otherwise ignored. The product name of the farm configured.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"restrict_pops": schema.StringAttribute{
				Description: "Cloud deployments only otherwise ignored. Restricts GWaaS traffic to the specified POP.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"farm_guid": schema.StringAttribute{
				Description: "A tag indicating the scope of the farm. Valid for cloud deployments only.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"zones": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of Zone names associated with the farm.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"server_urls": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The url to the service location used to provide web and SaaS apps via this farm.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"servers": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The list of servers in the Farm.",
				Required:    true,
			},
		},
	}
}

func (r *STFStoreFarmResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, farm citrixstorefront.StoreFarmModel) {
	if farm.FarmName.IsSet() {
		r.FarmName = types.StringValue(*farm.FarmName.Get())
	}
	if farm.FarmType.IsSet() {
		r.FarmType = types.StringValue(FarmTypeFromInt(*farm.FarmType.Get()))
	}
	if farm.Port.IsSet() {
		r.Port = types.Int64Value(*farm.Port.Get())
	}
	if farm.SSLRelayPort.IsSet() {
		r.SSLRelayPort = types.Int64Value(*farm.SSLRelayPort.Get())
	}
	if farm.TransportType.IsSet() {
		r.TransportType = types.Int64Value(*farm.TransportType.Get())
	}
	if farm.LoadBalance.IsSet() {
		r.LoadBalance = types.BoolValue(*farm.LoadBalance.Get())
	}
	if farm.XMLValidationEnabled.IsSet() {
		r.XMLValidationEnabled = types.BoolValue(*farm.XMLValidationEnabled.Get())
	}
	if farm.XMLValidationSecret.IsSet() {
		r.XMLValidationSecret = types.StringValue(*farm.XMLValidationSecret.Get())
	}
	if farm.AllFailedBypassDuration.IsSet() {
		r.AllFailedBypassDuration = types.Int64Value(*farm.AllFailedBypassDuration.Get())
	}
	if farm.BypassDuration.IsSet() {
		r.BypassDuration = types.Int64Value(*farm.BypassDuration.Get())
	}
	if farm.TicketTimeToLive.IsSet() {
		r.TicketTimeToLive = types.Int64Value(*farm.TicketTimeToLive.Get())
	}
	if farm.RadeTicketTimeToLive.IsSet() {
		r.RadeTicketTimeToLive = types.Int64Value(*farm.RadeTicketTimeToLive.Get())
	}
	if farm.MaxFailedServersPerRequest.IsSet() {
		r.MaxFailedServersPerRequest = types.Int64Value(*farm.MaxFailedServersPerRequest.Get())
	}
	if farm.Product.IsSet() {
		r.Product = types.StringValue(*farm.Product.Get())
	}
	if farm.RestrictPoPs.IsSet() {
		r.RestrictPoPs = types.StringValue(*farm.RestrictPoPs.Get())
	}
	if farm.FarmGuid.IsSet() {
		r.FarmGuid = types.StringValue(*farm.FarmGuid.Get())
	}

	r.Zones = util.RefreshListValues(ctx, diagnostics, r.Zones, farm.Zones)
	r.Servers = util.RefreshListValues(ctx, diagnostics, r.Servers, farm.Servers)
	r.ServiceUrls = util.RefreshListValues(ctx, diagnostics, r.ServiceUrls, farm.ServiceUrls)
}

func FarmTypeFromInt(farmTypeInt int64) string {
	switch farmTypeInt {
	case 0:
		return "XenApp"
	case 1:
		return "XenDesktop"
	case 2:
		return "AppController"
	case 3:
		return "VDIinaBox"
	case 4:
		return "Store"
	case 5:
		return "SPA"
	default:
		return "Unknown"
	}
}
