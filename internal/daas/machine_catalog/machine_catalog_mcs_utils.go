// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

var MappedCustomProperties = map[string]string{
	"Zones":                            "availability_zones",
	"StorageType":                      "storage_type",
	"ResourceGroups":                   "vda_resource_group",
	"UseManagedDisks":                  "use_managed_disks",
	"WBCDiskStorageType":               "wbc_disk_storage_type",
	"PersistWBC":                       "persist_wbc",
	"PersistOsDisk":                    "persist_os_disk",
	"PersistVm":                        "persist_vm",
	"CatalogZones":                     "availability_zones",
	"StorageTypeAtShutdown":            "storage_cost_saving",
	"LicenseType":                      "license_type",
	"UseSharedImageGallery":            "use_azure_compute_gallery",
	"SharedImageGalleryReplicaRatio":   "replica_ratio",
	"SharedImageGalleryReplicaMaximum": "replica_maximum",
	"UseEphemeralOsDisk":               "storage_type",
}

func getProvSchemeForMcsCatalog(plan MachineCatalogResourceModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, isOnPremises bool) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {
	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, plan.ProvisioningScheme)
	if !checkIfProvSchemeIsCloudOnly(provSchemeModel, isOnPremises, diagnostics, ctx) {
		return nil, fmt.Errorf("identity type %s is not supported in OnPremises environment. ", provSchemeModel.IdentityType.ValueString())
	}

	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, provSchemeModel.Hypervisor.ValueString())
	if err != nil {
		return nil, err
	}

	hypervisorResourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, provSchemeModel.Hypervisor.ValueString(), provSchemeModel.HypervisorResourcePool.ValueString())
	if err != nil {
		return nil, err
	}

	provisioningScheme, err := buildProvSchemeForMcsCatalog(ctx, client, diagnostics, util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, plan.ProvisioningScheme), hypervisor, hypervisorResourcePool)
	if err != nil {
		return nil, err
	}

	return provisioningScheme, nil
}

func buildProvSchemeForMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, provisioningSchemePlan ProvisioningSchemeModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {
	var machineAccountCreationRules citrixorchestration.MachineAccountCreationRulesRequestModel
	machineAccountCreationRulesModel := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, diag, provisioningSchemePlan.MachineAccountCreationRules)
	machineAccountCreationRules.SetNamingScheme(machineAccountCreationRulesModel.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(machineAccountCreationRulesModel.NamingSchemeType.ValueString())
	if err != nil {
		diag.AddError(
			"Error creating Machine Catalog",
			fmt.Sprintf("Unsupported machine account naming scheme type. Error: %s", err.Error()),
		)
		return nil, err
	}

	machineAccountCreationRules.SetNamingSchemeType(*namingScheme)
	if !provisioningSchemePlan.MachineDomainIdentity.IsNull() {
		machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, diag, provisioningSchemePlan.MachineDomainIdentity)
		machineAccountCreationRules.SetDomain(machineDomainIdentityModel.Domain.ValueString())
		machineAccountCreationRules.SetOU(machineDomainIdentityModel.Ou.ValueString())
	}

	azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diag, provisioningSchemePlan.AzureMachineConfig)

	var provisioningScheme citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel
	provisioningScheme.SetNumTotalMachines(int32(provisioningSchemePlan.NumTotalMachines.ValueInt64()))
	identityType := citrixorchestration.IdentityType(provisioningSchemePlan.IdentityType.ValueString())
	provisioningScheme.SetIdentityType(identityType)
	provisioningScheme.SetWorkGroupMachines(identityType == citrixorchestration.IDENTITYTYPE_AZURE_AD ||
		identityType == citrixorchestration.IDENTITYTYPE_WORKGROUP) // AzureAD and Workgroup identity types are non-domain joined
	if identityType == citrixorchestration.IDENTITYTYPE_AZURE_AD && azureMachineConfigModel.EnrollInIntune.ValueBool() {
		provisioningScheme.SetDeviceManagementType(citrixorchestration.DEVICEMANAGEMENTTYPE_INTUNE)
	}
	provisioningScheme.SetMachineAccountCreationRules(machineAccountCreationRules)
	provisioningScheme.SetResourcePool(provisioningSchemePlan.HypervisorResourcePool.ValueString())

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM || hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		customProperties := parseCustomPropertiesToClientModel(ctx, diag, provisioningSchemePlan, hypervisor.ConnectionType)
		provisioningScheme.SetCustomProperties(customProperties)
	}

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		serviceOffering := azureMachineConfigModel.ServiceOffering.ValueString()
		queryPath := "serviceoffering.folder"
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, util.ServiceOfferingResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetServiceOfferingPath(serviceOfferingPath)

		azureMasterImageModel := util.ObjectValueToTypedObject[AzureMasterImageModel](ctx, diag, azureMachineConfigModel.AzureMasterImage)
		sharedSubscription := azureMasterImageModel.SharedSubscription.ValueString()
		resourceGroup := azureMasterImageModel.ResourceGroup.ValueString()
		masterImage := azureMasterImageModel.MasterImage.ValueString()
		imagePath := ""
		imageBasePath := "image.folder"
		if sharedSubscription != "" {
			imageBasePath = fmt.Sprintf("image.folder\\%s.sharedsubscription", sharedSubscription)
		}
		if masterImage != "" {
			storageAccount := azureMasterImageModel.StorageAccount.ValueString()
			container := azureMasterImageModel.Container.ValueString()
			if storageAccount != "" && container != "" {
				queryPath = fmt.Sprintf(
					"%s\\%s.resourcegroup\\%s.storageaccount\\%s.container",
					imageBasePath,
					resourceGroup,
					storageAccount,
					container)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, masterImage, "", "")
				if err != nil {
					diag.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image VHD %s in container %s of storage account %s, error: %s", masterImage, container, storageAccount, err.Error()),
					)
					return nil, err
				}
			} else {
				queryPath = fmt.Sprintf(
					"%s\\%s.resourcegroup",
					imageBasePath,
					resourceGroup)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, masterImage, "", "")
				if err != nil {
					diag.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image Managed Disk or Snapshot %s, error: %s", masterImage, err.Error()),
					)
					return nil, err
				}
			}
		} else if !azureMasterImageModel.GalleryImage.IsNull() {
			azureGalleryImage := util.ObjectValueToTypedObject[GalleryImageModel](ctx, diag, azureMasterImageModel.GalleryImage)
			gallery := azureGalleryImage.Gallery.ValueString()
			definition := azureGalleryImage.Definition.ValueString()
			version := azureGalleryImage.Version.ValueString()
			if gallery != "" && definition != "" {
				queryPath = fmt.Sprintf(
					"%s\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
					imageBasePath,
					resourceGroup,
					gallery,
					definition)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
				if err != nil {
					diag.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", definition, version, gallery, err.Error()),
					)
					return nil, err
				}
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		machineProfile := azureMachineConfigModel.MachineProfile
		if !machineProfile.IsNull() {
			machineProfilePath, err := handleMachineProfileForAzureMcsCatalog(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), util.ObjectValueToTypedObject[AzureMachineProfileModel](ctx, diag, machineProfile), "creating")
			if err != nil {
				return nil, err
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if !azureMachineConfigModel.WritebackCache.IsNull() {
			azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, diag, azureMachineConfigModel.WritebackCache)
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(azureWbcModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !azureWbcModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(azureWbcModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}

		if !azureMachineConfigModel.DiskEncryptionSet.IsNull() {
			diskEncryptionSetModel := util.ObjectValueToTypedObject[AzureDiskEncryptionSetModel](ctx, diag, azureMachineConfigModel.DiskEncryptionSet)
			diskEncryptionSet := diskEncryptionSetModel.DiskEncryptionSetName.ValueString()
			diskEncryptionSetRg := diskEncryptionSetModel.DiskEncryptionSetResourceGroup.ValueString()
			des, err := util.GetSingleResourceFromHypervisor(ctx, client, hypervisor.GetId(), hypervisorResourcePool.GetId(), fmt.Sprintf("%s\\diskencryptionset.folder", hypervisorResourcePool.GetXDPath()), diskEncryptionSet, "", diskEncryptionSetRg)
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate disk encryption set %s in resource group %s, error: %s", diskEncryptionSet, diskEncryptionSetRg, err.Error()),
				)
			}

			customProp := provisioningScheme.GetCustomProperties()
			util.AppendNameValueStringPair(&customProp, "DiskEncryptionSetId", des.GetId())
			provisioningScheme.SetCustomProperties(customProp)
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, diag, provisioningSchemePlan.AwsMachineConfig)
		inputServiceOffering := awsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetServiceOfferingPath(serviceOffering)

		masterImage := awsMachineConfig.MasterImage.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, awsMachineConfig.ImageAmi.ValueString())
		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", masterImage, awsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		securityGroupPaths := []string{}
		for _, securityGroup := range util.StringListToStringArray(ctx, diag, awsMachineConfig.SecurityGroups) {
			securityGroupPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", securityGroup, util.SecurityGroupResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate security group %s, error: %s", securityGroup, err.Error()),
				)
				return nil, err
			}

			securityGroupPaths = append(securityGroupPaths, securityGroupPath)
		}
		provisioningScheme.SetSecurityGroups(securityGroupPaths)

		tenancyType, err := citrixorchestration.NewTenancyTypeFromValue(awsMachineConfig.TenancyType.ValueString())
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"Unsupported provisioning type.",
			)

			return nil, err
		}
		provisioningScheme.SetTenancyType(*tenancyType)

	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, diag, provisioningSchemePlan.GcpMachineConfig)
		imagePath := ""
		snapshot := gcpMachineConfig.MachineSnapshot.ValueString()
		imageVm := gcpMachineConfig.MasterImage.ValueString()
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", imageVm)
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, snapshot, util.SnapshotResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate snapshot %s of master image VM %s on GCP, error: %s", snapshot, imageVm, err.Error()),
				)
				return nil, err
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageVm, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", imageVm, err.Error()),
				)
				return nil, err
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)
		machineProfile := gcpMachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfile, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", gcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return nil, err
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if !gcpMachineConfig.WritebackCache.IsNull() {
			writeBackCacheModel := util.ObjectValueToTypedObject[GcpWritebackCacheModel](ctx, diag, gcpMachineConfig.WritebackCache)
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, diag, provisioningSchemePlan.VsphereMachineConfig)
		provisioningScheme.SetMemoryMB(int32(vSphereMachineConfig.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(vSphereMachineConfig.CpuCount.ValueInt64()))

		image := vSphereMachineConfig.MasterImageVm.ValueString()
		snapshot := vSphereMachineConfig.ImageSnapshot.ValueString()
		imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		if !vSphereMachineConfig.WritebackCache.IsNull() {
			provisioningScheme.SetUseWriteBackCache(true)
			writeBackCacheModel := util.ObjectValueToTypedObject[VsphereWritebackCacheModel](ctx, diag, vSphereMachineConfig.WritebackCache)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if !writeBackCacheModel.WriteBackCacheDriveLetter.IsNull() {
				provisioningScheme.SetWriteBackCacheDriveLetter(writeBackCacheModel.WriteBackCacheDriveLetter.ValueString())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, diag, provisioningSchemePlan.XenserverMachineConfig)
		provisioningScheme.SetCpuCount(int32(xenserverMachineConfig.CpuCount.ValueInt64()))
		provisioningScheme.SetMemoryMB(int32(xenserverMachineConfig.MemoryMB.ValueInt64()))

		image := xenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := xenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		if xenserverMachineConfig.WritebackCache.IsNull() {
			provisioningScheme.SetUseWriteBackCache(true)
			writeBackCacheModel := util.ObjectValueToTypedObject[XenserverWritebackCacheModel](ctx, diag, xenserverMachineConfig.WritebackCache)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, diag, provisioningSchemePlan.NutanixMachineConfigModel)
		if hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
			return nil, fmt.Errorf("unsupported hypervisor plugin %s", hypervisor.GetPluginId())
		}

		provisioningScheme.SetMemoryMB(int32(nutanixMachineConfig.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(nutanixMachineConfig.CpuCount.ValueInt64()))
		provisioningScheme.SetCoresPerCpuCount(int32(nutanixMachineConfig.CoresPerCpuCount.ValueInt64()))

		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.MasterImage.ValueString(), util.TemplateResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate master image %s on NUTANIX, error: %s", nutanixMachineConfig.MasterImage.ValueString(), err.Error()),
			)
			return nil, err
		}

		containerId, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.Container.ValueString(), util.StorageResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate container %s on NUTANIX, error: %s", nutanixMachineConfig.Container.ValueString(), err.Error()),
			)
			return nil, err
		}

		provisioningScheme.SetMasterImagePath(imagePath)
		customProperties := provisioningScheme.GetCustomProperties()
		util.AppendNameValueStringPair(&customProperties, "NutanixContainerId", containerId)
		provisioningScheme.SetCustomProperties(customProperties)
	}

	if !provisioningSchemePlan.NetworkMapping.IsNull() {
		networkMappingModel := util.ObjectListToTypedArray[NetworkMappingModel](ctx, diag, provisioningSchemePlan.NetworkMapping)
		networkMapping, err := parseNetworkMappingToClientModel(networkMappingModel, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to find hypervisor network, error: %s", err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetNetworkMapping(networkMapping)
	}

	return &provisioningScheme, nil
}

func setProvSchemePropertiesForUpdateCatalog(provisioningSchemePlan ProvisioningSchemeModel, body citrixorchestration.UpdateMachineCatalogRequestModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (citrixorchestration.UpdateMachineCatalogRequestModel, error) {
	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, provisioningSchemePlan.Hypervisor.ValueString())
	if err != nil {
		return body, err
	}

	hypervisorResourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, provisioningSchemePlan.Hypervisor.ValueString(), provisioningSchemePlan.HypervisorResourcePool.ValueString())
	if err != nil {
		return body, err
	}

	// Resolve resource path for service offering and master image
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diagnostics, provisioningSchemePlan.AzureMachineConfig)
		serviceOffering := azureMachineConfigModel.ServiceOffering.ValueString()
		queryPath := "serviceoffering.folder"
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, util.ServiceOfferingResourceType, "")
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, nil, provisioningSchemePlan.AwsMachineConfig)
		inputServiceOffering := awsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", inputServiceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOffering)

		securityGroupPaths := []string{}
		for _, securityGroup := range util.StringListToStringArray(ctx, diagnostics, awsMachineConfig.SecurityGroups) {
			securityGroupPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", securityGroup, util.SecurityGroupResourceType, "")
			if err != nil {
				diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate security group %s, error: %s", securityGroup, err.Error()),
				)
				return body, err
			}

			securityGroupPaths = append(securityGroupPaths, securityGroupPath)
		}
		body.SetSecurityGroups(securityGroupPaths)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, nil, provisioningSchemePlan.XenserverMachineConfig)
		body.SetCpuCount(int32(xenserverMachineConfig.CpuCount.ValueInt64()))
		body.SetMemoryMB(int32(xenserverMachineConfig.MemoryMB.ValueInt64()))
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, nil, provisioningSchemePlan.VsphereMachineConfig)
		body.SetCpuCount(int32(vSphereMachineConfig.CpuCount.ValueInt64()))
		body.SetMemoryMB(int32(vSphereMachineConfig.MemoryMB.ValueInt64()))
	}

	if !provisioningSchemePlan.NetworkMapping.IsNull() {
		networkMappingModel := util.ObjectListToTypedArray[NetworkMappingModel](ctx, diagnostics, provisioningSchemePlan.NetworkMapping)
		networkMapping, err := parseNetworkMappingToClientModel(networkMappingModel, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to parse network mapping, error: %s", err.Error()),
			)
			return body, err
		}
		body.SetNetworkMapping(networkMapping)
	}

	customProperties := parseCustomPropertiesToClientModel(ctx, diagnostics, provisioningSchemePlan, hypervisor.ConnectionType)
	body.SetCustomProperties(customProperties)

	return body, nil
}

func generateAdminCredentialHeader(machineDomainIdentityModel MachineDomainIdentityModel) string {
	credential := fmt.Sprintf("%s\\%s:%s", machineDomainIdentityModel.Domain.ValueString(), machineDomainIdentityModel.ServiceAccount.ValueString(), machineDomainIdentityModel.ServiceAccountPassword.ValueString())
	encodedData := base64.StdEncoding.EncodeToString([]byte(credential))
	header := fmt.Sprintf("Basic %s", encodedData)

	return header
}

func deleteMachinesFromMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, provisioningSchemePlan ProvisioningSchemeModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	if catalog.GetAllocationType() != citrixorchestration.ALLOCATIONTYPE_RANDOM {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"Deleting machine(s) is supported for machine catalogs with Random allocation type only.",
		)
		return fmt.Errorf("deleting machine(s) is supported for machine catalogs with Random allocation type only")
	}

	getMachinesResponse, err := util.GetMachineCatalogMachines(ctx, client, &resp.Diagnostics, catalogId)
	if err != nil {
		return err
	}

	machineDeleteRequestCount := int(catalog.GetTotalCount()) - int(provisioningSchemePlan.NumTotalMachines.ValueInt64())
	machinesToDelete := []citrixorchestration.MachineResponseModel{}

	for _, machine := range getMachinesResponse.GetItems() {
		if !machine.GetDeliveryGroup().Id.IsSet() || machine.GetSessionCount() == 0 {
			machinesToDelete = append(machinesToDelete, machine)
		}

		if len(machinesToDelete) == machineDeleteRequestCount {
			break
		}
	}

	machinesToDeleteCount := len(machinesToDelete)

	if machineDeleteRequestCount > machinesToDeleteCount {
		errorString := fmt.Sprintf("%d machine(s) requested to be deleted. %d machine(s) qualify for deletion.", machineDeleteRequestCount, machinesToDeleteCount)

		resp.Diagnostics.AddError(
			"Error deleting machine(s) from Machine Catalog "+catalogName,
			errorString+" Ensure machine that needs to be deleted has no active sessions.",
		)

		return err
	}

	return deleteMachinesFromCatalog(ctx, client, resp, provisioningSchemePlan, machinesToDelete, catalogName, true)
}

func addMachinesToMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, provisioningSchemePlan ProvisioningSchemeModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	addMachinesCount := int32(provisioningSchemePlan.NumTotalMachines.ValueInt64()) - catalog.GetTotalCount()

	var updateMachineAccountCreationRule citrixorchestration.UpdateMachineAccountCreationRulesRequestModel
	machineAccountCreationRulesModel := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, provisioningSchemePlan.MachineAccountCreationRules)
	updateMachineAccountCreationRule.SetNamingScheme(machineAccountCreationRulesModel.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(machineAccountCreationRulesModel.NamingSchemeType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"Unsupported machine account naming scheme type.",
		)
		return err
	}
	updateMachineAccountCreationRule.SetNamingSchemeType(*namingScheme)
	if !provisioningSchemePlan.MachineDomainIdentity.IsNull() {
		machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provisioningSchemePlan.MachineDomainIdentity)
		updateMachineAccountCreationRule.SetDomain(machineDomainIdentityModel.Domain.ValueString())
		updateMachineAccountCreationRule.SetOU(machineDomainIdentityModel.Ou.ValueString())
	}

	var addMachineRequestBody citrixorchestration.AddMachineToMachineCatalogDetailRequestModel
	addMachineRequestBody.SetMachineAccountCreationRules(updateMachineAccountCreationRule)

	addMachineRequestStringBody, err := util.ConvertToString(addMachineRequestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"An unexpected error occurred: "+err.Error(),
		)
		return err
	}

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, client, provisioningSchemePlan, true)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nCould not add machine to Machine Catalog, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/Machines?async=true", catalogId)
	for i := 0; i < int(addMachinesCount); i++ {
		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetReference(strconv.Itoa(i))
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetBody(addMachineRequestStringBody)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(batchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding machine(s) to Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < int(addMachinesCount) {
		errMsg := fmt.Sprintf("An error occurred while adding machine(s) to the Machine Catalog. %d of %d machines were added to the Machine Catalog.", successfulJobs, addMachinesCount)
		err = fmt.Errorf(errMsg)
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}

	return nil
}

func updateCatalogMachineProfile(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, machineProfilePath string) error {
	var body citrixorchestration.UpdateMachineCatalogRequestModel
	body.SetMachineProfilePath(machineProfilePath)

	updateMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalog(ctx, catalog.GetId())
	updateMachineCatalogRequest = updateMachineCatalogRequest.UpdateMachineCatalogRequestModel(body).Async(true)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateMachineCatalogRequest, client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalog.GetName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating machine profile for Machine Catalog "+catalog.GetName(), &resp.Diagnostics, 15, false)
	if err != nil {
		return err
	}

	return nil
}

func updateCatalogImageAndMachineProfile(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan MachineCatalogResourceModel) error {

	catalogName := catalog.GetName()
	catalogId := catalog.GetId()

	provScheme := catalog.GetProvisioningScheme()
	masterImage := provScheme.GetMasterImage()

	machineProfile := provScheme.GetMachineProfile()

	provisioningSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)

	hypervisor, errResp := util.GetHypervisor(ctx, client, &resp.Diagnostics, provisioningSchemePlan.Hypervisor.ValueString())
	if errResp != nil {
		return errResp
	}

	hypervisorResourcePool, errResp := util.GetHypervisorResourcePool(ctx, client, &resp.Diagnostics, provisioningSchemePlan.Hypervisor.ValueString(), provisioningSchemePlan.HypervisorResourcePool.ValueString())
	if errResp != nil {
		return errResp
	}

	// Check if XDPath has changed for the image
	imagePath := ""
	machineProfilePath := ""
	var err error
	updateCustomProperties := []citrixorchestration.NameValueStringPairModel{}

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.AzureMachineConfig)
		azureMasterImageModel := util.ObjectValueToTypedObject[AzureMasterImageModel](ctx, &resp.Diagnostics, azureMachineConfigModel.AzureMasterImage)
		newImage := azureMasterImageModel.MasterImage.ValueString()
		resourceGroup := azureMasterImageModel.ResourceGroup.ValueString()
		azureMachineProfile := azureMachineConfigModel.MachineProfile
		if newImage != "" {
			storageAccount := azureMasterImageModel.StorageAccount.ValueString()
			container := azureMasterImageModel.Container.ValueString()
			if storageAccount != "" && container != "" {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.storageaccount\\%s.container",
					resourceGroup,
					storageAccount,
					container)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image VHD %s in container %s of storage account %s, error: %s", newImage, container, storageAccount, err.Error()),
					)
					return err
				}
			} else {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup",
					resourceGroup)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to resolve master image Managed Disk or Snapshot %s, error: %s", newImage, err.Error()),
					)
					return err
				}
			}
		} else if !azureMasterImageModel.GalleryImage.IsNull() {
			azureGalleryImage := util.ObjectValueToTypedObject[GalleryImageModel](ctx, &resp.Diagnostics, azureMasterImageModel.GalleryImage)
			gallery := azureGalleryImage.Gallery.ValueString()
			definition := azureGalleryImage.Definition.ValueString()
			version := azureGalleryImage.Version.ValueString()
			if gallery != "" && definition != "" {
				queryPath := fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
					resourceGroup,
					gallery,
					definition)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", newImage, version, gallery, err.Error()),
					)
					return err
				}
			}
		}

		if !azureMachineProfile.IsNull() {
			machineProfilePath, err = handleMachineProfileForAzureMcsCatalog(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), util.ObjectValueToTypedObject[AzureMachineProfileModel](ctx, &resp.Diagnostics, azureMachineProfile), "updating")
			if err != nil {
				return err
			}
		}

		if !azureMachineConfigModel.UseAzureComputeGallery.IsNull() {
			azureComputeGalleryModel := util.ObjectValueToTypedObject[AzureComputeGallerySettings](ctx, &resp.Diagnostics, azureMachineConfigModel.UseAzureComputeGallery)
			util.AppendNameValueStringPair(&updateCustomProperties, "UseSharedImageGallery", "true")
			util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaRatio", strconv.Itoa(int(azureComputeGalleryModel.ReplicaRatio.ValueInt64())))
			util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaMaximum", strconv.Itoa(int(azureComputeGalleryModel.ReplicaMaximum.ValueInt64())))
		} else {
			util.AppendNameValueStringPair(&updateCustomProperties, "UseSharedImageGallery", "false")
			util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaRatio", "")
			util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaMaximum", "")
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.AwsMachineConfig)
		imageId := fmt.Sprintf("%s (%s)", awsMachineConfig.MasterImage.ValueString(), awsMachineConfig.ImageAmi.ValueString())
		imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", awsMachineConfig.MasterImage.ValueString(), awsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.GcpMachineConfig)
		newImage := gcpMachineConfig.MasterImage.ValueString()
		snapshot := gcpMachineConfig.MachineSnapshot.ValueString()
		gcpMachineProfile := gcpMachineConfig.MachineProfile.ValueString()

		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", newImage)
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, snapshot, util.SnapshotResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate snapshot %s of master image %s on GCP, error: %s", snapshot, newImage, err.Error()),
				)
				return err
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", newImage, util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", newImage, err.Error()),
				)
				return err
			}
		}
		if gcpMachineProfile != "" {
			machineProfilePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", gcpMachineConfig.MachineProfile.ValueString(), util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", gcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.VsphereMachineConfig)
		newImage := vSphereMachineConfig.MasterImageVm.ValueString()
		snapshot := vSphereMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "updating")
		if err != nil {
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.XenserverMachineConfig)
		newImage := xenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := xenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "updating")
		if err != nil {
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.NutanixMachineConfigModel)
		if hypervisor.GetPluginId() == util.NUTANIX_PLUGIN_ID {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.MasterImage.ValueString(), util.TemplateResourceType, "")

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image %s on NUTANIX, error: %s", nutanixMachineConfig.MasterImage.ValueString(), err.Error()),
				)
				return err
			}
		}
	}

	if machineProfile.GetXDPath() != machineProfilePath {
		err = updateCatalogMachineProfile(ctx, client, resp, catalog, machineProfilePath)
		if err != nil {
			return err
		}
	}

	if masterImage.GetXDPath() == imagePath {
		return nil
	}

	// Update Master Image for Machine Catalog
	var updateProvisioningSchemeModel citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel
	var rebootOption citrixorchestration.RebootMachinesRequestModel

	// Update the image immediately
	rebootOption.SetRebootDuration(60)
	rebootOption.SetWarningDuration(15)
	rebootOption.SetWarningMessage("Warning: An important update is about to be installed. To ensure that no loss of data occurs, save any outstanding work and close all applications.")

	functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
		)
		return err
	}

	updateProvisioningSchemeModel.SetRebootOptions(rebootOption)
	updateProvisioningSchemeModel.SetMinimumFunctionalLevel(*functionalLevel)
	updateProvisioningSchemeModel.SetMasterImagePath(imagePath)
	updateProvisioningSchemeModel.SetStoreOldImage(true)

	if len(updateCustomProperties) > 0 {
		updateProvisioningSchemeModel.SetCustomProperties(updateCustomProperties)
	}

	updateMasterImageRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalogProvisioningScheme(ctx, catalogId)
	updateMasterImageRequest = updateMasterImageRequest.UpdateMachineCatalogProvisioningSchemeRequestModel(updateProvisioningSchemeModel)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateMasterImageRequest, client).Async(true).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Image for Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating Image for Machine Catalog "+catalogName, &resp.Diagnostics, 60, false)
	if err != nil {
		return err
	}

	return nil
}

func (r MachineCatalogResourceModel) updateCatalogWithProvScheme(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, pluginId string) MachineCatalogResourceModel {
	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, r.ProvisioningScheme)
	provScheme := catalog.GetProvisioningScheme()
	resourcePool := provScheme.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	machineAccountCreateRules := provScheme.GetMachineAccountCreationRules()
	domain := machineAccountCreateRules.GetDomain()
	customProperties := provScheme.GetCustomProperties()

	// Refresh Hypervisor and Resource Pool
	provSchemeModel.Hypervisor = types.StringValue(hypervisor.GetId())
	provSchemeModel.HypervisorResourcePool = types.StringValue(resourcePool.GetId())

	switch *connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diagnostics, provSchemeModel.AzureMachineConfig)
		azureMachineConfigModel.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.AzureMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, azureMachineConfigModel)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" && !provSchemeModel.AvailabilityZones.IsNull() {
				provSchemeModel.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, diagnostics, provSchemeModel.AwsMachineConfig)
		if provSchemeModel.AwsMachineConfig.IsNull() {
			awsMachineConfig = AwsMachineConfigModel{}
		} else {

			serviceOfferingObject, err := util.GetSingleResourceFromHypervisor(ctx, client, hypervisor.GetId(), resourcePool.GetId(), "", provScheme.GetServiceOffering(), util.ServiceOfferingResourceType, "")
			if err == nil {
				provScheme.SetServiceOffering(serviceOfferingObject.GetId())
				catalog.SetProvisioningScheme(provScheme)
			}
		}
		awsMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.AwsMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, awsMachineConfig)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" {
				provSchemeModel.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, diagnostics, provSchemeModel.GcpMachineConfig)
		if provSchemeModel.GcpMachineConfig.IsNull() {
			gcpMachineConfig = GcpMachineConfigModel{}
		}

		gcpMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.GcpMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, gcpMachineConfig)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "CatalogZones" && !provSchemeModel.AvailabilityZones.IsNull() {
				provSchemeModel.AvailabilityZones = types.StringValue(stringPair.GetValue())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, diagnostics, provSchemeModel.VsphereMachineConfig)
		if provSchemeModel.VsphereMachineConfig.IsNull() {
			vSphereMachineConfig = VsphereMachineConfigModel{}
		}
		vSphereMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.VsphereMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, vSphereMachineConfig)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, diagnostics, provSchemeModel.XenserverMachineConfig)
		if provSchemeModel.XenserverMachineConfig.IsNull() {
			xenserverMachineConfig = XenserverMachineConfigModel{}
		}

		xenserverMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.XenserverMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, xenserverMachineConfig)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		if pluginId == util.NUTANIX_PLUGIN_ID {
			nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, diagnostics, provSchemeModel.NutanixMachineConfigModel)
			if provSchemeModel.NutanixMachineConfigModel.IsNull() {
				nutanixMachineConfig = NutanixMachineConfigModel{}
			}

			nutanixMachineConfig.RefreshProperties(*catalog)
			provSchemeModel.NutanixMachineConfigModel = util.TypedObjectToObjectValue(ctx, diagnostics, nutanixMachineConfig)
		}
	}

	remoteCustomProperties := map[string]string{}
	for _, customProperty := range customProperties {
		remoteCustomProperties[customProperty.GetName()] = customProperty.GetValue()
	}
	refreshedCustomProperties := []CustomPropertyModel{}
	customPropertiesModel := util.ObjectListToTypedArray[CustomPropertyModel](ctx, diagnostics, provSchemeModel.CustomProperties)
	for _, customProperty := range customPropertiesModel {
		if value, ok := remoteCustomProperties[customProperty.Name.ValueString()]; ok {
			newProperty := CustomPropertyModel{}
			newProperty.Name = customProperty.Name
			newProperty.Value = types.StringValue(value)
			refreshedCustomProperties = append(refreshedCustomProperties, newProperty)
		}
	}

	if len(refreshedCustomProperties) == 0 {
		attributesMap, err := util.AttributeMapFromObject(CustomPropertyModel{})
		if err != nil {
			diagnostics.AddError("Error converting schema to attribute map. Error: ", err.Error())
		}
		provSchemeModel.CustomProperties = types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	} else {
		provSchemeModel.CustomProperties = util.TypedArrayToObjectList[CustomPropertyModel](ctx, diagnostics, refreshedCustomProperties)
	}

	// Refresh Total Machine Count
	provSchemeModel.NumTotalMachines = types.Int64Value(int64(provScheme.GetMachineCount()))

	// Refresh Identity Type
	if identityType := types.StringValue(string(provScheme.GetIdentityType())); identityType.ValueString() != "" {
		provSchemeModel.IdentityType = identityType
	} else {
		provSchemeModel.IdentityType = types.StringNull()
	}

	// Refresh Network Mapping
	networkMaps := provScheme.GetNetworkMaps()

	if len(networkMaps) > 0 && !provSchemeModel.NetworkMapping.IsNull() {
		provSchemeModel.NetworkMapping = util.RefreshListValueProperties[NetworkMappingModel, citrixorchestration.NetworkMapResponseModel](ctx, diagnostics, provSchemeModel.NetworkMapping, "NetworkDevice", networkMaps, "DeviceId", "RefreshListItem")
	} else {
		attributesMap, err := util.AttributeMapFromObject(NetworkMappingModel{})
		if err != nil {
			diagnostics.AddError("Error converting schema to attribute map. Error: ", err.Error())
		}
		provSchemeModel.NetworkMapping = types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	}

	// Identity Pool Properties
	machineAccountCreationRulesModel := MachineAccountCreationRulesModel{}
	machineAccountCreationRulesModel.NamingScheme = types.StringValue(machineAccountCreateRules.GetNamingScheme())
	namingSchemeType := machineAccountCreateRules.GetNamingSchemeType()
	machineAccountCreationRulesModel.NamingSchemeType = types.StringValue(string(namingSchemeType))
	provSchemeModel.MachineAccountCreationRules = util.TypedObjectToObjectValue(ctx, diagnostics, machineAccountCreationRulesModel)

	// Domain Identity Properties
	if provScheme.GetIdentityType() == citrixorchestration.IDENTITYTYPE_AZURE_AD ||
		provScheme.GetIdentityType() == citrixorchestration.IDENTITYTYPE_WORKGROUP {
		return r
	}

	machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, diagnostics, provSchemeModel.MachineDomainIdentity)

	if domain.GetName() != "" {
		machineDomainIdentityModel.Domain = types.StringValue(domain.GetName())
	}
	if machineAccountCreateRules.GetOU() != "" {
		machineDomainIdentityModel.Ou = types.StringValue(machineAccountCreateRules.GetOU())
	}

	provSchemeModel.MachineDomainIdentity = util.TypedObjectToObjectValue(ctx, diagnostics, machineDomainIdentityModel)
	r.ProvisioningScheme = util.TypedObjectToObjectValue(ctx, diagnostics, provSchemeModel)
	return r
}

func parseCustomPropertiesToClientModel(ctx context.Context, diagnostics *diag.Diagnostics, provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType) []citrixorchestration.NameValueStringPairModel {
	var res = &[]citrixorchestration.NameValueStringPairModel{}
	switch connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diagnostics, provisioningScheme.AzureMachineConfig)
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
		} else {
			util.AppendNameValueStringPair(res, "Zones", "")
		}
		if !azureMachineConfigModel.StorageType.IsNull() {
			if azureMachineConfigModel.StorageType.ValueString() == util.AzureEphemeralOSDisk {
				util.AppendNameValueStringPair(res, "UseEphemeralOsDisk", "true")
			} else {
				util.AppendNameValueStringPair(res, "StorageType", azureMachineConfigModel.StorageType.ValueString())
			}
		}
		if !azureMachineConfigModel.VdaResourceGroup.IsNull() {
			util.AppendNameValueStringPair(res, "ResourceGroups", azureMachineConfigModel.VdaResourceGroup.ValueString())
		}
		if !azureMachineConfigModel.UseManagedDisks.IsNull() {
			if azureMachineConfigModel.UseManagedDisks.ValueBool() {
				util.AppendNameValueStringPair(res, "UseManagedDisks", "true")
			} else {
				util.AppendNameValueStringPair(res, "UseManagedDisks", "false")
			}
		}
		if !azureMachineConfigModel.WritebackCache.IsNull() {
			azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, diagnostics, azureMachineConfigModel.WritebackCache)
			if !azureWbcModel.WBCDiskStorageType.IsNull() {
				util.AppendNameValueStringPair(res, "WBCDiskStorageType", azureWbcModel.WBCDiskStorageType.ValueString())
			}
			if azureWbcModel.PersistWBC.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistWBC", "true")
				if azureWbcModel.StorageCostSaving.ValueBool() {
					util.AppendNameValueStringPair(res, "StorageTypeAtShutdown", "Standard_LRS")
				}
			}
			if azureWbcModel.PersistOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistOsDisk", "true")
				if azureWbcModel.PersistVm.ValueBool() {
					util.AppendNameValueStringPair(res, "PersistVm", "true")
				}
			}
		}

		licenseType := azureMachineConfigModel.LicenseType.ValueString()
		util.AppendNameValueStringPair(res, "LicenseType", licenseType)

		if !azureMachineConfigModel.UseAzureComputeGallery.IsNull() {
			azureComputeGalleryModel := util.ObjectValueToTypedObject[AzureComputeGallerySettings](ctx, diagnostics, azureMachineConfigModel.UseAzureComputeGallery)
			util.AppendNameValueStringPair(res, "UseSharedImageGallery", "true")
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaRatio", strconv.Itoa(int(azureComputeGalleryModel.ReplicaRatio.ValueInt64())))
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaMaximum", strconv.Itoa(int(azureComputeGalleryModel.ReplicaMaximum.ValueInt64())))
		} else {
			util.AppendNameValueStringPair(res, "UseSharedImageGallery", "false")
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaRatio", "")
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaMaximum", "")
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "Zones", provisioningScheme.AvailabilityZones.ValueString())
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](context.Background(), nil, provisioningScheme.GcpMachineConfig)
		if !provisioningScheme.AvailabilityZones.IsNull() {
			util.AppendNameValueStringPair(res, "CatalogZones", provisioningScheme.AvailabilityZones.ValueString())
		}
		if !gcpMachineConfig.StorageType.IsNull() {
			util.AppendNameValueStringPair(res, "StorageType", gcpMachineConfig.StorageType.ValueString())
		}
		if !gcpMachineConfig.WritebackCache.IsNull() {
			writebackCacheModel := util.ObjectValueToTypedObject[GcpWritebackCacheModel](context.Background(), nil, gcpMachineConfig.WritebackCache)
			if !writebackCacheModel.WBCDiskStorageType.IsNull() {
				util.AppendNameValueStringPair(res, "WBCDiskStorageType", writebackCacheModel.WBCDiskStorageType.ValueString())
			}
			if writebackCacheModel.PersistWBC.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistWBC", "true")
			}
			if writebackCacheModel.PersistOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "PersistOsDisk", "true")
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		return nil
	}

	if len(*res) == 0 {
		return nil
	}

	return *res
}

func parseNetworkMappingToClientModel(networkMappings []NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel, hypervisorPluginId string) ([]citrixorchestration.NetworkMapRequestModel, error) {
	var networks []citrixorchestration.HypervisorResourceRefResponseModel
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		networks = resourcePool.Subnets
	} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == util.NUTANIX_PLUGIN_ID {
		networks = resourcePool.Networks
	}

	var res = []citrixorchestration.NetworkMapRequestModel{}
	for _, networkMapping := range networkMappings {
		var networkName string
		if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER ||
			resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == util.NUTANIX_PLUGIN_ID {
			networkName = networkMapping.Network.ValueString()
		} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
			networkName = fmt.Sprintf("%s (%s)", networkMapping.Network.ValueString(), resourcePool.GetResourcePoolRootId())
		}
		network := slices.IndexFunc(networks, func(c citrixorchestration.HypervisorResourceRefResponseModel) bool {
			return strings.EqualFold(c.GetName(), networkName)
		})
		if network == -1 {
			return res, fmt.Errorf("network %s not found", networkName)
		}

		networkMapRequestModel := citrixorchestration.NetworkMapRequestModel{
			NetworkDeviceNameOrId: *citrixorchestration.NewNullableString(networkMapping.NetworkDevice.ValueStringPointer()),
			NetworkPath:           networks[network].GetXDPath(),
		}

		if resourcePool.GetConnectionType() == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER || (resourcePool.GetConnectionType() == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == util.NUTANIX_PLUGIN_ID) {
			networkMapRequestModel.SetDeviceNameOrId(networks[network].GetId())
		}

		res = append(res, networkMapRequestModel)
	}

	return res, nil
}

func getOnPremImagePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisorName, resourcePoolName, image, snapshot, action string) (string, error) {
	queryPath := ""
	resourceType := util.VirtualMachineResourceType
	resourceName := image
	errTemplate := fmt.Sprintf("Failed to locate master image machine %s", image)
	if snapshot != "" {
		queryPath = fmt.Sprintf("%s.vm", image)
		snapshotSegments := strings.Split(snapshot, "/")
		snapshotName := strings.Split(snapshotSegments[len(snapshotSegments)-1], ".")[0]
		for i := 0; i < len(snapshotSegments)-1; i++ {
			queryPath = queryPath + "\\" + snapshotSegments[i] + ".snapshot"
		}

		resourceType = util.SnapshotResourceType
		resourceName = snapshotName
		errTemplate = fmt.Sprintf("Failed to locate snapshot %s of master image VM %s", snapshotName, image)
	}

	imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisorName, resourcePoolName, queryPath, resourceName, resourceType, "")
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error %s Machine Catalog", action),
			fmt.Sprintf("%s, error: %s", errTemplate, err.Error()),
		)
		return "", err
	}

	return imagePath, nil
}

func handleMachineProfileForAzureMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, hypervisorName, resourcePoolName string, machineProfile AzureMachineProfileModel, action string) (string, error) {
	machineProfileResourceGroup := machineProfile.MachineProfileResourceGroup.ValueString()
	machineProfileVmOrTemplateSpecVersion := machineProfile.MachineProfileVmName.ValueString()
	resourceType := util.VirtualMachineResourceType
	queryPath := fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
	errorMessage := fmt.Sprintf("Failed to locate machine profile vm %s on Azure", machineProfile.MachineProfileVmName.ValueString())
	isUsingTemplateSpec := false
	if machineProfile.MachineProfileVmName.IsNull() {
		isUsingTemplateSpec = true
		machineProfileVmOrTemplateSpecVersion = machineProfile.MachineProfileTemplateSpecVersion.ValueString()
		queryPath = fmt.Sprintf("%s\\%s.templatespec", queryPath, machineProfile.MachineProfileTemplateSpecName.ValueString())
		resourceType = ""
		errorMessage = fmt.Sprintf("Failed to locate machine profile template spec %s with version %s on Azure", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfile.MachineProfileTemplateSpecVersion.ValueString())
	}
	machineProfileResource, err := util.GetSingleResourceFromHypervisor(ctx, client, hypervisorName, resourcePoolName, queryPath, machineProfileVmOrTemplateSpecVersion, resourceType, "")
	if err != nil {
		diag.AddError(
			fmt.Sprintf("Error %s Machine Catalog", action),
			fmt.Sprintf("%s, error: %s", errorMessage, err.Error()),
		)
		return "", err
	}
	if isUsingTemplateSpec {
		// validate the template spec
		isValid, errorMsg := util.ValidateHypervisorResource(ctx, client, hypervisorName, resourcePoolName, machineProfileResource.GetRelativePath())
		if !isValid {
			diag.AddError(
				fmt.Sprintf("Error %s Machine Catalog", action),
				fmt.Sprintf("Failed to validate template spec %s with version %s, %s", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfileVmOrTemplateSpecVersion, errorMsg),
			)
			return "", fmt.Errorf("failed to validate template spec %s with version %s, %s", machineProfile.MachineProfileTemplateSpecName.ValueString(), machineProfileVmOrTemplateSpecVersion, errorMsg)
		}
	}

	return machineProfileResource.GetXDPath(), nil
}
