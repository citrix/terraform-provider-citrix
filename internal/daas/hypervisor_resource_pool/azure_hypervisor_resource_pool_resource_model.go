// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Region         types.String   `tfsdk:"region"`
	VirtualNetwork types.String   `tfsdk:"virtual_network"`
	Subnets        []types.String `tfsdk:"subnets"`
	/** Azure Resource Pool **/
	VirtualNetworkResourceGroup types.String `tfsdk:"virtual_network_resource_group"`
}

func (r AzureHypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) AzureHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	region := resourcePool.GetRegion()
	if r.shouldSetRegion(region) {
		r.Region = types.StringValue(region.GetName())
	}
	virtualNetwork := resourcePool.GetVirtualNetwork()
	resourceGroupName := getResourceGroupNameFromVnetId(virtualNetwork.GetId())
	r.VirtualNetworkResourceGroup = types.StringValue(resourceGroupName)
	r.VirtualNetwork = types.StringValue(virtualNetwork.GetName())
	var res []string
	for _, model := range resourcePool.GetSubnets() {
		res = append(res, model.GetName())
	}
	r.Subnets = util.ConvertPrimitiveStringArrayToBaseStringArray(res)

	return r
}

func (r AzureHypervisorResourcePoolResourceModel) shouldSetRegion(region citrixorchestration.HypervisorResourceRefResponseModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.IsNull() || r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}

func getResourceGroupNameFromVnetId(vnetId string) string {
	resourceGroupAndVnetName := strings.Split(vnetId, "/")
	return resourceGroupAndVnetName[0]
}
