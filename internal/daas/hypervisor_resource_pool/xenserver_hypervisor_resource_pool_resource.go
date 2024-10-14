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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &xenserverHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &xenserverHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &xenserverHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &xenserverHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &xenserverHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewXenserverHypervisorResourcePoolResource() resource.Resource {
	return &xenserverHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type xenserverHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *xenserverHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xenserver_hypervisor_resource_pool"
}

func (r *xenserverHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = XenserverHypervisorResourcePoolResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *xenserverHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *xenserverHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan XenserverHypervisorResourcePoolResourceModel
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
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for XenServer",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)

	// storages, tempStorages, networks := SetResourceList(ctx, r.client, &resp.Diagnostics, hypervisorId, hypervisorConnectionType, plan)

	storages, tempStorages := plan.GetStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, true, false)
	networks := plan.GetNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, true)

	if len(storages) == 0 || len(tempStorages) == 0 || len(networks) == 0 {
		// Error handled in helper function.
		return
	}
	resourcePoolDetails.SetStorage(storages)
	resourcePoolDetails.SetTemporaryStorage(tempStorages)
	resourcePoolDetails.SetNetworks(networks)

	resourcePoolDetails.SetUseLocalStorageCaching(plan.UseLocalStorageCaching.ValueBool())

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	resourcePoolDetails.SetMetadata(metadata)

	resourcePool, err := CreateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, *hypervisor, resourcePoolDetails)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *xenserverHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state XenserverHypervisorResourcePoolResourceModel
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
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *xenserverHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan XenserverHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state XenserverHypervisorResourcePoolResourceModel
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

	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER {
		resp.Diagnostics.AddError(
			"Error creating Azure Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER)

	storagesToBeIncluded, tempStoragesToBeIncluded := plan.GetStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, false, false)
	storagesToBeSuperseded, tempStoragesToBeSuperseded := plan.GetStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, false, true)
	networks := plan.GetNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, false)
	if (len(storagesToBeIncluded) == 0 && len(storagesToBeSuperseded) == 0) || (len(tempStoragesToBeIncluded) == 0 && len(tempStoragesToBeSuperseded) == 0) || len(networks) == 0 {
		// Error handled in helper function.
		return
	}

	storageRequests := []citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
	for _, storage := range storagesToBeIncluded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		storageRequests = append(storageRequests, *request)
	}
	for _, storage := range storagesToBeSuperseded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		request.SetSuperseded(true)
		storageRequests = append(storageRequests, *request)
	}
	editHypervisorResourcePool.SetStorage(storageRequests)

	tempStorageRequests := []citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
	for _, storage := range tempStoragesToBeIncluded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		tempStorageRequests = append(tempStorageRequests, *request)
	}
	for _, storage := range tempStoragesToBeSuperseded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		request.SetSuperseded(true)
		tempStorageRequests = append(tempStorageRequests, *request)
	}
	editHypervisorResourcePool.SetTemporaryStorage(tempStorageRequests)

	editHypervisorResourcePool.SetNetworks(networks)

	editHypervisorResourcePool.SetUseLocalStorageCaching(plan.UseLocalStorageCaching.ValueBool())

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorResourcePool.SetMetadata(metadata)

	updatedResourcePool, err := UpdateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.Hypervisor.ValueString(), plan.Id.ValueString(), editHypervisorResourcePool)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedResourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *xenserverHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *xenserverHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state XenserverHypervisorResourcePoolResourceModel
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

func (plan XenserverHypervisorResourcePoolResourceModel) GetStorageList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool, forSuperseded bool) ([]string, []string) {
	action := "updating"
	if isCreate {
		action = "creating"
	}

	storage := []basetypes.StringValue{}
	storageFromPlan := util.ObjectListToTypedArray[HypervisorStorageModel](ctx, diags, plan.Storage)
	for _, storageModel := range storageFromPlan {
		if storageModel.Superseded.ValueBool() == forSuperseded {
			storage = append(storage, storageModel.StorageName)
		}
	}
	storageNames := util.ConvertBaseStringArrayToPrimitiveStringArray(storage)
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	storages, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, "", util.StorageResourceType, storageNames, hypervisorConnectionType, hypervisor.GetPluginId())

	if len(storage) > 0 && len(storages) == 0 {
		errDetail := "No storage found for the given storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Xenserver", action),
			errDetail,
		)
		return nil, nil
	}

	tempStorage := []basetypes.StringValue{}
	temporaryStorageFromPlan := util.ObjectListToTypedArray[HypervisorStorageModel](ctx, diags, plan.TemporaryStorage)
	for _, storageModel := range temporaryStorageFromPlan {
		if storageModel.Superseded.ValueBool() == forSuperseded {
			tempStorage = append(tempStorage, storageModel.StorageName)
		}
	}
	tempStorageNames := util.ConvertBaseStringArrayToPrimitiveStringArray(tempStorage)
	tempStorages, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, "", util.StorageResourceType, tempStorageNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(tempStorage) > 0 && len(tempStorages) == 0 {
		errDetail := "No storage found for the given temporary storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Xenserver", action),
			errDetail,
		)
		return nil, nil
	}

	return storages, tempStorages
}

func (plan XenserverHypervisorResourcePoolResourceModel) GetNetworksList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool) []string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	networkNames := util.StringListToStringArray(ctx, diags, plan.Networks)
	networks, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, "", util.NetworkResourceType, networkNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(networks) == 0 {
		errDetail := "No network found for the given network names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Xenserver", action),
			errDetail,
		)
		return nil
	}

	return networks
}

func (r *xenserverHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data XenserverHypervisorResourcePoolResourceModel
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

func (r *xenserverHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
