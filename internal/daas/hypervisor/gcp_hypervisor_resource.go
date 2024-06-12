// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &gcpHypervisorResource{}
	_ resource.ResourceWithConfigure   = &gcpHypervisorResource{}
	_ resource.ResourceWithImportState = &gcpHypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewGcpHypervisorResource() resource.Resource {
	return &gcpHypervisorResource{}
}

// hypervisorResource is the resource implementation.
type gcpHypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *gcpHypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_hypervisor"
}

// Schema defines the schema for the resource.
func (r *gcpHypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetGcpHypervisorSchema()
}

// Configure adds the provider configured client to the resource.
func (r *gcpHypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *gcpHypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan GcpHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM)
	if !plan.Scopes.IsNull() {
		connectionDetails.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	if plan.ServiceAccountId.IsNull() || plan.ServiceAccountCredentials.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor for GCP",
			"ServiceAccountId/ServiceAccountCredential is missing.",
		)
		return
	}
	connectionDetails.SetServiceAccountId(plan.ServiceAccountId.ValueString())
	connectionDetails.SetServiceAccountCredentials(plan.ServiceAccountCredentials.ValueString())

	// Generate API request body from plan
	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	// Create new hypervisor
	hypervisor, err := CreateHypervisor(ctx, r.client, &resp.Diagnostics, body)
	if err != nil {
		// Directly return. Error logs have been populated in common function.
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, hypervisor)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *gcpHypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state GcpHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := state.Id.ValueString()
	hypervisor, err := readHypervisor(ctx, r.client, resp, hypervisorId)
	if err != nil {
		return
	}

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not an GCP connection type hypervisor.",
		)
		return
	}

	// Overwrite hypervisor with refreshed state
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, hypervisor)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gcpHypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan GcpHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM)
	editHypervisorRequestBody.SetServiceAccountId(plan.ServiceAccountId.ValueString())
	editHypervisorRequestBody.SetServiceAccountCredential(plan.ServiceAccountCredentials.ValueString())
	if !plan.Scopes.IsNull() {
		editHypervisorRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	// Patch hypervisor
	updatedHypervisor, err := UpdateHypervisor(ctx, r.client, &resp.Diagnostics, hypervisor, editHypervisorRequestBody)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedHypervisor)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gcpHypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state GcpHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing hypervisor
	hypervisorId := state.Id.ValueString()
	hypervisorName := state.Name.ValueString()
	deleteHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisor(ctx, hypervisorId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *gcpHypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
