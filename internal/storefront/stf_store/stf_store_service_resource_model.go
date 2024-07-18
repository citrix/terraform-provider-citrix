// Copyright © 2024. Citrix Systems, Inc.

package stf_store

import (
	"context"
	"regexp"

	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoamingAccount struct {
	Published types.Bool `tfsdk:"published"` // Whether the roaming account is published
}

func (RoamingAccount) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Roaming account settings for the Store",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"published": schema.BoolAttribute{
				Description: "Whether the roaming account is published. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (RoamingAccount) GetAttributes() map[string]schema.Attribute {
	return RoamingAccount{}.GetSchema().Attributes
}

// GatewaySettings maps the STFStoreGatewayServiceSetRequestModel struct.
type GatewaySettings struct {
	Enabled                  types.Bool   `tfsdk:"enable"`                      // Enable use of the gateway service
	CustomerId               types.String `tfsdk:"customer_id"`                 // The CWC customer id
	GetGatewayServiceUrl     types.String `tfsdk:"get_gateway_service_url"`     // The URL of the service used to retrieve gateway address (FQDN)
	PrivateKey               types.String `tfsdk:"private_key"`                 // The private key for CWC trust
	ServiceName              types.String `tfsdk:"service_name"`                // The service name for CWC trust
	InstanceId               types.String `tfsdk:"instance_id"`                 // The instance id for CWC trust
	SecureTicketAuthorityUrl types.String `tfsdk:"secure_ticket_authority_url"` // The URL of the CWC STA service
	SecureTicketLifetime     types.String `tfsdk:"secure_ticket_lifetime"`      // The lifetime requested for CWC STA service tickets
	SessionReliability       types.Bool   `tfsdk:"session_reliability"`         // A value indicating whether session reliability should be enabled
	IgnoreZones              types.List   `tfsdk:"ignore_zones"`                // List[string] A value indicating that the gateway service should not be used for the specified zones
	HandleZones              types.List   `tfsdk:"handle_zones"`                // List[string] A value indicating that the gateway service should be used for the specified zones
}

func (GatewaySettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Gateway service settings for the Store",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enable": schema.BoolAttribute{
				Description: "Enable use of the gateway service. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"customer_id": schema.StringAttribute{
				Description: "The CWC customer id",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"get_gateway_service_url": schema.StringAttribute{
				Description: "The URL of the service used to retrieve gateway address (FQDN)",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "The private key for CWC trust",
				Sensitive:   true,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "The service name for CWC trust",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"instance_id": schema.StringAttribute{
				Description: "The instance id for CWC trust",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secure_ticket_authority_url": schema.StringAttribute{
				Description: "The URL of the CWC STA service",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secure_ticket_lifetime": schema.StringAttribute{
				Description: "The lifetime requested for CWC STA service tickets. Default is 00:01:00.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("00:01:00"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"session_reliability": schema.BoolAttribute{
				Description: "A value indicating whether session reliability should be enabled. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ignore_zones": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "A value indicating that the gateway service should not be used for the specified zones",
				Optional:    true,
			},
			"handle_zones": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "A value indicating that the gateway service should be used for the specified zones",
				Optional:    true,
			},
		},
	}
}

func (GatewaySettings) GetAttributes() map[string]schema.Attribute {
	return GatewaySettings{}.GetSchema().Attributes
}

// EnumerationOptions maps the STFStoreEnumerationOptionsRequestModel struct.
type EnumerationOptions struct {
	EnhancedEnumeration                          types.Bool  `tfsdk:"enhanced_enumeration"`
	MaximumConcurrentEnumerations                types.Int64 `tfsdk:"maximum_concurrent_enumerations"`
	MinimumFarmsRequiredForConcurrentEnumeration types.Int64 `tfsdk:"minimum_farms_required_for_concurrent_enumeration"`
	FilterByTypesInclude                         types.List  `tfsdk:"filter_by_types_include"`    // List[string]
	FilterByKeywordsInclude                      types.List  `tfsdk:"filter_by_keywords_include"` // List[string]
	FilterByKeywordsExclude                      types.List  `tfsdk:"filter_by_keywords_exclude"` // List[string]
}

// LaunchOptions maps the STFStoreLaunchOptionsRequestModel struct.
type LaunchOptions struct {
	AddressResolutionType                  types.String `tfsdk:"address_resolution_type"`
	RequestIcaClientSecureChannel          types.String `tfsdk:"request_ica_client_secure_channel"`
	AllowSpecialFolderRedirection          types.Bool   `tfsdk:"allow_special_folder_redirection"`
	AllowFontSmoothing                     types.Bool   `tfsdk:"allow_font_smoothing"`
	RequireLaunchReference                 types.Bool   `tfsdk:"require_launch_reference"`
	OverrideIcaClientName                  types.Bool   `tfsdk:"override_ica_client_name"`
	OverlayAutoLoginCredentialsWithTicket  types.Bool   `tfsdk:"overlay_auto_login_credentials_with_ticket"`
	IgnoreClientProvidedClientAddress      types.Bool   `tfsdk:"ignore_client_provided_client_address"`
	SetNoLoadBiasFlag                      types.Bool   `tfsdk:"set_no_load_bias_flag"`
	RDPOnly                                types.Bool   `tfsdk:"rdp_only"`
	VdaLogonDataProvider                   types.String `tfsdk:"vda_logon_data_provider"`
	IcaTemplateName                        types.String `tfsdk:"ica_template_name"`
	FederatedAuthenticationServiceFailover types.Bool   `tfsdk:"federated_authentication_service_failover"`
}

type FarmSettings struct {
	EnableFileTypeAssociation          types.Bool   `tfsdk:"enable_file_type_association"`
	CommunicationTimeout               types.String `tfsdk:"communication_timeout"`
	ConnectionTimeout                  types.String `tfsdk:"connection_timeout"`
	LeasingStatusExpiryFailed          types.String `tfsdk:"leasing_status_expiry_failed"`
	LeasingStatusExpiryLeasing         types.String `tfsdk:"leasing_status_expiry_leasing"`
	LeasingStatusExpiryPending         types.String `tfsdk:"leasing_status_expiry_pending"`
	PooledSockets                      types.Bool   `tfsdk:"pooled_sockets"`
	ServerCommunicationAttempts        types.Int64  `tfsdk:"server_communication_attempts"`
	BackgroundHealthCheckPollingPeriod types.String `tfsdk:"background_healthcheck_polling"`
	AdvancedHealthCheck                types.Bool   `tfsdk:"advanced_healthcheck"`
	CertRevocationPolicy               types.String `tfsdk:"cert_revocation_policy"`
}

func (LaunchOptions) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Launch options for the Store",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"address_resolution_type": schema.StringAttribute{
				Description: "Specifies the type of address(Dns, DnsPort, IPV4, IPV4Port, Dot, DotPort, Uri, NoChange) to use in the .ica launch file. Default is DnsPort.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("DnsPort"),
				Validators: []validator.String{
					stringvalidator.OneOf("Dns", "DnsPort", "IPV4", "IPV4Port", "Dot", "DotPort", "Uri", "NoChange"),
				},
			},
			"request_ica_client_secure_channel": schema.StringAttribute{
				Description: "Specifies TLS settings(SSLAnyCiphers, TLSGovCipers, DetectAnyCiphers). Default is DetectAnyCipher.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("DetectAnyCiphers"),
				Validators: []validator.String{
					stringvalidator.OneOf("SSLAnyCiphers", "TLSGovCipers", "DetectAnyCiphers"),
				},
			},
			"allow_special_folder_redirection": schema.BoolAttribute{
				Description: "Redirect special folders such as Documents, Computer and the Desktop. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"allow_font_smoothing": schema.BoolAttribute{
				Description: "Specifies whether or not font smoothing is permitted for ICA sessions. Default is true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"require_launch_reference": schema.BoolAttribute{
				Description: "Specifies whether or not the use of launch references is enforced. Default is true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"override_ica_client_name": schema.BoolAttribute{
				Description: "Specifies whether or not a Web Interface-generated ID must be passed in the client name entry of an .ica launch file. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"overlay_auto_login_credentials_with_ticket": schema.BoolAttribute{
				Description: "Specifies whether a logon ticket must be duplicated in a logon ticket entry or placed in a separate .ica launch file ticket entry only. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ignore_client_provided_client_address": schema.BoolAttribute{
				Description: "Specifies whether or not to ignore the address provided by the Citrix client. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"set_no_load_bias_flag": schema.BoolAttribute{
				Description: "Specifies whether XenApp load bias should be used. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"rdp_only": schema.BoolAttribute{
				Description: "Configure the Store to only launch use the RDP protocol. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"vda_logon_data_provider": schema.StringAttribute{
				Description: "The Vda logon data provider to use during launch. Default is empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""), // Default is empty string
			},
			"ica_template_name": schema.StringAttribute{
				Description: "Ica template to use when launching an application or desktop. Default is empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""), // Default is empty string
			},
			"federated_authentication_service_failover": schema.BoolAttribute{
				Description: "Specifies whether to failover to launch without the Federated Auth Service (FAS) should it become uncontactable. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (FarmSettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Store farm configuration settings for the Store.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enable_file_type_association": schema.BoolAttribute{
				Description: "Enable File Type Association so that content is seamlessly redirected to users subscribed applications when they open local files of the appropriate types. Default value is true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"communication_timeout": schema.StringAttribute{
				Description: "Communication timeout when using to the Xml service in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:0:30.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:0:30"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"connection_timeout": schema.StringAttribute{
				Description: "Connection timeout when using to the Xml service in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:0:6.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:0:6"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"leasing_status_expiry_failed": schema.StringAttribute{
				Description: "Period of time before retrying a XenDesktop 7 and greater farm in failed leasing mode in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:3:0.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:3:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"leasing_status_expiry_leasing": schema.StringAttribute{
				Description: "Period of time before retrying a XenDesktop 7 and greater farm in leasing mode in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:3:0.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:3:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"leasing_status_expiry_pending": schema.StringAttribute{
				Description: "Period of time before retrying a XenDesktop 7 and greater farm in pending leasing mode in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:3:0.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:3:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"pooled_sockets": schema.BoolAttribute{
				Description: "Use pooled sockets so that StoreFront maintains a pool of sockets. Default value is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"server_communication_attempts": schema.Int64Attribute{
				Description: "Number of server connection attempts before failure. Default value is 1.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"background_healthcheck_polling": schema.StringAttribute{
				Description: "Period of time between polling servers in timestamp format, which must be in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:1:0.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:1:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"advanced_healthcheck": schema.BoolAttribute{
				Description: "Indicates whether advanced healthcheck should be performed. Default value is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cert_revocation_policy": schema.StringAttribute{
				Description: "Certificate Revocation Policy to use when connecting to XML services using HTTPS. Valid values are NoCheck (Default), MustCheck, FullCheck or NoNetworkAccess.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("NoCheck"),
				Validators: []validator.String{
					stringvalidator.OneOf("NoCheck", "FullCheck", "MustCheck", "NoNetworkAccess"),
				},
			},
		},
	}
}

func (FarmSettings) GetAttributes() map[string]schema.Attribute {
	return FarmSettings{}.GetSchema().Attributes
}

func (EnumerationOptions) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Enumeration options for the Store",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enhanced_enumeration": schema.BoolAttribute{
				Description: "Enable enhanced enumeration. Enumerate multiple farms in parallel to reduce operation time. Default is true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"maximum_concurrent_enumerations": schema.Int64Attribute{
				Description: "Maximum farms enumerated in parallel. Default is 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"minimum_farms_required_for_concurrent_enumeration": schema.Int64Attribute{
				Description: "Minimum farms required for concurrent enumeration. Default is 3.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
			},
			"filter_by_types_include": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Inclusive resource filter by type (Applications, Desktops or Documents). Default is empty list.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.NoneOf("")),
				},
			},
			"filter_by_keywords_include": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Only include applications and desktops that match the keywords. Default is empty list.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.NoneOf("")),
				},
			},
			"filter_by_keywords_exclude": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Exclude applications and desktops that match the keywords. Default is empty list.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.NoneOf("")),
				},
			},
		},
	}
}

func (EnumerationOptions) GetAttributes() map[string]schema.Attribute {
	return EnumerationOptions{}.GetSchema().Attributes
}

func (LaunchOptions) GetAttributes() map[string]schema.Attribute {
	return LaunchOptions{}.GetSchema().Attributes
}

// SFStoreServiceResourceModel maps the resource schema data.
type STFStoreServiceResourceModel struct {
	VirtualPath           types.String `tfsdk:"virtual_path"`
	SiteId                types.String `tfsdk:"site_id"`
	FriendlyName          types.String `tfsdk:"friendly_name"`
	AuthenticationService types.String `tfsdk:"authentication_service_virtual_path"`
	Anonymous             types.Bool   `tfsdk:"anonymous"`
	LoadBalance           types.Bool   `tfsdk:"load_balance"`
	PNA                   types.Object `tfsdk:"pna"`                 // PNA
	EnumerationOptions    types.Object `tfsdk:"enumeration_options"` // EnumerationOptions
	LaunchOptions         types.Object `tfsdk:"launch_options"`      // LaunchOptions
	FarmSettings          types.Object `tfsdk:"farm_settings"`
	GatewaySettings       types.Object `tfsdk:"gateway_settings"` // GatewaySettings
	RoamingAccount        types.Object `tfsdk:"roaming_account"`  // RoamingAccount
}

func (*stfStoreServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "StoreFront StoreService.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront storeservice. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_path": schema.StringAttribute{
				Description: "The IIS VirtualPath at which the Store will be configured to be accessed by Receivers.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name of the Store",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_service_virtual_path": schema.StringAttribute{
				Description: "The Virtual Path of the StoreFront Authentication Service to use for authenticating users.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"anonymous": schema.BoolAttribute{
				Description: "Whether the Store is anonymous. Anonymous Store not requiring authentication.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"load_balance": schema.BoolAttribute{
				Description: "Whether the Store is load balanced.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enumeration_options": EnumerationOptions{}.GetSchema(),
			"pna":                 PNA{}.GetSchema(),
			"launch_options":      LaunchOptions{}.GetSchema(),
			"farm_settings":       FarmSettings{}.GetSchema(),
			"gateway_settings":    GatewaySettings{}.GetSchema(),
			"roaming_account":     RoamingAccount{}.GetSchema(),
		},
	}
}

// PNA maps the resource schema data.
type PNA struct {
	Enable types.Bool `tfsdk:"enable"`
}

func (PNA) GetAttributes() map[string]schema.Attribute {
	return PNA{}.GetSchema().Attributes
}

func (PNA) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "StoreFront PNA (Program Neighborhood Agent) state of the Store",
		Attributes: map[string]schema.Attribute{
			"enable": schema.BoolAttribute{
				Description: "Whether PNA is enabled for the Store.",
				Required:    true,
			},
		},
	}
}

func (r *STFStoreServiceResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, storeService *citrixstorefront.STFStoreDetailModel) {
	// Overwrite STFStoreServiceResourceModel with refreshed state
	if storeService.VirtualPath.IsSet() {
		r.VirtualPath = types.StringValue(*storeService.VirtualPath.Get())
	}
	if storeService.SiteId.IsSet() {
		r.SiteId = types.StringValue(strconv.Itoa(*storeService.SiteId.Get()))
	}
	if storeService.FriendlyName.IsSet() {
		r.FriendlyName = types.StringValue(*storeService.FriendlyName.Get())
	}
}

func (r *STFStoreServiceResourceModel) RefreshPnaValues(ctx context.Context, diagnostics *diag.Diagnostics, pna citrixstorefront.STFPna) {
	refreshedPna := util.ObjectValueToTypedObject[PNA](ctx, diagnostics, r.PNA)
	if pna.PnaEnabled.IsSet() {
		refreshedPna.Enable = types.BoolValue(*pna.PnaEnabled.Get())
	}
	r.PNA = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedPna)
}

func (r *STFStoreServiceResourceModel) RefreshEnumerationOptions(ctx context.Context, diagnostics *diag.Diagnostics, response *citrixstorefront.GetSTFStoreEnumerationOptionsResponseModel) {
	refreshedEnumerationOptions := util.ObjectValueToTypedObject[EnumerationOptions](ctx, diagnostics, r.EnumerationOptions)
	if response.EnhancedEnumeration.IsSet() {
		refreshedEnumerationOptions.EnhancedEnumeration = types.BoolValue(*response.EnhancedEnumeration.Get())
	}
	if response.MaximumConcurrentEnumerations.IsSet() {
		refreshedEnumerationOptions.MaximumConcurrentEnumerations = types.Int64Value(*response.MaximumConcurrentEnumerations.Get())
	}
	if response.MinimumFarmsRequiredForConcurrentEnumeration.IsSet() {
		refreshedEnumerationOptions.MinimumFarmsRequiredForConcurrentEnumeration = types.Int64Value(*response.MinimumFarmsRequiredForConcurrentEnumeration.Get())
	}
	if len(response.FilterByTypesInclude) >= 1 && response.FilterByTypesInclude[0] != "" {
		refreshedEnumerationOptions.FilterByTypesInclude = util.RefreshListValues(ctx, diagnostics, refreshedEnumerationOptions.FilterByTypesInclude, response.FilterByTypesInclude)
	}
	if len(response.FilterByKeywordsInclude) >= 1 && response.FilterByKeywordsInclude[0] != "" {
		refreshedEnumerationOptions.FilterByKeywordsInclude = util.RefreshListValues(ctx, diagnostics, refreshedEnumerationOptions.FilterByKeywordsInclude, response.FilterByKeywordsInclude)
	}
	if len(response.FilterByKeywordsExclude) >= 1 && response.FilterByKeywordsExclude[0] != "" {
		refreshedEnumerationOptions.FilterByKeywordsExclude = util.RefreshListValues(ctx, diagnostics, refreshedEnumerationOptions.FilterByKeywordsExclude, response.FilterByKeywordsExclude)
	}

	refreshedEnumerationOptionsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedEnumerationOptions)

	r.EnumerationOptions = refreshedEnumerationOptionsObject
}

func (r *STFStoreServiceResourceModel) RefreshLaunchOptions(ctx context.Context, diagnostics *diag.Diagnostics, response *citrixstorefront.GetSTFStoreLaunchOptionsResponseModel) {
	refreshedLaunchOptions := util.ObjectValueToTypedObject[LaunchOptions](ctx, diagnostics, r.LaunchOptions)

	refreshedLaunchOptions.AddressResolutionType = types.StringValue(response.AddressResolutionType)
	refreshedLaunchOptions.RequestIcaClientSecureChannel = types.StringValue(response.RequestICAClientSecureChannel)

	if response.AllowSpecialFolderRedirection.IsSet() {
		refreshedLaunchOptions.AllowSpecialFolderRedirection = types.BoolValue(*response.AllowSpecialFolderRedirection.Get())
	}
	if response.AllowFontSmoothing.IsSet() {
		refreshedLaunchOptions.AllowFontSmoothing = types.BoolValue(*response.AllowFontSmoothing.Get())
	}
	if response.RequireLaunchReference.IsSet() {
		refreshedLaunchOptions.RequireLaunchReference = types.BoolValue(*response.RequireLaunchReference.Get())
	}
	if response.OverrideIcaClientName.IsSet() {
		refreshedLaunchOptions.OverrideIcaClientName = types.BoolValue(*response.OverrideIcaClientName.Get())
	}
	if response.OverlayAutoLoginCredentialsWithTicket.IsSet() {
		refreshedLaunchOptions.OverlayAutoLoginCredentialsWithTicket = types.BoolValue(*response.OverlayAutoLoginCredentialsWithTicket.Get())
	}
	if response.IgnoreClientProvidedClientAddress.IsSet() {
		refreshedLaunchOptions.IgnoreClientProvidedClientAddress = types.BoolValue(*response.IgnoreClientProvidedClientAddress.Get())
	}
	if response.SetNoLoadBiasFlag.IsSet() {
		refreshedLaunchOptions.SetNoLoadBiasFlag = types.BoolValue(*response.SetNoLoadBiasFlag.Get())
	}
	if response.FederatedAuthenticationServiceFailover.IsSet() {
		refreshedLaunchOptions.FederatedAuthenticationServiceFailover = types.BoolValue(*response.FederatedAuthenticationServiceFailover.Get())
	}
	if response.VdaLogonDataProviderName.IsSet() {
		refreshedLaunchOptions.VdaLogonDataProvider = types.StringValue(*response.VdaLogonDataProviderName.Get())
	}

	refreshedLaunchOptionsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedLaunchOptions)

	r.LaunchOptions = refreshedLaunchOptionsObject
}
func (r *STFStoreServiceResourceModel) RefreshGatewaySettings(ctx context.Context, diagnostics *diag.Diagnostics, gatewayService *citrixstorefront.STFStoreGatewayServiceResponseModel) {
	refreshedGatewaySettings := util.ObjectValueToTypedObject[GatewaySettings](ctx, diagnostics, r.GatewaySettings)
	if gatewayService.Enabled.IsSet() {
		refreshedGatewaySettings.Enabled = types.BoolValue(*gatewayService.Enabled.Get())
	}
	if gatewayService.CustomerId.IsSet() && *gatewayService.CustomerId.Get() != "" {
		refreshedGatewaySettings.CustomerId = types.StringValue(*gatewayService.CustomerId.Get())
	}
	if gatewayService.GetGatewayServiceUrl.IsSet() && *gatewayService.GetGatewayServiceUrl.Get() != "" {
		refreshedGatewaySettings.GetGatewayServiceUrl = types.StringValue(*gatewayService.GetGatewayServiceUrl.Get())
	}
	if gatewayService.PrivateKey.IsSet() && *gatewayService.PrivateKey.Get() != "" {
		refreshedGatewaySettings.PrivateKey = types.StringValue(*gatewayService.PrivateKey.Get())
	}
	if gatewayService.ServiceName.IsSet() && *gatewayService.ServiceName.Get() != "" {
		refreshedGatewaySettings.ServiceName = types.StringValue(*gatewayService.ServiceName.Get())
	}
	if gatewayService.InstanceId.IsSet() && *gatewayService.InstanceId.Get() != "" {
		refreshedGatewaySettings.InstanceId = types.StringValue(*gatewayService.InstanceId.Get())
	}
	if gatewayService.SecureTicketAuthorityUrl.IsSet() && *gatewayService.SecureTicketAuthorityUrl.Get() != "" {
		refreshedGatewaySettings.SecureTicketAuthorityUrl = types.StringValue(*gatewayService.SecureTicketAuthorityUrl.Get())
	}

	if gatewayService.SessionReliability.IsSet() && *gatewayService.SessionReliability.Get() {
		refreshedGatewaySettings.SessionReliability = types.BoolValue(*gatewayService.SessionReliability.Get())
	}

	if gatewayService.IgnoreZones != nil && len(gatewayService.IgnoreZones) >= 1 {
		refreshedGatewaySettings.IgnoreZones = util.RefreshListValues(ctx, diagnostics, refreshedGatewaySettings.IgnoreZones, gatewayService.IgnoreZones)
	}

	if gatewayService.HandleZones != nil && len(gatewayService.HandleZones) >= 1 {
		refreshedGatewaySettings.HandleZones = util.RefreshListValues(ctx, diagnostics, refreshedGatewaySettings.HandleZones, gatewayService.HandleZones)
	}

	refreshedGatewaySettings.SecureTicketLifetime = types.StringValue(gatewayService.SecureTicketLifetime)
	r.GatewaySettings = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedGatewaySettings)
}

func (r *STFStoreServiceResourceModel) RefreshFarmSettings(ctx context.Context, diagnostics *diag.Diagnostics, response *citrixstorefront.StoreFarmConfigurationResponseModel) {
	refreshedStoreFarmSettings := util.ObjectValueToTypedObject[FarmSettings](ctx, diagnostics, r.FarmSettings)

	if response.EnableFileTypeAssociation.IsSet() {
		refreshedStoreFarmSettings.EnableFileTypeAssociation = types.BoolValue(*response.EnableFileTypeAssociation.Get())
	}

	refreshedStoreFarmSettings.CommunicationTimeout = types.StringValue(response.CommunicationTimeout)
	refreshedStoreFarmSettings.ConnectionTimeout = types.StringValue(response.ConnectionTimeout)
	refreshedStoreFarmSettings.LeasingStatusExpiryFailed = types.StringValue(response.LeasingStatusExpiryFailed)
	refreshedStoreFarmSettings.LeasingStatusExpiryPending = types.StringValue(response.LeasingStatusExpiryPending)
	refreshedStoreFarmSettings.LeasingStatusExpiryLeasing = types.StringValue(response.LeasingStatusExpiryLeasing)

	if response.PooledSockets.IsSet() {
		refreshedStoreFarmSettings.PooledSockets = types.BoolValue(*response.PooledSockets.Get())
	}

	if response.ServerCommunicationAttempts.IsSet() {
		int64Val := int64(*response.ServerCommunicationAttempts.Get())
		refreshedStoreFarmSettings.ServerCommunicationAttempts = types.Int64Value(int64Val)
	}

	refreshedStoreFarmSettings.BackgroundHealthCheckPollingPeriod = types.StringValue(response.BackgroundHealthCheckPollingPeriod)

	if response.AdvancedHealthCheck.IsSet() {
		refreshedStoreFarmSettings.AdvancedHealthCheck = types.BoolValue(*response.AdvancedHealthCheck.Get())
	}

	if response.CertRevocationPolicy.IsSet() {
		refreshedStoreFarmSettings.CertRevocationPolicy = types.StringValue((*response.CertRevocationPolicy.Get()))
	}
	refreshedStoreFarmSettingsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedStoreFarmSettings)

	r.FarmSettings = refreshedStoreFarmSettingsObject
}

func (r *STFStoreServiceResourceModel) RefreshRoamingAccount(ctx context.Context, diagnostics *diag.Diagnostics, roamingAccount *citrixstorefront.GetSTFRoamingAccountResponseModel) {
	refreshedRoamingAccount := util.ObjectValueToTypedObject[RoamingAccount](ctx, diagnostics, r.RoamingAccount)

	if roamingAccount.Published.IsSet() {
		refreshedRoamingAccount.Published = types.BoolValue(*roamingAccount.Published.Get())
	}
	refreshedRoamingObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedRoamingAccount)

	r.RoamingAccount = refreshedRoamingObject

}
