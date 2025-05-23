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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Gets the policy set and logs any errors
func GetPolicySets(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.PolicySetResponse, error) {
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

func GetPolicySet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.PolicySetResponse, error) {
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
func GetPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.CollectionEnvelopeOfPolicyResponse, error) {
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

func getPolicySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string) (*citrixorchestration.CollectionEnvelopeOfSettingResponse, error) {
	getPolicySettingsRequest := client.ApiClient.GpoDAAS.GpoReadGpoSettings(ctx)
	getPolicySettingsRequest = getPolicySettingsRequest.PolicyGuid(policyId)
	settings, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfSettingResponse](getPolicySettingsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Settings in Policy "+policyId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return settings, err
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

func constructCreatePolicyBatchRequestModel(ctx context.Context, diags *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policiesToCreate []PolicyModel, policySetGuid string, policySetName string, defaultSettingValueMap map[string]string) (citrixorchestration.BatchRequestModel, error) {
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	var batchRequestModel citrixorchestration.BatchRequestModel

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diags.AddError(
			"Error creating policy in policy set "+policySetName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not create policies within the policy set, unexpected error: "+util.ReadClientError(err),
		)
		return batchRequestModel, err
	}

	for policyIndex, policyToCreate := range policiesToCreate {
		var createPolicyRequest = citrixorchestration.PolicyRequest{}
		createPolicyRequest.SetName(policyToCreate.Name.ValueString())
		createPolicyRequest.SetDescription(policyToCreate.Description.ValueString())
		createPolicyRequest.SetIsEnabled(policyToCreate.Enabled.ValueBool())
		// Add Policy Settings
		policySettings := []citrixorchestration.SettingRequest{}
		policySettingsToCreate := util.ObjectSetToTypedArray[PolicySettingModel](ctx, diags, policyToCreate.PolicySettings)
		for _, policySetting := range policySettingsToCreate {
			settingRequest, err := constructSettingRequest(policySetting, diags, policyToCreate.Name.ValueString(), defaultSettingValueMap)
			if err != nil {
				return batchRequestModel, err
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

func constructPolicyFilterRequests(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policy PolicyModel) ([]citrixorchestration.FilterRequest, error) {
	filterRequests := []citrixorchestration.FilterRequest{}

	serverValue := getServerValue(client)

	if !policy.AccessControlFilters.IsNull() && len(policy.AccessControlFilters.Elements()) > 0 {
		accessControlFilters := util.ObjectSetToTypedArray[AccessControlFilterModel](ctx, diagnostics, policy.AccessControlFilters)
		for _, accessControlFilter := range accessControlFilters {
			filterRequest, err := accessControlFilter.GetFilterRequest(diagnostics, serverValue)
			if err != nil {
				return filterRequests, err
			}
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.BranchRepeaterFilter.IsNull() {
		branchRepeaterFilter := util.ObjectValueToTypedObject[BranchRepeaterFilterModel](ctx, diagnostics, policy.BranchRepeaterFilter)
		branchRepeaterFilterRequest, _ := branchRepeaterFilter.GetFilterRequest(diagnostics, serverValue)
		filterRequests = append(filterRequests, branchRepeaterFilterRequest)
	}

	if !policy.ClientIPFilters.IsNull() && len(policy.ClientIPFilters.Elements()) > 0 {
		clientIpFilters := util.ObjectSetToTypedArray[ClientIPFilterModel](ctx, diagnostics, policy.ClientIPFilters)
		for _, clientIpFilter := range clientIpFilters {
			filterRequest, _ := clientIpFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.ClientNameFilters.IsNull() && len(policy.ClientNameFilters.Elements()) > 0 {
		clientNameFilters := util.ObjectSetToTypedArray[ClientNameFilterModel](ctx, diagnostics, policy.ClientNameFilters)
		for _, clientNameFilter := range clientNameFilters {
			filterRequest, _ := clientNameFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.ClientPlatformFilters.IsNull() && len(policy.ClientPlatformFilters.Elements()) > 0 {
		clientPlatformFilters := util.ObjectSetToTypedArray[ClientPlatformFilterModel](ctx, diagnostics, policy.ClientPlatformFilters)
		for _, clientPlatformFilter := range clientPlatformFilters {
			filterRequest, _ := clientPlatformFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.DeliveryGroupFilters.IsNull() && len(policy.DeliveryGroupFilters.Elements()) > 0 {
		deliveryGroupFilters := util.ObjectSetToTypedArray[DeliveryGroupFilterModel](ctx, diagnostics, policy.DeliveryGroupFilters)
		for _, deliveryGroupFilter := range deliveryGroupFilters {
			filterRequest, err := deliveryGroupFilter.GetFilterRequest(diagnostics, serverValue)
			if err != nil {
				return filterRequests, err
			}

			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.DeliveryGroupTypeFilters.IsNull() && len(policy.DeliveryGroupTypeFilters.Elements()) > 0 {
		deliveryGroupTypeFilters := util.ObjectSetToTypedArray[DeliveryGroupTypeFilterModel](ctx, diagnostics, policy.DeliveryGroupTypeFilters)
		for _, deliveryGroupTypeFilter := range deliveryGroupTypeFilters {
			filterRequest, _ := deliveryGroupTypeFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.TagFilters.IsNull() && len(policy.TagFilters.Elements()) > 0 {
		tagFilters := util.ObjectSetToTypedArray[TagFilterModel](ctx, diagnostics, policy.TagFilters)
		for _, tagFilter := range tagFilters {
			filterRequest, err := tagFilter.GetFilterRequest(diagnostics, serverValue)
			if err != nil {
				return filterRequests, err
			}
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.OuFilters.IsNull() && len(policy.OuFilters.Elements()) > 0 {
		ouFilters := util.ObjectSetToTypedArray[OuFilterModel](ctx, diagnostics, policy.OuFilters)
		for _, ouFilter := range ouFilters {
			filterRequest, _ := ouFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	if !policy.UserFilters.IsNull() && len(policy.UserFilters.Elements()) > 0 {
		userFilters := util.ObjectSetToTypedArray[UserFilterModel](ctx, diagnostics, policy.UserFilters)
		for _, userFilter := range userFilters {
			filterRequest, _ := userFilter.GetFilterRequest(diagnostics, serverValue)
			filterRequests = append(filterRequests, filterRequest)
		}
	}

	return filterRequests, nil
}

func constructPolicyPriorityRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, policySet *citrixorchestration.PolicySetResponse, planedPolicies []PolicyModel) citrixorchestration.ApiGpoRankGpoPoliciesRequest {
	// 1. Construct map of policy name: policy id
	// 2. Construct array of policy id based on the policy name order
	// 3. post policy priority
	policyNameIdMap := map[types.String]string{}
	if policySet.GetPolicies() != nil {
		for _, policy := range policySet.GetPolicies() {
			policyNameIdMap[types.StringValue(policy.GetPolicyName())] = policy.GetPolicyGuid()
		}
	}
	policyPriority := []string{}
	for _, policyToCreate := range planedPolicies {
		policyPriority = append(policyPriority, policyNameIdMap[policyToCreate.Name])
	}

	return ConstructPolicyPriorityRequestWithIds(ctx, client, policySet.GetPolicySetGuid(), policyPriority)
}

func ConstructPolicyPriorityRequestWithIds(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, policySetId string, policyPriority []string) citrixorchestration.ApiGpoRankGpoPoliciesRequest {
	createPolicyPriorityRequest := client.ApiClient.GpoDAAS.GpoRankGpoPolicies(ctx)
	createPolicyPriorityRequest = createPolicyPriorityRequest.PolicySetGuid(policySetId)
	createPolicyPriorityRequest = createPolicyPriorityRequest.RequestBody(policyPriority)
	return createPolicyPriorityRequest
}

func constructSettingRequest(policySetting PolicySettingModel, diagnostics *diag.Diagnostics, policyName string, defaultBoolSettingValueMap map[string]string) (citrixorchestration.SettingRequest, error) {
	settingRequest := citrixorchestration.SettingRequest{}
	settingName := policySetting.Name.ValueString()
	settingRequest.SetSettingName(settingName)
	settingRequest.SetUseDefault(policySetting.UseDefault.ValueBool())
	if policySetting.UseDefault.ValueBool() {
		if defaultBoolSettingValueMap[settingName] != "" {
			settingRequest.SetSettingValue(defaultBoolSettingValueMap[settingName])
		}
	} else if !policySetting.Value.IsNull() {
		settingRequest.SetSettingValue(policySetting.Value.ValueString())
	} else if !policySetting.Enabled.IsNull() {
		if policySetting.Enabled.ValueBool() {
			settingRequest.SetSettingValue("1")
		} else {
			settingRequest.SetSettingValue("0")
		}
	} else {
		err := fmt.Errorf("Policy setting %s has `use_default` set to `false`, but does not have `value` or `enabled` field specified", policySetting.Name.ValueString())
		diagnostics.AddError(
			fmt.Sprintf("Error adding policy setting %s to policy %s", settingName, policyName),
			err.Error(),
		)
		return settingRequest, err
	}
	return settingRequest, nil
}

func updatePolicySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policyName string, policySettingsInPlan []PolicySettingModel, policySettingsInState []PolicySettingModel, defaultBoolSettingValueMap map[string]string) error {
	// Detect deleted settings
	policySettingsToDelete := []PolicySettingModel{}
	for _, policySetting := range policySettingsInState {
		if !slices.ContainsFunc(policySettingsInPlan, func(policySettingInPlan PolicySettingModel) bool {
			return strings.EqualFold(policySetting.Name.ValueString(), policySettingInPlan.Name.ValueString())
		}) {
			policySettingsToDelete = append(policySettingsToDelete, policySetting)
		}
	}

	policySettingsToCreate := []PolicySettingModel{}
	policySettingsToUpdate := []PolicySettingModel{}
	for _, policySetting := range policySettingsInPlan {
		if !policySetting.UseDefault.ValueBool() && policySetting.Value.IsNull() && policySetting.Enabled.IsNull() {
			settingName := policySetting.Name.ValueString()
			err := fmt.Errorf("Policy setting %s has `use_default` set to `false`, but does not have `value` or `enabled` field specified", settingName)
			diagnostics.AddError(
				fmt.Sprintf("Error adding policy setting %s to policy %s", settingName, policyName),
				err.Error(),
			)
			return err
		}
		policyInStateIndex := slices.IndexFunc(policySettingsInState, func(policySettingInState PolicySettingModel) bool {
			return strings.EqualFold(policySetting.Name.ValueString(), policySettingInState.Name.ValueString())
		})
		if policyInStateIndex != -1 {
			policySettingInState := policySettingsInState[policyInStateIndex]
			if policySetting.Enabled.ValueBool() != policySettingInState.Enabled.ValueBool() ||
				policySetting.Value.ValueString() != policySettingInState.Value.ValueString() ||
				policySetting.UseDefault.ValueBool() != policySettingInState.UseDefault.ValueBool() {
				policySettingsToUpdate = append(policySettingsToUpdate, policySetting)
			}
		} else {
			policySettingsToCreate = append(policySettingsToCreate, policySetting)
		}
	}

	// Delete policy settings
	if len(policySettingsToDelete) > 0 {
		err := deletePolicySettings(ctx, client, diagnostics, policyId, policySettingsToDelete)
		if err != nil {
			return err
		}
	}

	// Create policy settings
	if len(policySettingsToCreate) > 0 {
		err := createPolicySettings(ctx, client, diagnostics, policyId, policyName, policySettingsToCreate, defaultBoolSettingValueMap)
		if err != nil {
			return err
		}
	}

	// Update policy settings
	if len(policySettingsToUpdate) > 0 {
		err := updatePolicySettingDetails(ctx, client, diagnostics, policyId, policyName, policySettingsToUpdate, defaultBoolSettingValueMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPolicySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policyName string, policySettingsToCreate []PolicySettingModel, defaultBoolSettingValueMap map[string]string) error {
	// Batch create new policy settings
	addPolicySettingBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diagnostics.AddError(
			"Error creating policy settings in policy "+policyName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not create policy settings in the policy, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	for index, policySetting := range policySettingsToCreate {
		relativeUrl := fmt.Sprintf("/gpo/settings?policyGuid=%s", policyId)

		settingRequest, err := constructSettingRequest(policySetting, diagnostics, policyName, defaultBoolSettingValueMap)
		if err != nil {
			return err
		}
		settingRequestStringBody, err := util.ConvertToString(settingRequest)
		if err != nil {
			diagnostics.AddError(
				"Error adding policy setting to policy "+policyName,
				"An unexpected error occurred: "+err.Error(),
			)
			return err
		}

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("addPolicySetting%s", strconv.Itoa(index)))
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(settingRequestStringBody)
		addPolicySettingBatchRequestItems = append(addPolicySettingBatchRequestItems, batchRequestItem)
	}

	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(addPolicySettingBatchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		diagnostics.AddError(
			"Error adding Policy Settings to Policy "+policyName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(addPolicySettingBatchRequestItems) {
		errMsg := fmt.Sprintf("An error occurred while adding policy settings to the Policy. %d of %d policy settings were added to the Policy.", successfulJobs, len(addPolicySettingBatchRequestItems))
		diagnostics.AddError(
			"Error adding policy settings to Policy "+policyName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
		return err
	}
	return nil
}

func updatePolicySettingDetails(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policyName string, policySettingsToUpdate []PolicySettingModel, defaultBoolSettingValueMap map[string]string) error {
	// Batch create new policy settings
	updatePolicySettingBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diagnostics.AddError(
			"Error updating policy settings in policy "+policyName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not update policy settings in the policy, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	policySettings, err := getPolicySettings(ctx, client, diagnostics, policyId)
	if err != nil {
		return err
	}

	policySettingIdMap := map[string]PolicySettingModel{}
	for _, policySettingInPlan := range policySettingsToUpdate {
		for _, policySettingInRemote := range policySettings.GetItems() {
			if strings.EqualFold(policySettingInPlan.Name.ValueString(), policySettingInRemote.GetSettingName()) {
				policySettingIdMap[policySettingInRemote.GetSettingGuid()] = policySettingInPlan
				continue
			}
		}
	}

	// Update policy settings
	policySettingUpdateCounter := 0
	for id, policySetting := range policySettingIdMap {
		relativeUrl := fmt.Sprintf("/gpo/settings/%s", id)

		settingRequest, err := constructSettingRequest(policySetting, diagnostics, policyName, defaultBoolSettingValueMap)
		if err != nil {
			return err
		}
		settingRequestStringBody, err := util.ConvertToString(settingRequest)
		if err != nil {
			diagnostics.AddError(
				"Error updating policy setting in policy "+policyName,
				"An unexpected error occurred: "+err.Error(),
			)
			return err
		}

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("updatePolicySetting%s", strconv.Itoa(policySettingUpdateCounter)))
		batchRequestItem.SetMethod(http.MethodPatch)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(settingRequestStringBody)
		updatePolicySettingBatchRequestItems = append(updatePolicySettingBatchRequestItems, batchRequestItem)
		policySettingUpdateCounter++
	}

	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(updatePolicySettingBatchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		diagnostics.AddError(
			"Error updating Policy Settings in Policy "+policyName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(updatePolicySettingBatchRequestItems) {
		errMsg := fmt.Sprintf("An error occurred while updating policy settings to the Policy. %d of %d policy settings were updated.", successfulJobs, len(updatePolicySettingBatchRequestItems))
		diagnostics.AddError(
			"Error updating policy settings in Policy "+policyName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
		return err
	}
	return nil
}

func deletePolicySettings(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policySettingsToDelete []PolicySettingModel) error {
	// Setup batch requests
	deletePolicySettingBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diagnostics.AddError(
			"Error deleting policy settings from policy "+policyId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not delete policy settings from the policy, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	policySettings, err := getPolicySettings(ctx, client, diagnostics, policyId)
	if err != nil {
		return err
	}

	policySettingIdsToDelete := []string{}
	for _, policySetting := range policySettings.GetItems() {
		if slices.ContainsFunc(policySettingsToDelete, func(policySettingToDelete PolicySettingModel) bool {
			return strings.EqualFold(policySetting.GetSettingName(), policySettingToDelete.Name.ValueString())
		}) {
			policySettingIdsToDelete = append(policySettingIdsToDelete, policySetting.GetSettingGuid())
		}
	}

	// batch delete policy settings
	for index, policySettingId := range policySettingIdsToDelete {
		relativeUrl := fmt.Sprintf("/gpo/settings/%s", policySettingId)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("removeSetting%s", strconv.Itoa(index)))
		batchRequestItem.SetMethod(http.MethodDelete)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		deletePolicySettingBatchRequestItems = append(deletePolicySettingBatchRequestItems, batchRequestItem)
	}

	var deletePolicySettingBatchRequestModel citrixorchestration.BatchRequestModel
	deletePolicySettingBatchRequestModel.SetItems(deletePolicySettingBatchRequestItems)

	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, deletePolicySettingBatchRequestModel)
	if err != nil {
		diagnostics.AddError(
			"Error deleting policy settings from Policy "+policyId,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(deletePolicySettingBatchRequestItems) {
		errMsg := fmt.Sprintf("An error occurred while deleting policy settings from the Policy. %d of %d policy settings were deleted from the Policy.", successfulJobs, len(deletePolicySettingBatchRequestItems))
		diagnostics.AddError(
			"Error deleting policy settings from Policy "+policyId,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}
	return nil
}

func clearPolicyFilters(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policyName string) error {
	readFiltersRequest := client.ApiClient.GpoDAAS.GpoReadGpoFilters(ctx)
	readFiltersRequest = readFiltersRequest.PolicyGuid(policyId)

	filtersResponse, httpResp, err := citrixdaasclient.AddRequestData(readFiltersRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating Policy Filters for Policy "+policyId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	if filtersResponse != nil {
		policyFilterIdsToDelete := []string{}
		for _, policyFilter := range filtersResponse.GetItems() {
			policyFilterIdsToDelete = append(policyFilterIdsToDelete, policyFilter.GetFilterGuid())
		}
		err := deletePolicyFilters(ctx, client, diagnostics, policyName, policyFilterIdsToDelete)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPolicyFilters(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, policyName string, serverValue string, policyFiltersToCreate []PolicyFilterInterface) error {
	// Batch create new policy filters
	addPolicyFiltersBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diagnostics.AddError(
			"Error creating policy filters in policy "+policyName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not create policy filters in the policy, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	for index, policyFilter := range policyFiltersToCreate {
		relativeUrl := fmt.Sprintf("/gpo/filters?policyGuid=%s", policyId)

		filterRequest, err := policyFilter.GetFilterRequest(diagnostics, serverValue)
		if err != nil {
			return err
		}
		filterRequestStringBody, err := util.ConvertToString(filterRequest)
		if err != nil {
			diagnostics.AddError(
				"Error adding policy filter to policy "+policyName,
				"An unexpected error occurred: "+err.Error(),
			)
			return err
		}

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("addPolicyFilter%s", strconv.Itoa(index)))
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(filterRequestStringBody)
		addPolicyFiltersBatchRequestItems = append(addPolicyFiltersBatchRequestItems, batchRequestItem)
	}

	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(addPolicyFiltersBatchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		diagnostics.AddError(
			"Error adding Policy Filters to Policy "+policyName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(addPolicyFiltersBatchRequestItems) {
		errMsg := fmt.Sprintf("An error occurred while adding policy filters to the Policy. %d of %d policy filters were added to the Policy.", successfulJobs, len(addPolicyFiltersBatchRequestItems))
		diagnostics.AddError(
			"Error adding policy filters to Policy "+policyName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
		return err
	}
	return nil
}

func deletePolicyFilters(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyName string, policyFilterIdsToDelete []string) error {
	if len(policyFilterIdsToDelete) == 0 {
		return nil
	}

	// Setup batch requests
	deletePolicyFilterBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diagnostics.AddError(
			"Error deleting policy filters from policy "+policyName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not delete policy filters from the policy, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	// batch delete policy filters
	for index, policyFilterId := range policyFilterIdsToDelete {
		relativeUrl := fmt.Sprintf("/gpo/filters/%s", policyFilterId)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("removeFilter%s", strconv.Itoa(index)))
		batchRequestItem.SetMethod(http.MethodDelete)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		deletePolicyFilterBatchRequestItems = append(deletePolicyFilterBatchRequestItems, batchRequestItem)
	}

	var deletePolicyFilterBatchRequestModel citrixorchestration.BatchRequestModel
	deletePolicyFilterBatchRequestModel.SetItems(deletePolicyFilterBatchRequestItems)

	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, deletePolicyFilterBatchRequestModel)
	if err != nil {
		diagnostics.AddError(
			"Error deleting policy filters from Policy "+policyName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(deletePolicyFilterBatchRequestItems) {
		errMsg := fmt.Sprintf("An error occurred while deleting policy filters from the Policy. %d of %d policy filters were deleted from the Policy.", successfulJobs, len(deletePolicyFilterBatchRequestItems))
		diagnostics.AddError(
			"Error deleting policy filters from Policy "+policyName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}
	return nil
}

func constructEditDeliveryGroupPolicySetBatchRequestModel(diags *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetGuid string, deliveryGroupIds []string) (citrixorchestration.BatchRequestModel, error) {
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	var batchRequestModel citrixorchestration.BatchRequestModel

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error associated policy set %s to delivery groups ", policySetGuid),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError Message: "+util.ReadClientError(err),
		)
		return batchRequestModel, err
	}

	for index, deliveryGroupId := range deliveryGroupIds {
		var editDeliveryGroupRequest = citrixorchestration.EditDeliveryGroupRequestModel{}
		editDeliveryGroupRequest.SetPolicySetGuid(policySetGuid)

		editDeliveryGroupRequestBodyString, err := util.ConvertToString(editDeliveryGroupRequest)
		if err != nil {
			diags.AddError(
				"Error associate delivery group "+deliveryGroupId+" with Policy Set "+policySetGuid,
				"An unexpected error occurred: "+err.Error(),
			)
			return batchRequestModel, err
		}

		relativeUrl := fmt.Sprintf("/DeliveryGroups/%s", deliveryGroupId)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("editDeliveryGroup%d", index))
		batchRequestItem.SetMethod(http.MethodPatch)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(editDeliveryGroupRequestBodyString)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	batchRequestModel.SetItems(batchRequestItems)
	return batchRequestModel, nil
}

func UpdateDeliveryGroupsWithPolicySet(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetName string, policySetGuid string, deliveryGroups []string, errorMessage string) error {
	if len(deliveryGroups) == 0 {
		return nil
	}
	// Update Delivery Groups in the Policy Set
	batchRequestModel, err := constructEditDeliveryGroupPolicySetBatchRequestModel(diagnostics, client, policySetGuid, deliveryGroups)
	if err != nil {
		return err
	}

	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error %s.", errorMessage),
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(deliveryGroups) {
		errMsg := fmt.Sprintf("An error occurred while %s. %d of %d delivery groups were updated.", errorMessage, successfulJobs, len(deliveryGroups))
		diagnostics.AddError(
			fmt.Sprintf("Error %s.", errorMessage),
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
		return err
	}
	return nil
}

func getGpoUserSettingDefinitions(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient) ([]citrixorchestration.SettingDefinition, error) {
	getSettingDefinitionsRequest := client.ApiClient.GpoDAAS.GpoGetSettingDefinitions(ctx)
	getSettingDefinitionsRequest = getSettingDefinitionsRequest.IsLean(true)
	getSettingDefinitionsRequest = getSettingDefinitionsRequest.Limit(-1)
	getSettingDefinitionsRequest = getSettingDefinitionsRequest.IsUserSetting(true)
	getSettingDefinitionsRequest.Execute()
	settingResp, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.SettingDefinitionEnvelope](getSettingDefinitionsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Unable to fetch user setting definitions",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return []citrixorchestration.SettingDefinition{}, err
	}
	return settingResp.GetItems(), nil
}

func GetGpoBooleanSettingDefaultValueMap(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient) (map[string]string, error) {
	defaultValueMap := map[string]string{}
	getSettingDefinitionsRequest := client.ApiClient.GpoDAAS.GpoGetSettingDefinitions(ctx)
	getSettingDefinitionsRequest = getSettingDefinitionsRequest.IsLean(true)
	getSettingDefinitionsRequest = getSettingDefinitionsRequest.Limit(-1)
	getSettingDefinitionsRequest.Execute()
	settingResp, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.SettingDefinitionEnvelope](getSettingDefinitionsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Unable to fetch boolean setting definitions",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return defaultValueMap, err
	}
	for _, setting := range settingResp.GetItems() {
		if setting.GetValueType() == "State" || setting.GetValueType() == "StateAllowed" {
			defaultValueMap[setting.GetSettingName()] = setting.GetDefaultValue()
		}
	}

	return defaultValueMap, nil
}

func getServerValue(client *citrixdaasclient.CitrixDaasClient) string {
	if client.AuthConfig.OnPremises || !client.AuthConfig.ApiGateway {
		return client.ApiClient.GetConfig().Host
	} else {
		switch client.AuthConfig.Environment {
		case "Staging":
			return fmt.Sprintf("%s.xdstaging.net", client.ClientConfig.CustomerId)
		case "Japan":
			return fmt.Sprintf("%s.xendesktop.jp", client.ClientConfig.CustomerId)
		case "JapanStaging":
			return fmt.Sprintf("%s.xdstaging.jp", client.ClientConfig.CustomerId)
		case "Gov":
			return fmt.Sprintf("%s.xendesktop.us", client.ClientConfig.CustomerId)
		case "GovStaging":
			return fmt.Sprintf("%s.xdstaging.us", client.ClientConfig.CustomerId)
		default:
			return fmt.Sprintf("%s.xendesktop.net", client.ClientConfig.CustomerId)
		}
	}
}
