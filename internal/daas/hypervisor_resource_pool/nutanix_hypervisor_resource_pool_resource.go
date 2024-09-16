// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

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
	_ resource.Resource                   = &nutanixHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &nutanixHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &nutanixHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &nutanixHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &nutanixHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewNutanixHypervisorResourcePoolResource() resource.Resource {
	return &nutanixHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type nutanixHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (*nutanixHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nutanix_hypervisor_resource_pool"
}

func (*nutanixHypervisorResourcePoolResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = NutanixHypervisorResourcePoolResourceModel{}.GetSchema()
}

func (r *nutanixHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *nutanixHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan NutanixHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)

	if err != nil {
		return
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Nutanix",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)
	resourcePoolDetails.SetRootPath("")
	networks := plan.GetNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, true)
	if len(networks) == 0 {
		// Error handled in helper function.
		return
	}
	resourcePoolDetails.SetNetworks(networks)

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	resourcePoolDetails.SetMetadata(metadata)

	resourcePool, err := CreateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, *hypervisor, resourcePoolDetails)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &diags, resourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *nutanixHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state NutanixHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get hypervisor properties from Orchestration
	hypervisorId := state.Hypervisor.ValueString()

	// Get the resource pool
	resourcePool, err := ReadHypervisorResourcePool(ctx, r.client, resp, hypervisorId, state.Id.ValueString())
	if err != nil {
		return
	}

	// Override with refreshed state
	state = state.RefreshPropertyValues(ctx, &diags, resourcePool)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *nutanixHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan NutanixHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)

	if err != nil {
		return
	}

	resourcePoolId := plan.Id.ValueString()

	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Nutanix",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM)

	networks := plan.GetNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, false)
	editHypervisorResourcePool.SetNetworks(networks)

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorResourcePool.SetMetadata(metadata)

	_, err = UpdateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.Hypervisor.ValueString(), plan.Id.ValueString(), editHypervisorResourcePool)
	if err != nil {
		return
	}

	resourcePool, err := util.GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, hypervisorId, resourcePoolId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &diags, resourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *nutanixHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: hypervisorId,hypervisorResourcePoolId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hypervisor"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (r *nutanixHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state NutanixHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource pool
	hypervisorId := state.Hypervisor.ValueString()
	deleteHypervisorResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisorResourcePool(ctx, hypervisorId, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorResourcePoolRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Resource Pool for Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (plan NutanixHypervisorResourcePoolResourceModel) GetNetworksList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool) []string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	pluginId := hypervisor.GetPluginId()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	networkNames := util.StringListToStringArray(ctx, diags, plan.Networks)
	networks, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, "", util.NetworkResourceType, networkNames, hypervisorConnectionType, pluginId)
	if len(networks) == 0 {
		errDetail := "No network found for the given network names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Nutanix", action),
			errDetail,
		)
		return nil
	}

	return networks
}

func (r *nutanixHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data NutanixHypervisorResourcePoolResourceModel
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

func (r *nutanixHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
