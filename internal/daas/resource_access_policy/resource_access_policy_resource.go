// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
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
	_ resource.Resource                   = &ResourceAccessPolicyResource{}
	_ resource.ResourceWithConfigure      = &ResourceAccessPolicyResource{}
	_ resource.ResourceWithImportState    = &ResourceAccessPolicyResource{}
	_ resource.ResourceWithValidateConfig = &ResourceAccessPolicyResource{}
	_ resource.ResourceWithModifyPlan     = &ResourceAccessPolicyResource{}
)

// NewResourceAccessPolicyResource is a helper function to simplify the provider implementation.
func NewResourceAccessPolicyResource() resource.Resource {
	return &ResourceAccessPolicyResource{}
}

// ResourceAccessPolicyResource is the resource implementation.
type ResourceAccessPolicyResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *ResourceAccessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_access_policy"
}

// Schema defines the schema for the resource.
func (r *ResourceAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceAccessPolicyResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *ResourceAccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient) //nolint:forcetypeassert // framework guarantee
}

// Create implements resource.Resource.
func (r *ResourceAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyName := plan.Name.ValueString()
	errorSummary := "Error creating Resource Access Policy " + policyName

	tagUid, err := getTagUid(ctx, r.client, &resp.Diagnostics, plan.TagId, errorSummary)
	if err != nil {
		return
	}

	body := plan.buildRequest(ctx, &resp.Diagnostics, tagUid)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := createResourceAccessPolicy(ctx, r.client, &resp.Diagnostics, body, resourceAccessPolicyJobTimeout, errorSummary)
	if err != nil {
		return
	}

	tagId, err := getTagIdOfPolicy(ctx, r.client, &resp.Diagnostics, policy, errorSummary)
	if err != nil {
		// The policy was created. Record its id so Terraform tracks the resource as tainted instead of orphaning it.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.Itoa(int(policy.GetId())))...)
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policy, tagId)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read implements resource.Resource.
func (r *ResourceAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyId := state.Id.ValueString()
	errorSummary := "Error reading Resource Access Policy " + policyId

	policies, err := getResourceAccessPolicies(ctx, r.client, &resp.Diagnostics, errorSummary)
	if err != nil {
		return
	}

	// The API reports a missing policy as a bad request rather than not found, so look it up in the full list instead
	policy := findPolicyById(policies, policyId)
	if policy == nil {
		resp.Diagnostics.AddWarning(
			"Resource Access Policy not found",
			fmt.Sprintf("Resource Access Policy %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", policyId),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	tagId, err := getTagIdOfPolicy(ctx, r.client, &resp.Diagnostics, policy, errorSummary)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, policy, tagId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update implements resource.Resource.
func (r *ResourceAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorSummary := "Error updating Resource Access Policy " + state.Name.ValueString()

	policyId, err := parsePolicyId(&resp.Diagnostics, state.Id.ValueString(), errorSummary)
	if err != nil {
		return
	}

	tagUid, err := getTagUid(ctx, r.client, &resp.Diagnostics, plan.TagId, errorSummary)
	if err != nil {
		return
	}

	body := plan.buildRequest(ctx, &resp.Diagnostics, tagUid)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := updateResourceAccessPolicy(ctx, r.client, &resp.Diagnostics, policyId, body, resourceAccessPolicyJobTimeout, errorSummary)
	if err != nil {
		return
	}

	tagId, err := getTagIdOfPolicy(ctx, r.client, &resp.Diagnostics, policy, errorSummary)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policy, tagId)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete implements resource.Resource.
func (r *ResourceAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorSummary := "Error deleting Resource Access Policy " + state.Name.ValueString()

	policyId, err := parsePolicyId(&resp.Diagnostics, state.Id.ValueString(), errorSummary)
	if err != nil {
		return
	}

	err = deleteResourceAccessPolicy(ctx, r.client, &resp.Diagnostics, policyId, resourceAccessPolicyJobTimeout, errorSummary)
	if err != nil {
		return
	}
}

func (r *ResourceAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceAccessPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ResourceAccessPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ValidateConfig(ctx, &resp.Diagnostics)

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *ResourceAccessPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
	}
}

// createResourceAccessPolicy creates the policy as a background job and returns it once the job completes.
func createResourceAccessPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, body citrixorchestration.ResourceAccessPolicyRequestModel, timeout int32, errorSummary string) (*citrixorchestration.ResourceAccessPolicyResponseModel, error) {
	createRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsCreateResourceAccessPolicy(ctx)
	createRequest = createRequest.ResourceAccessPolicyRequestModel(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(createRequest, client).Async(true).Execute()
	if err != nil {
		addRequestError(diagnostics, httpResp, err, errorSummary)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, errorSummary, diagnostics, timeout)
	if err != nil {
		return nil, err
	}

	// The job does not save the created policy as its result, so look it up by its unique name
	policies, err := getResourceAccessPolicies(ctx, client, diagnostics, errorSummary)
	if err != nil {
		return nil, err
	}

	policy := findPolicyByName(policies, body.GetName())
	if policy == nil {
		err = fmt.Errorf("resource access policy %s was not found after it was created", body.GetName())
		addRequestError(diagnostics, httpResp, err, errorSummary)
	}
	return policy, err
}

// updateResourceAccessPolicy updates the policy as a background job and returns it once the job completes.
func updateResourceAccessPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId int32, body citrixorchestration.ResourceAccessPolicyRequestModel, timeout int32, errorSummary string) (*citrixorchestration.ResourceAccessPolicyResponseModel, error) {
	patchRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsPatchResourceAccessPolicy(ctx, policyId)
	patchRequest = patchRequest.ResourceAccessPolicyRequestModel(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(patchRequest, client).Async(true).Execute()
	if err != nil {
		addRequestError(diagnostics, httpResp, err, errorSummary)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, errorSummary, diagnostics, timeout)
	if err != nil {
		return nil, err
	}

	// The job does not save the updated policy as its result, so look it up by id
	policies, err := getResourceAccessPolicies(ctx, client, diagnostics, errorSummary)
	if err != nil {
		return nil, err
	}

	policy := findPolicyById(policies, strconv.Itoa(int(policyId)))
	if policy == nil {
		err = fmt.Errorf("resource access policy %d was not found after it was updated", policyId)
		addRequestError(diagnostics, httpResp, err, errorSummary)
	}
	return policy, err
}

// deleteResourceAccessPolicy deletes the policy as a background job and waits for the job to complete.
func deleteResourceAccessPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId int32, timeout int32, errorSummary string) error {
	deleteRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDeleteResourceAccessPolicy(ctx, policyId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteRequest, client).Async(true).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return nil
		}
		addRequestError(diagnostics, httpResp, err, errorSummary)
		return err
	}

	return util.ProcessAsyncJobResponse(ctx, client, httpResp, errorSummary, diagnostics, timeout)
}

func addRequestError(diagnostics *diag.Diagnostics, httpResp *http.Response, err error, errorSummary string) {
	diagnostics.AddError(
		errorSummary,
		"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
			"\nError message: "+util.ReadClientError(err),
	)
}

func parsePolicyId(diagnostics *diag.Diagnostics, policyId string, errorSummary string) (int32, error) {
	id, err := strconv.ParseInt(policyId, 10, 32)
	if err != nil {
		diagnostics.AddError(
			errorSummary,
			fmt.Sprintf("Resource Access Policy id %s is not a valid integer. Error message: %s", policyId, err.Error()),
		)
	}
	return int32(id), err
}

func getResourceAccessPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, errorSummary string) ([]citrixorchestration.ResourceAccessPolicyResponseModel, error) {
	getRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetResourceAccessPolicies(ctx)
	policies, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ResourceAccessPolicyResponseModelCollection](getRequest, client)
	if err != nil {
		diagnostics.AddError(
			errorSummary,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policies.GetItems(), nil
}

func findPolicyById(policies []citrixorchestration.ResourceAccessPolicyResponseModel, policyId string) *citrixorchestration.ResourceAccessPolicyResponseModel {
	for _, policy := range policies {
		if strconv.Itoa(int(policy.GetId())) == policyId {
			return &policy
		}
	}
	return nil
}

func findPolicyByName(policies []citrixorchestration.ResourceAccessPolicyResponseModel, policyName string) *citrixorchestration.ResourceAccessPolicyResponseModel {
	for _, policy := range policies {
		if strings.EqualFold(policy.GetName(), policyName) {
			return &policy
		}
	}
	return nil
}

func getTag(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, tagNameOrId string, errorSummary string) (*citrixorchestration.TagDetailResponseModel, error) {
	getTagRequest := client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagNameOrId)
	tag, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.TagDetailResponseModel](getTagRequest, client)
	if err != nil {
		diagnostics.AddError(
			errorSummary,
			fmt.Sprintf("Error reading tag %s.", tagNameOrId)+
				"\nTransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
	return tag, err
}

// getTagUid resolves the tag GUID in the configuration to the numeric UID the policy API takes.
// Returns nil when no tag is configured.
func getTagUid(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, tagId types.String, errorSummary string) (*int32, error) {
	if tagId.ValueString() == "" {
		return nil, nil //nolint:nilnil // a policy without a tag has no tag UID
	}

	tag, err := getTag(ctx, client, diagnostics, tagId.ValueString(), errorSummary)
	if err != nil {
		return nil, err
	}

	if !tag.Uid.IsSet() || tag.Uid.Get() == nil {
		err = fmt.Errorf("tag %s did not report a UID", tagId.ValueString())
		diagnostics.AddError(errorSummary, err.Error())
		return nil, err
	}

	return tag.Uid.Get(), nil
}

// getTagIdOfPolicy resolves the tag the policy reports by UID and name to the tag's GUID.
// Returns an empty string when the policy has no tag.
func getTagIdOfPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policy *citrixorchestration.ResourceAccessPolicyResponseModel, errorSummary string) (string, error) {
	if !policy.TagUid.IsSet() || policy.TagUid.Get() == nil {
		return "", nil
	}

	tag, err := getTag(ctx, client, diagnostics, policy.GetTagName(), errorSummary)
	if err != nil {
		return "", err
	}

	return tag.GetId(), nil
}
