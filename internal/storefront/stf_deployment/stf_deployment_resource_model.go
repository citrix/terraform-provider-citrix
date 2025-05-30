// Copyright © 2024. Citrix Systems, Inc.

package stf_deployment

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoamingBeacon struct {
	Internal types.String `tfsdk:"internal_address"`
	External types.List   `tfsdk:"external_addresses"`
}

func (RoamingBeacon) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Roaming Beacon configuration.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"internal_address": schema.StringAttribute{
				Description: "Internal IP address of the beacon. It can either be the hostname or the IP address of the beacon. The Internal IP must be either in `http(s)://<ip_address>/` OR `http(s)://<hostname>/` format.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.UrlValidator), "must be a valid URI format starting with http:// or https:// and ending with /."),
				},
			},
			"external_addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "External IP addresses of the beacon. It can either be the gateway url or the IP addresses of the beacon. If the user removes it from terraform, then the previously persisted values will be retained. When omitted, StoreFront server will use default value of `http://ping.citrix.com` and the gateway url. Each External IP must be either in `http(s)://<ip_address>/` OR `http(s)://<hostname>/` format.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.UrlValidator), "must be a valid URI format starting with http:// or https:// and ending with /."),
					),
				},
			},
		},
	}
}

func (RoamingBeacon) GetAttributes() map[string]schema.Attribute {
	return RoamingBeacon{}.GetSchema().Attributes
}

type STFSecureTicketAuthority struct {
	StaUrl               types.String `tfsdk:"sta_url"`
	StaValidationEnabled types.Bool   `tfsdk:"sta_validation_enabled"`
	StaValidationSecret  types.String `tfsdk:"sta_validation_secret"`
}

func (r STFSecureTicketAuthority) GetKey() string {
	return r.StaUrl.ValueString()
}

func (r STFSecureTicketAuthority) RefreshListItem(_ context.Context, _ *diag.Diagnostics, sta citrixstorefront.STFSTAUrlModel) util.ResourceModelWithAttributes {
	r.StaUrl = types.StringValue(*sta.StaUrl.Get())
	r.StaValidationEnabled = types.BoolValue(*sta.StaValidationEnabled.Get())
	if !sta.StaValidationEnabled.IsSet() || !*sta.StaValidationEnabled.Get() {
		r.StaValidationSecret = types.StringNull()
	} else if sta.StaValidationSecret.IsSet() && *sta.StaValidationSecret.Get() != "" {
		r.StaValidationSecret = types.StringValue(*sta.StaValidationSecret.Get())
	}
	return r
}

func (STFSecureTicketAuthority) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"sta_url": schema.StringAttribute{
				Description: "The URL of the Secure Ticket Authority (STA) server.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^.*\/scripts\/ctxsta\.dll$`), "must be a valid URL end with `/scripts/ctxsta.dll`."),
				},
			},
			"sta_validation_enabled": schema.BoolAttribute{
				Description: "Whether Secure Ticket Authority (STA) validation is enabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"sta_validation_secret": schema.StringAttribute{
				Description: "The Secure Ticket Authority (STA) validation secret.",
				Optional:    true,
				Sensitive:   true,
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

type RoamingGateway struct {
	Name                       types.String `tfsdk:"name"`
	LogonType                  types.String `tfsdk:"logon_type"`
	SmartCardFallbackLogonType types.String `tfsdk:"smart_card_fallback_logon_type"`
	GatewayUrl                 types.String `tfsdk:"gateway_url"`
	CallbackUrl                types.String `tfsdk:"callback_url"`
	Version                    types.String `tfsdk:"version"`
	SessionReliability         types.Bool   `tfsdk:"session_reliability"`
	RequestTicketTwoSTAs       types.Bool   `tfsdk:"request_ticket_from_two_stas"`
	SubnetIPAddress            types.String `tfsdk:"subnet_ip_address"`
	SecureTicketAuthorityUrls  types.List   `tfsdk:"secure_ticket_authority_urls"` // List[STFSecureTicketAuthority]
	StasUseLoadBalancing       types.Bool   `tfsdk:"stas_use_load_balancing"`
	StasBypassDuration         types.String `tfsdk:"stas_bypass_duration"`
	GslbUrl                    types.String `tfsdk:"gslb_url"`
	IsCloudGateway             types.Bool   `tfsdk:"is_cloud_gateway"`
}

func (r RoamingGateway) GetKey() string {
	return r.Name.ValueString()
}

func (r RoamingGateway) RefreshListItem(ctx context.Context, diagnostics *diag.Diagnostics, roamingGateway citrixstorefront.STFRoamingGatewayResponseModel) util.ResourceModelWithAttributes {
	r.Name = types.StringValue(*roamingGateway.Name.Get())
	r.LogonType = types.StringValue(*roamingGateway.LogonType.Get())
	if roamingGateway.SmartCardFallbackLogonType.IsSet() {
		r.SmartCardFallbackLogonType = types.StringValue(*roamingGateway.SmartCardFallbackLogonType.Get())
	}
	r.GatewayUrl = types.StringValue(*roamingGateway.GatewayUrl.Get())
	if roamingGateway.CallbackUrl.IsSet() {
		r.CallbackUrl = types.StringValue(*roamingGateway.CallbackUrl.Get())
	}
	if roamingGateway.Version.IsSet() {
		r.Version = types.StringValue(*roamingGateway.Version.Get())
	}
	if roamingGateway.SessionReliability.IsSet() {
		r.SessionReliability = types.BoolValue(*roamingGateway.SessionReliability.Get())
	}
	if roamingGateway.RequestTicketTwoSTAs.IsSet() {
		r.RequestTicketTwoSTAs = types.BoolValue(*roamingGateway.RequestTicketTwoSTAs.Get())
	}
	if roamingGateway.SubnetIPAddress.IsSet() && *roamingGateway.SubnetIPAddress.Get() != "" {
		r.SubnetIPAddress = types.StringValue(*roamingGateway.SubnetIPAddress.Get())
	}
	if roamingGateway.StasUseLoadBalancing.IsSet() {
		r.StasUseLoadBalancing = types.BoolValue(*roamingGateway.StasUseLoadBalancing.Get())
	}
	if roamingGateway.StasBypassDuration.IsSet() {
		r.StasBypassDuration = types.StringValue(*roamingGateway.StasBypassDuration.Get())
	}
	if roamingGateway.GslbUrl.IsSet() {
		r.GslbUrl = types.StringValue(*roamingGateway.GslbUrl.Get())
	}
	if roamingGateway.IsCloudGateway.IsSet() {
		r.IsCloudGateway = types.BoolValue(*roamingGateway.IsCloudGateway.Get())
	}
	if len(roamingGateway.SecureTicketAuthorityUrls) > 0 {
		r.SecureTicketAuthorityUrls = util.RefreshListValueProperties[STFSecureTicketAuthority, citrixstorefront.STFSTAUrlModel](ctx, diagnostics, r.SecureTicketAuthorityUrls, roamingGateway.SecureTicketAuthorityUrls, util.GetSTFSTAUrlKey)
	} else {
		r.SecureTicketAuthorityUrls = util.TypedArrayToObjectList[STFSecureTicketAuthority](ctx, diagnostics, nil)
	}
	return r
}

func (RoamingGateway) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the StoreFront roaming gateway.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"logon_type": schema.StringAttribute{
				Description: "The login type required and supported by the Gateway. Possible values are `UsedForHDXOnly`, `Domain`, `RSA`, `DomainAndRSA`, `SMS`, `GatewayKnows`, `SmartCard`, and `None`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixstorefront.LOGONTYPE_USED_FOR_HDX_ONLY),
						string(citrixstorefront.LOGONTYPE_DOMAIN),
						string(citrixstorefront.LOGONTYPE_RSA),
						string(citrixstorefront.LOGONTYPE_DOMAIN_AND_RSA),
						string(citrixstorefront.LOGONTYPE_SMS),
						string(citrixstorefront.LOGONTYPE_GATEWAY_KNOWS),
						string(citrixstorefront.LOGONTYPE_SMART_CARD),
						string(citrixstorefront.LOGONTYPE_NONE),
					),
				},
			},
			"gateway_url": schema.StringAttribute{
				Description: "The URL of the StoreFront gateway.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.UrlValidator), "must be a valid URI format starting with http:// or https:// and ending with /."),
				},
			},
			"callback_url": schema.StringAttribute{
				Description: "The Gateway authentication NetScaler call-back url. Must end with `/CitrixAuthService/AuthService.asmx`",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^.*\/CitrixAuthService\/AuthService.asmx$`), "must be a valid URL end with `/CitrixAuthService/AuthService.asmx`."),
				},
			},
			"smart_card_fallback_logon_type": schema.StringAttribute{
				Description: "The login type to use when SmartCard fails. Possible values are `UsedForHDXOnly`, `Domain`, `RSA`, `DomainAndRSA`, `SMS`, `GatewayKnows`, `SmartCard`, and `None`. Defaults to `None`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("None"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixstorefront.LOGONTYPE_USED_FOR_HDX_ONLY),
						string(citrixstorefront.LOGONTYPE_DOMAIN),
						string(citrixstorefront.LOGONTYPE_RSA),
						string(citrixstorefront.LOGONTYPE_DOMAIN_AND_RSA),
						string(citrixstorefront.LOGONTYPE_SMS),
						string(citrixstorefront.LOGONTYPE_GATEWAY_KNOWS),
						string(citrixstorefront.LOGONTYPE_SMART_CARD),
						string(citrixstorefront.LOGONTYPE_NONE),
					),
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
			"request_ticket_from_two_stas": schema.BoolAttribute{
				Description: "Request STA tickets from two STA servers (Requires two STA servers). Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"subnet_ip_address": schema.StringAttribute{
				Description: "The subnet IP address of the StoreFront gateway.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(7),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`),
						"must be in the format xxx.xxx.xxx.xxx where each segment is 1 to 3 digits",
					),
				},
			},
			"secure_ticket_authority_urls": schema.ListNestedAttribute{
				NestedObject: STFSecureTicketAuthority{}.GetSchema(),
				Description:  "The Secure Ticket Authority (STA) URLs. The STA servers validate the tickets that are issued by the StoreFront server. The STA servers must be reachable from the StoreFront server.",
				Optional:     true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
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
				Description: "An optional URL which corresponds to the Global Server Load Balancing domain used by multiple gateways.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.UrlValidator), "must be a valid URI format starting with http:// or https://, and ending with /."),
				},
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

func (RoamingGateway) GetAttributes() map[string]schema.Attribute {
	return RoamingGateway{}.GetSchema().Attributes
}

// SFDeploymentResourceModel maps the resource schema data.
type STFDeploymentResourceModel struct {
	SiteId         types.String `tfsdk:"site_id"`
	HostBaseUrl    types.String `tfsdk:"host_base_url"`
	RoamingGateway types.List   `tfsdk:"roaming_gateway"` // List[RoamingGateway]
	RoamingBeacon  types.Object `tfsdk:"roaming_beacon"`  // RoamingBeacon
}

func (r *STFDeploymentResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deployment *citrixstorefront.STFDeploymentDetailModel, roamingGateway []citrixstorefront.STFRoamingGatewayResponseModel, roamInt *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel, roamExt *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel) {
	// Overwrite SFDeploymentResourceModel with refreshed state
	r.SiteId = types.StringValue(strconv.Itoa(int(*deployment.SiteId.Get())))
	r.HostBaseUrl = types.StringValue(strings.TrimRight(*deployment.HostBaseUrl.Get(), "/"))

	// Roaming Gateway
	if len(roamingGateway) == 0 {
		r.RoamingGateway = util.TypedArrayToObjectList[RoamingGateway](ctx, diagnostics, nil)
	} else {
		r.RoamingGateway = util.RefreshListValueProperties[RoamingGateway, citrixstorefront.STFRoamingGatewayResponseModel](ctx, diagnostics, r.RoamingGateway, roamingGateway, util.GetSTFRoamingGatewayKey)
		if roamInt != nil && roamInt.Internal != "" {
			r.RefreshRoamingBeacon(ctx, diagnostics, roamInt, roamExt)
		}
	}
}

func (STFDeploymentResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront Deployment.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront deployment. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_base_url": schema.StringAttribute{
				Description: "Url used to access the StoreFront server group.",
				Required:    true,
			},
			"roaming_gateway": schema.ListNestedAttribute{
				Description:  "Roaming Gateway configuration.",
				Optional:     true,
				NestedObject: RoamingGateway{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"roaming_beacon": RoamingBeacon{}.GetSchema(),
		},
	}
}

func (STFDeploymentResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFDeploymentResourceModel{}.GetSchema().Attributes
}

func (STFDeploymentResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r *STFDeploymentResourceModel) RefreshRoamingBeacon(ctx context.Context, diagnostics *diag.Diagnostics, roamInt *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel, roamExt *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel) {
	refreshedRoamingBeacon := util.ObjectValueToTypedObject[RoamingBeacon](ctx, diagnostics, r.RoamingBeacon)
	if roamInt.Internal[len(roamInt.Internal)-1:] != "/" {
		roamInt.Internal += "/"
	}
	refreshedRoamingBeacon.Internal = types.StringValue(roamInt.Internal)
	if roamExt != nil && roamExt.External[0] != "" {
		for ext := 0; ext < len(roamExt.External); ext++ {
			if roamExt.External[ext][len(roamExt.External[ext])-1:] != "/" {
				roamExt.External[ext] += "/"
			}
		}
		refreshedRoamingBeacon.External = util.RefreshListValues(ctx, diagnostics, refreshedRoamingBeacon.External, roamExt.External)
	}

	refreshedRoamingBeaconObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedRoamingBeacon)
	r.RoamingBeacon = refreshedRoamingBeaconObject
}
