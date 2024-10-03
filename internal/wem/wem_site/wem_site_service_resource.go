// Copyright © 2024. Citrix Systems, Inc.

package wem_site

import (
	"context"
	"strconv"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &wemSiteServiceResource{}
	_ resource.ResourceWithConfigure   = &wemSiteServiceResource{}
	_ resource.ResourceWithImportState = &wemSiteServiceResource{}
	_ resource.ResourceWithModifyPlan  = &wemSiteServiceResource{}
)

// wemSiteServiceResource is the resource implementation.
type wemSiteServiceResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ImportState implements resource.ResourceWithImportState.
func (w *wemSiteServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Metadata implements resource.Resource.
func (w *wemSiteServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wem_configuration_set"
}

// Configure implements resource.ResourceWithConfigure.
func (w *wemSiteServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	w.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// NewWemSiteServiceResource is a helper function to simplify the provider implementation.
func NewWemSiteServiceResource() resource.Resource {
	return &wemSiteServiceResource{}
}

// Schema implements resource.Resource.
func (w *wemSiteServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WemSiteResourceModel{}.GetSchema()
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (w *wemSiteServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if w.client != nil && (w.client.ApiClient == nil || w.client.WemClient == nil) {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if w.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing WEM Configuration Sets", "Configuration Sets are only supported for Cloud customers.")
	}
}

// Create implements resource.Resource.
func (w *wemSiteServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan WemSiteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixwemservice.SiteModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())

	// Generate Create API request
	siteCreateRequest := w.client.WemClient.SiteDAAS.SiteCreate(ctx)
	siteCreateRequest = siteCreateRequest.Body(body)
	httpResp, err := citrixdaasclient.AddRequestData(siteCreateRequest, w.client).Execute()

	// In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating WEM site "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get Newly created site by name from remote (ID is not available yet)
	siteConfig, err := getSiteByName(ctx, w.client, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching WEM site",
			util.ReadClientError(err),
		)
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &siteConfig)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (w *wemSiteServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state WemSiteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert site Id to int64
	siteId, err := strconv.ParseInt(state.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting site Id to int64",
			err.Error(),
		)
		return
	}

	// Generate Delete API request
	siteDeleteRequest := w.client.WemClient.SiteDAAS.SiteDelete(ctx, siteId)
	httpResp, err := citrixdaasclient.AddRequestData(siteDeleteRequest, w.client).Execute()

	// In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting WEM site "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// Read implements resource.Resource.
func (w *wemSiteServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state WemSiteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get site from remote using site Id
	siteConfig, err := getSiteById(ctx, w.client, state)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, siteConfig)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (w *wemSiteServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get plan values
	var plan WemSiteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixwemservice.SiteModel
	siteId, err := strconv.ParseInt(plan.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting site Id to int64",
			err.Error(),
		)
		return
	}
	body.SetId(siteId)
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())

	// Generate Update API request
	siteUpdateRequest := w.client.WemClient.SiteDAAS.SiteUpdate(ctx)
	siteUpdateRequest = siteUpdateRequest.Body(body)
	httpResp, err := citrixdaasclient.AddRequestData(siteUpdateRequest, w.client).Execute()

	// In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating WEM site "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get Newly Updated site from remote by Id
	siteConfig, err := getSiteById(ctx, w.client, plan)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, siteConfig)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
