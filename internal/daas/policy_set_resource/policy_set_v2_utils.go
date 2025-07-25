// Copyright © 2024. Citrix Systems, Inc.

package policy_set_resource

import (
	"context"
	"fmt"
	"slices"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetDefaultPolicySetId(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient) (string, error) {
	policySets, err := policies.GetPolicySets(ctx, client, diagnostics)
	if err != nil {
		return "", err
	}

	for _, policySet := range policySets {
		if strings.EqualFold(policySet.GetName(), "DefaultSitePolicies") {
			return policySet.GetPolicySetGuid(), nil
		}
	}

	return "", fmt.Errorf("DefaultSitePolicies Policy Set not found")
}

func updatePolicySetBody(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan PolicySetV2Model) error {
	// Construct the update model
	var editPolicySetRequestBody = &citrixorchestration.PolicySetRequest{}
	editPolicySetRequestBody.SetName(plan.Name.ValueString())
	editPolicySetRequestBody.SetDescription(plan.Description.ValueString())
	editPolicySetRequestBody.SetScopes(util.StringSetToStringArray(ctx, diagnostics, plan.Scopes))

	editPolicySetRequest := client.ApiClient.GpoDAAS.GpoUpdateGpoPolicySet(ctx, plan.Id.ValueString())
	editPolicySetRequest = editPolicySetRequest.PolicySetRequest(*editPolicySetRequestBody)

	// Update Policy Set
	httpResp, err := citrixdaasclient.AddRequestData(editPolicySetRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Updating Policy Set",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}

func updatePolicySetDeliveryGroups(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan PolicySetV2Model, state PolicySetV2Model) error {
	// Update delivery groups with policy set
	deliveryGroupsToBeRemoved := []string{}
	deliveryGroupsToBeAdded := []string{}
	deliveryGroupsInPlan := util.StringSetToStringArray(ctx, diagnostics, plan.DeliveryGroups)
	deliveryGroupsInState := util.StringSetToStringArray(ctx, diagnostics, state.DeliveryGroups)
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

	policySetName := plan.Name.ValueString()
	err := policies.UpdateDeliveryGroupsWithPolicySet(ctx, diagnostics, client, policySetName, util.DefaultSitePolicySetIdForDeliveryGroup, deliveryGroupsToBeRemoved, fmt.Sprintf("removing Policy Set %s's associations with Delivery Group", policySetName))
	if err != nil {
		return err
	}

	err = policies.UpdateDeliveryGroupsWithPolicySet(ctx, diagnostics, client, policySetName, plan.Id.ValueString(), deliveryGroupsToBeAdded, fmt.Sprintf("associating Policy Set %s with Delivery Group", policySetName))
	if err != nil {
		return err
	}
	return nil
}

func getPolicySetDetailsForRefreshState(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetId string) (*citrixorchestration.PolicySetResponse, []string, []citrixorchestration.DeliveryGroupResponseModel, error) {
	policySet, err := policies.GetPolicySet(ctx, client, diagnostics, policySetId)
	if err != nil {
		// Check if this is a "policy set not found" error - return a specific error that can be handled by the caller
		if strings.Contains(err.Error(), "policy set not found") {
			return nil, nil, nil, fmt.Errorf("policy set not found: %w", err)
		}
		return nil, nil, nil, err
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, *diagnostics, client, policySet.GetScopes())
	if err != nil {
		return policySet, nil, nil, err
	}

	associatedDeliveryGroups, err := util.GetDeliveryGroups(ctx, client, diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return policySet, policySetScopes, nil, err
	}

	return policySet, policySetScopes, associatedDeliveryGroups, err
}

func validateDefaultPolicySetConfigs(ctx context.Context, diagnostics *diag.Diagnostics, plan PolicySetV2Model, state PolicySetV2Model) error {
	policySetId := plan.Id.ValueString()
	// Default policy set cannot change name/description/scopes
	if !strings.EqualFold(plan.Name.ValueString(), state.Name.ValueString()) ||
		plan.Description.ValueString() != "" {
		err := fmt.Errorf("Default Policy Set %s cannot be updated", policySetId)
		diagnostics.AddError(
			"Error Updating Policy Set "+policySetId,
			"Default Policy Set cannot be updated",
		)
		return err
	}

	deliveryGroups := util.StringSetToStringArray(ctx, diagnostics, plan.DeliveryGroups)
	if len(deliveryGroups) > 0 {
		err := fmt.Errorf("`delivery_groups` cannot be specified for the Default Policy Set")
		diagnostics.AddError(
			"Error Updating Policy Set "+policySetId,
			err.Error(),
		)
		return err
	}

	scopes := util.StringSetToStringArray(ctx, diagnostics, plan.Scopes)
	if len(scopes) > 0 {
		err := fmt.Errorf("`scopes` cannot be specified for the Default Policy Set")
		diagnostics.AddError(
			"Error Updating Policy Set "+policySetId,
			err.Error(),
		)
		return err
	}
	return nil
}
