// Copyright © 2023. Citrix Systems, Inc.
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfWebReceiverResource{}
	_ resource.ResourceWithConfigure   = &stfWebReceiverResource{}
	_ resource.ResourceWithImportState = &stfWebReceiverResource{}
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

// Schema defines the schema for the resource.
func (r *stfWebReceiverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Storefront WebReceiver.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the Storefront webreceiver. Defaults to 1.",
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
			"authentication_methods": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The authentication methods supported by the WebReceiver.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"plugin_assistant": schema.SingleNestedAttribute{
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
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *stfWebReceiverResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
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
			"Error creating Storefront WebReceiver ",
			"\nError message: "+err.Error(),
		)
		return
	}
	body.SetSiteId(siteIdInt)
	body.SetVirtualPath(plan.VirtualPath.String())
	body.SetFriendlyName(plan.FriendlyName.ValueString())
	body.SetStoreService("(Get-STFStoreService -VirtualPath " + plan.StoreService.ValueString() + " ) ")
	createWebReceiverRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverCreateSTFWebReceiver(ctx, body)
	// Create new STF WebReceiver
	WebReceiverDetail, err := createWebReceiverRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Storefront WebReceiver",
			"TransactionId: ",
		)
		return
	}

	// Create the authentication methods Body
	if plan.AuthenticationMethods != nil {
		var authMethodCreateBody citrixstorefront.UpdateSTFWebReceiverAuthenticationMethodsRequestModel
		authMethodCreateBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		authMethodCreateBody.SetAuthenticationMethods(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.AuthenticationMethods))
		creatAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverSetSTFWebReceiverAuthenticationMethods(ctx, authMethodCreateBody)
		// Create new STF WebReceiver Authentication Methods
		_, err = creatAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Storefront WebReceiver Authentication Methods",
				"TransactionId: ",
			)
			return
		}
	}

	// Create the Plugin Assistant
	if plan.PluginAssistant != nil {
		var pluginAssistantBody citrixstorefront.UpdateSTFWebReceiverPluginAssistantRequestModel
		pluginAssistantBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		pluginAssistantBody.SetEnabled(plan.PluginAssistant.Enabled.ValueBool())
		pluginAssistantBody.SetUpgradeAtLogin(plan.PluginAssistant.UpgradeAtLogin.ValueBool())
		pluginAssistantBody.SetShowAfterLogin(plan.PluginAssistant.ShowAfterLogin.ValueBool())
		pluginAssistantBody.SetWin32Path(plan.PluginAssistant.Win32Path.ValueString())
		pluginAssistantBody.SetMacOSPath(plan.PluginAssistant.MacOSPath.ValueString())
		pluginAssistantBody.SetMacOSMinimumSupportedVersion(plan.PluginAssistant.MacOSMinimumSupportedVersion.ValueString())
		pluginAssistantBody.SetHtml5SingleTabLaunch(plan.PluginAssistant.Html5SingleTabLaunch.ValueBool())
		pluginAssistantBody.SetHtml5Enabled(plan.PluginAssistant.Html5Enabled.ValueString())
		pluginAssistantBody.SetHtml5Platforms(plan.PluginAssistant.Html5Platforms.ValueString())
		pluginAssistantBody.SetHtml5Preferences(plan.PluginAssistant.Html5Preferences.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppOrigins(plan.PluginAssistant.Html5ChromeAppOrigins.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppPreferences(plan.PluginAssistant.Html5ChromeAppPreferences.ValueString())
		pluginAssistantBody.SetProtocolHandlerEnabled(plan.PluginAssistant.ProtocolHandlerEnabled.ValueBool())
		pluginAssistantBody.SetProtocolHandlerPlatforms(plan.PluginAssistant.ProtocolHandlerPlatforms.ValueString())
		pluginAssistantBody.SetProtocolHandlerSkipDoubleHopCheckWhenDisabled(plan.PluginAssistant.ProtocolHandlerSkipDoubleHopCheckWhenDisabled.ValueBool())
		pluginAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantUpdate(ctx, pluginAssistantBody)
		// Create new STF WebReceiver Plugin Assistant
		_, err = pluginAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Storefront WebReceiver Plugin Assistant",
				"TransactionId: ",
			)
			return
		}

	}

	// Refresh the authentication methods
	if plan.AuthenticationMethods != nil {
		var authMethodGetBody citrixstorefront.GetSTFWebReceiverAuthenticationMethodsRequestModel
		authMethodGetBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		getAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverGetSTFWebReceiverAuthenticationMethods(ctx, authMethodGetBody)
		authMethoResult, err := getAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching Storefront WebReceiver Authentication Methods",
				"TransactionId: ",
			)
			return
		}
		util.RefreshListDeprecated(plan.AuthenticationMethods, authMethoResult.Methods)
	}

	//Refresh Plugin Assistant
	if plan.PluginAssistant != nil {
		var pluginAssistantGetBody citrixstorefront.GetSTFWebReceiverPluginAssistantRequestModel
		pluginAssistantGetBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		getPlugInAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantGet(ctx, pluginAssistantGetBody)
		assistant, err := getPlugInAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching Storefront WebReceiver Plugin Assistant",
				"TransactionId: ",
			)
			return
		}
		plan.RefreshPlugInAssistant(&assistant)
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(&WebReceiverDetail)

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

	//Refresh Plugin Assistant
	if state.PluginAssistant != nil {
		var pluginAssistantGetBody citrixstorefront.GetSTFWebReceiverPluginAssistantRequestModel
		pluginAssistantGetBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + state.VirtualPath.ValueString() + " -SiteId " + state.SiteId.ValueString() + " )")
		getPlugInAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantGet(ctx, pluginAssistantGetBody)
		assistant, err := getPlugInAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching Storefront WebReceiver Plugin Assistant",
				"TransactionId: ",
			)
			return
		}
		state.RefreshPlugInAssistant(&assistant)
	}

	state.RefreshPropertyValues(STFWebReceiver)

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

	// Get current state
	var state STFWebReceiverResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Auth Methods
	if plan.AuthenticationMethods != nil {
		var authMethodCreateBody citrixstorefront.UpdateSTFWebReceiverAuthenticationMethodsRequestModel
		authMethodCreateBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		authMethodCreateBody.SetAuthenticationMethods(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.AuthenticationMethods))
		creatAuthProtocolRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverSetSTFWebReceiverAuthenticationMethods(ctx, authMethodCreateBody)
		// Create new STF WebReceiver Authentication Methods
		_, err := creatAuthProtocolRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Storefront WebReceiver Authentication Methods",
				"TransactionId: ",
			)
			return
		}
	}

	// update the Plugin Assistant
	if plan.PluginAssistant != nil {
		var pluginAssistantBody citrixstorefront.UpdateSTFWebReceiverPluginAssistantRequestModel
		pluginAssistantBody.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + plan.VirtualPath.ValueString() + " -SiteId " + plan.SiteId.ValueString() + " )")
		pluginAssistantBody.SetEnabled(plan.PluginAssistant.Enabled.ValueBool())
		pluginAssistantBody.SetUpgradeAtLogin(plan.PluginAssistant.UpgradeAtLogin.ValueBool())
		pluginAssistantBody.SetShowAfterLogin(plan.PluginAssistant.ShowAfterLogin.ValueBool())
		pluginAssistantBody.SetWin32Path(plan.PluginAssistant.Win32Path.ValueString())
		pluginAssistantBody.SetMacOSPath(plan.PluginAssistant.MacOSPath.ValueString())
		pluginAssistantBody.SetMacOSMinimumSupportedVersion(plan.PluginAssistant.MacOSMinimumSupportedVersion.ValueString())
		pluginAssistantBody.SetHtml5SingleTabLaunch(plan.PluginAssistant.Html5SingleTabLaunch.ValueBool())
		pluginAssistantBody.SetHtml5Enabled(plan.PluginAssistant.Html5Enabled.ValueString())
		pluginAssistantBody.SetHtml5Platforms(plan.PluginAssistant.Html5Platforms.ValueString())
		pluginAssistantBody.SetHtml5Preferences(plan.PluginAssistant.Html5Preferences.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppOrigins(plan.PluginAssistant.Html5ChromeAppOrigins.ValueString())
		pluginAssistantBody.SetHtml5ChromeAppPreferences(plan.PluginAssistant.Html5ChromeAppPreferences.ValueString())
		pluginAssistantBody.SetProtocolHandlerEnabled(plan.PluginAssistant.ProtocolHandlerEnabled.ValueBool())
		pluginAssistantBody.SetProtocolHandlerPlatforms(plan.PluginAssistant.ProtocolHandlerPlatforms.ValueString())
		pluginAssistantBody.SetProtocolHandlerSkipDoubleHopCheckWhenDisabled(plan.PluginAssistant.ProtocolHandlerSkipDoubleHopCheckWhenDisabled.ValueBool())
		pluginAssistantRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverPluginAssistantUpdate(ctx, pluginAssistantBody)
		// Create new STF WebReceiver Plugin Assistant
		_, err := pluginAssistantRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Storefront WebReceiver Plugin Assistant",
				"TransactionId: ",
			)
			return
		}

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

	var body citrixstorefront.ClearSTFWebReceiverRequestModel
	if state.SiteId.ValueString() != "" {
		body.SetWebReceiverService("(Get-STFWebReceiverService -VirtualPath " + state.VirtualPath.ValueString() + " -SiteId " + state.SiteId.ValueString() + " )")
	}

	// Delete existing STF WebReceiver
	deleteWebReceiverRequest := r.client.StorefrontClient.WebReceiverSF.STFWebReceiverClearSTFWebReceiver(ctx, body)
	_, err := deleteWebReceiverRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Storefront WebReceiver ",
			"\nError message: "+err.Error(),
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

// Gets the STFWebReceiver and logs any errors
func getSTFWebReceiver(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, state STFWebReceiverResourceModel) (*citrixstorefront.STFWebReceiverDetailModel, error) {
	var body citrixstorefront.GetSTFWebReceiverRequestModel
	if !state.SiteId.IsNull() {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of Storefront WebReceiver ",
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
		return &STFWebReceiver, err
	}
	return &STFWebReceiver, nil
}
