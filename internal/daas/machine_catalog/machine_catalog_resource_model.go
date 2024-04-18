// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"reflect"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MachineCatalogResourceModel maps the resource schema data.
type MachineCatalogResourceModel struct {
	Id                     types.String             `tfsdk:"id"`
	Name                   types.String             `tfsdk:"name"`
	Description            types.String             `tfsdk:"description"`
	IsPowerManaged         types.Bool               `tfsdk:"is_power_managed"`
	IsRemotePc             types.Bool               `tfsdk:"is_remote_pc"`
	AllocationType         types.String             `tfsdk:"allocation_type"`
	SessionSupport         types.String             `tfsdk:"session_support"`
	Zone                   types.String             `tfsdk:"zone"`
	VdaUpgradeType         types.String             `tfsdk:"vda_upgrade_type"`
	ProvisioningType       types.String             `tfsdk:"provisioning_type"`
	ProvisioningScheme     *ProvisioningSchemeModel `tfsdk:"provisioning_scheme"`
	MachineAccounts        []MachineAccountsModel   `tfsdk:"machine_accounts"`
	RemotePcOus            []RemotePcOuModel        `tfsdk:"remote_pc_ous"`
	MinimumFunctionalLevel types.String             `tfsdk:"minimum_functional_level"`
}

type MachineAccountsModel struct {
	Hypervisor types.String                 `tfsdk:"hypervisor"`
	Machines   []MachineCatalogMachineModel `tfsdk:"machines"`
}

type MachineCatalogMachineModel struct {
	MachineAccount    types.String `tfsdk:"machine_account"`
	MachineName       types.String `tfsdk:"machine_name"`
	Region            types.String `tfsdk:"region"`
	ResourceGroupName types.String `tfsdk:"resource_group_name"`
	ProjectName       types.String `tfsdk:"project_name"`
	AvailabilityZone  types.String `tfsdk:"availability_zone"`
	Datacenter        types.String `tfsdk:"datacenter"`
	Cluster           types.String `tfsdk:"cluster"`
	Host              types.String `tfsdk:"host"`
}

// ProvisioningSchemeModel maps the nested provisioning scheme resource schema data.
type ProvisioningSchemeModel struct {
	Hypervisor                  types.String                      `tfsdk:"hypervisor"`
	HypervisorResourcePool      types.String                      `tfsdk:"hypervisor_resource_pool"`
	AzureMachineConfig          *AzureMachineConfigModel          `tfsdk:"azure_machine_config"`
	AwsMachineConfig            *AwsMachineConfigModel            `tfsdk:"aws_machine_config"`
	GcpMachineConfig            *GcpMachineConfigModel            `tfsdk:"gcp_machine_config"`
	VsphereMachineConfig        *VsphereMachineConfigModel        `tfsdk:"vsphere_machine_config"`
	XenserverMachineConfig      *XenserverMachineConfigModel      `tfsdk:"xenserver_machine_config"`
	NutanixMachineConfigModel   *NutanixMachineConfigModel        `tfsdk:"nutanix_machine_config"`
	NumTotalMachines            types.Int64                       `tfsdk:"number_of_total_machines"`
	NetworkMapping              *NetworkMappingModel              `tfsdk:"network_mapping"`
	AvailabilityZones           types.String                      `tfsdk:"availability_zones"`
	IdentityType                types.String                      `tfsdk:"identity_type"`
	MachineDomainIdentity       *MachineDomainIdentityModel       `tfsdk:"machine_domain_identity"`
	MachineAccountCreationRules *MachineAccountCreationRulesModel `tfsdk:"machine_account_creation_rules"`
	CustomProperties            []CustomPropertyModel             `tfsdk:"custom_properties"`
}

type CustomPropertyModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type MachineDomainIdentityModel struct {
	Domain                 types.String `tfsdk:"domain"`
	Ou                     types.String `tfsdk:"domain_ou"`
	ServiceAccount         types.String `tfsdk:"service_account"`
	ServiceAccountPassword types.String `tfsdk:"service_account_password"`
}

// MachineAccountCreationRulesModel maps the nested machine account creation rules resource schema data.
type MachineAccountCreationRulesModel struct {
	NamingScheme     types.String `tfsdk:"naming_scheme"`
	NamingSchemeType types.String `tfsdk:"naming_scheme_type"`
}

// NetworkMappingModel maps the nested network mapping resource schema data.
type NetworkMappingModel struct {
	NetworkDevice types.String `tfsdk:"network_device"`
	Network       types.String `tfsdk:"network"`
}

type RemotePcOuModel struct {
	IncludeSubFolders types.Bool   `tfsdk:"include_subfolders"`
	OUName            types.String `tfsdk:"ou_name"`
}

func (r MachineCatalogResourceModel) RefreshPropertyValues(ctx context.Context, client *citrixclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, machines *citrixorchestration.MachineResponseModelCollection, pluginId string) MachineCatalogResourceModel {
	// Machine Catalog Properties
	r.Id = types.StringValue(catalog.GetId())
	r.Name = types.StringValue(catalog.GetName())
	if catalog.GetDescription() != "" {
		r.Description = types.StringValue(catalog.GetDescription())
	} else {
		r.Description = types.StringNull()
	}
	allocationType := catalog.GetAllocationType()
	r.AllocationType = types.StringValue(allocationTypeEnumToString(allocationType))
	sessionSupport := catalog.GetSessionSupport()
	r.SessionSupport = types.StringValue(reflect.ValueOf(sessionSupport).String())

	minimumFunctionalLevel := catalog.GetMinimumFunctionalLevel()
	r.MinimumFunctionalLevel = types.StringValue(string(minimumFunctionalLevel))

	catalogZone := catalog.GetZone()
	r.Zone = types.StringValue(catalogZone.GetId())

	if catalog.UpgradeInfo != nil {
		if *catalog.UpgradeInfo.UpgradeType != citrixorchestration.VDAUPGRADETYPE_NOT_SET || !r.VdaUpgradeType.IsNull() {
			r.VdaUpgradeType = types.StringValue(string(*catalog.UpgradeInfo.UpgradeType))
		}
	} else {
		r.VdaUpgradeType = types.StringNull()
	}

	provtype := catalog.GetProvisioningType()
	r.ProvisioningType = types.StringValue(string(provtype))
	if provtype == citrixorchestration.PROVISIONINGTYPE_MANUAL || !r.IsPowerManaged.IsNull() {
		r.IsPowerManaged = types.BoolValue(catalog.GetIsPowerManaged())
	}

	if catalog.ProvisioningType == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		// Handle machines
		r = r.updateCatalogWithMachines(ctx, client, machines)
	}

	r = r.updateCatalogWithRemotePcConfig(catalog)

	if catalog.ProvisioningScheme == nil {
		r.ProvisioningScheme = nil
		return r
	}

	// Provisioning Scheme Properties
	r = r.updateCatalogWithProvScheme(ctx, client, catalog, connectionType, pluginId)

	return r
}
