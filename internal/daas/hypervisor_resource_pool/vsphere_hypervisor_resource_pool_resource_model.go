// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VsphereHypervisorStorageModel struct {
	StorageName types.String `tfsdk:"storage_name"`
	Superseded  types.Bool   `tfsdk:"superseded"`
}

type VsphereHypervisorClusterModel struct {
	Datacenter  types.String `tfsdk:"datacenter"`
	ClusterName types.String `tfsdk:"cluster_name"`
	Host        types.String `tfsdk:"host"`
}

type VsphereHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Cluster                *VsphereHypervisorClusterModel  `tfsdk:"cluster"`
	Networks               []types.String                  `tfsdk:"networks"`
	Storage                []VsphereHypervisorStorageModel `tfsdk:"storage"`
	TemporaryStorage       []VsphereHypervisorStorageModel `tfsdk:"temporary_storage"`
	UseLocalStorageCaching types.Bool                      `tfsdk:"use_local_storage_caching"`
}

func (r VsphereHypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) VsphereHypervisorResourcePoolResourceModel {

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
	r.Storage = util.RefreshListProperties[VsphereHypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](r.Storage, "StorageName", resourcePool.GetStorage(), "Name", "RefreshListItem")
	r.TemporaryStorage = util.RefreshListProperties[VsphereHypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](r.TemporaryStorage, "StorageName", resourcePool.GetTemporaryStorage(), "Name", "RefreshListItem")

	return r
}

func (v VsphereHypervisorStorageModel) RefreshListItem(remote citrixorchestration.HypervisorStorageResourceResponseModel) VsphereHypervisorStorageModel {
	v.StorageName = types.StringValue(remote.GetName())
	v.Superseded = types.BoolValue(remote.GetSuperseded())
	return v
}
