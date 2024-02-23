// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type XenserverHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Networks               []types.String `tfsdk:"networks"`
	Storage                []types.String `tfsdk:"storage"`
	TemporaryStorage       []types.String `tfsdk:"temporary_storage"`
	UseLocalStorageCaching types.Bool     `tfsdk:"use_local_storage_caching"`
}

func (r XenserverHypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) XenserverHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	r.UseLocalStorageCaching = types.BoolValue(resourcePool.GetUseLocalStorageCaching())

	remoteNetwork := []string{}
	for _, network := range resourcePool.GetNetworks() {
		remoteNetwork = append(remoteNetwork, network.GetName())
	}
	r.Networks = util.RefreshList(r.Networks, remoteNetwork)

	remoteStorage := []string{}
	for _, storage := range resourcePool.GetStorage() {
		remoteStorage = append(remoteStorage, storage.GetName())
	}
	r.Storage = util.RefreshList(r.Storage, remoteStorage)

	remoteTempStorage := []string{}
	for _, storage := range resourcePool.GetTemporaryStorage() {
		remoteTempStorage = append(remoteTempStorage, storage.GetName())
	}
	r.TemporaryStorage = util.RefreshList(r.TemporaryStorage, remoteTempStorage)

	return r
}
