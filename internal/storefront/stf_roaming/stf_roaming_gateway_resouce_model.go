// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"
	"regexp"

	citrixstorefrontModels "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type STFSecureTicketAuthority struct {
	AuthorityId          types.String `tfsdk:"authority_id"`
	StaUrl               types.String `tfsdk:"sta_url"`
	StaValidationEnabled types.Bool   `tfsdk:"sta_validation_enabled"`
	StaValidationSecret  types.String `tfsdk:"sta_validation_secret"`
}

func (STFSecureTicketAuthority) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"authority_id": schema.StringAttribute{
				Description: "The ID of the Secure Ticket Authority (STA) server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sta_url": schema.StringAttribute{
				Description: "The URL of the Secure Ticket Authority (STA) server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^.*\/scripts\/ctxsta\.dll$`), "must be a valid URL end with `/scripts/ctxsta.dll`."),
				},
			},
			"sta_validation_enabled": schema.BoolAttribute{
				Description: "Whether Secure Ticket Authority (STA) validation is enabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Default: booldefault.StaticBool(false),
			},
			"sta_validation_secret": schema.StringAttribute{
				Description: "The Secure Ticket Authority (STA) validation secret.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (STFSecureTicketAuthority) GetAttributes() map[string]schema.Attribute {
	return STFSecureTicketAuthority{}.GetSchema().Attributes
}

type STFRoamingGatewayResourceModel struct {
	SiteId                     types.String `tfsdk:"site_id"`
	Name                       types.String `tfsdk:"name"`
	LogonType                  types.String `tfsdk:"logon_type"`
	SmartCardFallbackLogonType types.String `tfsdk:"smart_card_fallback_logon_type"`
	GatewayUrl                 types.String `tfsdk:"gateway_url"`
	CallbackUrl                types.String `tfsdk:"callback_url"`
	Edition                    types.String `tfsdk:"edition"`
	Deployment                 types.String `tfsdk:"deployment"`
	Version                    types.String `tfsdk:"version"`
	SessionReliability         types.Bool   `tfsdk:"session_reliability"`
	RequestTicketTwoSTAs       types.Bool   `tfsdk:"request_ticket_two_stas"`
	SubnetIPAddress            types.String `tfsdk:"subnet_ip_address"`
	SecureTicketAuthorityUrls  types.List   `tfsdk:"secure_ticket_authority_urls"` // List[STFSecureTicketAuthority]
	StasUseLoadBalancing       types.Bool   `tfsdk:"stas_use_load_balancing"`
	StasBypassDuration         types.String `tfsdk:"stas_bypass_duration"`
	GslbUrl                    types.String `tfsdk:"gslb_url"`
	IsCloudGateway             types.Bool   `tfsdk:"is_cloud_gateway"`
}

func (r *STFRoamingGatewayResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, roamingGateway *citrixstorefrontModels.STFRoamingGatewayResponseModel) {
	r.SiteId = types.StringValue(roamingGateway.SiteId)
	if roamingGateway.Name.IsSet() {
		r.Name = types.StringValue(*roamingGateway.Name.Get())
	}

	if roamingGateway.LogonType.IsSet() {
		r.LogonType = types.StringValue(*roamingGateway.LogonType.Get())
	}

	if roamingGateway.SmartCardFallbackLogonType.IsSet() {
		r.SmartCardFallbackLogonType = types.StringValue(*roamingGateway.SmartCardFallbackLogonType.Get())
	} else {
		r.SmartCardFallbackLogonType = types.StringValue("None")
	}

	if roamingGateway.GatewayUrl.IsSet() {
		r.GatewayUrl = types.StringValue(*roamingGateway.GatewayUrl.Get())
	}

	if roamingGateway.CallbackUrl.IsSet() {
		r.CallbackUrl = types.StringValue(*roamingGateway.CallbackUrl.Get())
	}

	if roamingGateway.Edition.IsSet() {
		r.Edition = types.StringValue(*roamingGateway.Edition.Get())
	}

	if roamingGateway.Deployment.IsSet() {
		r.Deployment = types.StringValue(*roamingGateway.Deployment.Get())
	}

	if roamingGateway.Version.IsSet() {
		r.Version = types.StringValue(*roamingGateway.Version.Get())
	}

	if roamingGateway.SessionReliability.IsSet() {
		r.SessionReliability = types.BoolValue(*roamingGateway.SessionReliability.Get())
	} else {
		r.SessionReliability = types.BoolValue(false)
	}

	if roamingGateway.RequestTicketTwoSTAs.IsSet() {
		r.RequestTicketTwoSTAs = types.BoolValue(*roamingGateway.RequestTicketTwoSTAs.Get())
	} else {
		r.RequestTicketTwoSTAs = types.BoolValue(false)
	}

	if roamingGateway.SubnetIPAddress.IsSet() {
		r.SubnetIPAddress = types.StringValue(*roamingGateway.SubnetIPAddress.Get())
	}

	if len(roamingGateway.SecureTicketAuthorityUrls) > 0 {
		authUrls := []STFSecureTicketAuthority{}
		for _, staUrl := range roamingGateway.SecureTicketAuthorityUrls {
			sta := STFSecureTicketAuthority{}
			if staUrl.AuthorityId.IsSet() {
				sta.AuthorityId = types.StringValue(*staUrl.AuthorityId.Get())
			}
			if staUrl.StaUrl.IsSet() {
				sta.StaUrl = types.StringValue(*staUrl.StaUrl.Get())
			}
			if staUrl.StaValidationEnabled.IsSet() {
				sta.StaValidationEnabled = types.BoolValue(*staUrl.StaValidationEnabled.Get())
			}
			if !sta.StaValidationEnabled.ValueBool() {
				sta.StaValidationSecret = types.StringNull()
			} else if staUrl.StaValidationSecret.IsSet() && *staUrl.StaValidationSecret.Get() != "" {
				sta.StaValidationSecret = types.StringValue(*staUrl.StaValidationSecret.Get())
			}
			authUrls = append(authUrls, sta)
		}

		staUrlSet := util.TypedArrayToObjectList[STFSecureTicketAuthority](ctx, diagnostics, authUrls)
		r.SecureTicketAuthorityUrls = staUrlSet
	}

	if roamingGateway.StasUseLoadBalancing.IsSet() {
		r.StasUseLoadBalancing = types.BoolValue(*roamingGateway.StasUseLoadBalancing.Get())
	} else {
		r.StasUseLoadBalancing = types.BoolValue(false)
	}

	if roamingGateway.StasBypassDuration.IsSet() {
		r.StasBypassDuration = types.StringValue(*roamingGateway.StasBypassDuration.Get())
	} else {
		r.StasBypassDuration = types.StringValue("0.1:0:0")
	}

	if roamingGateway.GslbUrl.IsSet() {
		r.GslbUrl = types.StringValue(*roamingGateway.GslbUrl.Get())
	} else {
		r.GslbUrl = types.StringValue("")
	}

	if roamingGateway.IsCloudGateway.IsSet() {
		r.IsCloudGateway = types.BoolValue(*roamingGateway.IsCloudGateway.Get())
	} else {
		r.IsCloudGateway = types.BoolValue(false)
	}
}

func (STFRoamingGatewayResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- Manage a StoreFront roaming gateway in the global gateway list. Gateways for remote access and authentication are added to Stores from the globally managed list.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront roaming gateway.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the StoreFront roaming gateway.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"logon_type": schema.StringAttribute{
				Description: "The login type required and supported by the Gateway. Possible values are `UsedForHDXOnly`, `Domain`, `RSA`, `DomainAndRSA`, `SMS`, `GatewayKnows`, `SmartCard`, and `None`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("UsedForHDXOnly", "Domain", "RSA", "DomainAndRSA", "SMS", "GatewayKnows", "SmartCard", "None"),
				},
			},
			"gateway_url": schema.StringAttribute{
				Description: "The URL of the StoreFront gateway.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"callback_url": schema.StringAttribute{
				Description: "The Gateway authentication NetScaler call-back Url.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"edition": schema.StringAttribute{
				Description: "The Citrix NetScaler Gateway edition.",
				Computed:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "The deployment type of the StoreFront gateway.",
				Computed:    true,
			},
			"smart_card_fallback_logon_type": schema.StringAttribute{
				Description: "The login type to use when SmartCard fails. Possible values are `UsedForHDXOnly`, `Domain`, `RSA`, `DomainAndRSA`, `SMS`, `GatewayKnows`, `SmartCard`, and `None`. Defaults to `None`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("None"),
				Validators: []validator.String{
					stringvalidator.OneOf("UsedForHDXOnly", "Domain", "RSA", "DomainAndRSA", "SMS", "GatewayKnows", "SmartCard", "None"),
				},
			},
			"version": schema.StringAttribute{
				Description: "The Citrix NetScaler Gateway version. Possible values are `Version10_0_69_4` and `Version9x`. Defaults to `Version10_0_69_4`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Version10_0_69_4"),
				Validators: []validator.String{
					stringvalidator.OneOf("Version10_0_69_4", "Version9x"),
				},
			},
			"session_reliability": schema.BoolAttribute{
				Description: "Enable session reliability. Session Reliability keeps sessions active and on the user’s screen when network connectivity is interrupted. Users continue to see the application they are using until network connectivity resumes. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"request_ticket_two_stas": schema.BoolAttribute{
				Description: "Request STA tickets from two STA servers (Requires two STA servers). Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"subnet_ip_address": schema.StringAttribute{
				Description: "The subnet IP address of the StoreFront gateway.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(8),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`),
						"must be in the format xx.xxx.xxx.xxx where each segment is 1 to 3 digits",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secure_ticket_authority_urls": schema.ListNestedAttribute{
				NestedObject: STFSecureTicketAuthority{}.GetSchema(),
				Description:  "The Secure Ticket Authority (STA) URLs. The STA servers validate the tickets that are issued by the StoreFront server. The STA servers must be reachable from the StoreFront server.",
				Optional:     true,
			},
			"stas_use_load_balancing": schema.BoolAttribute{
				Description: "Use load balancing for the Secure Ticket Authority (STA) servers. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"stas_bypass_duration": schema.StringAttribute{
				Description: "Time before retrying a failed STA server in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.1:0:0`",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.1:0:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"gslb_url": schema.StringAttribute{
				Description: "An optional URL which corresponds to the GSLB domain used by multiple gateways. Defaults to an empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"is_cloud_gateway": schema.BoolAttribute{
				Description: "Whether the Gateway is an instance of Citrix Gateway Service in the cloud. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (STFRoamingGatewayResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFRoamingGatewayResourceModel{}.GetSchema().Attributes
}
