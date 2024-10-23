// Copyright © 2024. Citrix Systems, Inc.
package tags

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &TagResource{}
	_ resource.ResourceWithConfigure      = &TagResource{}
	_ resource.ResourceWithImportState    = &TagResource{}
	_ resource.ResourceWithValidateConfig = &TagResource{}
	_ resource.ResourceWithModifyPlan     = &TagResource{}
)

// NewTagResource is a helper function to simplify the provider implementation.
func NewTagResource() resource.Resource {
	return &TagResource{}
}

// TagResource is the resource implementation.
type TagResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

// Schema defines the schema for the resource.
func (r *TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TagResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create implements resource.Resource.
func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan TagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedScopes := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes)

	var body citrixorchestration.TagRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetScopes(plannedScopes)

	createTagRequest := r.client.ApiClient.TagsAPIsDAAS.TagsCreateTag(ctx)
	createTagRequest = createTagRequest.TagRequestModel(body)

	// Create new tag
	tagResponse, httpResp, err := citrixdaasclient.AddRequestData(createTagRequest, r.client).Execute()

	// In case of error, add it to diagnostics so that the resource gets marked as tainted
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Tag: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new tag detail from remote
	tagDetailResponse, err := getTag(ctx, r.client, &resp.Diagnostics, tagResponse.GetId())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, tagDetailResponse)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state TagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the tag from remote
	tag, err := readTag(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, tag)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan TagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tagId = plan.Id.ValueString()
	var tagName = plan.Name.ValueString()

	plannedScopes := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes)

	// Generate Update API request body from plan
	var body citrixorchestration.TagRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetScopes(plannedScopes)

	// Update tag using orchestration call
	patchTagRequest := r.client.ApiClient.TagsAPIsDAAS.TagsPatchTag(ctx, tagId)
	patchTagRequest = patchTagRequest.TagRequestModel(body)

	tagResponse, httpResp, err := citrixdaasclient.AddRequestData(patchTagRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tag: "+tagName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Fetch updated tag detail using orchestration.
	tagDetailResponse, err := getTag(ctx, r.client, &resp.Diagnostics, tagResponse.GetId())
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, tagDetailResponse)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state TagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing tag
	tagId := state.Id.ValueString()
	tagName := state.Name.ValueString()

	deleteTagRequest := r.client.ApiClient.TagsAPIsDAAS.TagsDeleteTag(ctx, tagId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteTagRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting tag: "+tagName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TagResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data TagResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *TagResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Skip modify plan when doing destroy action
	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan TagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	operation := "updating"
	if create {
		operation = "creating"
	}

	// Only validate tag name availability against tag ID during update operation.
	tagId := ""
	if !create {
		tagId = plan.Id.ValueString()
	}

	isTagNameAvailable, err := checkTagNameAvailability(ctx, r.client, &resp.Diagnostics, tagId, plan.Name.ValueString())
	if err != nil {
		return
	}
	if !isTagNameAvailable && create {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error %s Tag: %s", operation, plan.Name.ValueString()),
			fmt.Sprintf("Tag with name %s already exist", plan.Name.ValueString()),
		)
		return
	}
}

func checkTagNameAvailability(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, tagId string, tagName string) (bool, error) {
	getTagRequest := client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagName)
	tag, httpResp, err := citrixdaasclient.AddRequestData(getTagRequest, client).Execute()
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			return true, nil
		}
		diagnostics.AddError(
			"Error checking tag name availability: "+tagName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return false, err
	}

	// Only fail availability check if the tag name is already in used by another tag
	if !strings.EqualFold(tagId, tag.GetId()) {
		return false, nil
	}

	return true, nil
}

func getTag(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, tagNameOrId string) (*citrixorchestration.TagDetailResponseModel, error) {
	getTagRequest := client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagNameOrId)
	tagResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.TagDetailResponseModel](getTagRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading tag: "+tagNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return tagResponse, err
}

func readTag(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, tagNameOrId string) (*citrixorchestration.TagDetailResponseModel, error) {
	getTagRequest := client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagNameOrId)
	tag, _, err := util.ReadResource[*citrixorchestration.TagDetailResponseModel](getTagRequest, ctx, client, resp, "Tag", tagNameOrId)
	return tag, err
}
