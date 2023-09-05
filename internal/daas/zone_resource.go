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

	// Generate API request body from plan
	var body citrixorchestration.CreateZoneRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	if plan.Metadata != nil {
		metadata := util.ParseNameValueStringPairToClientModel(*plan.Metadata)
		body.SetMetadata(*metadata)
	}

	createZoneRequest := r.client.ApiClient.ZonesTPApi.ZonesTPCreateZone(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	token, _ := r.client.SignIn()
	createZoneRequest = createZoneRequest.Authorization(token)
	createZoneRequest = createZoneRequest.Request(body)

	// Create new zone
	_, err := createZoneRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating zone",
			"Error message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new zone with zone name
	zone, err := getZone(ctx, r.client, resp.Diagnostics, plan.Name.ValueString())
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
	zone, err := getZone(ctx, r.client, resp.Diagnostics, state.Id.ValueString())
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
	_, err := getZone(ctx, r.client, resp.Diagnostics, zoneId)
	if err != nil {
		return
	}

	// Construct the update model
	var editZoneRequestBody = &citrixorchestration.EditZoneRequestModel{}
	editZoneRequestBody.SetName(plan.Name.ValueString())
	editZoneRequestBody.SetDescription(plan.Description.ValueString())
	metadata := util.ParseNameValueStringPairToClientModel(*plan.Metadata)
	editZoneRequestBody.SetMetadata(*metadata)

	// Update zone
	editZoneRequest := r.client.ApiClient.ZonesTPApi.ZonesTPEditZone(ctx, zoneId, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	token, _ := r.client.SignIn()
	editZoneRequest = editZoneRequest.Authorization(token)
	editZoneRequest = editZoneRequest.Request(*editZoneRequestBody)
	_, err = editZoneRequest.Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Zone",
			"Error message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated zone from GetZone.
	updatedZone, err := getZone(ctx, r.client, resp.Diagnostics, zoneId)
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

	// Delete existing zone
	zoneId := state.Id.ValueString()
	deleteZoneRequest := r.client.ApiClient.ZonesTPApi.ZonesTPDeleteZone(ctx, zoneId, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	token, _ := r.client.SignIn()
	deleteZoneRequest = deleteZoneRequest.Authorization(token)
	_, err := deleteZoneRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Zone with Id "+state.Id.ValueString(),
			"Error message: "+util.ReadClientError(err),
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

		if !r.client.AuthConfig.OnPremise && plan.Metadata != nil {
			resp.Diagnostics.AddWarning(
				"Cannot Modify Cloud Zone Metadata",
				"Skipping zone metadata changes. In DaaS cloud, zones are synchronized with Citrix Cloud Resource Locations. "+
					"Zone metadata is readonly.",
			)
		}
	}
}

// Gets the zone and logs any errors
func getZone(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, zoneId string) (*citrixorchestration.ZoneDetailResponseModel, error) {
	getZoneRequest := client.ApiClient.ZonesTPApi.ZonesTPGetZone(ctx, zoneId, client.ClientConfig.CustomerId, client.ClientConfig.SiteId)
	token, _ := client.SignIn()
	getZoneRequest = getZoneRequest.Authorization(token)
	zone, resp, err := getZoneRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Reading Zone with name or Id "+zoneId,
			"TransactionId: "+resp.Header.Get("Citrix-TransactionId")+"\nError message: "+util.ReadClientError(err),
		)
	}
	return zone, err
}

func zoneNameOrIdCheck[T models.ResourceWithZoneModel](client *citrixdaasclient.CitrixDaasClient, ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() && !req.State.Raw.IsNull() {
		var plan T
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Retrieve values from state
		var state T
		diags2 := req.State.Get(ctx, &state)
		resp.Diagnostics.Append(diags2...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.GetZone().IsNull() && !state.GetZone().IsNull() && plan.GetZone().ValueString() != state.GetZone().ValueString() {
			zone, err := getZone(ctx, client, resp.Diagnostics, state.GetZone().ValueString())
			if err != nil {
				return
			}

			if zone.Id == plan.GetZone().ValueString() || zone.Name == plan.GetZone().ValueString() {
				// plan and state are referring to the same zone, fix up the plane
				resp.Plan.SetAttribute(ctx, path.Root("zone"), state.GetZone().ValueString())
			}
		}
	}
}
