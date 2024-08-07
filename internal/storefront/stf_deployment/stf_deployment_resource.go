// Copyright © 2024. Citrix Systems, Inc.
package stf_deployment

import (
	"context"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &stfDeploymentResource{}
	_ resource.ResourceWithConfigure      = &stfDeploymentResource{}
	_ resource.ResourceWithImportState    = &stfDeploymentResource{}
	_ resource.ResourceWithValidateConfig = &stfDeploymentResource{}
)

// stfDeploymentResource is a helper function to simplify the provider implementation.
func NewSTFDeploymentResource() resource.Resource {
	return &stfDeploymentResource{}
}

// stfDeploymentResource is the resource implementation.
type stfDeploymentResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (*stfDeploymentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data STFDeploymentResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Metadata returns the resource type name.
func (r *stfDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_deployment"
}

// Schema defines the schema for the resource.
func (r *stfDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = STFDeploymentResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *stfDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.CreateSTFDeploymentRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
		return
	}

	body.SetSiteId(siteIdInt)
	body.SetHostBaseUrl(plan.HostBaseUrl.ValueString())

	createDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentCreateSTFDeployment(ctx, body)

	// Create new STF Deployment
	DeploymentDetail, err := createDeploymentRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Deployment",
			"Error message: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(&DeploymentDetail)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stfDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil {
		return
	}
	if deployment == nil {
		resp.Diagnostics.AddWarning(
			"StoreFront Deployment not found",
			"StoreFront Deployment was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
		)
		resp.State.RemoveResource(ctx)
		return
	}
	state.RefreshPropertyValues(deployment)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed STFDeployment
	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	// Construct the update model
	var editSTFDeploymentBody citrixstorefront.SetSTFDeploymentRequestModel
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching state of StoreFront Authentication Service ",
			"Error message: "+err.Error(),
		)
		return
	}
	editSTFDeploymentBody.SetSiteId(siteIdInt)
	editSTFDeploymentBody.SetHostBaseUrl(plan.HostBaseUrl.ValueString())

	// Update STFDeployment
	editDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentSetSTFDeployment(ctx, editSTFDeploymentBody)
	_, err = editDeploymentRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
	}

	// Fetch updated STFDeployment
	updatedSTFDeployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan.RefreshPropertyValues(updatedSTFDeployment)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if STFDeployment exists
	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	// Delete existing STF Deployment
	var body citrixstorefront.ClearSTFDeploymentRequestModel
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting StoreFront Deployment ",
				"Error message: "+err.Error(),
			)
			return
		}
		body.SetSiteId(siteIdInt)
	}
	deleteDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentClearSTFDeployment(ctx, body)
	_, err = deleteDeploymentRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
		return
	}
}

func (r *stfDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("site_id"), req, resp)
}

// Gets the STFDeployment and logs any errors
func GetSTFDeployment(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId *string) (*citrixstorefront.STFDeploymentDetailModel, error) {
	var body citrixstorefront.GetSTFDeploymentRequestModel
	if siteId != nil {
		siteIdInt, err := strconv.ParseInt(*siteId, 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of StoreFront Deployment ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}
	getSTFDeploymentRequest := client.StorefrontClient.DeploymentSF.STFDeploymentGetSTFDeployment(ctx, body)

	// Get refreshed STFDeployment properties from Orchestration
	STFDeployment, err := getSTFDeploymentRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			return nil, nil
		}
		return &STFDeployment, err
	}
	return &STFDeployment, nil
}
