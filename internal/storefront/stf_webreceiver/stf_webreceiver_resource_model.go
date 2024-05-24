// Copyright © 2023. Citrix Systems, Inc.

package stf_webreceiver

import (
	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SFWebReceiverResourceModel maps the resource schema data.
type STFWebReceiverResourceModel struct {
	VirtualPath           types.String     `tfsdk:"virtual_path"`
	SiteId                types.String     `tfsdk:"site_id"`
	FriendlyName          types.String     `tfsdk:"friendly_name"`
	StoreService          types.String     `tfsdk:"store_service"`
	PluginAssistant       *PluginAssistant `tfsdk:"plugin_assistant"`
	AuthenticationMethods []types.String   `tfsdk:"authentication_methods"`
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

func (r *STFWebReceiverResourceModel) RefreshPropertyValues(webreceiver *citrixstorefront.STFWebReceiverDetailModel) {
	// Overwrite SFWebReceiverResourceModel with refreshed state
	r.VirtualPath = types.StringValue(*webreceiver.VirtualPath.Get())
	r.SiteId = types.StringValue(strconv.Itoa(*webreceiver.SiteId.Get()))
	r.FriendlyName = types.StringValue(*webreceiver.FriendlyName.Get())

}

func (r *STFWebReceiverResourceModel) RefreshPlugInAssistant(assistant *citrixstorefront.WebReceiverPluginAssistantModel) {
	// Overwrite SFWebReceiverResourceModel with refreshed state
	r.PluginAssistant.Enabled = types.BoolValue(*assistant.Enabled.Get())
	r.PluginAssistant.UpgradeAtLogin = types.BoolValue(*assistant.UpgradeAtLogin.Get())
	r.PluginAssistant.ShowAfterLogin = types.BoolValue(*assistant.ShowAfterLogin.Get())
	if !r.PluginAssistant.Win32Path.IsNull() {
		r.PluginAssistant.Win32Path = types.StringValue(*assistant.Win32.Path.Get())
	}
	if !r.PluginAssistant.MacOSPath.IsNull() {
		r.PluginAssistant.MacOSPath = types.StringValue(*assistant.MacOS.Path.Get())
	}
	if !r.PluginAssistant.MacOSMinimumSupportedVersion.IsNull() {
		r.PluginAssistant.MacOSMinimumSupportedVersion = types.StringValue(*assistant.MacOS.MinimumSupportedVersion.Get())
	}
	if !r.PluginAssistant.Html5SingleTabLaunch.IsNull() {
		r.PluginAssistant.Html5SingleTabLaunch = types.BoolValue(*assistant.HTML5.SingleTabLaunch.Get())
	}
	if !r.PluginAssistant.Html5Enabled.IsNull() {
		if *assistant.HTML5.Enabled.Get() == 0 {
			r.PluginAssistant.Html5Enabled = types.StringValue("Off")
		} else if *assistant.HTML5.Enabled.Get() == 1 {
			r.PluginAssistant.Html5Enabled = types.StringValue("Always")
		} else {
			r.PluginAssistant.Html5Enabled = types.StringValue("Fallback")
		}
	}
	if !r.PluginAssistant.Html5Platforms.IsNull() {
		r.PluginAssistant.Html5Platforms = types.StringValue(*assistant.HTML5.Platforms.Get())
	}
	if !r.PluginAssistant.Html5Preferences.IsNull() {
		r.PluginAssistant.Html5Preferences = types.StringValue(*assistant.HTML5.Preferences.Get())
	}
	if !r.PluginAssistant.Html5ChromeAppOrigins.IsNull() {
		r.PluginAssistant.Html5ChromeAppOrigins = types.StringValue(*assistant.HTML5.ChromeAppOrigins.Get())
	}
	if !r.PluginAssistant.Html5ChromeAppPreferences.IsNull() {
		r.PluginAssistant.Html5ChromeAppPreferences = types.StringValue(*assistant.HTML5.ChromeAppPreferences.Get())
	}
	if !r.PluginAssistant.ProtocolHandlerEnabled.IsNull() {
		r.PluginAssistant.ProtocolHandlerEnabled = types.BoolValue(*assistant.ProtocolHandler.Enabled.Get())
	}
	if !r.PluginAssistant.ProtocolHandlerPlatforms.IsNull() {
		r.PluginAssistant.ProtocolHandlerPlatforms = types.StringValue(*assistant.ProtocolHandler.Platforms.Get())
	}

}
