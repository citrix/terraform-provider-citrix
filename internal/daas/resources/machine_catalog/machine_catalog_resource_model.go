// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"golang.org/x/exp/slices"
)

// MachineCatalogResourceModel maps the resource schema data.
type MachineCatalogResourceModel struct {
	Id                 types.String             `tfsdk:"id"`
	Name               types.String             `tfsdk:"name"`
	Description        types.String             `tfsdk:"description"`
	IsPowerManaged     types.Bool               `tfsdk:"is_power_managed"`
	IsRemotePc         types.Bool               `tfsdk:"is_remote_pc"`
	AllocationType     types.String             `tfsdk:"allocation_type"`
	SessionSupport     types.String             `tfsdk:"session_support"`
	Zone               types.String             `tfsdk:"zone"`
	VdaUpgradeType     types.String             `tfsdk:"vda_upgrade_type"`
	ProvisioningType   types.String             `tfsdk:"provisioning_type"`
	ProvisioningScheme *ProvisioningSchemeModel `tfsdk:"provisioning_scheme"`
	MachineAccounts    []MachineAccountsModel   `tfsdk:"machine_accounts"`
	RemotePcOus        []RemotePcOuModel        `tfsdk:"remote_pc_ous"`
}

type MachineAccountsModel struct {
	Hypervisor types.String                 `tfsdk:"hypervisor"`
	Machines   []MachineCatalogMachineModel `tfsdk:"machines"`
}

type MachineCatalogMachineModel struct {
	MachineName       types.String `tfsdk:"machine_name"`
	Region            types.String `tfsdk:"region"`
	ResourceGroupName types.String `tfsdk:"resource_group_name"`
	ProjectName       types.String `tfsdk:"project_name"`
	AvailabilityZone  types.String `tfsdk:"availability_zone"`
}

// ProvisioningSchemeModel maps the nested provisioning scheme resource schema data.
type ProvisioningSchemeModel struct {
	Hypervisor                  types.String                      `tfsdk:"hypervisor"`
	HypervisorResourcePool      types.String                      `tfsdk:"hypervisor_resource_pool"`
	AzureMachineConfig          *AzureMachineConfigModel          `tfsdk:"azure_machine_config"`
	AwsMachineConfig            *AwsMachineConfigModel            `tfsdk:"aws_machine_config"`
	GcpMachineConfig            *GcpMachineConfigModel            `tfsdk:"gcp_machine_config"`
	NumTotalMachines            types.Int64                       `tfsdk:"number_of_total_machines"`
	NetworkMapping              *NetworkMappingModel              `tfsdk:"network_mapping"`
	AvailabilityZones           types.String                      `tfsdk:"availability_zones"`
	IdentityType                types.String                      `tfsdk:"identity_type"`
	MachineDomainIdentity       *MachineDomainIdentityModel       `tfsdk:"machine_domain_identity"`
	MachineAccountCreationRules *MachineAccountCreationRulesModel `tfsdk:"machine_account_creation_rules"`
}

type MachineProfileModel struct {
	MachineProfileVmName        types.String `tfsdk:"machine_profile_vm_name"`
	MachineProfileResourceGroup types.String `tfsdk:"machine_profile_resource_group"`
}

type MachineDomainIdentityModel struct {
	Domain                 types.String `tfsdk:"domain"`
	Ou                     types.String `tfsdk:"domain_ou"`
	ServiceAccount         types.String `tfsdk:"service_account"`
	ServiceAccountPassword types.String `tfsdk:"service_account_password"`
}

type GalleryImageModel struct {
	Gallery    types.String `tfsdk:"gallery"`
	Definition types.String `tfsdk:"definition"`
	Version    types.String `tfsdk:"version"`
}

// WritebackCacheModel maps the write back cacheconfiguration schema data.
type WritebackCacheModel struct {
	PersistWBC                 types.Bool   `tfsdk:"persist_wbc"`
	WBCDiskStorageType         types.String `tfsdk:"wbc_disk_storage_type"`
	PersistOsDisk              types.Bool   `tfsdk:"persist_os_disk"`
	PersistVm                  types.Bool   `tfsdk:"persist_vm"`
	StorageCostSaving          types.Bool   `tfsdk:"storage_cost_saving"`
	WriteBackCacheDiskSizeGB   types.Int64  `tfsdk:"writeback_cache_disk_size_gb"`
	WriteBackCacheMemorySizeMB types.Int64  `tfsdk:"writeback_cache_memory_size_mb"`
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

func (r MachineCatalogResourceModel) RefreshPropertyValues(ctx context.Context, client *citrixclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, machines *citrixorchestration.MachineResponseModelCollection) MachineCatalogResourceModel {
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

	if r.ProvisioningScheme == nil {
		r.ProvisioningScheme = &ProvisioningSchemeModel{}
	}

	provScheme := catalog.GetProvisioningScheme()
	resourcePool := provScheme.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	machineAccountCreateRules := provScheme.GetMachineAccountCreationRules()
	domain := machineAccountCreateRules.GetDomain()
	customProperties := provScheme.GetCustomProperties()

	// Refresh Hypervisor and Resource Pool
	r.ProvisioningScheme.Hypervisor = types.StringValue(hypervisor.GetId())
	r.ProvisioningScheme.HypervisorResourcePool = types.StringValue(resourcePool.GetId())

	switch *connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		if r.ProvisioningScheme.AzureMachineConfig == nil {
			r.ProvisioningScheme.AzureMachineConfig = &AzureMachineConfigModel{}
		}

		r.ProvisioningScheme.AzureMachineConfig.RefreshProperties(*catalog)

		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" && !r.ProvisioningScheme.AvailabilityZones.IsNull() {
				r.ProvisioningScheme.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if r.ProvisioningScheme.AwsMachineConfig == nil {
			r.ProvisioningScheme.AwsMachineConfig = &AwsMachineConfigModel{}
		}
		r.ProvisioningScheme.AwsMachineConfig.RefreshProperties(*catalog)

		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" {
				r.ProvisioningScheme.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if r.ProvisioningScheme.GcpMachineConfig == nil {
			r.ProvisioningScheme.GcpMachineConfig = &GcpMachineConfigModel{}
		}

		r.ProvisioningScheme.GcpMachineConfig.RefreshProperties(*catalog)

		for _, stringPair := range customProperties {
			if stringPair.GetName() == "CatalogZones" && !r.ProvisioningScheme.AvailabilityZones.IsNull() {
				r.ProvisioningScheme.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	}

	// Refresh Total Machine Count
	r.ProvisioningScheme.NumTotalMachines = types.Int64Value(int64(provScheme.GetMachineCount()))

	// Refresh Total Machine Count
	if identityType := types.StringValue(reflect.ValueOf(provScheme.GetIdentityType()).String()); identityType.ValueString() != "" {
		r.ProvisioningScheme.IdentityType = identityType
	} else {
		r.ProvisioningScheme.IdentityType = types.StringNull()
	}

	// Refresh Network Mapping
	networkMaps := provScheme.GetNetworkMaps()

	if len(networkMaps) > 0 && r.ProvisioningScheme.NetworkMapping != nil {
		r.ProvisioningScheme.NetworkMapping = &NetworkMappingModel{}
		r.ProvisioningScheme.NetworkMapping.NetworkDevice = types.StringValue(networkMaps[0].GetDeviceId())
		network := networkMaps[0].GetNetwork()
		segments := strings.Split(network.GetXDPath(), "\\")
		lastIndex := len(segments)
		r.ProvisioningScheme.NetworkMapping.Network = types.StringValue(strings.Split((strings.Split(segments[lastIndex-1], "."))[0], " ")[0])
	} else {
		r.ProvisioningScheme.NetworkMapping = nil
	}

	// Identity Pool Properties
	if r.ProvisioningScheme.MachineAccountCreationRules == nil {
		r.ProvisioningScheme.MachineAccountCreationRules = &MachineAccountCreationRulesModel{}
	}
	r.ProvisioningScheme.MachineAccountCreationRules.NamingScheme = types.StringValue(machineAccountCreateRules.GetNamingScheme())
	namingSchemeType := machineAccountCreateRules.GetNamingSchemeType()
	r.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType = types.StringValue(reflect.ValueOf(namingSchemeType).String())

	// Domain Identity Properties
	if r.ProvisioningScheme.MachineDomainIdentity == nil {
		r.ProvisioningScheme.MachineDomainIdentity = &MachineDomainIdentityModel{}
	}

	if domain.GetName() != "" {
		r.ProvisioningScheme.MachineDomainIdentity.Domain = types.StringValue(domain.GetName())
	}
	if machineAccountCreateRules.GetOU() != "" {
		r.ProvisioningScheme.MachineDomainIdentity.Ou = types.StringValue(machineAccountCreateRules.GetOU())
	}

	return r
}

func allocationTypeEnumToString(conn citrixorchestration.AllocationType) string {
	switch conn {
	case citrixorchestration.ALLOCATIONTYPE_UNKNOWN:
		return "Unknown"
	case citrixorchestration.ALLOCATIONTYPE_RANDOM:
		return "Random"
	case citrixorchestration.ALLOCATIONTYPE_STATIC:
		return "Static"
	default:
		return ""
	}
}

func ParseCustomPropertiesToClientModel(provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType) []citrixorchestration.NameValueStringPairModel {
	var res = &[]citrixorchestration.NameValueStringPairModel{}
	switch connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
		} else {
			util.AppendNameValueStringPair(res, "Zones", "")
		}
		if !provisioningScheme.AzureMachineConfig.StorageType.IsNull() {
			util.AppendNameValueStringPair(res, "StorageType", provisioningScheme.AzureMachineConfig.StorageType.ValueString())
		}
		if !provisioningScheme.AzureMachineConfig.VdaResourceGroup.IsNull() {
			util.AppendNameValueStringPair(res, "ResourceGroups", provisioningScheme.AzureMachineConfig.VdaResourceGroup.ValueString())
		}
		if !provisioningScheme.AzureMachineConfig.UseManagedDisks.IsNull() {
			if provisioningScheme.AzureMachineConfig.UseManagedDisks.ValueBool() {
				util.AppendNameValueStringPair(res, "UseManagedDisks", "true")
			} else {
				util.AppendNameValueStringPair(res, "UseManagedDisks", "false")
			}
		}
		if provisioningScheme.AzureMachineConfig.WritebackCache != nil {
			if !provisioningScheme.AzureMachineConfig.WritebackCache.WBCDiskStorageType.IsNull() {
				util.AppendNameValueStringPair(res, "WBCDiskStorageType", provisioningScheme.AzureMachineConfig.WritebackCache.WBCDiskStorageType.ValueString())
			}
			if provisioningScheme.AzureMachineConfig.WritebackCache.PersistWBC.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistWBC", "true")
				if provisioningScheme.AzureMachineConfig.WritebackCache.StorageCostSaving.ValueBool() {
					util.AppendNameValueStringPair(res, "StorageTypeAtShutdown", "Standard_LRS")
				}
			}
			if provisioningScheme.AzureMachineConfig.WritebackCache.PersistOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistOsDisk", "true")
				if provisioningScheme.AzureMachineConfig.WritebackCache.PersistVm.ValueBool() {
					util.AppendNameValueStringPair(res, "PersistVm", "true")
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "CatalogZones", provisioningScheme.AvailabilityZones.ValueString())
		}
		if !provisioningScheme.GcpMachineConfig.StorageType.IsNull() {
			util.AppendNameValueStringPair(res, "StorageType", provisioningScheme.GcpMachineConfig.StorageType.ValueString())
		}
		if provisioningScheme.GcpMachineConfig.WritebackCache != nil {
			if !provisioningScheme.GcpMachineConfig.WritebackCache.WBCDiskStorageType.IsNull() {
				util.AppendNameValueStringPair(res, "WBCDiskStorageType", provisioningScheme.GcpMachineConfig.WritebackCache.WBCDiskStorageType.ValueString())
			}
			if provisioningScheme.GcpMachineConfig.WritebackCache.PersistWBC.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistWBC", "true")
			}
			if provisioningScheme.GcpMachineConfig.WritebackCache.PersistOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistOsDisk", "true")
			}
		}
	}

	return *res
}

func ParseNetworkMappingToClientModel(networkMapping NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) ([]citrixorchestration.NetworkMapRequestModel, error) {
	var networks []citrixorchestration.HypervisorResourceRefResponseModel
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		networks = resourcePool.Subnets
	} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS || resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM {
		networks = resourcePool.Networks
	}

	var res = []citrixorchestration.NetworkMapRequestModel{}
	var networkName string
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM || resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM {
		networkName = networkMapping.Network.ValueString()
	} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
		networkName = fmt.Sprintf("%s (%s)", networkMapping.Network.ValueString(), resourcePool.GetResourcePoolRootId())
	}
	network := slices.IndexFunc(networks, func(c citrixorchestration.HypervisorResourceRefResponseModel) bool { return c.GetName() == networkName })
	if network == -1 {
		return res, fmt.Errorf("network %s not found", networkName)
	}

	res = append(res, citrixorchestration.NetworkMapRequestModel{
		NetworkDeviceNameOrId: *citrixorchestration.NewNullableString(networkMapping.NetworkDevice.ValueStringPointer()),
		NetworkPath:           networks[network].GetXDPath(),
	})
	return res, nil
}

func (r MachineCatalogResourceModel) updateCatalogWithMachines(ctx context.Context, client *citrixclient.CitrixDaasClient, machines *citrixorchestration.MachineResponseModelCollection) MachineCatalogResourceModel {
	if machines == nil {
		r.MachineAccounts = nil
		return r
	}

	machineMapFromRemote := map[string]citrixorchestration.MachineResponseModel{}
	for _, machine := range machines.GetItems() {
		machineMapFromRemote[strings.ToLower(machine.GetName())] = machine
	}

	if r.MachineAccounts != nil {
		machinesNotPresetInRemote := map[string]bool{}
		for _, machineAccount := range r.MachineAccounts {
			for _, machineFromPlan := range machineAccount.Machines {
				machineFromPlanName := machineFromPlan.MachineName.ValueString()
				machineFromRemote, exists := machineMapFromRemote[strings.ToLower(machineFromPlanName)]
				if !exists {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				hosting := machineFromRemote.GetHosting()
				hypervisor := hosting.GetHypervisorConnection()
				hypervisorId := hypervisor.GetId()

				if !strings.EqualFold(hypervisorId, machineAccount.Hypervisor.ValueString()) {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				if hypervisorId == "" {
					delete(machineMapFromRemote, strings.ToLower(machineFromPlanName))
					continue
				}

				hyp, err := util.GetHypervisor(ctx, client, nil, hypervisorId)
				if err != nil {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				connectionType := hyp.GetConnectionType()
				hostedMachineId := hosting.GetHostedMachineId()
				switch connectionType {
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
					if hostedMachineId != "" {
						resourceGroupName := strings.Split(hostedMachineId, "/")[0] // hosted machine id is resourcegroupname/vmname
						if !strings.EqualFold(machineFromPlan.ResourceGroupName.ValueString(), resourceGroupName) {
							machineFromPlan.ResourceGroupName = types.StringValue(resourceGroupName)
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
					if hostedMachineId != "" {
						machineIdArray := strings.Split(hostedMachineId, ":") // hosted machine id is projectname:region:vmname
						if !strings.EqualFold(machineFromPlan.Region.ValueString(), machineIdArray[1]) {
							machineFromPlan.Region = types.StringValue(machineIdArray[1])
						}
					}
					// case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS: AvailabilityZone is not available from remote
				}

				delete(machineMapFromRemote, strings.ToLower(machineFromPlanName))
			}
		}

		machineAccounts := []MachineAccountsModel{}
		for _, machineAccount := range r.MachineAccounts {
			machines := []MachineCatalogMachineModel{}
			for _, machine := range machineAccount.Machines {
				if machinesNotPresetInRemote[strings.ToLower(machine.MachineName.ValueString())] {
					continue
				}
				machines = append(machines, machine)
			}
			machineAccount.Machines = machines
			machineAccounts = append(machineAccounts, machineAccount)
		}

		r.MachineAccounts = machineAccounts
	}

	// go over any machines that are in remote but were not in plan
	newMachines := map[string][]MachineCatalogMachineModel{}
	for machineName, machineFromRemote := range machineMapFromRemote {
		hosting := machineFromRemote.GetHosting()
		hypConnection := hosting.GetHypervisorConnection()
		hypId := hypConnection.GetId()

		var machineModel MachineCatalogMachineModel
		machineModel.MachineName = types.StringValue(machineName)

		if hypId != "" {
			hyp, err := util.GetHypervisor(ctx, client, nil, hypId)
			if err != nil {
				continue
			}

			connectionType := hyp.GetConnectionType()
			hostedMachineId := hosting.GetHostedMachineId()
			switch connectionType {
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
				if hostedMachineId != "" {
					resourceGroupName := strings.Split(hostedMachineId, "/")[0] // hosted machine id is resourcegroupname/vmname
					machineModel.ResourceGroupName = types.StringValue(resourceGroupName)
					// region is not available from remote
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
				if hostedMachineId != "" {
					machineIdArray := strings.Split(hostedMachineId, ":") // hosted machine id is projectname:region:vmname
					machineModel.ProjectName = types.StringValue(machineIdArray[0])
					machineModel.Region = types.StringValue(machineIdArray[1])
				}
				// case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS: AvailabilityZone is not available from remote
			}
		}

		_, exists := newMachines[hypId]
		if !exists {
			newMachines[hypId] = []MachineCatalogMachineModel{}
		}

		newMachines[hypId] = append(newMachines[hypId], machineModel)
	}

	if len(newMachines) > 0 && r.MachineAccounts == nil {
		r.MachineAccounts = []MachineAccountsModel{}
	}

	machineAccountMap := map[string]int{}
	for index, machineAccount := range r.MachineAccounts {
		machineAccountMap[machineAccount.Hypervisor.ValueString()] = index
	}

	for hypId, machines := range newMachines {
		machineAccIndex, exists := machineAccountMap[hypId]
		if exists {
			machAccounts := r.MachineAccounts
			machineAccount := machAccounts[machineAccIndex]
			if machineAccount.Machines == nil {
				machineAccount.Machines = []MachineCatalogMachineModel{}
			}
			machineAccountMachines := machineAccount.Machines
			machineAccountMachines = append(machineAccountMachines, machines...)
			machineAccount.Machines = machineAccountMachines
			machAccounts[machineAccIndex] = machineAccount
			r.MachineAccounts = machAccounts
			continue
		}
		var machineAccount MachineAccountsModel
		machineAccount.Hypervisor = types.StringValue(hypId)
		machineAccount.Machines = machines
		machAccounts := r.MachineAccounts
		machAccounts = append(machAccounts, machineAccount)
		machineAccountMap[hypId] = len(machAccounts) - 1
		r.MachineAccounts = machAccounts
	}

	return r
}

func (r MachineCatalogResourceModel) updateCatalogWithRemotePcConfig(catalog *citrixorchestration.MachineCatalogDetailResponseModel) MachineCatalogResourceModel {
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL || !r.IsRemotePc.IsNull() {
		r.IsRemotePc = types.BoolValue(catalog.GetIsRemotePC())
	}
	rpcOUs := util.RefreshListProperties[RemotePcOuModel, citrixorchestration.RemotePCEnrollmentScopeResponseModel](r.RemotePcOus, "OUName", catalog.GetRemotePCEnrollmentScopes(), "OU", "RefreshListItem")
	r.RemotePcOus = rpcOUs
	return r
}

func (scope RemotePcOuModel) RefreshListItem(remote citrixorchestration.RemotePCEnrollmentScopeResponseModel) RemotePcOuModel {
	scope.OUName = types.StringValue(remote.GetOU())
	scope.IncludeSubFolders = types.BoolValue(remote.GetIncludeSubfolders())

	return scope
}
