// Copyright © 2024. Citrix Systems, Inc.

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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &openshiftHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &openshiftHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &openshiftHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &openshiftHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &openshiftHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewOpenShiftHypervisorResourcePoolResource() resource.Resource {
	return &openshiftHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type openshiftHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *openshiftHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan OpenShiftHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !create {
		return
	}

	storage := util.ObjectListToTypedArray[HypervisorStorageModel](ctx, &resp.Diagnostics, plan.Storage)
	for _, storageModel := range storage {
		if storageModel.Superseded.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("storage"),
				"Incorrect attribute value",
				"Storage cannot be superseded when creating a new resource pool. Use only when updating the resource pool.",
			)
		}
	}

	temporaryStorage := util.ObjectListToTypedArray[HypervisorStorageModel](ctx, &resp.Diagnostics, plan.TemporaryStorage)
	for _, storageModel := range temporaryStorage {
		if storageModel.Superseded.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("temporary_storage"),
				"Incorrect attribute value",
				"Storage cannot be superseded when creating a new resource pool. Use only when updating the resource pool.",
			)
		}
	}
}

// Metadata returns the resource type name.
func (r *openshiftHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openshift_hypervisor_resource_pool"
}

func (r *openshiftHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = OpenShiftHypervisorResourcePoolResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *openshiftHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *openshiftHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan OpenShiftHypervisorResourcePoolResourceModel
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
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for OpenShift",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)

	namespace := plan.GetNamespace(ctx, r.client, &resp.Diagnostics, hypervisor, true)
	if namespace == "" {
		// Error handled in helper function.
		return
	}
	resourcePoolDetails.SetRootPath(namespace)

	storages, tempStorages := plan.GetStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, true, false)
	networks := plan.GetNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, true)
	if len(storages) == 0 || len(tempStorages) == 0 || len(networks) == 0 {
		// Error handled in helper function.
		return
	}
	resourcePoolDetails.SetStorage(storages)
	resourcePoolDetails.SetTemporaryStorage(tempStorages)
	resourcePoolDetails.SetNetworks(networks)

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

func (r *openshiftHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state OpenShiftHypervisorResourcePoolResourceModel
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

func (r *openshiftHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan OpenShiftHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OpenShiftHypervisorResourcePoolResourceModel
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
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT)

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

func (r *openshiftHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	hypervisorAndResourcePools, httpResp, err := citrixdaasclient.AddRequestData(getHypervisorsAndResourcePoolsRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Hypervisor and Resource Pools",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	for _, hypervisorAndResourcePool := range hypervisorAndResourcePools.GetItems() {
		if hypervisorAndResourcePool.ConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT {
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

func (r *openshiftHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state OpenShiftHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource pool
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

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Resource Pool for Hypervisor "+hypervisorId, &resp.Diagnostics, 5)
	if err != nil {
		return
	}
}

func (plan OpenShiftHypervisorResourcePoolResourceModel) GetNamespace(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool) string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	namespaceFromPlan := []string{plan.Namespace.ValueString()}

	namespaces, err := util.GetFilteredResourcePathListWithNoCacheRetry(ctx, client, diags, hypervisorId, "", util.NamespaceResourceType, namespaceFromPlan, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(namespaces) == 0 {
		errDetail := "No namespaces found for the given namespace value"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for OpenShift", action),
			errDetail,
		)
		return ""
	}

	return namespaces[0]
}

func (plan OpenShiftHypervisorResourcePoolResourceModel) GetStorageList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool, forSuperseded bool) ([]string, []string) {
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
	folderPath := fmt.Sprintf("%s\\%s.namespace", hypervisor.GetXDPath(), plan.Namespace.ValueString())
	storages, err := util.GetFilteredResourcePathListWithNoCacheRetry(ctx, client, diags, hypervisorId, folderPath, util.StorageResourceType, storageNames, hypervisorConnectionType, hypervisor.GetPluginId())

	if len(storage) > 0 && len(storages) == 0 {
		errDetail := "No storage found for the given storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for OpenShift", action),
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
	tempStorages, err := util.GetFilteredResourcePathListWithNoCacheRetry(ctx, client, diags, hypervisorId, folderPath, util.StorageResourceType, tempStorageNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(tempStorage) > 0 && len(tempStorages) == 0 {
		errDetail := "No storage found for the given temporary storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for OpenShift", action),
			errDetail,
		)
		return nil, nil
	}

	return storages, tempStorages
}

func (plan OpenShiftHypervisorResourcePoolResourceModel) GetNetworksList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool) []string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	folderPath := fmt.Sprintf("%s.namespace", plan.Namespace.ValueString())

	networkNames := util.StringListToStringArray(ctx, diags, plan.Networks)
	networks, err := util.GetFilteredResourcePathListWithNoCacheRetry(ctx, client, diags, hypervisorId, folderPath, util.NetworkResourceType, networkNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(networks) == 0 {
		errDetail := "No network found for the given network names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for OpenShift", action),
			errDetail,
		)
		return nil
	}

	return networks
}

func (r *openshiftHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data OpenShiftHypervisorResourcePoolResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
