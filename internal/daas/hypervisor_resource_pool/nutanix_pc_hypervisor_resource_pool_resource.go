// Copyright © 2026. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"fmt"
	"net/http"
	"slices"
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
	_ resource.Resource                   = &nutanixPCHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &nutanixPCHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &nutanixPCHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &nutanixPCHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &nutanixPCHypervisorResourcePoolResource{}
)

// NewNutanixPCHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewNutanixPCHypervisorResourcePoolResource() resource.Resource {
	return &nutanixPCHypervisorResourcePoolResource{}
}

type nutanixPCHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (*nutanixPCHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nutanix_pc_hypervisor_resource_pool"
}

func (*nutanixPCHypervisorResourcePoolResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = NutanixPCHypervisorResourcePoolResourceModel{}.GetSchema()
}

func (r *nutanixPCHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient) //nolint:forcetypeassert // framework guarantee
}

func (r *nutanixPCHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan NutanixPCHypervisorResourcePoolResourceModel
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

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Nutanix Prism Central",
			"Hypervisor "+hypervisor.GetName()+" is not a Nutanix AHV Prism Central hypervisor.",
		)
		return
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC)
	resourcePoolDetails.SetVmTagging(plan.VmTagging.ValueBool())

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	resourcePoolDetails.SetMetadata(metadata)

	timeoutConfigs := util.ObjectValueToTypedObject[ResourcePoolTimeout](ctx, &resp.Diagnostics, plan.Timeout)
	createTimeout := timeoutConfigs.Create.ValueInt32()
	if createTimeout == 0 {
		createTimeout = getResourcePoolTimeoutConfigs().CreateDefault
	}

	resourcePool, err := CreateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, *hypervisor, resourcePoolDetails, createTimeout)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nutanixPCHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state NutanixPCHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourcePool, err := ReadHypervisorResourcePool(ctx, r.client, resp, state.Hypervisor.ValueString(), state.Id.ValueString())
	if err != nil {
		return
	}

	if resourcePool.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor Resource Pool",
			"Resource pool "+resourcePool.GetName()+" is not a Nutanix AHV Prism Central resource pool.",
		)
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nutanixPCHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan NutanixPCHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NutanixPCHypervisorResourcePoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Nutanix Prism Central",
			"Hypervisor "+hypervisor.GetName()+" is not a Nutanix AHV Prism Central hypervisor.",
		)
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC)
	editHypervisorResourcePool.SetVmTagging(plan.VmTagging.ValueBool())

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorResourcePool.SetMetadata(metadata)

	timeoutConfigs := util.ObjectValueToTypedObject[ResourcePoolTimeout](ctx, &resp.Diagnostics, plan.Timeout)
	updateTimeout := timeoutConfigs.Update.ValueInt32()
	if updateTimeout == 0 {
		updateTimeout = getResourcePoolTimeoutConfigs().UpdateDefault
	}

	resourcePool, err := UpdateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, hypervisorId, plan.Id.ValueString(), editHypervisorResourcePool, updateTimeout)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *nutanixPCHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) > 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: hypervisorResourcePoolId or hypervisorId,hypervisorResourcePoolId. Got: %q", req.ID),
		)
		return
	}

	if len(idParts) == 2 {
		if idParts[0] == "" || idParts[1] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: hypervisorId,hypervisorResourcePoolId. Got: %q", req.ID),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hypervisor"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
		return
	}

	// Support for importing hypervisor resource pool with only resource pool id
	getHypervisorsAndResourcePoolsRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorsAndResourcePools(ctx)
	hypervisorAndResourcePools, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.HypervisorsAndResourcePoolsResponseModelCollection](getHypervisorsAndResourcePoolsRequest, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Hypervisor and Resource Pools",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	for _, hypervisorAndResourcePool := range hypervisorAndResourcePools.GetItems() {
		if hypervisorAndResourcePool.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC {
			continue
		}

		resourcePools := hypervisorAndResourcePool.GetResourcePools()
		if slices.ContainsFunc(resourcePools, func(resourcePool citrixorchestration.HypervisorBaseResponseModel) bool {
			return strings.EqualFold(resourcePool.GetId(), req.ID)
		}) {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hypervisor"), hypervisorAndResourcePool.GetId())...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Error importing Hypervisor Resource Pool",
		fmt.Sprintf("Hypervisor Resource Pool with ID %q not found", req.ID),
	)
}

func (r *nutanixPCHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state NutanixPCHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := state.Hypervisor.ValueString()
	deleteHypervisorResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisorResourcePool(ctx, hypervisorId, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorResourcePoolRequest, r.client).Async(true).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Resource Pool for Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	timeoutConfigs := util.ObjectValueToTypedObject[ResourcePoolTimeout](ctx, &resp.Diagnostics, state.Timeout)
	deleteTimeout := timeoutConfigs.Delete.ValueInt32()
	if deleteTimeout == 0 {
		deleteTimeout = getResourcePoolTimeoutConfigs().DeleteDefault
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Resource Pool for Hypervisor "+hypervisorId, &resp.Diagnostics, deleteTimeout)
	if err != nil {
		return
	}
}

func (r *nutanixPCHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data NutanixPCHypervisorResourcePoolResourceModel
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

func (r *nutanixPCHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
