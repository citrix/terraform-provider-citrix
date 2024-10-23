// Copyright © 2024. Citrix Systems, Inc.

package desktop_icon

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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &desktopIconResource{}
	_ resource.ResourceWithConfigure      = &desktopIconResource{}
	_ resource.ResourceWithImportState    = &desktopIconResource{}
	_ resource.ResourceWithValidateConfig = &desktopIconResource{}
	_ resource.ResourceWithModifyPlan     = &desktopIconResource{}
)

// NewDesktopIconResource is a helper function to simplify the provider implementation.
func NewDesktopIconResource() resource.Resource {
	return &desktopIconResource{}
}

// desktopIconResource is the resource implementation.
type desktopIconResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *desktopIconResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_desktop_icon"
}

// Configure adds the provider configured client to the data source.
func (r *desktopIconResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema returns the resource schema.
func (r *desktopIconResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DesktopIconResourceModel{}.GetSchema()
}

func (r *desktopIconResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan DesktopIconResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createDesktopIconRequest citrixorchestration.AddIconRequestModel
	if !plan.RawData.IsNull() {
		createDesktopIconRequest.SetRawData(plan.RawData.ValueString())
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
		createDesktopIconRequest.SetRawData(base64String)
	}

	// Set default icon format to 32x32x24 png format
	createDesktopIconRequest.SetIconFormat("image/png;32x32x24")

	// Create new desktop icon
	addDesktopIconRequest := r.client.ApiClient.IconsAPIsDAAS.IconsAddIcon(ctx)
	addDesktopIconRequest = addDesktopIconRequest.AddIconRequestModel(createDesktopIconRequest)

	desktopIcon, httpResp, err := citrixdaasclient.AddRequestData(addDesktopIconRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Desktop Icon",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(desktopIcon)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *desktopIconResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DesktopIconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	desktopIconId := state.Id.ValueString()
	// Get refreshed desktop icon properties from Orchestration
	desktopIcon, err := readDesktopIcon(ctx, r.client, resp, desktopIconId)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(desktopIcon)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *desktopIconResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	resp.Diagnostics.AddError("Unsupported Operation", "Update is not supported for this resource")
}

func (r *desktopIconResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DesktopIconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	desktopIconId, err := strconv.ParseInt(state.Id.ValueString(), 10, 32)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Icon", "Invalid Icon Id")
		return
	}

	deleteDesktopIconRequest := r.client.ApiClient.IconsAPIsDAAS.IconsRemoveIcon(ctx, int32(desktopIconId))
	httpResp, err := citrixdaasclient.AddRequestData(deleteDesktopIconRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Desktop Icon "+state.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *desktopIconResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readDesktopIcon(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, desktopIconId string) (*citrixorchestration.IconResponseModel, error) {
	getDesktopIconRequest := client.ApiClient.IconsAPIsDAAS.IconsGetIcon(ctx, desktopIconId)
	desktopIcon, _, err := util.ReadResource[*citrixorchestration.IconResponseModel](getDesktopIconRequest, ctx, client, resp, "Desktop Icon", desktopIconId)
	return desktopIcon, err
}

func (r *desktopIconResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data DesktopIconResourceModel
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

func (r *desktopIconResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
