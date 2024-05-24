// Copyright © 2023. Citrix Systems, Inc.

package gac_settings

import (
	"fmt"
	"reflect"

	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GACSettingsResourceModel struct {
	ServiceUrl      types.String `tfsdk:"service_url"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	UseForAppConfig types.Bool   `tfsdk:"use_for_app_config"`
	AppSettings     *AppSettings `tfsdk:"app_settings"`
}

type AppSettings struct {
	Windows  []Windows  `tfsdk:"windows"`
	Ios      []Ios      `tfsdk:"ios"`
	Android  []Android  `tfsdk:"android"`
	Chromeos []Chromeos `tfsdk:"chromeos"`
	Html5    []Html5    `tfsdk:"html5"`
	Macos    []Macos    `tfsdk:"macos"`
}

type Windows struct {
	Category     types.String      `tfsdk:"category"`
	UserOverride types.Bool        `tfsdk:"user_override"`
	Settings     []WindowsSettings `tfsdk:"settings"`
}

type Ios struct {
	Category     types.String  `tfsdk:"category"`
	UserOverride types.Bool    `tfsdk:"user_override"`
	Settings     []IosSettings `tfsdk:"settings"`
}

type Android struct {
	Category     types.String      `tfsdk:"category"`
	UserOverride types.Bool        `tfsdk:"user_override"`
	Settings     []AndroidSettings `tfsdk:"settings"`
}

type Chromeos struct {
	Category     types.String       `tfsdk:"category"`
	UserOverride types.Bool         `tfsdk:"user_override"`
	Settings     []ChromeosSettings `tfsdk:"settings"`
}

type Html5 struct {
	Category     types.String    `tfsdk:"category"`
	UserOverride types.Bool      `tfsdk:"user_override"`
	Settings     []Html5Settings `tfsdk:"settings"`
}

type Macos struct {
	Category     types.String    `tfsdk:"category"`
	UserOverride types.Bool      `tfsdk:"user_override"`
	Settings     []MacosSettings `tfsdk:"settings"`
}

type WindowsSettings struct {
	Name        types.String   `tfsdk:"name"`
	ValueString types.String   `tfsdk:"value_string"`
	ValueList   []types.String `tfsdk:"value_list"`
}

type IosSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
}

type AndroidSettings struct {
	Name        types.String   `tfsdk:"name"`
	ValueString types.String   `tfsdk:"value_string"`
	ValueList   []types.String `tfsdk:"value_list"`
}

type ChromeosSettings struct {
	Name        types.String   `tfsdk:"name"`
	ValueString types.String   `tfsdk:"value_string"`
	ValueList   []types.String `tfsdk:"value_list"`
}

type Html5Settings struct {
	Name        types.String   `tfsdk:"name"`
	ValueString types.String   `tfsdk:"value_string"`
	ValueList   []types.String `tfsdk:"value_list"`
}

type MacosSettings struct {
	Name        types.String   `tfsdk:"name"`
	ValueString types.String   `tfsdk:"value_string"`
	ValueList   []types.String `tfsdk:"value_list"`
}

func (r GACSettingsResourceModel) RefreshPropertyValues(settingsRecordModel globalappconfiguration.SettingsRecordModel, diagnostics *diag.Diagnostics) GACSettingsResourceModel {

	var serviceUrlModel = settingsRecordModel.GetServiceURL()
	r.ServiceUrl = types.StringValue(serviceUrlModel.GetUrl())

	var settings = settingsRecordModel.GetSettings()
	r.Name = types.StringValue(settings.GetName())
	r.Description = types.StringValue(settings.GetDescription())
	r.UseForAppConfig = types.BoolValue(settings.GetUseForAppConfig())

	var appSettings = settings.GetAppSettings()
	var windowsSettings = appSettings.GetWindows()
	var iosSettings = appSettings.GetIos()
	var androidSettings = appSettings.GetAndroid()
	var chromeosSettings = appSettings.GetChromeos()
	var html5Settings = appSettings.GetHtml5()
	var macosSettings = appSettings.GetMacos()

	if r.AppSettings == nil {
		r.AppSettings = &AppSettings{}
	}
	r.AppSettings.Windows = r.getWindowsSettings(windowsSettings, diagnostics)
	r.AppSettings.Ios = r.getIosSettings(iosSettings, diagnostics)
	r.AppSettings.Android = r.getAndroidSettings(androidSettings, diagnostics)
	r.AppSettings.Chromeos = r.getChromeosSettings(chromeosSettings, diagnostics)
	r.AppSettings.Html5 = r.getHtml5Settings(html5Settings, diagnostics)
	r.AppSettings.Macos = r.getMacosSettings(macosSettings, diagnostics)

	return r
}

func (r GACSettingsResourceModel) getWindowsSettings(remoteWindowsSettings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Windows {
	var stateWindowsSettings []Windows
	if r.AppSettings != nil && r.AppSettings.Windows != nil {
		stateWindowsSettings = r.AppSettings.Windows
	}

	type RemoteWindowsSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteWindowsSettingsTracker for remote
	remoteWindowsSettingsMap := map[string]*RemoteWindowsSettingsTracker{}
	for _, remoteWindowsSetting := range remoteWindowsSettings {
		remoteWindowsSettingsMap[remoteWindowsSetting.GetCategory()] = &RemoteWindowsSettingsTracker{
			platformSetting: remoteWindowsSetting,
			IsVisited:       false,
		}
	}

	// Prepare the windows settings list to be stored in the state
	var windowsSettingsForState []Windows
	for _, stateWindowsSetting := range stateWindowsSettings {
		remoteWindowsSetting, exists := remoteWindowsSettingsMap[stateWindowsSetting.Category.ValueString()]
		if !exists {
			// If windows setting is not present in the remote, then don't add it to the state
			continue
		}

		windowsSettingsForState = append(windowsSettingsForState, Windows{
			Category:     types.StringValue(remoteWindowsSetting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteWindowsSetting.platformSetting.GetUserOverride()),
			Settings:     getWindowsCategorySettings(stateWindowsSetting.Settings, remoteWindowsSetting.platformSetting.GetSettings(), diagnostics),
		})

		remoteWindowsSetting.IsVisited = true

	}

	// Add the windows settings from remote which are not present in the state
	for _, remoteWindowsSetting := range remoteWindowsSettingsMap {
		if !remoteWindowsSetting.IsVisited {
			windowsSettingsForState = append(windowsSettingsForState, Windows{
				Category:     types.StringValue(remoteWindowsSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteWindowsSetting.platformSetting.GetUserOverride()),
				Settings:     parseWindowsSettings(remoteWindowsSetting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return windowsSettingsForState
}

func (r GACSettingsResourceModel) getIosSettings(remoteIosSettings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Ios {
	var stateIosSettings []Ios
	if r.AppSettings != nil && r.AppSettings.Ios != nil {
		stateIosSettings = r.AppSettings.Ios
	}

	type RemoteIosSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteIosSettingsTracker for remote
	remoteIosSettingsMap := map[string]*RemoteIosSettingsTracker{}
	for _, remoteIosSetting := range remoteIosSettings {
		remoteIosSettingsMap[remoteIosSetting.GetCategory()] = &RemoteIosSettingsTracker{
			platformSetting: remoteIosSetting,
			IsVisited:       false,
		}
	}

	// Prepare the ios settings list to be stored in the state
	var iosSettingsForState []Ios
	for _, stateIosSetting := range stateIosSettings {
		remoteIosSetting, exists := remoteIosSettingsMap[stateIosSetting.Category.ValueString()]
		if !exists {
			// If ios setting is not present in the remote, then don't add it to the state
			continue
		}

		iosSettingsForState = append(iosSettingsForState, Ios{
			Category:     types.StringValue(remoteIosSetting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteIosSetting.platformSetting.GetUserOverride()),
			Settings:     getIosCategorySettings(stateIosSetting.Settings, remoteIosSetting.platformSetting.GetSettings(), diagnostics),
		})

		remoteIosSetting.IsVisited = true

	}

	// Add the ios settings from remote which are not present in the state
	for _, remoteIosSetting := range remoteIosSettingsMap {
		if !remoteIosSetting.IsVisited {
			iosSettingsForState = append(iosSettingsForState, Ios{
				Category:     types.StringValue(remoteIosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteIosSetting.platformSetting.GetUserOverride()),
				Settings:     parseIosSettings(remoteIosSetting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return iosSettingsForState
}

func (r GACSettingsResourceModel) getAndroidSettings(remoteAndroidSettings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Android {
	var stateAndroidSettings []Android
	if r.AppSettings != nil && r.AppSettings.Android != nil {
		stateAndroidSettings = r.AppSettings.Android
	}

	type RemoteAndroidSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteAndroidSettingsTracker for remote
	remoteAndroidSettingsMap := map[string]*RemoteAndroidSettingsTracker{}
	for _, remoteAndroidSetting := range remoteAndroidSettings {
		remoteAndroidSettingsMap[remoteAndroidSetting.GetCategory()] = &RemoteAndroidSettingsTracker{
			platformSetting: remoteAndroidSetting,
			IsVisited:       false,
		}
	}

	// Prepare the android settings list to be stored in the state
	var androidSettingsForState []Android
	for _, stateAndroidSetting := range stateAndroidSettings {
		remoteAndroidSetting, exists := remoteAndroidSettingsMap[stateAndroidSetting.Category.ValueString()]
		if !exists {
			// If android setting is not present in the remote, then don't add it to the state
			continue
		}

		androidSettingsForState = append(androidSettingsForState, Android{
			Category:     types.StringValue(remoteAndroidSetting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteAndroidSetting.platformSetting.GetUserOverride()),
			Settings:     getAndroidCategorySettings(stateAndroidSetting.Settings, remoteAndroidSetting.platformSetting.GetSettings(), diagnostics),
		})

		remoteAndroidSetting.IsVisited = true

	}

	// Add the android settings from remote which are not present in the state
	for _, remoteAndroidSetting := range remoteAndroidSettingsMap {
		if !remoteAndroidSetting.IsVisited {
			androidSettingsForState = append(androidSettingsForState, Android{
				Category:     types.StringValue(remoteAndroidSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteAndroidSetting.platformSetting.GetUserOverride()),
				Settings:     parseAndroidSettings(remoteAndroidSetting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return androidSettingsForState
}

func (r GACSettingsResourceModel) getHtml5Settings(remoteHtml5Settings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Html5 {
	var stateHtml5Settings []Html5
	if r.AppSettings != nil && r.AppSettings.Html5 != nil {
		stateHtml5Settings = r.AppSettings.Html5
	}

	type RemoteHtml5SettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteHtml5SettingsTracker for remote
	remoteHtml5SettingsMap := map[string]*RemoteHtml5SettingsTracker{}
	for _, remoteHtml5Setting := range remoteHtml5Settings {
		remoteHtml5SettingsMap[remoteHtml5Setting.GetCategory()] = &RemoteHtml5SettingsTracker{
			platformSetting: remoteHtml5Setting,
			IsVisited:       false,
		}
	}

	// Prepare the html5 settings list to be stored in the state
	var html5SettingsForState []Html5
	for _, stateHtml5Setting := range stateHtml5Settings {
		remoteHtml5Setting, exists := remoteHtml5SettingsMap[stateHtml5Setting.Category.ValueString()]
		if !exists {
			// If html5 setting is not present in the remote, then don't add it to the state
			continue
		}

		html5SettingsForState = append(html5SettingsForState, Html5{
			Category:     types.StringValue(remoteHtml5Setting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteHtml5Setting.platformSetting.GetUserOverride()),
			Settings:     getHtml5CategorySettings(stateHtml5Setting.Settings, remoteHtml5Setting.platformSetting.GetSettings(), diagnostics),
		})

		remoteHtml5Setting.IsVisited = true

	}

	// Add the html5 settings from remote which are not present in the state
	for _, remoteHtml5Setting := range remoteHtml5SettingsMap {
		if !remoteHtml5Setting.IsVisited {
			html5SettingsForState = append(html5SettingsForState, Html5{
				Category:     types.StringValue(remoteHtml5Setting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteHtml5Setting.platformSetting.GetUserOverride()),
				Settings:     parseHtml5Settings(remoteHtml5Setting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return html5SettingsForState
}

func (r GACSettingsResourceModel) getChromeosSettings(remoteChromeosSettings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Chromeos {
	var stateChromeosSettings []Chromeos
	if r.AppSettings != nil && r.AppSettings.Chromeos != nil {
		stateChromeosSettings = r.AppSettings.Chromeos
	}

	type RemoteChromeosSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteChromeosSettingsTracker for remote
	remoteChromeosSettingsMap := map[string]*RemoteChromeosSettingsTracker{}
	for _, remoteChromeosSetting := range remoteChromeosSettings {
		remoteChromeosSettingsMap[remoteChromeosSetting.GetCategory()] = &RemoteChromeosSettingsTracker{
			platformSetting: remoteChromeosSetting,
			IsVisited:       false,
		}
	}

	// Prepare the chromeos settings list to be stored in the state
	var chromeosSettingsForState []Chromeos
	for _, stateChromeosSetting := range stateChromeosSettings {
		remoteChromeosSetting, exists := remoteChromeosSettingsMap[stateChromeosSetting.Category.ValueString()]
		if !exists {
			// If chromeos setting is not present in the remote, then don't add it to the state
			continue
		}

		chromeosSettingsForState = append(chromeosSettingsForState, Chromeos{
			Category:     types.StringValue(remoteChromeosSetting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteChromeosSetting.platformSetting.GetUserOverride()),
			Settings:     getChromeosCategorySettings(stateChromeosSetting.Settings, remoteChromeosSetting.platformSetting.GetSettings(), diagnostics),
		})

		remoteChromeosSetting.IsVisited = true

	}

	// Add the chromeos settings from remote which are not present in the state
	for _, remoteChromeosSetting := range remoteChromeosSettingsMap {
		if !remoteChromeosSetting.IsVisited {
			chromeosSettingsForState = append(chromeosSettingsForState, Chromeos{
				Category:     types.StringValue(remoteChromeosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteChromeosSetting.platformSetting.GetUserOverride()),
				Settings:     parseChromeosSettings(remoteChromeosSetting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return chromeosSettingsForState
}

func (r GACSettingsResourceModel) getMacosSettings(remoteMacosSettings []globalappconfiguration.PlatformSettings, diagnostics *diag.Diagnostics) []Macos {
	var stateMacosSettings []Macos
	if r.AppSettings != nil && r.AppSettings.Macos != nil {
		stateMacosSettings = r.AppSettings.Macos
	}

	type RemoteMacosSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteMacosSettingsTracker for remote
	remoteMacosSettingsMap := map[string]*RemoteMacosSettingsTracker{}
	for _, remoteMacosSetting := range remoteMacosSettings {
		remoteMacosSettingsMap[remoteMacosSetting.GetCategory()] = &RemoteMacosSettingsTracker{
			platformSetting: remoteMacosSetting,
			IsVisited:       false,
		}
	}

	// Prepare the macos settings list to be stored in the state
	var macosSettingsForState []Macos
	for _, stateMacosSetting := range stateMacosSettings {
		remoteMacosSetting, exists := remoteMacosSettingsMap[stateMacosSetting.Category.ValueString()]
		if !exists {
			// If macos setting is not present in the remote, then don't add it to the state
			continue
		}

		macosSettingsForState = append(macosSettingsForState, Macos{
			Category:     types.StringValue(remoteMacosSetting.platformSetting.GetCategory()),
			UserOverride: types.BoolValue(remoteMacosSetting.platformSetting.GetUserOverride()),
			Settings:     getMacosCategorySettings(stateMacosSetting.Settings, remoteMacosSetting.platformSetting.GetSettings(), diagnostics),
		})

		remoteMacosSetting.IsVisited = true

	}

	// Add the macos settings from remote which are not present in the state
	for _, remoteMacosSetting := range remoteMacosSettingsMap {
		if !remoteMacosSetting.IsVisited {
			macosSettingsForState = append(macosSettingsForState, Macos{
				Category:     types.StringValue(remoteMacosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteMacosSetting.platformSetting.GetUserOverride()),
				Settings:     parseMacosSettings(remoteMacosSetting.platformSetting.GetSettings(), diagnostics),
			})
		}
	}

	return macosSettingsForState
}

func parseWindowsSettings(remoteWindowsSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []WindowsSettings {
	var windowsSettings []WindowsSettings
	var errMsg string

	for _, remoteWindowsSetting := range remoteWindowsSettings {
		var windowsSetting WindowsSettings
		windowsSetting.Name = types.StringValue(remoteWindowsSetting.GetName())
		valueType := reflect.TypeOf(remoteWindowsSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.GetValue().(string))
		case reflect.Slice:
			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteWindowsSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for windows setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+windowsSetting.Name.ValueString(),
				errMsg,
			)
		}

		windowsSettings = append(windowsSettings, windowsSetting)
	}

	return windowsSettings
}

func parseIosSettings(remoteIosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []IosSettings {
	var iosSettings []IosSettings
	var errMsg string

	for _, remoteIosSetting := range remoteIosSettings {
		var iosSetting IosSettings
		iosSetting.Name = types.StringValue(remoteIosSetting.GetName())
		valueType := reflect.TypeOf(remoteIosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			iosSetting.ValueString = types.StringValue(remoteIosSetting.GetValue().(string))
		default:
			errMsg = fmt.Sprintf("Unsupported type for ios setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+iosSetting.Name.ValueString(),
				errMsg,
			)
		}

		iosSettings = append(iosSettings, iosSetting)
	}

	return iosSettings
}

func parseAndroidSettings(remoteAndroidSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []AndroidSettings {
	var androidSettings []AndroidSettings
	var errMsg string

	for _, remoteAndroidSetting := range remoteAndroidSettings {
		var androidSetting AndroidSettings
		androidSetting.Name = types.StringValue(remoteAndroidSetting.GetName())
		valueType := reflect.TypeOf(remoteAndroidSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			androidSetting.ValueString = types.StringValue(remoteAndroidSetting.GetValue().(string))
		case reflect.Slice:
			androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteAndroidSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for android setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+androidSetting.Name.ValueString(),
				errMsg,
			)
		}

		androidSettings = append(androidSettings, androidSetting)
	}

	return androidSettings
}

func parseHtml5Settings(remoteHtml5Settings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []Html5Settings {
	var html5Settings []Html5Settings
	var errMsg string

	for _, remoteHtml5Setting := range remoteHtml5Settings {
		var html5Setting Html5Settings
		html5Setting.Name = types.StringValue(remoteHtml5Setting.GetName())
		valueType := reflect.TypeOf(remoteHtml5Setting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			html5Setting.ValueString = types.StringValue(remoteHtml5Setting.GetValue().(string))
		case reflect.Slice:
			html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteHtml5Setting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for html5 setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+html5Setting.Name.ValueString(),
				errMsg,
			)
		}

		html5Settings = append(html5Settings, html5Setting)
	}

	return html5Settings
}

func parseMacosSettings(remoteMacosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []MacosSettings {
	var macosSettings []MacosSettings
	var errMsg string

	for _, remoteMacosSetting := range remoteMacosSettings {
		var macosSetting MacosSettings
		macosSetting.Name = types.StringValue(remoteMacosSetting.GetName())
		valueType := reflect.TypeOf(remoteMacosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.GetValue().(string))
		case reflect.Slice:
			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteMacosSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for macos setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+macosSetting.Name.ValueString(),
				errMsg,
			)
		}

		macosSettings = append(macosSettings, macosSetting)
	}

	return macosSettings
}

func parseChromeosSettings(remoteChromeosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []ChromeosSettings {
	var chromeosSettings []ChromeosSettings
	var errMsg string

	for _, remoteChromeosSetting := range remoteChromeosSettings {
		var chromeosSetting ChromeosSettings
		chromeosSetting.Name = types.StringValue(remoteChromeosSetting.GetName())
		valueType := reflect.TypeOf(remoteChromeosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.GetValue().(string))
		case reflect.Slice:
			chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteChromeosSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for chrome os setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+chromeosSetting.Name.ValueString(),
				errMsg,
			)
		}

		chromeosSettings = append(chromeosSettings, chromeosSetting)
	}

	return chromeosSettings
}

func getWindowsCategorySettings(stateWindowsSettings []WindowsSettings, remoteWindowsSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []WindowsSettings {

	var windowsSettingsForState []WindowsSettings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteWindowsCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteWindowsSetting := range remoteWindowsSettings {
		remoteWindowsCategorySettingsMap[remoteWindowsSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteWindowsSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the windows settings list to be stored in the state
	for _, stateWindowsSetting := range stateWindowsSettings {
		remoteWindowsSetting, exists := remoteWindowsCategorySettingsMap[stateWindowsSetting.Name.ValueString()]
		if !exists {
			// If windows setting is not present in the remote, then don't add it to the state
			continue
		}

		var windowsSetting WindowsSettings
		errMsg = ""

		windowsSetting.Name = stateWindowsSetting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteWindowsSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
		case reflect.Slice:
			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteWindowsSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for windows setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+windowsSetting.Name.ValueString(),
				errMsg,
			)
		}

		windowsSettingsForState = append(windowsSettingsForState, windowsSetting)
		remoteWindowsSetting.IsVisited = true
	}

	// Add the windows settings from remote which are not present in the state
	for settingName, remoteWindowsSetting := range remoteWindowsCategorySettingsMap {
		if !remoteWindowsSetting.IsVisited {
			var windowsSetting WindowsSettings
			errMsg = ""
			windowsSetting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteWindowsSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
			case reflect.Slice:
				windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteWindowsSetting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for windows setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+windowsSetting.Name.ValueString(),
					errMsg,
				)
			}

			windowsSettingsForState = append(windowsSettingsForState, windowsSetting)
		}
	}

	return windowsSettingsForState
}

func getIosCategorySettings(stateIosSettings []IosSettings, remoteIosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []IosSettings {

	var iosSettingsForState []IosSettings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteIosCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteIosSetting := range remoteIosSettings {
		remoteIosCategorySettingsMap[remoteIosSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteIosSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the ios settings list to be stored in the state
	for _, stateIosSetting := range stateIosSettings {
		remoteIosSetting, exists := remoteIosCategorySettingsMap[stateIosSetting.Name.ValueString()]
		if !exists {
			// If ios setting is not present in the remote, then don't add it to the state
			continue
		}

		var iosSetting IosSettings
		errMsg = ""

		iosSetting.Name = stateIosSetting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteIosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			iosSetting.ValueString = types.StringValue(remoteIosSetting.Value.(string))
		default:
			errMsg = fmt.Sprintf("Unsupported type for ios setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+iosSetting.Name.ValueString(),
				errMsg,
			)
		}

		iosSettingsForState = append(iosSettingsForState, iosSetting)
		remoteIosSetting.IsVisited = true
	}

	// Add the ios settings from remote which are not present in the state
	for settingName, remoteIosSetting := range remoteIosCategorySettingsMap {
		if !remoteIosSetting.IsVisited {
			var iosSetting IosSettings
			errMsg = ""
			iosSetting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteIosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				iosSetting.ValueString = types.StringValue(remoteIosSetting.Value.(string))
			default:
				fmt.Errorf("Unsupported type for ios setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+iosSetting.Name.ValueString(),
					errMsg,
				)
			}

			iosSettingsForState = append(iosSettingsForState, iosSetting)
		}
	}

	return iosSettingsForState
}

func getAndroidCategorySettings(stateAndroidSettings []AndroidSettings, remoteAndroidSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []AndroidSettings {

	var androidSettingsForState []AndroidSettings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteAndroidCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteAndroidSetting := range remoteAndroidSettings {
		remoteAndroidCategorySettingsMap[remoteAndroidSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteAndroidSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the android settings list to be stored in the state
	for _, stateAndroidSetting := range stateAndroidSettings {
		remoteAndroidSetting, exists := remoteAndroidCategorySettingsMap[stateAndroidSetting.Name.ValueString()]
		if !exists {
			// If android setting is not present in the remote, then don't add it to the state
			continue
		}

		var androidSetting AndroidSettings
		androidSetting.Name = stateAndroidSetting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteAndroidSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			androidSetting.ValueString = types.StringValue(remoteAndroidSetting.Value.(string))
		case reflect.Slice:
			androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteAndroidSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for android setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+androidSetting.Name.ValueString(),
				errMsg,
			)
		}

		androidSettingsForState = append(androidSettingsForState, androidSetting)
		remoteAndroidSetting.IsVisited = true
	}

	// Add the android settings from remote which are not present in the state
	for settingName, remoteAndroidSetting := range remoteAndroidCategorySettingsMap {
		if !remoteAndroidSetting.IsVisited {
			var androidSetting AndroidSettings
			errMsg = ""
			androidSetting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteAndroidSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				androidSetting.ValueString = types.StringValue(remoteAndroidSetting.Value.(string))
			case reflect.Slice:
				androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteAndroidSetting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for android setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+androidSetting.Name.ValueString(),
					errMsg,
				)
			}

			androidSettingsForState = append(androidSettingsForState, androidSetting)
		}
	}

	return androidSettingsForState
}

func getChromeosCategorySettings(stateChromeosSettings []ChromeosSettings, remoteChromeosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []ChromeosSettings {

	var chromeosSettingsForState []ChromeosSettings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteChromeosCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteChromeosSetting := range remoteChromeosSettings {
		remoteChromeosCategorySettingsMap[remoteChromeosSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteChromeosSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the chromeos settings list to be stored in the state
	for _, stateChromeosSetting := range stateChromeosSettings {
		remoteChromeosSetting, exists := remoteChromeosCategorySettingsMap[stateChromeosSetting.Name.ValueString()]
		if !exists {
			// If chromeos setting is not present in the remote, then don't add it to the state
			continue
		}

		var chromeosSetting ChromeosSettings
		errMsg = ""

		chromeosSetting.Name = stateChromeosSetting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteChromeosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.Value.(string))
		case reflect.Slice:
			chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteChromeosSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for chromeos setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+chromeosSetting.Name.ValueString(),
				errMsg,
			)
		}

		chromeosSettingsForState = append(chromeosSettingsForState, chromeosSetting)
		remoteChromeosSetting.IsVisited = true
	}

	// Add the chromeos settings from remote which are not present in the state
	for settingName, remoteChromeosSetting := range remoteChromeosCategorySettingsMap {
		if !remoteChromeosSetting.IsVisited {
			var chromeosSetting ChromeosSettings
			errMsg = ""

			chromeosSetting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteChromeosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.Value.(string))
			case reflect.Slice:
				chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteChromeosSetting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for chromeos setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+chromeosSetting.Name.ValueString(),
					errMsg,
				)
			}

			chromeosSettingsForState = append(chromeosSettingsForState, chromeosSetting)
		}
	}

	return chromeosSettingsForState
}

func getHtml5CategorySettings(stateHtml5Settings []Html5Settings, remoteHtml5Settings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []Html5Settings {

	var html5SettingsForState []Html5Settings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteHtml5CategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteHtml5Setting := range remoteHtml5Settings {
		remoteHtml5CategorySettingsMap[remoteHtml5Setting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteHtml5Setting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the html5 settings list to be stored in the state
	for _, stateHtml5Setting := range stateHtml5Settings {
		remoteHtml5Setting, exists := remoteHtml5CategorySettingsMap[stateHtml5Setting.Name.ValueString()]
		if !exists {
			// If html5 setting is not present in the remote, then don't add it to the state
			continue
		}

		var html5Setting Html5Settings
		errMsg = ""

		html5Setting.Name = stateHtml5Setting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteHtml5Setting.Value)
		switch valueType.Kind() {
		case reflect.String:
			html5Setting.ValueString = types.StringValue(remoteHtml5Setting.Value.(string))
		case reflect.Slice:
			html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteHtml5Setting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for html5 setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+html5Setting.Name.ValueString(),
				errMsg,
			)
		}

		html5SettingsForState = append(html5SettingsForState, html5Setting)
		remoteHtml5Setting.IsVisited = true
	}

	// Add the html5 settings from remote which are not present in the state
	for settingName, remoteHtml5Setting := range remoteHtml5CategorySettingsMap {
		if !remoteHtml5Setting.IsVisited {
			var html5Setting Html5Settings
			errMsg = ""

			html5Setting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteHtml5Setting.Value)
			switch valueType.Kind() {
			case reflect.String:
				html5Setting.ValueString = types.StringValue(remoteHtml5Setting.Value.(string))
			case reflect.Slice:
				html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteHtml5Setting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for html5 setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+html5Setting.Name.ValueString(),
					errMsg,
				)
			}

			html5SettingsForState = append(html5SettingsForState, html5Setting)
		}
	}

	return html5SettingsForState
}

func getMacosCategorySettings(stateMacosSettings []MacosSettings, remoteMacosSettings []globalappconfiguration.CategorySettings, diagnostics *diag.Diagnostics) []MacosSettings {

	var macosSettingsForState []MacosSettings
	var errMsg string

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteMacosCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteMacosSetting := range remoteMacosSettings {
		remoteMacosCategorySettingsMap[remoteMacosSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteMacosSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the macos settings list to be stored in the state
	for _, stateMacosSetting := range stateMacosSettings {
		remoteMacosSetting, exists := remoteMacosCategorySettingsMap[stateMacosSetting.Name.ValueString()]
		if !exists {
			// If macos setting is not present in the remote, then don't add it to the state
			continue
		}

		var macosSetting MacosSettings
		errMsg = ""

		macosSetting.Name = stateMacosSetting.Name // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteMacosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
		case reflect.Slice:
			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteMacosSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for macos setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+macosSetting.Name.ValueString(),
				errMsg,
			)
		}

		macosSettingsForState = append(macosSettingsForState, macosSetting)
		remoteMacosSetting.IsVisited = true
	}

	// Add the macos settings from remote which are not present in the state
	for settingName, remoteMacosSetting := range remoteMacosCategorySettingsMap {
		if !remoteMacosSetting.IsVisited {
			var macosSetting MacosSettings
			errMsg = ""

			macosSetting.Name = types.StringValue(settingName)
			valueType := reflect.TypeOf(remoteMacosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
			case reflect.Slice:
				macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToBaseStringArray(remoteMacosSetting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for macos setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+macosSetting.Name.ValueString(),
					errMsg,
				)
			}

			macosSettingsForState = append(macosSettingsForState, macosSetting)
		}
	}

	return macosSettingsForState
}
