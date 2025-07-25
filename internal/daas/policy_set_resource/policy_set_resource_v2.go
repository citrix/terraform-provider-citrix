// Copyright © 2024. Citrix Systems, Inc.

package policy_set_resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &policySetV2Resource{}
	_ resource.ResourceWithConfigure      = &policySetV2Resource{}
	_ resource.ResourceWithImportState    = &policySetV2Resource{}
	_ resource.ResourceWithValidateConfig = &policySetV2Resource{}
	_ resource.ResourceWithModifyPlan     = &policySetV2Resource{}
)

func NewPolicySetV2Resource() resource.Resource {
	return &policySetV2Resource{}
}

type policySetV2Resource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *policySetV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set_v2"
}

// Schema defines the schema for the resource.
func (r *policySetV2Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicySetV2Model{}.GetSchema()
}

func (r *policySetV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySets, err := policies.GetPolicySets(ctx, r.client, &resp.Diagnostics)
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
	createPolicySetRequestBody.SetPolicySetType(string(citrixorchestration.SDKGPOPOLICYSETTYPE_DELIVERY_GROUP_POLICIES))

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
	policySetId := policySetResponse.GetPolicySetGuid()

	// Associated the created policy set with the delivery groups
	deliveryGroups := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.DeliveryGroups)
	err = policies.UpdateDeliveryGroupsWithPolicySet(ctx, &resp.Diagnostics, r.client, policySetResponse.GetName(), policySetId, deliveryGroups, fmt.Sprintf("associating Policy Set %s with Delivery Group", policySetResponse.GetName()))
	if err != nil {
		return
	}

	// Try getting the new policy set with policy set GUID
	policySet, policySetScopes, associatedDeliveryGroups, err := getPolicySetDetailsForRefreshState(ctx, &resp.Diagnostics, r.client, policySetId)
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// isDefaultPolicySet is set to `false` as the default policy set cannot be created
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, false, policySet, policySetScopes, associatedDeliveryGroups)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySetV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySetV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultPolicySetId, err := GetDefaultPolicySetId(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	policySet, policySetScopes, associatedDeliveryGroups, err := getPolicySetDetailsForRefreshState(ctx, &resp.Diagnostics, r.client, state.Id.ValueString())
	if err != nil {
		// Check if this is a "policy set not found" error
		if errors.Is(err, util.ErrPolicySetNotFound) {
			resp.Diagnostics.AddWarning(
				"Policy Set not found",
				fmt.Sprintf("Policy Set %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.Id.ValueString()),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	isDefaultPolicySet := strings.EqualFold(policySet.GetPolicySetGuid(), defaultPolicySetId)

	// Map response body to schema and populate Computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, isDefaultPolicySet, policySet, policySetScopes, associatedDeliveryGroups)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySetV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state PolicySetV2Model
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultPolicySetId, err := GetDefaultPolicySetId(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySetId := plan.Id.ValueString()
	policySetName := plan.Name.ValueString()

	// Validation for default policy set
	isDefaultPolicySet := strings.EqualFold(policySetId, defaultPolicySetId)
	if isDefaultPolicySet {
		err := validateDefaultPolicySetConfigs(ctx, &resp.Diagnostics, plan, state)
		if err != nil {
			return
		}
	}

	policySets, err := policies.GetPolicySets(ctx, r.client, &resp.Diagnostics)
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

	if !isDefaultPolicySet {
		err := updatePolicySetBody(ctx, &resp.Diagnostics, r.client, plan)
		if err != nil {
			return
		}
		err = updatePolicySetDeliveryGroups(ctx, &resp.Diagnostics, r.client, plan, state)
		if err != nil {
			return
		}
	}

	// Refresh values
	policySet, policySetScopes, associatedDeliveryGroups, err := getPolicySetDetailsForRefreshState(ctx, &resp.Diagnostics, r.client, policySetId)
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, isDefaultPolicySet, policySet, policySetScopes, associatedDeliveryGroups)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySetV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySetV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultPolicySetId, err := GetDefaultPolicySetId(ctx, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	if strings.EqualFold(state.Id.ValueString(), defaultPolicySetId) {
		resp.Diagnostics.AddError(
			"Error Deleting Policy Set",
			"Default Policy Set cannot be deleted",
		)
		return
	}

	policySet, err := policies.GetPolicySet(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil {
		return
	}

	// Move all the policies to the default policy set
	remainingPolicyIds := []string{}
	for _, policy := range policySet.GetPolicies() {
		remainingPolicyIds = append(remainingPolicyIds, policy.GetPolicyGuid())
	}
	remainingPolicyCount := len(remainingPolicyIds)
	if remainingPolicyCount > 0 {
		resp.Diagnostics.AddError(
			"Deleting Policy Set "+state.Id.ValueString(),
			fmt.Sprintf("Policy Set %s cannot be deleted. It still has %d policies associated with it. ", state.Id.ValueString(), remainingPolicyCount),
		)
		return
	}

	// Remove associations with delivery groups before attempting to delete the policy set
	err = updatePolicySetDeliveryGroups(ctx, &resp.Diagnostics, r.client, PolicySetV2Model{}, state)
	if err != nil {
		return
	}

	// Delete Policy Set
	deletePolicySetRequest := r.client.ApiClient.GpoDAAS.GpoDeleteGpoPolicySet(ctx, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deletePolicySetRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Policy Set "+state.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *policySetV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *policySetV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *policySetV2Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicySetV2Model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *policySetV2Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Validate DDC Version
	errorSummary := "Error managing Policy Set"
	feature := "Policy Set resource"
	isDdcVersionSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 120, 118, 7, 41, errorSummary, feature)
	if !isDdcVersionSupported {
		return
	}
}
