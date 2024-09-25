// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &adminScopeResource{}
	_ resource.ResourceWithConfigure      = &adminScopeResource{}
	_ resource.ResourceWithImportState    = &adminScopeResource{}
	_ resource.ResourceWithValidateConfig = &adminScopeResource{}
	_ resource.ResourceWithModifyPlan     = &adminScopeResource{}
)

// NewAdminScopeResource is a helper function to simplify the provider implementation.
func NewAdminScopeResource() resource.Resource {
	return &adminScopeResource{}
}

// adminScopeResource is the resource implementation.
type adminScopeResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *adminScopeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_scope"
}

// Schema defines the schema for the resource.
func (r *adminScopeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AdminScopeResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *adminScopeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *adminScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateAdminScopeRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetIsTenantScope(plan.IsTenantScope.ValueBool())

	createAdminScopeRequest := r.client.ApiClient.AdminAPIsDAAS.AdminCreateAdminScope(ctx)
	createAdminScopeRequest = createAdminScopeRequest.CreateAdminScopeRequestModel(body)

	// Create new admin scope
	httpResp, err := citrixdaasclient.AddRequestData(createAdminScopeRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Admin Scope: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new admin scope with scope name
	adminScope, err := getAdminScope(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, adminScope)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *adminScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AdminScopeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the new admin scope with scope name
	adminScope, err := readAdminScope(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, adminScope)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *adminScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var adminScopeId = plan.Id.ValueString()
	var adminScopeName = plan.Name.ValueString()

	// Generate Update API request body from plan
	var body citrixorchestration.EditAdminScopeRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())

	// Update admin scope using orchestration call
	updateAdminScopeRequest := r.client.ApiClient.AdminAPIsDAAS.AdminUpdateAdminScope(ctx, adminScopeId)
	updateAdminScopeRequest = updateAdminScopeRequest.EditAdminScopeRequestModel(body)

	httpResp, err := citrixdaasclient.AddRequestData(updateAdminScopeRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Admin Scope: "+adminScopeName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated admin scope using orchestration.
	updatedAdminScope, err := getAdminScope(ctx, r.client, &resp.Diagnostics, adminScopeId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedAdminScope)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *adminScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AdminScopeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing admin scope
	adminScopeId := state.Id.ValueString()
	adminScopeName := state.Name.ValueString()
	deleteAdminScopeRequest := r.client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminScope(ctx, adminScopeId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteAdminScopeRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Admin Scope: "+adminScopeName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *adminScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getAdminScope(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, adminScopeName string) (*citrixorchestration.ScopeResponseModel, error) {
	getAdminScopeRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScope(ctx, adminScopeName)
	adminScope, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ScopeResponseModel](getAdminScopeRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Admin Scope: "+adminScopeName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return adminScope, err
}

func readAdminScope(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, adminScopeName string) (*citrixorchestration.ScopeResponseModel, error) {
	getAdminScopeRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScope(ctx, adminScopeName)
	adminScope, _, err := util.ReadResource[*citrixorchestration.ScopeResponseModel](getAdminScopeRequest, ctx, client, resp, "Admin Scope", adminScopeName)
	return adminScope, err
}

func (r *adminScopeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AdminScopeResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *adminScopeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		// No plan to modify. Return
		return
	}

	create := req.State.Raw.IsNull()

	var plan AdminScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	operation := "updating"
	if create {
		operation = "creating"
	}

	if plan.IsTenantScope.ValueBool() && r.client.ClientConfig.IsCspCustomer {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_tenant_scope"),
			"Error "+operation+" Admin Scope "+plan.Name.ValueString(),
			"Tenant scopes are created automatically when the tenants are onboarded to the Citrix Service Provider customer. Add a tenant to this customer to create the tenant scope.",
		)
		return
	}
}
