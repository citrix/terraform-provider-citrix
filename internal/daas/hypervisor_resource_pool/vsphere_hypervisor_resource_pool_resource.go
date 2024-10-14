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
	_ resource.Resource                   = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &vsphereHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewVsphereHypervisorResourcePoolResource() resource.Resource {
	return &vsphereHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type vsphereHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *vsphereHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan VsphereHypervisorResourcePoolResourceModel
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
func (r *vsphereHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vsphere_hypervisor_resource_pool"
}

func (r *vsphereHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = VsphereHypervisorResourcePoolResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *vsphereHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *vsphereHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan VsphereHypervisorResourcePoolResourceModel
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

	folderPath := hypervisor.GetXDPath()
	var relativePath string

	cluster := util.ObjectValueToTypedObject[VsphereHypervisorClusterModel](ctx, &resp.Diagnostics, plan.Cluster)
	resource, httpResp, err := util.GetSingleHypervisorResource(ctx, r.client, &resp.Diagnostics, hypervisorId, folderPath, cluster.Datacenter.ValueString(), "datacenter", "", hypervisor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for vSphere",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\nFailed to resolve resource %s, error: %s", cluster.Datacenter.ValueString(), err.Error()),
		)
		return
	}
	folderPath = resource.GetXDPath()
	relativePath = resource.GetRelativePath()

	if !cluster.ClusterName.IsNull() {
		resource, httpResp, err = util.GetSingleHypervisorResource(ctx, r.client, &resp.Diagnostics, hypervisorId, folderPath, cluster.ClusterName.ValueString(), "cluster", "", hypervisor)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for vSphere",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve resource %s, error: %s", cluster.ClusterName.ValueString(), err.Error()),
			)
			return
		}
		folderPath = resource.GetXDPath()
		relativePath = resource.GetRelativePath()
	}

	if !cluster.Host.IsNull() {
		resource, httpResp, err = util.GetSingleHypervisorResource(ctx, r.client, &resp.Diagnostics, hypervisorId, folderPath, cluster.Host.ValueString(), "computeresource", "", hypervisor)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for vSphere",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve resource %s, error: %s", cluster.Host.ValueString(), err.Error()),
			)
			return
		}

		relativePath = resource.GetRelativePath()
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for vSphere",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)
	resourcePoolDetails.SetRootPath(relativePath)
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

func (r *vsphereHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state VsphereHypervisorResourcePoolResourceModel
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

func (r *vsphereHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan VsphereHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VsphereHypervisorResourcePoolResourceModel
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
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER)

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

func (r *vsphereHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *vsphereHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state VsphereHypervisorResourcePoolResourceModel
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

func (plan VsphereHypervisorResourcePoolResourceModel) GetStorageList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool, forSuperseded bool) ([]string, []string) {
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
	cluster := util.ObjectValueToTypedObject[VsphereHypervisorClusterModel](ctx, diags, plan.Cluster)
	folderPath := fmt.Sprintf("%s\\%s.datacenter", hypervisor.GetXDPath(), cluster.Datacenter.ValueString())
	if !cluster.ClusterName.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.cluster", folderPath, cluster.ClusterName.ValueString())
	}

	if !cluster.Host.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.computeresource", folderPath, cluster.Host.ValueString())
	}
	storages, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, folderPath, util.StorageResourceType, storageNames, hypervisorConnectionType, hypervisor.GetPluginId())

	if len(storage) > 0 && len(storages) == 0 {
		errDetail := "No storage found for the given storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for vSphere", action),
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
	tempStorages, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, folderPath, util.StorageResourceType, tempStorageNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(tempStorage) > 0 && len(tempStorages) == 0 {
		errDetail := "No storage found for the given temporary storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for vSphere", action),
			errDetail,
		)
		return nil, nil
	}

	return storages, tempStorages
}

func (plan VsphereHypervisorResourcePoolResourceModel) GetNetworksList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, isCreate bool) []string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	cluster := util.ObjectValueToTypedObject[VsphereHypervisorClusterModel](ctx, diags, plan.Cluster)
	folderPath := fmt.Sprintf("%s\\%s.datacenter", hypervisor.GetXDPath(), cluster.Datacenter.ValueString())
	if !cluster.ClusterName.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.cluster", folderPath, cluster.ClusterName.ValueString())
	}

	if !cluster.Host.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.computeresource", folderPath, cluster.Host.ValueString())
	}

	networkNames := util.StringListToStringArray(ctx, diags, plan.Networks)
	networks, err := util.GetFilteredResourcePathList(ctx, client, diags, hypervisorId, folderPath, util.NetworkResourceType, networkNames, hypervisorConnectionType, hypervisor.GetPluginId())
	if len(networks) == 0 {
		errDetail := "No network found for the given network names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for vSphere", action),
			errDetail,
		)
		return nil
	}

	return networks
}

func (r *vsphereHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data VsphereHypervisorResourcePoolResourceModel
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
