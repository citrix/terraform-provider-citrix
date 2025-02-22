// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"
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
	_ resource.Resource                   = &applicationIconResource{}
	_ resource.ResourceWithConfigure      = &applicationIconResource{}
	_ resource.ResourceWithImportState    = &applicationIconResource{}
	_ resource.ResourceWithValidateConfig = &applicationIconResource{}
	_ resource.ResourceWithModifyPlan     = &applicationIconResource{}
)

// NewApplicationIconResource is a helper function to simplify the provider implementation.
func NewApplicationIconResource() resource.Resource {
	return &applicationIconResource{}
}

// applicationIconResource is the resource implementation.
type applicationIconResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *applicationIconResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_icon"
}

// Configure adds the provider configured client to the data source.
func (r *applicationIconResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema returns the resource schema.
func (r *applicationIconResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApplicationIconResourceModel{}.GetSchema()
}

func (r *applicationIconResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationIconResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	icons, iconErr := GetAllExistingIcons(ctx, r.client, &resp.Diagnostics)
	if iconErr != nil {
		return
	}

	// Generate API request body from plan
	var createApplicationIconRequest citrixorchestration.AddIconRequestModel
	if !plan.RawData.IsNull() {
		createApplicationIconRequest.SetRawData(plan.RawData.ValueString())
	} else {
		bytes, err := os.ReadFile(plan.FilePath.ValueString())
		if err != nil {
			if os.IsPermission(err) {
				resp.Diagnostics.AddError(
					"Error reading icon file",
					"Permission denied to read icon file: "+plan.FilePath.ValueString()+
						"\nError message: "+err.Error(),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error reading file",
				err.Error(),
			)
			return
		}
		base64String := base64.StdEncoding.EncodeToString(bytes)
		createApplicationIconRequest.SetRawData(base64String)
	}

	// Set default icon format to 32x32x24 png format
	createApplicationIconRequest.SetIconFormat("image/png;32x32x24")

	// Create new application icon
	addApplicationIconRequest := r.client.ApiClient.IconsAPIsDAAS.IconsAddIcon(ctx)
	addApplicationIconRequest = addApplicationIconRequest.AddIconRequestModel(createApplicationIconRequest)

	applicationIcon, httpResp, err := citrixdaasclient.AddRequestData(addApplicationIconRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Application Icon",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	for _, item := range icons {
		if item.Id == applicationIcon.Id {
			resp.Diagnostics.AddError(
				"Icon already exists.",
				"\nIcon ID: "+applicationIcon.Id,
			)
			return
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(applicationIcon)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *applicationIconResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ApplicationIconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationIconId := state.Id.ValueString()
	// Get refreshed application properties from Orchestration
	applicationIcon, err := readApplicationIcon(ctx, r.client, resp, applicationIconId)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(applicationIcon)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *applicationIconResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	resp.Diagnostics.AddError("Unsupported Operation", "Update is not supported for this resource")
}

func (r *applicationIconResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ApplicationIconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationIconId, err := strconv.ParseInt(state.Id.ValueString(), 10, 32)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Icon", "Invalid Icon Id")
		return
	}

	deleteApplicationIconRequest := r.client.ApiClient.IconsAPIsDAAS.IconsRemoveIcon(ctx, int32(applicationIconId))
	httpResp, err := citrixdaasclient.AddRequestData(deleteApplicationIconRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Application Icon "+state.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *applicationIconResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readApplicationIcon(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, applicationIconId string) (*citrixorchestration.IconResponseModel, error) {
	getApplicationIconRequest := client.ApiClient.IconsAPIsDAAS.IconsGetIcon(ctx, applicationIconId)
	applicationIcon, _, err := util.ReadResource[*citrixorchestration.IconResponseModel](getApplicationIconRequest, ctx, client, resp, "Application Icon", applicationIconId)
	return applicationIcon, err
}

func GetAllExistingIcons(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.IconResponseModel, error) {
	req := client.ApiClient.IconsAPIsDAAS.IconsGetIcons(ctx)

	responses := []citrixorchestration.IconResponseModel{}
	continuationToken := ""
	for {
		req = req.ContinuationToken(continuationToken)
		responseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.IconResponseModelCollection](req, client)
		if err != nil {
			diagnostics.AddError(
				"Error getting all the existing icons",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return responses, err
		}
		responses = append(responses, responseModel.GetItems()...)
		if responseModel.GetContinuationToken() == "" {
			return responses, nil
		}
		continuationToken = responseModel.GetContinuationToken()
	}

}

func (r *applicationIconResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ApplicationIconResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.FilePath.IsNull() && !strings.HasSuffix(strings.ToLower(data.FilePath.ValueString()), ".ico") {
		resp.Diagnostics.AddError(
			"Invalid file format",
			"Only `.ico` icon file format is supported",
		)
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *applicationIconResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
