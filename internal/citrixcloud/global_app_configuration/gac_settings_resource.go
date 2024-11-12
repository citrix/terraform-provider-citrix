// Copyright © 2024. Citrix Systems, Inc.

package global_app_configuration

import (
	"context"
	b64 "encoding/base64"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &gacSettingsResource{}
	_ resource.ResourceWithConfigure      = &gacSettingsResource{}
	_ resource.ResourceWithImportState    = &gacSettingsResource{}
	_ resource.ResourceWithModifyPlan     = &gacSettingsResource{}
	_ resource.ResourceWithValidateConfig = &gacSettingsResource{}
)

// NewGacSettingsResource is a helper function to simplify the provider implementation.
func NewGacSettingsResource() resource.Resource {
	return &gacSettingsResource{}
}

// gacSettingsResource is the resource implementation.
type gacSettingsResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *gacSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gac_settings"
}

// Schema defines the schema for the resource.
func (r *gacSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GACSettingsResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *gacSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *gacSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GACSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUrlModel globalappconfiguration.ServiceURL
	serviceUrlModel.SetUrl(plan.ServiceUrl.ValueString())

	planAppSettings := util.ObjectValueToTypedObject[AppSettings](ctx, &resp.Diagnostics, plan.AppSettings)

	var appSettings globalappconfiguration.AppSettings
	appSettings.SetAndroid(GetAppSettingsForAndroid(ctx, &resp.Diagnostics, planAppSettings.Android))
	appSettings.SetChromeos(GetAppSettingsForChromeos(ctx, &resp.Diagnostics, planAppSettings.Chromeos))
	appSettings.SetHtml5(GetAppSettingsForHtml5(ctx, &resp.Diagnostics, planAppSettings.Html5))
	appSettings.SetIos(GetAppSettingsForIos(ctx, &resp.Diagnostics, planAppSettings.Ios))
	appSettings.SetMacos(GetAppSettingsForMacos(ctx, &resp.Diagnostics, planAppSettings.Macos))
	appSettings.SetLinux(GetAppSettingsForLinux(ctx, &resp.Diagnostics, planAppSettings.Linux))
	appSettings.SetWindows(GetAppSettingsForWindows(ctx, &resp.Diagnostics, planAppSettings.Windows))

	var settings globalappconfiguration.Settings
	settings.SetName(plan.Name.ValueString())
	settings.SetDescription(plan.Description.ValueString())
	settings.SetUseForAppConfig(plan.UseForAppConfig.ValueBool())
	settings.SetAppSettings(appSettings)

	var body globalappconfiguration.SettingsRecordModel

	body.SetServiceURL(serviceUrlModel)
	body.SetSettings(settings)

	// Call the API
	createSettingsRequest := r.client.GacClient.SettingsControllerDAAS.PostSettingsApiUsingPOST(ctx, util.GacAppName)
	createSettingsRequest = createSettingsRequest.SettingsRecord(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(createSettingsRequest, r.client).Execute()

	//In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Settings configuration for ServiceUrl: "+plan.ServiceUrl.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

	//Try to get the new settings configuration from remote
	settingsConfiguration, err := getSettingsConfiguration(ctx, r.client, &resp.Diagnostics, plan.ServiceUrl.ValueString())
	if err != nil {
		return
	}

	//Set the new state
	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, settingsConfiguration.GetItems()[0])

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *gacSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state GACSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get Service Url settings from remote
	settingsConfiguration, err := readSettingsConfiguration(ctx, r.client, resp, state.ServiceUrl.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, settingsConfiguration.GetItems()[0])

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gacSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GACSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serviceUrlModel globalappconfiguration.ServiceURL
	serviceUrlModel.SetUrl(plan.ServiceUrl.ValueString())

	planAppSettings := util.ObjectValueToTypedObject[AppSettings](ctx, &resp.Diagnostics, plan.AppSettings)

	var appSettings globalappconfiguration.AppSettings
	appSettings.SetAndroid(GetAppSettingsForAndroid(ctx, &resp.Diagnostics, planAppSettings.Android))
	appSettings.SetChromeos(GetAppSettingsForChromeos(ctx, &resp.Diagnostics, planAppSettings.Chromeos))
	appSettings.SetHtml5(GetAppSettingsForHtml5(ctx, &resp.Diagnostics, planAppSettings.Html5))
	appSettings.SetIos(GetAppSettingsForIos(ctx, &resp.Diagnostics, planAppSettings.Ios))
	appSettings.SetMacos(GetAppSettingsForMacos(ctx, &resp.Diagnostics, planAppSettings.Macos))
	appSettings.SetWindows(GetAppSettingsForWindows(ctx, &resp.Diagnostics, planAppSettings.Windows))
	appSettings.SetLinux(GetAppSettingsForLinux(ctx, &resp.Diagnostics, planAppSettings.Linux))

	var settings globalappconfiguration.Settings
	settings.SetName(plan.Name.ValueString())
	settings.SetDescription(plan.Description.ValueString())
	settings.SetUseForAppConfig(plan.UseForAppConfig.ValueBool())
	settings.SetAppSettings(appSettings)

	var body globalappconfiguration.SettingsRecordModel

	body.SetServiceURL(serviceUrlModel)
	body.SetSettings(settings)

	// Call the API
	updateSettingsRequest := r.client.GacClient.SettingsControllerDAAS.PutSettingsApiUsingPUT(ctx, util.GacAppName, b64.StdEncoding.EncodeToString([]byte(plan.ServiceUrl.ValueString())))
	updateSettingsRequest = updateSettingsRequest.SettingsRecord(body)
	httpResp, err := citrixdaasclient.AddRequestData(updateSettingsRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating settings configuration for service url: "+plan.ServiceUrl.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

	// Try to get Service Url settings from remote
	updatedSettingsConfiguration, err := getSettingsConfiguration(ctx, r.client, &resp.Diagnostics, plan.ServiceUrl.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedSettingsConfiguration.GetItems()[0])

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gacSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state GACSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Delete settings configuration for the service url
	encodedServiceUrl := b64.StdEncoding.EncodeToString([]byte(state.ServiceUrl.ValueString()))
	deleteSettingsRequest := r.client.GacClient.SettingsControllerDAAS.DeleteSettingsApiUsingDELETE(ctx, util.GacAppName, encodedServiceUrl)
	httpResp, err := citrixdaasclient.AddRequestData(deleteSettingsRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting settings configuration for service url: "+state.ServiceUrl.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

}

func (r *gacSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("service_url"), req, resp)
}

func getSettingsConfiguration(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, serviceUrl string) (*globalappconfiguration.GetAllSettingResponse, error) {
	encodedServiceUrl := b64.StdEncoding.EncodeToString([]byte(serviceUrl))
	getSettingsRequest := client.GacClient.SettingsControllerDAAS.GetSettingsApiUsingGET(ctx, util.GacAppName, encodedServiceUrl)
	getSettingsResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*globalappconfiguration.GetAllSettingResponse](getSettingsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error fetching settings configuration for service url: "+serviceUrl,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
	}

	return getSettingsResponse, nil
}

func readSettingsConfiguration(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, serviceUrl string) (*globalappconfiguration.GetAllSettingResponse, error) {
	encodedServiceUrl := b64.StdEncoding.EncodeToString([]byte(serviceUrl))
	getSettingsRequest := client.GacClient.SettingsControllerDAAS.GetSettingsApiUsingGET(ctx, util.GacAppName, encodedServiceUrl)
	getSettingsResponse, _, err := util.ReadResource[*globalappconfiguration.GetAllSettingResponse](getSettingsRequest, ctx, client, resp, "ServiceUrl Settings Configuration", serviceUrl)
	return getSettingsResponse, err
}

func GetAppSettingsForWindows(ctx context.Context, diagnostics *diag.Diagnostics, windowsList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	windows := util.ObjectSetToTypedArray[Windows](ctx, diagnostics, windowsList)

	for _, windowsInstance := range windows {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(windowsInstance.Category.ValueString())
		platformSetting.SetUserOverride(windowsInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForWindows(ctx, diagnostics, windowsInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForLinux(ctx context.Context, diagnostics *diag.Diagnostics, linuxList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	linux := util.ObjectSetToTypedArray[Linux](ctx, diagnostics, linuxList)

	for _, linuxInstance := range linux {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(linuxInstance.Category.ValueString())
		platformSetting.SetUserOverride(linuxInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForLinux(ctx, diagnostics, linuxInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForIos(ctx context.Context, diagnostics *diag.Diagnostics, iosList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	ios := util.ObjectSetToTypedArray[Ios](ctx, diagnostics, iosList)

	for _, iosInstance := range ios {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(iosInstance.Category.ValueString())
		platformSetting.SetUserOverride(iosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForIos(ctx, diagnostics, iosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForAndroid(ctx context.Context, diagnostics *diag.Diagnostics, androidList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	android := util.ObjectSetToTypedArray[Android](ctx, diagnostics, androidList)

	for _, androidInstance := range android {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(androidInstance.Category.ValueString())
		platformSetting.SetUserOverride(androidInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForAndroid(ctx, diagnostics, androidInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForChromeos(ctx context.Context, diagnostics *diag.Diagnostics, chromeosList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	chromeos := util.ObjectSetToTypedArray[Chromeos](ctx, diagnostics, chromeosList)

	for _, chromeosInstance := range chromeos {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(chromeosInstance.Category.ValueString())
		platformSetting.SetUserOverride(chromeosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForChromeos(ctx, diagnostics, chromeosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForHtml5(ctx context.Context, diagnostics *diag.Diagnostics, html5List types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	html5 := util.ObjectSetToTypedArray[Html5](ctx, diagnostics, html5List)

	for _, html5Instance := range html5 {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(html5Instance.Category.ValueString())
		platformSetting.SetUserOverride(html5Instance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForHtml5(ctx, diagnostics, html5Instance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForMacos(ctx context.Context, diagnostics *diag.Diagnostics, macosList types.Set) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	macos := util.ObjectSetToTypedArray[Macos](ctx, diagnostics, macosList)

	for _, macosInstance := range macos {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(macosInstance.Category.ValueString())
		platformSetting.SetUserOverride(macosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForMacos(ctx, diagnostics, macosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func CreateCategorySettingsForWindows(ctx context.Context, diagnostics *diag.Diagnostics, windowsSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	windowsSettings := util.ObjectListToTypedArray[WindowsSettings](ctx, diagnostics, windowsSettingsList)

	for _, windowsSetting := range windowsSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(windowsSetting.Name.ValueString())
		if !windowsSetting.ValueString.IsNull() {
			categorySetting.SetValue(windowsSetting.ValueString.ValueString())
		} else if len(windowsSetting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, windowsSetting.ValueList))
		} else if len(windowsSetting.LocalAppAllowList.Elements()) > 0 {
			var localAppAllowListGo []LocalAppAllowListModel_Go
			localAppAllowList := util.ObjectListToTypedArray[LocalAppAllowListModel](ctx, diagnostics, windowsSetting.LocalAppAllowList)
			for _, localApp := range localAppAllowList {
				var localAppAllowListGoItem LocalAppAllowListModel_Go
				ConvertStruct(localApp, &localAppAllowListGoItem)
				localAppAllowListGo = append(localAppAllowListGo, localAppAllowListGoItem)
			}
			categorySetting.SetValue(localAppAllowListGo)
		} else if !windowsSetting.EnterpriseBroswerSSO.IsNull() {
			var enterpriseBrowserSSO CitrixEnterpriseBrowserModel_Go
			browserModel := util.ObjectValueToTypedObject[CitrixEnterpriseBrowserModel](ctx, diagnostics, windowsSetting.EnterpriseBroswerSSO)
			domains := util.StringListToStringArray(ctx, diagnostics, browserModel.CitrixEnterpriseBrowserSSODomains)
			enterpriseBrowserSSO.CitrixEnterpriseBrowserSSOEnabled = browserModel.CitrixEnterpriseBrowserSSOEnabled.ValueBool()
			enterpriseBrowserSSO.CitrixEnterpriseBrowserSSODomains = domains
			categorySetting.SetValue(enterpriseBrowserSSO)
		} else if !windowsSetting.ExtensionInstallAllowList.IsNull() {
			var extensionInstallAllowListGo []ExtensionInstallAllowListModel_Go
			extensionInstallAllowList := util.ObjectListToTypedArray[ExtensionInstallAllowListModel](ctx, diagnostics, windowsSetting.ExtensionInstallAllowList)
			for _, extensionInstall := range extensionInstallAllowList {
				var extensionInstallAllowListGoItem ExtensionInstallAllowListModel_Go
				ConvertStruct(extensionInstall, &extensionInstallAllowListGoItem)
				extensionInstallAllowListGo = append(extensionInstallAllowListGo, extensionInstallAllowListGoItem)
			}
			categorySetting.SetValue(extensionInstallAllowListGo)
		} else if !windowsSetting.AutoLaunchProtocolsFromOrigins.IsNull() {
			var autoLaunchProtocolsFromOriginsGo []AutoLaunchProtocolsFromOriginsModel_Go
			autoLaunchProtocolsFromOriginsList := util.ObjectListToTypedArray[AutoLaunchProtocolsFromOriginsModel](ctx, diagnostics, windowsSetting.AutoLaunchProtocolsFromOrigins)
			for _, autoLaunchProtocolsFromOrigins := range autoLaunchProtocolsFromOriginsList {
				var autoLaunchProtocolItem AutoLaunchProtocolsFromOriginsModel_Go
				autoLaunchProtocolItem.Protocol = autoLaunchProtocolsFromOrigins.Protocol.ValueString()
				if !autoLaunchProtocolsFromOrigins.AllowedOrigins.IsNull() {
					autoLaunchProtocolItem.AllowedOrigins = util.StringListToStringArray(ctx, diagnostics, autoLaunchProtocolsFromOrigins.AllowedOrigins)
				}
				autoLaunchProtocolsFromOriginsGo = append(autoLaunchProtocolsFromOriginsGo, autoLaunchProtocolItem)
			}
			categorySetting.SetValue(autoLaunchProtocolsFromOriginsGo)
		} else if !windowsSetting.ManagedBookmarks.IsNull() {
			var managedBookmarksGo []BookMarkValueModel_Go
			managedBookmarksList := util.ObjectListToTypedArray[BookMarkValueModel](ctx, diagnostics, windowsSetting.ManagedBookmarks)
			for _, managedBookmark := range managedBookmarksList {
				var managedBookmarkItem BookMarkValueModel_Go
				managedBookmarkItem.Name = managedBookmark.Name.ValueString()
				managedBookmarkItem.Url = managedBookmark.Url.ValueString()
				managedBookmarksGo = append(managedBookmarksGo, managedBookmarkItem)
			}
			categorySetting.SetValue(managedBookmarksGo)
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForLinux(ctx context.Context, diagnostics *diag.Diagnostics, linuxSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	linuxSettings := util.ObjectListToTypedArray[LinuxSettings](ctx, diagnostics, linuxSettingsList)

	for _, linuxSetting := range linuxSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(linuxSetting.Name.ValueString())
		if !linuxSetting.ValueString.IsNull() {
			categorySetting.SetValue(linuxSetting.ValueString.ValueString())
		} else if len(linuxSetting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, linuxSetting.ValueList))
		} else if !linuxSetting.ExtensionInstallAllowList.IsNull() {
			var extensionInstallAllowListGo []ExtensionInstallAllowListModel_Go
			extensionInstallAllowList := util.ObjectListToTypedArray[ExtensionInstallAllowListModel](ctx, diagnostics, linuxSetting.ExtensionInstallAllowList)
			for _, extensionInstall := range extensionInstallAllowList {
				var extensionInstallAllowListGoItem ExtensionInstallAllowListModel_Go
				ConvertStruct(extensionInstall, &extensionInstallAllowListGoItem)
				extensionInstallAllowListGo = append(extensionInstallAllowListGo, extensionInstallAllowListGoItem)
			}
			categorySetting.SetValue(extensionInstallAllowListGo)
		} else if !linuxSetting.AutoLaunchProtocolsFromOrigins.IsNull() {
			var autoLaunchProtocolsFromOriginsGo []AutoLaunchProtocolsFromOriginsModel_Go
			autoLaunchProtocolsFromOriginsList := util.ObjectListToTypedArray[AutoLaunchProtocolsFromOriginsModel](ctx, diagnostics, linuxSetting.AutoLaunchProtocolsFromOrigins)
			for _, autoLaunchProtocolsFromOrigins := range autoLaunchProtocolsFromOriginsList {
				var autoLaunchProtocolItem AutoLaunchProtocolsFromOriginsModel_Go
				autoLaunchProtocolItem.Protocol = autoLaunchProtocolsFromOrigins.Protocol.ValueString()
				if !autoLaunchProtocolsFromOrigins.AllowedOrigins.IsNull() {
					autoLaunchProtocolItem.AllowedOrigins = util.StringListToStringArray(ctx, diagnostics, autoLaunchProtocolsFromOrigins.AllowedOrigins)
				}
				autoLaunchProtocolsFromOriginsGo = append(autoLaunchProtocolsFromOriginsGo, autoLaunchProtocolItem)
			}
			categorySetting.SetValue(autoLaunchProtocolsFromOriginsGo)
		} else if !linuxSetting.ManagedBookmarks.IsNull() {
			var managedBookmarksGo []BookMarkValueModel_Go
			managedBookmarksList := util.ObjectListToTypedArray[BookMarkValueModel](ctx, diagnostics, linuxSetting.ManagedBookmarks)
			for _, managedBookmark := range managedBookmarksList {
				var managedBookmarkItem BookMarkValueModel_Go
				managedBookmarkItem.Name = managedBookmark.Name.ValueString()
				managedBookmarkItem.Url = managedBookmark.Url.ValueString()
				managedBookmarksGo = append(managedBookmarksGo, managedBookmarkItem)
			}
			categorySetting.SetValue(managedBookmarksGo)
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForIos(ctx context.Context, diagnostics *diag.Diagnostics, iosSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	iosSettings := util.ObjectListToTypedArray[IosSettings](ctx, diagnostics, iosSettingsList)

	for _, iosSetting := range iosSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(iosSetting.Name.ValueString())
		if !iosSetting.ValueString.IsNull() {
			categorySetting.SetValue(iosSetting.ValueString.ValueString())
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForAndroid(ctx context.Context, diagnostics *diag.Diagnostics, androidSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	androidSettings := util.ObjectListToTypedArray[AndroidSettings](ctx, diagnostics, androidSettingsList)

	for _, androidSetting := range androidSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(androidSetting.Name.ValueString())
		if !androidSetting.ValueString.IsNull() {
			categorySetting.SetValue(androidSetting.ValueString.ValueString())
		} else if len(androidSetting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, androidSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForHtml5(ctx context.Context, diagnostics *diag.Diagnostics, html5SettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	html5Settings := util.ObjectListToTypedArray[Html5Settings](ctx, diagnostics, html5SettingsList)

	for _, html5Setting := range html5Settings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(html5Setting.Name.ValueString())
		if !html5Setting.ValueString.IsNull() {
			categorySetting.SetValue(html5Setting.ValueString.ValueString())
		} else if len(html5Setting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, html5Setting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForChromeos(ctx context.Context, diagnostics *diag.Diagnostics, chromeosSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	chromeosSettings := util.ObjectListToTypedArray[ChromeosSettings](ctx, diagnostics, chromeosSettingsList)

	for _, chromeosSetting := range chromeosSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(chromeosSetting.Name.ValueString())
		if !chromeosSetting.ValueString.IsNull() {
			categorySetting.SetValue(chromeosSetting.ValueString.ValueString())
		} else if len(chromeosSetting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, chromeosSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForMacos(ctx context.Context, diagnostics *diag.Diagnostics, macosSettingsList types.List) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	macosSettings := util.ObjectListToTypedArray[MacosSettings](ctx, diagnostics, macosSettingsList)

	for _, macosSetting := range macosSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(macosSetting.Name.ValueString())
		if !macosSetting.ValueString.IsNull() {
			categorySetting.SetValue(macosSetting.ValueString.ValueString())
		} else if len(macosSetting.ValueList.Elements()) > 0 {
			categorySetting.SetValue(util.StringListToStringArray(ctx, diagnostics, macosSetting.ValueList))
		} else if !macosSetting.AutoLaunchProtocolsFromOrigins.IsNull() {
			var autoLaunchProtocolsFromOriginsGo []AutoLaunchProtocolsFromOriginsModel_Go
			autoLaunchProtocolsFromOriginsList := util.ObjectListToTypedArray[AutoLaunchProtocolsFromOriginsModel](ctx, diagnostics, macosSetting.AutoLaunchProtocolsFromOrigins)
			for _, autoLaunchProtocolsFromOrigins := range autoLaunchProtocolsFromOriginsList {
				var autoLaunchProtocolItem AutoLaunchProtocolsFromOriginsModel_Go
				autoLaunchProtocolItem.Protocol = autoLaunchProtocolsFromOrigins.Protocol.ValueString()
				if !autoLaunchProtocolsFromOrigins.AllowedOrigins.IsNull() {
					autoLaunchProtocolItem.AllowedOrigins = util.StringListToStringArray(ctx, diagnostics, autoLaunchProtocolsFromOrigins.AllowedOrigins)
				}
				autoLaunchProtocolsFromOriginsGo = append(autoLaunchProtocolsFromOriginsGo, autoLaunchProtocolItem)
			}
			categorySetting.SetValue(autoLaunchProtocolsFromOriginsGo)
		} else if !macosSetting.EnterpriseBroswerSSO.IsNull() {
			var enterpriseBrowserSSO CitrixEnterpriseBrowserModel_Go
			browserModel := util.ObjectValueToTypedObject[CitrixEnterpriseBrowserModel](ctx, diagnostics, macosSetting.EnterpriseBroswerSSO)
			domains := util.StringListToStringArray(ctx, diagnostics, browserModel.CitrixEnterpriseBrowserSSODomains)
			enterpriseBrowserSSO.CitrixEnterpriseBrowserSSOEnabled = browserModel.CitrixEnterpriseBrowserSSOEnabled.ValueBool()
			enterpriseBrowserSSO.CitrixEnterpriseBrowserSSODomains = domains
			categorySetting.SetValue(enterpriseBrowserSSO)
		} else if !macosSetting.ExtensionInstallAllowList.IsNull() {
			var extensionInstallAllowListGo []ExtensionInstallAllowListModel_Go
			extensionInstallAllowList := util.ObjectListToTypedArray[ExtensionInstallAllowListModel](ctx, diagnostics, macosSetting.ExtensionInstallAllowList)
			for _, extensionInstall := range extensionInstallAllowList {
				var extensionInstallAllowListGoItem ExtensionInstallAllowListModel_Go
				ConvertStruct(extensionInstall, &extensionInstallAllowListGoItem)
				extensionInstallAllowListGo = append(extensionInstallAllowListGo, extensionInstallAllowListGoItem)
			}
			categorySetting.SetValue(extensionInstallAllowListGo)
		} else if !macosSetting.ManagedBookmarks.IsNull() {
			var managedBookmarksGo []BookMarkValueModel_Go
			managedBookmarksList := util.ObjectListToTypedArray[BookMarkValueModel](ctx, diagnostics, macosSetting.ManagedBookmarks)
			for _, managedBookmark := range managedBookmarksList {
				var managedBookmarkItem BookMarkValueModel_Go
				managedBookmarkItem.Name = managedBookmark.Name.ValueString()
				managedBookmarkItem.Url = managedBookmark.Url.ValueString()
				managedBookmarksGo = append(managedBookmarksGo, managedBookmarkItem)
			}
			categorySetting.SetValue(managedBookmarksGo)
		}

		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func (r *gacSettingsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.GacClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func (r *gacSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data GACSettingsResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the configuration for windows
	appSettings := util.ObjectValueToTypedObject[AppSettings](ctx, &resp.Diagnostics, data.AppSettings)
	if !appSettings.Windows.IsNull() && !appSettings.Windows.IsUnknown() {
		windowsList := util.ObjectSetToTypedArray[Windows](ctx, &resp.Diagnostics, appSettings.Windows)
		for _, windowsInstance := range windowsList {
			appSettings := util.ObjectListToTypedArray[WindowsSettings](ctx, &resp.Diagnostics, windowsInstance.Settings)
			for _, appSetting := range appSettings {
				if appSetting.LocalAppAllowList.IsNull() && appSetting.EnterpriseBroswerSSO.IsNull() && appSetting.ExtensionInstallAllowList.IsNull() && appSetting.ValueList.IsNull() && appSetting.ValueString.IsNull() && appSetting.AutoLaunchProtocolsFromOrigins.IsNull() && appSetting.ManagedBookmarks.IsNull() {
					resp.Diagnostics.AddError("Error in Windows Settings", "At least one value should be specified for Windows Settings")
				}
			}
		}
	}

	// Validate the configuration for macOs
	if !appSettings.Macos.IsNull() && !appSettings.Macos.IsUnknown() {
		macosList := util.ObjectSetToTypedArray[Macos](ctx, &resp.Diagnostics, appSettings.Macos)
		for _, macosInstance := range macosList {
			appSettings := util.ObjectListToTypedArray[MacosSettings](ctx, &resp.Diagnostics, macosInstance.Settings)
			for _, appSetting := range appSettings {
				if appSetting.EnterpriseBroswerSSO.IsNull() && appSetting.ExtensionInstallAllowList.IsNull() && appSetting.ValueList.IsNull() && appSetting.ValueString.IsNull() && appSetting.AutoLaunchProtocolsFromOrigins.IsNull() && appSetting.ManagedBookmarks.IsNull() {
					resp.Diagnostics.AddError("Error in MacOs Settings", "At least one value should be specified for MacOs Settings")
				}
			}
		}
	}

	// Validate the configuration for Linux
	if !appSettings.Linux.IsNull() && !appSettings.Linux.IsUnknown() {
		htmls5List := util.ObjectSetToTypedArray[Linux](ctx, &resp.Diagnostics, appSettings.Linux)
		for _, LinuxInstance := range htmls5List {
			appSettings := util.ObjectListToTypedArray[LinuxSettings](ctx, &resp.Diagnostics, LinuxInstance.Settings)
			for _, appSetting := range appSettings {
				if appSetting.ValueString.IsNull() && appSetting.ValueList.IsNull() && appSetting.AutoLaunchProtocolsFromOrigins.IsNull() && appSetting.ManagedBookmarks.IsNull() && appSetting.ExtensionInstallAllowList.IsNull() {
					resp.Diagnostics.AddError("Error in Linux Settings", "At least one value should be specified for Linux Settings")
				}
			}
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
