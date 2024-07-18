// Copyright © 2024. Citrix Systems, Inc.
package stf_store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Define a map to store mutexes for each siteId
var mutexes = make(map[string]*sync.Mutex)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfXenappDefaultStoreResource{}
	_ resource.ResourceWithConfigure   = &stfXenappDefaultStoreResource{}
	_ resource.ResourceWithImportState = &stfXenappDefaultStoreResource{}
)

// NewXenappDefaultStoreResource is a helper function to simplify the provider implementation.
func NewXenappDefaultStoreResource() resource.Resource {
	return &stfXenappDefaultStoreResource{}
}

// XenappDefaultStoreResource is the resource implementation.
type stfXenappDefaultStoreResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfXenappDefaultStoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_xenapp_default_store"
}

// Configure adds the provider configured client to the resource.
func (r *stfXenappDefaultStoreResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfXenappDefaultStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFXenappDefaultStoreResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the mutex for each deployment because only one instance are allowed for each store deployment
	mutex, ok := mutexes[plan.StoreSiteID.ValueString()]
	if !ok {
		// If the mutex does not exist, create it
		mutex = &sync.Mutex{}
		mutexes[plan.StoreSiteID.ValueString()] = mutex
	}

	if !mutex.TryLock() {
		resp.Diagnostics.AddError(
			"Error creating XenappDefaultStoreResource",
			"Another instance is already being created/updated",
		)
		return
	}
	defer mutex.Unlock()

	var storeBody citrixstorefront.GetSTFStoreRequestModel
	storeBody.SetVirtualPath(plan.StoreSiteID.ValueString())
	storeBody.SetVirtualPath(plan.StoreVirtualPath.ValueString())

	var enableDefaultStoreBody citrixstorefront.STFStorePnaSetRequestModel
	enableDefaultStoreBody.SetDefaultPnaService(true)
	enablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreEnableStorePna(ctx, enableDefaultStoreBody, storeBody)
	err := enablePnaRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Default XenApp Store ",
			"\nError message: "+err.Error(),
		)
	}

	// Fetch Default Store
	defaultStore, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching updated XenApp Store",
			"\nError message: "+err.Error(),
		)
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, defaultStore)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *stfXenappDefaultStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFXenappDefaultStoreResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storeBody citrixstorefront.GetSTFStoreRequestModel
	storeBody.SetVirtualPath(state.StoreSiteID.ValueString())
	storeBody.SetVirtualPath(state.StoreVirtualPath.ValueString())

	getSTFStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreGetSTFStore(ctx, storeBody)

	// Make sure store exists before fetching default store
	_, err := getSTFStoreServiceRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			resp.Diagnostics.AddWarning(
				"StoreFront Store not found for the XenApp Default Store",
				"StoreFront Store was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
	}

	// Fetch Default Store
	defaultStore, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching updated XenApp Store",
			"Error message: "+err.Error(),
		)
	}

	state.RefreshPropertyValues(ctx, &resp.Diagnostics, defaultStore)
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfXenappDefaultStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	// Retrieve values from plan
	var plan STFXenappDefaultStoreResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the mutex for each deployment because only one instace are allowed for each store deployment
	mutex, ok := mutexes[plan.StoreSiteID.ValueString()]
	if !ok {
		// If the mutex does not exist, create it
		mutex = &sync.Mutex{}
		mutexes[plan.StoreSiteID.ValueString()] = mutex
	}

	if !mutex.TryLock() {
		resp.Diagnostics.AddError(
			"Error Updating XenappDefaultStoreResource",
			"Another instance is already being created/updated",
		)
		return
	}
	defer mutex.Unlock()

	var storeBody citrixstorefront.GetSTFStoreRequestModel
	storeBody.SetVirtualPath(plan.StoreSiteID.ValueString())
	storeBody.SetVirtualPath(plan.StoreVirtualPath.ValueString())

	var enableDefaultStoreBody citrixstorefront.STFStorePnaSetRequestModel
	enableDefaultStoreBody.SetDefaultPnaService(true)
	enablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreEnableStorePna(ctx, enableDefaultStoreBody, storeBody)
	err := enablePnaRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Default XenApp Store ",
			"Error message: "+err.Error(),
		)
	}

	// Fetch Default Store
	defaultStore, err := r.client.StorefrontClient.StoreSF.STFStoreGetStorePna(ctx, storeBody).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching updated XenApp Store",
			"Error message: "+err.Error(),
		)
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, defaultStore)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfXenappDefaultStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFXenappDefaultStoreResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the mutex for each deployment because only one instace are allowed for each store deployment
	mutex, ok := mutexes[state.StoreSiteID.ValueString()]
	if !ok {
		// If the mutex does not exist, create it
		mutex = &sync.Mutex{}
		mutexes[state.StoreSiteID.ValueString()] = mutex
	}

	if !mutex.TryLock() {
		resp.Diagnostics.AddError(
			"Error creating XenappDefaultStoreResource",
			"Another instance is already being created/updated",
		)
		return
	}
	defer mutex.Unlock()

	var storeBody citrixstorefront.GetSTFStoreRequestModel
	siteIdInt, err := strconv.ParseInt(state.StoreSiteID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront StoreService ",
			"Error message: "+err.Error(),
		)
		return
	}

	storeBody.SetSiteId(siteIdInt)
	storeBody.SetVirtualPath(state.StoreVirtualPath.ValueString())
	disablePnaRequest := r.client.StorefrontClient.StoreSF.STFStoreDisableStorePna(ctx, storeBody)
	err = disablePnaRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error disabling XenApp Default Store",
			"Error message: "+err.Error(),
		)
	}
}

func (r *stfXenappDefaultStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("store_site_id"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("store_virtual_path"), idSegments[1])...)
}
