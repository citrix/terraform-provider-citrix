// Copyright © 2024. Citrix Systems, Inc.

package stf_store

import (
	"context"
	"regexp"

	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

type StoreFarm struct {
	// StoreService               types.String `tfsdk:"store_virtual_path"`             // The virtual path of the StoreService.
	FarmName                   types.String `tfsdk:"farm_name"`                      // The name of the Farm.
	FarmType                   types.String `tfsdk:"farm_type"`                      // The type of the Farm.
	Servers                    types.List   `tfsdk:"servers"`                        // List[string] The list of servers in the Farm.
	Port                       types.Int64  `tfsdk:"port"`                           // Service communication port.
	SSLRelayPort               types.Int64  `tfsdk:"ssl_relay_port"`                 // The SSL Relay port
	TransportType              types.String `tfsdk:"transport_type"`                 // Type of transport to use. Http, Https, SSL for example
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

func (StoreFarm) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
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
					int64validator.Between(1, 65535),
				},
			},
			"ssl_relay_port": schema.Int64Attribute{
				Description: "The SSL Relay port. Default is 443",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(443),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.Between(1, 65535),
				},
			},
			"transport_type": schema.StringAttribute{
				Description: "Type of transport to use. Http, Https, SSL for example. Default to HTTPs.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTPS", "HTTP", "SSL",
					),
				},
				Default: stringdefault.StaticString("HTTPS"),
			},
			"load_balance": schema.BoolAttribute{
				Description: "Round robin load balance the xml service servers. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"xml_validation_enabled": schema.BoolAttribute{
				Description: "Enable XML service endpoint validation. Defaults to false.",
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
				Description: "Period of time to skip all xml service requests should all servers fail to respond. Defaults to 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
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
				Description: "Period of time an ICA launch ticket is valid once requested on pre 7.0 XenApp and XenDesktop farms. Defaults to 200",
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
				Description: "Cloud deployments only otherwise ignored. The product name of the farm configured. Defaults to empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"restrict_pops": schema.StringAttribute{
				Description: "Cloud deployments only otherwise ignored. Restricts GWaaS traffic to the specified POP. Defaults to empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"farm_guid": schema.StringAttribute{
				Description: "A tag indicating the scope of the farm. Valid for cloud deployments only. Defaults to empty string.",
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

func (StoreFarm) GetAttributes() map[string]schema.Attribute {
	return StoreFarm{}.GetSchema().Attributes
}

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
	StoreFarm             types.List   `tfsdk:"farms"`               // List[StoreFarm]
	PNA                   types.Object `tfsdk:"pna"`                 // PNA
	EnumerationOptions    types.Object `tfsdk:"enumeration_options"` // EnumerationOptions
	LaunchOptions         types.Object `tfsdk:"launch_options"`      // LaunchOptions
	FarmSettings          types.Object `tfsdk:"farm_settings"`
	RoamingAccount        types.Object `tfsdk:"roaming_account"` // RoamingAccount
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

func (r *STFStoreServiceResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, storeService *citrixstorefront.STFStoreDetailModel, farms []citrixstorefront.StoreFarmModel) {
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
	var farmList []StoreFarm
	for _, farm := range farms {
		storefarm := StoreFarm{
			FarmName:                   types.StringValue(*farm.FarmName.Get()),
			FarmType:                   types.StringValue(FarmTypeFromInt(*farm.FarmType.Get())),
			Port:                       types.Int64Value(*farm.Port.Get()),
			SSLRelayPort:               types.Int64Value(*farm.SSLRelayPort.Get()),
			Product:                    types.StringValue(*farm.Product.Get()),
			RestrictPoPs:               types.StringValue(*farm.RestrictPoPs.Get()),
			FarmGuid:                   types.StringValue(*farm.FarmGuid.Get()),
			TransportType:              types.StringValue(TransportTypeFromInt(*farm.TransportType.Get())),
			LoadBalance:                types.BoolValue(*farm.LoadBalance.Get()),
			XMLValidationEnabled:       types.BoolValue(*farm.XMLValidationEnabled.Get()),
			XMLValidationSecret:        types.StringValue(*farm.XMLValidationSecret.Get()),
			AllFailedBypassDuration:    types.Int64Value(*farm.AllFailedBypassDuration.Get()),
			BypassDuration:             types.Int64Value(*farm.BypassDuration.Get()),
			TicketTimeToLive:           types.Int64Value(*farm.TicketTimeToLive.Get()),
			RadeTicketTimeToLive:       types.Int64Value(*farm.RadeTicketTimeToLive.Get()),
			MaxFailedServersPerRequest: types.Int64Value(*farm.MaxFailedServersPerRequest.Get()),
		}
		storefarm.Servers = util.RefreshListValues(ctx, diagnostics, storefarm.Servers, farm.Servers)
		storefarm.Zones = util.RefreshListValues(ctx, diagnostics, storefarm.Zones, farm.Zones)
		storefarm.ServiceUrls = util.RefreshListValues(ctx, diagnostics, storefarm.ServiceUrls, farm.ServiceUrls)
		farmList = append(farmList, storefarm)
	}
	r.StoreFarm = util.TypedArrayToObjectList[StoreFarm](ctx, diagnostics, farmList)
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

func TransportTypeFromInt(TransportTypeInt int64) string {
	switch TransportTypeInt {
	case 0:
		return "HTTP"
	case 1:
		return "HTTPS"
	case 2:
		return "SSL"
	default:
		return "Unknown"
	}
}

func (r *STFStoreServiceResourceModel) RefreshPnaValues(ctx context.Context, diagnostics *diag.Diagnostics, pna *citrixstorefront.STFPna) {
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

func (STFStoreServiceResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront StoreService.",
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
			"farms": schema.ListNestedAttribute{
				Description:  "A list of StoreFront Controller.",
				Required:     true,
				NestedObject: StoreFarm{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"enumeration_options": EnumerationOptions{}.GetSchema(),
			"pna":                 PNA{}.GetSchema(),
			"launch_options":      LaunchOptions{}.GetSchema(),
			"farm_settings":       FarmSettings{}.GetSchema(),
			"roaming_account":     RoamingAccount{}.GetSchema(),
		},
	}
}
