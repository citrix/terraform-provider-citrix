// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GcpHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Region  types.String   `tfsdk:"region"`
	Vpc     types.String   `tfsdk:"vpc"`
	Subnets []types.String `tfsdk:"subnets"`
	/** GCP Resource Pool **/
	ProjectName types.String `tfsdk:"project_name"`
	SharedVpc   types.Bool   `tfsdk:"shared_vpc"`
}

func (r GcpHypervisorResourcePoolResourceModel) RefreshPropertyValues(resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) GcpHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	region := resourcePool.GetRegion()
	if r.shouldSetRegion(region) {
		r.Region = types.StringValue(region.GetName())
	}
	project := resourcePool.GetProject()
	r.ProjectName = types.StringValue(project.GetName())
	vpc := resourcePool.GetVirtualPrivateCloud()
	r.Vpc = types.StringValue(vpc.GetName())
	var res []string
	for _, model := range resourcePool.GetNetworks() {
		res = append(res, model.GetName())
	}
	r.Subnets = util.ConvertPrimitiveStringArrayToBaseStringArray(res)

	vpcType := vpc.GetObjectTypeName()
	if vpcType == "sharedvirtualprivatecloud" {
		r.SharedVpc = types.BoolValue(true)
	}

	return r
}

func (r GcpHypervisorResourcePoolResourceModel) shouldSetRegion(region citrixorchestration.HypervisorResourceRefResponseModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.IsNull() || r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}
