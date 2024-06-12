// Copyright © 2024. Citrix Systems, Inc.

package stf_webreceiver

import (
	"context"
	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

// SFWebReceiverResourceModel maps the resource schema data.
type STFWebReceiverResourceModel struct {
	VirtualPath           types.String `tfsdk:"virtual_path"`
	SiteId                types.String `tfsdk:"site_id"`
	FriendlyName          types.String `tfsdk:"friendly_name"`
	StoreService          types.String `tfsdk:"store_service"`
	PluginAssistant       types.Object `tfsdk:"plugin_assistant"`       // PluginAssistant
	AuthenticationMethods types.Set    `tfsdk:"authentication_methods"` // Set[string]
}

// Schema defines the schema for the resource.
func (r *stfWebReceiverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "StoreFront WebReceiver.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront webreceiver. Defaults to 1.",
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
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name of the WebReceiver",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store_service": schema.StringAttribute{
				Description: "The StoreFront Store Service linked to the WebReceiver.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_methods": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The authentication methods supported by the WebReceiver.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetNull(types.StringType)),
			},
			"plugin_assistant": PluginAssistant{}.GetSchema(),
		},
	}
}

func (r *STFWebReceiverResourceModel) RefreshPropertyValues(webreceiver *citrixstorefront.STFWebReceiverDetailModel) {
	// Overwrite SFWebReceiverResourceModel with refreshed state
	r.VirtualPath = types.StringValue(*webreceiver.VirtualPath.Get())
	r.SiteId = types.StringValue(strconv.Itoa(*webreceiver.SiteId.Get()))
	r.FriendlyName = types.StringValue(*webreceiver.FriendlyName.Get())

}

func (r *STFWebReceiverResourceModel) RefreshPlugInAssistant(ctx context.Context, diagnostics *diag.Diagnostics, assistant *citrixstorefront.WebReceiverPluginAssistantModel) {
	// Overwrite SFWebReceiverResourceModel with refreshed state
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
