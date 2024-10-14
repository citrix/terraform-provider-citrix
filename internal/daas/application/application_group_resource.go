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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &applicationGroupResource{}
	_ resource.ResourceWithConfigure      = &applicationGroupResource{}
	_ resource.ResourceWithImportState    = &applicationGroupResource{}
	_ resource.ResourceWithValidateConfig = &applicationGroupResource{}
	_ resource.ResourceWithModifyPlan     = &applicationGroupResource{}
)

// NewApplicationGroupResource is a helper function to simplify the provider implementation.
func NewApplicationGroupResource() resource.Resource {
	return &applicationGroupResource{}
}

// applicationResource is the resource implementation.
type applicationGroupResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *applicationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_group"
}

// Configure adds the provider configured client to the resource.
func (r *applicationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the resource.
func (r *applicationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApplicationGroupResourceModel{}.GetSchema()
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

	createApplicationGroupRequest.SetAdminFolder(plan.ApplicationGroupFolderPath.ValueString())

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	createApplicationGroupRequest.SetMetadata(metadata)

	addApplicationsGroupRequest := r.client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsCreateApplicationGroup(ctx)
	addApplicationsGroupRequest = addApplicationsGroupRequest.CreateApplicationGroupRequestModel(createApplicationGroupRequest)

	// Create new application group
	addAppGroupResp, httpResp, err := citrixdaasclient.AddRequestData(addApplicationsGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Application Group "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	applicationGroupId := addAppGroupResp.GetId()
	// Update application group tags
	setApplicationGroupTags(ctx, &resp.Diagnostics, r.client, applicationGroupId, plan.Tags)

	// Map response body to schema and populate Computed attribute values
	// Get updated applicationGroup with getApplicationGroup
	applicationGroup, err := getApplicationGroup(ctx, r.client, &resp.Diagnostics, applicationGroupId)
	if err != nil {
		return
	}

	// Create AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationGroupId)
	if err != nil {
		return
	}

	tags := getApplicationGroupTags(ctx, &resp.Diagnostics, r.client, applicationGroupId)

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, applicationGroup, dgs, tags)
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

	// AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationGroup.GetId())
	if err != nil {
		return
	}

	tags := getApplicationGroupTags(ctx, &resp.Diagnostics, r.client, applicationGroup.GetId())

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, applicationGroup, dgs, tags)

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

	// Get current state
	var state ApplicationGroupResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationGroupId := plan.Id.ValueString()
	applicationGroupName := plan.Name.ValueString()

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
				"Error fetching user details for application group"+applicationGroupName,
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

	editApplicationGroupRequestBody.SetAdminFolder(plan.ApplicationGroupFolderPath.ValueString())

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editApplicationGroupRequestBody.SetMetadata(metadata)

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

	// Update application group tags
	setApplicationGroupTags(ctx, &resp.Diagnostics, r.client, applicationGroupId, plan.Tags)

	// Get updated applicationGroup from getApplicationGroup
	applicationGroup, err := getApplicationGroup(ctx, r.client, &resp.Diagnostics, applicationGroupId)
	if err != nil {
		return
	}

	// Get AppGroup response does not return delivery groups so we are making another call to fetch delivery groups
	dgs, err := getDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationGroup.GetId())
	if err != nil {
		return
	}

	tags := getApplicationGroupTags(ctx, &resp.Diagnostics, r.client, applicationGroupId)

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, applicationGroup, dgs, tags)

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
			"Error deleting Application Group "+applicationGroupName,
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
			"Error Reading Application Group "+applicationGroupId,
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

func (r *applicationGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ApplicationGroupResourceModel
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

func (r *applicationGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func setApplicationGroupTags(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, appGroupId string, tagSet types.Set) {
	setTagsRequestBody := util.ConstructTagsRequestModel(ctx, diagnostics, tagSet)

	setTagsRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsSetApplicationGroupTags(ctx, appGroupId)
	setTagsRequest = setTagsRequest.TagsRequestModel(setTagsRequestBody)

	httpResp, err := citrixdaasclient.AddRequestData(setTagsRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error set tags for Application Group "+appGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		// Continue without return in order to get other attributes refreshed in state
	}
}

func getApplicationGroupTags(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, applicationGroupId string) []string {
	getTagsRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsGetApplicationGroupTags(ctx, applicationGroupId)
	getTagsRequest = getTagsRequest.Fields("Id,Name,Description")
	tagsResp, httpResp, err := citrixdaasclient.AddRequestData(getTagsRequest, client).Execute()
	return util.ProcessTagsResponseCollection(diagnostics, tagsResp, httpResp, err, "Application Group", applicationGroupId)
}
