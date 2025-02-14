// Copyright © 2024. Citrix Systems, Inc.
package stf_store

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_deployment"
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

func (*stfStoreServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = STFStoreServiceResourceModel{}.GetSchema()
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

	//Create Store Farms
	var farms []citrixstorefront.StoreFarmModel
	storeFarm := util.ObjectListToTypedArray[StoreFarm](ctx, &resp.Diagnostics, plan.StoreFarm)
	createAndUpdateStoreFarms(ctx, r.client, &resp.Diagnostics, farms, siteIdInt, plan.VirtualPath.ValueString(), storeFarm)
	farms, err = plan.getStoreFarms(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &StoreServiceDetail, farms)
	// Create StoreFarmConfiguration
	if !plan.FarmSettings.IsNull() {
		plannedFarmSettings := util.ObjectValueToTypedObject[FarmSettings](ctx, &resp.Diagnostics, plan.FarmSettings)
		err := setFarmSettingsSetRequest(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), plannedFarmSettings)

		if err != nil {
			return
		}

		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations in Create",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}

	// Create StoreFront Store Enumeration Options
	if !plan.EnumerationOptions.IsNull() {
		enumerationOptions := util.ObjectValueToTypedObject[EnumerationOptions](ctx, &resp.Diagnostics, plan.EnumerationOptions)
		// Update Storefront Store Enumeration Options
		setSTFStoreEnumerationOptions(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), enumerationOptions)

		// Get updated STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
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

	// Set PNA properties
	if !plan.PNA.IsNull() {
		pna := util.ObjectValueToTypedObject[PNA](ctx, &resp.Diagnostics, plan.PNA)
		setSTFStorePNA(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), pna)

		updatedPna, err := getSTFStorePNA(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString())
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
		launchOptions := util.ObjectValueToTypedObject[LaunchOptions](ctx, &resp.Diagnostics, plan.LaunchOptions)

		// Update Storefront Store Launch Options
		setSTFStoreLaunchOptions(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), launchOptions)

		// Get updated STFStoreService Launch Options
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
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
		roamingAccount := util.ObjectValueToTypedObject[RoamingAccount](ctx, &resp.Diagnostics, plan.RoamingAccount)
		// Update Storefront Roaming Account
		err := setSTFRoamingAccount(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), roamingAccount)
		if err != nil {
			return
		}
		// Get updated Roaming Account
		getResponse, err := getSTFRoamingAccount(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
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

	siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Store Service ",
			"Error message: "+err.Error(),
		)
		return
	}

	STFStoreService, err := getSTFStoreService(ctx, r.client, siteIdInt, state.VirtualPath.ValueString())
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

	farms, err := state.getStoreFarms(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}
	state.RefreshPropertyValues(ctx, &resp.Diagnostics, STFStoreService, farms)
	// Refresh StoreFarmConfiguration
	if !state.FarmSettings.IsNull() {
		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, siteIdInt, state.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations in Read",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}
	//Refresh Storefront StoreService Enumerations
	if !state.EnumerationOptions.IsNull() {
		// Get STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, siteIdInt, state.VirtualPath.ValueString())
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
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, siteIdInt, state.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Launch Options",
				"Error message: "+err.Error(),
			)
			return
		}
		state.RefreshLaunchOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Fetch Pna
	if !state.PNA.IsNull() {
		updatedPna, err := getSTFStorePNA(ctx, r.client, &resp.Diagnostics, siteIdInt, state.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
		state.RefreshPnaValues(ctx, &resp.Diagnostics, updatedPna)
	}

	// Fetch Roaming Account
	if !state.RoamingAccount.IsNull() {
		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getSTFRoamingAccount(ctx, r.client, siteIdInt, state.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF Roaming Account",
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

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
		return
	}
	// Get refreshed STFStoreService
	_, err = getSTFStoreService(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
	if err != nil {
		return
	}

	// Construct the update model
	var getSTFStoreServiceBody = &citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetSiteId(siteIdInt)
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
	updatedSTFStoreService, err := getSTFStoreService(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
	if err != nil {
		return
	}

	var state STFStoreServiceResourceModel
	req.State.Get(ctx, &state)

	existingFarms, err := state.getStoreFarms(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}
	//Update farms
	storeFarms := util.ObjectListToTypedArray[StoreFarm](ctx, &resp.Diagnostics, plan.StoreFarm)
	createAndUpdateStoreFarms(ctx, r.client, &resp.Diagnostics, existingFarms, siteIdInt, plan.VirtualPath.ValueString(), storeFarms)
	farms, err := plan.getStoreFarms(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedSTFStoreService, farms)

	// Update StoreFarmConfiguration
	if !plan.FarmSettings.IsNull() {
		plannedFarmSettings := util.ObjectValueToTypedObject[FarmSettings](ctx, &resp.Diagnostics, plan.FarmSettings)
		err := setFarmSettingsSetRequest(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), plannedFarmSettings)

		if err != nil {
			return
		}

		// Get updated STFStoreFarmConfiguration Settings
		getResponse, err := getFarmSettingsGetRequest(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching STF StoreFarmConfigurations in Update",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshFarmSettings(ctx, &resp.Diagnostics, getResponse)
	}

	//  updated STFStoreService Enumeration Options
	if !plan.EnumerationOptions.IsNull() {
		enumerationOptions := util.ObjectValueToTypedObject[EnumerationOptions](ctx, &resp.Diagnostics, plan.EnumerationOptions)
		// Update Storefront Store Enumeration Options
		setSTFStoreEnumerationOptions(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), enumerationOptions)

		// Get updated STFStoreService Enumeration Options
		getResponse, err := getSTFStoreEnumerationOptions(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching StoreFront Store Enumeration Options",
				"Error message: "+err.Error(),
			)
			return
		}
		plan.RefreshEnumerationOptions(ctx, &resp.Diagnostics, getResponse)
	}

	// Set PNA properties
	if !plan.PNA.IsNull() {
		pna := util.ObjectValueToTypedObject[PNA](ctx, &resp.Diagnostics, plan.PNA)
		setSTFStorePNA(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), pna)

		updatedPna, err := getSTFStorePNA(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString())
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
		launchOptions := util.ObjectValueToTypedObject[LaunchOptions](ctx, &resp.Diagnostics, plan.LaunchOptions)
		// Update Storefront Store Launch Options
		setSTFStoreLaunchOptions(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), launchOptions)

		// Get updated STFStoreService Launch Options
		getResponse, err := getSTFStoreLaunchOptions(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
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
		roamingAccount := util.ObjectValueToTypedObject[RoamingAccount](ctx, &resp.Diagnostics, plan.RoamingAccount)
		// Update Storefront Roaming Account
		update_err := setSTFRoamingAccount(ctx, r.client, &resp.Diagnostics, siteIdInt, plan.VirtualPath.ValueString(), roamingAccount)

		if update_err != nil {
			return
		}

		// Get updated Roaming Account
		getResponse, err := getSTFRoamingAccount(ctx, r.client, siteIdInt, plan.VirtualPath.ValueString())
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
	body.SetVirtualPath(state.VirtualPath.ValueString())

	// Get refreshed STFDeployment, if no STFDeployment found, return
	deployment, err := stf_deployment.GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	// Delete existing STF StoreService
	deleteStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreClearSTFStore(ctx, body)
	_, err = deleteStoreServiceRequest.Execute()
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
func getSTFStoreService(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, siteIdInt int64, VirtualPath string) (*citrixstorefront.STFStoreDetailModel, error) {
	var body citrixstorefront.GetSTFStoreRequestModel

	body.SetSiteId(siteIdInt)

	body.SetVirtualPath(VirtualPath)
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

func setFarmSettingsSetRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId int64, VirtualPath string, plannedFarmSettings FarmSettings) error {
	var farmSettingBody citrixstorefront.SetStoreFarmConfigurationRequestModel

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
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

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
func getFarmSettingsGetRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, siteId int64, VirtualPath string) (*citrixstorefront.StoreFarmConfigurationResponseModel, error) {
	// Generate farmSetting body

	getFarmSettingsBody := citrixstorefront.GetSTFStoreRequestModel{}
	getFarmSettingsBody.SetSiteId(siteId)
	getFarmSettingsBody.SetVirtualPath(VirtualPath)

	// Get Request for STF StoreFarm Configurations
	farmSettingsRequest := client.StorefrontClient.StoreSF.STFStoreFarmGetStoreConfiguration(ctx, getFarmSettingsBody)

	// Execute the request
	getResponse, err := farmSettingsRequest.Execute()
	return &getResponse, err
}

func getSTFStorePNA(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteIdInt int64, VirtualPath string) (*citrixstorefront.STFPna, error) {
	var storeBody citrixstorefront.GetSTFStoreRequestModel

	storeBody.SetSiteId(siteIdInt)
	storeBody.SetVirtualPath(VirtualPath)
	// Fetch updated PNA
	updatedPna, err := client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching updated PNA for StoreFront StoreService",
			"Error message: "+err.Error(),
		)
	}
	return &updatedPna, err
}

// Set STF Store PNA
func setSTFStorePNA(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteIdInt int64, VirtualPath string, pna PNA) error {

	var storeBody citrixstorefront.GetSTFStoreRequestModel
	storeBody.SetSiteId(siteIdInt)
	storeBody.SetVirtualPath(VirtualPath)

	if pna.Enable.ValueBool() {
		// Disable PNA first because of the existing problem from PNA cmdlet
		disablePnaRequest := client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
		err := disablePnaRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error disabling PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}

		// Enable PNA
		var pnaSetBody citrixstorefront.STFStorePnaSetRequestModel
		enablePnaRequest := client.StorefrontClient.StoreSF.STFStoreEnableStorePna(ctx, pnaSetBody, storeBody)
		err = enablePnaRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error enabling PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
	} else {
		// Disable PNA
		disablePnaRequest := client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
		err := disablePnaRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error disabling PNA for StoreFront StoreService",
				"Error message: "+err.Error(),
			)
		}
	}
	return nil
}

// Set Storefront Store Enumeration Options
func setSTFStoreEnumerationOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId int64, VirtualPath string, enumerationOptions EnumerationOptions) error {

	// Generate API request body
	var enumerationOptionsBody citrixstorefront.SetSTFStoreEnumerationOptionsRequestModel

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
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

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
func getSTFStoreEnumerationOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, siteId int64, VirtualPath string) (*citrixstorefront.GetSTFStoreEnumerationOptionsResponseModel, error) {

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

	// Create the client request to Get StoreFront Enumeration Options
	enumerationOptionsRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStoreEnumerationOptions(ctx, getSTFStoreServiceBody)
	// Execute the request
	getResponse, err := enumerationOptionsRequest.Execute()

	return &getResponse, err
}

// Set Storefront Store Launch Options
func setSTFStoreLaunchOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId int64, VirtualPath string, launchOptions LaunchOptions) error {

	// Generate API request body
	var launchOptionsBody citrixstorefront.SetSTFStoreLaunchOptionsRequestModel

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
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

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
func getSTFStoreLaunchOptions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, siteId int64, VirtualPath string) (*citrixstorefront.GetSTFStoreLaunchOptionsResponseModel, error) {
	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

	// Create the client request to Get StoreFront Launch Options
	launchOptionsRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStoreLaunchOptions(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := launchOptionsRequest.Execute()

	return &getResponse, err
}

func setSTFRoamingAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId int64, VirtualPath string, roamingAccountSettings RoamingAccount) error {

	var roamingAccountSettingsBody citrixstorefront.SetSTFRoamingAccountRequestModel

	if !roamingAccountSettings.Published.IsNull() {
		roamingAccountSettingsBody.SetPublished(roamingAccountSettings.Published.ValueBool())
	}

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

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

func getSTFRoamingAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, siteId int64, VirtualPath string) (*citrixstorefront.GetSTFRoamingAccountResponseModel, error) {

	// Generate getSTFStoreService body
	getSTFStoreServiceBody := citrixstorefront.GetSTFStoreRequestModel{}
	getSTFStoreServiceBody.SetSiteId(siteId)
	getSTFStoreServiceBody.SetVirtualPath(VirtualPath)

	// Create the client request to Get StoreFront Gateway Settings
	roamingAccRequest := client.StorefrontClient.StoreSF.STFRoamingAccountGet(ctx, getSTFStoreServiceBody)

	// Execute the request
	getResponse, err := roamingAccRequest.Execute()

	return &getResponse, err
}

func createAndUpdateStoreFarms(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, existingFarms []citrixstorefront.StoreFarmModel, siteId int64, VirtualPath string, farms []StoreFarm) error {

	//set Store for StoreFarm Set Request
	getStoreBody := citrixstorefront.GetSTFStoreRequestModel{}
	getStoreBody.SetSiteId(siteId)
	getStoreBody.SetVirtualPath(VirtualPath)

	planFarmArray := []string{}
	existingFarmArray := []string{}

	for _, farm := range farms {
		planFarmArray = append(planFarmArray, farm.FarmName.ValueString())
	}

	// Delete Roaming Farms that are in not in the plan but are in the existing farms
	for _, existingFarm := range existingFarms {
		var existingFarmName string
		found := false
		if existingFarm.FarmName.Get() == nil || *existingFarm.FarmName.Get() == "" {
			continue
		}
		existingFarmArray = append(existingFarmArray, *existingFarm.FarmName.Get())
		existingFarmName = *existingFarm.FarmName.Get()

		for _, planFarm := range planFarmArray {
			if existingFarmName == planFarm {
				found = true
				break
			}
		}
		//delete farm
		if !found {
			var storeFarmDeleteBody citrixstorefront.GetSTFStoreFarmRequestModel
			storeFarmDeleteBody.SetFarmName(existingFarmName)
			deleteStoreFarmRequest := client.StorefrontClient.StoreSF.STFStoreRemoveStoreFarm(ctx, storeFarmDeleteBody, getStoreBody)
			err := deleteStoreFarmRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error deleting Store Farm during Update",
					"Error message: "+err.Error(),
				)
				return err
			}
		}
	}

	//update or create farms
	for _, farm := range farms {
		found := false
		var storeFarmSetBody = farm.buildStoreFarmBody(ctx, diagnostics)
		//fetch existing StoreFarm to see if a new farm need to be created

		// Check if the farm is already present in the existing farms
		for _, existingFarm := range existingFarmArray {
			if existingFarm == farm.FarmName.ValueString() {
				found = true
				break
			}
		}

		if !found { //create new farm
			createStoreFarmRequest := client.StorefrontClient.StoreSF.STFStoreNewStoreFarm(ctx, storeFarmSetBody, getStoreBody)
			_, err := createStoreFarmRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error creating Store Farm during Update",
					"Error message: "+err.Error(),
				)
				return err
			}
		} else { //otherwise update the existing farm
			setStoreFarmRequest := client.StorefrontClient.StoreSF.STFStoreSetStoreFarm(ctx, storeFarmSetBody, getStoreBody)
			_, err := setStoreFarmRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error updating Store Farm during Update",
					"Error message: "+err.Error(),
				)
				return err
			}
		}
	}
	return nil
}

func (farm StoreFarm) buildStoreFarmBody(ctx context.Context, diagnostics *diag.Diagnostics) citrixstorefront.AddSTFStoreFarmRequestModel {
	var storeFarmSetBody citrixstorefront.AddSTFStoreFarmRequestModel
	if !farm.AllFailedBypassDuration.IsNull() {
		storeFarmSetBody.SetAllFailedBypassDuration(farm.AllFailedBypassDuration.ValueInt64())
	}
	if !farm.FarmName.IsNull() {
		storeFarmSetBody.SetFarmName(farm.FarmName.ValueString())
	}
	if !farm.FarmType.IsNull() {
		storeFarmSetBody.SetFarmType(farm.FarmType.ValueString())
	}
	if !farm.Servers.IsNull() {
		storeFarmSetBody.SetServers(util.StringListToStringArray(ctx, diagnostics, farm.Servers))
	}
	if !farm.Zones.IsNull() {
		storeFarmSetBody.SetZones(util.StringListToStringArray(ctx, diagnostics, farm.Zones))
	}
	if !farm.ServiceUrls.IsNull() {
		urls := util.StringListToStringArray(ctx, diagnostics, farm.ServiceUrls)
		if len(urls) != 0 {
			storeFarmSetBody.SetServiceUrls(util.StringListToStringArray(ctx, diagnostics, farm.ServiceUrls))
		}
	}
	if !farm.LoadBalance.IsNull() {
		storeFarmSetBody.SetLoadBalance(farm.LoadBalance.ValueBool())
	}
	if !farm.BypassDuration.IsNull() {
		storeFarmSetBody.SetBypassDuration(farm.BypassDuration.ValueInt64())
	}
	if !farm.TicketTimeToLive.IsNull() {
		storeFarmSetBody.SetTicketTimeToLive(farm.TicketTimeToLive.ValueInt64())
	}
	if !farm.TransportType.IsNull() {
		storeFarmSetBody.SetTransportType(farm.TransportType.ValueString())
	}
	if !farm.RadeTicketTimeToLive.IsNull() {
		storeFarmSetBody.SetRadeTicketTimeToLive(farm.RadeTicketTimeToLive.ValueInt64())
	}
	if !farm.MaxFailedServersPerRequest.IsNull() {
		storeFarmSetBody.SetMaxFailedServersPerRequest(farm.MaxFailedServersPerRequest.ValueInt64())
	}
	if !farm.Product.IsNull() {
		storeFarmSetBody.SetProduct(farm.Product.ValueString())
	}
	if !farm.FarmGuid.IsNull() {
		storeFarmSetBody.SetFarmGuid(farm.FarmGuid.ValueString())
	}

	if !farm.RestrictPoPs.IsNull() {
		storeFarmSetBody.SetRestrictPoPs(farm.RestrictPoPs.ValueString())
	}
	if !farm.Port.IsNull() {
		storeFarmSetBody.SetPort(farm.Port.ValueInt64())
	}
	if !farm.SSLRelayPort.IsNull() {
		storeFarmSetBody.SetSSLRelayPort(farm.SSLRelayPort.ValueInt64())
	}
	if !farm.TransportType.IsNull() {
		storeFarmSetBody.SetTransportType(farm.TransportType.ValueString())
	}
	if !farm.XMLValidationEnabled.IsNull() {
		storeFarmSetBody.SetXMLValidationEnabled(farm.XMLValidationEnabled.ValueBool())
	}
	if !farm.XMLValidationSecret.IsNull() {
		storeFarmSetBody.SetXMLValidationSecret(farm.XMLValidationSecret.ValueString())
	}
	return storeFarmSetBody
}

func (plan STFStoreServiceResourceModel) getStoreFarms(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixstorefront.StoreFarmModel, error) {
	siteId, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error getting StoreFront StoreService SiteId",
			"Error message: "+err.Error(),
		)
		return nil, err
	}
	getStoreBody := citrixstorefront.GetSTFStoreRequestModel{}
	getStoreBody.SetSiteId(siteId)
	getStoreBody.SetVirtualPath(plan.VirtualPath.ValueString())

	getStoreFarmRequest := client.StorefrontClient.StoreSF.STFStoreGetStoreFarm(ctx, getStoreBody)
	farmWithoutFarmName, err := getStoreFarmRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching Store Farm",
			"Error message: "+err.Error(),
		)
		return nil, err
	}
	return farmWithoutFarmName, err

}
