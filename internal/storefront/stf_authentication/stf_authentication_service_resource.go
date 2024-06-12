// Copyright © 2024. Citrix Systems, Inc.

package stf_authentication

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfAuthenticationServiceResource{}
	_ resource.ResourceWithConfigure   = &stfAuthenticationServiceResource{}
	_ resource.ResourceWithImportState = &stfAuthenticationServiceResource{}
)

// NewSTFAuthenticationServiceResource is a helper function to simplify the provider implementation.
func NewSTFAuthenticationServiceResource() resource.Resource {
	return &stfAuthenticationServiceResource{}
}

// stfAuthenticationServiceResource is the resource implementation.
type stfAuthenticationServiceResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfAuthenticationServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_authentication_service"
}

// Schema defines the schema for the resource.
func (r *stfAuthenticationServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "StoreFront Authentication Service.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site to configure the authentication service for. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_path": schema.StringAttribute{
				Description: "The IIS virtual path to use for the authentication service. Defaults to `/Citrix/Authentication`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/Citrix/Authentication"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name the authentication service should be known as. Defaults to `Authentication Service`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Authentication Service"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *stfAuthenticationServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create implements resource.Resource.
func (r *stfAuthenticationServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFAuthenticationServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.AddSTFAuthenticationServiceRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Authentication Service ",
			"\nError message: "+err.Error(),
		)
		return
	}

	body.SetSiteId(siteIdInt)
	body.SetVirtualPath(plan.VirtualPath.ValueString())
	body.SetFriendlyName(plan.FriendlyName.ValueString())

	addAuthenticationServiceRequest := r.client.StorefrontClient.AuthenticationServiceSF.STFAuthenticationCreateSTFAuthenticationService(ctx, body)

	// Create new STF Deployment
	authenticationServiceDetail, err := addAuthenticationServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding StoreFront Authentication Service",
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(&authenticationServiceDetail)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *stfAuthenticationServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFAuthenticationServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	STFAuthenticationService, err := getSTFAuthenticationService(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		return
	}

	if STFAuthenticationService == nil {
		resp.Diagnostics.AddWarning(
			"Authentication Service not found",
			"Authentication Service was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	state.RefreshPropertyValues(STFAuthenticationService)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *stfAuthenticationServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFAuthenticationServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state STFAuthenticationServiceResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var getBody citrixstorefront.GetSTFAuthenticationServiceRequestModel
	if !state.SiteId.IsNull() {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating StoreFront Authentication Service ",
				"\nError message: "+err.Error(),
			)
			return
		}
		getBody.SetSiteId(siteIdInt)
	}
	if state.VirtualPath.ValueString() != "" {
		getBody.SetVirtualPath(state.VirtualPath.ValueString())
	}

	// Remove existing STF Authentication Service
	removeAuthenticationServiceRequest := r.client.StorefrontClient.AuthenticationServiceSF.STFAuthenticationRemoveSTFAuthenticationService(ctx, getBody)
	err := removeAuthenticationServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Authentication Service ",
			"\nError message: "+err.Error(),
		)
		return
	}

	// Add updated STF Authentication Service
	var createBody citrixstorefront.AddSTFAuthenticationServiceRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Authentication Service ",
			"\nError message: "+err.Error(),
		)
		return
	}

	createBody.SetSiteId(siteIdInt)
	createBody.SetVirtualPath(plan.VirtualPath.ValueString())
	createBody.SetFriendlyName(plan.FriendlyName.ValueString())

	addAuthenticationServiceRequest := r.client.StorefrontClient.AuthenticationServiceSF.STFAuthenticationCreateSTFAuthenticationService(ctx, createBody)

	authenticationServiceDetail, err := addAuthenticationServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding StoreFront Authentication Service",
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(&authenticationServiceDetail)

	// Set refreshed state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *stfAuthenticationServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFAuthenticationServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var getBody citrixstorefront.GetSTFAuthenticationServiceRequestModel
	if !state.SiteId.IsNull() {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing StoreFront Authentication Service ",
				"\nError message: "+err.Error(),
			)
			return
		}
		getBody.SetSiteId(siteIdInt)
	}
	if state.VirtualPath.ValueString() != "" {
		getBody.SetVirtualPath(state.VirtualPath.ValueString())
	}

	// Remove STF Authentication Service
	removeAuthenticationServiceRequest := r.client.StorefrontClient.AuthenticationServiceSF.STFAuthenticationRemoveSTFAuthenticationService(ctx, getBody)
	err := removeAuthenticationServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing StoreFront Authentication Service ",
			"\nError message: "+err.Error(),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *stfAuthenticationServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `site_id,virtual_path`, got: %q", req.ID),
		)
		return
	}

	_, err := strconv.Atoi(idSegments[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Site ID in Import Identifier",
			fmt.Sprintf("Site ID should be an integer, got: %q", idSegments[0]),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_path"), idSegments[1])...)
}

// Gets the getSTFAuthenticationService and logs any errors
func getSTFAuthenticationService(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, state STFAuthenticationServiceResourceModel) (*citrixstorefront.STFAuthenticationServiceResponseModel, error) {
	var body citrixstorefront.GetSTFAuthenticationServiceRequestModel
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of StoreFront Authentication Service ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}

	if state.VirtualPath.ValueString() != "" {
		body.SetVirtualPath(state.VirtualPath.ValueString())
	}

	getSTFAuthenticationServiceRequest := client.StorefrontClient.AuthenticationServiceSF.STFAuthenticationGetSTFAuthenticationService(ctx, body)

	// Get refreshed STFAuthenticationService properties from Orchestration
	STFAuthenticationService, err := getSTFAuthenticationServiceRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			return nil, nil
		}
		diagnostics.AddError(
			"Error fetching state of StoreFront Authentication Service ",
			"Error message: "+err.Error(),
		)
		return &STFAuthenticationService, err
	}
	return &STFAuthenticationService, nil
}
