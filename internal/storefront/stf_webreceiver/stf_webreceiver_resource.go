// Copyright © 2024. Citrix Systems, Inc.
package stf_webreceiver

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfWebReceiverResource{}
	_ resource.ResourceWithConfigure   = &stfWebReceiverResource{}
	_ resource.ResourceWithImportState = &stfWebReceiverResource{}
	_ resource.ResourceWithModifyPlan  = &stfWebReceiverResource{}
)

// stfWebReceiverResource is a helper function to simplify the provider implementation.
func NewSTFWebReceiverResource() resource.Resource {
	return &stfWebReceiverResource{}
}

// stfWebReceiverResource is the resource implementation.
type stfWebReceiverResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfWebReceiverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_webreceiver_service"
}

// Configure adds the provider configured client to the resource.
func (r *stfWebReceiverResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// ModifyPlan modifies the resource plan before it is applied.
func (r *stfWebReceiverResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Skip modify plan when doing destroy action
	if req.Plan.Raw.IsNull() {
		return
	}

	operation := "updating"
	if req.State.Raw.IsNull() {
		operation = "creating"
	}

	// Retrieve values from plan
	var plan STFWebReceiverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webReceiverUI := util.ObjectValueToTypedObject[UserInterface](ctx, &resp.Diagnostics, plan.UserInterface)
	if !webReceiverUI.PreventIcaDownloads.IsUnknown() && !webReceiverUI.PreventIcaDownloads.IsNull() {
		stfSupportPreventIcaDownload := util.CheckStoreFrontVersion(r.client.StorefrontClient.VersionSF, ctx, &resp.Diagnostics, 3, 27)
		if !stfSupportPreventIcaDownload && webReceiverUI.PreventIcaDownloads.ValueBool() {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error %s StoreFront Web Receiver resource", operation),
				"`prevent_ica_download` cannot be set to `true` for the targeted StoreFront, StoreFront version 2402 or higher is required",
			)
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfWebReceiverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFWebReceiverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.CreateSTFWebReceiverRequestModel
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return
	}
	body.SetSiteId(siteIdInt)
	body.SetVirtualPath(plan.VirtualPath.ValueString())
	body.SetFriendlyName(plan.FriendlyName.ValueString())

	var getSTFStoreBody citrixstorefront.GetSTFStoreRequestModel
	getSTFStoreBody.SetVirtualPath(plan.StoreServiceVirtualPath.ValueString())
	createWebReceiverRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverCreateSTFWebReceiver(ctx, body, getSTFStoreBody)
	// Create new STF WebReceiver
	WebReceiverDetail, err := createWebReceiverRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront WebReceiver",
			"Error message: "+err.Error(),
		)
		return
	}

	var getWebReceiverRequestBody citrixstorefront.GetSTFWebReceiverRequestModel
	getWebReceiverRequestBody.SetVirtualPath(plan.VirtualPath.ValueString())
	if plan.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating StoreFront Authentication Service ",
				"Error message: "+err.Error(),
			)
			return
		}
		getWebReceiverRequestBody.SetSiteId(siteIdInt)
	}

	// Create the authentication methods Body
	if !plan.AuthenticationMethods.IsNull() {
		var authMethodCreateBody citrixstorefront.UpdateSTFWebReceiverAuthenticationMethodsRequestModel
		authMethods := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.AuthenticationMethods)
		authMethodCreateBody.SetAuthenticationMethods(authMethods)

		creatAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverSetSTFWebReceiverAuthenticationMethods(ctx, authMethodCreateBody, getWebReceiverRequestBody)
		// Create new STF WebReceiver Authentication Methods
		_, err = creatAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating StoreFront WebReceiver Authentication Methods",
				"Error message: "+err.Error(),
			)
			return
		}
	}

	// Create the Plugin Assistant
	if !plan.PluginAssistant.IsNull() {
		var pluginAssistantBody citrixstorefront.UpdateSTFWebReceiverPluginAssistantRequestModel

		plannedPluginAssistant := util.ObjectValueToTypedObject[PluginAssistant](ctx, &resp.Diagnostics, plan.PluginAssistant)
		pluginAssistantBody.SetEnabled(plannedPluginAssistant.Enabled.ValueBool())
		pluginAssistantBody.SetUpgradeAtLogin(plannedPluginAssistant.UpgradeAtLogin.ValueBool())
		pluginAssistantBody.SetShowAfterLogin(plannedPluginAssistant.ShowAfterLogin.ValueBool())
		pluginAssistantBody.SetWin32Path(plannedPluginAssistant.Win32Path.ValueString())
		pluginAssistantBody.SetMacOSPath(plannedPluginAssistant.MacOSPath.ValueString())
		pluginAssistantBody.SetMacOSMinimumSupportedVersion(plannedPluginAssistant.MacOSMinimumSupportedVersion.ValueString())
		pluginAssistantBody.SetHtml5SingleTabLaunch(plannedPluginAssistant.Html5SingleTabLaunch.ValueBool())
		pluginAssistantBody.SetHtml5Enabled(plannedPluginAssistant.Html5Enabled.ValueString())
		pluginAssistantBody.SetHtml5Platforms(plannedPluginAssistant.Html5Platforms.ValueString())
		pluginAssistantBody.SetHtml5Preferences(plannedPluginAssistant.Html5Preferences.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppOrigins(plannedPluginAssistant.Html5ChromeAppOrigins.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppPreferences(plannedPluginAssistant.Html5ChromeAppPreferences.ValueString())
		pluginAssistantBody.SetProtocolHandlerEnabled(plannedPluginAssistant.ProtocolHandlerEnabled.ValueBool())
		pluginAssistantBody.SetProtocolHandlerPlatforms(plannedPluginAssistant.ProtocolHandlerPlatforms.ValueString())
		pluginAssistantBody.SetProtocolHandlerSkipDoubleHopCheckWhenDisabled(plannedPluginAssistant.ProtocolHandlerSkipDoubleHopCheckWhenDisabled.ValueBool())
		pluginAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantUpdate(ctx, pluginAssistantBody, getWebReceiverRequestBody)
		// Create new STF WebReceiver Plugin Assistant
		_, err = pluginAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating StoreFront WebReceiver Plugin Assistant",
				"Error message: "+err.Error(),
			)
			return
		}

	}

	// Refresh the authentication methods
	if !plan.AuthenticationMethods.IsNull() {
		getAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverGetSTFWebReceiverAuthenticationMethods(ctx, getWebReceiverRequestBody)
		authMethoResult, err := getAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront WebReceiver Authentication Methods",
				"Error message: "+err.Error(),
			)
			return
		}

		plan.AuthenticationMethods = util.StringArrayToStringSet(ctx, &resp.Diagnostics, authMethoResult.Methods)
	}

	//Refresh Plugin Assistant
	if !plan.PluginAssistant.IsNull() {

		getPlugInAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantGet(ctx, getWebReceiverRequestBody)
		assistant, err := getPlugInAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront WebReceiver Plugin Assistant",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshPlugInAssistant(ctx, &resp.Diagnostics, &assistant)
	}

	var appShortcutsResponse citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel
	if !plan.ApplicationShortcuts.IsNull() {
		appShortcutsResponse, err = setAndGetSTFWebReceiverApplicationShortcuts(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.ApplicationShortcuts)
		if err != nil {
			return
		}
	}

	var communicationResponse citrixstorefront.GetWebReceiverCommunicationResponseModel
	if !plan.Communication.IsNull() {
		communicationResponse, err = setAndGetSTFWebReceiverCommunication(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.Communication)
		if err != nil {
			return
		}
	}

	var stsResponse citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel
	if !plan.StrictTransportSecurity.IsNull() {
		stsResponse, err = setAndGetSTFWebReceiverStrictTransportSecurity(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.StrictTransportSecurity)
		if err != nil {
			return
		}
	}

	var authManagerResponse citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel
	if !plan.AuthenticationManager.IsNull() {
		authManagerResponse, err = setAndGetSTFWebReceiverAuthenticationManager(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.AuthenticationManager)
		if err != nil {
			return
		}
	}

	// WebReceiverUserInterface config
	var uiResponse citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel
	if !plan.UserInterface.IsNull() {
		uiResponse, err = setAndGetSTFWebReceiverUserInterface(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.UserInterface)
		if err != nil {
			return
		}
	}

	//  Resources Service settings Config
	var resourcesServiceResponse citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel
	if !plan.ResourcesService.IsNull() {
		resourcesServiceResponse, err = setAndGetSTFWebReceiverResourcesService(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.ResourcesService)
		if err != nil {
			return
		}
	}

	var sitestyle citrixstorefront.STFWebReceiverSiteStyleResponseModel
	if !plan.WebReceiverSiteStyle.IsNull() {
		sitestyle, err = setAndGetSTFWebReceiverSiteStyle(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.WebReceiverSiteStyle)
		if err != nil {
			return
		}

	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &WebReceiverDetail, &appShortcutsResponse, &communicationResponse, &stsResponse, &authManagerResponse, &uiResponse, &resourcesServiceResponse, &sitestyle)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stfWebReceiverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFWebReceiverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	STFWebReceiver, err := getSTFWebReceiver(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		return
	}

	if STFWebReceiver == nil {
		resp.Diagnostics.AddWarning(
			"StoreFront Web Receiver Service not found",
			"StoreFront Web Receiver Service was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	var getWebReceiverRequestBody citrixstorefront.GetSTFWebReceiverRequestModel
	getWebReceiverRequestBody.SetVirtualPath(state.VirtualPath.ValueString())
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading StoreFront Authentication Service ",
				"Error message: "+err.Error(),
			)
			return
		}
		getWebReceiverRequestBody.SetSiteId(siteIdInt)
	}

	//Refresh Plugin Assistant
	if !state.PluginAssistant.IsNull() {
		getPlugInAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantGet(ctx, getWebReceiverRequestBody)
		assistant, err := getPlugInAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront WebReceiver Plugin Assistant",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshPlugInAssistant(ctx, &resp.Diagnostics, &assistant)
	}

	var appShortcutsResponse citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel
	if !state.ApplicationShortcuts.IsNull() {
		appShortcutsResponse, err = getSTFWebReceiverApplicationShortcuts(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	var communicationResponse citrixstorefront.GetWebReceiverCommunicationResponseModel
	if !state.Communication.IsNull() {
		communicationResponse, err = getSTFWebReceiverCommunication(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	var stsResponse citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel
	if !state.StrictTransportSecurity.IsNull() {
		stsResponse, err = getSTFWebReceiverStrictTransportSecurity(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	var authManagerResponse citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel
	if !state.AuthenticationManager.IsNull() {
		authManagerResponse, err = getSTFWebReceiverAuthenticationManager(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	// WebReceiverUserInterface config
	var uiResponse citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel
	if !state.UserInterface.IsNull() {
		uiResponse, err = getSTFWebReceiverUserInterface(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	//  Resources Service settings Config
	var resourcesServiceResponse citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel
	if !state.ResourcesService.IsNull() {
		resourcesServiceResponse, err = getSTFWebReceiverResourcesService(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	// Refresh Site Style
	var sitestyle citrixstorefront.STFWebReceiverSiteStyleResponseModel
	if !state.WebReceiverSiteStyle.IsNull() {
		sitestyle, err = getSTFWebReceiverSiteStyle(ctx, &resp.Diagnostics, r.client, state.SiteId.ValueString(), state.VirtualPath.ValueString())
		if err != nil {
			return
		}
	}

	state.RefreshPropertyValues(ctx, &resp.Diagnostics, STFWebReceiver, &appShortcutsResponse, &communicationResponse, &stsResponse, &authManagerResponse, &uiResponse, &resourcesServiceResponse, &sitestyle)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfWebReceiverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFWebReceiverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var getWebReceiverRequestBody citrixstorefront.GetSTFWebReceiverRequestModel
	getWebReceiverRequestBody.SetVirtualPath(plan.VirtualPath.ValueString())
	if plan.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating StoreFront Authentication Service ",
				"Error message: "+err.Error(),
			)
			return
		}
		getWebReceiverRequestBody.SetSiteId(siteIdInt)
	}

	// Update the Auth Methods
	if !plan.AuthenticationMethods.IsNull() {
		var authMethodCreateBody citrixstorefront.UpdateSTFWebReceiverAuthenticationMethodsRequestModel
		authMethods := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.AuthenticationMethods)
		authMethodCreateBody.SetAuthenticationMethods(authMethods)

		creatAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverSetSTFWebReceiverAuthenticationMethods(ctx, authMethodCreateBody, getWebReceiverRequestBody)
		// Create new STF WebReceiver Authentication Methods
		_, err := creatAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating StoreFront WebReceiver Authentication Methods",
				"Error message: "+err.Error(),
			)
			return
		}
	}

	// update the Plugin Assistant
	if !plan.PluginAssistant.IsNull() {
		var pluginAssistantBody citrixstorefront.UpdateSTFWebReceiverPluginAssistantRequestModel

		plannedPluginAssistant := util.ObjectValueToTypedObject[PluginAssistant](ctx, &resp.Diagnostics, plan.PluginAssistant)
		pluginAssistantBody.SetEnabled(plannedPluginAssistant.Enabled.ValueBool())
		pluginAssistantBody.SetUpgradeAtLogin(plannedPluginAssistant.UpgradeAtLogin.ValueBool())
		pluginAssistantBody.SetShowAfterLogin(plannedPluginAssistant.ShowAfterLogin.ValueBool())
		pluginAssistantBody.SetWin32Path(plannedPluginAssistant.Win32Path.ValueString())
		pluginAssistantBody.SetMacOSPath(plannedPluginAssistant.MacOSPath.ValueString())
		pluginAssistantBody.SetMacOSMinimumSupportedVersion(plannedPluginAssistant.MacOSMinimumSupportedVersion.ValueString())
		pluginAssistantBody.SetHtml5SingleTabLaunch(plannedPluginAssistant.Html5SingleTabLaunch.ValueBool())
		pluginAssistantBody.SetHtml5Enabled(plannedPluginAssistant.Html5Enabled.ValueString())
		pluginAssistantBody.SetHtml5Platforms(plannedPluginAssistant.Html5Platforms.ValueString())
		pluginAssistantBody.SetHtml5Preferences(plannedPluginAssistant.Html5Preferences.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppOrigins(plannedPluginAssistant.Html5ChromeAppOrigins.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppPreferences(plannedPluginAssistant.Html5ChromeAppPreferences.ValueString())
		pluginAssistantBody.SetProtocolHandlerEnabled(plannedPluginAssistant.ProtocolHandlerEnabled.ValueBool())
		pluginAssistantBody.SetProtocolHandlerPlatforms(plannedPluginAssistant.ProtocolHandlerPlatforms.ValueString())
		pluginAssistantBody.SetProtocolHandlerSkipDoubleHopCheckWhenDisabled(plannedPluginAssistant.ProtocolHandlerSkipDoubleHopCheckWhenDisabled.ValueBool())
		pluginAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantUpdate(ctx, pluginAssistantBody, getWebReceiverRequestBody)
		// Create new STF WebReceiver Plugin Assistant
		_, err := pluginAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating StoreFront WebReceiver Plugin Assistant",
				"Error message: "+err.Error(),
			)
			return
		}
	}

	if !plan.ApplicationShortcuts.IsNull() {
		appShortcutsResponse, err := setAndGetSTFWebReceiverApplicationShortcuts(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.ApplicationShortcuts)
		if err != nil {
			return
		}

		plan.ApplicationShortcuts = plan.RefreshApplicationShortcuts(ctx, &resp.Diagnostics, &appShortcutsResponse)
	}

	if !plan.Communication.IsNull() {
		communicationResponse, err := setAndGetSTFWebReceiverCommunication(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.Communication)
		if err != nil {
			return
		}

		plan.Communication = plan.RefreshCommunication(ctx, &resp.Diagnostics, &communicationResponse)
	}

	if !plan.StrictTransportSecurity.IsNull() {
		stsResponse, err := setAndGetSTFWebReceiverStrictTransportSecurity(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.StrictTransportSecurity)
		if err != nil {
			return
		}
		plan.StrictTransportSecurity = plan.RefreshStrictTransportSecurity(ctx, &resp.Diagnostics, &stsResponse)
	}

	if !plan.AuthenticationManager.IsNull() {
		authManagerResponse, err := setAndGetSTFWebReceiverAuthenticationManager(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.AuthenticationManager)
		if err != nil {
			return
		}
		plan.AuthenticationManager = plan.RefreshAuthenticationManager(ctx, &resp.Diagnostics, &authManagerResponse)
	}

	// WebReceiverUserInterface config
	if !plan.UserInterface.IsNull() {
		uiResponse, err := setAndGetSTFWebReceiverUserInterface(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.UserInterface)
		if err != nil {
			return
		}
		plan.UserInterface = plan.RefreshUserInterface(ctx, &resp.Diagnostics, &uiResponse)
	}

	//  Resources Service settings Config
	if !plan.ResourcesService.IsNull() {
		resourcesServiceResponse, err := setAndGetSTFWebReceiverResourcesService(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.ResourcesService)
		if err != nil {
			return
		}
		plan.ResourcesService = plan.RefreshResourcesService(ctx, &resp.Diagnostics, &resourcesServiceResponse)
	}

	if !plan.WebReceiverSiteStyle.IsNull() {
		sitestyle, err := setAndGetSTFWebReceiverSiteStyle(ctx, &resp.Diagnostics, r.client, plan.SiteId.ValueString(), plan.VirtualPath.ValueString(), plan.WebReceiverSiteStyle)
		if err != nil {
			return
		}

		plan.WebReceiverSiteStyle = plan.RefreshWebReceiverSiteStyle(ctx, &resp.Diagnostics, &sitestyle)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfWebReceiverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFWebReceiverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var body citrixstorefront.GetSTFWebReceiverRequestModel
	body.SetVirtualPath(state.VirtualPath.ValueString())
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting StoreFront Authentication Service ",
				"Error message: "+err.Error(),
			)
			return
		}
		body.SetSiteId(siteIdInt)
	}

	// Delete existing STF WebReceiver SiteStyle
	deleteWebReceiverSiteStyleRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverClearSTFWebReceiverSiteStyle(ctx, body)
	_, res_error := deleteWebReceiverSiteStyleRequest.Execute()
	if res_error != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront WebReceiver SiteStyle",
			"Error message: "+res_error.Error(),
		)
		return
	}

	// Delete existing STF WebReceiver
	deleteWebReceiverRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverClearSTFWebReceiver(ctx, body)
	_, err := deleteWebReceiverRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return
	}
}

func (r *stfWebReceiverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `site_id,virtual_path`, got: %q", req.ID),
		)
		return
	}

	_, err := strconv.Atoi(idSegments[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Site ID in Import Identifier",
			fmt.Sprintf("Site ID should be an integer, got: %q", idSegments[0]),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_path"), idSegments[1])...)
}

func constructGetWebReceiverRequestBody(diagnostics *diag.Diagnostics, siteId string, virtualPath string) (citrixstorefront.GetSTFWebReceiverRequestModel, error) {
	var getWebReceiverRequestBody citrixstorefront.GetSTFWebReceiverRequestModel
	siteIdInt, err := strconv.ParseInt(siteId, 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error parsing Site Id of the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetSTFWebReceiverRequestModel{}, err
	}
	getWebReceiverRequestBody.SetSiteId(siteIdInt)
	getWebReceiverRequestBody.SetVirtualPath(virtualPath)
	return getWebReceiverRequestBody, nil
}

// Gets the STFWebReceiver and logs any errors
func getSTFWebReceiver(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, state STFWebReceiverResourceModel) (*citrixstorefront.STFWebReceiverDetailModel, error) {
	var body citrixstorefront.GetSTFWebReceiverRequestModel
	if !state.SiteId.IsNull() {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of StoreFront WebReceiver ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}
	if !state.VirtualPath.IsNull() {
		body.SetVirtualPath(state.VirtualPath.ValueString())
	}

	getSTFWebReceiverRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverGetSTFWebReceiver(ctx, body)

	// Get refreshed STFWebReceiver properties from Orchestration
	STFWebReceiver, err := getSTFWebReceiverRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			return nil, nil
		}
		return &STFWebReceiver, err
	}
	return &STFWebReceiver, nil
}

func setAndGetSTFWebReceiverApplicationShortcuts(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, appShortcuts basetypes.ObjectValue) (citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel{}, err
	}

	plannedAppShortcuts := util.ObjectValueToTypedObject[ApplicationShortcuts](ctx, diagnostics, appShortcuts)
	gatewayUrls := util.StringSetToStringArray(ctx, diagnostics, plannedAppShortcuts.GatewayUrls)
	trustedUrls := util.StringSetToStringArray(ctx, diagnostics, plannedAppShortcuts.TrustedUrls)

	var setAppShortcutsBody citrixstorefront.SetWebReceiverApplicationShortcutsRequestModel
	setAppShortcutsBody.SetPromptForUntrustedShortcuts(plannedAppShortcuts.PromptForUntrustedShortcuts.ValueBool())
	setAppShortcutsBody.SetGatewayUrls(gatewayUrls)
	setAppShortcutsBody.SetTrustedUrls(trustedUrls)

	appShortcutsRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverApplicationShortcutsSet(ctx, getWebReceiverRequestBody, setAppShortcutsBody)
	err = appShortcutsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting application shortcuts for the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel{}, err
	}

	appShortcutsResponse, err := getSTFWebReceiverApplicationShortcuts(ctx, diagnostics, client, siteId, virtualPath)
	return appShortcutsResponse, err
}

func getSTFWebReceiverApplicationShortcuts(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverApplicationShortcutsResponseModel{}, err
	}

	getAppShortcutsRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverApplicationShortcutsGet(ctx, getWebReceiverRequestBody)
	appShortcutsResponse, err := getAppShortcutsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver Application Shortcuts details",
			"Error Message: "+err.Error(),
		)
	}
	return appShortcutsResponse, err
}

func setAndGetSTFWebReceiverCommunication(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, communication basetypes.ObjectValue) (citrixstorefront.GetWebReceiverCommunicationResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverCommunicationResponseModel{}, err
	}

	plannedCommunication := util.ObjectValueToTypedObject[Communication](ctx, diagnostics, communication)

	var setCommunicationBody citrixstorefront.SetWebReceiverCommunicationRequestModel
	setCommunicationBody.SetAttempts(int(plannedCommunication.Attempts.ValueInt64()))
	setCommunicationBody.SetTimeout(plannedCommunication.Timeout.ValueString())
	setCommunicationBody.SetLoopback(plannedCommunication.Loopback.ValueString())
	setCommunicationBody.SetLoopbackPortUsingHttp(int(plannedCommunication.LoopbackPortUsingHttp.ValueInt64()))
	setCommunicationBody.SetProxyEnabled(plannedCommunication.ProxyEnabled.ValueBool())
	setCommunicationBody.SetProxyPort(int(plannedCommunication.ProxyPort.ValueInt64()))
	setCommunicationBody.SetProxyProcessName(plannedCommunication.ProxyProcessName.ValueString())

	communicationRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverCommunicationSet(ctx, getWebReceiverRequestBody, setCommunicationBody)
	err = communicationRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting communication configuration for the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetWebReceiverCommunicationResponseModel{}, err
	}

	communicationResponse, err := getSTFWebReceiverCommunication(ctx, diagnostics, client, siteId, virtualPath)
	return communicationResponse, err
}

func getSTFWebReceiverCommunication(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetWebReceiverCommunicationResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverCommunicationResponseModel{}, err
	}

	getCommunicationRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverCommunicationGet(ctx, getWebReceiverRequestBody)
	communicationResponse, err := getCommunicationRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver Communicationconfiguration details",
			"Error Message: "+err.Error(),
		)
	}
	return communicationResponse, err
}

func setAndGetSTFWebReceiverStrictTransportSecurity(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, sts basetypes.ObjectValue) (citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel{}, err
	}

	plannedSts := util.ObjectValueToTypedObject[StrictTransportSecurity](ctx, diagnostics, sts)

	var setStsBody citrixstorefront.SetWebReceiverStrictTransportSecurityRequestModel
	setStsBody.SetEnabled(plannedSts.Enabled.ValueBool())
	setStsBody.SetPolicyDuration(plannedSts.PolicyDuration.ValueString())

	stsRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverStrictTransportSecuritySet(ctx, getWebReceiverRequestBody, setStsBody)
	err = stsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting Strict Transport Security configuration for the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel{}, err
	}

	stsResponse, err := getSTFWebReceiverStrictTransportSecurity(ctx, diagnostics, client, siteId, virtualPath)
	return stsResponse, err
}

func getSTFWebReceiverStrictTransportSecurity(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverStrictTransportSecurityResponseModel{}, err
	}

	getStsRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverStrictTransportSecurityGet(ctx, getWebReceiverRequestBody)
	stsResponse, err := getStsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver StrictTransportSecurity details",
			"Error Message: "+err.Error(),
		)
	}
	return stsResponse, err
}

func setAndGetSTFWebReceiverAuthenticationManager(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, authManager basetypes.ObjectValue) (citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel{}, err
	}

	plannedAuthManager := util.ObjectValueToTypedObject[AuthenticationManager](ctx, diagnostics, authManager)

	var setAuthManagerBody citrixstorefront.SetWebReceiverAuthenticationManagerRequestModel
	setAuthManagerBody.SetLoginFormTimeout(int(plannedAuthManager.LoginFormTimeout.ValueInt64()))

	authManagerRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverAuthenticationManagerSet(ctx, getWebReceiverRequestBody, setAuthManagerBody)
	err = authManagerRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting Authentication Manager configuration for the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel{}, err
	}

	authManagerResponse, err := getSTFWebReceiverAuthenticationManager(ctx, diagnostics, client, siteId, virtualPath)
	return authManagerResponse, err
}

func getSTFWebReceiverAuthenticationManager(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetWebReceiverAuthenticationManagerResponseModel{}, err
	}

	getAuthManagerRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverAuthenticationManagerGet(ctx, getWebReceiverRequestBody)
	authManagerResponse, err := getAuthManagerRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver AuthenticationManager details",
			"Error Message: "+err.Error(),
		)
	}
	return authManagerResponse, err
}

func setAndGetSTFWebReceiverUserInterface(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, userInterface basetypes.ObjectValue) (citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel{}, err
	}

	plannedUserInterface := util.ObjectValueToTypedObject[UserInterface](ctx, diagnostics, userInterface)

	var setUserInterfaceBody citrixstorefront.SetSTFWebReceiverUserInterfaceRequestModel

	setUserInterfaceBody.SetAutoLaunchDesktop(plannedUserInterface.AutoLaunchDesktop.ValueBool())
	setUserInterfaceBody.SetMultiClickTimeout(int(plannedUserInterface.MultiClickTimeout.ValueInt64()))
	setUserInterfaceBody.SetEnableAppsFolderView(plannedUserInterface.EnableAppsFolderView.ValueBool())
	setUserInterfaceBody.SetCategoryViewCollapsed(plannedUserInterface.CategoryViewCollapsed.ValueBool())
	setUserInterfaceBody.SetMoveAppToUncategorized(plannedUserInterface.MoveAppToUncategorized.ValueBool())
	setUserInterfaceBody.SetShowActivityManager(plannedUserInterface.ShowActivityManager.ValueBool())
	setUserInterfaceBody.SetShowFirstTimeUse(plannedUserInterface.ShowFirstTimeUse.ValueBool())

	stfSupportPreventIcaDownload := util.CheckStoreFrontVersion(client.StorefrontClient.VersionSF, ctx, diagnostics, 3, 27)
	if !plannedUserInterface.PreventIcaDownloads.IsNull() && stfSupportPreventIcaDownload {
		setUserInterfaceBody.SetPreventIcaDownloads(plannedUserInterface.PreventIcaDownloads.ValueBool())
	}

	if !plannedUserInterface.UIViews.IsNull() {
		plannedUiViews := util.ObjectValueToTypedObject[UIViews](ctx, diagnostics, plannedUserInterface.UIViews)
		setUserInterfaceBody.SetShowAppsView(plannedUiViews.ShowAppsView.ValueBool())
		setUserInterfaceBody.SetShowDesktopsView(plannedUiViews.ShowDesktopsView.ValueBool())
		setUserInterfaceBody.SetDefaultView(plannedUiViews.DefaultView.ValueString())
	}

	if !plannedUserInterface.WorkspaceControl.IsNull() {
		plannedWorkspaceControl := util.ObjectValueToTypedObject[WorkspaceControl](ctx, diagnostics, plannedUserInterface.WorkspaceControl)
		setUserInterfaceBody.SetWorkspaceControlEnabled(plannedWorkspaceControl.Enabled.ValueBool())
		setUserInterfaceBody.SetWorkspaceControlAutoReconnectAtLogon(plannedWorkspaceControl.AutoReconnectAtLogon.ValueBool())
		setUserInterfaceBody.SetWorkspaceControlLogoffAction(plannedWorkspaceControl.LogoffAction.ValueString())
		setUserInterfaceBody.SetWorkspaceControlShowReconnectButton(plannedWorkspaceControl.ShowReconnectButton.ValueBool())
		setUserInterfaceBody.SetWorkspaceControlShowDisconnectButton(plannedWorkspaceControl.ShowDisconnectButton.ValueBool())
	}

	if !plannedUserInterface.ReceiverConfiguration.IsNull() {
		plannedReceiverConfiguration := util.ObjectValueToTypedObject[ReceiverConfiguration](ctx, diagnostics, plannedUserInterface.ReceiverConfiguration)
		setUserInterfaceBody.SetReceiverConfigurationEnabled(plannedReceiverConfiguration.Enabled.ValueBool())
	}

	if !plannedUserInterface.AppShortcuts.IsNull() {
		plannedAppShortcuts := util.ObjectValueToTypedObject[AppShortcuts](ctx, diagnostics, plannedUserInterface.AppShortcuts)
		setUserInterfaceBody.SetAppShortcutsEnabled(plannedAppShortcuts.Enabled.ValueBool())
		setUserInterfaceBody.SetAppShortcutsAllowSessionReconnect(plannedAppShortcuts.AllowSessionReconnect.ValueBool())
	}

	if !plannedUserInterface.ProgressiveWebApp.IsNull() {
		plannedProgressiveWebApp := util.ObjectValueToTypedObject[ProgressiveWebApp](ctx, diagnostics, plannedUserInterface.ProgressiveWebApp)
		setUserInterfaceBody.SetProgressiveWebAppEnabled(plannedProgressiveWebApp.Enabled.ValueBool())
		setUserInterfaceBody.SetShowProgressiveWebAppInstallPrompt(plannedProgressiveWebApp.ShowInstallPrompt.ValueBool())
	}

	setUserInterfaceRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverUserInterfaceSet(ctx, getWebReceiverRequestBody, setUserInterfaceBody)
	err = setUserInterfaceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting User Interface configuration for the StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel{}, err
	}

	userInterfaceResponse, err := getSTFWebReceiverUserInterface(ctx, diagnostics, client, siteId, virtualPath)
	return userInterfaceResponse, err
}

func getSTFWebReceiverUserInterface(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetSTFWebReceiverUserInterfaceResponseModel{}, err
	}

	getUserInterfaceRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverUserInterfaceGet(ctx, getWebReceiverRequestBody)
	userInterfaceResponse, err := getUserInterfaceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver UserInterface details",
			"Error Message: "+err.Error(),
		)
	}
	return userInterfaceResponse, err
}

func getSTFWebReceiverResourcesService(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel{}, err
	}

	getWebReceiverResourcesRequest := client.StorefrontClient.WebReceiverSF.GetSTFWebReceiverResourcesService(ctx, getWebReceiverRequestBody)
	webReceiverResources, err := getWebReceiverResourcesRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver Resources details",
			"Error Message: "+err.Error(),
		)
	}
	return webReceiverResources, err
}

func setAndGetSTFWebReceiverResourcesService(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, resourcesService basetypes.ObjectValue) (citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel{}, err
	}

	plannedResourcesService := util.ObjectValueToTypedObject[ResourcesService](ctx, diagnostics, resourcesService)

	var setResourcesServiceBody citrixstorefront.SetSTFWebReceiverResourcesServiceRequestModel
	if !plannedResourcesService.PersistentIconCacheEnabled.IsNull() {
		setResourcesServiceBody.SetPersistentIconCacheEnabled(plannedResourcesService.PersistentIconCacheEnabled.ValueBool())
	}
	if !plannedResourcesService.IcaFileCacheExpiry.IsNull() {
		setResourcesServiceBody.SetIcaFileCacheExpiry(int(plannedResourcesService.IcaFileCacheExpiry.ValueInt64()))
	}
	if !plannedResourcesService.IconSize.IsNull() {
		setResourcesServiceBody.SetIconSize(int(plannedResourcesService.IconSize.ValueInt64()))
	}
	if !plannedResourcesService.ShowDesktopViewer.IsNull() {
		setResourcesServiceBody.SetShowDesktopViewer(plannedResourcesService.ShowDesktopViewer.ValueBool())
	}

	resourcesServiceRequest := client.StorefrontClient.WebReceiverSF.SetSTFWebReceiverResourcesService(ctx, getWebReceiverRequestBody, setResourcesServiceBody)
	err = resourcesServiceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting Resources Service settings for the StoreFront WebReceiver",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.GetSTFWebReceiverResourcesServiceResponseModel{}, err
	}

	resourcesServiceResponse, err := getSTFWebReceiverResourcesService(ctx, diagnostics, client, siteId, virtualPath)
	return resourcesServiceResponse, err
}

func getSTFWebReceiverSiteStyle(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string) (citrixstorefront.STFWebReceiverSiteStyleResponseModel, error) {
	getWebReceiverRequestBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.STFWebReceiverSiteStyleResponseModel{}, err
	}

	getWebReceiverSiteStyleRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverGetSTFWebReceiverSiteStyle(ctx, getWebReceiverRequestBody)
	sitestyle, err := getWebReceiverSiteStyleRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching StoreFront WebReceiver Site Style",
			"Error message: "+err.Error(),
		)
	}
	return sitestyle, err
}

func setAndGetSTFWebReceiverSiteStyle(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteId string, virtualPath string, siteStyle basetypes.ObjectValue) (citrixstorefront.STFWebReceiverSiteStyleResponseModel, error) {
	getSTFWebReceiverSiteStyleBody, err := constructGetWebReceiverRequestBody(diagnostics, siteId, virtualPath)
	if err != nil {
		return citrixstorefront.STFWebReceiverSiteStyleResponseModel{}, err
	}

	plannedSiteStyle := util.ObjectValueToTypedObject[WebReceiverSiteStyle](ctx, diagnostics, siteStyle)

	var setSiteStyleBody citrixstorefront.SetSTFWebReceiverSiteStyleRequestModel

	setSiteStyleBody.SetHeaderBackgroundColor(plannedSiteStyle.HeaderBackgroundColor.ValueString())

	setSiteStyleBody.SetHeaderForegroundColor(plannedSiteStyle.HeaderForegroundColor.ValueString())

	setSiteStyleBody.SetHeaderLogoPath(plannedSiteStyle.HeaderLogoPath.ValueString())

	setSiteStyleBody.SetLogonLogoPath(plannedSiteStyle.LogonLogoPath.ValueString())

	setSiteStyleBody.SetLinkColor(plannedSiteStyle.LinkColor.ValueString())

	setSiteStyleBody.SetIgnoreNonExistentLogos(plannedSiteStyle.IgnoreNonExistentLogos.ValueBool())

	siteStyleRequest := client.StorefrontClient.WebReceiverSF.STFWebReceiverSetSTFWebReceiverSiteStyle(ctx, getSTFWebReceiverSiteStyleBody, setSiteStyleBody)
	err = siteStyleRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting Site Style for the StoreFront WebReceiver",
			"Error message: "+err.Error(),
		)
		return citrixstorefront.STFWebReceiverSiteStyleResponseModel{}, err
	}

	siteStyleResponse, err := getSTFWebReceiverSiteStyle(ctx, diagnostics, client, siteId, virtualPath)
	return siteStyleResponse, err
}
