// Copyright © 2024. Citrix Systems, Inc.
package cvad_site

import (
	"context"
	"fmt"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &siteSettingsResource{}
	_ resource.ResourceWithConfigure      = &siteSettingsResource{}
	_ resource.ResourceWithImportState    = &siteSettingsResource{}
	_ resource.ResourceWithValidateConfig = &siteSettingsResource{}
	_ resource.ResourceWithModifyPlan     = &siteSettingsResource{}
)

// NewSiteSettingsResource is a helper function to simplify the provider implementation.
func NewSiteSettingsResource() resource.Resource {
	return &siteSettingsResource{}
}

// siteSettingsResource is the resource implementation.
type siteSettingsResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *siteSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_settings"
}

// Configure adds the provider configured client to the resource.
func (r *siteSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	// Remove Site ID from the URL
	configuredURL := r.client.ApiClient.GetConfig().Servers[0].URL
	updatedUrl := strings.ReplaceAll(configuredURL, fmt.Sprintf("/%s/%s", r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId), "/"+r.client.ClientConfig.CustomerId)
	r.client.ApiClient.GetConfig().Servers[0].URL = updatedUrl
}

// Schema defines the schema for the resource.
func (r *siteSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SiteSettingsModel{}.GetSchema()
}

// Create implements resource.ResourceWithModifyPlan.
func (r *siteSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan SiteSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteSettings, multipleRemotePCAssignments, err := updateAndReturnSiteSettings(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	plan = plan.RefreshResourcePropertyValues(ctx, &resp.Diagnostics, r.client, siteSettings, multipleRemotePCAssignments)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.ResourceWithModifyPlan.
func (r *siteSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state SiteSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed site settings properties from Orchestration
	siteSettings, err := getSiteSettings(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Get multiple remote PC assignments settings from Orchestration
	multipleRemotePCAssignments, err := getMultipleRemotePCAssignmentsSetting(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	state = state.RefreshResourcePropertyValues(ctx, &resp.Diagnostics, r.client, siteSettings, multipleRemotePCAssignments)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.ResourceWithModifyPlan.
func (r *siteSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan SiteSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteSettings, multipleRemotePCAssignments, err := updateAndReturnSiteSettings(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	plan = plan.RefreshResourcePropertyValues(ctx, &resp.Diagnostics, r.client, siteSettings, multipleRemotePCAssignments)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.ResourceWithModifyPlan.
func (r *siteSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state SiteSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Attempting to delete Site Settings",
		"The requested site settings resource will be deleted from terraform state but there will be no configuration change for the site",
	)

	// Remove the SiteSettings from state file
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *siteSettingsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan SiteSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.WebUiPolicySetEnabled.IsUnknown() &&
		!plan.WebUiPolicySetEnabled.IsNull() &&
		!plan.WebUiPolicySetEnabled.ValueBool() {
		policySets, err := policies.GetPolicySets(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}
		for _, policySet := range policySets {
			if policySet.GetPolicySetType() == citrixorchestration.SDKGPOPOLICYSETTYPE_DELIVERY_GROUP_POLICIES {
				err := fmt.Errorf("`WebUiPolicySetEnabled` cannot be set to `false` when there are any policy sets of type `DeliveryGroupPolicies`")
				resp.Diagnostics.AddError(
					"Error validating Site Settings",
					err.Error(),
				)
				return
			}
		}
	}
}

func (r *siteSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	resp.Diagnostics.AddError(
		"Importing state is not supported",
		"Importing state for site settings is not supported, please specify the intended settings in the configuration. All the settings that are not specified will not be modified.",
	)
}

func (r *siteSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data SiteSettingsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func getSiteSettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (*citrixorchestration.SiteSettingsResponseModel, error) {
	siteId := client.ClientConfig.SiteId
	getSiteSettingsRequest := client.ApiClient.SitesAPIsDAAS.SitesGetSiteSettings(ctx, siteId)
	getSiteSettingsRequest = citrixdaasclient.AddRequestData(getSiteSettingsRequest, client)
	if client.AuthConfig.OnPremises {
		authToken, err := getAuthTokenForOnPremSiteSettingsRequest(client, diagnostics, siteId, "fetching")
		if err != nil {
			return nil, err
		}
		getSiteSettingsRequest = getSiteSettingsRequest.Authorization(authToken)
	}
	getSiteSettingsResult, httpResp, err := getSiteSettingsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error fetching Site Settings for site %s", siteId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
	return getSiteSettingsResult, err
}

func getMultipleRemotePCAssignmentsSetting(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (bool, error) {
	siteId := client.ClientConfig.SiteId
	getMultipleRemotePCAssignmentsRequest := client.ApiClient.SitesAPIsDAAS.SitesGetMultipleRemotePCAssignments(ctx, siteId)
	getMultipleRemotePCAssignmentsResult, httpResp, err := citrixdaasclient.ExecuteWithRetry[bool](getMultipleRemotePCAssignmentsRequest, client)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error fetching Multiple Remote PC Assignments Settings for site %s", siteId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
	return getMultipleRemotePCAssignmentsResult, err
}

func updateAndReturnSiteSettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan SiteSettingsModel) (*citrixorchestration.SiteSettingsResponseModel, bool, error) {

	validateOnPremSiteSettings(ctx, client, diagnostics, plan)
	if diagnostics.HasError() {
		err := fmt.Errorf("Error validating Site Settings configuration")
		return nil, false, err
	}
	siteId := client.ClientConfig.SiteId

	settingsConfigured := false
	body := citrixorchestration.EditSiteSettingsRequestModel{}
	if !plan.WebUiPolicySetEnabled.IsNull() {
		settingsConfigured = true
		body.SetWebUiPolicySetEnabled(plan.WebUiPolicySetEnabled.ValueBool())
	}

	if !plan.TrustRequestsSentToTheXmlServicePortEnabled.IsNull() {
		settingsConfigured = true
		body.SetTrustRequestsSentToTheXmlServicePortEnabled(plan.TrustRequestsSentToTheXmlServicePortEnabled.ValueBool())
	}

	if !plan.DnsResolutionEnabled.IsNull() {
		settingsConfigured = true
		body.SetDnsResolutionEnabled(plan.DnsResolutionEnabled.ValueBool())
	}

	if !plan.UseVerticalScalingForRdsLaunches.IsNull() {
		settingsConfigured = true
		body.SetUseVerticalScalingForRdsLaunches(plan.UseVerticalScalingForRdsLaunches.ValueBool())
	}

	if !plan.ConsoleInactivityTimeoutMinutes.IsNull() {
		settingsConfigured = true
		body.SetConsoleInactivityTimeoutMinutes(plan.ConsoleInactivityTimeoutMinutes.ValueInt32())
	}

	if !plan.SupportedAuthenticators.IsNull() {
		settingsConfigured = true
		authenticators, err := citrixorchestration.NewAuthenticatorFromValue(plan.SupportedAuthenticators.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Error updating Site Settings",
				"Unsupported authenticator type.",
			)
			return nil, false, err
		}
		body.SetSupportedAuthenticators(*authenticators)
	}

	if !plan.AllowedCorsOriginsForIwa.IsNull() {
		settingsConfigured = true
		allowedCorsOriginsForIwa := util.StringSetToStringArray(ctx, diagnostics, plan.AllowedCorsOriginsForIwa)
		body.SetAllowedCorsOriginsForIwa(allowedCorsOriginsForIwa)
	}

	if settingsConfigured {
		updateSiteSettingsRequest := client.ApiClient.SitesAPIsDAAS.SitesPatchSiteSettings(ctx, siteId)
		updateSiteSettingsRequest = updateSiteSettingsRequest.EditSiteSettingsRequestModel(body)
		updateSiteSettingsRequest = citrixdaasclient.AddRequestData(updateSiteSettingsRequest, client)
		if client.AuthConfig.OnPremises {
			authToken, err := getAuthTokenForOnPremSiteSettingsRequest(client, diagnostics, siteId, "updating")
			if err != nil {
				return nil, false, err
			}
			updateSiteSettingsRequest = updateSiteSettingsRequest.Authorization(authToken)
		}
		httpResp, err := updateSiteSettingsRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				fmt.Sprintf("Error updating Site Settings for site %s", siteId),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, false, err
		}
	}

	if !plan.MultipleRemotePCAssignments.IsNull() {
		multipleRpcAssignmentsReq := client.ApiClient.SitesAPIsDAAS.SitesPatchMultipleRemotePCAssignments(ctx, siteId)
		multipleRpcAssignmentsReq = multipleRpcAssignmentsReq.Allow(plan.MultipleRemotePCAssignments.ValueBool())
		multipleRpcAssignmentsReq = citrixdaasclient.AddRequestData(multipleRpcAssignmentsReq, client)
		if client.AuthConfig.OnPremises {
			authToken, err := getAuthTokenForOnPremSiteSettingsRequest(client, diagnostics, siteId, "updating")
			if err != nil {
				return nil, false, err
			}
			multipleRpcAssignmentsReq = multipleRpcAssignmentsReq.Authorization(authToken)
		}
		httpResp, err := multipleRpcAssignmentsReq.Execute()
		if err != nil {
			diagnostics.AddError(
				fmt.Sprintf("Error updating Multiple Remote PC Assignments Setting for site %s", siteId),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, false, err
		}
	}

	// Get refreshed site settings properties from Orchestration
	siteSettings, err := getSiteSettings(ctx, client, diagnostics)
	if err != nil {
		return siteSettings, false, err
	}

	// Get multiple remote PC assignments settings from Orchestration
	multipleRemotePCAssignments, err := getMultipleRemotePCAssignmentsSetting(ctx, client, diagnostics)
	if err != nil {
		return siteSettings, multipleRemotePCAssignments, err
	}

	return siteSettings, multipleRemotePCAssignments, nil
}

func validateOnPremSiteSettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan SiteSettingsModel) {
	if !client.AuthConfig.OnPremises {
		if !plan.ConsoleInactivityTimeoutMinutes.IsNull() {
			diagnostics.AddAttributeError(
				path.Root("console_inactivity_timeout_minutes"),
				"Incorrect Attribute Configuration",
				"console_inactivity_timeout_minutes cannot be configured for cloud environments",
			)
		}

		if !plan.SupportedAuthenticators.IsNull() {
			diagnostics.AddAttributeError(
				path.Root("supported_authenticators"),
				"Incorrect Attribute Configuration",
				"supported_authenticators cannot be configured for cloud environments",
			)
		}

		if !plan.AllowedCorsOriginsForIwa.IsNull() {
			diagnostics.AddAttributeError(
				path.Root("allowed_cors_origins_for_iwa"),
				"Incorrect Attribute Configuration",
				"allowed_cors_origins_for_iwa cannot be configured for cloud environments",
			)
		}
	}
}

func getAuthTokenForOnPremSiteSettingsRequest(client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId string, operation string) (string, error) {
	cwsAuthToken, httpResp, err := client.SignIn()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error %s Site Settings for site %s", operation, siteId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return "", err
	}

	if cwsAuthToken == "" {
		err := fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) +
			"\nError message: Unable to fetch Auth Token")
		diagnostics.AddError(
			fmt.Sprintf("Error %s Site Settings for site %s", operation, siteId),
			err.Error(),
		)
		return "", err
	}
	token := strings.Split(cwsAuthToken, "=")[1]
	return "Bearer " + token, nil
}
