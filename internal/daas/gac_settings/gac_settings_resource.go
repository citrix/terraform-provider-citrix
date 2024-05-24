// Copyright © 2023. Citrix Systems, Inc.

package gac_settings

import (
	"context"
	b64 "encoding/base64"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &gacSettingsResource{}
	_ resource.ResourceWithConfigure   = &gacSettingsResource{}
	_ resource.ResourceWithImportState = &gacSettingsResource{}
)

// NewAdminUserResource is a helper function to simplify the provider implementation.
func NewAGacSettingsResource() resource.Resource {
	return &gacSettingsResource{}
}

// adminUserResource is the resource implementation.
type gacSettingsResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *gacSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gac_settings"
}

// Schema defines the schema for the resource.
func (r *gacSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Global App Configuration settings for a service url.",

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
			"app_settings": schema.SingleNestedAttribute{
				Description: "Defines the device platform and the associated settings. Currently, only settings objects with value type of integer, boolean, strings and list of strings is supported.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"windows": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using windows platform.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"ios": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using ios platform.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"android": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using android platform.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"html5": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using html5.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"chromeos": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using chrome os platform.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"macos": schema.ListNestedAttribute{
						Description: "Settings to be applied for users using mac os platform.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
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
									NestedObject: schema.NestedAttributeObject{
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
									},
								},
							},
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
				},
			},
		},
	}
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

	var appSettings globalappconfiguration.AppSettings
	appSettings.SetAndroid(GetAppSettingsForAndroid(plan.AppSettings.Android))
	appSettings.SetChromeos(GetAppSettingsForChromeos(plan.AppSettings.Chromeos))
	appSettings.SetHtml5(GetAppSettingsForHtml5(plan.AppSettings.Html5))
	appSettings.SetIos(GetAppSettingsForIos(plan.AppSettings.Ios))
	appSettings.SetMacos(GetAppSettingsForMacos(plan.AppSettings.Macos))
	appSettings.SetWindows(GetAppSettingsForWindows(plan.AppSettings.Windows))

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
				"\nError message: "+util.ReadClientError(err),
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
	plan = plan.RefreshPropertyValues(settingsConfiguration.GetItems()[0], &resp.Diagnostics)

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

	state = state.RefreshPropertyValues(settingsConfiguration.GetItems()[0], &resp.Diagnostics)

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

	var appSettings globalappconfiguration.AppSettings
	appSettings.SetAndroid(GetAppSettingsForAndroid(plan.AppSettings.Android))
	appSettings.SetChromeos(GetAppSettingsForChromeos(plan.AppSettings.Chromeos))
	appSettings.SetHtml5(GetAppSettingsForHtml5(plan.AppSettings.Html5))
	appSettings.SetIos(GetAppSettingsForIos(plan.AppSettings.Ios))
	appSettings.SetMacos(GetAppSettingsForMacos(plan.AppSettings.Macos))
	appSettings.SetWindows(GetAppSettingsForWindows(plan.AppSettings.Windows))

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
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Try to get Service Url settings from remote
	updatedSettingsConfiguration, err := getSettingsConfiguration(ctx, r.client, &resp.Diagnostics, plan.ServiceUrl.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(updatedSettingsConfiguration.GetItems()[0], &resp.Diagnostics)

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
				"\nError message: "+util.ReadClientError(err),
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
				"\nError message: "+util.ReadClientError(err),
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

func GetAppSettingsForWindows(windows []Windows) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, windowsInstance := range windows {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(windowsInstance.Category.ValueString())
		platformSetting.SetUserOverride(windowsInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForWindows(windowsInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForIos(ios []Ios) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, iosInstance := range ios {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(iosInstance.Category.ValueString())
		platformSetting.SetUserOverride(iosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForIos(iosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForAndroid(android []Android) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, androidInstance := range android {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(androidInstance.Category.ValueString())
		platformSetting.SetUserOverride(androidInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForAndroid(androidInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForChromeos(chromeos []Chromeos) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, chromeosInstance := range chromeos {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(chromeosInstance.Category.ValueString())
		platformSetting.SetUserOverride(chromeosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForChromeos(chromeosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForHtml5(html5 []Html5) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, html5Instance := range html5 {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(html5Instance.Category.ValueString())
		platformSetting.SetUserOverride(html5Instance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForHtml5(html5Instance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func GetAppSettingsForMacos(macos []Macos) []globalappconfiguration.PlatformSettings {
	var platformSettings []globalappconfiguration.PlatformSettings
	for _, macosInstance := range macos {
		var platformSetting globalappconfiguration.PlatformSettings
		platformSetting.SetCategory(macosInstance.Category.ValueString())
		platformSetting.SetUserOverride(macosInstance.UserOverride.ValueBool())
		platformSetting.SetAssignmentPriority(util.AssignmentPriority)
		platformSetting.SetAssignedTo(util.PlatformSettingsAssignedTo)
		platformSetting.SetSettings(CreateCategorySettingsForMacos(macosInstance.Settings))
		platformSettings = append(platformSettings, platformSetting)
	}
	return platformSettings
}

func CreateCategorySettingsForWindows(windowsSettings []WindowsSettings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	for _, windowsSetting := range windowsSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(windowsSetting.Name.ValueString())
		if !windowsSetting.ValueString.IsNull() {
			categorySetting.SetValue(windowsSetting.ValueString.ValueString())
		} else if len(windowsSetting.ValueList) > 0 {
			categorySetting.SetValue(util.ConvertBaseStringArrayToPrimitiveStringArray(windowsSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForIos(iosSettings []IosSettings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
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

func CreateCategorySettingsForAndroid(androidSettings []AndroidSettings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	for _, androidSetting := range androidSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(androidSetting.Name.ValueString())
		if !androidSetting.ValueString.IsNull() {
			categorySetting.SetValue(androidSetting.ValueString.ValueString())
		} else if len(androidSetting.ValueList) > 0 {
			categorySetting.SetValue(util.ConvertBaseStringArrayToPrimitiveStringArray(androidSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForHtml5(html5Settings []Html5Settings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	for _, html5Setting := range html5Settings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(html5Setting.Name.ValueString())
		if !html5Setting.ValueString.IsNull() {
			categorySetting.SetValue(html5Setting.ValueString.ValueString())
		} else if len(html5Setting.ValueList) > 0 {
			categorySetting.SetValue(util.ConvertBaseStringArrayToPrimitiveStringArray(html5Setting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForChromeos(chromeosSettings []ChromeosSettings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	for _, chromeosSetting := range chromeosSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(chromeosSetting.Name.ValueString())
		if !chromeosSetting.ValueString.IsNull() {
			categorySetting.SetValue(chromeosSetting.ValueString.ValueString())
		} else if len(chromeosSetting.ValueList) > 0 {
			categorySetting.SetValue(util.ConvertBaseStringArrayToPrimitiveStringArray(chromeosSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}

func CreateCategorySettingsForMacos(macosSettings []MacosSettings) []globalappconfiguration.CategorySettings {
	var categorySettings []globalappconfiguration.CategorySettings
	for _, macosSetting := range macosSettings {
		var categorySetting globalappconfiguration.CategorySettings

		categorySetting.SetName(macosSetting.Name.ValueString())
		if !macosSetting.ValueString.IsNull() {
			categorySetting.SetValue(macosSetting.ValueString.ValueString())
		} else if len(macosSetting.ValueList) > 0 {
			categorySetting.SetValue(util.ConvertBaseStringArrayToPrimitiveStringArray(macosSetting.ValueList))
		}
		categorySettings = append(categorySettings, categorySetting)
	}
	return categorySettings
}
