// Copyright © 2024. Citrix Systems, Inc.

package policy_setting

import (
	"context"

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
	_ resource.Resource                   = &policySettingResource{}
	_ resource.ResourceWithConfigure      = &policySettingResource{}
	_ resource.ResourceWithImportState    = &policySettingResource{}
	_ resource.ResourceWithValidateConfig = &policySettingResource{}
	_ resource.ResourceWithModifyPlan     = &policySettingResource{}
)

func NewPolicySettingResource() resource.Resource {
	return &policySettingResource{}
}

type policySettingResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *policySettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_setting"
}

// Schema defines the schema for the resource.
func (r *policySettingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PolicySettingModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *policySettingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *policySettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingRequest, err := buildSettingRequest(ctx, r.client, &resp.Diagnostics, plan, "creating")
	createPolicySettingReq := r.client.ApiClient.GpoDAAS.GpoCreateGpoSetting(ctx)
	createPolicySettingReq = createPolicySettingReq.SettingRequest(settingRequest)
	createPolicySettingReq = createPolicySettingReq.PolicyGuid(plan.PolicyId.ValueString())
	policySetting, httpResp, err := citrixdaasclient.AddRequestData(createPolicySettingReq, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy Setting",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	policySetting, err = getPolicySetting(ctx, r.client, &resp.Diagnostics, policySetting.GetSettingGuid())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(policySetting)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySettingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetting, err := getPolicySetting(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(policySetting)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySettingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := updatePolicySetting(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	policySetting, err := getPolicySetting(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(policySetting)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *policySettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySettingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySettingId := state.Id.ValueString()
	deletePolicySettingReq := r.client.ApiClient.GpoDAAS.GpoDeleteGpoSetting(ctx, policySettingId)
	httpResp, err := citrixdaasclient.AddRequestData(deletePolicySettingReq, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Policy Setting "+policySettingId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *policySettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *policySettingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PolicySettingModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *policySettingResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func getPolicySetting(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySettingId string) (*citrixorchestration.SettingResponse, error) {
	getPolicySettingReq := client.ApiClient.GpoDAAS.GpoReadGpoSetting(ctx, policySettingId)
	policySetting, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.SettingResponse](getPolicySettingReq, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Setting "+policySettingId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policySetting, nil
}

func updatePolicySetting(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetting PolicySettingModel) error {
	settingRequest, err := buildSettingRequest(ctx, client, diagnostics, policySetting, "updating")
	editPolicySettingRequest := client.ApiClient.GpoDAAS.GpoUpdateGpoSetting(ctx, policySetting.Id.ValueString())
	editPolicySettingRequest = editPolicySettingRequest.SettingRequest(settingRequest)

	// Update policy setting
	httpResp, err := citrixdaasclient.AddRequestData(editPolicySettingRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Updating Policy Setting "+policySetting.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}
