// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &applicationFolderResource{}
	_ resource.ResourceWithConfigure      = &applicationFolderResource{}
	_ resource.ResourceWithImportState    = &applicationFolderResource{}
	_ resource.ResourceWithValidateConfig = &applicationFolderResource{}
	_ resource.ResourceWithModifyPlan     = &applicationFolderResource{}
)

// NewApplicationFolderResource is a helper function to simplify the provider implementation.
func NewApplicationFolderResource() resource.Resource {
	return &applicationFolderResource{}
}

// applicationFolderResource is the resource implementation.
type applicationFolderResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *applicationFolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_folder"
}

// Configure adds the provider configured client to the data source.
func (r *applicationFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *applicationFolderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApplicationFolderResourceModel{}.GetSchema()
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createApplicationFolderRequest citrixorchestration.CreateAdminFolderRequestModel
	createApplicationFolderRequest.SetName(plan.Name.ValueString())
	createApplicationFolderRequest.SetPath(plan.ParentPath.ValueString())
	createApplicationFolderRequest.SetObjectIdentifiers([]citrixorchestration.AdminFolderObjectIdentifier{citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATIONS})

	addApplicationFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersCreateAdminFolder(ctx)
	addApplicationFolderRequest = addApplicationFolderRequest.CreateAdminFolderRequestModel(createApplicationFolderRequest)

	// Create new application folder
	application_folder, httpResp, err := citrixdaasclient.AddRequestData(addApplicationFolderRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Application Folder",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(application_folder)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ApplicationFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed application properties from Orchestration
	application_folder, err := readApplicationFolder(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(application_folder)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ApplicationFolderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationFolderId := state.Id.ValueString()
	applicationFoldeName := state.Name.ValueString()

	// Construct the update model
	var editApplicationFolderRequestBody = &citrixorchestration.EditAdminFolderRequestModel{}
	editApplicationFolderRequestBody.SetName(plan.Name.ValueString())
	editApplicationFolderRequestBody.SetParent(plan.ParentPath.ValueString())

	// Update Application Folder
	editApplicationFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersUpdateAdminFolder(ctx, applicationFolderId)
	editApplicationFolderRequest = editApplicationFolderRequest.EditAdminFolderRequestModel(*editApplicationFolderRequestBody)
	application_folder, httpResp, err := citrixdaasclient.AddRequestData(editApplicationFolderRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Application Folder "+applicationFoldeName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(application_folder)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ApplicationFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationFolderId := state.Id.ValueString()
	applicationFolderName := state.Name.ValueString()

	deleteApplicationFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersDeleteAdminFolder(ctx, applicationFolderId)
	deleteApplicationFolderRequest = deleteApplicationFolderRequest.ObjectsToRemove([]citrixorchestration.AdminFolderObjects{citrixorchestration.ADMINFOLDEROBJECTS_APPLICATIONS})
	httpResp, err := citrixdaasclient.AddRequestData(deleteApplicationFolderRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Application Folder "+applicationFolderName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *applicationFolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readApplicationFolder(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, applicationFolderId string) (*citrixorchestration.AdminFolderResponseModel, error) {
	getApplicationFolderRequest := client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolder(ctx, applicationFolderId)
	applicationFolderResource, _, err := util.ReadResource[*citrixorchestration.AdminFolderResponseModel](getApplicationFolderRequest, ctx, client, resp, "Application Folder", applicationFolderId)
	return applicationFolderResource, err
}

func (r *applicationFolderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ApplicationFolderResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *applicationFolderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
