// Copyright © 2024. Citrix Systems, Inc.
package stf_store

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
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfStoreServiceResource{}
	_ resource.ResourceWithConfigure   = &stfStoreServiceResource{}
	_ resource.ResourceWithImportState = &stfStoreServiceResource{}
	_ resource.ResourceWithModifyPlan  = &stfStoreServiceResource{}
)

// stfStoreServiceResource is a helper function to simplify the provider implementation.
func NewSTFStoreServiceResource() resource.Resource {
	return &stfStoreServiceResource{}
}

// stfStoreServiceResource is the resource implementation.
type stfStoreServiceResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfStoreServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_store_service"
}

// Configure adds the provider configured client to the resource.
func (r *stfStoreServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// ModifyPlan modifies the resource plan before it is applied.
func (r *stfStoreServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
	var plan STFStoreServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Anonymous.IsUnknown() && !plan.AuthenticationService.IsUnknown() &&
		plan.Anonymous.IsNull() && plan.AuthenticationService.IsNull() {
		resp.Diagnostics.AddError(
			"Error "+operation+" StoreFront StoreService",
			"Either `anonymous` or `authentication_service_virtual_path` should be provided",
		)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfStoreServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.CreateSTFStoreRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
		return
	}

	body.SetSiteId(siteIdInt)
	body.SetVirtualPath(plan.VirtualPath.ValueString())
	body.SetFriendlyName(plan.FriendlyName.ValueString())

	var getAuthenticationServiceBody citrixstorefront.GetSTFAuthenticationServiceRequestModel
	if !plan.Anonymous.IsNull() && plan.Anonymous.ValueBool() {
		body.SetAnonymous(true)
	} else {
		getAuthenticationServiceBody.SetVirtualPath(plan.AuthenticationService.ValueString())
	}

	if !plan.LoadBalance.IsNull() {
		body.SetLoadBalance(plan.LoadBalance.ValueBool())
	}

	createStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreCreateSTFStore(ctx, body, getAuthenticationServiceBody)

	// Create new STF StoreService
	StoreServiceDetail, err := createStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront StoreService",
			"Error message: "+err.Error(),
		)
		return
	}
	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &StoreServiceDetail)

	// Create StoreFarmConfiguration
	if !plan.FarmSettings.IsNull() {

		err := setFarmSettingsSetRequest(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			return
		}

		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}

	// Create StoreFront Store Enumeration Options
	if !plan.EnumerationOptions.IsNull() {

		// Update Storefront Store Enumeration Options
		setSTFStoreEnumerationOptions(ctx, r.client, &resp.Diagnostics, plan)

		// Get updated STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Enumeration Options",
				"Error message: "+err.Error(),
			)
			return
		}

		// Refresh Storefront StoreService Enumerations
		plan.RefreshEnumerationOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Create StoreFront Gateway Settings
	if !plan.GatewaySettings.IsNull() {

		// Update Storefront Store Gateway Settings
		plan.setSTFStoreGatewaySettings(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}

		// Get updated STFStoreService Gateway Settings
		getResponse, err := plan.getSTFStoreGatewaySettings(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}

		// Refresh Storefront StoreService Gateway Settings
		plan.RefreshGatewaySettings(ctx, &resp.Diagnostics, getResponse)
	}

	// Set PNA properties
	if !plan.PNA.IsNull() {
		var storeBody citrixstorefront.GetSTFStoreRequestModel
		storeBody.SetVirtualPath(plan.VirtualPath.ValueString())

		pna := util.ObjectValueToTypedObject[PNA](ctx, &resp.Diagnostics, plan.PNA)
		if pna.Enable.ValueBool() {

			// Disable PNA first because of the existing problem from PNA cmdlet
			disablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
			err := disablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}

			// Enable PNA
			var pnaSetBody citrixstorefront.STFStorePnaSetRequestModel
			enablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreEnableStorePna(ctx, pnaSetBody, storeBody)
			err = enablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error enabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}
		} else {
			// Disable PNA
			disablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
			err := disablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}
		}

		// Fetch updated PNA
		updatedPna, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching updated PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
		plan.RefreshPnaValues(ctx, &resp.Diagnostics, updatedPna)
	}

	// Create StoreFront Store Launch Options
	if !plan.LaunchOptions.IsNull() {

		// Update Storefront Store Launch Options
		setSTFStoreLaunchOptions(ctx, r.client, &resp.Diagnostics, plan)

		// Get updated STFStoreService Launch Options
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Launch Options",
				"Error message: "+err.Error(),
			)
			return
		}

		// Refresh Storefront StoreService Launch Options
		plan.RefreshLaunchOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Create StoreFront Roaming Account

	if !plan.RoamingAccount.IsNull() {
		// Update Storefront Roaming Account
		err := setSTFRoamingAccount(ctx, r.client, &resp.Diagnostics, plan)
		if err != nil {
			return
		}
		// Get updated Roaming Account
		getResponse, err := getSTFRoamingAccount(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFront Roaming Account",
				"Error message: "+err.Error(),
			)
			return
		}

		// Refresh Storefront StoreService Launch Options
		plan.RefreshRoamingAccount(ctx, &resp.Diagnostics, getResponse)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *stfStoreServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFStoreServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	STFStoreService, err := state.getSTFStoreService(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}
	if STFStoreService == nil {
		resp.Diagnostics.AddWarning(
			"StoreFront Store Service not found",
			"StoreFront Store Service was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
		)
		resp.State.RemoveResource(ctx)
		return
	}
	state.RefreshPropertyValues(ctx, &resp.Diagnostics, STFStoreService)
	// Refresh StoreFarmConfiguration
	if !state.FarmSettings.IsNull() {
		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, state)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}
	//Refresh Storefront StoreService Enumerations
	if !state.EnumerationOptions.IsNull() {
		// Get STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, state)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Enumeration Options",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshEnumerationOptions(ctx, &resp.Diagnostics, getResponse)
	}

	//Refresh Storefront StoreService Launch Options
	if !state.LaunchOptions.IsNull() {
		// Get STFStoreService Launch Options
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, state)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Launch Options",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshLaunchOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Fetch StoreService Gateway Settings
	if !state.GatewaySettings.IsNull() {
		// Get STFStoreService Gateway Settings
		getResponse, err := state.getSTFStoreGatewaySettings(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}
		state.RefreshGatewaySettings(ctx, &resp.Diagnostics, getResponse)
	}

	// Fetch Pna
	if !state.PNA.IsNull() {
		var storeBody citrixstorefront.GetSTFStoreRequestModel
		storeBody.SetVirtualPath(state.VirtualPath.ValueString())
		updatedPna, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching updated PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
		state.RefreshPnaValues(ctx, &resp.Diagnostics, updatedPna)
	}

	// Fetch Roaming Account

	if !state.RoamingAccount.IsNull() {
		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getSTFRoamingAccount(ctx, r.client, state)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshRoamingAccount(ctx, &resp.Diagnostics, getResponse)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfStoreServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed STFStoreService
	_, err := plan.getSTFStoreService(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Construct the update model
	var getSTFStoreServiceBody = &citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Update STFStoreService
	editStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreSetSTFStore(ctx, *getSTFStoreServiceBody)
	_, err = editStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
	}

	// Fetch updated STFStoreService
	updatedSTFStoreService, err := plan.getSTFStoreService(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedSTFStoreService)

	// Update StoreFarmConfiguration
	if !plan.FarmSettings.IsNull() {

		err := setFarmSettingsSetRequest(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			return
		}

		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}

	//  updated STFStoreService Enumeration Options
	if !plan.EnumerationOptions.IsNull() {

		// Update Storefront Store Enumeration Options
		setSTFStoreEnumerationOptions(ctx, r.client, &resp.Diagnostics, plan)

		// Get updated STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Enumeration Options",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshEnumerationOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Create StoreFront Gateway Settings
	if !plan.GatewaySettings.IsNull() {

		// Update Storefront Store Gateway Settings
		plan.setSTFStoreGatewaySettings(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}

		// Get updated STFStoreService Gateway Settings
		getResponse, err := plan.getSTFStoreGatewaySettings(ctx, r.client, &resp.Diagnostics)
		if err != nil {
			return
		}

		// Refresh Storefront StoreService Gateway Settings
		plan.RefreshGatewaySettings(ctx, &resp.Diagnostics, getResponse)
	}

	// Set PNA properties
	if !plan.PNA.IsNull() {
		var storeBody citrixstorefront.GetSTFStoreRequestModel
		storeBody.SetVirtualPath(plan.VirtualPath.ValueString())

		pna := util.ObjectValueToTypedObject[PNA](ctx, &resp.Diagnostics, plan.PNA)
		if pna.Enable.ValueBool() {
			// Disable PNA first because of the existing problem from PNA cmdlet
			disablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
			err := disablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}

			// Enable PNA
			var pnaSetBody citrixstorefront.STFStorePnaSetRequestModel
			enablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreEnableStorePna(ctx, pnaSetBody, storeBody)
			err = enablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error enabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}
		} else {
			// Disable PNA
			disablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
			err := disablePnaRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling PNA for StoreFront StoreService",
					"Error message: "+err.Error(),
				)
			}
		}

		// Fetch updated PNA
		updatedPna, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching updated PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
		plan.RefreshPnaValues(ctx, &resp.Diagnostics, updatedPna)
	}

	// Update StoreFront Store Launch Options
	if !plan.LaunchOptions.IsNull() {

		// Update Storefront Store Launch Options
		setSTFStoreLaunchOptions(ctx, r.client, &resp.Diagnostics, plan)

		// Get updated STFStoreService Launch Options
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Launch Options",
				"Error message: "+err.Error(),
			)
			return
		}

		// Refresh Storefront StoreService Launch Options
		plan.RefreshLaunchOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Update StoreFront Roaming Account

	if !plan.RoamingAccount.IsNull() {
		// Update Storefront Roaming Account
		update_err := setSTFRoamingAccount(ctx, r.client, &resp.Diagnostics, plan)

		if update_err != nil {
			return
		}

		// Get updated Roaming Account
		getResponse, err := getSTFRoamingAccount(ctx, r.client, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFront Roaming Account",
				"Error message: "+err.Error(),
			)
			return
		}

		// Refresh Storefront StoreService Launch Options
		plan.RefreshRoamingAccount(ctx, &resp.Diagnostics, getResponse)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfStoreServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFStoreServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var body citrixstorefront.GetSTFStoreRequestModel
	body.SetVirtualPath(state.VirtualPath.ValueString())
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting StoreFront Store Service ",
				"Error message: "+err.Error(),
			)
			return
		}
		body.SetSiteId(siteIdInt)
	}

	// Delete existing STF StoreService
	deleteStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreClearSTFStore(ctx, body)
	_, err := deleteStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Store Service ",
			"Error message: "+err.Error(),
		)
		return
	}
}

func (r *stfStoreServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

// Gets the STFStoreService and logs any errors
func (plan STFStoreServiceResourceModel) getSTFStoreService(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (*citrixstorefront.STFStoreDetailModel, error) {
	var body citrixstorefront.GetSTFStoreRequestModel
	if !plan.SiteId.IsNull() {
		siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of StoreFront StoreService ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}
	body.SetVirtualPath(plan.VirtualPath.ValueString())
	getSTFStoreServiceRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStore(ctx, body)

	// Get refreshed STFStoreService properties from Orchestration
	STFStoreService, err := getSTFStoreServiceRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			return nil, nil
		}
		return &STFStoreService, err
	}
	return &STFStoreService, nil
}

func setFarmSettingsSetRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFStoreServiceResourceModel) error {
	var farmSettingBody citrixstorefront.SetStoreFarmConfigurationRequestModel

	plannedFarmSettings := util.ObjectValueToTypedObject[FarmSettings](ctx, diagnostics, plan.FarmSettings)

	farmSettingBody.SetEnableFileTypeAssociation(plannedFarmSettings.EnableFileTypeAssociation.ValueBool())
	farmSettingBody.SetCommunicationTimeout(plannedFarmSettings.CommunicationTimeout.ValueString())
	farmSettingBody.SetConnectionTimeout(plannedFarmSettings.ConnectionTimeout.ValueString())
	farmSettingBody.SetLeasingStatusExpiryFailed(plannedFarmSettings.LeasingStatusExpiryFailed.ValueString())
	farmSettingBody.SetLeasingStatusExpiryLeasing(plannedFarmSettings.LeasingStatusExpiryLeasing.ValueString())
	farmSettingBody.SetLeasingStatusExpiryPending(plannedFarmSettings.LeasingStatusExpiryPending.ValueString())
	farmSettingBody.SetPooledSockets(plannedFarmSettings.PooledSockets.ValueBool())
	farmSettingBody.SetServerCommunicationAttempts(int(plannedFarmSettings.ServerCommunicationAttempts.ValueInt64()))
	farmSettingBody.SetBackgroundHealthCheckPollingPeriod(plannedFarmSettings.BackgroundHealthCheckPollingPeriod.ValueString())
	farmSettingBody.SetAdvancedHealthCheck(plannedFarmSettings.AdvancedHealthCheck.ValueBool())
	farmSettingBody.SetCertRevocationPolicy(plannedFarmSettings.CertRevocationPolicy.ValueString())

	// Generate STFStoreFarmConfig body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Set StoreFront Enumeration Options
	farmSettingsRequest := client.StorefrontClient.StoreSF.STFStoreFarmSetSTFStoreConfiguration(ctx, farmSettingBody, getSTFStoreServiceBody)

	// Execute the request
	_, err := farmSettingsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error creating StoreFront Store Farm Configurations",
			"Error message: "+err.Error(),
		)
		return err
	}
	return nil

}

// Get STF-StoreFarm Config
func getFarmSettingsGetRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan STFStoreServiceResourceModel) (*citrixstorefront.StoreFarmConfigurationResponseModel, error) {
	// Generate getSTFStoreService body
	getFarmSettingsBody := citrixstorefront.GetSTFStoreRequestModel{}
	getFarmSettingsBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Get Request for STF StoreFarm Configurations
	farmSettingsRequest := client.StorefrontClient.StoreSF.STFStoreFarmGetStoreConfiguration(ctx, getFarmSettingsBody)

	// Execute the request
	getResponse, err := farmSettingsRequest.Execute()
	return &getResponse, err
}

// Set Storefront Store Enumeration Options
func setSTFStoreEnumerationOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFStoreServiceResourceModel) error {
	// Generate API request body
	var enumerationOptionsBody citrixstorefront.SetSTFStoreEnumerationOptionsRequestModel
	enumerationOptions := util.ObjectValueToTypedObject[EnumerationOptions](ctx, diagnostics, plan.EnumerationOptions)
	if !enumerationOptions.EnhancedEnumeration.IsNull() {
		enumerationOptionsBody.SetEnhancedEnumeration(enumerationOptions.EnhancedEnumeration.ValueBool())
	}
	if !enumerationOptions.MaximumConcurrentEnumerations.IsNull() {
		enumerationOptionsBody.SetMaximumConcurrentEnumerations(enumerationOptions.MaximumConcurrentEnumerations.ValueInt64())
	}
	if !enumerationOptions.FilterByTypesInclude.IsNull() {
		enumerationOptionsBody.SetFilterByTypesInclude(util.StringListToStringArray(ctx, diagnostics, enumerationOptions.FilterByTypesInclude))
	}
	if !enumerationOptions.FilterByKeywordsInclude.IsNull() {
		enumerationOptionsBody.SetFilterByKeywordsInclude(util.StringListToStringArray(ctx, diagnostics, enumerationOptions.FilterByKeywordsInclude))
	}
	if !enumerationOptions.FilterByKeywordsExclude.IsNull() {
		enumerationOptionsBody.SetFilterByKeywordsExclude(util.StringListToStringArray(ctx, diagnostics, enumerationOptions.FilterByKeywordsExclude))
	}

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Set StoreFront Enumeration Options
	enumerationOptionsRequest := client.StorefrontClient.StoreSF.STFStoreSetSTFStoreEnumerationOptions(ctx, enumerationOptionsBody, getSTFStoreServiceBody)

	// Execute the request
	_, err := enumerationOptionsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting StoreFront Store Enumeration Options",
			"Error message: "+err.Error(),
		)
		return err
	}
	return nil
}

// Get Storefront store Enumeration Options
func getSTFStoreEnumerationOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan STFStoreServiceResourceModel) (*citrixstorefront.GetSTFStoreEnumerationOptionsResponseModel, error) {
	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Get StoreFront Enumeration Options
	enumerationOptionsRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStoreEnumerationOptions(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := enumerationOptionsRequest.Execute()

	return &getResponse, err
}

// Set Storefront Store Launch Options
func setSTFStoreLaunchOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFStoreServiceResourceModel) error {
	// Generate API request body
	var launchOptionsBody citrixstorefront.SetSTFStoreLaunchOptionsRequestModel
	launchOptions := util.ObjectValueToTypedObject[LaunchOptions](ctx, diagnostics, plan.LaunchOptions)

	if !launchOptions.AddressResolutionType.IsNull() {
		launchOptionsBody.SetAddressResolutionType(launchOptions.AddressResolutionType.ValueString())
	}
	if !launchOptions.AllowFontSmoothing.IsNull() {
		launchOptionsBody.SetAllowFontSmoothing(launchOptions.AllowFontSmoothing.ValueBool())
	}
	if !launchOptions.AllowSpecialFolderRedirection.IsNull() {
		launchOptionsBody.SetAllowSpecialFolderRedirection(launchOptions.AllowSpecialFolderRedirection.ValueBool())
	}
	if !launchOptions.FederatedAuthenticationServiceFailover.IsNull() {
		launchOptionsBody.SetFederatedAuthenticationServiceFailover(launchOptions.FederatedAuthenticationServiceFailover.ValueBool())
	}
	if !launchOptions.IcaTemplateName.IsNull() {
		launchOptionsBody.SetIcaTemplateName(launchOptions.IcaTemplateName.ValueString())
	}
	if !launchOptions.IgnoreClientProvidedClientAddress.IsNull() {
		launchOptionsBody.SetIgnoreClientProvidedClientAddress(launchOptions.IgnoreClientProvidedClientAddress.ValueBool())
	}
	if !launchOptions.OverlayAutoLoginCredentialsWithTicket.IsNull() {
		launchOptionsBody.SetOverlayAutoLoginCredentialsWithTicket(launchOptions.OverlayAutoLoginCredentialsWithTicket.ValueBool())
	}
	if !launchOptions.OverrideIcaClientName.IsNull() {
		launchOptionsBody.SetOverrideIcaClientName(launchOptions.OverrideIcaClientName.ValueBool())
	}
	if !launchOptions.RDPOnly.IsNull() {
		launchOptionsBody.SetRDPOnly(launchOptions.RDPOnly.ValueBool())
	}
	if !launchOptions.RequestIcaClientSecureChannel.IsNull() {
		launchOptionsBody.SetRequestICAClientSecureChannel(launchOptions.RequestIcaClientSecureChannel.ValueString())
	}
	if !launchOptions.RequireLaunchReference.IsNull() {
		launchOptionsBody.SetRequireLaunchReference(launchOptions.RequireLaunchReference.ValueBool())
	}
	if !launchOptions.SetNoLoadBiasFlag.IsNull() {
		launchOptionsBody.SetSetNoLoadBiasFlag(launchOptions.SetNoLoadBiasFlag.ValueBool())
	}
	if !launchOptions.VdaLogonDataProvider.IsNull() {
		launchOptionsBody.SetVdaLogonDataProvider(launchOptions.VdaLogonDataProvider.ValueString())
	}

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Set StoreFront Enumeration Options
	launchOptionsRequest := client.StorefrontClient.StoreSF.STFStoreSetSTFStoreLaunchOptions(ctx, launchOptionsBody, getSTFStoreServiceBody)

	// Execute the request
	err := launchOptionsRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting StoreFront Store Launch Options",
			"Error message: "+err.Error(),
		)
		return err
	}
	return nil
}

// Get Storefront store Launch Options
func getSTFStoreLaunchOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan STFStoreServiceResourceModel) (*citrixstorefront.GetSTFStoreLaunchOptionsResponseModel, error) {
	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Get StoreFront Launch Options
	launchOptionsRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStoreLaunchOptions(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := launchOptionsRequest.Execute()

	return &getResponse, err
}

func (plan *STFStoreServiceResourceModel) setSTFStoreGatewaySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) error {
	var gatewaySettingsBody citrixstorefront.STFStoreGatewayServiceSetRequestModel
	gatewaySettings := util.ObjectValueToTypedObject[GatewaySettings](ctx, diagnostics, plan.GatewaySettings)

	if !gatewaySettings.Enabled.IsNull() {
		gatewaySettingsBody.SetEnabled(gatewaySettings.Enabled.ValueBool())
	}
	if !gatewaySettings.CustomerId.IsNull() && gatewaySettings.CustomerId.ValueString() != "" {
		gatewaySettingsBody.SetCustomerId(gatewaySettings.CustomerId.ValueString())
	}
	if !gatewaySettings.GetGatewayServiceUrl.IsNull() && gatewaySettings.GetGatewayServiceUrl.ValueString() != "" {
		gatewaySettingsBody.SetGatewayDiscoveryProtocol(gatewaySettings.GetGatewayServiceUrl.ValueString())
	}
	if !gatewaySettings.PrivateKey.IsNull() && gatewaySettings.PrivateKey.ValueString() != "" {
		gatewaySettingsBody.SetPrivateKey(gatewaySettings.PrivateKey.ValueString())
	}
	if !gatewaySettings.ServiceName.IsNull() && gatewaySettings.ServiceName.ValueString() != "" {
		gatewaySettingsBody.SetServiceName(gatewaySettings.ServiceName.ValueString())
	}
	if !gatewaySettings.InstanceId.IsNull() && gatewaySettings.InstanceId.ValueString() != "" {
		gatewaySettingsBody.SetInstanceId(gatewaySettings.InstanceId.ValueString())
	}
	if !gatewaySettings.SecureTicketAuthorityUrl.IsNull() {
		gatewaySettingsBody.SetSecureTicketAuthorityUrl(gatewaySettings.SecureTicketAuthorityUrl.ValueString())
	}
	if !gatewaySettings.SecureTicketLifetime.IsNull() {
		gatewaySettingsBody.SetSecureTicketLifetime(gatewaySettings.SecureTicketLifetime.ValueString())
	}
	if !gatewaySettings.SessionReliability.IsNull() {
		gatewaySettingsBody.SetSessionReliability(gatewaySettings.SessionReliability.ValueBool())
	}
	if !gatewaySettings.IgnoreZones.IsNull() {
		gatewaySettingsBody.SetIgnoreZones(util.StringListToStringArray(ctx, diagnostics, gatewaySettings.IgnoreZones))
	}
	if !gatewaySettings.HandleZones.IsNull() {
		gatewaySettingsBody.SetHandleZones(util.StringListToStringArray(ctx, diagnostics, gatewaySettings.HandleZones))
	}

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error creating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
		return err
	}
	getSTFStoreServiceBody.SetSiteId(siteIdInt)
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Set Storefront Gateway Settings
	gatewaySerivceRequest := client.StorefrontClient.StoreSF.STFStoreSETStoreGatewayService(ctx, gatewaySettingsBody, getSTFStoreServiceBody)

	// Execute the request
	err = gatewaySerivceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting StoreFront Store Gateway Settings",
			"Error message: "+err.Error(),
		)
		return err
	}
	return nil
}

func (plan *STFStoreServiceResourceModel) getSTFStoreGatewaySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (*citrixstorefront.STFStoreGatewayServiceResponseModel, error) {
	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error creating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
		return nil, err
	}
	getSTFStoreServiceBody.SetSiteId(siteIdInt)
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Get StoreFront Gateway Settings
	gatewayServiceRequest := client.StorefrontClient.StoreSF.STFStoreGETStoreGatewayService(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := gatewayServiceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error getting StoreFront Store Gateway Settings",
			"Error message: "+err.Error(),
		)
		return nil, err
	}
	return &getResponse, err
}

func setSTFRoamingAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFStoreServiceResourceModel) error {
	var roamingAccountSettingsBody citrixstorefront.SetSTFRoamingAccountRequestModel

	roamingAccountSettings := util.ObjectValueToTypedObject[RoamingAccount](ctx, diagnostics, plan.RoamingAccount)

	if !roamingAccountSettings.Published.IsNull() {
		roamingAccountSettingsBody.SetPublished(roamingAccountSettings.Published.ValueBool())
	}

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Set Storefront Gateway Settings
	gatewaySerivceRequest := client.StorefrontClient.StoreSF.STFRoamingAccountSet(ctx, roamingAccountSettingsBody, getSTFStoreServiceBody)

	// Execute the request
	err := gatewaySerivceRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting StoreFront Store Gateway Settings",
			"Error message: "+err.Error(),
		)
		return err
	}
	return nil

}

func getSTFRoamingAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan STFStoreServiceResourceModel) (*citrixstorefront.GetSTFRoamingAccountResponseModel, error) {
	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetVirtualPath(plan.VirtualPath.ValueString())

	// Create the client request to Get StoreFront Gateway Settings
	roamingAccRequest := client.StorefrontClient.StoreSF.STFRoamingAccountGet(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := roamingAccRequest.Execute()

	return &getResponse, err
}
