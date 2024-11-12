// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	"context"

	resourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &resourceLocationResource{}
	_ resource.ResourceWithConfigure      = &resourceLocationResource{}
	_ resource.ResourceWithImportState    = &resourceLocationResource{}
	_ resource.ResourceWithValidateConfig = &resourceLocationResource{}
	_ resource.ResourceWithModifyPlan     = &resourceLocationResource{}
)

// NewResourceLocationResource is a helper function to simplify the provider implementation.
func NewResourceLocationResource() resource.Resource {
	return &resourceLocationResource{}
}

// resourceLocationResource is the resource implementation.
type resourceLocationResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *resourceLocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_resource_location"
}

// Schema defines the schema for the resource.
func (r *resourceLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceLocationModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *resourceLocationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *resourceLocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ResourceLocationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource location
	var body resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel
	body.SetName(plan.Name.ValueString())
	body.SetInternalOnly(plan.InternalOnly.ValueBool())
	body.SetTimeZone(plan.TimeZone.ValueString())

	createResourceLocationRequest := r.client.ResourceLocationsClient.LocationsDAAS.LocationsCreate(ctx)
	createResourceLocationRequest = createResourceLocationRequest.Model(body)

	// Create resource location
	resourceLocation, httpResp, err := citrixdaasclient.AddRequestData(createResourceLocationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource location",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	resourceLocationId := resourceLocation.GetId()

	// Get resource location from remote using id
	resourceLocation, err = GetResourceLocation(ctx, r.client, &resp.Diagnostics, resourceLocationId)
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(resourceLocation)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *resourceLocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ResourceLocationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get resource location from remote using id
	resourceLocation, err := readResourceLocation(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(resourceLocation)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *resourceLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ResourceLocationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var body resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationUpdateModel
	body.SetName(plan.Name.ValueString())
	body.SetInternalOnly(plan.InternalOnly.ValueBool())
	body.SetTimeZone(plan.TimeZone.ValueString())

	updateResourceLocationRequest := r.client.ResourceLocationsClient.LocationsDAAS.LocationsUpdate(ctx, plan.Id.ValueString())
	updateResourceLocationRequest = updateResourceLocationRequest.Model(body)

	// Update resource location
	_, httpResp, err := citrixdaasclient.AddRequestData(updateResourceLocationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating resource location with id: "+plan.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get resource location from remote using id
	updatedResourceLocation, err := GetResourceLocation(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(updatedResourceLocation)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *resourceLocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ResourceLocationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResourceLocationRequest := r.client.ResourceLocationsClient.LocationsDAAS.LocationsDelete(ctx, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteResourceLocationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting resource location with id: "+state.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *resourceLocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readResourceLocation(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, resourceLocationId string) (*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel, error) {
	getResourceLocationRequest := client.ResourceLocationsClient.LocationsDAAS.LocationsGet(ctx, resourceLocationId)
	resourceLocation, _, err := util.ReadResource[*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel](getResourceLocationRequest, ctx, client, resp, "Resource Location", resourceLocationId)
	return resourceLocation, err
}

func (r *resourceLocationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ResourceLocationModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Resource Location is a cloud concept which is not supported for on-prem environment
func (r *resourceLocationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ResourceLocationsClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing resource location", "Resource locations are only supported for Cloud customers. On-premises customers can use the Zone resource directly.")
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan ResourceLocationModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
