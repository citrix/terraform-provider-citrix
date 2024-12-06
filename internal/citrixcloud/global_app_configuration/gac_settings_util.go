// Copyright © 2024. Citrix Systems, Inc.

package global_app_configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// region BookMarkValueModel
type BookMarkValueModel struct {
	Name types.String `tfsdk:"name" json:"name"`
	Url  types.String `tfsdk:"url" json:"url"`
}

type BookMarkValueModel_Go struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (BookMarkValueModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for the bookmark",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL for the bookmark",
				Required:    true,
			},
		},
	}
}

func (BookMarkValueModel) GetAttributes() map[string]schema.Attribute {
	return BookMarkValueModel{}.GetSchema().Attributes
}

// endregion BookMarkValueModel

// region AutoLaunchProtocolsFromOriginsModel
type AutoLaunchProtocolsFromOriginsModel struct {
	Protocol       types.String `tfsdk:"protocol"`
	AllowedOrigins types.List   `tfsdk:"allowed_origins"`
}

type AutoLaunchProtocolsFromOriginsModel_Go struct {
	Protocol       string   `json:"protocol"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

func (AutoLaunchProtocolsFromOriginsModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{
				Description: "Auto launch protocol",
				Required:    true,
			},
			"allowed_origins": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of origins urls",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}

func (AutoLaunchProtocolsFromOriginsModel) GetAttributes() map[string]schema.Attribute {
	return AutoLaunchProtocolsFromOriginsModel{}.GetSchema().Attributes
}

// endregion AutoLaunchProtocolsFromOriginsModel

// region CitrixEnterpriseBrowserModel
type CitrixEnterpriseBrowserModel struct {
	CitrixEnterpriseBrowserSSOEnabled types.Bool `tfsdk:"citrix_enterprise_browser_sso_enabled"`
	CitrixEnterpriseBrowserSSODomains types.List `tfsdk:"citrix_enterprise_browser_sso_domains"`
}

type CitrixEnterpriseBrowserModel_Go struct {
	CitrixEnterpriseBrowserSSOEnabled bool     `json:"CitrixEnterpriseBrowserSSOEnabled"`
	CitrixEnterpriseBrowserSSODomains []string `json:"CitrixEnterpriseBrowserSSODomains"`
}

func (CitrixEnterpriseBrowserModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Enables Single Sign-on (SSO) for all the web and SaaS apps for the selected Operating System for the IdP domains added as long as the same IdP is used to sign in to the Citrix Workspace app and the relevant web or SaaS app.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"citrix_enterprise_browser_sso_enabled": schema.BoolAttribute{
				Description: "Enables Single Sign-on (SSO) for all the web and SaaS apps.",
				Required:    true,
			},
			"citrix_enterprise_browser_sso_domains": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of IdP domains for which SSO is enabled.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (CitrixEnterpriseBrowserModel) GetAttributes() map[string]schema.Attribute {
	return CitrixEnterpriseBrowserModel{}.GetSchema().Attributes
}

//endregion CitrixEnterpriseBrowserModel

// region LocalAppAllowListModel
type LocalAppAllowListModel struct {
	Name      types.String `tfsdk:"name"`
	Path      types.String `tfsdk:"path"`
	Arguments types.String `tfsdk:"arguments"`
}

type LocalAppAllowListModel_Go struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Arguments string `json:"arguments"`
}

func (LocalAppAllowListModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for Local App Discovery.",
				Required:    true,
			},
			"path": schema.StringAttribute{
				Description: "Path for Local App Discovery.",
				Required:    true,
			},
			"arguments": schema.StringAttribute{
				Description: "Arguments for Local App Discovery.",
				Required:    true,
			},
		},
	}
}

func (LocalAppAllowListModel) GetAttributes() map[string]schema.Attribute {
	return LocalAppAllowListModel{}.GetSchema().Attributes
}

//endregion LocalAppAllowListModel

// region ExtensionInstallAllowListModel
type ExtensionInstallAllowListModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	InstallLink types.String `tfsdk:"install_link"`
}

type ExtensionInstallAllowListModel_Go struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	InstallLink string `json:"install link"`
}

func (ExtensionInstallAllowListModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the allowed extensions.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the allowed extensions.",
				Required:    true,
			},
			"install_link": schema.StringAttribute{
				Description: "Install link for the allowed extensions.",
				Required:    true,
			},
		},
	}
}

func (ExtensionInstallAllowListModel) GetAttributes() map[string]schema.Attribute {
	return ExtensionInstallAllowListModel{}.GetSchema().Attributes
}

//endregion ExtensionInstallAllowListModel

// Generic converter function
func ConvertStruct(src interface{}, dst interface{}) error {
	if src == nil || dst == nil {
		return fmt.Errorf("src or dst is nil")
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("dst must be a non-nil pointer")
	}

	dstElem := dstVal.Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		if !dstElem.IsValid() {
			return fmt.Errorf("destination element is invalid")
		}

		if !srcVal.IsValid() {
			return fmt.Errorf("source value is invalid")
		}

		fieldName := srcVal.Type().Field(i).Name
		dstField := dstElem.FieldByName(fieldName)

		if dstField.IsValid() && dstField.CanSet() {
			if srcField.Type() == reflect.TypeOf(types.String{}) && dstField.Type() == reflect.TypeOf("") {
				// Retrieve the method by name
				method := srcField.MethodByName("ValueString")
				// Call the method and get the first return value
				result := method.Call(nil)[0]
				// Convert the result to a string
				fmt.Println(result.String()) // Output: Hello, World!
				dstField.SetString(result.String())
			} else if srcField.Type().ConvertibleTo(dstField.Type()) {
				dstField.Set(srcField.Convert(dstField.Type()))
			}
		}
	}

	return nil
}

// Generic converter function to convert from src (with string fields) to dst (with types.String fields)
func ConvertStructReverse(src interface{}, dst interface{}) error {
	if src == nil || dst == nil {
		return fmt.Errorf("src or dst is nil")
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() != reflect.Ptr || srcVal.IsNil() {
		return fmt.Errorf("src must be a non-nil pointer")
	}

	srcElem := srcVal.Elem()
	dstElem := dstVal.Elem()

	for i := 0; i < srcElem.NumField(); i++ {
		srcField := srcElem.Field(i)
		fieldName := srcElem.Type().Field(i).Name
		dstField := dstElem.FieldByName(fieldName)

		if dstField.IsValid() && dstField.CanSet() {
			if srcField.Type() == reflect.TypeOf("") && dstField.Type() == reflect.TypeOf(types.String{}) {
				dstField.Set(reflect.ValueOf(types.StringValue(srcField.String())))
			} else if srcField.Type() == reflect.TypeOf([]string{}) && dstField.Type() == reflect.TypeOf(types.List{}) {
				// Convert []string to types.List
				stringList := srcField.Interface().([]string)
				listValue := make([]attr.Value, len(stringList))
				for i, v := range stringList {
					listValue[i] = types.StringValue(v)
				}
				listVal, diags := types.ListValue(types.StringType, listValue)
				if diags.HasError() {
					return fmt.Errorf("error creating ListValue: %v", diags)
				}
				dstField.Set(reflect.ValueOf(listVal))
			} else if srcField.Type().ConvertibleTo(dstField.Type()) {
				dstField.Set(srcField.Convert(dstField.Type()))
			}
		}
	}

	return nil
}

func GACSettingsUpdate[tfType any, goType any](ctx context.Context, diagnostics *diag.Diagnostics, ls interface{}) []tfType {
	listGoType := make([]goType, 0, reflect.ValueOf(ls).Len())
	sliceValue := reflect.ValueOf(ls)
	for i := 0; i < sliceValue.Len(); i++ {
		element := sliceValue.Index(i).Interface()
		goTypeInstance := new(goType)
		mapElement, _ := element.(map[string]interface{})
		if i == 0 && !verifyStruct(mapElement, goTypeInstance) { //only verify once the struct
			return nil
		}
		elementJSON, _ := json.Marshal(element)
		var app goType
		if err := json.Unmarshal(elementJSON, &app); err != nil {
			return nil
		}
		listGoType = append(listGoType, app)
	}
	listTFType := make([]tfType, 0, reflect.ValueOf(ls).Len())
	if len(listGoType) > 0 {
		for _, app := range listGoType {
			var allowListItem tfType
			ConvertStructReverse(&app, &allowListItem)
			listTFType = append(listTFType, allowListItem)
		}
	}
	return listTFType
}

func verifyStruct(mapData map[string]interface{}, goType interface{}) bool {
	val := reflect.ValueOf(goType).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}
		mapValue, ok := mapData[fieldName]
		if !ok {
			fmt.Printf("Field %s not found in map\n", fieldName)
			return false
		}
		if fieldName == "allowedOrigins" && field.Type == reflect.TypeOf([]string{}) && reflect.TypeOf(mapValue) == reflect.TypeOf([]interface{}{}) {
			continue
		}
		if reflect.TypeOf(mapValue) != field.Type {
			fmt.Printf("Type mismatch for field %s: expected %s, got %s\n", fieldName, field.Type, reflect.TypeOf(mapValue))
			return false
		}
	}
	return true
}

func WindowsSettingsDefaultValues(ctx context.Context, diagnostics *diag.Diagnostics, windowsSetting *WindowsSettings) {
	windowsSetting.ValueList = types.ListNull(types.StringType)
	//setting null for LocalAppAllowList
	localAppAllowListAttributesMap, err := util.ResourceAttributeMapFromObject(LocalAppAllowListModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	windowsSetting.LocalAppAllowList = types.SetNull(types.ObjectType{AttrTypes: localAppAllowListAttributesMap})

	//setting null for ExtensionInstallAllowList
	installAllowListAttributesMap, err := util.ResourceAttributeMapFromObject(ExtensionInstallAllowListModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	windowsSetting.ExtensionInstallAllowList = types.SetNull(types.ObjectType{AttrTypes: installAllowListAttributesMap})

	//setting null for EnterpriseBroswerSSO
	enterpriseBrowserAttributesMap, err := util.ResourceAttributeMapFromObject(CitrixEnterpriseBrowserModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	windowsSetting.EnterpriseBroswerSSO = types.ObjectNull(enterpriseBrowserAttributesMap)

	//setting null for AutoLaunchProtocolsFromOrigins
	autoLaunchProtocolAttributesMap, err := util.ResourceAttributeMapFromObject(AutoLaunchProtocolsFromOriginsModel{AllowedOrigins: types.ListNull(types.StringType)})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	windowsSetting.AutoLaunchProtocolsFromOrigins = types.SetNull(types.ObjectType{AttrTypes: autoLaunchProtocolAttributesMap})

	//setting null for ManagedBookmarks
	bookMarkAttributesMap, err := util.ResourceAttributeMapFromObject(BookMarkValueModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	windowsSetting.ManagedBookmarks = types.SetNull(types.ObjectType{AttrTypes: bookMarkAttributesMap})
}

func MacosSettingsDefaultValues(ctx context.Context, diagnostics *diag.Diagnostics, macosSetting *MacosSettings) {
	macosSetting.ValueList = types.ListNull(types.StringType)
	//setting null for AutoLaunchProtocolsFromOrigins
	autoLaunchProtocolAttributesMap, err := util.ResourceAttributeMapFromObject(AutoLaunchProtocolsFromOriginsModel{AllowedOrigins: types.ListNull(types.StringType)})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	macosSetting.AutoLaunchProtocolsFromOrigins = types.SetNull(types.ObjectType{AttrTypes: autoLaunchProtocolAttributesMap})

	//setting null for ManagedBookmarks
	bookMarkAttributesMap, err := util.ResourceAttributeMapFromObject(BookMarkValueModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	macosSetting.ManagedBookmarks = types.SetNull(types.ObjectType{AttrTypes: bookMarkAttributesMap})

	//setting null for ExtensionInstallAllowList
	installAllowListAttributesMap, err := util.ResourceAttributeMapFromObject(ExtensionInstallAllowListModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	macosSetting.ExtensionInstallAllowList = types.SetNull(types.ObjectType{AttrTypes: installAllowListAttributesMap})

	//setting null for EnterpriseBroswerSSO
	enterpriseBrowserAttributesMap, err := util.ResourceAttributeMapFromObject(CitrixEnterpriseBrowserModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	macosSetting.EnterpriseBroswerSSO = types.ObjectNull(enterpriseBrowserAttributesMap)
}

func LinuxSettingsDefaultValues(ctx context.Context, diagnostics *diag.Diagnostics, linuxSetting *LinuxSettings) {
	linuxSetting.ValueList = types.ListNull(types.StringType)
	//setting null for AutoLaunchProtocolsFromOrigins
	autoLaunchProtocolAttributesMap, err := util.ResourceAttributeMapFromObject(AutoLaunchProtocolsFromOriginsModel{AllowedOrigins: types.ListNull(types.StringType)})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	linuxSetting.AutoLaunchProtocolsFromOrigins = types.SetNull(types.ObjectType{AttrTypes: autoLaunchProtocolAttributesMap})

	//setting null for ManagedBookmarks
	bookMarkAttributesMap, err := util.ResourceAttributeMapFromObject(BookMarkValueModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	linuxSetting.ManagedBookmarks = types.SetNull(types.ObjectType{AttrTypes: bookMarkAttributesMap})

	//setting null for ExtensionInstallAllowList
	installAllowListAttributesMap, err := util.ResourceAttributeMapFromObject(ExtensionInstallAllowListModel{})
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return
	}
	linuxSetting.ExtensionInstallAllowList = types.SetNull(types.ObjectType{AttrTypes: installAllowListAttributesMap})

}
