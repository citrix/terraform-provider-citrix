// Copyright © 2024. Citrix Systems, Inc.

package storefront_server

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &storeFrontServerResource{}
	_ resource.ResourceWithConfigure   = &storeFrontServerResource{}
	_ resource.ResourceWithImportState = &storeFrontServerResource{}
)

// NewStoreFrontServerResource is a helper function to simplify the provider implementation.
func NewStoreFrontServerResource() resource.Resource {
	return &storeFrontServerResource{}
}

// storeFrontServerResource is the resource implementation.
type storeFrontServerResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *storeFrontServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storefront_server"
}

// Schema defines the schema for the resource.
func (r *storeFrontServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StoreFront server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the StoreFront server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the StoreFront server.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the StoreFront server.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL for connecting to the StoreFront server.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates if the StoreFront server is enabled.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *storeFrontServerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *storeFrontServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan StoreFrontServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixorchestration.StoreFrontServerRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetUrl(plan.Url.ValueString())
	body.SetEnabled(plan.Enabled.ValueBool())

	createStoreFrontServerRequest := r.client.ApiClient.StoreFrontServersAPIsDAAS.StoreFrontServersCreateStoreFrontServer(ctx)
	createStoreFrontServerRequest = createStoreFrontServerRequest.StoreFrontServerRequestModel(body).Async(true)

	// Create new StoreFront server
	_, httpResp, err := citrixdaasclient.AddRequestData(createStoreFrontServerRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Server "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error creating StoreFront Server "+plan.Name.ValueString(), &resp.Diagnostics, 10, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Server "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new StoreFront server with StoreFront server name
	sfserver, _, err := getStoreFrontServer(ctx, r.client, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(sfserver)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *storeFrontServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state StoreFrontServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sfServerId := state.Id.ValueString()

	// Get refreshed StoreFront server properties from Orchestration
	sfServer, httpResp, err := getStoreFrontServer(ctx, r.client, sfServerId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Machine Catalog "+sfServerId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)

		return
	}

	state = state.RefreshPropertyValues(sfServer)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *storeFrontServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan StoreFrontServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed StoreFront server properties from Orchestration
	sfServerId := plan.Id.ValueString()
	sfServerName := plan.Name.ValueString()

	// Construct the update model
	var editStoreFrontServerRequestBody = &citrixorchestration.StoreFrontServerRequestModel{}
	editStoreFrontServerRequestBody.SetName(plan.Name.ValueString())
	editStoreFrontServerRequestBody.SetDescription(plan.Description.ValueString())
	editStoreFrontServerRequestBody.SetUrl(plan.Url.ValueString())
	editStoreFrontServerRequestBody.SetEnabled(plan.Enabled.ValueBool())

	// Update StoreFront server
	editStoreFronteServerRequest := r.client.ApiClient.StoreFrontServersAPIsDAAS.StoreFrontServersUpdateStoreFrontServer(ctx, sfServerId)
	editStoreFronteServerRequest = editStoreFronteServerRequest.StoreFrontServerRequestModel(*editStoreFrontServerRequestBody).Async(true)
	_, httpResp, err := citrixdaasclient.AddRequestData(editStoreFronteServerRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Server "+sfServerName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error updating StoreFront Server "+sfServerName, &resp.Diagnostics, 10, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Server "+sfServerName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Fetch updated StoreFront server from getStoreFrontServer.
	updatedStoreFrontServer, _, err := getStoreFrontServer(ctx, r.client, sfServerId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated StoreFront Server "+sfServerName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(updatedStoreFrontServer)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *storeFrontServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state StoreFrontServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing StoreFront server
	sfServerId := state.Id.ValueString()
	sfServerName := state.Name.ValueString()
	deleteStoreFrontServerRequest := r.client.ApiClient.StoreFrontServersAPIsDAAS.StoreFrontServersDeleteStoreFrontServer(ctx, sfServerId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteStoreFrontServerRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Server "+sfServerName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *storeFrontServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Gets the StoreFront server and logs any errors
func getStoreFrontServer(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, storeFrontServerId string) (*citrixorchestration.StoreFrontServerResponseModel, *http.Response, error) {
	getStoreFrontRequest := client.ApiClient.StoreFrontServersAPIsDAAS.StoreFrontServersGetStoreFrontServer(ctx, storeFrontServerId)
	sfServer, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.StoreFrontServerResponseModel](getStoreFrontRequest, client)

	return sfServer, httpResp, err
}
