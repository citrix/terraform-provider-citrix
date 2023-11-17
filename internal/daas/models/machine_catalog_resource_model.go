package models

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"golang.org/x/exp/slices"
)

// MachineCatalogResourceModel maps the resource schema data.
type MachineCatalogResourceModel struct {
	Id                     types.String             `tfsdk:"id"`
	Name                   types.String             `tfsdk:"name"`
	Description            types.String             `tfsdk:"description"`
	ServiceAccount         types.String             `tfsdk:"service_account"`
	ServiceAccountPassword types.String             `tfsdk:"service_account_password"`
	AllocationType         types.String             `tfsdk:"allocation_type"`
	SessionSupport         types.String             `tfsdk:"session_support"`
	Zone                   types.String             `tfsdk:"zone"`
	VdaUpgradeType         types.String             `tfsdk:"vda_upgrade_type"`
	ProvisioningScheme     *ProvisioningSchemeModel `tfsdk:"provisioning_scheme"`
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
	ServiceOffering        types.String `tfsdk:"service_offering"`
	MasterImage            types.String `tfsdk:"master_image"`
	/** Azure Hypervisor **/
	ResourceGroup  types.String `tfsdk:"resource_group"`
	StorageAccount types.String `tfsdk:"storage_account"`
	Container      types.String `tfsdk:"container"`
	/** AWS Hypervisor **/
	ImageAmi types.String `tfsdk:"image_ami"`
	/** GCP Hypervisor **/
	MachineProfile types.String `tfsdk:"machine_profile"`
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

func (r MachineCatalogResourceModel) RefreshPropertyValues(ctx context.Context, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType citrixorchestration.HypervisorConnectionType) MachineCatalogResourceModel {
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
		r.VdaUpgradeType = types.StringValue(string(*catalog.UpgradeInfo.UpgradeType))
	} else {
		r.VdaUpgradeType = types.StringValue(string(citrixorchestration.VDAUPGRADETYPE_NOT_SET))
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
	if catalog.ProvisioningScheme.GetServiceOffering() != "" {
		r.ProvisioningScheme.MachineConfig.ServiceOffering = types.StringValue(provScheme.GetServiceOffering())
	} else {
		r.ProvisioningScheme.MachineConfig.ServiceOffering = types.StringNull()
	}

	// Refresh Master Image
	r.ProvisioningScheme.MachineConfig.MasterImage = types.StringValue(masterImage.GetName())
	if connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		masterImageXdPath := masterImage.GetXDPath()
		if masterImageXdPath != "" {
			segments := strings.Split(masterImage.GetXDPath(), "\\")
			lastIndex := len(segments)
			if lastIndex == 8 {
				// VHD image
				r.ProvisioningScheme.MachineConfig.Container = types.StringValue(strings.Split(segments[lastIndex-2], ".")[0])
				r.ProvisioningScheme.MachineConfig.StorageAccount = types.StringValue(strings.Split(segments[lastIndex-3], ".")[0])
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
		case "ResourceGroups":
			res.VdaResourceGroup = types.StringValue(stringPair.GetValue())
		case "WBCDiskStorageType":
			if res.WritebackCache == nil {
				res.WritebackCache = &WritebackCacheModel{WBCDiskStorageType: types.StringValue(stringPair.GetValue())}
			} else {
				res.WritebackCache.WBCDiskStorageType = types.StringValue(stringPair.GetValue())
			}
		case "PersistWBC":
			if res.WritebackCache == nil {
				res.WritebackCache = &WritebackCacheModel{PersistWBC: util.StringToTypeBool(stringPair.GetValue())}
			} else {
				res.WritebackCache.PersistWBC = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistOsDisk":
			if res.WritebackCache == nil {
				res.WritebackCache = &WritebackCacheModel{PersistOsDisk: util.StringToTypeBool(stringPair.GetValue())}
			} else {
				res.WritebackCache.PersistOsDisk = util.StringToTypeBool(stringPair.GetValue())
			}
		case "PersistVm":
			if res.WritebackCache == nil {
				res.WritebackCache = &WritebackCacheModel{PersistVm: util.StringToTypeBool(stringPair.GetValue())}
			} else {
				res.WritebackCache.PersistVm = util.StringToTypeBool(stringPair.GetValue())
			}
		case "StorageTypeAtShutdown":
			if res.WritebackCache == nil {
				res.WritebackCache = &WritebackCacheModel{StorageCostSaving: types.BoolValue(true)}
			} else {
				res.WritebackCache.StorageCostSaving = types.BoolValue(true)
			}
		default:
		}
	}
}

func ParseCustomPropertiesToClientModel(provisioningScheme ProvisioningSchemeModel) *[]citrixorchestration.NameValueStringPairModel {
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
		util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
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
