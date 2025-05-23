// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &deliveryGroupTypeFilterResource{}
	_ resource.ResourceWithConfigure      = &deliveryGroupTypeFilterResource{}
	_ resource.ResourceWithImportState    = &deliveryGroupTypeFilterResource{}
	_ resource.ResourceWithValidateConfig = &deliveryGroupTypeFilterResource{}
	_ resource.ResourceWithModifyPlan     = &deliveryGroupTypeFilterResource{}
)

func NewDeliveryGroupTypeFilterResource() resource.Resource {
	return &deliveryGroupTypeFilterResource{}
}

type deliveryGroupTypeFilterResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *deliveryGroupTypeFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_group_type_policy_filter"
}

// Schema defines the schema for the resource.
func (r *deliveryGroupTypeFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DeliveryGroupTypeFilterModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *deliveryGroupTypeFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *deliveryGroupTypeFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan DeliveryGroupTypeFilterModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyFilter, err := createPolicyFilter(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	policyFilter, err = getPolicyFilter(ctx, r.client, &resp.Diagnostics, policyFilter.GetFilterGuid())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, *policyFilter)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupTypeFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DeliveryGroupTypeFilterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyFilter, err := getPolicyFilter(ctx, r.client, &resp.Diagnostics, state.GetId())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, *policyFilter)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupTypeFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan DeliveryGroupTypeFilterModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := updatePolicyFilter(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		return
	}

	policyFilter, err := getPolicyFilter(ctx, r.client, &resp.Diagnostics, plan.GetId())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, *policyFilter)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupTypeFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DeliveryGroupTypeFilterModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := deletePolicyFilter(ctx, r.client, &resp.Diagnostics, state.GetId())
	if err != nil {
		return
	}
}

func (r *deliveryGroupTypeFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *deliveryGroupTypeFilterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data DeliveryGroupTypeFilterModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *deliveryGroupTypeFilterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
