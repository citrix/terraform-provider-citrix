// Copyright © 2024. Citrix Systems, Inc.

package global_app_configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	TestChannel     types.Bool   `tfsdk:"test_channel"`
}

func (GACSettingsResourceModel) GetAttributes() map[string]schema.Attribute {
	return GACSettingsResourceModel{}.GetSchema().Attributes
}

type AppSettings struct {
	Windows  types.Set `tfsdk:"windows"`  //Set[Windows]
	Ios      types.Set `tfsdk:"ios"`      //Set[Ios]
	Android  types.Set `tfsdk:"android"`  //Set[Android]
	Chromeos types.Set `tfsdk:"chromeos"` //Set[Chromeos]
	Html5    types.Set `tfsdk:"html5"`    //Set[Html5]
	Macos    types.Set `tfsdk:"macos"`    //Set[Macos]
	Linux    types.Set `tfsdk:"linux"`    //Set[Linux]
}

func (AppSettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the device platform and the associated settings. Currently, only settings objects with value type of integer, boolean, strings and list of strings is supported.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"windows": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using windows platform.",
				Optional:     true,
				NestedObject: Windows{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"ios": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using ios platform.",
				Optional:     true,
				NestedObject: Ios{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"android": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using android platform.",
				Optional:     true,
				NestedObject: Android{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"html5": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using html5.",
				Optional:     true,
				NestedObject: Html5{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"chromeos": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using chrome os platform.",
				Optional:     true,
				NestedObject: Chromeos{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"macos": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using mac os platform.",
				Optional:     true,
				NestedObject: Macos{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"linux": schema.SetNestedAttribute{
				Description:  "Settings to be applied for users using linux platform.",
				Optional:     true,
				NestedObject: Linux{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[WindowsSettings]
}

func (Windows) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[IosSettings]
}

func (Ios) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[AndroidSettings]
}

func (Android) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[ChromeosSettings]
}

func (Chromeos) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[Html5Settings]
}

func (Html5) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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
	Settings     types.Set    `tfsdk:"settings"` //Set[MacosSettings]
}

func (Macos) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: MacosSettings{}.GetSchema(),
			},
		},
	}
}

func (Macos) GetAttributes() map[string]schema.Attribute {
	return Macos{}.GetSchema().Attributes
}

type Linux struct {
	Category     types.String `tfsdk:"category"`
	UserOverride types.Bool   `tfsdk:"user_override"`
	Settings     types.Set    `tfsdk:"settings"` //Set[LinuxSettings]
}

func (Linux) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Description: "Defines the category of the setting.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "\nCategory name must be all lowercase letters "),
				},
			},
			"user_override": schema.BoolAttribute{
				Description: "Defines if users can modify or change the value of as obtained settings from the Global App Citrix Workspace configuration service.",
				Required:    true,
			},
			"settings": schema.SetNestedAttribute{
				Description: "A set of name value pairs for the settings. Please refer to the following [table](https://developer-docs.citrix.com/en-us/server-integration/global-app-configuration-service/getting-started#supported-settings-and-their-values-per-platform) for the supported settings name and their values per platform.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: LinuxSettings{}.GetSchema(),
			},
		},
	}
}

func (Linux) GetAttributes() map[string]schema.Attribute {
	return Linux{}.GetSchema().Attributes
}

type WindowsSettings struct {
	Name                           types.String `tfsdk:"name"`
	ValueString                    types.String `tfsdk:"value_string"`
	ValueList                      types.List   `tfsdk:"value_list"`
	LocalAppAllowList              types.Set    `tfsdk:"local_app_allow_list"`               //Set[LocalAppAllowSetModel]
	ExtensionInstallAllowList      types.Set    `tfsdk:"extension_install_allow_list"`       //Set[ExtensionInstallAllowSetModel]
	AutoLaunchProtocolsFromOrigins types.Set    `tfsdk:"auto_launch_protocols_from_origins"` //Set[AutoLaunchProtocolsFromOriginsModel]
	ManagedBookmarks               types.Set    `tfsdk:"managed_bookmarks"`                  //Set[BookMarkValueModel]
	EnterpriseBroswerSSO           types.Object `tfsdk:"enterprise_browser_sso"`             //CitrixEnterpriseBrowserModel
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
			"local_app_allow_list": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "Set of App Object to allow list for Local App Discovery.",
				NestedObject: LocalAppAllowListModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"extension_install_allow_list": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "An allowed list of extensions that users can add to the Citrix Enterprise Browser. This list uses the Chrome Web Store.",
				NestedObject: ExtensionInstallAllowListModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"auto_launch_protocols_from_origins": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "A set of protocols that can launch an external application from the listed origins without prompting the user.",
				NestedObject: AutoLaunchProtocolsFromOriginsModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"managed_bookmarks": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "A set of bookmarks to push to the Citrix Enterprise Browser.",
				NestedObject: BookMarkValueModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"enterprise_browser_sso": CitrixEnterpriseBrowserModel{}.GetSchema(),
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
				Required:    true,
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

type LinuxSettings struct {
	Name                           types.String `tfsdk:"name"`
	ValueString                    types.String `tfsdk:"value_string"`
	ValueList                      types.List   `tfsdk:"value_list"`
	ExtensionInstallAllowList      types.Set    `tfsdk:"extension_install_allow_list"`       //Set[ExtensionInstallAllowListModel]
	AutoLaunchProtocolsFromOrigins types.Set    `tfsdk:"auto_launch_protocols_from_origins"` //Set[AutoLaunchProtocolsFromOriginsModel]
	ManagedBookmarks               types.Set    `tfsdk:"managed_bookmarks"`                  //Set[BookMarkValueModel]
}

func (LinuxSettings) GetSchema() schema.NestedAttributeObject {
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
			"value_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List value (if any) associated with the setting.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"extension_install_allow_list": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "An allowed list of extensions that users can add to the Citrix Enterprise Browser. This list uses the Chrome Web Store.",
				NestedObject: ExtensionInstallAllowListModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"auto_launch_protocols_from_origins": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "A set of protocols that can launch an external application from the listed origins without prompting the user.",
				NestedObject: AutoLaunchProtocolsFromOriginsModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"managed_bookmarks": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "A set of bookmarks to push to the Citrix Enterprise Browser.",
				NestedObject: BookMarkValueModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (LinuxSettings) GetAttributes() map[string]schema.Attribute {
	return LinuxSettings{}.GetSchema().Attributes
}

type MacosSettings struct {
	Name                           types.String `tfsdk:"name"`
	ValueString                    types.String `tfsdk:"value_string"`
	ValueList                      types.List   `tfsdk:"value_list"`
	AutoLaunchProtocolsFromOrigins types.Set    `tfsdk:"auto_launch_protocols_from_origins"` //Set[AutoLaunchProtocolsFromOrigins]
	ManagedBookmarks               types.Set    `tfsdk:"managed_bookmarks"`                  //Set[BookMarkValue]
	ExtensionInstallAllowList      types.Set    `tfsdk:"extension_install_allow_list"`       //Set[ExtensionInstallAllowList]
	EnterpriseBroswerSSO           types.Object `tfsdk:"enterprise_browser_sso"`             //CitrixEnterpriseBrowserModel
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
			"auto_launch_protocols_from_origins": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "Specify a list of protocols that can launch an external application from the listed origins without prompting the user.",
				NestedObject: AutoLaunchProtocolsFromOriginsModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"managed_bookmarks": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "Array of objects of type ManagedBookmarks. For example: {name:\"bookmark_name1\",url:\"bookmark_url1\"}",
				NestedObject: BookMarkValueModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"extension_install_allow_list": schema.SetNestedAttribute{
				Optional:     true,
				Description:  "Array of objects of type ExtensionInstallAllowlist. For example: {id:\"extension_id1\",name:\"extension_name1\",install link:\"chrome store url for the extension\"}",
				NestedObject: ExtensionInstallAllowListModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"enterprise_browser_sso": CitrixEnterpriseBrowserModel{}.GetSchema(),
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

	if settingsRecordModel.SettingsChannel != nil && settingsRecordModel.SettingsChannel.GetChannelName() != "" {
		r.TestChannel = types.BoolValue(true)
	} else {
		r.TestChannel = types.BoolValue(false)
	}

	var appSettings = settings.GetAppSettings()
	var windowsSettings = appSettings.GetWindows()
	var iosSettings = appSettings.GetIos()
	var androidSettings = appSettings.GetAndroid()
	var chromeosSettings = appSettings.GetChromeos()
	var html5Settings = appSettings.GetHtml5()
	var macosSettings = appSettings.GetMacos()
	var linuxSettings = appSettings.GetLinux()

	planAppSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)

	planAppSettings.Windows = r.getWindowsSettings(ctx, diagnostics, windowsSettings)
	planAppSettings.Ios = r.getIosSettings(ctx, diagnostics, iosSettings)
	planAppSettings.Android = r.getAndroidSettings(ctx, diagnostics, androidSettings)
	planAppSettings.Chromeos = r.getChromeosSettings(ctx, diagnostics, chromeosSettings)
	planAppSettings.Html5 = r.getHtml5Settings(ctx, diagnostics, html5Settings)
	planAppSettings.Macos = r.getMacosSettings(ctx, diagnostics, macosSettings)
	planAppSettings.Linux = r.getLinuxSettings(ctx, diagnostics, linuxSettings)

	r.AppSettings = util.TypedObjectToObjectValue(ctx, diagnostics, planAppSettings)

	return r
}

func (r GACSettingsResourceModel) getWindowsSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteWindowsSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateWindowsSettings []Windows
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Windows.IsNull() {
			stateWindowsSettings = util.ObjectSetToTypedArray[Windows](ctx, diagnostics, appSettings.Windows)
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
		remoteWindowsSetting, exists := remoteWindowsSettingsMap[strings.ToLower(stateWindowsSetting.Category.ValueString())]
		if !exists {
			// If windows setting is not present in the remote, then don't add it to the state
			continue
		}

		windowsSettingsForState = append(windowsSettingsForState, Windows{
			Category:     types.StringValue(strings.ToLower(remoteWindowsSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteWindowsSetting.platformSetting.GetUserOverride()),
			Settings:     getWindowsCategorySettings(ctx, diagnostics, stateWindowsSetting.Settings, remoteWindowsSetting.platformSetting.GetSettings()),
		})

		remoteWindowsSetting.IsVisited = true

	}

	// Add the windows settings from remote which are not present in the state
	for _, remoteWindowsSetting := range remoteWindowsSettingsMap {
		if !remoteWindowsSetting.IsVisited {
			windowsSettingsForState = append(windowsSettingsForState, Windows{
				Category:     types.StringValue(strings.ToLower(remoteWindowsSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteWindowsSetting.platformSetting.GetUserOverride()),
				Settings:     parseWindowsSettings(ctx, diagnostics, remoteWindowsSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Windows](ctx, diagnostics, windowsSettingsForState)
}

func (r GACSettingsResourceModel) getLinuxSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteLinuxSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateLinuxSettings []Linux
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Linux.IsNull() {
			stateLinuxSettings = util.ObjectSetToTypedArray[Linux](ctx, diagnostics, appSettings.Linux)
		}
	}

	type RemoteLinuxSettingsTracker struct {
		platformSetting globalappconfiguration.PlatformSettings
		IsVisited       bool
	}

	// Create a map of category -> RemoteLinuxSettingsTracker for remote
	remoteLinuxSettingsMap := map[string]*RemoteLinuxSettingsTracker{}
	for _, remoteLinuxSetting := range remoteLinuxSettings {
		remoteLinuxSettingsMap[remoteLinuxSetting.GetCategory()] = &RemoteLinuxSettingsTracker{
			platformSetting: remoteLinuxSetting,
			IsVisited:       false,
		}
	}

	// Prepare the linux settings list to be stored in the state
	var linuxSettingsForState []Linux
	for _, stateLinuxSetting := range stateLinuxSettings {
		remoteLinuxSetting, exists := remoteLinuxSettingsMap[stateLinuxSetting.Category.ValueString()]
		if !exists {
			// If linux setting is not present in the remote, then don't add it to the state
			continue
		}

		linuxSettingsForState = append(linuxSettingsForState, Linux{
			Category:     types.StringValue(strings.ToLower(remoteLinuxSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteLinuxSetting.platformSetting.GetUserOverride()),
			Settings:     getLinuxCategorySettings(ctx, diagnostics, stateLinuxSetting.Settings, remoteLinuxSetting.platformSetting.GetSettings()),
		})

		remoteLinuxSetting.IsVisited = true

	}

	// Add the linux settings from remote which are not present in the state
	for _, remoteLinuxSetting := range remoteLinuxSettingsMap {
		if !remoteLinuxSetting.IsVisited {
			linuxSettingsForState = append(linuxSettingsForState, Linux{
				Category:     types.StringValue(strings.ToLower(remoteLinuxSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteLinuxSetting.platformSetting.GetUserOverride()),
				Settings:     parseLinuxSettings(ctx, diagnostics, remoteLinuxSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Linux](ctx, diagnostics, linuxSettingsForState)
}

func (r GACSettingsResourceModel) getIosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteIosSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateIosSettings []Ios
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Ios.IsNull() {
			stateIosSettings = util.ObjectSetToTypedArray[Ios](ctx, diagnostics, appSettings.Ios)
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
			Category:     types.StringValue(strings.ToLower(remoteIosSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteIosSetting.platformSetting.GetUserOverride()),
			Settings:     getIosCategorySettings(ctx, diagnostics, stateIosSetting.Settings, remoteIosSetting.platformSetting.GetSettings()),
		})

		remoteIosSetting.IsVisited = true

	}

	// Add the ios settings from remote which are not present in the state
	for _, remoteIosSetting := range remoteIosSettingsMap {
		if !remoteIosSetting.IsVisited {
			iosSettingsForState = append(iosSettingsForState, Ios{
				Category:     types.StringValue(strings.ToLower(remoteIosSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteIosSetting.platformSetting.GetUserOverride()),
				Settings:     parseIosSettings(ctx, diagnostics, remoteIosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Ios](ctx, diagnostics, iosSettingsForState)
}

func (r GACSettingsResourceModel) getAndroidSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteAndroidSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateAndroidSettings []Android
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Android.IsNull() {
			stateAndroidSettings = util.ObjectSetToTypedArray[Android](ctx, diagnostics, appSettings.Android)
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
			Category:     types.StringValue(strings.ToLower(remoteAndroidSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteAndroidSetting.platformSetting.GetUserOverride()),
			Settings:     getAndroidCategorySettings(ctx, diagnostics, stateAndroidSetting.Settings, remoteAndroidSetting.platformSetting.GetSettings()),
		})

		remoteAndroidSetting.IsVisited = true

	}

	// Add the android settings from remote which are not present in the state
	for _, remoteAndroidSetting := range remoteAndroidSettingsMap {
		if !remoteAndroidSetting.IsVisited {
			androidSettingsForState = append(androidSettingsForState, Android{
				Category:     types.StringValue(strings.ToLower(remoteAndroidSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteAndroidSetting.platformSetting.GetUserOverride()),
				Settings:     parseAndroidSettings(ctx, diagnostics, remoteAndroidSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Android](ctx, diagnostics, androidSettingsForState)
}

func (r GACSettingsResourceModel) getHtml5Settings(ctx context.Context, diagnostics *diag.Diagnostics, remoteHtml5Settings []globalappconfiguration.PlatformSettings) types.Set {
	var stateHtml5Settings []Html5
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Html5.IsNull() {
			stateHtml5Settings = util.ObjectSetToTypedArray[Html5](ctx, diagnostics, appSettings.Html5)
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
			Category:     types.StringValue(strings.ToLower(remoteHtml5Setting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteHtml5Setting.platformSetting.GetUserOverride()),
			Settings:     getHtml5CategorySettings(ctx, diagnostics, stateHtml5Setting.Settings, remoteHtml5Setting.platformSetting.GetSettings()),
		})

		remoteHtml5Setting.IsVisited = true

	}

	// Add the html5 settings from remote which are not present in the state
	for _, remoteHtml5Setting := range remoteHtml5SettingsMap {
		if !remoteHtml5Setting.IsVisited {
			html5SettingsForState = append(html5SettingsForState, Html5{
				Category:     types.StringValue(strings.ToLower(remoteHtml5Setting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteHtml5Setting.platformSetting.GetUserOverride()),
				Settings:     parseHtml5Settings(ctx, diagnostics, remoteHtml5Setting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Html5](ctx, diagnostics, html5SettingsForState)
}

func (r GACSettingsResourceModel) getChromeosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteChromeosSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateChromeosSettings []Chromeos
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Chromeos.IsNull() {
			stateChromeosSettings = util.ObjectSetToTypedArray[Chromeos](ctx, diagnostics, appSettings.Chromeos)
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
			Category:     types.StringValue(strings.ToLower(remoteChromeosSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteChromeosSetting.platformSetting.GetUserOverride()),
			Settings:     getChromeosCategorySettings(ctx, diagnostics, stateChromeosSetting.Settings, remoteChromeosSetting.platformSetting.GetSettings()),
		})

		remoteChromeosSetting.IsVisited = true

	}

	// Add the chromeos settings from remote which are not present in the state
	for _, remoteChromeosSetting := range remoteChromeosSettingsMap {
		if !remoteChromeosSetting.IsVisited {
			chromeosSettingsForState = append(chromeosSettingsForState, Chromeos{
				Category:     types.StringValue(strings.ToLower(remoteChromeosSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteChromeosSetting.platformSetting.GetUserOverride()),
				Settings:     parseChromeosSettings(ctx, diagnostics, remoteChromeosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Chromeos](ctx, diagnostics, chromeosSettingsForState)
}

func (r GACSettingsResourceModel) getMacosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteMacosSettings []globalappconfiguration.PlatformSettings) types.Set {
	var stateMacosSettings []Macos
	if !r.AppSettings.IsNull() {
		appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, diagnostics, r.AppSettings)
		if !appSettings.Macos.IsNull() {
			stateMacosSettings = util.ObjectSetToTypedArray[Macos](ctx, diagnostics, appSettings.Macos)
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
			Category:     types.StringValue(strings.ToLower(remoteMacosSetting.platformSetting.GetCategory())),
			UserOverride: types.BoolValue(remoteMacosSetting.platformSetting.GetUserOverride()),
			Settings:     getMacosCategorySettings(ctx, diagnostics, stateMacosSetting.Settings, remoteMacosSetting.platformSetting.GetSettings()),
		})

		remoteMacosSetting.IsVisited = true

	}

	// Add the macos settings from remote which are not present in the state
	for _, remoteMacosSetting := range remoteMacosSettingsMap {
		if !remoteMacosSetting.IsVisited {
			macosSettingsForState = append(macosSettingsForState, Macos{
				Category:     types.StringValue(strings.ToLower(remoteMacosSetting.platformSetting.GetCategory())),
				UserOverride: types.BoolValue(remoteMacosSetting.platformSetting.GetUserOverride()),
				Settings:     parseMacosSettings(ctx, diagnostics, remoteMacosSetting.platformSetting.GetSettings()),
			})
		}
	}

	return util.TypedArrayToObjectSet[Macos](ctx, diagnostics, macosSettingsForState)
}

func parseWindowsSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteWindowsSettings []globalappconfiguration.CategorySettings) types.Set {
	var windowsSettings []WindowsSettings
	var errMsg string

	for _, remoteWindowsSetting := range remoteWindowsSettings {
		var windowsSetting WindowsSettings
		WindowsSettingsDefaultValues(ctx, diagnostics, &windowsSetting)
		windowsSetting.Name = types.StringValue(remoteWindowsSetting.GetName())
		valueType := reflect.TypeOf(remoteWindowsSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.GetValue().(string))
		case reflect.Slice:
			localAppAllowList := GACSettingsUpdate[LocalAppAllowListModel, LocalAppAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if localAppAllowList != nil {
				windowsSetting.LocalAppAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, localAppAllowList)
				break
			}

			extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if extentionInstallAllowList != nil {
				windowsSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
				break
			}

			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if autoLaunchProtocolsList != nil {
				windowsSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if managedBookmarkList != nil {
				windowsSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}

			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))
		case reflect.Map:
			v := remoteWindowsSetting.Value.(map[string]interface{})
			elementJSON, err := json.Marshal(v)
			if err != nil {
				errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var app CitrixEnterpriseBrowserModel_Go
			if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for the conversion
				errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var browserModel CitrixEnterpriseBrowserModel
			browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
			browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
			windowsSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)
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

	return util.TypedArrayToObjectSet[WindowsSettings](ctx, diagnostics, windowsSettings)
}

func parseLinuxSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteLinuxSettings []globalappconfiguration.CategorySettings) types.Set {
	var linuxSettings []LinuxSettings
	var errMsg string

	for _, remoteLinuxSetting := range remoteLinuxSettings {
		var linuxSetting LinuxSettings
		LinuxSettingsDefaultValues(ctx, diagnostics, &linuxSetting)
		linuxSetting.Name = types.StringValue(remoteLinuxSetting.GetName())
		valueType := reflect.TypeOf(remoteLinuxSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			linuxSetting.ValueString = types.StringValue(remoteLinuxSetting.GetValue().(string))
		case reflect.Slice:
			extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if extentionInstallAllowList != nil {
				linuxSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
				break
			}

			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if autoLaunchProtocolsList != nil {
				linuxSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if managedBookmarkList != nil {
				linuxSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}

			linuxSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteLinuxSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for linux setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+linuxSetting.Name.ValueString(),
				errMsg,
			)
		}

		linuxSettings = append(linuxSettings, linuxSetting)
	}

	return util.TypedArrayToObjectSet[LinuxSettings](ctx, diagnostics, linuxSettings)
}

func parseIosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteIosSettings []globalappconfiguration.CategorySettings) types.Set {
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

	return util.TypedArrayToObjectSet[IosSettings](ctx, diagnostics, iosSettings)
}

func parseAndroidSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteAndroidSettings []globalappconfiguration.CategorySettings) types.Set {
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

	return util.TypedArrayToObjectSet[AndroidSettings](ctx, diagnostics, androidSettings)
}

func parseHtml5Settings(ctx context.Context, diagnostics *diag.Diagnostics, remoteHtml5Settings []globalappconfiguration.CategorySettings) types.Set {
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

	return util.TypedArrayToObjectSet[Html5Settings](ctx, diagnostics, html5Settings)
}

func parseMacosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteMacosSettings []globalappconfiguration.CategorySettings) types.Set {
	var macosSettings []MacosSettings
	var errMsg string

	for _, remoteMacosSetting := range remoteMacosSettings {
		var macosSetting MacosSettings
		MacosSettingsDefaultValues(ctx, diagnostics, &macosSetting) // Set the default list null for macos settings
		macosSetting.Name = types.StringValue(remoteMacosSetting.GetName())
		valueType := reflect.TypeOf(remoteMacosSetting.GetValue())
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.GetValue().(string))
		case reflect.Slice:
			// Check if the value is of type LocalAppAllowList
			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if autoLaunchProtocolsList != nil {
				macosSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			// Check if the value is of type BookMarkValue
			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if managedBookmarkList != nil {
				macosSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}
			// Check if the value is of type ExtensionInstallAllowList
			extensionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if extensionInstallAllowList != nil {
				macosSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extensionInstallAllowList)
				break
			}

			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))

		case reflect.Map:
			v := remoteMacosSetting.Value.(map[string]interface{})
			elementJSON, err := json.Marshal(v)
			if err != nil {
				errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var app CitrixEnterpriseBrowserModel_Go
			if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for the conversion
				errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var browserModel CitrixEnterpriseBrowserModel
			browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
			browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
			macosSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)

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

	return util.TypedArrayToObjectSet[MacosSettings](ctx, diagnostics, macosSettings)
}

func parseChromeosSettings(ctx context.Context, diagnostics *diag.Diagnostics, remoteChromeosSettings []globalappconfiguration.CategorySettings) types.Set {
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

	return util.TypedArrayToObjectSet[ChromeosSettings](ctx, diagnostics, chromeosSettings)
}

func getWindowsCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, windowsSettings types.Set, remoteWindowsSettings []globalappconfiguration.CategorySettings) types.Set {

	var windowsSettingsForState []WindowsSettings
	var errMsg string

	stateWindowsSettings := util.ObjectSetToTypedArray[WindowsSettings](ctx, diagnostics, windowsSettings)

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
		windowsSetting.Name = stateWindowsSetting.Name
		WindowsSettingsDefaultValues(ctx, diagnostics, &windowsSetting)
		valueType := reflect.TypeOf(remoteWindowsSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
		case reflect.Slice:
			localAppAllowList := GACSettingsUpdate[LocalAppAllowListModel, LocalAppAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if localAppAllowList != nil {
				windowsSetting.LocalAppAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, localAppAllowList)
				break
			}

			extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if extentionInstallAllowList != nil {
				windowsSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
				break
			}

			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if autoLaunchProtocolsList != nil {
				windowsSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
			if managedBookmarkList != nil {
				windowsSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}

			windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))

		case reflect.Map:
			v := remoteWindowsSetting.Value.(map[string]interface{})
			elementJSON, err := json.Marshal(v)
			if err != nil {
				errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var app CitrixEnterpriseBrowserModel_Go
			if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for the conversion
				errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var browserModel CitrixEnterpriseBrowserModel
			browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
			browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
			windowsSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)

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
			WindowsSettingsDefaultValues(ctx, diagnostics, &windowsSetting)
			valueType := reflect.TypeOf(remoteWindowsSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				windowsSetting.ValueString = types.StringValue(remoteWindowsSetting.Value.(string))
			case reflect.Slice:
				localAppAllowList := GACSettingsUpdate[LocalAppAllowListModel, LocalAppAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
				if localAppAllowList != nil {
					windowsSetting.LocalAppAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, localAppAllowList)
					break
				}

				extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
				if extentionInstallAllowList != nil {
					windowsSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
					break
				}

				autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
				if autoLaunchProtocolsList != nil {
					windowsSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
					break
				}

				managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteWindowsSetting.Value)
				if managedBookmarkList != nil {
					windowsSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
					break
				}

				windowsSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteWindowsSetting.Value.([]interface{}))
			case reflect.Map:
				v := remoteWindowsSetting.Value.(map[string]interface{})
				elementJSON, err := json.Marshal(v)
				if err != nil {
					errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
					break
				}
				var app CitrixEnterpriseBrowserModel_Go
				if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for the conversion
					errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
					break
				}
				var browserModel CitrixEnterpriseBrowserModel
				browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
				browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
				windowsSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)

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

	return util.TypedArrayToObjectSet[WindowsSettings](ctx, diagnostics, windowsSettingsForState)
}

func getLinuxCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, linuxSettings types.Set, remoteLinuxSettings []globalappconfiguration.CategorySettings) types.Set {

	var linuxSettingsForState []LinuxSettings
	var errMsg string

	stateLinuxSettings := util.ObjectSetToTypedArray[LinuxSettings](ctx, diagnostics, linuxSettings)

	type RemoteSettingsTracker struct {
		Value     interface{}
		IsVisited bool
	}

	// Create a map of name -> RemoteSettingsTracker for remote
	remoteLinuxCategorySettingsMap := map[string]*RemoteSettingsTracker{}
	for _, remoteLinuxSetting := range remoteLinuxSettings {
		remoteLinuxCategorySettingsMap[remoteLinuxSetting.GetName()] = &RemoteSettingsTracker{
			Value:     remoteLinuxSetting.GetValue(),
			IsVisited: false,
		}
	}

	// Prepare the linux settings list to be stored in the state
	for _, stateLinuxSetting := range stateLinuxSettings {
		remoteLinuxSetting, exists := remoteLinuxCategorySettingsMap[stateLinuxSetting.Name.ValueString()]
		if !exists {
			// If linux setting is not present in the remote, then don't add it to the state
			continue
		}

		var linuxSetting LinuxSettings
		errMsg = ""
		linuxSetting.Name = stateLinuxSetting.Name
		LinuxSettingsDefaultValues(ctx, diagnostics, &linuxSetting)
		valueType := reflect.TypeOf(remoteLinuxSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			linuxSetting.ValueString = types.StringValue(remoteLinuxSetting.Value.(string))
		case reflect.Slice:
			extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if extentionInstallAllowList != nil {
				linuxSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
				break
			}

			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if autoLaunchProtocolsList != nil {
				linuxSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
			if managedBookmarkList != nil {
				linuxSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}

			linuxSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteLinuxSetting.Value.([]interface{}))
		default:
			errMsg = fmt.Sprintf("Unsupported type for linux setting value: %v", valueType.Kind())
		}
		if errMsg != "" {
			diagnostics.AddError(
				"Could not parse value for the setting:"+linuxSetting.Name.ValueString(),
				errMsg,
			)
		}

		linuxSettingsForState = append(linuxSettingsForState, linuxSetting)
		remoteLinuxSetting.IsVisited = true
	}
	// Add the linux settings from remote which are not present in the state
	for settingName, remoteLinuxSetting := range remoteLinuxCategorySettingsMap {
		if !remoteLinuxSetting.IsVisited {
			var linuxSetting LinuxSettings
			errMsg = ""

			linuxSetting.Name = types.StringValue(settingName)
			LinuxSettingsDefaultValues(ctx, diagnostics, &linuxSetting)
			valueType := reflect.TypeOf(remoteLinuxSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				linuxSetting.ValueString = types.StringValue(remoteLinuxSetting.Value.(string))
			case reflect.Slice:
				extentionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
				if extentionInstallAllowList != nil {
					linuxSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extentionInstallAllowList)
					break
				}

				autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
				if autoLaunchProtocolsList != nil {
					linuxSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
					break
				}

				managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteLinuxSetting.Value)
				if managedBookmarkList != nil {
					linuxSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
					break
				}

				linuxSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteLinuxSetting.Value.([]interface{}))
			default:
				errMsg = fmt.Sprintf("Unsupported type for linux setting value: %v", valueType.Kind())
			}
			if errMsg != "" {
				diagnostics.AddError(
					"Could not parse value for the setting:"+linuxSetting.Name.ValueString(),
					errMsg,
				)
			}

			linuxSettingsForState = append(linuxSettingsForState, linuxSetting)
		}
	}

	return util.TypedArrayToObjectSet[LinuxSettings](ctx, diagnostics, linuxSettingsForState)
}

func getIosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, iosSettings types.Set, remoteIosSettings []globalappconfiguration.CategorySettings) types.Set {

	var iosSettingsForState []IosSettings
	var errMsg string

	stateIosSettings := util.ObjectSetToTypedArray[IosSettings](ctx, diagnostics, iosSettings)

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

	return util.TypedArrayToObjectSet[IosSettings](ctx, diagnostics, iosSettingsForState)
}

func getAndroidCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, androidSettings types.Set, remoteAndroidSettings []globalappconfiguration.CategorySettings) types.Set {

	var androidSettingsForState []AndroidSettings
	var errMsg string

	stateAndroidSettings := util.ObjectSetToTypedArray[AndroidSettings](ctx, diagnostics, androidSettings)

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

	return util.TypedArrayToObjectSet[AndroidSettings](ctx, diagnostics, androidSettingsForState)
}

func getChromeosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, chromeosSettings types.Set, remoteChromeosSettings []globalappconfiguration.CategorySettings) types.Set {

	var chromeosSettingsForState []ChromeosSettings
	var errMsg string

	stateChromeosSettings := util.ObjectSetToTypedArray[ChromeosSettings](ctx, diagnostics, chromeosSettings)

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

	return util.TypedArrayToObjectSet[ChromeosSettings](ctx, diagnostics, chromeosSettingsForState)
}

func getHtml5CategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, html5Settings types.Set, remoteHtml5Settings []globalappconfiguration.CategorySettings) types.Set {

	var html5SettingsForState []Html5Settings
	var errMsg string

	stateHtml5Settings := util.ObjectSetToTypedArray[Html5Settings](ctx, diagnostics, html5Settings)

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

	return util.TypedArrayToObjectSet[Html5Settings](ctx, diagnostics, html5SettingsForState)
}

func getMacosCategorySettings(ctx context.Context, diagnostics *diag.Diagnostics, macosSettings types.Set, remoteMacosSettings []globalappconfiguration.CategorySettings) types.Set {

	var macosSettingsForState []MacosSettings
	var errMsg string

	stateMacosSettings := util.ObjectSetToTypedArray[MacosSettings](ctx, diagnostics, macosSettings)

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
		MacosSettingsDefaultValues(ctx, diagnostics, &macosSetting) // Set the default list null for macos settings
		macosSetting.Name = stateMacosSetting.Name                  // Since this value is present as the map key, it is same as remote
		valueType := reflect.TypeOf(remoteMacosSetting.Value)
		switch valueType.Kind() {
		case reflect.String:
			macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
		case reflect.Slice:
			// Check if the value is of type LocalAppAllowList
			autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if autoLaunchProtocolsList != nil {
				macosSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
				break
			}

			// Check if the value is of type BookMarkValue
			managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if managedBookmarkList != nil {
				macosSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
				break
			}
			// Check if the value is of type ExtensionInstallAllowList
			extensionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
			if extensionInstallAllowList != nil {
				macosSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extensionInstallAllowList)
				break
			}

			macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))

		case reflect.Map:
			v := remoteMacosSetting.Value.(map[string]interface{})
			elementJSON, err := json.Marshal(v)
			if err != nil {
				errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var app CitrixEnterpriseBrowserModel_Go
			if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for the conversion
				errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
				break
			}
			var browserModel CitrixEnterpriseBrowserModel
			browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
			browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
			macosSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)

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
			MacosSettingsDefaultValues(ctx, diagnostics, &macosSetting) // Set the default list null for macos settings
			valueType := reflect.TypeOf(remoteMacosSetting.Value)
			switch valueType.Kind() {
			case reflect.String:
				macosSetting.ValueString = types.StringValue(remoteMacosSetting.Value.(string))
			case reflect.Slice:
				// Check if the value is of type LocalAppAllowList
				autoLaunchProtocolsList := GACSettingsUpdate[AutoLaunchProtocolsFromOriginsModel, AutoLaunchProtocolsFromOriginsModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
				if autoLaunchProtocolsList != nil {
					macosSetting.AutoLaunchProtocolsFromOrigins = util.TypedArrayToObjectSet(ctx, diagnostics, autoLaunchProtocolsList)
					break
				}
				// Check if the value is of type BookMarkValue
				managedBookmarkList := GACSettingsUpdate[BookMarkValueModel, BookMarkValueModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
				if managedBookmarkList != nil {
					macosSetting.ManagedBookmarks = util.TypedArrayToObjectSet(ctx, diagnostics, managedBookmarkList)
					break
				}
				// Check if the value is of type ExtensionInstallAllowList
				extensionInstallAllowList := GACSettingsUpdate[ExtensionInstallAllowListModel, ExtensionInstallAllowListModel_Go](ctx, diagnostics, remoteMacosSetting.Value)
				if extensionInstallAllowList != nil {
					macosSetting.ExtensionInstallAllowList = util.TypedArrayToObjectSet(ctx, diagnostics, extensionInstallAllowList)
					break
				}
				macosSetting.ValueList, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, remoteMacosSetting.Value.([]interface{}))

			case reflect.Map:
				v := remoteMacosSetting.Value.(map[string]interface{})
				elementJSON, err := json.Marshal(v)
				if err != nil {
					errMsg = fmt.Sprintf("Error marshaling element to CitrixEnterpriseBrowserModel: %v", err)
					break
				}
				var app CitrixEnterpriseBrowserModel_Go
				if err := json.Unmarshal(elementJSON, &app); err != nil { // marshal and unmarshal for conversion
					errMsg = fmt.Sprintf("Error unmarshaling element to CitrixEnterpriseBrowserModel: %v", err)
					break
				}
				var browserModel CitrixEnterpriseBrowserModel
				browserModel.CitrixEnterpriseBrowserSSOEnabled = types.BoolValue(app.CitrixEnterpriseBrowserSSOEnabled)
				browserModel.CitrixEnterpriseBrowserSSODomains, errMsg = util.ConvertPrimitiveInterfaceArrayToStringList(ctx, diagnostics, v["CitrixEnterpriseBrowserSSODomains"].([]interface{}))
				macosSetting.EnterpriseBroswerSSO = util.TypedObjectToObjectValue(ctx, diagnostics, browserModel)
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

	return util.TypedArrayToObjectSet[MacosSettings](ctx, diagnostics, macosSettingsForState)
}

func (GACSettingsResourceModel) GetSchema() schema.Schema {
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
			"test_channel": schema.BoolAttribute{
				Description: "Defines whether to use the test channel for settings or not. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
