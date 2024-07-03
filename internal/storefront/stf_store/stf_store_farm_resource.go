// Copyright © 2024. Citrix Systems, Inc.
package stf_store

import (
	"context"
	"fmt"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfStoreFarmResource{}
	_ resource.ResourceWithConfigure   = &stfStoreFarmResource{}
	_ resource.ResourceWithImportState = &stfStoreFarmResource{}
)

// stfStoreFarmResource is a helper function to simplify the provider implementation.
func NewSTFStoreFarmResource() resource.Resource {
	return &stfStoreFarmResource{}
}

// stfStoreFarmResource is the resource implementation.
type stfStoreFarmResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfStoreFarmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_store_farm"
}

// Configure adds the provider configured client to the resource.
func (r *stfStoreFarmResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfStoreFarmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreFarmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storeFarmAddBody citrixstorefront.AddSTFStoreFarmRequestModel
	if !plan.AllFailedBypassDuration.IsNull() {
		storeFarmAddBody.SetAllFailedBypassDuration(plan.AllFailedBypassDuration.ValueInt64())
	}
	if !plan.FarmName.IsNull() {
		storeFarmAddBody.SetFarmName(plan.FarmName.ValueString())
	}
	if !plan.FarmType.IsNull() {
		storeFarmAddBody.SetFarmType(plan.FarmType.ValueString())
	}
	if !plan.Servers.IsNull() {
		storeFarmAddBody.SetServers(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.Servers))
	}
	if !plan.Zones.IsNull() {
		storeFarmAddBody.SetZones(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.Zones))
	}
	if !plan.ServiceUrls.IsNull() {
		urls := util.StringListToStringArray(ctx, &resp.Diagnostics, plan.ServiceUrls)
		if len(urls) != 0 {
			storeFarmAddBody.SetServiceUrls(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.ServiceUrls))
		}
	}
	if !plan.LoadBalance.IsNull() {
		storeFarmAddBody.SetLoadBalance(plan.LoadBalance.ValueBool())
	}
	if !plan.BypassDuration.IsNull() {
		storeFarmAddBody.SetBypassDuration(plan.BypassDuration.ValueInt64())
	}
	if !plan.TicketTimeToLive.IsNull() {
		storeFarmAddBody.SetTicketTimeToLive(plan.TicketTimeToLive.ValueInt64())
	}
	if !plan.TransportType.IsNull() {
		storeFarmAddBody.SetTransportType(plan.TransportType.ValueInt64())
	}
	if !plan.RadeTicketTimeToLive.IsNull() {
		storeFarmAddBody.SetRadeTicketTimeToLive(plan.RadeTicketTimeToLive.ValueInt64())
	}
	if !plan.MaxFailedServersPerRequest.IsNull() {
		storeFarmAddBody.SetMaxFailedServersPerRequest(plan.MaxFailedServersPerRequest.ValueInt64())
	}
	if !plan.Product.IsNull() {
		storeFarmAddBody.SetProduct(plan.Product.ValueString())
	}
	if !plan.FarmGuid.IsNull() {
		storeFarmAddBody.SetFarmGuid(plan.FarmGuid.ValueString())
	}

	if !plan.RestrictPoPs.IsNull() {
		storeFarmAddBody.SetRestrictPoPs(plan.RestrictPoPs.ValueString())
	}
	if !plan.Port.IsNull() {
		storeFarmAddBody.SetPort(plan.Port.ValueInt64())
	}
	if !plan.SSLRelayPort.IsNull() {
		storeFarmAddBody.SetSSLRelayPort(plan.SSLRelayPort.ValueInt64())
	}
	if !plan.TransportType.IsNull() {
		storeFarmAddBody.SetTransportType(plan.TransportType.ValueInt64())
	}
	if !plan.XMLValidationEnabled.IsNull() {
		storeFarmAddBody.SetXMLValidationEnabled(plan.XMLValidationEnabled.ValueBool())
	}
	if !plan.XMLValidationSecret.IsNull() {
		storeFarmAddBody.SetXMLValidationSecret(plan.XMLValidationSecret.ValueString())
	}

	//set Store for StoreFarm Add Request
	getStoreBody := citrixstorefront.GetSTFStoreRequestModel{}
	getStoreBody.SetVirtualPath(plan.StoreService.ValueString())
	addStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreNewStoreFarm(ctx, storeFarmAddBody, getStoreBody)

	_, err := addStoreFarmRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Store Farm",
			"\nError message: "+err.Error(),
		)
		return
	}

	// Get Store Farm
	var storeFarmGetBody citrixstorefront.GetSTFStoreFarmRequestModel
	storeFarmGetBody.SetFarmName(plan.FarmName.ValueString())
	getStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreGetStoreFarm(ctx, storeFarmGetBody, getStoreBody)
	storeFarm, err := getStoreFarmRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Store Farm",
			"\nError message: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, storeFarm)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stfStoreFarmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFStoreFarmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storeFarmGetBody citrixstorefront.GetSTFStoreFarmRequestModel
	storeFarmGetBody.SetFarmName(state.FarmName.ValueString())
	var getStoreBody citrixstorefront.GetSTFStoreRequestModel
	getStoreBody.SetVirtualPath(state.StoreService.ValueString())

	getStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreGetStoreFarm(ctx, storeFarmGetBody, getStoreBody)
	storeFarm, err := getStoreFarmRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			resp.Diagnostics.AddWarning(
				"StoreFront Store Farm not found",
				"StoreFront Store Farm was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
			)
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(
				"Error reading Store Farm",
				"\nError message: "+err.Error(),
			)
			return
		}
	}

	state.RefreshPropertyValues(ctx, &resp.Diagnostics, storeFarm)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfStoreFarmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreFarmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storeFarmSetBody citrixstorefront.SetSTFStoreFarmRequestModel
	if !plan.AllFailedBypassDuration.IsNull() {
		storeFarmSetBody.SetAllFailedBypassDuration(plan.AllFailedBypassDuration.ValueInt64())
	}
	if !plan.FarmName.IsNull() {
		storeFarmSetBody.SetFarmName(plan.FarmName.ValueString())
	}
	if !plan.FarmType.IsNull() {
		storeFarmSetBody.SetFarmType(plan.FarmType.ValueString())
	}
	if !plan.Servers.IsNull() {
		storeFarmSetBody.SetServers(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.Servers))
	}
	if !plan.Zones.IsNull() {
		storeFarmSetBody.SetZones(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.Zones))
	}
	if !plan.ServiceUrls.IsNull() {
		urls := util.StringListToStringArray(ctx, &resp.Diagnostics, plan.ServiceUrls)
		if len(urls) != 0 {
			storeFarmSetBody.SetServiceUrls(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.ServiceUrls))
		}
	}
	if !plan.LoadBalance.IsNull() {
		storeFarmSetBody.SetLoadBalance(plan.LoadBalance.ValueBool())
	}
	if !plan.BypassDuration.IsNull() {
		storeFarmSetBody.SetBypassDuration(plan.BypassDuration.ValueInt64())
	}
	if !plan.TicketTimeToLive.IsNull() {
		storeFarmSetBody.SetTicketTimeToLive(plan.TicketTimeToLive.ValueInt64())
	}
	if !plan.TransportType.IsNull() {
		storeFarmSetBody.SetTransportType(plan.TransportType.ValueInt64())
	}
	if !plan.RadeTicketTimeToLive.IsNull() {
		storeFarmSetBody.SetRadeTicketTimeToLive(plan.RadeTicketTimeToLive.ValueInt64())
	}
	if !plan.MaxFailedServersPerRequest.IsNull() {
		storeFarmSetBody.SetMaxFailedServersPerRequest(plan.MaxFailedServersPerRequest.ValueInt64())
	}
	if !plan.Product.IsNull() {
		storeFarmSetBody.SetProduct(plan.Product.ValueString())
	}
	if !plan.FarmGuid.IsNull() {
		storeFarmSetBody.SetFarmGuid(plan.FarmGuid.ValueString())
	}

	if !plan.RestrictPoPs.IsNull() {
		storeFarmSetBody.SetRestrictPoPs(plan.RestrictPoPs.ValueString())
	}
	if !plan.Port.IsNull() {
		storeFarmSetBody.SetPort(plan.Port.ValueInt64())
	}
	if !plan.SSLRelayPort.IsNull() {
		storeFarmSetBody.SetSSLRelayPort(plan.SSLRelayPort.ValueInt64())
	}
	if !plan.TransportType.IsNull() {
		storeFarmSetBody.SetTransportType(plan.TransportType.ValueInt64())
	}
	if !plan.XMLValidationEnabled.IsNull() {
		storeFarmSetBody.SetXMLValidationEnabled(plan.XMLValidationEnabled.ValueBool())
	}
	if !plan.XMLValidationSecret.IsNull() {
		storeFarmSetBody.SetXMLValidationSecret(plan.XMLValidationSecret.ValueString())
	}

	//set Store for StoreFarm Set Request
	getStoreBody := citrixstorefront.GetSTFStoreRequestModel{}
	getStoreBody.SetVirtualPath(plan.StoreService.ValueString())

	setStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreSetStoreFarm(ctx, storeFarmSetBody, getStoreBody)
	_, err := setStoreFarmRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Store Farm",
			"\nError message: "+err.Error(),
		)
		return
	}

	var storeFarmGetBody citrixstorefront.GetSTFStoreFarmRequestModel
	getStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreGetStoreFarm(ctx, storeFarmGetBody, getStoreBody)
	storeFarm, err := getStoreFarmRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Store Farm after Upate",
			"\nError message: "+err.Error(),
		)
		return
	}

	// Update resource state with updated property values
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, storeFarm)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfStoreFarmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFStoreFarmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deleteStoreFarmbody citrixstorefront.GetSTFStoreFarmRequestModel
	deleteStoreFarmbody.SetFarmName(state.FarmName.ValueString())
	var getStoreBody citrixstorefront.GetSTFStoreRequestModel
	getStoreBody.SetVirtualPath(state.StoreService.ValueString())

	// Delete existing STF Store
	deleteStoreFarmRequest := r.client.StorefrontClient.StoreSF.STFStoreRemoveStoreFarm(ctx, deleteStoreFarmbody, getStoreBody)
	err := deleteStoreFarmRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Store Farm",
			"\nError message: "+err.Error(),
		)
		return
	}
}

func (r *stfStoreFarmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `store_virtual_path,farm_name`, got: %q", req.ID),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("store_virtual_path"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("farm_name"), idSegments[1])...)
}
