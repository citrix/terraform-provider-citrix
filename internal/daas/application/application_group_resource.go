// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &applicationGroupResource{}
	_ resource.ResourceWithConfigure   = &applicationGroupResource{}
	_ resource.ResourceWithImportState = &applicationGroupResource{}
)

// NewApplicationGroupResource is a helper function to simplify the provider implementation.
func NewApplicationGroupResource() resource.Resource {
	return &applicationGroupResource{}
}

// applicationResource is the resource implementation.
type applicationGroupResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *applicationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_group"
}

// Configure adds the provider configured client to the data source.
func (r *applicationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *applicationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetApplicationGroupSchema()
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createApplicationGroupRequest citrixorchestration.CreateApplicationGroupRequestModel
	createApplicationGroupRequest.SetName(plan.Name.ValueString())
	createApplicationGroupRequest.SetDescription(plan.Description.ValueString())
	createApplicationGroupRequest.SetRestrictToTag(plan.RestrictToTag.ValueString())

	if !plan.IncludedUsers.IsNull() {
		createApplicationGroupRequest.SetIncludedUserFilterEnabled(true)
		includedUsers := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.IncludedUsers)
		includedUserIds, httpResp, err := util.GetUserIdsUsingIdentity(ctx, r.client, includedUsers)
		if err != nil {
			diags.AddError(
				"Error fetching user details for application group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
		createApplicationGroupRequest.SetIncludedUsers(includedUserIds)
	} else {
		createApplicationGroupRequest.SetIncludedUserFilterEnabled(false)
	}

	if !plan.Scopes.IsNull() {
		createApplicationGroupRequest.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	var deliveryGroups []citrixorchestration.PriorityRefRequestModel
	for _, value := range util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups) {
		var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
		deliveryGroupRequestModel.SetItem(value)
		deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
	}

	createApplicationGroupRequest.SetDeliveryGroups(deliveryGroups)

	addApplicationsGroupRequest := r.client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsCreateApplicationGroup(ctx)
	addApplicationsGroupRequest = addApplicationsGroupRequest.CreateApplicationGroupRequestModel(createApplicationGroupRequest)

	// Create new application group
	addAppGroupResp, httpResp, err := citrixdaasclient.AddRequestData(addApplicationsGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Application",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
	// Map response body to schema and populate Computed attribute values

	//Create AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, addAppGroupResp.GetId())
	if err != nil {
		return
	}
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, addAppGroupResp, dgs)
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ApplicationGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed application properties from Orchestration
	applicationGroup, err := readApplicationGroup(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	//AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationGroup.GetId())
	if err != nil {
		return
	}
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, applicationGroup, dgs)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed application properties from Orchestration
	applicationGroupId := plan.Id.ValueString()
	applicationGroupName := plan.Name.ValueString()

	_, err := getApplicationGroup(ctx, r.client, &resp.Diagnostics, applicationGroupId)
	if err != nil {
		return
	}

	// Construct the update model
	var editApplicationGroupRequestBody = &citrixorchestration.EditApplicationGroupRequestModel{}
	editApplicationGroupRequestBody.SetName(plan.Name.ValueString())
	editApplicationGroupRequestBody.SetDescription(plan.Description.ValueString())

	editApplicationGroupRequestBody.SetRestrictToTag(plan.RestrictToTag.ValueString())
	if !plan.IncludedUsers.IsNull() {
		editApplicationGroupRequestBody.SetIncludedUserFilterEnabled(true)
		includedUsers := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.IncludedUsers)
		includedUserIds, httpResp, err := util.GetUserIdsUsingIdentity(ctx, r.client, includedUsers)
		if err != nil {
			diags.AddError(
				"Error fetching user details for application group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
		editApplicationGroupRequestBody.SetIncludedUsers(includedUserIds)
	} else {
		editApplicationGroupRequestBody.SetIncludedUserFilterEnabled(false)
		editApplicationGroupRequestBody.SetIncludedUsers([]string{})
	}

	if !plan.Scopes.IsNull() {
		editApplicationGroupRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	var deliveryGroups []citrixorchestration.PriorityRefRequestModel
	for _, value := range util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups) {
		var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
		deliveryGroupRequestModel.SetItem(value)
		deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
	}

	editApplicationGroupRequestBody.SetDeliveryGroups(deliveryGroups)

	// Update Application
	editApplicationRequest := r.client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsUpdateApplicationGroup(ctx, applicationGroupId)
	editApplicationRequest = editApplicationRequest.EditApplicationGroupRequestModel(*editApplicationGroupRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(editApplicationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Application "+applicationGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Get updated applicationGroup from GetApplication
	applicationGroup, err := getApplicationGroup(ctx, r.client, &resp.Diagnostics, applicationGroupId)
	if err != nil {
		return
	}

	//Create AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationGroup.GetId())
	if err != nil {
		return
	}
	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, applicationGroup, dgs)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ApplicationGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing delivery group
	applicationGroupId := state.Id.ValueString()
	applicationGroupName := state.Name.ValueString()
	deleteApplicationRequest := r.client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsDeleteApplicationGroup(ctx, applicationGroupId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteApplicationRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Application "+applicationGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *applicationGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readApplicationGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, applicationGroupId string) (*citrixorchestration.ApplicationGroupDetailResponseModel, error) {
	getApplicationGroupRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsGetApplicationGroup(ctx, applicationGroupId)
	applicationGroupResource, _, err := util.ReadResource[*citrixorchestration.ApplicationGroupDetailResponseModel](getApplicationGroupRequest, ctx, client, resp, "ApplicationGroup", applicationGroupId)
	return applicationGroupResource, err
}

func getApplicationGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, applicationGroupId string) (*citrixorchestration.ApplicationGroupDetailResponseModel, error) {
	getApplicationRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsGetApplicationGroup(ctx, applicationGroupId)
	applicationGroup, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ApplicationGroupDetailResponseModel](getApplicationRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Application "+applicationGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return applicationGroup, err
}

func getDeliveryGroups(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, applicationGroupId string) (*citrixorchestration.ApplicationGroupDeliveryGroupResponseModelCollection, error) {
	getDeliveryGroupsRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsGetApplicationGroupDeliveryGroups(ctx, applicationGroupId)
	deliveryGroups, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ApplicationGroupDeliveryGroupResponseModelCollection](getDeliveryGroupsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Delivery Groups",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroups, err
}
