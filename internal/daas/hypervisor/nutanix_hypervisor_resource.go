// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &nutanixHypervisorResource{}
	_ resource.ResourceWithConfigure      = &nutanixHypervisorResource{}
	_ resource.ResourceWithImportState    = &nutanixHypervisorResource{}
	_ resource.ResourceWithValidateConfig = &nutanixHypervisorResource{}
	_ resource.ResourceWithModifyPlan     = &nutanixHypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewNutanixHypervisorResource() resource.Resource {
	return &nutanixHypervisorResource{}
}

type nutanixHypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata implements resource.Resource.
func (*nutanixHypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nutanix_hypervisor"
}

// Configure implements resource.ResourceWithConfigure.
func (r *nutanixHypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (r *nutanixHypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = NutanixHypervisorResourceModel{}.GetSchema()
}

// ImportState implements resource.ResourceWithImportState.
func (*nutanixHypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create implements resource.Resource.
func (r *nutanixHypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan NutanixHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM)
	connectionDetails.SetPluginId(util.NUTANIX_PLUGIN_ID)
	connectionDetails.SetUserName(plan.Username.ValueString())
	connectionDetails.SetPassword(plan.Password.ValueString())
	pwdFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.PasswordFormat.ValueString())
	if err != nil || pwdFormat == nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor for Nutanix",
			"Unsupported password format: "+plan.PasswordFormat.ValueString(),
		)
	}
	connectionDetails.SetPasswordFormat(*pwdFormat)

	addresses := util.StringListToStringArray(ctx, &diags, plan.Addresses)
	connectionDetails.SetAddresses(addresses)
	connectionDetails.SetMaxAbsoluteActiveActions(int32(plan.MaxAbsoluteActiveActions.ValueInt64()))
	connectionDetails.SetMaxAbsoluteNewActionsPerMinute(int32(plan.MaxAbsoluteNewActionsPerMinute.ValueInt64()))
	connectionDetails.SetMaxPowerActionsPercentageOfMachines(int32(plan.MaxPowerActionsPercentageOfMachines.ValueInt64()))
	if !plan.Scopes.IsNull() {
		connectionDetails.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	connectionDetails.SetMetadata(metadata)

	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	hypervisor, err := CreateHypervisor(ctx, r.client, &resp.Diagnostics, body)
	if err != nil {
		// Directly return. Error logs have been populated in common function.
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &diags, hypervisor)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *nutanixHypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state NutanixHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := state.Id.ValueString()
	hypervisor, err := readHypervisor(ctx, r.client, resp, hypervisorId)
	if err != nil {
		return
	}

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM || hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not a Nutanix connection type hypervisor.",
		)
		return
	}

	// Overwrite hypervisor with refreshed state
	state = state.RefreshPropertyValues(ctx, &diags, hypervisor)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *nutanixHypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan NutanixHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state NutanixHypervisorResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM)
	editHypervisorRequestBody.SetUserName(plan.Username.ValueString())
	editHypervisorRequestBody.SetPassword(plan.Password.ValueString())
	pwdFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.PasswordFormat.ValueString())
	if err != nil || pwdFormat == nil {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor for Nutanix",
			"Unsupported password format: "+plan.PasswordFormat.ValueString(),
		)
	}
	editHypervisorRequestBody.SetPasswordFormat(*pwdFormat)

	addresses := util.StringListToStringArray(ctx, &diags, plan.Addresses)
	editHypervisorRequestBody.SetAddresses(addresses)

	editHypervisorRequestBody.SetMaxAbsoluteActiveActions(int32(plan.MaxAbsoluteActiveActions.ValueInt64()))
	editHypervisorRequestBody.SetMaxAbsoluteNewActionsPerMinute(int32(plan.MaxAbsoluteNewActionsPerMinute.ValueInt64()))
	editHypervisorRequestBody.SetMaxPowerActionsPercentageOfMachines(int32(plan.MaxPowerActionsPercentageOfMachines.ValueInt64()))
	if !plan.Scopes.IsNull() {
		editHypervisorRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorRequestBody.SetMetadata(metadata)

	// Patch hypervisor
	updatedHypervisor, err := UpdateHypervisor(ctx, r.client, &resp.Diagnostics, editHypervisorRequestBody, plan.Id.ValueString(), plan.Name.ValueString())
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &diags, updatedHypervisor)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *nutanixHypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state NutanixHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing hypervisor
	hypervisorId := state.Id.ValueString()
	hypervisorName := state.Name.ValueString()
	deleteHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisor(ctx, hypervisorId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorRequest, r.client).Async(true).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Hypervisor "+hypervisorName, &resp.Diagnostics, 5)
	if err != nil {
		return
	}
}

func (r *nutanixHypervisorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data NutanixHypervisorResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *nutanixHypervisorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
