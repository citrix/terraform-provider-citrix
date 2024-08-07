// Copyright © 2024. Citrix Systems, Inc.

package gac_settings

import (
	"context"
	"fmt"
	"reflect"

	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GACSettingsResourceModel struct {
	ServiceUrl      types.String `tfsdk:"service_url"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	UseForAppConfig types.Bool   `tfsdk:"use_for_app_config"`
	AppSettings     types.Object `tfsdk:"app_settings"` // AppSettings
}

type AppSettings struct {
	Windows  types.List `tfsdk:"windows"`  //[]Windows
	Ios      types.List `tfsdk:"ios"`      //[]Ios
	Android  types.List `tfsdk:"android"`  //[]Android
	Chromeos types.List `tfsdk:"chromeos"` //[]Chromeos
	Html5    types.List `tfsdk:"html5"`    //[]Html5
	Macos    types.List `tfsdk:"macos"`    //[]Macos
}

func (AppSettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the device platform and the associated settings. Currently, only settings objects with value type of integer, boolean, strings and list of strings is supported.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"windows": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using windows platform.",
				Optional:     true,
				NestedObject: Windows{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"ios": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using ios platform.",
				Optional:     true,
				NestedObject: Ios{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"android": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using android platform.",
				Optional:     true,
				NestedObject: Android{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"html5": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using html5.",
				Optional:     true,
				NestedObject: Html5{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"chromeos": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using chrome os platform.",
				Optional:     true,
				NestedObject: Chromeos{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"macos": schema.ListNestedAttribute{
				Description:  "Settings to be applied for users using mac os platform.",
				Optional:     true,
				NestedObject: Macos{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (AppSettings) GetAttributes() map[string]schema.Attribute {
	return AppSettings{}.GetSchema().Attributes
}

type Windows struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]WindowsSettings
}

func (Windows) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: WindowsSettings{}.GetSchema(),
			},
		},
	}
}

func (Windows) GetAttributes() map[string]schema.Attribute {
	return Windows{}.GetSchema().Attributes
}

type Ios struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]IosSettings
}

func (Ios) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: IosSettings{}.GetSchema(),
			},
		},
	}
}

func (Ios) GetAttributes() map[string]schema.Attribute {
	return Ios{}.GetSchema().Attributes
}

type Android struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]AndroidSettings
}

func (Android) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: AndroidSettings{}.GetSchema(),
			},
		},
	}
}

func (Android) GetAttributes() map[string]schema.Attribute {
	return Android{}.GetSchema().Attributes
}

type Chromeos struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]ChromeosSettings
}

func (Chromeos) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: ChromeosSettings{}.GetSchema(),
			},
		},
	}
}

func (Chromeos) GetAttributes() map[string]schema.Attribute {
	return Chromeos{}.GetSchema().Attributes
}

type Html5 struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]Html5Settings
}

func (Html5) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: Html5Settings{}.GetSchema(),
			},
		},
	}
}

func (Html5) GetAttributes() map[string]schema.Attribute {
	return Html5{}.GetSchema().Attributes
}

type Macos struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.List   `tfsdk:"settings"` //[]MacosSettings
}

func (Macos) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.ListNestedAttribute{
				Description: "A list of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: MacosSettings{}.GetSchema(),
			},
		},
	}
}

func (Macos) GetAttributes() map[string]schema.Attribute {
	return Macos{}.GetSchema().Attributes
}

type WindowsSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
	ValueList   types.List   `tfsdk:"value_list"`
}

func (WindowsSettings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_string"), path.MatchRelative().AtParent().AtName("value_list")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (WindowsSettings) GetAttributes() map[string]schema.Attribute {
	return WindowsSettings{}.GetSchema().Attributes
}

type IosSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
}

func (IosSettings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (IosSettings) GetAttributes() map[string]schema.Attribute {
	return IosSettings{}.GetSchema().Attributes
}

type AndroidSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
	ValueList   types.List   `tfsdk:"value_list"`
}

func (AndroidSettings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_string"), path.MatchRelative().AtParent().AtName("value_list")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (AndroidSettings) GetAttributes() map[string]schema.Attribute {
	return AndroidSettings{}.GetSchema().Attributes
}

type ChromeosSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
	ValueList   types.List   `tfsdk:"value_list"`
}

func (ChromeosSettings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_string"), path.MatchRelative().AtParent().AtName("value_list")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (ChromeosSettings) GetAttributes() map[string]schema.Attribute {
	return ChromeosSettings{}.GetSchema().Attributes
}

type Html5Settings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
	ValueList   types.List   `tfsdk:"value_list"`
}

func (Html5Settings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_string"), path.MatchRelative().AtParent().AtName("value_list")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (Html5Settings) GetAttributes() map[string]schema.Attribute {
	return Html5Settings{}.GetSchema().Attributes
}

type MacosSettings struct {
	Name        types.String `tfsdk:"name"`
	ValueString types.String `tfsdk:"value_string"`
	ValueList   types.List   `tfsdk:"value_list"`
}

func (MacosSettings) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the setting.",
				Required:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "String value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("value_string"), path.MatchRelative().AtParent().AtName("value_list")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (MacosSettings) GetAttributes() map[string]schema.Attribute {
	return MacosSettings{}.GetSchema().Attributes
}

func (r GACSettingsResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, settingsRecordModel globalappconfiguration.SettingsRecordModel) GACSettingsResourceModel {

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

	planAppSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)

	planAppSettings.Windows = r.getWindowsSettings(ctx, diagnostics, windowsSettings)
	planAppSettings.Ios = r.getIosSettings(ctx, diagnostics, iosSettings)
	planAppSettings.Android = r.getAndroidSettings(ctx, diagnostics, androidSettings)
	planAppSettings.Chromeos = r.getChromeosSettings(ctx, diagnostics, chromeosSettings)
	planAppSettings.Html5 = r.getHtml5Settings(ctx, diagnostics, html5Settings)
	planAppSettings.Macos = r.getMacosSettings(ctx, diagnostics, macosSettings)

	r.AppSettings = util.TypedObjectToObjectValue(ctx, diagnostics, planAppSettings)

	return r
}

func (r GACSettingsResourceModel) getWindowsSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteWindowsSettings []globalappconfiguration.PlatformSettings) types.List {
	var stateWindowsSettings []Windows
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Windows.IsNull() {
			stateWindowsSettings = util.ObjectListToTypedArray[Windows](ctx, diagnostics, appSettings.Windows)
		}
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
			Settings:     getWindowsCategorySettings(ctx, diagnostics, stateWindowsSetting.Settings, remoteWindowsSetting.platformSetting.GetSettings()),
		})

		remoteWindowsSetting.IsVisited = true

	}

	// Add the windows settings from remote which are not present in the state
	for _, remoteWindowsSetting := range remoteWindowsSettingsMap {
		if !remoteWindowsSetting.IsVisited {
			windowsSettingsForState = append(windowsSettingsForState, Windows{
				Category:     types.StringValue(remoteWindowsSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteWindowsSetting.platformSetting.GetUserOverride()),
				Settings:     parseWindowsSettings(ctx, diagnostics, remoteWindowsSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Windows](ctx, diagnostics, windowsSettingsForState)
}

func (r GACSettingsResourceModel) getIosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteIosSettings []globalappconfiguration.PlatformSettings) types.List {
	var stateIosSettings []Ios
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Ios.IsNull() {
			stateIosSettings = util.ObjectListToTypedArray[Ios](ctx, diagnostics, appSettings.Ios)
		}
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
			Settings:     getIosCategorySettings(ctx, diagnostics, stateIosSetting.Settings, remoteIosSetting.platformSetting.GetSettings()),
		})

		remoteIosSetting.IsVisited = true

	}

	// Add the ios settings from remote which are not present in the state
	for _, remoteIosSetting := range remoteIosSettingsMap {
		if !remoteIosSetting.IsVisited {
			iosSettingsForState = append(iosSettingsForState, Ios{
				Category:     types.StringValue(remoteIosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteIosSetting.platformSetting.GetUserOverride()),
				Settings:     parseIosSettings(ctx, diagnostics, remoteIosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Ios](ctx, diagnostics, iosSettingsForState)
}

func (r GACSettingsResourceModel) getAndroidSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteAndroidSettings []globalappconfiguration.PlatformSettings) types.List {
	var stateAndroidSettings []Android
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Android.IsNull() {
			stateAndroidSettings = util.ObjectListToTypedArray[Android](ctx, diagnostics, appSettings.Android)
		}
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
			Settings:     getAndroidCategorySettings(ctx, diagnostics, stateAndroidSetting.Settings, remoteAndroidSetting.platformSetting.GetSettings()),
		})

		remoteAndroidSetting.IsVisited = true

	}

	// Add the android settings from remote which are not present in the state
	for _, remoteAndroidSetting := range remoteAndroidSettingsMap {
		if !remoteAndroidSetting.IsVisited {
			androidSettingsForState = append(androidSettingsForState, Android{
				Category:     types.StringValue(remoteAndroidSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteAndroidSetting.platformSetting.GetUserOverride()),
				Settings:     parseAndroidSettings(ctx, diagnostics, remoteAndroidSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Android](ctx, diagnostics, androidSettingsForState)
}

func (r GACSettingsResourceModel) getHtml5Settings(ctx context.Context, diagnostics *diag.Diagnostics, remoteHtml5Settings []globalappconfiguration.PlatformSettings) types.List {
	var stateHtml5Settings []Html5
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Html5.IsNull() {
			stateHtml5Settings = util.ObjectListToTypedArray[Html5](ctx, diagnostics, appSettings.Html5)
		}
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
			Settings:     getHtml5CategorySettings(ctx, diagnostics, stateHtml5Setting.Settings, remoteHtml5Setting.platformSetting.GetSettings()),
		})

		remoteHtml5Setting.IsVisited = true

	}

	// Add the html5 settings from remote which are not present in the state
	for _, remoteHtml5Setting := range remoteHtml5SettingsMap {
		if !remoteHtml5Setting.IsVisited {
			html5SettingsForState = append(html5SettingsForState, Html5{
				Category:     types.StringValue(remoteHtml5Setting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteHtml5Setting.platformSetting.GetUserOverride()),
				Settings:     parseHtml5Settings(ctx, diagnostics, remoteHtml5Setting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Html5](ctx, diagnostics, html5SettingsForState)
}

func (r GACSettingsResourceModel) getChromeosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteChromeosSettings []globalappconfiguration.PlatformSettings) types.List {
	var stateChromeosSettings []Chromeos
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Chromeos.IsNull() {
			stateChromeosSettings = util.ObjectListToTypedArray[Chromeos](ctx, diagnostics, appSettings.Chromeos)
		}
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
			Settings:     getChromeosCategorySettings(ctx, diagnostics, stateChromeosSetting.Settings, remoteChromeosSetting.platformSetting.GetSettings()),
		})

		remoteChromeosSetting.IsVisited = true

	}

	// Add the chromeos settings from remote which are not present in the state
	for _, remoteChromeosSetting := range remoteChromeosSettingsMap {
		if !remoteChromeosSetting.IsVisited {
			chromeosSettingsForState = append(chromeosSettingsForState, Chromeos{
				Category:     types.StringValue(remoteChromeosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteChromeosSetting.platformSetting.GetUserOverride()),
				Settings:     parseChromeosSettings(ctx, diagnostics, remoteChromeosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Chromeos](ctx, diagnostics, chromeosSettingsForState)
}

func (r GACSettingsResourceModel) getMacosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteMacosSettings []globalappconfiguration.PlatformSettings) types.List {
	var stateMacosSettings []Macos
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Macos.IsNull() {
			stateMacosSettings = util.ObjectListToTypedArray[Macos](ctx, diagnostics, appSettings.Macos)
		}
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
			Settings:     getMacosCategorySettings(ctx, diagnostics, stateMacosSetting.Settings, remoteMacosSetting.platformSetting.GetSettings()),
		})

		remoteMacosSetting.IsVisited = true

	}

	// Add the macos settings from remote which are not present in the state
	for _, remoteMacosSetting := range remoteMacosSettingsMap {
		if !remoteMacosSetting.IsVisited {
			macosSettingsForState = append(macosSettingsForState, Macos{
				Category:     types.StringValue(remoteMacosSetting.platformSetting.GetCategory()),
				UserOverride: types.BoolValue(remoteMacosSetting.platformSetting.GetUserOverride()),
				Settings:     parseMacosSettings(ctx, diagnostics, remoteMacosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectList[Macos](ctx, diagnostics, macosSettingsForState)
}

func parseWindowsSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteWindowsSettings []globalappconfiguration.CategorySettings) types.List {
	var windowsSettings []WindowsSettings
	var errMsg string

	for _, remoteWindowsSetting := range remoteWindowsSettings {
		var windowsSetting WindowsSettings
		windowsSetting.Name = types.StringValue(remoteWindowsSetting.GetName())
		windowsSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteWindowsSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.GetValue().(string))
		case reflect.Slice:
			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[WindowsSettings](ctx, diagnostics, windowsSettings)
}

func parseIosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteIosSettings []globalappconfiguration.CategorySettings) types.List {
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

	return util.TypedArrayToObjectList[IosSettings](ctx, diagnostics, iosSettings)
}

func parseAndroidSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteAndroidSettings []globalappconfiguration.CategorySettings) types.List {
	var androidSettings []AndroidSettings
	var errMsg string

	for _, remoteAndroidSetting := range remoteAndroidSettings {
		var androidSetting AndroidSettings
		androidSetting.Name = types.StringValue(remoteAndroidSetting.GetName())
		androidSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteAndroidSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			androidSetting.ValueString = types.StringValue(remoteAndroidSetting.GetValue().(string))
		case reflect.Slice:
			androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteAndroidSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[AndroidSettings](ctx, diagnostics, androidSettings)
}

func parseHtml5Settings(ctx context.Context, diagnostics *diag.Diagnostics, remoteHtml5Settings []globalappconfiguration.CategorySettings) types.List {
	var html5Settings []Html5Settings
	var errMsg string

	for _, remoteHtml5Setting := range remoteHtml5Settings {
		var html5Setting Html5Settings
		html5Setting.Name = types.StringValue(remoteHtml5Setting.GetName())
		html5Setting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteHtml5Setting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			html5Setting.ValueString = types.StringValue(remoteHtml5Setting.GetValue().(string))
		case reflect.Slice:
			html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteHtml5Setting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[Html5Settings](ctx, diagnostics, html5Settings)
}

func parseMacosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteMacosSettings []globalappconfiguration.CategorySettings) types.List {
	var macosSettings []MacosSettings
	var errMsg string

	for _, remoteMacosSetting := range remoteMacosSettings {
		var macosSetting MacosSettings
		macosSetting.Name = types.StringValue(remoteMacosSetting.GetName())
		macosSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteMacosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.GetValue().(string))
		case reflect.Slice:
			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[MacosSettings](ctx, diagnostics, macosSettings)
}

func parseChromeosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteChromeosSettings []globalappconfiguration.CategorySettings) types.List {
	var chromeosSettings []ChromeosSettings
	var errMsg string

	for _, remoteChromeosSetting := range remoteChromeosSettings {
		var chromeosSetting ChromeosSettings
		chromeosSetting.Name = types.StringValue(remoteChromeosSetting.GetName())
		chromeosSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteChromeosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.GetValue().(string))
		case reflect.Slice:
			chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteChromeosSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[ChromeosSettings](ctx, diagnostics, chromeosSettings)
}

func getWindowsCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, windowsSettings types.List, remoteWindowsSettings []globalappconfiguration.CategorySettings) types.List {

	var windowsSettingsForState []WindowsSettings
	var errMsg string

	stateWindowsSettings := util.ObjectListToTypedArray[WindowsSettings](ctx, diagnostics, windowsSettings)

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
		windowsSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteWindowsSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
		case reflect.Slice:
			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))
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
			windowsSetting.ValueList = types.ListNull(types.StringType)
			valueType := reflect.TypeOf(remoteWindowsSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
			case reflect.Slice:
				windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[WindowsSettings](ctx, diagnostics, windowsSettingsForState)
}

func getIosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, iosSettings types.List, remoteIosSettings []globalappconfiguration.CategorySettings) types.List {

	var iosSettingsForState []IosSettings
	var errMsg string

	stateIosSettings := util.ObjectListToTypedArray[IosSettings](ctx, diagnostics, iosSettings)

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
				errMsg = fmt.Sprintf("Unsupported type for ios setting value: %v", valueType.Kind())
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

	return util.TypedArrayToObjectList[IosSettings](ctx, diagnostics, iosSettingsForState)
}

func getAndroidCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, androidSettings types.List, remoteAndroidSettings []globalappconfiguration.CategorySettings) types.List {

	var androidSettingsForState []AndroidSettings
	var errMsg string

	stateAndroidSettings := util.ObjectListToTypedArray[AndroidSettings](ctx, diagnostics, androidSettings)

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
		androidSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteAndroidSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			androidSetting.ValueString = types.StringValue(remoteAndroidSetting.Value.(string))
		case reflect.Slice:
			androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteAndroidSetting.Value.([]interface{}))
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
			androidSetting.ValueList = types.ListNull(types.StringType)
			valueType := reflect.TypeOf(remoteAndroidSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				androidSetting.ValueString = types.StringValue(remoteAndroidSetting.Value.(string))
			case reflect.Slice:
				androidSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteAndroidSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[AndroidSettings](ctx, diagnostics, androidSettingsForState)
}

func getChromeosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, chromeosSettings types.List, remoteChromeosSettings []globalappconfiguration.CategorySettings) types.List {

	var chromeosSettingsForState []ChromeosSettings
	var errMsg string

	stateChromeosSettings := util.ObjectListToTypedArray[ChromeosSettings](ctx, diagnostics, chromeosSettings)

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
		chromeosSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteChromeosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.Value.(string))
		case reflect.Slice:
			chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteChromeosSetting.Value.([]interface{}))
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
			chromeosSetting.ValueList = types.ListNull(types.StringType)
			valueType := reflect.TypeOf(remoteChromeosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				chromeosSetting.ValueString = types.StringValue(remoteChromeosSetting.Value.(string))
			case reflect.Slice:
				chromeosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteChromeosSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[ChromeosSettings](ctx, diagnostics, chromeosSettingsForState)
}

func getHtml5CategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, html5Settings types.List, remoteHtml5Settings []globalappconfiguration.CategorySettings) types.List {

	var html5SettingsForState []Html5Settings
	var errMsg string

	stateHtml5Settings := util.ObjectListToTypedArray[Html5Settings](ctx, diagnostics, html5Settings)

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
		html5Setting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteHtml5Setting.Value)
		switch valueType.Kind() {
		case reflect.String:
			html5Setting.ValueString = types.StringValue(remoteHtml5Setting.Value.(string))
		case reflect.Slice:
			html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteHtml5Setting.Value.([]interface{}))
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
			html5Setting.ValueList = types.ListNull(types.StringType)
			valueType := reflect.TypeOf(remoteHtml5Setting.Value)
			switch valueType.Kind() {
			case reflect.String:
				html5Setting.ValueString = types.StringValue(remoteHtml5Setting.Value.(string))
			case reflect.Slice:
				html5Setting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteHtml5Setting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[Html5Settings](ctx, diagnostics, html5SettingsForState)
}

func getMacosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, macosSettings types.List, remoteMacosSettings []globalappconfiguration.CategorySettings) types.List {

	var macosSettingsForState []MacosSettings
	var errMsg string

	stateMacosSettings := util.ObjectListToTypedArray[MacosSettings](ctx, diagnostics, macosSettings)

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
		macosSetting.ValueList = types.ListNull(types.StringType)
		valueType := reflect.TypeOf(remoteMacosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
		case reflect.Slice:
			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))
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
			macosSetting.ValueList = types.ListNull(types.StringType)
			valueType := reflect.TypeOf(remoteMacosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
			case reflect.Slice:
				macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))
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

	return util.TypedArrayToObjectList[MacosSettings](ctx, diagnostics, macosSettingsForState)
}

func GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Manages the Global App Configuration settings for a service url.",
		Attributes: map[string]schema.Attribute{
			"service_url": schema.StringAttribute{
				Description: "Citrix workspace application store url for which settings are to be configured. The value is case sensitive and requires the protocol (\"https\" or \"http\") and port number.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the settings record.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the settings record.",
				Required:    true, //Check if this can be made into Optional
			},
			"use_for_app_config": schema.BoolAttribute{
				Description: "Defines whether to use the settings for app configuration or not. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"app_settings": AppSettings{}.GetSchema(),
		},
	}
}
