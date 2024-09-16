// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &policySetResource{}
	_ resource.ResourceWithConfigure      = &policySetResource{}
	_ resource.ResourceWithImportState    = &policySetResource{}
	_ resource.ResourceWithValidateConfig = &policySetResource{}
	_ resource.ResourceWithModifyPlan     = &policySetResource{}
)

// NewPolicySetResource is a helper function to simplify the provider implementation.
func NewPolicySetResource() resource.Resource {
	return &policySetResource{}
}

// policySetResource is the resource implementation.
type policySetResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *policySetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
	operation := "updating"
	if create {
		operation = "creating"
	}

	// Retrieve values from plan
	var plan PolicySetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate DDC Version
	errorSummary := fmt.Sprintf("Error %s Policy Set", operation)
	feature := "Policy Set resource"
	isDdcVersionSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 118, 7, 41, errorSummary, feature)

	if !isDdcVersionSupported {
		return
	}

	if !plan.Scopes.IsNull() {
		if slices.Contains(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes), util.AllScopeId) {
			resp.Diagnostics.AddError(
				"Error "+operation+" Policy Set",
				fmt.Sprintf("Id `%s` for Scope `All` should not be added to the policy set scopes.", util.AllScopeId),
			)
		}
	}

	plannedPolicies := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)

	for _, policy := range plannedPolicies {
		policySettings := util.ObjectSetToTypedArray[PolicySettingModel](ctx, &resp.Diagnostics, policy.PolicySettings)

		for _, setting := range policySettings {
			if strings.EqualFold(setting.Value.ValueString(), "true") ||
				strings.EqualFold(setting.Value.ValueString(), "1") ||
				strings.EqualFold(setting.Value.ValueString(), "false") ||
				strings.EqualFold(setting.Value.ValueString(), "0") {
				resp.Diagnostics.AddError(
					"Error "+operation+" Policy Set",
					"Please specify boolean policy setting value with the 'enabled' attribute.",
				)
			}
		}
	}
}

// Metadata returns the data source type name.
func (r *policySetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set"
}

// Configure implements resource.ResourceWithConfigure.
func (r *policySetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (*policySetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicySetResourceModel{}.GetSchema()
}

// Create implements resource.Resource.
func (r *policySetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySets, err := getPolicySets(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	for _, policySet := range policySets {
		if strings.EqualFold(policySet.GetName(), plan.Name.ValueString()) {
			resp.Diagnostics.AddError(
				"Error Creating Policy Set",
				"Policy Set with name "+plan.Name.ValueString()+" already exists",
			)
			return
		}
	}

	var createPolicySetRequestBody = &citrixorchestration.PolicySetRequest{}
	createPolicySetRequestBody.SetName(plan.Name.ValueString())
	createPolicySetRequestBody.SetDescription(plan.Description.ValueString())
	createPolicySetRequestBody.SetPolicySetType(plan.Type.ValueString())

	// Use scope names instead of IDs for create request to support 2311
	plannedScopeNames, err := util.FetchScopeNamesByIds(ctx, resp.Diagnostics, r.client, util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	if err != nil {
		return
	}
	createPolicySetRequestBody.SetScopes(plannedScopeNames)

	createPolicySetRequest := r.client.ApiClient.GpoDAAS.GpoCreateGpoPolicySet(ctx)
	createPolicySetRequest = createPolicySetRequest.PolicySetRequest(*createPolicySetRequestBody)

	// Create new Policy Set
	policySetResponse, httpResp, err := citrixdaasclient.AddRequestData(createPolicySetRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy Set",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	plannedPolicies := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
	// Create new policies
	batchRequestModel, err := constructCreatePolicyBatchRequestModel(ctx, &resp.Diagnostics, r.client, plannedPolicies, policySetResponse.GetPolicySetGuid(), policySetResponse.GetName())
	if err != nil {
		return
	}

	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Policies to Policy Set "+policySetResponse.GetName(),
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	if successfulJobs < len(plan.Policies.Elements()) {
		errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(plan.Policies.Elements()))
		resp.Diagnostics.AddError(
			"Error adding Policies to Policy Set "+policySetResponse.GetName(),
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
	}

	// Try getting the new policy set with policy set GUID
	policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	if len(policySet.Policies) > 0 {
		// Update Policy Priority
		plannedPolicies = util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
		policyPriorityRequest := constructPolicyPriorityRequest(ctx, r.client, policySet, plannedPolicies)
		// Update policy priorities in the Policy Set
		policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, r.client).Execute()
		if err != nil || !policyPriorityResponse {
			resp.Diagnostics.AddError(
				"Error Changing Policy Priorities in Policy Set "+policySet.GetPolicySetGuid(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}
	}

	// Try getting the new policy set with policy set GUID
	policySet, err = getPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, r.client, policySet.GetScopes())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policies, policySetScopes)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *policySetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state PolicySetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySet, err := readPolicySet(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	policies, err := readPolicies(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, r.client, policySet.GetScopes())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policies, policySetScopes)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *policySetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySetId := plan.Id.ValueString()
	policySetName := plan.Name.ValueString()

	policySets, err := getPolicySets(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	for _, policySet := range policySets {
		if strings.EqualFold(policySet.GetName(), policySetName) && !strings.EqualFold(policySet.GetPolicySetGuid(), policySetId) {
			resp.Diagnostics.AddError(
				"Error Updating Policy Set "+policySetId,
				"Policy Set with name "+policySetName+" already exists",
			)
			return
		}
	}

	stateAndPlanDiff, _ := req.State.Raw.Diff(req.Plan.Raw)
	var policiesModified bool
	for _, diff := range stateAndPlanDiff {
		if diff.Path.Steps()[0].Equal(tftypes.AttributeName("policies")) {
			policiesModified = true
			break
		}
	}

	if policiesModified {
		// Get Remote Policies
		policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetId)
		if err != nil {
			return
		}

		// Setup batch requests
		deletePolicyBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating policies in policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not update policies within the policy set, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		// Clean up all the policies, settings, and filters in policy set
		for index, policy := range policies.Items {
			relativeUrl := fmt.Sprintf("/gpo/policies/%s", policy.GetPolicyGuid())

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(fmt.Sprintf("deletePolicy%s", strconv.Itoa(index)))
			batchRequestItem.SetMethod(http.MethodDelete)
			batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetHeaders(batchApiHeaders)
			deletePolicyBatchRequestItems = append(deletePolicyBatchRequestItems, batchRequestItem)
		}

		var deletePolicyBatchRequestModel citrixorchestration.BatchRequestModel
		deletePolicyBatchRequestModel.SetItems(deletePolicyBatchRequestItems)

		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, deletePolicyBatchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error cleanup Policies in Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(deletePolicyBatchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while deleting policies in the Policy Set. %d of %d policies were deleted from the Policy Set.", successfulJobs, len(deletePolicyBatchRequestItems))
			resp.Diagnostics.AddError(
				"Error deleting Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return
		}

		// Create all the policies, settings, and filters in the plan
		plannedPolicies := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
		createPoliciesBatchRequestModel, err := constructCreatePolicyBatchRequestModel(ctx, &resp.Diagnostics, r.client, plannedPolicies, plan.Id.ValueString(), plan.Name.ValueString())
		if err != nil {
			return
		}

		successfulJobs, txId, err = citrixdaasclient.PerformBatchOperation(ctx, r.client, createPoliciesBatchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(createPoliciesBatchRequestModel.Items) {
			errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(createPoliciesBatchRequestModel.Items))
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return
		}

		// Update policy priority
		policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
		if err != nil {
			return
		}

		if len(policySet.Policies) > 0 {
			plannedPolicies = util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
			policyPriorityRequest := constructPolicyPriorityRequest(ctx, r.client, policySet, plannedPolicies)
			// Update policy priorities in the Policy Set
			policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, r.client).Execute()
			if err != nil || !policyPriorityResponse {
				resp.Diagnostics.AddError(
					"Error updating Policy Priorities in Policy Set "+policySet.GetPolicySetGuid(),
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}
		}
	}

	// Construct the update model
	var editPolicySetRequestBody = &citrixorchestration.PolicySetRequest{}
	editPolicySetRequestBody.SetName(policySetName)
	editPolicySetRequestBody.SetDescription(plan.Description.ValueString())
	editPolicySetRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))

	editPolicySetRequest := r.client.ApiClient.GpoDAAS.GpoUpdateGpoPolicySet(ctx, policySetId)
	editPolicySetRequest = editPolicySetRequest.PolicySetRequest(*editPolicySetRequestBody)

	// Update Policy Set
	httpResp, err := citrixdaasclient.AddRequestData(editPolicySetRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy Set",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new policy set with policy set GUID
	policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, r.client, policySet.GetScopes())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policies, policySetScopes)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *policySetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetId := state.Id.ValueString()
	policySetName := state.Name.ValueString()
	// Get delivery groups and check if the current policy set is assigned to one of them
	getDeliveryGroupsRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroups(ctx)
	deliveryGroups, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupsRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not get delivery group associated with the policy set, unexpected error: "+util.ReadClientError(err),
		)
		return
	}
	associatedDeliveryGroupIds := []string{}
	for _, deliveryGroup := range deliveryGroups.Items {
		if deliveryGroup.GetPolicySetGuid() == policySetId {
			associatedDeliveryGroupIds = append(associatedDeliveryGroupIds, deliveryGroup.GetId())
		}
	}

	if len(associatedDeliveryGroupIds) > 0 {
		// Unassign policy set from delivery groups to unblock delete operation
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not remove policy set from delivery groups, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		var editDeliveryGroupRequestBody citrixorchestration.EditDeliveryGroupRequestModel
		editDeliveryGroupRequestBody.SetPolicySetGuid(util.DefaultSitePolicySetId)
		editDeliveryGroupStringBody, err := util.ConvertToString(editDeliveryGroupRequestBody)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error policy set "+policySetName+" from delivery groups",
				"An unexpected error occurred: "+err.Error(),
			)
			return
		}

		for index, deliveryGroupId := range associatedDeliveryGroupIds {
			relativeUrl := fmt.Sprintf("/DeliveryGroups/%s", deliveryGroupId)
			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(strconv.Itoa(index))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetBody(editDeliveryGroupStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}

		if len(batchRequestItems) > 0 {
			// If there are any machines that need to be put in maintenance mode
			var batchRequestModel citrixorchestration.BatchRequestModel
			batchRequestModel.SetItems(batchRequestItems)
			successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
					"TransactionId: "+txId+
						"\nError Message: "+util.ReadClientError(err),
				)
				return
			}

			if successfulJobs < len(batchRequestItems) {
				errMsg := fmt.Sprintf("An error occurred removing policy set %s from delivery groups. Unassigned from %d of %d delivery groups.", policySetName, successfulJobs, len(batchRequestItems))
				resp.Diagnostics.AddError(
					"Error deleting Policy Set "+policySetName,
					"TransactionId: "+txId+
						"\n"+errMsg,
				)

				return
			}
		}
	}

	// Delete existing Policy Set
	deletePolicySetRequest := r.client.ApiClient.GpoDAAS.GpoDeleteGpoPolicySet(ctx, policySetId)
	httpResp, err = citrixdaasclient.AddRequestData(deletePolicySetRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error Deleting Policy Set "+policySetName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *policySetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Gets the policy set and logs any errors
func getPolicySets(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.PolicySetResponse, error) {
	getPolicySetsRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySets(ctx)
	policySets, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfPolicySetResponse](getPolicySetsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Sets",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policySets.Items, err
}

func getPolicySet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.PolicySetResponse, error) {
	getPolicySetRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySet(ctx, policySetId)
	getPolicySetRequest = getPolicySetRequest.WithPolicies(true)
	policySet, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.PolicySetResponse](getPolicySetRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Set "+policySetId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return policySet, err
}

func readPolicySet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, policySetId string) (*citrixorchestration.PolicySetResponse, error) {
	getPolicySetRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySet(ctx, policySetId)
	getPolicySetRequest = getPolicySetRequest.WithPolicies(true)
	policySet, _, err := util.ReadResource[*citrixorchestration.PolicySetResponse](getPolicySetRequest, ctx, client, resp, "PolicySet", policySetId)
	return policySet, err
}

// Gets the policy set and logs any errors
func getPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.CollectionEnvelopeOfPolicyResponse, error) {
	getPoliciesRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicies(ctx)
	getPoliciesRequest = getPoliciesRequest.PolicySetGuid(policySetId)
	getPoliciesRequest = getPoliciesRequest.WithFilters(true)
	getPoliciesRequest = getPoliciesRequest.WithSettings(true)
	policies, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfPolicyResponse](getPoliciesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policies in Policy Set "+policySetId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return policies, err
}

func readPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, policySetId string) (*citrixorchestration.CollectionEnvelopeOfPolicyResponse, error) {
	getPoliciesRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicies(ctx)
	getPoliciesRequest = getPoliciesRequest.PolicySetGuid(policySetId)
	getPoliciesRequest = getPoliciesRequest.WithFilters(true)
	getPoliciesRequest = getPoliciesRequest.WithSettings(true)
	policies, _, err := util.ReadResource[*citrixorchestration.CollectionEnvelopeOfPolicyResponse](getPoliciesRequest, ctx, client, resp, "Policies", policySetId)
	return policies, err
}

func generateBatchApiHeaders(client *citrixdaasclient.CitrixDaasClient) ([]citrixorchestration.NameValueStringPairModel, *http.Response, error) {
	headers := []citrixorchestration.NameValueStringPairModel{}

	cwsAuthToken, httpResp, err := client.SignIn()
	var token string
	if err != nil {
		return headers, httpResp, err
	}

	if cwsAuthToken != "" {
		token = strings.Split(cwsAuthToken, "=")[1]
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("Authorization")
		header.SetValue("Bearer " + token)
		headers = append(headers, header)
	}

	return headers, httpResp, err
}

func constructCreatePolicyBatchRequestModel(ctx context.Context, diags *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policiesToCreate []PolicyModel, policySetGuid string, policySetName string) (citrixorchestration.BatchRequestModel, error) {
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	var batchRequestModel citrixorchestration.BatchRequestModel

	for policyIndex, policyToCreate := range policiesToCreate {
		var createPolicyRequest = citrixorchestration.PolicyRequest{}
		createPolicyRequest.SetName(policyToCreate.Name.ValueString())
		createPolicyRequest.SetDescription(policyToCreate.Description.ValueString())
		createPolicyRequest.SetIsEnabled(policyToCreate.Enabled.ValueBool())
		// Add Policy Settings
		policySettings := []citrixorchestration.SettingRequest{}
		policySettingsToCreate := util.ObjectSetToTypedArray[PolicySettingModel](ctx, diags, policyToCreate.PolicySettings)
		for _, policySetting := range policySettingsToCreate {
			settingRequest := citrixorchestration.SettingRequest{}
			settingRequest.SetSettingName(policySetting.Name.ValueString())
			settingRequest.SetUseDefault(policySetting.UseDefault.ValueBool())
			if policySetting.Value.ValueString() != "" {
				settingRequest.SetSettingValue(policySetting.Value.ValueString())
			} else {
				if policySetting.Enabled.ValueBool() {
					settingRequest.SetSettingValue("1")
				} else {
					settingRequest.SetSettingValue("0")
				}
			}
			policySettings = append(policySettings, settingRequest)
		}
		createPolicyRequest.SetSettings(policySettings)

		// Add Policy Filters
		policyFilters, err := constructPolicyFilterRequests(ctx, diags, client, policyToCreate)
		if err != nil {
			return batchRequestModel, err
		}
		createPolicyRequest.SetFilters(policyFilters)

		createPolicyRequestBodyString, err := util.ConvertToString(createPolicyRequest)
		if err != nil {
			diags.AddError(
				"Error adding Policy "+policyToCreate.Name.ValueString()+" to Policy Set "+policySetName,
				"An unexpected error occurred: "+err.Error(),
			)
			return batchRequestModel, err
		}

		batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
		if err != nil {
			diags.AddError(
				"Error deleting policy from policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not delete policies within the policy set to be updated, unexpected error: "+util.ReadClientError(err),
			)
			return batchRequestModel, err
		}

		relativeUrl := fmt.Sprintf("/gpo/policies?policySetGuid=%s", policySetGuid)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("createPolicy%d", policyIndex))
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(createPolicyRequestBodyString)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	batchRequestModel.SetItems(batchRequestItems)
	return batchRequestModel, nil
}

func constructPolicyFilterRequests(ctx context.Context, diags *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policy PolicyModel) ([]citrixorchestration.FilterRequest, error) {
	filterRequests := []citrixorchestration.FilterRequest{}

	serverValue := ""
	if client.AuthConfig.OnPremises || !client.AuthConfig.ApiGateway {
		serverValue = client.ApiClient.GetConfig().Host
	} else {
		serverValue = fmt.Sprintf("%s.xendesktop.net", client.ClientConfig.CustomerId)
	}

	if !policy.AccessControlFilters.IsNull() && len(policy.AccessControlFilters.Elements()) > 0 {
		accessControlFilters := util.ObjectSetToTypedArray[AccessControlFilterModel](ctx, diags, policy.AccessControlFilters)
		for _, accessControlFilter := range accessControlFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("AccessControl")

			policyFilterDataClientModel := PolicyFilterGatewayDataClientModel{
				Connection: accessControlFilter.Connection,
				Condition:  accessControlFilter.Condition,
				Gateway:    accessControlFilter.Gateway,
			}

			policyFilterDataJson, err := json.Marshal(policyFilterDataClientModel)
			if err != nil {
				diags.AddError(
					"Error adding Access Control Policy Filter to Policy Set. ",
					"An unexpected error occurred: "+err.Error(),
				)
				return filterRequests, err
			}
			filterRequest.SetFilterData(string(policyFilterDataJson))
			filterRequest.SetIsAllowed(accessControlFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(accessControlFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.BranchRepeaterFilter.IsNull() {
		branchRepeaterFilter := util.ObjectValueToTypedObject[BranchRepeaterFilterModel](ctx, diags, policy.BranchRepeaterFilter)
		branchnRepeaterFilterRequest := citrixorchestration.FilterRequest{}
		branchnRepeaterFilterRequest.SetFilterType("BranchRepeater")
		branchnRepeaterFilterRequest.SetIsAllowed(branchRepeaterFilter.Allowed.ValueBool())
		branchnRepeaterFilterRequest.SetIsEnabled(branchRepeaterFilter.Enabled.ValueBool())
		filterRequests = append(filterRequests, branchnRepeaterFilterRequest)
	}

	if !policy.ClientIPFilters.IsNull() && len(policy.ClientIPFilters.Elements()) > 0 {
		clientIpFilters := util.ObjectSetToTypedArray[ClientIPFilterModel](ctx, diags, policy.ClientIPFilters)
		for _, clientIpFilter := range clientIpFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("ClientIP")

			filterRequest.SetFilterData(clientIpFilter.IpAddress.ValueString())
			filterRequest.SetIsAllowed(clientIpFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(clientIpFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.ClientNameFilters.IsNull() && len(policy.ClientNameFilters.Elements()) > 0 {
		clientNameFilters := util.ObjectSetToTypedArray[ClientNameFilterModel](ctx, diags, policy.ClientNameFilters)
		for _, clientName := range clientNameFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("ClientName")

			filterRequest.SetFilterData(clientName.ClientName.ValueString())
			filterRequest.SetIsAllowed(clientName.Allowed.ValueBool())
			filterRequest.SetIsEnabled(clientName.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.DeliveryGroupFilters.IsNull() && len(policy.DeliveryGroupFilters.Elements()) > 0 {
		deliveryGroupFilters := util.ObjectSetToTypedArray[DeliveryGroupFilterModel](ctx, diags, policy.DeliveryGroupFilters)
		for _, deliveryGroupFilter := range deliveryGroupFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("DesktopGroup")

			policyFilterDataClientModel := PolicyFilterUuidDataClientModel{
				Uuid:   deliveryGroupFilter.DeliveryGroupId.ValueString(),
				Server: serverValue,
			}

			policyFilterDataJson, err := json.Marshal(policyFilterDataClientModel)
			if err != nil {
				diags.AddError(
					"Error adding Access Control Policy Filter to Policy Set. ",
					"An unexpected error occurred: "+err.Error(),
				)
				return filterRequests, err
			}

			filterRequest.SetFilterData(string(policyFilterDataJson))
			filterRequest.SetIsAllowed(deliveryGroupFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(deliveryGroupFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.DeliveryGroupTypeFilters.IsNull() && len(policy.DeliveryGroupTypeFilters.Elements()) > 0 {
		deliveryGroupTypeFilters := util.ObjectSetToTypedArray[DeliveryGroupTypeFilterModel](ctx, diags, policy.DeliveryGroupTypeFilters)
		for _, deliveryGroupTypeFilter := range deliveryGroupTypeFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("DesktopKind")

			filterRequest.SetFilterData(deliveryGroupTypeFilter.DeliveryGroupType.ValueString())
			filterRequest.SetIsAllowed(deliveryGroupTypeFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(deliveryGroupTypeFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.TagFilters.IsNull() && len(policy.TagFilters.Elements()) > 0 {
		tagFilters := util.ObjectSetToTypedArray[TagFilterModel](ctx, diags, policy.TagFilters)
		for _, tagFilter := range tagFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("DesktopTag")

			policyFilterDataClientModel := PolicyFilterUuidDataClientModel{
				Uuid:   tagFilter.Tag.ValueString(),
				Server: serverValue,
			}

			policyFilterDataJson, err := json.Marshal(policyFilterDataClientModel)
			if err != nil {
				diags.AddError(
					"Error adding Access Control Policy Filter to Policy Set. ",
					"An unexpected error occurred: "+err.Error(),
				)
				return filterRequests, err
			}

			filterRequest.SetFilterData(string(policyFilterDataJson))
			filterRequest.SetIsAllowed(tagFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(tagFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.OuFilters.IsNull() && len(policy.OuFilters.Elements()) > 0 {
		ouFilters := util.ObjectSetToTypedArray[OuFilterModel](ctx, diags, policy.OuFilters)
		for _, ouFilter := range ouFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("OU")

			filterRequest.SetFilterData(ouFilter.Ou.ValueString())
			filterRequest.SetIsAllowed(ouFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(ouFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.UserFilters.IsNull() && len(policy.UserFilters.Elements()) > 0 {
		userFilters := util.ObjectSetToTypedArray[UserFilterModel](ctx, diags, policy.UserFilters)
		for _, userFilter := range userFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType("User")

			filterRequest.SetFilterData(userFilter.UserSid.ValueString())
			filterRequest.SetIsAllowed(userFilter.Allowed.ValueBool())
			filterRequest.SetIsEnabled(userFilter.Enabled.ValueBool())
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	return filterRequests, nil
}

func constructPolicyPriorityRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, policySet *citrixorchestration.PolicySetResponse, planedPolicies []PolicyModel) citrixorchestration.ApiGpoRankGpoPoliciesRequest {
	// 1. Construct map of policy name: policy id
	// 2. Construct array of policy id based on the policy name order
	// 3. post policy priority
	policyNameIdMap := map[types.String]types.String{}
	if policySet.GetPolicies() != nil {
		for _, policy := range policySet.GetPolicies() {
			policyNameIdMap[types.StringValue(policy.GetPolicyName())] = types.StringValue(policy.GetPolicyGuid())
		}
	}
	policyPriority := []types.String{}
	for _, policyToCreate := range planedPolicies {
		policyPriority = append(policyPriority, policyNameIdMap[policyToCreate.Name])
	}

	policySetId := policySet.GetPolicySetGuid()
	createPolicyPriorityRequest := client.ApiClient.GpoDAAS.GpoRankGpoPolicies(ctx)
	createPolicyPriorityRequest = createPolicyPriorityRequest.PolicySetGuid(policySetId)
	createPolicyPriorityRequest = createPolicyPriorityRequest.RequestBody(util.ConvertBaseStringArrayToPrimitiveStringArray(policyPriority))
	return createPolicyPriorityRequest
}

func (r *policySetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicySetResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
