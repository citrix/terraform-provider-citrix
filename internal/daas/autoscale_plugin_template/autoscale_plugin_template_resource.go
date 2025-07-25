// Copyright © 2024. Citrix Systems, Inc.

package autoscale_plugin_template

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
	_ resource.Resource                   = &autoscalePluginTemplateResource{}
	_ resource.ResourceWithConfigure      = &autoscalePluginTemplateResource{}
	_ resource.ResourceWithImportState    = &autoscalePluginTemplateResource{}
	_ resource.ResourceWithValidateConfig = &autoscalePluginTemplateResource{}
	_ resource.ResourceWithModifyPlan     = &autoscalePluginTemplateResource{}
)

// NewautoscalePluginTemplateResource is a helper function to simplify the provider implementation.
func NewAutoscalePluginTemplateResource() resource.Resource {
	return &autoscalePluginTemplateResource{}
}

// autoscalePluginTemplateResource is the resource implementation.
type autoscalePluginTemplateResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Configure implements resource.ResourceWithConfigure.
func (r *autoscalePluginTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Metadata implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_autoscale_plugin_template"
}

// Schema implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AutoscalePluginTemplateResourceModel{}.GetSchema()
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (r *autoscalePluginTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AutoscalePluginTemplateResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *autoscalePluginTemplateResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	isFeatureSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 125, 124, 7, 44, "Error managing Autoscale Plugin Template resource", "Autoscale Plugin Template resource")

	if !isFeatureSupported {
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan AutoscalePluginTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if req.State.Raw.IsNull() {
		// For create, we want to check if the autoscale plugin template already exists
		autoscalePluginTemplate, _ := getAutoscalePluginTemplate(ctx, r.client, nil, plan.Type.ValueString(), plan.Name.ValueString())
		if autoscalePluginTemplate != nil {
			resp.Diagnostics.AddError(
				"Error creating Autoscale Plugin Template",
				"Autoscale Plugin Template with name '"+plan.Name.ValueString()+"' already exists.",
			)
		}

		return
	}

	var state AutoscalePluginTemplateResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !strings.EqualFold(state.Name.ValueString(), plan.Name.ValueString()) {
		// If the name is changed, we need to check if the new name already exists
		autoscalePluginTemplate, _ := getAutoscalePluginTemplate(ctx, r.client, nil, plan.Type.ValueString(), plan.Name.ValueString())
		if autoscalePluginTemplate != nil {
			resp.Diagnostics.AddError(
				"Error updating Autoscale Plugin Template",
				"Autoscale Plugin Template with name '"+plan.Name.ValueString()+"' already exists.",
			)
		}
	}
}

// Create implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan AutoscalePluginTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var autoscalePluginTemplateRequestModel citrixorchestration.CreateAutoscalePluginTemplateRequestModel
	autoscalePluginTemplateRequestModel.SetName(plan.Name.ValueString())
	autoscalePluginTemplateRequestModel.SetDates(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Dates))

	autoscalePluginTemplateRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsCreateAutoscalePluginTemplate(ctx, plan.Type.ValueString())
	autoscalePluginTemplateRequest = autoscalePluginTemplateRequest.CreateAutoscalePluginTemplateRequestModel(autoscalePluginTemplateRequestModel)
	httpResp, err := citrixdaasclient.AddRequestData(autoscalePluginTemplateRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Autoscale Plugin Template",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	autoscalePlugintemplate, err := getAutoscalePluginTemplate(ctx, r.client, &resp.Diagnostics, plan.Type.ValueString(), plan.Name.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, autoscalePlugintemplate)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AutoscalePluginTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed autoscalePluginTemplate properties from Orchestration
	autoscalePluginTemplate, err := readAutoscalePluginTemplate(ctx, r.client, resp, state.Type.ValueString(), state.Name.ValueString())
	if err != nil {
		return
	}

	// Overwrite autoscalePluginTemplate with refreshed state
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, autoscalePluginTemplate)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state AutoscalePluginTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	autoscalePluginTemplateName := state.Name.ValueString()
	deleteAutoscalePluginTemplateRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDeleteAutoscalePluginTemplate(ctx, state.Type.ValueString(), autoscalePluginTemplateName)
	httpResp, err := citrixdaasclient.AddRequestData(deleteAutoscalePluginTemplateRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Autoscale Plugin Template "+autoscalePluginTemplateName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: templateType,templateName. Got: %q", req.ID),
		)
		return
	}

	if idParts[0] != string(citrixorchestration.AUTOSCALEPLUGINTYPE_HOLIDAY) {
		resp.Diagnostics.AddError(
			"Invalid Template Type",
			fmt.Sprintf("Expected template type: %q. Got: %q", string(citrixorchestration.AUTOSCALEPLUGINTYPE_HOLIDAY), idParts[0]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
}

// Update implements resource.ResourceWithImportState.
func (r *autoscalePluginTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan AutoscalePluginTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AutoscalePluginTemplateResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var updateAutoscalePluginTemplateRequestModel citrixorchestration.UpdateAutoscalePluginTemplateRequestModel
	updateAutoscalePluginTemplateRequestModel.SetName(plan.Name.ValueString())
	updateAutoscalePluginTemplateRequestModel.SetDates(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Dates))

	updateAutoscalePluginTemplateRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsUpdateAutoscalePluginTemplate(ctx, state.Type.ValueString(), state.Name.ValueString())
	updateAutoscalePluginTemplateRequest = updateAutoscalePluginTemplateRequest.UpdateAutoscalePluginTemplateRequestModel(updateAutoscalePluginTemplateRequestModel)
	httpResp, err := citrixdaasclient.AddRequestData(updateAutoscalePluginTemplateRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Autoscale Plugin Template",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	autoscalePluginTemplate, err := getAutoscalePluginTemplate(ctx, r.client, &resp.Diagnostics, plan.Type.ValueString(), plan.Name.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, autoscalePluginTemplate)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func getAutoscalePluginTemplate(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, templateType string, autoscalePluginTemplateName string) (*citrixorchestration.AutoscalePluginTemplateResponseModel, error) {
	// Resolve resource path for service offering and master image
	getAutoscalePluginTemplateReq := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetAutoscalePluginTemplate(ctx, templateType, autoscalePluginTemplateName)
	autoscalePluginTemplate, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.AutoscalePluginTemplateResponseModel](getAutoscalePluginTemplateReq, client)
	if err != nil && diagnostics != nil {
		diagnostics.AddError(
			"Error reading Autoscale Plugin Template "+autoscalePluginTemplateName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return autoscalePluginTemplate, err
}

func readAutoscalePluginTemplate(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, templateType string, autoscalePluginTemplateName string) (*citrixorchestration.AutoscalePluginTemplateResponseModel, error) {
	getAutoscalePluginTemplateReq := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetAutoscalePluginTemplate(ctx, templateType, autoscalePluginTemplateName)
	autoscalePluginTemplate, _, err := util.ReadResource[*citrixorchestration.AutoscalePluginTemplateResponseModel](getAutoscalePluginTemplateReq, ctx, client, resp, "Autoscale Plugin Template", autoscalePluginTemplateName)
	return autoscalePluginTemplate, err
}
