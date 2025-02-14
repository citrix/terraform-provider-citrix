// Copyright © 2024. Citrix Systems, Inc.

package zone

import (
	"context"
	"net/http"
	"time"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/citrixcloud/resource_locations"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &zoneResource{}
	_ resource.ResourceWithConfigure      = &zoneResource{}
	_ resource.ResourceWithImportState    = &zoneResource{}
	_ resource.ResourceWithValidateConfig = &zoneResource{}
	_ resource.ResourceWithModifyPlan     = &zoneResource{}
)

// NewZoneResource is a helper function to simplify the provider implementation.
func NewZoneResource() resource.Resource {
	return &zoneResource{}
}

// zoneResource is the resource implementation.
type zoneResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *zoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

// Schema defines the schema for the resource.
func (r *zoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ZoneResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *zoneResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *zoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.client.AuthConfig.OnPremises && !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		// For Cloud customers, resource location ID is required and name is not allowed
		resp.Diagnostics.AddError(
			"Error Managing Zone",
			"For Citrix Cloud customers, Zone has to be created in terraform by providing Resource Location ID. Zone name is not allowed.",
		)
		return
	}

	if r.client.AuthConfig.OnPremises && !plan.ResourceLocationId.IsNull() && !plan.ResourceLocationId.IsUnknown() {
		// For On-prem customers, name is required and resource location ID is not allowed
		resp.Diagnostics.AddError(
			"Error Managing Zone",
			"For On-prem customers, Zone has to be created in terraform by providing zone name. Resource location ID is not allowed.",
		)
		return
	}

	if !r.client.AuthConfig.OnPremises {
		// Zone creation is not allowed for cloud. Check if zone exists and import if it does.
		// If zone does not exist, throw an error
		if plan.ResourceLocationId.IsNull() || plan.ResourceLocationId.IsUnknown() {
			resp.Diagnostics.AddError(
				"Error creating Zone",
				"Resource Location id is required to create a Zone.",
			)
			return
		}

		// If resource location id is provided, get resource location from Citrix Cloud and extract zone name to import zone
		// Get resource location from remote using id
		resourceLocation, err := resource_locations.GetResourceLocation(ctx, r.client, &resp.Diagnostics, plan.ResourceLocationId.ValueString())
		if err != nil {
			return
		}

		zoneName := resourceLocation.GetName()

		zone, err := pollZone(ctx, r.client, zoneName)
		if err == nil && zone != nil {
			// zone exists. run an update on zone so that the description and metadata are updated
			zone, err = plan.updateZoneAfterCreate(ctx, r.client, &resp.Diagnostics, zone.GetId(), zoneName)

			if err != nil {
				return
			}

			// Add it to the state file
			plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, zone, false)

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
		} else {
			resp.Diagnostics.AddError(
				"Error creating Zone",
				"Zones and Cloud Connectors are managed only by Citrix Cloud. Ensure you have a resource location created manually or via terraform, then try again. If zone is for on-premises hypervisor, make sure you have cloud connectors deployed in it.",
			)
		}

		return
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateZoneRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	body.SetMetadata(metadata)

	createZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesCreateZone(ctx)
	createZoneRequest = createZoneRequest.CreateZoneRequestModel(body)

	// Create new zone
	httpResp, err := citrixdaasclient.AddRequestData(createZoneRequest, r.client).Async(true).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Zone",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error creating zone "+plan.Name.ValueString(), &resp.Diagnostics, 5, true)
	if err != nil {
		return
	}

	// Try getting the new zone with zone name
	zone, err := getZone(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, zone, r.client.AuthConfig.OnPremises)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *zoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed zone properties from Orchestration
	zone, err := readZone(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, zone, r.client.AuthConfig.OnPremises)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *zoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve state from plan
	var state ZoneResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the update model
	var editZoneRequestBody = &citrixorchestration.EditZoneRequestModel{}
	zoneName := plan.Name.ValueString()
	if !r.client.AuthConfig.OnPremises {
		// for cloud, there will be no name in plan, use name from state
		zoneName = state.Name.ValueString()
	}
	editZoneRequestBody.SetName(zoneName)
	editZoneRequestBody.SetDescription(plan.Description.ValueString())
	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editZoneRequestBody.SetMetadata(metadata)

	// Update zone
	editZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesEditZone(ctx, state.Id.ValueString())
	editZoneRequest = editZoneRequest.EditZoneRequestModel(*editZoneRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(editZoneRequest, r.client).Async(true).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Zone "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error updating zone "+plan.Name.ValueString(), &resp.Diagnostics, 5, true)
	if err != nil {
		return
	}

	// Fetch updated zone from GetZone.
	updatedZone, err := getZone(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedZone, r.client.AuthConfig.OnPremises)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *zoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For cloud, just delete from state file. A warning was already added in Modify Plan
	if !r.client.AuthConfig.OnPremises {
		return
	}

	// Delete existing zone
	zoneId := state.Id.ValueString()
	zoneName := state.Name.ValueString()
	deleteZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesDeleteZone(ctx, zoneId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteZoneRequest, r.client).Async(true).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Zone "+zoneName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting zone "+state.Name.ValueString(), &resp.Diagnostics, 5, true)
	if err != nil {
		return
	}
}

func (r *zoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *zoneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan ZoneResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if req.Plan.Raw.IsNull() && !r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddWarning(
			"Attempting to delete Zone",
			"Zones and Cloud Connectors are managed only by Citrix Cloud. The requested zone will be deleted from terraform state but you will still need to manually delete these resources "+
				"by logging into Citrix Cloud.",
		)
	}
}

func (r *zoneResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ZoneResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Gets the zone and logs any errors
func getZone(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, zoneId string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	getZoneRequest := client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, zoneId)
	zone, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ZoneDetailResponseModel](getZoneRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Zone "+zoneId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return zone, err
}

func readZone(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, zoneId string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	getZoneRequest := client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, zoneId)
	zone, _, err := util.ReadResource[*citrixorchestration.ZoneDetailResponseModel](getZoneRequest, ctx, client, resp, "Zone", zoneId)
	return zone, err
}

func pollZone(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, zoneName string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	// default polling to every 10 seconds
	pollInterval := 10
	startTime := time.Now()
	getZoneRequest := client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, zoneName)

	var zone *citrixorchestration.ZoneDetailResponseModel
	var err error
	for {
		// Zone sync should be completed within 8 minutes
		if time.Since(startTime) > time.Minute*time.Duration(8) {
			break
		}

		zone, httpResp, err := citrixdaasclient.AddRequestData(getZoneRequest, client).Execute()
		if err == nil {
			// Zone sync completed. Return the zone
			return zone, nil
		} else if httpResp.StatusCode != http.StatusNotFound {
			// GET Zone call failed with an error other than 404. Return the error
			return zone, err
		}

		time.Sleep(time.Second * time.Duration(pollInterval))
		continue
	}

	return zone, err
}

func (r ZoneResourceModel) updateZoneAfterCreate(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, zoneId, zoneName string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	var editZoneRequestBody citrixorchestration.EditZoneRequestModel
	editZoneRequestBody.SetName(zoneName)
	editZoneRequestBody.SetDescription(r.Description.ValueString())
	metadata := util.GetMetadataRequestModel(ctx, diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata))
	editZoneRequestBody.SetMetadata(metadata)

	// Update zone
	editZoneRequest := client.ApiClient.ZonesAPIsDAAS.ZonesEditZone(ctx, zoneId)
	editZoneRequest = editZoneRequest.EditZoneRequestModel(editZoneRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(editZoneRequest, client).Async(true).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating Zone "+zoneName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating zone "+zoneName, diagnostics, 5, true)
	if err != nil {
		return nil, err
	}

	// Fetch updated zone from GetZone.
	updatedZone, err := getZone(ctx, client, diagnostics, zoneId)
	if err != nil {
		return nil, err
	}

	return updatedZone, err
}
