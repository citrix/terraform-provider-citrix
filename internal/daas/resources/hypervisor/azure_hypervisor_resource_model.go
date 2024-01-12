// Copyright © 2023. Citrix Systems, Inc.

package hypervisor

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type AzureHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Zone types.String `tfsdk:"zone"`
	/** Azure Connection **/
	ApplicationId                   types.String `tfsdk:"application_id"`
	ApplicationSecret               types.String `tfsdk:"application_secret"`
	ApplicationSecretExpirationDate types.String `tfsdk:"application_secret_expiration_date"`
	SubscriptionId                  types.String `tfsdk:"subscription_id"`
	ActiveDirectoryId               types.String `tfsdk:"active_directory_id"`
}

func (r AzureHypervisorResourceModel) RefreshPropertyValues(hypervisor *citrixorchestration.HypervisorDetailResponseModel) AzureHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.ApplicationId = types.StringValue(hypervisor.GetApplicationId())
	r.SubscriptionId = types.StringValue(hypervisor.GetSubscriptionId())
	r.ActiveDirectoryId = types.StringValue(hypervisor.GetActiveDirectoryId())

	return r
}
