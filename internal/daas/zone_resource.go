package daas

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &zoneResource{}
	_ resource.ResourceWithConfigure   = &zoneResource{}
	_ resource.ResourceWithImportState = &zoneResource{}
	_ resource.ResourceWithModifyPlan  = &zoneResource{}
)

// NewZoneResource is a helper function to simplify the provider implementation.
func NewZoneResource() resource.Resource {
	return &zoneResource{}
}

// zoneResource is the resource implementation.
type zoneResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *zoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_zone"
}

// Schema defines the schema for the data source.
func (r *zoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the zone.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the zone.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the zone.",
				Optional:    true,
			},
			"metadata": schema.ListNestedAttribute{
				Description: "Metadata of the zone. Cannot be modified in DaaS cloud.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Metadata name.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Metadata value.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *zoneResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *zoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan models.ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.client.AuthConfig.OnPremise {
		// Zone creation is not allowed for cloud. Check if zone exists and import if it does.
		// If zone does not exist, throw an error
		getZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, plan.Name.ValueString())
		zone, _, err := citrixdaasclient.AddRequestData(getZoneRequest, r.client).Execute()

		if err == nil && zone != nil {
			// zone exists. Add it to the state file
			plan = plan.RefreshPropertyValues(zone, false)

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
		} else {
			resp.Diagnostics.AddError(
				"Error creating Zone",
				"Zones and Cloud Connectors are managed only by Citrix Cloud. Ensure you have a resource location manually created and connectors deployed in it and then try again.",
			)
		}

		return
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateZoneRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	if plan.Metadata != nil {
		metadata := util.ParseNameValueStringPairToClientModel(*plan.Metadata)
		body.SetMetadata(*metadata)
	}

	createZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesCreateZone(ctx)
	createZoneRequest = createZoneRequest.CreateZoneRequestModel(body)

	// Create new zone
	httpResp, err := citrixdaasclient.AddRequestData(createZoneRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Zone",
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new zone with zone name
	zone, err := getZone(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(zone, r.client.AuthConfig.OnPremise)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *zoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state models.ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed zone properties from Orchestration
	zone, err := getZone(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(zone, r.client.AuthConfig.OnPremise)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *zoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan models.ZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed zone properties from Orchestration
	zoneId := plan.Id.ValueString()
	zoneName := plan.Name.ValueString()
	_, err := getZone(ctx, r.client, &resp.Diagnostics, zoneId)
	if err != nil {
		return
	}

	// Construct the update model
	var editZoneRequestBody = &citrixorchestration.EditZoneRequestModel{}
	editZoneRequestBody.SetName(plan.Name.ValueString())
	editZoneRequestBody.SetDescription(plan.Description.ValueString())

	if plan.Metadata != nil {
		metadata := util.ParseNameValueStringPairToClientModel(*plan.Metadata)
		editZoneRequestBody.SetMetadata(*metadata)
	}

	// Update zone
	editZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesEditZone(ctx, zoneId)
	editZoneRequest = editZoneRequest.EditZoneRequestModel(*editZoneRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(editZoneRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zone "+zoneName,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated zone from GetZone.
	updatedZone, err := getZone(ctx, r.client, &resp.Diagnostics, zoneId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(updatedZone, r.client.AuthConfig.OnPremise)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *zoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state models.ZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For cloud, just delete from state file. A warning was already added in Modify Plan
	if !r.client.AuthConfig.OnPremise {
		return
	}

	// Delete existing zone
	zoneId := state.Id.ValueString()
	zoneName := state.Name.ValueString()
	deleteZoneRequest := r.client.ApiClient.ZonesAPIsDAAS.ZonesDeleteZone(ctx, zoneId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteZoneRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Zone "+zoneName,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *zoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *zoneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan models.ZoneResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !r.client.AuthConfig.OnPremise && !req.State.Raw.IsNull() {
			resp.Diagnostics.AddWarning(
				"Attempting to modify Zone",
				"Zones and Cloud Connectors are managed only by Citrix Cloud. You may update the description but any metadata changes will be skipped.",
			)
		}
	}

	if req.Plan.Raw.IsNull() && !r.client.AuthConfig.OnPremise {
		resp.Diagnostics.AddWarning(
			"Attempting to delete Zone",
			"Zones and Cloud Connectors are managed only by Citrix Cloud. The requested zone will be deleted from terraform state but you will still need to manually delete these resources "+
				"by logging into Citrix Cloud.",
		)
	}
}

// Gets the zone and logs any errors
func getZone(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, zoneId string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	getZoneRequest := client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, zoneId)
	zone, httpResp, err := citrixdaasclient.AddRequestData(getZoneRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Reading Zone "+zoneId,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
	return zone, err
}
