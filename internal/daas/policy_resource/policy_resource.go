// Copyright © 2024. Citrix Systems, Inc.

package policy_resource

import (
	"context"
	"errors"
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
	_ resource.Resource                   = &policyResource{}
	_ resource.ResourceWithConfigure      = &policyResource{}
	_ resource.ResourceWithImportState    = &policyResource{}
	_ resource.ResourceWithValidateConfig = &policyResource{}
	_ resource.ResourceWithModifyPlan     = &policyResource{}
)

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

type policyResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the resource.
func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicyModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := createPolicy(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	policy, err = GetPolicy(ctx, r.client, &resp.Diagnostics, policy.GetPolicyGuid(), false, false)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policy)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := GetPolicy(ctx, r.client, &resp.Diagnostics, state.Id.ValueString(), false, false)
	if err != nil {
		// Check if this is a "policy not found" error
		if errors.Is(err, util.ErrPolicyNotFound) {
			resp.Diagnostics.AddWarning(
				"Policy not found",
				fmt.Sprintf("Policy %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.Id.ValueString()),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, policy)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicyModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if strings.EqualFold(state.Name.ValueString(), "Unfiltered") {
		resp.Diagnostics.AddError(
			"Error updating Policy "+state.Id.ValueString(),
			"The site default `Unfiltered` policy cannot be updated.",
		)
		return
	}

	err := updatePolicy(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	if !strings.EqualFold(plan.PolicySetId.ValueString(), state.PolicySetId.ValueString()) {
		err = movePoliciesToPolicySet(ctx, &resp.Diagnostics, r.client, plan.PolicySetId.ValueString(), plan.Id.ValueString())
		if err != nil {
			return
		}
	}

	policy, err := GetPolicy(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString(), false, false)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, policy)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if strings.EqualFold(state.Name.ValueString(), "Unfiltered") {
		resp.Diagnostics.AddError(
			"Error deleting Policy "+state.Id.ValueString(),
			"The site default `Unfiltered` policy cannot be deleted.",
		)
		return
	}

	policy, err := GetPolicy(ctx, r.client, &resp.Diagnostics, state.Id.ValueString(), true, true)
	if err != nil {
		return
	}

	remainingFilterCount := len(policy.GetFilters())
	if remainingFilterCount > 0 {
		resp.Diagnostics.AddError(
			"Error Deleting Policy "+state.Id.ValueString(),
			fmt.Sprintf("Policy has filters %d attached to it. Please delete the filters first.", remainingFilterCount),
		)
		return
	}

	remainingSettingCount := len(policy.GetSettings())
	if remainingSettingCount > 0 {
		resp.Diagnostics.AddError(
			"Error Deleting Policy "+state.Id.ValueString(),
			fmt.Sprintf("Policy has %d settings attached to it. Please delete the settings first.", remainingSettingCount),
		)
		return
	}

	policyId := state.Id.ValueString()
	deletePolicyReq := r.client.ApiClient.GpoDAAS.GpoDeleteGpoPolicy(ctx, policyId)
	httpResp, err := citrixdaasclient.AddRequestData(deletePolicyReq, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Policy "+policyId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *policyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicyModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *policyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func createPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyToCreate PolicyModel) (*citrixorchestration.PolicyResponse, error) {
	var createPolicyBody = citrixorchestration.PolicyRequest{}
	createPolicyBody.SetName(policyToCreate.Name.ValueString())
	createPolicyBody.SetDescription(policyToCreate.Description.ValueString())
	createPolicyBody.SetIsEnabled(policyToCreate.Enabled.ValueBool())

	createPolicyReq := client.ApiClient.GpoDAAS.GpoCreateGpoPolicy(ctx)
	createPolicyReq = createPolicyReq.PolicySetGuid(policyToCreate.PolicySetId.ValueString())
	createPolicyReq = createPolicyReq.PolicyRequest(createPolicyBody)
	policy, httpResp, err := citrixdaasclient.AddRequestData(createPolicyReq, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error creating Policy "+policyToCreate.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return policy, nil
}

func GetPolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyId string, withFilters bool, withSettings bool) (*citrixorchestration.PolicyResponse, error) {
	getPolicyReq := client.ApiClient.GpoDAAS.GpoReadGpoPolicy(ctx, policyId)
	getPolicyReq = getPolicyReq.WithFilters(withFilters)
	getPolicyReq = getPolicyReq.WithSettings(withSettings)
	policy, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.PolicyResponse](getPolicyReq, client)
	if err != nil {
		// Check if this is a 404 Not Found error - return a specific error that can be handled by the caller
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %v", util.ErrPolicyNotFound, err)
		}

		diagnostics.AddError(
			"Error reading Policy "+policyId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policy, nil
}

func updatePolicy(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyToUpdate PolicyModel) error {
	body := citrixorchestration.PolicyBodyRequest{}
	body.SetName(policyToUpdate.Name.ValueString())
	body.SetDescription(policyToUpdate.Description.ValueString())
	body.SetIsEnabled(policyToUpdate.Enabled.ValueBool())

	updatePolicyReq := client.ApiClient.GpoDAAS.GpoUpdateGpoPolicy(ctx, policyToUpdate.Id.ValueString())
	updatePolicyReq = updatePolicyReq.PolicyBodyRequest(body)
	httpResp, err := citrixdaasclient.AddRequestData(updatePolicyReq, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating Policy "+policyToUpdate.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}

func movePoliciesToPolicySet(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, policySetId string, policy string) error {
	movePoliciesRequest := client.ApiClient.GpoDAAS.GpoMoveGpoPolicies(ctx)
	movePoliciesRequest = movePoliciesRequest.RequestBody([]string{policy})
	movePoliciesRequest = movePoliciesRequest.ToPolicySet(policySetId)
	httpResp, err := citrixdaasclient.AddRequestData(movePoliciesRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error moving policies to Policy Set "+policySetId,
			fmt.Sprintf("Failed to move policy %s to Policy Set %s", policy, policySetId)+
				"\nTransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}
