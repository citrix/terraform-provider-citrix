// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Vpc     types.String   `tfsdk:"vpc"`
	Subnets []types.String `tfsdk:"subnets"`
	/** AWS Resource Pool **/
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

func (r AwsHypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) AwsHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	virtualNetwork := resourcePool.GetVirtualPrivateCloud()
	r.Vpc = types.StringValue(virtualNetwork.GetName())
	availabilityZone := resourcePool.GetAvailabilityZone()
	r.AvailabilityZone = types.StringValue(strings.Split(availabilityZone.GetName(), " ")[0])
	var res []string
	for _, model := range resourcePool.GetNetworks() {
		name := model.GetName()
		res = append(res, strings.Split(name, " ")[0])
	}
	r.Subnets = util.ConvertPrimitiveStringArrayToBaseStringArray(res)

	return r
}
