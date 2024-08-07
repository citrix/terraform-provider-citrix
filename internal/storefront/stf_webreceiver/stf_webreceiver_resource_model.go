// Copyright © 2024. Citrix Systems, Inc.

package stf_webreceiver

import (
	"context"
	"regexp"
	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebReceiverSiteStyle struct {
	HeaderLogoPath         types.String `tfsdk:"header_logo_path"`
	LogonLogoPath          types.String `tfsdk:"logon_logo_path"`
	HeaderBackgroundColor  types.String `tfsdk:"header_background_color"`
	HeaderForegroundColor  types.String `tfsdk:"header_foreground_color"`
	LinkColor              types.String `tfsdk:"link_color"`
	IgnoreNonExistentLogos types.Bool   `tfsdk:"ignore_non_existent_logos"`
}

func (WebReceiverSiteStyle) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Site Styles for the Web Receiver for Website.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"header_logo_path": schema.StringAttribute{
				Description: "Points to the Header Logo's path in the system.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFrontReceiverLogo_Home@2x_B07AF017CEE39553.png"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^.*\.(png|jpg|jpeg|gif|tiff|bmp)$`), "must be a valid image file"),
				},
			},
			"logon_logo_path": schema.StringAttribute{
				Description: "Points to the Logon Logo's path in the system.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("C:\\inetpub\\wwwroot\\Citrix\\StoreWeb\\receiver\\images\\2x\\CitrixStoreFront_auth@2x_CB5D9D1BADB08AFF.png"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^.*\.(png|jpg|jpeg|gif|tiff|bmp)$`), "must be a valid image file"),
				},
			},
			"header_background_color": schema.StringAttribute{
				Description: "Sets the background color of the header.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("#312139"),
			},
			"header_foreground_color": schema.StringAttribute{
				Description: "Sets the foreground color of the header.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("#fff"),
			},
			"link_color": schema.StringAttribute{
				Description: "Sets the link color of the page.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("#67397a"),
			},
			"ignore_non_existent_logos": schema.BoolAttribute{
				Description: "Whether to ignore non-existent logo files and continue to set colors.",
				Optional:    true,
			},
		},
	}
}

func (WebReceiverSiteStyle) GetAttributes() map[string]schema.Attribute {
	return WebReceiverSiteStyle{}.GetSchema().Attributes
}

type PluginAssistant struct {
	Enabled                                       types.Bool   `tfsdk:"enabled"`                                              //Enable Receiver client detection.
	UpgradeAtLogin                                types.Bool   `tfsdk:"upgrade_at_login"`                                     //Prompt to upgrade older clients.
	ShowAfterLogin                                types.Bool   `tfsdk:"show_after_login"`                                     //Show Receiver client detection after the user logs in.
	Win32Path                                     types.String `tfsdk:"win32_path"`                                           //Path to the Windows Receiver.
	MacOSPath                                     types.String `tfsdk:"macos_path"`                                           //Path to the MacOS Receiver.
	MacOSMinimumSupportedVersion                  types.String `tfsdk:"macos_minimum_supported_version"`                      //Minimum version of the MacOS supported.
	Html5SingleTabLaunch                          types.Bool   `tfsdk:"html5_single_tab_launch"`                              //Launch Html5 Receiver in the same browser tab.
	Html5Enabled                                  types.String `tfsdk:"html5_enabled"`                                        //Method of deploying and using the Html5 Receiver.
	Html5Platforms                                types.String `tfsdk:"html5_platforms"`                                      //The supported Html5 platforms.
	Html5Preferences                              types.String `tfsdk:"html5_preferences"`                                    //Html5 Receiver preferences.
	Html5ChromeAppOrigins                         types.String `tfsdk:"html5_chrome_app_origins"`                             //The Html5 Chrome Application Origins settings.
	Html5ChromeAppPreferences                     types.String `tfsdk:"html5_chrome_app_preferences"`                         //The Html5 Chrome Application preferences.
	ProtocolHandlerEnabled                        types.Bool   `tfsdk:"protocol_handler_enabled"`                             //Enable the Receiver Protocol Handler.
	ProtocolHandlerPlatforms                      types.String `tfsdk:"protocol_handler_platforms"`                           //The supported Protocol Handler platforms.
	ProtocolHandlerSkipDoubleHopCheckWhenDisabled types.Bool   `tfsdk:"protocol_handler_skip_double_hop_check_when_disabled"` //Skip the Protocol Handle double hop check.
}

func (PluginAssistant) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Pluin Assistant configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Enable the Plugin Assistant.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"upgrade_at_login": schema.BoolAttribute{
				Description: "Prompt to upgrade older clients.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"show_after_login": schema.BoolAttribute{
				Description: "Show Plugin Assistant after the user logs in.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"win32_path": schema.StringAttribute{
				Description: "Path to the Windows Receiver.",
				Optional:    true,
			},
			"macos_path": schema.StringAttribute{
				Description: "Path to the MacOS Receiver.",
				Optional:    true,
			},
			"macos_minimum_supported_version": schema.StringAttribute{
				Description: "Minimum version of the MacOS supported.",
				Optional:    true,
			},
			"html5_single_tab_launch": schema.BoolAttribute{
				Description: "Launch Html5 Receiver in the same browser tab.",
				Optional:    true,
			},
			"html5_enabled": schema.StringAttribute{
				Description: "Method of deploying and using the Html5 Receiver.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Off"),
			},
			"html5_platforms": schema.StringAttribute{
				Description: "The supported Html5 platforms.",
				Optional:    true,
			},
			"html5_preferences": schema.StringAttribute{
				Description: "Html5 Receiver preferences.",
				Optional:    true,
			},
			"html5_chrome_app_origins": schema.StringAttribute{
				Description: "The Html5 Chrome Application Origins settings.",
				Optional:    true,
			},
			"html5_chrome_app_preferences": schema.StringAttribute{
				Description: "The Html5 Chrome Application preferences.",
				Optional:    true,
			},
			"protocol_handler_enabled": schema.BoolAttribute{
				Description: "Enable the Receiver Protocol Handler.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"protocol_handler_platforms": schema.StringAttribute{
				Description: "The supported Protocol Handler platforms.",
				Optional:    true,
			},
			"protocol_handler_skip_double_hop_check_when_disabled": schema.BoolAttribute{
				Description: "Skip the Protocol Handle double hop check.",
				Optional:    true,
			},
		},
	}
}

func (PluginAssistant) GetAttributes() map[string]schema.Attribute {
	return PluginAssistant{}.GetSchema().Attributes
}

type ApplicationShortcuts struct {
	PromptForUntrustedShortcuts types.Bool `tfsdk:"prompt_for_untrusted_shortcuts"`
	TrustedUrls                 types.Set  `tfsdk:"trusted_urls"` // Set[string]
	GatewayUrls                 types.Set  `tfsdk:"gateway_urls"` // Set[string]
}

func (ApplicationShortcuts) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Application shortcuts configurations for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"prompt_for_untrusted_shortcuts": schema.BoolAttribute{
				Description: "Display confirmation dialog when Receiver for Web cannot determine if an app shortcut originated from a trusted internal site.",
				Required:    true,
			},
			"trusted_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Set of internal web sites that will provide app shortcuts to users.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"gateway_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Set of gateways through which shortcuts will be provided to users.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}

func (ApplicationShortcuts) GetAttributes() map[string]schema.Attribute {
	return ApplicationShortcuts{}.GetSchema().Attributes
}

type Communication struct {
	Attempts              types.Int64  `tfsdk:"attempts"`                 // Number of attempts to connect to the Store Service.
	Timeout               types.String `tfsdk:"timeout"`                  // Timeout for the connection to the Store Service.
	Loopback              types.String `tfsdk:"loopback"`                 // Enable loopback communication.
	LoopbackPortUsingHttp types.Int64  `tfsdk:"loopback_port_using_http"` // Use HTTP for loopback communication.
	ProxyEnabled          types.Bool   `tfsdk:"proxy_enabled"`            // Enable proxy for communication.
	ProxyPort             types.Int64  `tfsdk:"proxy_port"`               // Port for the proxy.
	ProxyProcessName      types.String `tfsdk:"proxy_process_name"`       // Process name for the proxy.
}

func (Communication) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Communication settings used for the WebReceiver proxy.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"attempts": schema.Int64Attribute{
				Description: "The number of attempts WebReceiver should make to contact StoreFront before it gives up. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"timeout": schema.StringAttribute{
				Description: "Timeout value for communicating with StoreFront in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `0.0:3:0`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0.0:3:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
			"loopback": schema.StringAttribute{
				Description: "Whether to use the loopback address for communications with the store service, rather than the actual StoreFront server URL. Available values are `On`, `Off`, `OnUsingHttp`. Defaults to `Off`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Off"),
				Validators: []validator.String{
					stringvalidator.OneOf("On", "Off", "OnUsingHttp"),
				},
			},
			"loopback_port_using_http": schema.Int64Attribute{
				Description: "When loopback is set to `OnUsingHttp`, the port number to use for loopback communications. Defaults to `80`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(80),
			},
			"proxy_enabled": schema.BoolAttribute{
				Description: "Whether the communications proxy is enabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"proxy_port": schema.Int64Attribute{
				Description: "The port to use for the communications proxy. Defaults to `8888`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8888),
			},
			"proxy_process_name": schema.StringAttribute{
				Description: "The name of the process acting as proxy. Defaults to `Fiddler`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Fiddler"),
			},
		},
	}
}

func (Communication) GetAttributes() map[string]schema.Attribute {
	return Communication{}.GetSchema().Attributes
}

type StrictTransportSecurity struct {
	Enabled        types.Bool   `tfsdk:"enabled"`         // Enable Strict Transport Security.
	PolicyDuration types.String `tfsdk:"policy_duration"` // Duration of the policy.
}

func (StrictTransportSecurity) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Communication settings used for the WebReceiver proxy.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether to enable the HTTP Strict Transport Security feature. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"policy_duration": schema.StringAttribute{
				Description: "The time period for which browsers should apply HSTS to the RfWeb site in `dd.hh:mm:ss` format with 0's trimmed. Defaults to `90.0:0:0`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("90.0:0:0"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeSpanRegex), "must be in `dd.hh:mm:ss` format with 0's trimmed."),
				},
			},
		},
	}
}

func (StrictTransportSecurity) GetAttributes() map[string]schema.Attribute {
	return StrictTransportSecurity{}.GetSchema().Attributes
}

type UserInterface struct {
	AutoLaunchDesktop      types.Bool   `tfsdk:"auto_launch_desktop"`
	MultiClickTimeout      types.Int64  `tfsdk:"multi_click_timeout"`
	EnableAppsFolderView   types.Bool   `tfsdk:"enable_apps_folder_view"`
	WorkspaceControl       types.Object `tfsdk:"workspace_control"`      // WorkspaceControl
	ReceiverConfiguration  types.Object `tfsdk:"receiver_configuration"` // ReceiverConfiguration
	AppShortcuts           types.Object `tfsdk:"app_shortcuts"`          // AppShortcuts
	UIViews                types.Object `tfsdk:"ui_views"`               // UIViews
	CategoryViewCollapsed  types.Bool   `tfsdk:"category_view_collapsed"`
	MoveAppToUncategorized types.Bool   `tfsdk:"move_app_to_uncategorized"`
	ProgressiveWebApp      types.Object `tfsdk:"progressive_web_app"` // ProgressiveWebApp
	ShowActivityManager    types.Bool   `tfsdk:"show_activity_manager"`
	ShowFirstTimeUse       types.Bool   `tfsdk:"show_first_time_use"`
	PreventIcaDownloads    types.Bool   `tfsdk:"prevent_ica_downloads"`
}

func (UserInterface) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "User interface configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"auto_launch_desktop": schema.BoolAttribute{
				Description: "Whether to auto-launch desktop at login if there is only one desktop available for the user. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"multi_click_timeout": schema.Int64Attribute{
				Description: "The time period in seconds for which the spinner control is displayed, after the user clicks on the App/Desktop icon within Receiver for Web. Defaults to `3`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
			},
			"enable_apps_folder_view": schema.BoolAttribute{
				Description: "Allows the user to turn off folder view when in a locked-down store or unauthenticated store. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"workspace_control":      WorkspaceControl{}.GetSchema(),
			"receiver_configuration": ReceiverConfiguration{}.GetSchema(),
			"app_shortcuts":          AppShortcuts{}.GetSchema(),
			"ui_views":               UIViews{}.GetSchema(),
			"category_view_collapsed": schema.BoolAttribute{
				Description: "Collapse the category view so that only the immediate contents of the selected category/sub-catagory are displayed. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"move_app_to_uncategorized": schema.BoolAttribute{
				Description: "Move uncategorized apps into a folder named ‘Uncategorized’ when the category view is collapsed. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"progressive_web_app": ProgressiveWebApp{}.GetSchema(),
			"show_activity_manager": schema.BoolAttribute{
				Description: "Enable the Activity Manager within the end user interface. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"show_first_time_use": schema.BoolAttribute{
				Description: "Enable the showing of the First Time Use screen within the end user interface. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"prevent_ica_downloads": schema.BoolAttribute{
				Description: "Prevent download of ICA Files. Defaults to `false`. StoreFront version 2402 or higher is required to modify this setting.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (UserInterface) GetAttributes() map[string]schema.Attribute {
	return UserInterface{}.GetSchema().Attributes
}

type WorkspaceControl struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	AutoReconnectAtLogon types.Bool   `tfsdk:"auto_reconnect_at_logon"`
	LogoffAction         types.String `tfsdk:"logoff_action"`
	ShowReconnectButton  types.Bool   `tfsdk:"show_reconnect_button"`
	ShowDisconnectButton types.Bool   `tfsdk:"show_disconnect_button"`
}

func (WorkspaceControl) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Workspace control configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether to enable workspace control. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"auto_reconnect_at_logon": schema.BoolAttribute{
				Description: "Whether to perform auto-reconnect at login. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"logoff_action": schema.StringAttribute{
				Description: "Whether to disconnect or terminate HDX sessions when actively logging off Receiver for Web. Available values are `Disconnect`, `Terminate`, and `None`. Defaults to `Disconnect`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Disconnect"),
				Validators: []validator.String{
					stringvalidator.OneOf("Disconnect", "Terminate", "None"),
				},
			},
			"show_reconnect_button": schema.BoolAttribute{
				Description: "Whether to show the reconnect button/link. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"show_disconnect_button": schema.BoolAttribute{
				Description: "Whether to show the disconnect button/link. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (WorkspaceControl) GetAttributes() map[string]schema.Attribute {
	return WorkspaceControl{}.GetSchema().Attributes
}

type ReceiverConfiguration struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	DownloadUrl types.String `tfsdk:"download_url"`
}

func (ReceiverConfiguration) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Receiver configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Enable the Receiver Configuration .cr file download. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"download_url": schema.StringAttribute{
				Description: "The URL to download the Receiver Configuration .cr file.",
				Computed:    true,
			},
		},
	}
}

func (ReceiverConfiguration) GetAttributes() map[string]schema.Attribute {
	return ReceiverConfiguration{}.GetSchema().Attributes
}

type AppShortcuts struct {
	Enabled               types.Bool `tfsdk:"enabled"`
	AllowSessionReconnect types.Bool `tfsdk:"allow_session_reconnect"`
}

func (AppShortcuts) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "App shortcuts configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Enable app shortcuts. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"allow_session_reconnect": schema.BoolAttribute{
				Description: "Enable App Shortcuts to support session reconnect. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (AppShortcuts) GetAttributes() map[string]schema.Attribute {
	return AppShortcuts{}.GetSchema().Attributes
}

type UIViews struct {
	ShowAppsView     types.Bool   `tfsdk:"show_apps_view"`
	ShowDesktopsView types.Bool   `tfsdk:"show_desktops_view"`
	DefaultView      types.String `tfsdk:"default_view"`
}

func (UIViews) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "UI view configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"show_apps_view": schema.BoolAttribute{
				Description: "Whether to show the apps view tab. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"show_desktops_view": schema.BoolAttribute{
				Description: "Whether to show the desktops tab. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"default_view": schema.StringAttribute{
				Description: "The view to show after logon. Available values are `Auto`, `Desktops`, and `Apps`. Defaults to `Auto`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Auto"),
				Validators: []validator.String{
					stringvalidator.OneOf("Auto", "Desktops", "Apps"),
				},
			},
		},
	}
}

func (UIViews) GetAttributes() map[string]schema.Attribute {
	return UIViews{}.GetSchema().Attributes
}

type AuthenticationManager struct {
	LoginFormTimeout     types.Int64  `tfsdk:"login_form_timeout"`
	GetUserNameUrl       types.String `tfsdk:"get_user_name_url"`
	LogoffUrl            types.String `tfsdk:"logoff_url"`
	ChangeCredentialsUrl types.String `tfsdk:"change_credentials_url"`
}

func (AuthenticationManager) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "WebReceiver Authentication Manager client options.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"login_form_timeout": schema.Int64Attribute{
				Description: "The WebReceiver login form timeout in minutes. Defaults to `5`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(5),
			},
			"get_user_name_url": schema.StringAttribute{
				Description: "The URL to obtain the full username. Defaults to `Authentication/GetUserName`.",
				Computed:    true,
			},
			"logoff_url": schema.StringAttribute{
				Description: "The URL to log off the Citrix Receiver for Web session. Defaults to `Authentication/Logoff`.",
				Computed:    true,
			},
			"change_credentials_url": schema.StringAttribute{
				Description: "The URL to initiate a change password operation. Defaults to `ExplicitAuth/GetChangeCredentialForm`.",
				Computed:    true,
			},
		},
	}
}

func (AuthenticationManager) GetAttributes() map[string]schema.Attribute {
	return AuthenticationManager{}.GetSchema().Attributes
}

type ProgressiveWebApp struct {
	Enabled           types.Bool `tfsdk:"enabled"`
	ShowInstallPrompt types.Bool `tfsdk:"show_install_prompt"`
}

func (ProgressiveWebApp) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Progressive Web App configuration for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Enable Progressive Web App support. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"show_install_prompt": schema.BoolAttribute{
				Description: "Enable prompt to install Progressive Web App. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (ProgressiveWebApp) GetAttributes() map[string]schema.Attribute {
	return ProgressiveWebApp{}.GetSchema().Attributes
}

// SFWebReceiverResourceModel maps the resource schema data.
type STFWebReceiverResourceModel struct {
	VirtualPath             types.String `tfsdk:"virtual_path"`
	SiteId                  types.String `tfsdk:"site_id"`
	FriendlyName            types.String `tfsdk:"friendly_name"`
	StoreServiceVirtualPath types.String `tfsdk:"store_virtual_path"`
	PluginAssistant         types.Object `tfsdk:"plugin_assistant"`          // PluginAssistant
	AuthenticationMethods   types.Set    `tfsdk:"authentication_methods"`    // Set[string]
	ApplicationShortcuts    types.Object `tfsdk:"application_shortcuts"`     // ApplicationShortcuts
	Communication           types.Object `tfsdk:"communication"`             // Communication
	StrictTransportSecurity types.Object `tfsdk:"strict_transport_security"` // StrictTransportSecurity
	AuthenticationManager   types.Object `tfsdk:"authentication_manager"`    // AuthenticationManager
	UserInterface           types.Object `tfsdk:"user_interface"`            // UserInterface
	ResourcesService        types.Object `tfsdk:"resources_service"`         // ResourcesServiceModel
	WebReceiverSiteStyle    types.Object `tfsdk:"web_receiver_site_style"`   // WebReceiverSiteStyle
}

type ResourcesService struct {
	PersistentIconCacheEnabled types.Bool  `tfsdk:"persistent_icon_cache_enabled"`
	IcaFileCacheExpiry         types.Int64 `tfsdk:"ica_file_cache_expiry"`
	IconSize                   types.Int64 `tfsdk:"icon_size"`
	ShowDesktopViewer          types.Bool  `tfsdk:"show_desktop_viewer"`
}

func (ResourcesService) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Resources Service settings for the WebReceiver.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"persistent_icon_cache_enabled": schema.BoolAttribute{
				Description: "Whether to cache icon data in the local file system. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ica_file_cache_expiry": schema.Int64Attribute{
				Description: "How long the ICA file data is cached in the memory of the Web Proxy. Defaults to `90`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(90),
			},
			"icon_size": schema.Int64Attribute{
				Description: "The desired icon size sent to the Store Service in icon requests. Defaults to `128`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(128),
				Validators: []validator.Int64{
					int64validator.OneOf(16, 24, 32, 48, 64, 96, 128, 256, 512),
				},
			},
			"show_desktop_viewer": schema.BoolAttribute{
				Description: "Shows the Citrix Desktop Viewer window and toolbar when users access their desktops from legacy clients. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (ResourcesService) GetAttributes() map[string]schema.Attribute {
	return ResourcesService{}.GetSchema().Attributes
}

func (r *STFWebReceiverResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, webreceiver *citrixstorefront.STFWebReceiverDetailModel, appShortcuts *citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel, communication *citrixstorefront.GetWebReceiverCommunicationResponseModel, sts *citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel, authManager *citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel, ui *citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel, resourcesService *citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel, siteStyle *citrixstorefront.STFWebReceiverSiteStyleResponseModel) {
	// Overwrite SFWebReceiverResourceModel with refreshed state
	r.VirtualPath = types.StringValue(*webreceiver.VirtualPath.Get())
	r.SiteId = types.StringValue(strconv.Itoa(*webreceiver.SiteId.Get()))
	r.FriendlyName = types.StringValue(*webreceiver.FriendlyName.Get())

	if !r.ApplicationShortcuts.IsNull() {
		r.ApplicationShortcuts = r.RefreshApplicationShortcuts(ctx, diagnostics, appShortcuts)
	}

	if !r.Communication.IsNull() {
		r.Communication = r.RefreshCommunication(ctx, diagnostics, communication)
	}

	if !r.StrictTransportSecurity.IsNull() {
		r.StrictTransportSecurity = r.RefreshStrictTransportSecurity(ctx, diagnostics, sts)
	}

	if !r.AuthenticationManager.IsNull() {
		r.AuthenticationManager = r.RefreshAuthenticationManager(ctx, diagnostics, authManager)
	}

	if !r.UserInterface.IsNull() {
		r.UserInterface = r.RefreshUserInterface(ctx, diagnostics, ui)
	}

	if !r.ResourcesService.IsNull() {
		r.ResourcesService = r.RefreshResourcesService(ctx, diagnostics, resourcesService)
	}

	if !r.WebReceiverSiteStyle.IsNull() {
		r.WebReceiverSiteStyle = r.RefreshWebReceiverSiteStyle(ctx, diagnostics, siteStyle)
	}
}

func (r *STFWebReceiverResourceModel) RefreshApplicationShortcuts(ctx context.Context, diagnostics *diag.Diagnostics, appShortcuts *citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel) types.Object {
	refreshedApplicationShortcuts := ApplicationShortcuts{}
	refreshedApplicationShortcuts.PromptForUntrustedShortcuts = types.BoolValue(*appShortcuts.PromptForUntrustedShortcuts.Get())
	if len(appShortcuts.GetTrustedUrls()) > 0 {
		refreshedApplicationShortcuts.TrustedUrls = util.StringArrayToStringSet(ctx, diagnostics, appShortcuts.GetTrustedUrls())
	} else {
		refreshedApplicationShortcuts.TrustedUrls = types.SetValueMust(types.StringType, []attr.Value{})
	}

	if len(appShortcuts.GetGatewayUrls()) > 0 {
		refreshedApplicationShortcuts.GatewayUrls = util.StringArrayToStringSet(ctx, diagnostics, appShortcuts.GetGatewayUrls())
	} else {
		refreshedApplicationShortcuts.GatewayUrls = types.SetValueMust(types.StringType, []attr.Value{})
	}
	refreshedApplicationShortcutsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedApplicationShortcuts)

	return refreshedApplicationShortcutsObject
}

func (r *STFWebReceiverResourceModel) RefreshResourcesService(ctx context.Context, diagnostics *diag.Diagnostics, resourcesService *citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel) types.Object {
	refreshedResourcesService := ResourcesService{}
	refreshedResourcesService.PersistentIconCacheEnabled = types.BoolValue(*resourcesService.PersistentIconCacheEnabled.Get())
	refreshedResourcesService.IcaFileCacheExpiry = types.Int64Value(int64(*resourcesService.IcaFileCacheExpiry.Get()))
	refreshedResourcesService.IconSize = types.Int64Value(int64(*resourcesService.IconSize.Get()))
	refreshedResourcesService.ShowDesktopViewer = types.BoolValue(*resourcesService.ShowDesktopViewer.Get())
	refreshedResourcesServiceObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedResourcesService)

	return refreshedResourcesServiceObject
}

func (r *STFWebReceiverResourceModel) RefreshCommunication(ctx context.Context, diagnostics *diag.Diagnostics, communication *citrixstorefront.GetWebReceiverCommunicationResponseModel) types.Object {
	refreshedCommunication := Communication{}
	refreshedCommunication.Attempts = types.Int64Value(int64(*communication.Attempts.Get()))
	refreshedCommunication.Timeout = types.StringValue(communication.Timeout)
	refreshedCommunication.Loopback = types.StringValue(communication.Loopback)
	refreshedCommunication.LoopbackPortUsingHttp = types.Int64Value(int64(*communication.LoopbackPortUsingHttp.Get()))
	refreshedCommunication.ProxyEnabled = types.BoolValue(*communication.Proxy.Enabled.Get())
	refreshedCommunication.ProxyPort = types.Int64Value(int64(*communication.Proxy.Port.Get()))
	refreshedCommunication.ProxyProcessName = types.StringValue(*communication.Proxy.ProcessName.Get())
	refreshedCommunicationObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedCommunication)

	return refreshedCommunicationObject

}

func (r *STFWebReceiverResourceModel) RefreshWebReceiverSiteStyle(ctx context.Context, diagnostics *diag.Diagnostics, ss *citrixstorefront.STFWebReceiverSiteStyleResponseModel) types.Object {
	refreshedSiteStyle := WebReceiverSiteStyle{}
	refreshedSiteStyle.HeaderBackgroundColor = types.StringValue(*ss.HeaderBackgroundColor.Get())
	refreshedSiteStyle.HeaderForegroundColor = types.StringValue(*ss.HeaderForegroundColor.Get())
	refreshedSiteStyle.HeaderLogoPath = types.StringValue(*ss.HeaderLogoPath.Get())
	refreshedSiteStyle.LogonLogoPath = types.StringValue(*ss.LogonLogoPath.Get())
	refreshedSiteStyle.LinkColor = types.StringValue(*ss.LinkColor.Get())
	refreshedSiteStyleObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedSiteStyle)
	return refreshedSiteStyleObject
}

func (r *STFWebReceiverResourceModel) RefreshStrictTransportSecurity(ctx context.Context, diagnostics *diag.Diagnostics, sts *citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel) types.Object {
	refreshedSts := StrictTransportSecurity{}
	refreshedSts.Enabled = types.BoolValue(sts.Enabled)
	refreshedSts.PolicyDuration = types.StringValue(sts.PolicyDuration)
	refreshedStsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedSts)

	return refreshedStsObject
}

func (r *STFWebReceiverResourceModel) RefreshAuthenticationManager(ctx context.Context, diagnostics *diag.Diagnostics, authManager *citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel) types.Object {
	refreshedAuthManager := AuthenticationManager{}
	if authManager.LoginFormTimeout.IsSet() {
		refreshedAuthManager.LoginFormTimeout = types.Int64Value(int64(*authManager.LoginFormTimeout.Get()))
	}
	if authManager.GetUserNameUrl.IsSet() {
		refreshedAuthManager.GetUserNameUrl = types.StringValue(*authManager.GetUserNameUrl.Get())
	}
	if authManager.LogoffUrl.IsSet() {
		refreshedAuthManager.LogoffUrl = types.StringValue(*authManager.LogoffUrl.Get())
	}
	if authManager.ChangeCredentialsUrl.IsSet() {
		refreshedAuthManager.ChangeCredentialsUrl = types.StringValue(*authManager.ChangeCredentialsUrl.Get())
	}
	refreshedAuthManagerObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedAuthManager)

	return refreshedAuthManagerObject
}

func (r *STFWebReceiverResourceModel) RefreshPlugInAssistant(ctx context.Context, diagnostics *diag.Diagnostics, assistant *citrixstorefront.WebReceiverPluginAssistantModel) {
	refreshedPluginAssistant := util.ObjectValueToTypedObject[PluginAssistant](ctx, diagnostics, r.PluginAssistant)
	refreshedPluginAssistant.Enabled = types.BoolValue(*assistant.Enabled.Get())
	refreshedPluginAssistant.UpgradeAtLogin = types.BoolValue(*assistant.UpgradeAtLogin.Get())
	refreshedPluginAssistant.ShowAfterLogin = types.BoolValue(*assistant.ShowAfterLogin.Get())
	if !refreshedPluginAssistant.Win32Path.IsNull() {
		refreshedPluginAssistant.Win32Path = types.StringValue(*assistant.Win32.Path.Get())
	}
	if !refreshedPluginAssistant.MacOSPath.IsNull() {
		refreshedPluginAssistant.MacOSPath = types.StringValue(*assistant.MacOS.Path.Get())
	}
	if !refreshedPluginAssistant.MacOSMinimumSupportedVersion.IsNull() {
		refreshedPluginAssistant.MacOSMinimumSupportedVersion = types.StringValue(*assistant.MacOS.MinimumSupportedVersion.Get())
	}
	if !refreshedPluginAssistant.Html5SingleTabLaunch.IsNull() {
		refreshedPluginAssistant.Html5SingleTabLaunch = types.BoolValue(*assistant.HTML5.SingleTabLaunch.Get())
	}
	if !refreshedPluginAssistant.Html5Enabled.IsNull() {
		if *assistant.HTML5.Enabled.Get() == 0 {
			refreshedPluginAssistant.Html5Enabled = types.StringValue("Off")
		} else if *assistant.HTML5.Enabled.Get() == 1 {
			refreshedPluginAssistant.Html5Enabled = types.StringValue("Always")
		} else {
			refreshedPluginAssistant.Html5Enabled = types.StringValue("Fallback")
		}
	}
	if !refreshedPluginAssistant.Html5Platforms.IsNull() {
		refreshedPluginAssistant.Html5Platforms = types.StringValue(*assistant.HTML5.Platforms.Get())
	}
	if !refreshedPluginAssistant.Html5Preferences.IsNull() {
		refreshedPluginAssistant.Html5Preferences = types.StringValue(*assistant.HTML5.Preferences.Get())
	}
	if !refreshedPluginAssistant.Html5ChromeAppOrigins.IsNull() {
		refreshedPluginAssistant.Html5ChromeAppOrigins = types.StringValue(*assistant.HTML5.ChromeAppOrigins.Get())
	}
	if !refreshedPluginAssistant.Html5ChromeAppPreferences.IsNull() {
		refreshedPluginAssistant.Html5ChromeAppPreferences = types.StringValue(*assistant.HTML5.ChromeAppPreferences.Get())
	}
	if !refreshedPluginAssistant.ProtocolHandlerEnabled.IsNull() {
		refreshedPluginAssistant.ProtocolHandlerEnabled = types.BoolValue(*assistant.ProtocolHandler.Enabled.Get())
	}
	if !refreshedPluginAssistant.ProtocolHandlerPlatforms.IsNull() {
		refreshedPluginAssistant.ProtocolHandlerPlatforms = types.StringValue(*assistant.ProtocolHandler.Platforms.Get())
	}
	refreshedPluginAssistantObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedPluginAssistant)

	r.PluginAssistant = refreshedPluginAssistantObject
}

func (r *STFWebReceiverResourceModel) RefreshUserInterface(ctx context.Context, diagnostics *diag.Diagnostics, ui *citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel) types.Object {
	refreshedUserInterface := util.ObjectValueToTypedObject[UserInterface](ctx, diagnostics, r.UserInterface)
	if ui.AutoLaunchDesktop.IsSet() {
		refreshedUserInterface.AutoLaunchDesktop = types.BoolValue(*ui.AutoLaunchDesktop.Get())
	}

	if ui.MultiClickTimeout.IsSet() {
		refreshedUserInterface.MultiClickTimeout = types.Int64Value(int64(*ui.MultiClickTimeout.Get()))
	}

	if ui.EnableAppsFolderView.IsSet() {
		refreshedUserInterface.EnableAppsFolderView = types.BoolValue(*ui.EnableAppsFolderView.Get())
	}

	if ui.CategoryViewCollapsed.IsSet() {
		refreshedUserInterface.CategoryViewCollapsed = types.BoolValue(*ui.CategoryViewCollapsed.Get())
	}

	if ui.MoveAppToUncategorized.IsSet() {
		refreshedUserInterface.MoveAppToUncategorized = types.BoolValue(*ui.MoveAppToUncategorized.Get())
	}

	if ui.ShowActivityManager.IsSet() {
		refreshedUserInterface.ShowActivityManager = types.BoolValue(*ui.ShowActivityManager.Get())
	}

	if ui.ShowFirstTimeUse.IsSet() {
		refreshedUserInterface.ShowFirstTimeUse = types.BoolValue(*ui.ShowFirstTimeUse.Get())
	}

	if ui.PreventIcaDownloads.IsSet() {
		refreshedUserInterface.PreventIcaDownloads = types.BoolValue(*ui.PreventIcaDownloads.Get())
	} else {
		refreshedUserInterface.PreventIcaDownloads = types.BoolValue(false)
	}

	if !refreshedUserInterface.WorkspaceControl.IsNull() {
		refreshedWorkspaceControl := WorkspaceControl{}
		if ui.WorkspaceControl.Enabled.IsSet() {
			refreshedWorkspaceControl.Enabled = types.BoolValue(*ui.WorkspaceControl.Enabled.Get())
		}

		if ui.WorkspaceControl.AutoReconnectAtLogon.IsSet() {
			refreshedWorkspaceControl.AutoReconnectAtLogon = types.BoolValue(*ui.WorkspaceControl.AutoReconnectAtLogon.Get())
		}

		refreshedWorkspaceControl.LogoffAction = types.StringValue(ui.WorkspaceControl.LogoffAction)

		if ui.WorkspaceControl.ShowReconnectButton.IsSet() {
			refreshedWorkspaceControl.ShowReconnectButton = types.BoolValue(*ui.WorkspaceControl.ShowReconnectButton.Get())
		}

		if ui.WorkspaceControl.ShowDisconnectButton.IsSet() {
			refreshedWorkspaceControl.ShowDisconnectButton = types.BoolValue(*ui.WorkspaceControl.ShowDisconnectButton.Get())
		}

		refreshedUserInterface.WorkspaceControl = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedWorkspaceControl)
	}

	if !refreshedUserInterface.ReceiverConfiguration.IsNull() {
		refreshedReceiverConfiguration := ReceiverConfiguration{}

		if ui.ReceiverConfiguration.Enabled.IsSet() {
			refreshedReceiverConfiguration.Enabled = types.BoolValue(*ui.ReceiverConfiguration.Enabled.Get())
		}

		if ui.ReceiverConfiguration.DownloadUrl.IsSet() {
			refreshedReceiverConfiguration.DownloadUrl = types.StringValue(*ui.ReceiverConfiguration.DownloadUrl.Get())
		}

		refreshedUserInterface.ReceiverConfiguration = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedReceiverConfiguration)
	}

	if !refreshedUserInterface.AppShortcuts.IsNull() {
		refreshedAppShortcuts := AppShortcuts{}

		if ui.AppShortcuts.Enabled.IsSet() {
			refreshedAppShortcuts.Enabled = types.BoolValue(*ui.AppShortcuts.Enabled.Get())
		}
		if ui.AppShortcuts.AllowSessionReconnect.IsSet() {
			refreshedAppShortcuts.AllowSessionReconnect = types.BoolValue(*ui.AppShortcuts.AllowSessionReconnect.Get())
		}

		refreshedUserInterface.AppShortcuts = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedAppShortcuts)
	}

	if !refreshedUserInterface.UIViews.IsNull() {
		refreshedUIViews := UIViews{}

		if ui.UIViews.ShowAppsView.IsSet() {
			refreshedUIViews.ShowAppsView = types.BoolValue(*ui.UIViews.ShowAppsView.Get())
		}
		if ui.UIViews.ShowDesktopsView.IsSet() {
			refreshedUIViews.ShowDesktopsView = types.BoolValue(*ui.UIViews.ShowDesktopsView.Get())
		}
		refreshedUIViews.DefaultView = types.StringValue(ui.UIViews.DefaultView)

		refreshedUserInterface.UIViews = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedUIViews)
	}

	if !refreshedUserInterface.ProgressiveWebApp.IsNull() {
		refreshedProgressiveWebApp := ProgressiveWebApp{}

		if ui.ProgressiveWebApp.Enabled.IsSet() {
			refreshedProgressiveWebApp.Enabled = types.BoolValue(*ui.ProgressiveWebApp.Enabled.Get())
		}
		if ui.ProgressiveWebApp.ShowInstallPrompt.IsSet() {
			refreshedProgressiveWebApp.ShowInstallPrompt = types.BoolValue(*ui.ProgressiveWebApp.ShowInstallPrompt.Get())
		}

		refreshedUserInterface.ProgressiveWebApp = util.TypedObjectToObjectValue(ctx, diagnostics, refreshedProgressiveWebApp)
	}

	refreshedUserInterfaceObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedUserInterface)

	return refreshedUserInterfaceObject
}

func (STFWebReceiverResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront WebReceiver.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront WebReceiver. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_path": schema.StringAttribute{
				Description: "The IIS VirtualPath at which the WebReceiver will be configured to be accessed by Receivers.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name of the WebReceiver",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store_virtual_path": schema.StringAttribute{
				Description: "The Virtual Path of the StoreFront Store Service linked to the WebReceiver.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"authentication_methods": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The authentication methods supported by the WebReceiver.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetNull(types.StringType)),
			},
			"plugin_assistant":          PluginAssistant{}.GetSchema(),
			"application_shortcuts":     ApplicationShortcuts{}.GetSchema(),
			"communication":             Communication{}.GetSchema(),
			"strict_transport_security": StrictTransportSecurity{}.GetSchema(),
			"authentication_manager":    AuthenticationManager{}.GetSchema(),
			"user_interface":            UserInterface{}.GetSchema(),
			"resources_service":         ResourcesService{}.GetSchema(),
			"web_receiver_site_style":   WebReceiverSiteStyle{}.GetSchema(),
		},
	}
}

func (STFWebReceiverResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFWebReceiverResourceModel{}.GetSchema().Attributes
}
