// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	if req.Plan.Raw.IsNull() || !req.Plan.Raw.IsFullyKnown() {
		return
	}

	create := req.State.Raw.IsNull()
	operation := "updating"
	if create {
		operation = "creating"
	}

	// Retrieve values from plan
	var plan PolicySetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate DDC Version
	errorSummary := fmt.Sprintf("Error %s Policy Set", operation)
	feature := "Policy Set resource"
	isDdcVersionSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 120, 118, 7, 41, errorSummary, feature)

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

	if !plan.DeliveryGroups.IsNull() {
		deliveryGroupsToCheck := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups)
		if len(deliveryGroupsToCheck) > 0 {
			deliveryGroups, err := util.GetDeliveryGroups(ctx, r.client, &resp.Diagnostics, "Id")
			if err != nil {
				return
			}
			for _, deliveryGroupToCheck := range deliveryGroupsToCheck {
				if !slices.ContainsFunc(deliveryGroups, func(deliveryGroup citrixorchestration.DeliveryGroupResponseModel) bool {
					return strings.EqualFold(deliveryGroup.GetId(), deliveryGroupToCheck)
				}) {
					resp.Diagnostics.AddError(
						"Error "+operation+" Policy Set",
						fmt.Sprintf("Specified Delivery Group with Id `%s` does not exist.", deliveryGroupToCheck),
					)
					return
				}
			}
		}
	}

	userSettings := []string{}
	settingDefinitions, err := getGpoUserSettingDefinitions(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}
	for _, setting := range settingDefinitions {
		if setting.GetIsUserSetting() {
			userSettings = append(userSettings, setting.GetSettingName())
		}
	}

	plannedPolicies := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)

	for _, policy := range plannedPolicies {
		policyContainsUserSetting := false
		existingPolicySettingNames := map[string]bool{}
		policySettings := util.ObjectSetToTypedArray[PolicySettingModel](ctx, &resp.Diagnostics, policy.PolicySettings)
		for _, setting := range policySettings {
			if existingPolicySettingNames[setting.Name.ValueString()] {
				resp.Diagnostics.AddError(
					"Error "+operation+" Policy Set",
					"Each type of policy settings can only be specified once in the same group policy.",
				)
				return
			} else {
				existingPolicySettingNames[setting.Name.ValueString()] = true
			}
			if slices.ContainsFunc(userSettings, func(userSetting string) bool {
				return strings.EqualFold(userSetting, setting.Name.ValueString())
			}) {
				policyContainsUserSetting = true
			}

			if strings.EqualFold(setting.Value.ValueString(), "true") ||
				strings.EqualFold(setting.Value.ValueString(), "1") ||
				strings.EqualFold(setting.Value.ValueString(), "false") ||
				strings.EqualFold(setting.Value.ValueString(), "0") {
				resp.Diagnostics.AddError(
					"Error "+operation+" Policy Set",
					"Please specify boolean policy setting value with the 'enabled' attribute.",
				)
				return
			}
		}

		if !policyContainsUserSetting {
			if !policy.AccessControlFilters.IsNull() ||
				!policy.BranchRepeaterFilter.IsNull() ||
				!policy.ClientIPFilters.IsNull() ||
				!policy.ClientNameFilters.IsNull() ||
				!policy.UserFilters.IsNull() {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Error configuring Policy %s in Policy Set %s", policy.Name.ValueString(), plan.Name.ValueString()),
					"None of `access_control_filters`, `branch_repeater_filter`, `client_ip_filters`, `client_name_filters`, and `user_filters` can be specified when policy does not contain any user setting.",
				)
				return
			}
		}

		if !policy.ClientPlatformFilters.IsNull() {
			clientPlatformFilters := util.ObjectSetToTypedArray[ClientPlatformFilterModel](ctx, &resp.Diagnostics, policy.ClientPlatformFilters)
			if len(clientPlatformFilters) > 0 {
				isDdcVersionSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 124, 124, 7, 44, errorSummary, "Policy Set Resource Client Platform Filters")
				if !isDdcVersionSupported {
					return
				}
			}
		}

		if !policy.DeliveryGroupFilters.IsNull() {
			deliveryGroupFilters := util.ObjectSetToTypedArray[DeliveryGroupFilterModel](ctx, &resp.Diagnostics, policy.DeliveryGroupFilters)
			for _, deliveryGroupFilter := range deliveryGroupFilters {
				deliveryGroupId := deliveryGroupFilter.DeliveryGroupId.ValueString()
				getDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupId)
				_, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, r.client)
				if err != nil {
					resp.Diagnostics.AddError(
						fmt.Sprintf("Delivery group %s specified in the delivery group filter does not exist.", deliveryGroupId),
						"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
							"\nError message: "+util.ReadClientError(err),
					)
					return
				}
			}
		}

		if !policy.TagFilters.IsNull() {
			tagFilters := util.ObjectSetToTypedArray[TagFilterModel](ctx, &resp.Diagnostics, policy.TagFilters)
			for _, tagFilter := range tagFilters {
				tagId := tagFilter.Tag.ValueString()
				getTagRequest := r.client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagId)
				_, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.TagDetailResponseModel](getTagRequest, r.client)
				if err != nil {
					resp.Diagnostics.AddError(
						fmt.Sprintf("Tag %s specified in the tag filter does not exist.", tagId),
						"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
							"\nError message: "+util.ReadClientError(err),
					)
					return
				}
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
	resp.Schema = PolicySetModel{}.GetSchema()
}

// Create implements resource.Resource.
func (r *policySetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySets, err := GetPolicySets(ctx, r.client, &resp.Diagnostics)
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

	defaultBoolSettingValueMap, err := GetGpoBooleanSettingDefaultValueMap(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}
	plannedPolicies := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
	// Create new policies
	batchRequestModel, err := constructCreatePolicyBatchRequestModel(ctx, &resp.Diagnostics, r.client, plannedPolicies, policySetResponse.GetPolicySetGuid(), policySetResponse.GetName(), defaultBoolSettingValueMap)
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
		return
	}

	if successfulJobs < len(plan.Policies.Elements()) {
		errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(plan.Policies.Elements()))
		resp.Diagnostics.AddError(
			"Error adding Policies to Policy Set "+policySetResponse.GetName(),
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
		return
	}

	// Associated the created policy set with the delivery groups
	deliveryGroups := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups)
	err = UpdateDeliveryGroupsWithPolicySet(ctx, &resp.Diagnostics, r.client, policySetResponse.GetName(), policySetResponse.GetPolicySetGuid(), deliveryGroups, fmt.Sprintf("associating Policy Set %s with Delivery Group", policySetResponse.GetName()))
	if err != nil {
		return
	}

	// Try getting the new policy set with policy set GUID
	policySet, err := GetPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
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
			return
		}
	}

	// Try getting the new policy set with policy set GUID
	policySet, err = GetPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	policies, err := GetPolicies(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, r.client, policySet.GetScopes())
	if err != nil {
		return
	}

	associatedDeliveryGroups, err := util.GetDeliveryGroups(ctx, r.client, &resp.Diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, policySet, policies, policySetScopes, associatedDeliveryGroups)

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

	var state PolicySetModel
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

	deliveryGroups, err := util.GetDeliveryGroups(ctx, r.client, &resp.Diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, true, policySet, policies, policySetScopes, deliveryGroups)

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
	var plan PolicySetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicySetModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySetId := plan.Id.ValueString()
	policySetName := plan.Name.ValueString()

	policySets, err := GetPolicySets(ctx, r.client, &resp.Diagnostics)
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

	policiesInPlan := util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, plan.Policies)
	policiesToCreate := []PolicyModel{}
	policiesToUpdate := []PolicyModel{}
	policiesWithUpdatedNames := []PolicyModel{}
	policyIdMapInPlan := map[string]PolicyModel{}
	for _, policy := range policiesInPlan {
		if policy.Id.ValueString() == "" {
			policiesToCreate = append(policiesToCreate, policy)
		} else {
			policiesToUpdate = append(policiesToUpdate, policy)
			policyIdMapInPlan[policy.Id.ValueString()] = policy
		}
	}

	policyIdsInState := []string{}
	policyIdMapFromState := map[string]PolicyModel{}
	for _, policy := range util.ObjectListToTypedArray[PolicyModel](ctx, &resp.Diagnostics, state.Policies) {
		policyIdMapFromState[strings.ToLower(policy.Id.ValueString())] = policy
		policyIdsInState = append(policyIdsInState, policy.Id.ValueString())
		if policyInPlan, ok := policyIdMapInPlan[policy.Id.ValueString()]; ok && policy.Name.ValueString() != policyInPlan.Name.ValueString() {
			policiesWithUpdatedNames = append(policiesWithUpdatedNames, policy)
		}
	}

	policyIdsToDelete := []string{}
	// Check if any policies are to be deleted
	for _, policyId := range policyIdsInState {
		if _, ok := policyIdMapInPlan[policyId]; !ok {
			policyIdsToDelete = append(policyIdsToDelete, policyId)
		}
	}

	// Rename policies to update with their policy id to avoid naming collision
	if len(policiesWithUpdatedNames) > 0 {
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating policies in policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not update policies within the policy set to be updated, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		var batchRequestModel citrixorchestration.BatchRequestModel
		for policyIndex, policy := range policiesWithUpdatedNames {
			var updatePolicyRequest = citrixorchestration.PolicyRequest{}
			updatePolicyRequest.SetName(policy.Id.ValueString())
			updatePolicyRequestBodyString, err := util.ConvertToString(updatePolicyRequest)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Policy "+policy.Name.ValueString()+" to Policy Set "+policySetName,
					"An unexpected error occurred: "+err.Error(),
				)
				return
			}

			relativeUrl := fmt.Sprintf("/gpo/policies/%s", policy.Id.ValueString())

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(fmt.Sprintf("renamePolicy%d", policyIndex))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItem.SetBody(updatePolicyRequestBodyString)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
		batchRequestModel.SetItems(batchRequestItems)
		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Policies in Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(batchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while updating policies in the Policy Set. %d of %d policies were updated to the Policy Set.", successfulJobs, len(batchRequestItems))
			resp.Diagnostics.AddError(
				"Error updating Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)
			return
		}
	}

	if len(policyIdsToDelete) > 0 {
		// Setup batch requests
		deletePolicyBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting policies from policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not delete policies from the policy set, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		// batch delete policies
		for index, policyId := range policyIdsToDelete {
			relativeUrl := fmt.Sprintf("/gpo/policies/%s", policyId)

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
				"Error deleting Policies from Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(deletePolicyBatchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while deleting policies from the Policy Set. %d of %d policies were deleted from the Policy Set.", successfulJobs, len(deletePolicyBatchRequestItems))
			resp.Diagnostics.AddError(
				"Error deleting Policies from Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return
		}
	}

	defaultBoolSettingValueMap, err := GetGpoBooleanSettingDefaultValueMap(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	if len(policiesToCreate) > 0 {
		// Create new policies
		batchRequestModel, err := constructCreatePolicyBatchRequestModel(ctx, &resp.Diagnostics, r.client, policiesToCreate, policySetId, policySetName, defaultBoolSettingValueMap)
		if err != nil {
			return
		}

		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(batchRequestModel.GetItems()) {
			errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(batchRequestModel.GetItems()))
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)
			return
		}
	}

	if len(policiesToUpdate) > 0 {
		// Update policies in policy set
		for _, policy := range policiesToUpdate {
			var editPolicyRequestModel = &citrixorchestration.PolicyBodyRequest{}
			editPolicyRequestModel.SetName(policy.Name.ValueString())
			editPolicyRequestModel.SetDescription(policy.Description.ValueString())
			editPolicyRequestModel.SetIsEnabled(policy.Enabled.ValueBool())

			editPolicyRequest := r.client.ApiClient.GpoDAAS.GpoUpdateGpoPolicy(ctx, policy.Id.ValueString())
			editPolicyRequest = editPolicyRequest.PolicyBodyRequest(*editPolicyRequestModel)

			// Update policy
			httpResp, err := citrixdaasclient.AddRequestData(editPolicyRequest, r.client).Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Policy "+policy.Name.ValueString(),
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}

			policyInState := policyIdMapFromState[strings.ToLower(policy.Id.ValueString())]
			// Perform policy setting updates
			policySettingsInPlan := util.ObjectSetToTypedArray[PolicySettingModel](ctx, &resp.Diagnostics, policy.PolicySettings)
			policySettingsInState := util.ObjectSetToTypedArray[PolicySettingModel](ctx, &resp.Diagnostics, policyInState.PolicySettings)
			err = updatePolicySettings(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), policySettingsInPlan, policySettingsInState, defaultBoolSettingValueMap)
			if err != nil {
				return
			}

			// Perform policy filter updates
			// Clear the policy filters
			err = clearPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString())
			if err != nil {
				return
			}

			serverValue := getServerValue(r.client)
			// Update Access Control Filters
			accessControlFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[AccessControlFilterModel](ctx, &resp.Diagnostics, policy.AccessControlFilters) {
				accessControlFilterInterfaceInPlan = append(accessControlFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, accessControlFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Branch Repeater Filter
			branchRepeaterFilterInterfaceInPlan := []PolicyFilterInterface{}
			if !policy.BranchRepeaterFilter.IsNull() {
				filter := util.ObjectValueToTypedObject[BranchRepeaterFilterModel](ctx, &resp.Diagnostics, policy.BranchRepeaterFilter)
				branchRepeaterFilterInterfaceInPlan = append(branchRepeaterFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, branchRepeaterFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Client IP Filters
			clientIpFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[ClientIPFilterModel](ctx, &resp.Diagnostics, policy.ClientIPFilters) {
				clientIpFilterInterfaceInPlan = append(clientIpFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, clientIpFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Client Name Filters
			clientNameFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[ClientNameFilterModel](ctx, &resp.Diagnostics, policy.ClientNameFilters) {
				clientNameFilterInterfaceInPlan = append(clientNameFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, clientNameFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Delivery Group Filters
			deliveryGroupFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[DeliveryGroupFilterModel](ctx, &resp.Diagnostics, policy.DeliveryGroupFilters) {
				deliveryGroupFilterInterfaceInPlan = append(deliveryGroupFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, deliveryGroupFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Delivery Group Type Filters
			deliveryGroupTypeFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[DeliveryGroupTypeFilterModel](ctx, &resp.Diagnostics, policy.DeliveryGroupTypeFilters) {
				deliveryGroupTypeFilterInterfaceInPlan = append(deliveryGroupTypeFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, deliveryGroupTypeFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Tag Filters
			tagFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[TagFilterModel](ctx, &resp.Diagnostics, policy.TagFilters) {
				tagFilterInterfaceInPlan = append(tagFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, tagFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update Ou Filters
			ouFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[OuFilterModel](ctx, &resp.Diagnostics, policy.OuFilters) {
				tagFilterInterfaceInPlan = append(tagFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, ouFilterInterfaceInPlan)
			if err != nil {
				return
			}

			// Update User Filters
			userFilterInterfaceInPlan := []PolicyFilterInterface{}
			for _, filter := range util.ObjectSetToTypedArray[UserFilterModel](ctx, &resp.Diagnostics, policy.UserFilters) {
				userFilterInterfaceInPlan = append(userFilterInterfaceInPlan, filter)
			}
			err = createPolicyFilters(ctx, r.client, &resp.Diagnostics, policy.Id.ValueString(), policy.Name.ValueString(), serverValue, userFilterInterfaceInPlan)
			if err != nil {
				return
			}
		}
	}

	// Update policy priority
	// Try getting the new policy set with policy set GUID
	policySet, err := GetPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	if len(policySet.Policies) > 0 {
		// Update Policy Priority
		policyPriorityRequest := constructPolicyPriorityRequest(ctx, r.client, policySet, policiesInPlan)
		// Update policy priorities in the Policy Set
		policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, r.client).Execute()
		if err != nil || !policyPriorityResponse {
			resp.Diagnostics.AddError(
				"Error Changing Policy Priorities in Policy Set "+policySet.GetPolicySetGuid(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
	}

	// Update delivery groups with policy set
	deliveryGroupsToBeRemoved := []string{}
	deliveryGroupsToBeAdded := []string{}
	deliveryGroupsInPlan := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups)
	deliveryGroupsInState := util.StringSetToStringArray(ctx, &resp.Diagnostics, state.DeliveryGroups)
	for _, deliveryGroupInState := range deliveryGroupsInState {
		if !slices.Contains(deliveryGroupsInPlan, deliveryGroupInState) {
			deliveryGroupsToBeRemoved = append(deliveryGroupsToBeRemoved, deliveryGroupInState)
		}
	}

	for _, deliveryGroupInPlan := range deliveryGroupsInPlan {
		if !slices.Contains(deliveryGroupsInState, deliveryGroupInPlan) {
			deliveryGroupsToBeAdded = append(deliveryGroupsToBeAdded, deliveryGroupInPlan)
		}
	}

	err = UpdateDeliveryGroupsWithPolicySet(ctx, &resp.Diagnostics, r.client, policySet.GetName(), util.DefaultSitePolicySetIdForDeliveryGroup, deliveryGroupsToBeRemoved, fmt.Sprintf("removing Policy Set %s's associations with Delivery Group", policySet.GetName()))
	if err != nil {
		return
	}

	err = UpdateDeliveryGroupsWithPolicySet(ctx, &resp.Diagnostics, r.client, policySet.GetName(), policySet.GetPolicySetGuid(), deliveryGroupsToBeAdded, fmt.Sprintf("associating Policy Set %s with Delivery Group", policySet.GetName()))
	if err != nil {
		return
	}

	policies, err := GetPolicies(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, r.client, policySet.GetScopes())
	if err != nil {
		return
	}

	associatedDeliveryGroups, err := util.GetDeliveryGroups(ctx, r.client, &resp.Diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, policySet, policies, policySetScopes, associatedDeliveryGroups)

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
	var state PolicySetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetId := state.Id.ValueString()
	policySetName := state.Name.ValueString()
	// Get delivery groups and check if the current policy set is assigned to one of them
	deliveryGroups, err := util.GetDeliveryGroups(ctx, r.client, &resp.Diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return
	}

	associatedDeliveryGroupIds := []string{}
	for _, deliveryGroup := range deliveryGroups {
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
		editDeliveryGroupRequestBody.SetPolicySetGuid(util.DefaultSitePolicySetIdForDeliveryGroup)
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
	httpResp, err := citrixdaasclient.AddRequestData(deletePolicySetRequest, r.client).Execute()
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

func (r *policySetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicySetModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
