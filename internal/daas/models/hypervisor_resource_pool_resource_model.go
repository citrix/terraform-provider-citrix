package models

import (
	"reflect"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HypervisorResourcePoolResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Hypervisor               types.String `tfsdk:"hypervisor"`
	HypervisorConnectionType types.String `tfsdk:"hypervisor_connection_type"`
	/**** Resource Pool Details ****/
	Region         types.String   `tfsdk:"region"`
	VirtualNetwork types.String   `tfsdk:"virtual_network"`
	Subnets        []types.String `tfsdk:"subnets"`
	/** Azure Resource Pool **/
	VirtualNetworkResourceGroup types.String `tfsdk:"virtual_network_resource_group"`
	/** AWS Resource Pool **/
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	/** GCP Resource Pool **/
	ProjectName types.String `tfsdk:"project_name"`
}

func (r HypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool citrixorchestration.HypervisorResourcePoolDetailResponseModel, hypervisorConnectionType citrixorchestration.HypervisorConnectionType) HypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())
	r.HypervisorConnectionType = types.StringValue(reflect.ValueOf(resourcePool.GetConnectionType()).String())

	switch hypervisorConnectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
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

	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		virtualNetwork := resourcePool.GetVirtualPrivateCloud()
		r.VirtualNetwork = types.StringValue(virtualNetwork.GetName())
		availabilityZone := resourcePool.GetAvailabilityZone()
		r.AvailabilityZone = types.StringValue(strings.Split(availabilityZone.GetName(), " ")[0])
		var res []string
		for _, model := range resourcePool.GetNetworks() {
			name := model.GetName()
			res = append(res, strings.Split(name, " ")[0])
		}
		r.Subnets = util.ConvertPrimitiveStringArrayToBaseStringArray(res)

	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		region := resourcePool.GetRegion()
		if r.shouldSetRegion(region) {
			r.Region = types.StringValue(region.GetName())
		}
		project := resourcePool.GetProject()
		r.ProjectName = types.StringValue(project.GetName())
		virtualNetwork := resourcePool.GetVirtualPrivateCloud()
		r.VirtualNetwork = types.StringValue(virtualNetwork.GetName())
		var res []string
		for _, model := range resourcePool.GetNetworks() {
			res = append(res, model.GetName())
		}
		r.Subnets = util.ConvertPrimitiveStringArrayToBaseStringArray(res)
	}

	return r
}

func (r HypervisorResourcePoolResourceModel) shouldSetRegion(region citrixorchestration.HypervisorResourceRefResponseModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.IsNull() || r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}
