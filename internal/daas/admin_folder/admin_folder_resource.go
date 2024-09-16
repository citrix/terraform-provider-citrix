// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	adminFolderTypesArray, err := getAdminFolderObjectIdentifierArrayFromTypeSet(ctx, &resp.Diagnostics, plan.Name.ValueString(), plan.Type)
	if err != nil {
		return
	}
	createAdminFolderRequest.SetObjectIdentifiers(adminFolderTypesArray)

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

	adminFolderId := plan.Id.ValueString()

	adminFolderResource, err := getAdminFolder(ctx, r.client, &resp.Diagnostics, adminFolderId)
	if err != nil {
		return
	}

	// Construct the update model
	var editAdminFolderRequestBody = &citrixorchestration.EditAdminFolderRequestModel{}
	editAdminFolderRequestBody.SetName(plan.Name.ValueString())

	var parentPath = strings.TrimSuffix(adminFolderResource.GetPath(), adminFolderResource.GetName()+"\\")
	if plan.ParentPath.ValueString() != parentPath {
		editAdminFolderRequestBody.SetParent(plan.ParentPath.ValueString())
	}

	metadataArray := getAdminFolderMetadataArrayFromTypeSet(ctx, &resp.Diagnostics, plan.Type)
	editAdminFolderRequestBody.SetMetadata(metadataArray)

	// Update Admin Folder
	editAdminFolderRequest := r.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersUpdateAdminFolder(ctx, adminFolderId)
	editAdminFolderRequest = editAdminFolderRequest.EditAdminFolderRequestModel(*editAdminFolderRequestBody)
	adminFolderResponse, httpResp, err := citrixdaasclient.AddRequestData(editAdminFolderRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Admin Folder "+plan.Name.ValueString(),
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

	adminFolderObjectsEnumArray, err := getAdminFolderObjectsEnumArrayFromAdminFolderTypeSet(ctx, &resp.Diagnostics, state.Name.ValueString(), state.Type)
	if err != nil {
		return
	}
	deleteAdminFolderRequest = deleteAdminFolderRequest.ObjectsToRemove(adminFolderObjectsEnumArray)
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

func getAdminFolder(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, adminFolderIdOrPath string) (*citrixorchestration.AdminFolderResponseModel, error) {
	getAdminFolderRequest := client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolder(ctx, adminFolderIdOrPath)
	adminFolderResource, httpResp, err := citrixdaasclient.AddRequestData(getAdminFolderRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error reading Admin Folder "+adminFolderIdOrPath,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return adminFolderResource, nil
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

func getAdminFolderObjectsEnumArrayFromAdminFolderTypeSet(ctx context.Context, diagnostics *diag.Diagnostics, adminFolderName string, stateTypes types.Set) ([]citrixorchestration.AdminFolderObjects, error) {
	types := util.StringSetToStringArray(ctx, diagnostics, stateTypes)
	adminFolderObjectsEnumArray := []citrixorchestration.AdminFolderObjects{}
	for _, adminFolderType := range types {
		objectEnum, err := getAdminFolderObjectsEnumFromAdminFolderType(adminFolderType)
		if err != nil {
			diagnostics.AddError(
				fmt.Sprintf("Unable to get admin folder objects enum with type %s for admin folder %s", adminFolderType, adminFolderName),
				"Error message: "+err.Error(),
			)
			return adminFolderObjectsEnumArray, err
		}
		adminFolderObjectsEnumArray = append(adminFolderObjectsEnumArray, objectEnum)
	}
	return adminFolderObjectsEnumArray, nil
}

func getAdminFolderObjectIdentifierArrayFromTypeSet(ctx context.Context, diagnostics *diag.Diagnostics, adminFolderName string, plannedTypes types.Set) ([]citrixorchestration.AdminFolderObjectIdentifier, error) {
	types := util.StringSetToStringArray(ctx, diagnostics, plannedTypes)
	adminFolderTypesArray := []citrixorchestration.AdminFolderObjectIdentifier{}
	for _, adminFolderType := range types {
		objectIdentifier, err := citrixorchestration.NewAdminFolderObjectIdentifierFromValue(adminFolderType)
		if err != nil {
			diagnostics.AddError(
				fmt.Sprintf("Unable to create admin folder %s with type %s", adminFolderName, adminFolderType),
				"Error message: "+err.Error(),
			)
			return adminFolderTypesArray, err
		}
		adminFolderTypesArray = append(adminFolderTypesArray, *objectIdentifier)
	}
	return adminFolderTypesArray, nil
}

func getAdminFolderMetadataArrayFromTypeSet(ctx context.Context, diagnostics *diag.Diagnostics, plannedTypes types.Set) []citrixorchestration.NameValueStringPairModel {
	types := util.StringSetToStringArray(ctx, diagnostics, plannedTypes)
	metadataArray := []citrixorchestration.NameValueStringPairModel{}
	for _, adminFolderType := range types {
		metadata := citrixorchestration.NameValueStringPairModel{}
		metadata.SetName(adminFolderType)
		metadata.SetValue("true")

		metadataArray = append(metadataArray, metadata)
	}
	return metadataArray
}
