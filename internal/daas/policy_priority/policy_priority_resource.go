// Copyright © 2024. Citrix Systems, Inc.

package policy_priority

import (
	"context"
	"fmt"
	"slices"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &policyPriorityResource{}
	_ resource.ResourceWithConfigure      = &policyPriorityResource{}
	_ resource.ResourceWithImportState    = &policyPriorityResource{}
	_ resource.ResourceWithValidateConfig = &policyPriorityResource{}
	_ resource.ResourceWithModifyPlan     = &policyPriorityResource{}
)

func NewPolicyPriorityResource() resource.Resource {
	return &policyPriorityResource{}
}

type policyPriorityResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *policyPriorityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_priority"
}

// Schema defines the schema for the resource.
func (r *policyPriorityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicyPriorityModel{}.GetSchema()
}

func (r *policyPriorityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicyPriorityModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySet, err := policies.GetPolicySet(ctx, r.client, &resp.Diagnostics, plan.PolicySetId.ValueString())
	if err != nil {
		return
	}

	err = validateAndUpdatePolicyPriorities(ctx, &resp.Diagnostics, r.client, policySet.GetPolicySetGuid(), plan)
	if err != nil {
		return
	}

	policiesInRemote, err := policies.GetPolicies(ctx, r.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policiesInRemote)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyPriorityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicyPriorityModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySet, err := policies.GetPolicySet(ctx, r.client, &resp.Diagnostics, state.PolicySetId.ValueString())
	if err != nil {
		return
	}

	policiesInPolicySet, err := policies.GetPolicies(ctx, r.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policiesInPolicySet)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyPriorityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicyPriorityModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySet, err := policies.GetPolicySet(ctx, r.client, &resp.Diagnostics, plan.PolicySetId.ValueString())
	if err != nil {
		return
	}

	err = validateAndUpdatePolicyPriorities(ctx, &resp.Diagnostics, r.client, policySet.GetPolicySetGuid(), plan)
	if err != nil {
		return
	}

	policiesInRemote, err := policies.GetPolicies(ctx, r.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policiesInRemote)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyPriorityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicyPriorityModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *policyPriorityResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *policyPriorityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("policy_set_id"), req, resp)
}

func (r *policyPriorityResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicyPriorityModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *policyPriorityResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func updatePolicySetPolicyPriorities(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetId string, policyPriority []string) error {
	if len(policyPriority) > 0 {
		// Update Policy Priority
		policyPriorityRequest := policies.ConstructPolicyPriorityRequestWithIds(ctx, client, policySetId, policyPriority)
		// Update policy priorities in the Policy Set
		policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, client).Execute()
		if err != nil || !policyPriorityResponse {
			diagnostics.AddError(
				"Error Changing Policy Priorities in Policy Set "+policySetId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return err
		}
	}
	return nil
}

func validateAndUpdatePolicyPriorities(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetId string, plan PolicyPriorityModel) error {
	policiesInRemote, err := policies.GetPolicies(ctx, client, diagnostics, policySetId)
	if err != nil {
		return err
	}
	policyIdsInRemote := []string{}
	for _, policy := range policiesInRemote.GetItems() {
		policyIdsInRemote = append(policyIdsInRemote, policy.GetPolicyGuid())
	}

	// If not containing policies specified, throw error
	policyPriority := util.StringListToStringArray(ctx, diagnostics, plan.PolicyPriority)
	policiesNotInRemote := []string{}
	policiesNotInPlan := []string{}
	for _, policyId := range policyPriority {
		if !slices.ContainsFunc(policyIdsInRemote, func(policyIdInRemote string) bool {
			return strings.EqualFold(policyIdInRemote, policyId)
		}) {
			policiesNotInRemote = append(policiesNotInRemote, policyId)
		}
	}

	if len(policiesNotInRemote) > 0 {
		err := fmt.Errorf("policy IDs %s in the `policy_priority` list are not found in the policy set %s", strings.Join(policiesNotInRemote, ", "), policySetId)
		diagnostics.AddError(
			"Error managing Policy Priority in Policy Set "+policySetId,
			err.Error(),
		)
		return err
	}

	for _, policyId := range policyIdsInRemote {
		if !slices.ContainsFunc(policyPriority, func(policyPriorityId string) bool {
			return strings.EqualFold(policyPriorityId, policyId)
		}) {
			policiesNotInPlan = append(policiesNotInPlan, policyId)
		}
	}
	if len(policiesNotInPlan) > 0 {
		err := fmt.Errorf("policy IDs [%s] in the policy set are not found in the `policy_priority` list", strings.Join(policiesNotInPlan, ", "))
		diagnostics.AddError(
			"Error managing Policy Priority in Policy Set "+policySetId,
			err.Error(),
		)
		return err
	}

	err = updatePolicySetPolicyPriorities(ctx, diagnostics, client, policySetId, policyPriority)
	if err != nil {
		return err
	}

	return nil
}
