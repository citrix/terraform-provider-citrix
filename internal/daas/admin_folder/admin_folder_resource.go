// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"
	"fmt"
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
	_ resource.Resource                   = &adminFolderResource{}
	_ resource.ResourceWithConfigure      = &adminFolderResource{}
	_ resource.ResourceWithImportState    = &adminFolderResource{}
	_ resource.ResourceWithValidateConfig = &adminFolderResource{}
	_ resource.ResourceWithModifyPlan     = &adminFolderResource{}
)

// NewAdminFolderResource is a helper function to simplify the provider implementation.
func NewAdminFolderResource() resource.Resource {
	return &adminFolderResource{}
}

// adminFolderResource is the resource implementation.
type adminFolderResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *adminFolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_folder"
}

// Configure adds the provider configured client to the data source.
func (r *adminFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *adminFolderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AdminFolderResourceModel{}.GetSchema()
}

// Create creates the resource and sets the initial Terraform state.
func (r *adminFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createAdminFolderRequest citrixorchestration.CreateAdminFolderRequestModel
	createAdminFolderRequest.SetName(plan.Name.ValueString())
	createAdminFolderRequest.SetPath(plan.ParentPath.ValueString())
	objectIdentifier, err := citrixorchestration.NewAdminFolderObjectIdentifierFromValue(plan.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to create admin folder %s with type %s", plan.Name.ValueString(), plan.Type.ValueString()),
			"Error message: "+err.Error(),
		)
		return
	}
	createAdminFolderRequest.SetObjectIdentifiers([]citrixorchestration.AdminFolderObjectIdentifier{*objectIdentifier})

	addAdminFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersCreateAdminFolder(ctx)
	addAdminFolderRequest = addAdminFolderRequest.CreateAdminFolderRequestModel(createAdminFolderRequest)

	// Create new admin folder
	adminFolderResponse, httpResp, err := citrixdaasclient.AddRequestData(addAdminFolderRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Admin Folder",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, adminFolderResponse)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *adminFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AdminFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed admin properties from Orchestration
	adminFolder, err := readAdminFolder(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, adminFolder)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *adminFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AdminFolderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminFolderId := state.Id.ValueString()
	adminFolderName := state.Name.ValueString()

	// Construct the update model
	var editAdminFolderRequestBody = &citrixorchestration.EditAdminFolderRequestModel{}
	editAdminFolderRequestBody.SetName(plan.Name.ValueString())
	editAdminFolderRequestBody.SetParent(plan.ParentPath.ValueString())

	// Update Admin Folder
	editAdminFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersUpdateAdminFolder(ctx, adminFolderId)
	editAdminFolderRequest = editAdminFolderRequest.EditAdminFolderRequestModel(*editAdminFolderRequestBody)
	adminFolderResponse, httpResp, err := citrixdaasclient.AddRequestData(editAdminFolderRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Admin Folder "+adminFolderName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, adminFolderResponse)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *adminFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AdminFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	adminFolderId := state.Id.ValueString()
	adminFolderName := state.Name.ValueString()

	deleteAdminFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersDeleteAdminFolder(ctx, adminFolderId)
	adminFolderObjects, err := getAdminFolderObjectsEnumFromAdminFolderType(state.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Admin Folder "+adminFolderName,
			"Error message: "+util.ReadClientError(err),
		)
		return
	}
	deleteAdminFolderRequest = deleteAdminFolderRequest.ObjectsToRemove([]citrixorchestration.AdminFolderObjects{adminFolderObjects})
	httpResp, err := citrixdaasclient.AddRequestData(deleteAdminFolderRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Admin Folder "+adminFolderName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *adminFolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *adminFolderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AdminFolderResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

	if !data.Id.IsUnknown() && data.Id.ValueString() == "0" {
		resp.Diagnostics.AddError(
			"Unable to manage admin folder with id `0`",
			"Managing admin folder with id `0` is not supported",
		)
	}
}

func (r *adminFolderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func readAdminFolder(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, adminFolderId string) (*citrixorchestration.AdminFolderResponseModel, error) {
	getAdminFolderRequest := client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolder(ctx, adminFolderId)
	adminFolderResource, _, err := util.ReadResource[*citrixorchestration.AdminFolderResponseModel](getAdminFolderRequest, ctx, client, resp, "Admin Folder", adminFolderId)
	return adminFolderResource, err
}

func getAdminFolderObjectsEnumFromAdminFolderType(adminFolderType string) (citrixorchestration.AdminFolderObjects, error) {
	switch adminFolderType {
	case string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATIONS):
		return citrixorchestration.ADMINFOLDEROBJECTS_APPLICATIONS, nil
	case string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATION_GROUPS):
		return citrixorchestration.ADMINFOLDEROBJECTS_APPLICATION_GROUPS, nil
	case string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_DELIVERY_GROUPS):
		return citrixorchestration.ADMINFOLDEROBJECTS_DELIVERY_GROUPS, nil
	case string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_MACHINE_CATALOGS):
		return citrixorchestration.ADMINFOLDEROBJECTS_MACHINE_CATALOGS, nil
	default:
		return "", fmt.Errorf("unable to parse admin folder object type %s", adminFolderType)
	}
}
