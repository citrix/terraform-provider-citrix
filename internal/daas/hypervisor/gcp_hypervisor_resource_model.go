// Copyright © 2023. Citrix Systems, Inc.

package hypervisor

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type GcpHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Zone types.String `tfsdk:"zone"`
	/** GCP Connection **/
	ServiceAccountId          types.String `tfsdk:"service_account_id"`
	ServiceAccountCredentials types.String `tfsdk:"service_account_credentials"`
}

func (r GcpHypervisorResourceModel) RefreshPropertyValues(hypervisor *citrixorchestration.HypervisorDetailResponseModel) GcpHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.ServiceAccountId = types.StringValue(hypervisor.GetServiceAccountId())

	return r
}
