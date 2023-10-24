package models

import (
	"reflect"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type HypervisorResourceModel struct {
	/**** Connection Details ****/
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Zone           types.String `tfsdk:"zone"`
	/** Azure Connection **/
	ApplicationId                   types.String `tfsdk:"application_id"`
	ApplicationSecret               types.String `tfsdk:"application_secret"`
	ApplicationSecretExpirationDate types.String `tfsdk:"application_secret_expiration_date"`
	SubscriptionId                  types.String `tfsdk:"subscription_id"`
	ActiveDirectoryId               types.String `tfsdk:"active_directory_id"`
	/** AWS EC2 Connection **/
	AwsRegion types.String `tfsdk:"aws_region"`
	ApiKey    types.String `tfsdk:"api_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	/** GCP Connection **/
	ServiceAccountId          types.String `tfsdk:"service_account_id"`
	ServiceAccountCredentials types.String `tfsdk:"service_account_credentials"`
}

func (r HypervisorResourceModel) RefreshPropertyValues(hypervisor *citrixorchestration.HypervisorDetailResponseModel) HypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	connectionType := hypervisor.GetConnectionType()
	r.ConnectionType = types.StringValue(reflect.ValueOf(connectionType).String())
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		r.ApplicationId = types.StringValue(hypervisor.GetApplicationId())
		r.SubscriptionId = types.StringValue(hypervisor.GetSubscriptionId())
		r.ActiveDirectoryId = types.StringValue(hypervisor.GetActiveDirectoryId())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		r.AwsRegion = types.StringValue(hypervisor.GetRegion())
		r.ApiKey = types.StringValue(hypervisor.GetApiKey())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		r.ServiceAccountId = types.StringValue(hypervisor.GetServiceAccountId())
	}

	return r
}

func getResourceGroupNameFromVnetId(vnetId string) string {
	resourceGroupAndVnetName := strings.Split(vnetId, "/")
	return resourceGroupAndVnetName[0]
}
