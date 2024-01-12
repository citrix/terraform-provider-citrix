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
	MachineAccounts    *[]MachineAccountsModel  `tfsdk:"machine_accounts"`
	RemotePcOus        *[]RemotePcOuModel       `tfsdk:"remote_pc_ous"`
}

type MachineAccountsModel struct {
	Hypervisor types.String                  `tfsdk:"hypervisor"`
	Machines   *[]MachineCatalogMachineModel `tfsdk:"machines"`
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
	MachineConfig               *MachineConfigModel               `tfsdk:"machine_config"`
	NumTotalMachines            types.Int64                       `tfsdk:"number_of_total_machines"`
	NetworkMapping              *NetworkMappingModel              `tfsdk:"network_mapping"`
	MachineAccountCreationRules *MachineAccountCreationRulesModel `tfsdk:"machine_account_creation_rules"`
	AvailabilityZones           types.String                      `tfsdk:"availability_zones"`
	StorageType                 types.String                      `tfsdk:"storage_type"`
	VdaResourceGroup            types.String                      `tfsdk:"vda_resource_group"`
	UseManagedDisks             types.Bool                        `tfsdk:"use_managed_disks"`
	WritebackCache              *WritebackCacheModel              `tfsdk:"writeback_cache"`
}

type MachineConfigModel struct {
	Hypervisor             types.String `tfsdk:"hypervisor"`
	HypervisorResourcePool types.String `tfsdk:"hypervisor_resource_pool"`
	ServiceAccount         types.String `tfsdk:"service_account"`
	ServiceAccountPassword types.String `tfsdk:"service_account_password"`
	ServiceOffering        types.String `tfsdk:"service_offering"`
	MasterImage            types.String `tfsdk:"master_image"`
	/** Azure Hypervisor **/
	ResourceGroup  types.String       `tfsdk:"resource_group"`
	StorageAccount types.String       `tfsdk:"storage_account"`
	Container      types.String       `tfsdk:"container"`
	GalleryImage   *GalleryImageModel `tfsdk:"gallery_image"`
	/** AWS Hypervisor **/
	ImageAmi types.String `tfsdk:"image_ami"`
	/** GCP Hypervisor **/
	MachineProfile  types.String `tfsdk:"machine_profile"`
	MachineSnapshot types.String `tfsdk:"machine_snapshot"`
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
	Domain           types.String `tfsdk:"domain"`
	Ou               types.String `tfsdk:"domain_ou"`
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

	r.IsPowerManaged = types.BoolValue(catalog.GetIsPowerManaged())
	r.ProvisioningType = types.StringValue(string(catalog.GetProvisioningType()))

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
		r.ProvisioningScheme.MachineConfig = &MachineConfigModel{}
	}

	provScheme := catalog.GetProvisioningScheme()
	resourcePool := provScheme.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	machineProfile := provScheme.GetMachineProfile()
	masterImage := provScheme.GetMasterImage()
	machineAccountCreateRules := provScheme.GetMachineAccountCreationRules()
	domain := machineAccountCreateRules.GetDomain()

	// Refresh Machine Config Properties
	if r.ProvisioningScheme.MachineConfig.Hypervisor == types.StringNull() {
		r.ProvisioningScheme.MachineConfig.Hypervisor = types.StringValue(hypervisor.GetId())
	}

	if r.ProvisioningScheme.MachineConfig.HypervisorResourcePool == types.StringNull() {
		r.ProvisioningScheme.MachineConfig.HypervisorResourcePool = types.StringValue(resourcePool.GetId())
	}
	if catalog.ProvisioningScheme.GetServiceOffering() != "" && r.ProvisioningScheme.MachineConfig.ServiceOffering.ValueString() != "" {
		r.ProvisioningScheme.MachineConfig.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	} else {
		// For GCP, service offering will be nil in plan / state. Skip service offering refresh for GCP catalog.
		r.ProvisioningScheme.MachineConfig.ServiceOffering = types.StringNull()
	}

	// Refresh Master Image
	if r.ProvisioningScheme.MachineConfig.GalleryImage != nil {
		/* For Azure Image Gallery image, the XDPath looks like:
		 * XDHyp:\\HostingUnits\\{resource pool}\\image.folder\\{resource group}.resourcegroup\\{gallery name}.gallery\\{image name}.imagedefinition\\{image version}.imageversion
		 * The Name property in MasterImage will be image version instead of image definition (name of the image)
		 */
		r.ProvisioningScheme.MachineConfig.GalleryImage.Version = types.StringValue(masterImage.GetName())
	} else if r.ProvisioningScheme.MachineConfig.MachineSnapshot.ValueString() != "" {
		/* For GCP snapshot image, the XDPath looks like:
		 * XDHyp:\\HostingUnits\\{resource pool}\\{VM name}.vm\\{VM snapshot name}.snapshot
		 * The Name property in MasterImage will be VM snapshot name instead of VM name
		 */
		r.ProvisioningScheme.MachineConfig.MachineSnapshot = types.StringValue(masterImage.GetName())
		masterImageXdPath := masterImage.GetXDPath()
		if masterImageXdPath != "" {
			segments := strings.Split(masterImage.GetXDPath(), "\\")
			lastIndex := len(segments)
			// Snapshot or Managed Disk
			r.ProvisioningScheme.MachineConfig.MasterImage = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
		}
	} else {
		r.ProvisioningScheme.MachineConfig.MasterImage = types.StringValue(masterImage.GetName())
	}

	if *connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		masterImageXdPath := masterImage.GetXDPath()
		if masterImageXdPath != "" {
			segments := strings.Split(masterImage.GetXDPath(), "\\")
			lastIndex := len(segments)
			if lastIndex == 8 {
				resourceTag := strings.Split(segments[lastIndex-1], ".")
				resourceType := resourceTag[len(resourceTag)-1]

				if resourceType == "vhd" {
					// VHD image
					r.ProvisioningScheme.MachineConfig.Container = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
					r.ProvisioningScheme.MachineConfig.StorageAccount = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
				} else if resourceType == "imageversion" {
					// Gallery image
					r.ProvisioningScheme.MachineConfig.GalleryImage.Definition = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
					r.ProvisioningScheme.MachineConfig.GalleryImage.Gallery = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
				}
				r.ProvisioningScheme.MachineConfig.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-4], ".")[0])
			} else {
				// Snapshot or Managed Disk
				r.ProvisioningScheme.MachineConfig.ResourceGroup = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
			}
		}
	}

	if machineProfileName := machineProfile.GetName(); machineProfileName != "" {
		r.ProvisioningScheme.MachineConfig.MachineProfile = types.StringValue(machineProfileName)
	} else {
		r.ProvisioningScheme.MachineConfig.MachineProfile = types.StringNull()
	}

	// Refresh Total Machine Count
	r.ProvisioningScheme.NumTotalMachines = types.Int64Value(int64(provScheme.GetMachineCount()))

	// Refresh Network Mapping
	networkMaps := catalog.ProvisioningScheme.GetNetworkMaps()

	if len(networkMaps) > 0 && r.ProvisioningScheme.NetworkMapping != nil {
		r.ProvisioningScheme.NetworkMapping.NetworkDevice = types.StringValue(networkMaps[0].GetDeviceId())
		network := networkMaps[0].GetNetwork()
		segments := strings.Split(network.GetXDPath(), "\\")
		lastIndex := len(segments)
		r.ProvisioningScheme.NetworkMapping.Network = types.StringValue(strings.Split((strings.Split(segments[lastIndex-1], "."))[0], " ")[0])
	} else {
		r.ProvisioningScheme.NetworkMapping = nil
	}

	//Refresh custom properties
	customProperties := provScheme.GetCustomProperties()
	r.ProvisioningScheme.RefreshProperties(customProperties)

	if r.ProvisioningScheme.WritebackCache != nil {
		r.ProvisioningScheme.WritebackCache.WriteBackCacheDiskSizeGB = types.Int64Value(int64(provScheme.GetWriteBackCacheDiskSizeGB()))
		if !r.ProvisioningScheme.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
			r.ProvisioningScheme.WritebackCache.WriteBackCacheMemorySizeMB = types.Int64Value(int64(provScheme.GetWriteBackCacheMemorySizeMB()))
		}
	}

	// Identity Pool Properties
	if r.ProvisioningScheme.MachineAccountCreationRules == nil {
		r.ProvisioningScheme.MachineAccountCreationRules = &MachineAccountCreationRulesModel{}
	}
	r.ProvisioningScheme.MachineAccountCreationRules.NamingScheme = types.StringValue(machineAccountCreateRules.GetNamingScheme())
	namingSchemeType := machineAccountCreateRules.GetNamingSchemeType()
	r.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType = types.StringValue(reflect.ValueOf(namingSchemeType).String())
	r.ProvisioningScheme.MachineAccountCreationRules.Domain = types.StringValue(domain.GetName())
	if machineAccountCreateRules.GetOU() != "" {
		r.ProvisioningScheme.MachineAccountCreationRules.Ou = types.StringValue(machineAccountCreateRules.GetOU())
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

func (res *ProvisioningSchemeModel) RefreshProperties(stringPairs []citrixorchestration.NameValueStringPairModel) {

	for _, stringPair := range stringPairs {
		switch stringPair.GetName() {
		case "StorageType":
			res.StorageType = types.StringValue(stringPair.GetValue())
		case "UseManagedDisks":
			res.UseManagedDisks = util.StringToTypeBool(stringPair.GetValue())
		case "Zones":
			res.AvailabilityZones = types.StringValue(stringPair.GetValue())
		case "CatalogZones":
			res.AvailabilityZones = types.StringValue(stringPair.GetValue())
		case "ResourceGroups":
			res.VdaResourceGroup = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			if res.WritebackCache != nil {
				res.WritebackCache.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
			}
		case "PersistWBC":
			if res.WritebackCache != nil {
				res.WritebackCache.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistOsDisk":
			if res.WritebackCache != nil {
				res.WritebackCache.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistVm":
			if res.WritebackCache != nil {
				res.WritebackCache.PersistVm = util.StringToTypeBool(stringPair.GetValue())
			}
		case "StorageTypeAtShutdown":
			if res.WritebackCache != nil {
				res.WritebackCache.StorageCostSaving = types.BoolValue(true)
			}
		default:
		}
	}
}

func ParseCustomPropertiesToClientModel(provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType) *[]citrixorchestration.NameValueStringPairModel {
	var res = &[]citrixorchestration.NameValueStringPairModel{}
	if !provisioningScheme.StorageType.IsNull() {
		util.AppendNameValueStringPair(res, "StorageType", provisioningScheme.StorageType.ValueString())
	}
	if !provisioningScheme.VdaResourceGroup.IsNull() {
		util.AppendNameValueStringPair(res, "ResourceGroups", provisioningScheme.VdaResourceGroup.ValueString())
	}
	if !provisioningScheme.UseManagedDisks.IsNull() {
		if provisioningScheme.UseManagedDisks.ValueBool() {
			util.AppendNameValueStringPair(res, "UseManagedDisks", "true")
		} else {
			util.AppendNameValueStringPair(res, "UseManagedDisks", "false")
		}
	}
	if !provisioningScheme.AvailabilityZones.IsNull() {
		if connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
			util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
		} else if connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM {
			util.AppendNameValueStringPair(res, "CatalogZones", provisioningScheme.AvailabilityZones.ValueString())
		}
	}
	if provisioningScheme.WritebackCache != nil {
		if !provisioningScheme.WritebackCache.WBCDiskStorageType.IsNull() {
			util.AppendNameValueStringPair(res, "WBCDiskStorageType", provisioningScheme.WritebackCache.WBCDiskStorageType.ValueString())
		}
		if provisioningScheme.WritebackCache.PersistWBC.ValueBool() {
			util.AppendNameValueStringPair(res, "PersistWBC", "true")
		}
		if provisioningScheme.WritebackCache.PersistOsDisk.ValueBool() {
			util.AppendNameValueStringPair(res, "PersistOsDisk", "true")
			if provisioningScheme.WritebackCache.PersistOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistVm", "true")
			}
			if provisioningScheme.WritebackCache.StorageCostSaving.ValueBool() {
				util.AppendNameValueStringPair(res, "StorageTypeAtShutdown", "Standard_LRS")
			}
		}
	}
	return res
}

func ParseNetworkMappingToClientModel(networkMapping NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) []citrixorchestration.NetworkMapRequestModel {
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
	res = append(res, citrixorchestration.NetworkMapRequestModel{
		NetworkDeviceNameOrId: *citrixorchestration.NewNullableString(networkMapping.NetworkDevice.ValueStringPointer()),
		NetworkPath:           networks[network].GetXDPath(),
	})
	return res
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
		for _, machineAccount := range *r.MachineAccounts {
			for _, machineFromPlan := range *machineAccount.Machines {
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
						if !strings.EqualFold(machineFromPlan.ProjectName.ValueString(), machineIdArray[0]) {
							machineFromPlan.ProjectName = types.StringValue(machineIdArray[0])
						}
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
		for _, machineAccount := range *r.MachineAccounts {
			machines := []MachineCatalogMachineModel{}
			for _, machine := range *machineAccount.Machines {
				if machinesNotPresetInRemote[strings.ToLower(machine.MachineName.ValueString())] {
					continue
				}
				machines = append(machines, machine)
			}
			machineAccount.Machines = &machines
			machineAccounts = append(machineAccounts, machineAccount)
		}

		r.MachineAccounts = &machineAccounts
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
		r.MachineAccounts = &[]MachineAccountsModel{}
	}

	machineAccountMap := map[string]int{}
	for index, machineAccount := range *r.MachineAccounts {
		machineAccountMap[machineAccount.Hypervisor.ValueString()] = index
	}

	for hypId, machines := range newMachines {
		machineAccIndex, exists := machineAccountMap[hypId]
		if exists {
			machAccounts := *r.MachineAccounts
			machineAccount := machAccounts[machineAccIndex]
			if machineAccount.Machines == nil {
				machineAccount.Machines = &[]MachineCatalogMachineModel{}
			}
			machineAccountMachines := *machineAccount.Machines
			machineAccountMachines = append(machineAccountMachines, machines...)
			machineAccount.Machines = &machineAccountMachines
			machAccounts[machineAccIndex] = machineAccount
			r.MachineAccounts = &machAccounts
			continue
		}
		var machineAccount MachineAccountsModel
		machineAccount.Hypervisor = types.StringValue(hypId)
		machineAccount.Machines = &machines
		machAccounts := *r.MachineAccounts
		machAccounts = append(machAccounts, machineAccount)
		machineAccountMap[hypId] = len(machAccounts) - 1
		r.MachineAccounts = &machAccounts
	}

	return r
}

func (r MachineCatalogResourceModel) updateCatalogWithRemotePcConfig(catalog *citrixorchestration.MachineCatalogDetailResponseModel) MachineCatalogResourceModel {
	r.IsRemotePc = types.BoolValue(catalog.GetIsRemotePC())
	remotePcEnrollemntScopes := catalog.GetRemotePCEnrollmentScopes()
	if len(remotePcEnrollemntScopes) == 0 {
		if r.RemotePcOus != nil && len(*r.RemotePcOus) > 0 {
			r.RemotePcOus = nil
		}
		return r
	}

	if r.RemotePcOus == nil {
		r.RemotePcOus = &[]RemotePcOuModel{}
	}

	existingEnrollmentScopes := map[string]int{}
	for index, ou := range *r.RemotePcOus {
		existingEnrollmentScopes[ou.OUName.ValueString()] = index
	}

	remotePCOUs := *r.RemotePcOus
	for _, enrollmentScope := range catalog.GetRemotePCEnrollmentScopes() {
		ouName := enrollmentScope.GetOU()
		index, exists := existingEnrollmentScopes[ouName]
		if exists {
			existingEnrollmentScope := remotePCOUs[index]
			existingEnrollmentScope.IncludeSubFolders = types.BoolValue(enrollmentScope.GetIncludeSubfolders())
			remotePCOUs[index] = existingEnrollmentScope
		} else {
			var remotePCOU RemotePcOuModel
			remotePCOU.IncludeSubFolders = types.BoolValue(enrollmentScope.GetIncludeSubfolders())
			remotePCOU.OUName = types.StringValue(ouName)
			remotePCOUs = append(remotePCOUs, remotePCOU)
		}

		existingEnrollmentScopes[ouName] = -1 // Mark as visited. The ones not visited should be removed.
	}
	r.RemotePcOus = &remotePCOUs

	rpcOUs := []RemotePcOuModel{}
	for _, ou := range *r.RemotePcOus {
		if existingEnrollmentScopes[ou.OUName.ValueString()] == -1 {
			rpcOUs = append(rpcOUs, ou) // if visited, include. Not visited ones will not be included.
		}
	}

	r.RemotePcOus = &rpcOUs
	return r
}
