// Copyright © 2023. Citrix Systems, Inc.

package hypervisor

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type AwsHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Zone types.String `tfsdk:"zone"`
	/** AWS EC2 Connection **/
	Region    types.String `tfsdk:"region"`
	ApiKey    types.String `tfsdk:"api_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func (r AwsHypervisorResourceModel) RefreshPropertyValues(hypervisor *citrixorchestration.HypervisorDetailResponseModel) AwsHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.Region = types.StringValue(hypervisor.GetRegion())
	r.ApiKey = types.StringValue(hypervisor.GetApiKey())

	return r
}
