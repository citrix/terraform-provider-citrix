// Copyright © 2024. Citrix Systems, Inc.

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
	"github.com/citrix/terraform-provider-citrix/internal/daas/image_definition"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func getProvSchemeForCatalog(plan MachineCatalogResourceModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, isOnPremises bool, provisioningType *citrixorchestration.ProvisioningType) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {
	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, plan.ProvisioningScheme)
	if !checkIfProvSchemeIsCloudOnly(ctx, diagnostics, provSchemeModel, isOnPremises) {
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

	provisioningScheme, err := buildProvSchemeForCatalog(ctx, client, diagnostics, util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, plan.ProvisioningScheme), hypervisor, hypervisorResourcePool, provisioningType)
	if err != nil {
		return nil, err
	}

	return provisioningScheme, nil
}

func buildProvSchemeForCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, provisioningSchemePlan ProvisioningSchemeModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel, provisioningType *citrixorchestration.ProvisioningType) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {
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

	if !provisioningSchemePlan.MachineAccountCreationRules.IsNull() {
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
		provisioningScheme.SetMachineAccountCreationRules(machineAccountCreationRules)
	}

	if !provisioningSchemePlan.MachineADAccounts.IsNull() {
		machineADAccounts := util.ObjectListToTypedArray[MachineADAccountModel](ctx, diag, provisioningSchemePlan.MachineADAccounts)
		availableMachineAccounts, err := constructAvailableMachineAccountsRequestModel(diag, machineADAccounts, "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetAvailableMachineAccounts(availableMachineAccounts)
	}

	provisioningScheme.SetResourcePool(provisioningSchemePlan.HypervisorResourcePool.ValueString())

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM || hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		customProperties, err := parseCustomPropertiesToClientModel(ctx, diag, client, provisioningSchemePlan, hypervisor.ConnectionType, provisioningType, false)
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetCustomProperties(customProperties)
	}

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		serviceOffering := azureMachineConfigModel.ServiceOffering.ValueString()
		queryPath := "serviceoffering.folder"
		serviceOfferingPath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, util.ServiceOfferingResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetServiceOfferingPath(serviceOfferingPath)

		if *provisioningType == citrixorchestration.PROVISIONINGTYPE_MCS {
			err = setProvisioningSchemeForMcsCatalog(ctx, client, azureMachineConfigModel, diag, &provisioningScheme, hypervisor, hypervisorResourcePool)
		} else if *provisioningType == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING {
			err = setProvisioningSchemeForPvsCatalog(ctx, azureMachineConfigModel, diag, &provisioningScheme)
		}

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to set provisioning scheme for catalog with provisioning type %s on Azure, error: %s", *provisioningType, err.Error()),
			)
			return nil, err
		}

		machineProfile := azureMachineConfigModel.MachineProfile
		if !machineProfile.IsNull() {
			machineProfilePath, err := util.HandleMachineProfileForAzureMcsPvsCatalog(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), util.ObjectValueToTypedObject[util.AzureMachineProfileModel](ctx, diag, machineProfile), "Error creating Machine Catalog")
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
			diskEncryptionSetModel := util.ObjectValueToTypedObject[util.AzureDiskEncryptionSetModel](ctx, diag, azureMachineConfigModel.DiskEncryptionSet)
			diskEncryptionSet := diskEncryptionSetModel.DiskEncryptionSetName.ValueString()
			diskEncryptionSetRg := diskEncryptionSetModel.DiskEncryptionSetResourceGroup.ValueString()
			des, httpResp, err := util.GetSingleResourceFromHypervisor(ctx, client, diag, hypervisor.GetId(), hypervisorResourcePool.GetId(), fmt.Sprintf("%s\\diskencryptionset.folder", hypervisorResourcePool.GetXDPath()), diskEncryptionSet, "", diskEncryptionSetRg)
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate disk encryption set %s in resource group %s, error: %s", diskEncryptionSet, diskEncryptionSetRg, err.Error()),
				)
			}

			customProp := provisioningScheme.GetCustomProperties()
			util.AppendNameValueStringPair(&customProp, "DiskEncryptionSetId", des.GetId())
			provisioningScheme.SetCustomProperties(customProp)
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, diag, provisioningSchemePlan.AwsMachineConfig)
		inputServiceOffering := awsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetServiceOfferingPath(serviceOffering)

		masterImage := awsMachineConfig.MasterImage.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, awsMachineConfig.ImageAmi.ValueString())
		imagePath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate AWS image %s with AMI %s, error: %s", masterImage, awsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		masterImageNote := awsMachineConfig.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)

		securityGroupPaths := []string{}
		for _, securityGroup := range util.StringListToStringArray(ctx, diag, awsMachineConfig.SecurityGroups) {
			securityGroupPath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", securityGroup, util.SecurityGroupResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate security group %s, error: %s", securityGroup, err.Error()),
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
		var httpResp *http.Response
		var err error
		snapshot := gcpMachineConfig.MachineSnapshot.ValueString()
		imageVm := gcpMachineConfig.MasterImage.ValueString()
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", imageVm)
			imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, snapshot, util.SnapshotResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate snapshot %s of master image VM %s on GCP, error: %s", snapshot, imageVm, err.Error()),
				)
				return nil, err
			}
		} else {
			imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageVm, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate master image machine %s on GCP, error: %s", imageVm, err.Error()),
				)
				return nil, err
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		masterImageNote := gcpMachineConfig.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)

		machineProfile := gcpMachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfile, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate machine profile %s on GCP, error: %s", gcpMachineConfig.MachineProfile.ValueString(), err.Error()),
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

		if !vSphereMachineConfig.VspherePreparedImage.IsNull() {
			// Add support for prepared image
			provisioningScheme.SetPrepareImage(true)
			preparedImageConfig := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diag, vSphereMachineConfig.VspherePreparedImage)

			var assignImageVersionToProvScheme citrixorchestration.AssignImageVersionToProvisioningSchemeRequestModel
			assignImageVersionToProvScheme.SetImageDefinition(preparedImageConfig.ImageDefinition.ValueString())
			assignImageVersionToProvScheme.SetImageVersion(preparedImageConfig.ImageVersion.ValueString())
			provisioningScheme.SetAssignImageVersionToProvisioningScheme(assignImageVersionToProvScheme)
			provisioningScheme.SetResourcePool(hypervisorResourcePool.GetName()) // Override with name to adapt to image version workflow
		} else {
			image := vSphereMachineConfig.MasterImageVm.ValueString()
			snapshot := vSphereMachineConfig.ImageSnapshot.ValueString()
			resourcePoolPath := vSphereMachineConfig.ResourcePoolPath.ValueString()
			imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, resourcePoolPath, "creating")
			if err != nil {
				return nil, err
			}
			provisioningScheme.SetMasterImagePath(imagePath)

			masterImageNote := vSphereMachineConfig.MasterImageNote.ValueString()
			provisioningScheme.SetMasterImageNote(masterImageNote)
		}

		if !vSphereMachineConfig.WritebackCache.IsNull() {
			provisioningScheme.SetUseWriteBackCache(true)
			writeBackCacheModel := util.ObjectValueToTypedObject[VsphereAndSCVMMWritebackCacheModel](ctx, diag, vSphereMachineConfig.WritebackCache)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if !writeBackCacheModel.WriteBackCacheDriveLetter.IsNull() {
				provisioningScheme.SetWriteBackCacheDriveLetter(writeBackCacheModel.WriteBackCacheDriveLetter.ValueString())
			}
		}

		if !vSphereMachineConfig.MachineProfile.IsNull() {
			machineProfile, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", vSphereMachineConfig.MachineProfile.ValueString(), util.TemplateResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate machine profile %s on vSphere, error: %s", vSphereMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return nil, err
			}

			provisioningScheme.SetMachineProfilePath(machineProfile)
		}
		provisioningScheme.SetUseFullDiskCloneProvisioning(vSphereMachineConfig.UseFullDiskCloneProvisioning.ValueBool())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, diag, provisioningSchemePlan.XenserverMachineConfig)
		provisioningScheme.SetCpuCount(int32(xenserverMachineConfig.CpuCount.ValueInt64()))
		provisioningScheme.SetMemoryMB(int32(xenserverMachineConfig.MemoryMB.ValueInt64()))

		image := xenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := xenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "", "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		masterImageNote := xenserverMachineConfig.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)

		if !xenserverMachineConfig.WritebackCache.IsNull() {
			provisioningScheme.SetUseWriteBackCache(true)
			writeBackCacheModel := util.ObjectValueToTypedObject[XenserverWritebackCacheModel](ctx, diag, xenserverMachineConfig.WritebackCache)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
		provisioningScheme.SetUseFullDiskCloneProvisioning(xenserverMachineConfig.UseFullDiskCloneProvisioning.ValueBool())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
		scvmmMachineConfig := util.ObjectValueToTypedObject[SCVMMMachineConfigModel](ctx, diag, provisioningSchemePlan.SCVMMMachineConfigModel)
		provisioningScheme.SetMemoryMB(int32(scvmmMachineConfig.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(scvmmMachineConfig.CpuCount.ValueInt64()))

		image := scvmmMachineConfig.MasterImage.ValueString()
		snapshot := scvmmMachineConfig.ImageSnapshot.ValueString()
		imageResource, err := getOnPremImage(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "", "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imageResource.GetXDPath())

		masterImageNote := scvmmMachineConfig.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)

		if !scvmmMachineConfig.WritebackCache.IsNull() {
			provisioningScheme.SetUseWriteBackCache(true)
			writeBackCacheModel := util.ObjectValueToTypedObject[VsphereAndSCVMMWritebackCacheModel](ctx, diag, scvmmMachineConfig.WritebackCache)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(writeBackCacheModel.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !writeBackCacheModel.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(writeBackCacheModel.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if !writeBackCacheModel.WriteBackCacheDriveLetter.IsNull() {
				provisioningScheme.SetWriteBackCacheDriveLetter(writeBackCacheModel.WriteBackCacheDriveLetter.ValueString())
			}
		}

		networkMapping, err := getNetworkMappingForSCVMMCatalog(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), imageResource.GetRelativePath(), provisioningSchemePlan)
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetNetworkMapping(networkMapping)

		provisioningScheme.SetUseFullDiskCloneProvisioning(scvmmMachineConfig.UseFullDiskCloneProvisioning.ValueBool())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, diag, provisioningSchemePlan.NutanixMachineConfig)
		if hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
			return nil, fmt.Errorf("unsupported hypervisor plugin %s", hypervisor.GetPluginId())
		}

		provisioningScheme.SetMemoryMB(int32(nutanixMachineConfig.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(nutanixMachineConfig.CpuCount.ValueInt64()))
		provisioningScheme.SetCoresPerCpuCount(int32(nutanixMachineConfig.CoresPerCpuCount.ValueInt64()))

		imagePath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.MasterImage.ValueString(), util.TemplateResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate master image %s on NUTANIX, error: %s", nutanixMachineConfig.MasterImage.ValueString(), err.Error()),
			)
			return nil, err
		}

		containerId, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.Container.ValueString(), util.StorageResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate container %s on NUTANIX, error: %s", nutanixMachineConfig.Container.ValueString(), err.Error()),
			)
			return nil, err
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		masterImageNote := nutanixMachineConfig.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)

		customProperties := provisioningScheme.GetCustomProperties()
		util.AppendNameValueStringPair(&customProperties, "NutanixContainerId", containerId)
		provisioningScheme.SetCustomProperties(customProperties)
	}

	if !provisioningSchemePlan.NetworkMapping.IsNull() && hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM {
		networkMappingModel := util.ObjectListToTypedArray[util.NetworkMappingModel](ctx, diag, provisioningSchemePlan.NetworkMapping)
		networkMapping, err := util.ParseNetworkMappingToClientModel(networkMappingModel, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to find hypervisor network, error: %s", err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetNetworkMapping(networkMapping)
	}

	metadata := util.GetMetadataRequestModel(ctx, diag, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diag, provisioningSchemePlan.Metadata))
	provisioningScheme.SetMetadata(metadata)

	return &provisioningScheme, nil
}

func setProvisioningSchemeForMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, azureMachineConfigModel AzureMachineConfigModel, diagnostics *diag.Diagnostics, provisioningScheme *citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) error {
	if !azureMachineConfigModel.AzureMasterImage.IsNull() {
		azureMasterImageModel := util.ObjectValueToTypedObject[AzureMasterImageModel](ctx, diagnostics, azureMachineConfigModel.AzureMasterImage)
		sharedSubscription := azureMasterImageModel.SharedSubscription.ValueString()
		resourceGroup := azureMasterImageModel.ResourceGroup.ValueString()
		masterImage := azureMasterImageModel.MasterImage.ValueString()
		storageAccount := azureMasterImageModel.StorageAccount.ValueString()
		container := azureMasterImageModel.Container.ValueString()
		err := error(nil)
		imagePath, err := util.BuildAzureMasterImagePath(ctx, client, diagnostics, azureMasterImageModel.GalleryImage, sharedSubscription, resourceGroup, storageAccount, container, masterImage, hypervisor.GetName(), hypervisorResourcePool.GetName(), "Error creating Machine Catalog")
		if err != nil {
			return err
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		masterImageNote := azureMachineConfigModel.MasterImageNote.ValueString()
		provisioningScheme.SetMasterImageNote(masterImageNote)
	} else if !azureMachineConfigModel.AzurePreparedImage.IsNull() {
		preparedImageConfig := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, azureMachineConfigModel.AzurePreparedImage)
		provisioningScheme.SetPrepareImage(true)

		var assignImageVersionToProvScheme citrixorchestration.AssignImageVersionToProvisioningSchemeRequestModel
		assignImageVersionToProvScheme.SetImageDefinition(preparedImageConfig.ImageDefinition.ValueString())
		assignImageVersionToProvScheme.SetImageVersion(preparedImageConfig.ImageVersion.ValueString())
		provisioningScheme.SetAssignImageVersionToProvisioningScheme(assignImageVersionToProvScheme)
		provisioningScheme.SetResourcePool(hypervisorResourcePool.GetName()) // Override with name to adapt to image version workflow
	}

	return nil
}

func setProvisioningSchemeForPvsCatalog(ctx context.Context, azureMachineConfigModel AzureMachineConfigModel, diagnostics *diag.Diagnostics, provisioningScheme *citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel) error {
	pvsConfigurationModel := util.ObjectValueToTypedObject[AzurePvsConfigurationModel](ctx, diagnostics, azureMachineConfigModel.AzurePvsConfiguration)
	provisioningScheme.SetPVSSite(pvsConfigurationModel.PvsSiteId.ValueString())
	provisioningScheme.SetPVSVDisk(pvsConfigurationModel.PvsVdiskId.ValueString())

	return nil
}

func setProvSchemePropertiesForUpdateCatalog(provisioningSchemePlan ProvisioningSchemeModel, body citrixorchestration.UpdateMachineCatalogRequestModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, provisioningType *citrixorchestration.ProvisioningType) (citrixorchestration.UpdateMachineCatalogRequestModel, error) {
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
		serviceOfferingPath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, util.ServiceOfferingResourceType, "")
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, nil, provisioningSchemePlan.AwsMachineConfig)
		inputServiceOffering := awsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on AWS, error: %s", inputServiceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOffering)

		securityGroupPaths := []string{}
		for _, securityGroup := range util.StringListToStringArray(ctx, diagnostics, awsMachineConfig.SecurityGroups) {
			securityGroupPath, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", securityGroup, util.SecurityGroupResourceType, "")
			if err != nil {
				diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate security group %s, error: %s", securityGroup, err.Error()),
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

		if !vSphereMachineConfig.MachineProfile.IsNull() {
			machineProfile, httpResp, err := util.GetSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", vSphereMachineConfig.MachineProfile.ValueString(), util.TemplateResourceType, "")
			if err != nil {
				diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to resolve machine profile %s on vSphere, error: %s", vSphereMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return body, err
			}

			body.SetMachineProfilePath(machineProfile)
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
		scvmmMachineConfig := util.ObjectValueToTypedObject[SCVMMMachineConfigModel](ctx, nil, provisioningSchemePlan.SCVMMMachineConfigModel)
		body.SetCpuCount(int32(scvmmMachineConfig.CpuCount.ValueInt64()))
		body.SetMemoryMB(int32(scvmmMachineConfig.MemoryMB.ValueInt64()))

		image := scvmmMachineConfig.MasterImage.ValueString()
		snapshot := scvmmMachineConfig.ImageSnapshot.ValueString()
		imageResource, err := getOnPremImage(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "", "updating")
		if err != nil {
			return body, err
		}

		networkMapping, err := getNetworkMappingForSCVMMCatalog(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), imageResource.GetRelativePath(), provisioningSchemePlan)
		if err != nil {
			return body, err
		}
		body.SetNetworkMapping(networkMapping)
	}

	if !provisioningSchemePlan.NetworkMapping.IsNull() && hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM {
		networkMappingModel := util.ObjectListToTypedArray[util.NetworkMappingModel](ctx, diagnostics, provisioningSchemePlan.NetworkMapping)
		networkMapping, err := util.ParseNetworkMappingToClientModel(networkMappingModel, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to parse network mapping, error: %s", err.Error()),
			)
			return body, err
		}
		body.SetNetworkMapping(networkMapping)
	}

	customProperties, err := parseCustomPropertiesToClientModel(ctx, diagnostics, client, provisioningSchemePlan, hypervisor.ConnectionType, provisioningType, true)
	if err != nil {
		return body, err
	}
	body.SetCustomProperties(customProperties)

	return body, nil
}

func generateAdminCredentialHeader(machineDomainIdentityModel MachineDomainIdentityModel) string {
	domain := machineDomainIdentityModel.Domain.ValueString()
	if !machineDomainIdentityModel.ServiceAccountDomain.IsNull() {
		domain = machineDomainIdentityModel.ServiceAccountDomain.ValueString()
	}
	credential := fmt.Sprintf("%s\\%s:%s", domain, machineDomainIdentityModel.ServiceAccount.ValueString(), machineDomainIdentityModel.ServiceAccountPassword.ValueString())
	encodedData := base64.StdEncoding.EncodeToString([]byte(credential))
	header := fmt.Sprintf("Basic %s", encodedData)

	return header
}

func deleteMachinesFromMcsPvsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, provisioningSchemePlan ProvisioningSchemeModel, machineAccountsInPlan []MachineADAccountModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	getMachinesResponses, err := util.GetMachineCatalogMachines(ctx, client, &resp.Diagnostics, catalogId)
	if err != nil {
		return err
	}

	machineDeleteRequestCount := int(catalog.GetTotalCount()) - int(provisioningSchemePlan.NumTotalMachines.ValueInt64())
	machinesToDelete := []citrixorchestration.MachineResponseModel{}

	for _, machine := range getMachinesResponses {

		if machine.GetAllocationType() == citrixorchestration.ALLOCATIONTYPE_STATIC && len(machine.GetAssignedUsers()) > 0 {
			continue
		}

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
			errorString+" Ensure machines that need to be deleted have no active sessions. For machines with `Static` allocation type, also ensure there are no assigned users.",
		)

		return err
	}

	return deleteMachinesFromCatalog(ctx, client, resp, provisioningSchemePlan, machinesToDelete, catalogName, true, machineAccountsInPlan)
}

func addMachinesToMcsPvsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, provisioningSchemePlan ProvisioningSchemeModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	addMachinesCount := int32(provisioningSchemePlan.NumTotalMachines.ValueInt64()) - catalog.GetTotalCount()
	var addMachineRequestBody citrixorchestration.AddMachineToMachineCatalogDetailRequestModel

	if !provisioningSchemePlan.MachineAccountCreationRules.IsNull() {
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
		addMachineRequestBody.SetMachineAccountCreationRules(updateMachineAccountCreationRule)
	}

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
	relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/Machines", catalogId)

	allMachineADAccounts, err := getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, client, catalogId)
	if err != nil {
		return err
	}
	availableMachineAccounts := []string{}
	for _, machineADAccount := range allMachineADAccounts {
		if machineADAccount.GetState() == citrixorchestration.PROVISIONINGSCHEMEMACHINEACCOUNTSTATE_AVAILABLE {
			availableMachineAccounts = append(availableMachineAccounts, machineADAccount.GetSamName())
		}
	}
	if len(availableMachineAccounts) > int(addMachinesCount) {
		for i := 0; i < int(addMachinesCount); i++ {
			batchRequestItem, err := constructAddMachineWithADAccountRequestBatchItem(&resp.Diagnostics, client, relativeUrl, batchApiHeaders, catalogName, i, availableMachineAccounts)
			if err != nil {
				return err
			}
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
	} else {
		addMachinesWithRuleCount := addMachinesCount - int32(len(availableMachineAccounts))
		for i := 0; i < len(availableMachineAccounts); i++ {
			batchRequestItem, err := constructAddMachineWithADAccountRequestBatchItem(&resp.Diagnostics, client, relativeUrl, batchApiHeaders, catalogName, i, availableMachineAccounts)
			if err != nil {
				return err
			}
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}

		for i := 0; i < int(addMachinesWithRuleCount); i++ {
			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetMethod(http.MethodPost)
			batchRequestItem.SetReference(strconv.Itoa(i))
			batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetBody(addMachineRequestStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
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

func updateCatalogImageAndMachineProfile(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan MachineCatalogResourceModel, provisioningType *citrixorchestration.ProvisioningType) error {

	catalogName := catalog.GetName()
	catalogId := catalog.GetId()

	provScheme := catalog.GetProvisioningScheme()
	masterImage := provScheme.GetMasterImage()
	currentDiskImage := provScheme.GetCurrentDiskImage()
	machineProfile := provScheme.GetMachineProfile()
	customProps := provScheme.GetCustomProperties()
	currentImageDetails := provScheme.GetCurrentImageVersion()

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
	masterImageNote := ""
	machineProfilePath := ""
	usePreparedImage := false
	imageDefinition := ""
	imageVersion := ""
	currentImageDefinition := ""
	currentImageVersion := ""
	preparedImageUseSharedGallery := false
	var err error
	var httpResp *http.Response
	updateCustomProperties := []citrixorchestration.NameValueStringPairModel{}

	// Set default reboot options
	var rebootOption citrixorchestration.RebootMachinesRequestModel
	rebootOption.SetRebootDuration(-1) // Default to update on next shutdown
	rebootOption.SetWarningDuration(0) // Default to no warning

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.AzureMachineConfig)
		azureMachineProfile := azureMachineConfigModel.MachineProfile
		if !(*provisioningType == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING) {
			if !azureMachineConfigModel.AzureMasterImage.IsNull() {
				azureMasterImageModel := util.ObjectValueToTypedObject[AzureMasterImageModel](ctx, &resp.Diagnostics, azureMachineConfigModel.AzureMasterImage)
				sharedSubscription := azureMasterImageModel.SharedSubscription.ValueString()
				newImage := azureMasterImageModel.MasterImage.ValueString()
				resourceGroup := azureMasterImageModel.ResourceGroup.ValueString()
				imageBasePath := "image.folder"
				if sharedSubscription != "" {
					imageBasePath = fmt.Sprintf("image.folder\\%s.sharedsubscription", sharedSubscription)
				}

				if newImage != "" {
					storageAccount := azureMasterImageModel.StorageAccount.ValueString()
					container := azureMasterImageModel.Container.ValueString()
					if storageAccount != "" && container != "" {
						queryPath := fmt.Sprintf(
							"%s\\%s.resourcegroup\\%s.storageaccount\\%s.container",
							imageBasePath,
							resourceGroup,
							storageAccount,
							container)
						imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
						if err != nil {
							resp.Diagnostics.AddError(
								"Error updating Machine Catalog",
								"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
									fmt.Sprintf("\nFailed to resolve master image VHD %s in container %s of storage account %s, error: %s", newImage, container, storageAccount, err.Error()),
							)
							return err
						}
					} else {
						queryPath := fmt.Sprintf(
							"%s\\%s.resourcegroup",
							imageBasePath,
							resourceGroup)
						imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, newImage, "", "")
						if err != nil {
							resp.Diagnostics.AddError(
								"Error updating Machine Catalog",
								"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
									fmt.Sprintf("\nFailed to resolve master image Managed Disk or Snapshot %s, error: %s", newImage, err.Error()),
							)
							return err
						}
					}
				} else if !azureMasterImageModel.GalleryImage.IsNull() {
					azureGalleryImage := util.ObjectValueToTypedObject[util.GalleryImageModel](ctx, &resp.Diagnostics, azureMasterImageModel.GalleryImage)
					gallery := azureGalleryImage.Gallery.ValueString()
					definition := azureGalleryImage.Definition.ValueString()
					version := azureGalleryImage.Version.ValueString()
					if gallery != "" && definition != "" {
						queryPath := fmt.Sprintf(
							"%s\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
							imageBasePath,
							resourceGroup,
							gallery,
							definition)
						imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
						if err != nil {
							resp.Diagnostics.AddError(
								"Error updating Machine Catalog",
								"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
									fmt.Sprintf("\nFailed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", newImage, version, gallery, err.Error()),
							)
							return err
						}
					}
				}

				masterImageNote = azureMachineConfigModel.MasterImageNote.ValueString()
			} else if !azureMachineConfigModel.AzurePreparedImage.IsNull() {
				// Handle prepared image
				usePreparedImage = true
				preparedImageModel := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, &resp.Diagnostics, azureMachineConfigModel.AzurePreparedImage)
				currentImageVersionDetails := currentImageDetails.GetImageVersion()
				currentImageDefinitionDetails := currentImageVersionDetails.GetImageDefinition()
				imageDefinition = preparedImageModel.ImageDefinition.ValueString()
				imageVersion = preparedImageModel.ImageVersion.ValueString()
				currentImageVersion = currentImageVersionDetails.GetId()
				currentImageDefinition = currentImageDefinitionDetails.GetId()
				imageDefinitionResp, err := image_definition.GetImageDefinition(ctx, client, &resp.Diagnostics, imageDefinition)
				if err != nil {
					return err
				}
				preparedImageUseSharedGallery = IsAzureImageDefinitionUsingSharedImageGallery(imageDefinitionResp)
			}

			// Set reboot options if configured
			if !azureMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
				rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, azureMachineConfigModel.ImageUpdateRebootOptions)
				rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
				warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
				rebootOption.SetWarningDuration(warningDuration)
				if warningDuration > 0 || warningDuration == -1 {
					// if warning duration is not 0, it's set in plan and requires warning message body
					rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
					if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
						rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
					}
				}
			}

			updateCustomProperties = appendUseAzureComputeGalleryCustomProperties(ctx, &resp.Diagnostics, updateCustomProperties, azureMachineConfigModel, preparedImageUseSharedGallery)
		}

		// For both MCS and PVS
		if !azureMachineProfile.IsNull() {
			machineProfilePath, err = util.HandleMachineProfileForAzureMcsPvsCatalog(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), util.ObjectValueToTypedObject[util.AzureMachineProfileModel](ctx, &resp.Diagnostics, azureMachineProfile), "Error updating Machine Catalog")
			if err != nil {
				return err
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.AwsMachineConfig)
		imageId := fmt.Sprintf("%s (%s)", awsMachineConfig.MasterImage.ValueString(), awsMachineConfig.ImageAmi.ValueString())
		imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate AWS image %s with AMI %s, error: %s", awsMachineConfig.MasterImage.ValueString(), awsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return err
		}

		masterImageNote = awsMachineConfig.MasterImageNote.ValueString()

		// Set reboot options if configured
		if !awsMachineConfig.ImageUpdateRebootOptions.IsNull() {
			rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, awsMachineConfig.ImageUpdateRebootOptions)
			rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
			warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
			rebootOption.SetWarningDuration(warningDuration)
			if warningDuration > 0 || warningDuration == -1 {
				// if warning duration is not 0, it's set in plan and requires warning message body
				rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
				if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
					rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.GcpMachineConfig)
		newImage := gcpMachineConfig.MasterImage.ValueString()
		snapshot := gcpMachineConfig.MachineSnapshot.ValueString()
		gcpMachineProfile := gcpMachineConfig.MachineProfile.ValueString()

		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", newImage)
			imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, snapshot, util.SnapshotResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate snapshot %s of master image %s on GCP, error: %s", snapshot, newImage, err.Error()),
				)
				return err
			}
		} else {
			imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", newImage, util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate master image machine %s on GCP, error: %s", newImage, err.Error()),
				)
				return err
			}
		}

		masterImageNote = gcpMachineConfig.MasterImageNote.ValueString()

		// Set reboot options if configured
		if !gcpMachineConfig.ImageUpdateRebootOptions.IsNull() {
			rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, gcpMachineConfig.ImageUpdateRebootOptions)
			rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
			warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
			rebootOption.SetWarningDuration(warningDuration)
			if warningDuration > 0 || warningDuration == -1 {
				// if warning duration is not 0, it's set in plan and requires warning message body
				rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
				if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
					rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
				}
			}
		}

		if gcpMachineProfile != "" {
			machineProfilePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", gcpMachineConfig.MachineProfile.ValueString(), util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate machine profile %s on GCP, error: %s", gcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.VsphereMachineConfig)
		if !vSphereMachineConfig.VspherePreparedImage.IsNull() {
			usePreparedImage = true
			preparedImageModel := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, &resp.Diagnostics, vSphereMachineConfig.VspherePreparedImage)
			currentImageVersionDetails := currentImageDetails.GetImageVersion()
			currentImageDefinitionDetails := currentImageVersionDetails.GetImageDefinition()
			imageDefinition = preparedImageModel.ImageDefinition.ValueString()
			imageVersion = preparedImageModel.ImageVersion.ValueString()
			currentImageVersion = currentImageVersionDetails.GetId()
			currentImageDefinition = currentImageDefinitionDetails.GetId()
		} else {
			newImage := vSphereMachineConfig.MasterImageVm.ValueString()
			snapshot := vSphereMachineConfig.ImageSnapshot.ValueString()
			resourcePoolPath := vSphereMachineConfig.ResourcePoolPath.ValueString()
			imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, resourcePoolPath, "updating")
			if err != nil {
				return err
			}

			masterImageNote = vSphereMachineConfig.MasterImageNote.ValueString()
		}
		// Set reboot options if configured
		if !vSphereMachineConfig.ImageUpdateRebootOptions.IsNull() {
			rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, vSphereMachineConfig.ImageUpdateRebootOptions)
			rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
			warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
			rebootOption.SetWarningDuration(warningDuration)
			if warningDuration > 0 || warningDuration == -1 {
				// if warning duration is not 0, it's set in plan and requires warning message body
				rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
				if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
					rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.XenserverMachineConfig)
		newImage := xenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := xenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "", "updating")
		if err != nil {
			return err
		}

		masterImageNote = xenserverMachineConfig.MasterImageNote.ValueString()

		// Set reboot options if configured
		if !xenserverMachineConfig.ImageUpdateRebootOptions.IsNull() {
			rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, xenserverMachineConfig.ImageUpdateRebootOptions)
			rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
			warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
			rebootOption.SetWarningDuration(warningDuration)
			if warningDuration > 0 || warningDuration == -1 {
				// if warning duration is not 0, it's set in plan and requires warning message body
				rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
				if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
					rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
		scvmmMachineConfig := util.ObjectValueToTypedObject[SCVMMMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.SCVMMMachineConfigModel)
		newImage := scvmmMachineConfig.MasterImage.ValueString()
		snapshot := scvmmMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "", "updating")
		if err != nil {
			return err
		}

		masterImageNote = scvmmMachineConfig.MasterImageNote.ValueString()

		// Set reboot options if configured
		if !scvmmMachineConfig.ImageUpdateRebootOptions.IsNull() {
			rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, scvmmMachineConfig.ImageUpdateRebootOptions)
			rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
			warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
			rebootOption.SetWarningDuration(warningDuration)
			if warningDuration > 0 || warningDuration == -1 {
				// if warning duration is not 0, it's set in plan and requires warning message body
				rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
				if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
					rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
				}
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, &resp.Diagnostics, provisioningSchemePlan.NutanixMachineConfig)
		if hypervisor.GetPluginId() == util.NUTANIX_PLUGIN_ID {
			imagePath, httpResp, err = util.GetSingleResourcePathFromHypervisor(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", nutanixMachineConfig.MasterImage.ValueString(), util.TemplateResourceType, "")

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate master image %s on NUTANIX, error: %s", nutanixMachineConfig.MasterImage.ValueString(), err.Error()),
				)
				return err
			}

			masterImageNote = nutanixMachineConfig.MasterImageNote.ValueString()

			// Set reboot options if configured
			if !nutanixMachineConfig.ImageUpdateRebootOptions.IsNull() {
				rebootOptionsPlan := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, nutanixMachineConfig.ImageUpdateRebootOptions)
				rebootOption.SetRebootDuration(int32(rebootOptionsPlan.RebootDuration.ValueInt64()))
				warningDuration := int32(rebootOptionsPlan.WarningDuration.ValueInt64())
				rebootOption.SetWarningDuration(warningDuration)
				if warningDuration > 0 || warningDuration == -1 {
					// if warning duration is not 0, it's set in plan and requires warning message body
					rebootOption.SetWarningMessage(rebootOptionsPlan.WarningMessage.ValueString())
					if !rebootOptionsPlan.WarningRepeatInterval.IsNull() {
						rebootOption.SetWarningRepeatInterval(int32(rebootOptionsPlan.WarningRepeatInterval.ValueInt64()))
					}
				}
			}
		}
	}

	if machineProfile.GetXDPath() != machineProfilePath {
		err = updateCatalogMachineProfile(ctx, client, resp, catalog, machineProfilePath)
		if err != nil {
			return err
		}
	}

	// Updating image is not supported for PVSStreaming catalog
	if !(*provisioningType == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING) {

		replicaRatio, replicaMaximum, useSharedGallery := setComputeGalleryValues(customProps)

		updateReplicaRatio, updateReplicaMaximum, updateUseSharedGallery := setComputeGalleryValues(updateCustomProperties)

		if masterImage.GetXDPath() == imagePath && currentDiskImage.GetMasterImageNote() == masterImageNote && updateReplicaRatio == replicaRatio && updateReplicaMaximum == replicaMaximum && updateUseSharedGallery == useSharedGallery && currentImageDefinition == imageDefinition && currentImageVersion == imageVersion {
			return nil
		}

		// Update Master Image for Machine Catalog
		var updateProvisioningSchemeModel citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel

		functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog "+catalogName,
				fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
			)
			return err
		}

		updateProvisioningSchemeModel.SetMinimumFunctionalLevel(*functionalLevel)
		updateProvisioningSchemeModel.SetStoreOldImage(true)

		if !usePreparedImage {
			updateProvisioningSchemeModel.SetMasterImagePath(imagePath)
			updateProvisioningSchemeModel.SetMasterImageNote(masterImageNote)
		} else {
			var assignImageVersionToProvScheme citrixorchestration.AssignImageVersionToProvisioningSchemeRequestModel
			assignImageVersionToProvScheme.SetImageDefinition(imageDefinition)
			assignImageVersionToProvScheme.SetImageVersion(imageVersion)
			updateProvisioningSchemeModel.SetAssignImageVersionToProvisioningScheme(assignImageVersionToProvScheme)
		}
		updateProvisioningSchemeModel.SetRebootOptions(rebootOption)

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
	}

	return nil
}

func setComputeGalleryValues(customProperties []citrixorchestration.NameValueStringPairModel) (string, string, string) {
	replicaRatio := ""
	replicaMaximum := ""
	useSharedGallery := "false"
	for _, customProp := range customProperties {
		if customProp.GetName() == util.ReplicaRatio {
			replicaRatio = customProp.GetValue()
		}
		if customProp.GetName() == util.ReplicaMaximum {
			replicaMaximum = customProp.GetValue()
		}
		if customProp.GetName() == util.SharedGallery {
			useSharedGallery = customProp.GetValue()
		}
	}
	return replicaRatio, replicaMaximum, useSharedGallery
}

func (r MachineCatalogResourceModel) updateCatalogWithProvScheme(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, pluginId string, provScheme citrixorchestration.ProvisioningSchemeResponseModel, machineAdAccounts []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel) MachineCatalogResourceModel {
	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, r.ProvisioningScheme)
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
		provisioningType, _ := citrixorchestration.NewProvisioningTypeFromValue(r.ProvisioningType.ValueString())
		azureMachineConfigModel.RefreshProperties(ctx, diagnostics, *catalog, provisioningType)
		provSchemeModel.AzureMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, azureMachineConfigModel)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" && stringPair.GetValue() != "" {
				availability_zones := strings.Split(stringPair.GetValue(), ",")
				provSchemeModel.AvailabilityZones = util.StringArrayToStringList(ctx, diagnostics, availability_zones)
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		awsMachineConfig := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, diagnostics, provSchemeModel.AwsMachineConfig)
		if !provSchemeModel.AwsMachineConfig.IsNull() {
			if serviceOfferingObject, httpResp, err := util.GetSingleResourceFromHypervisor(ctx, client, diagnostics, hypervisor.GetId(), resourcePool.GetId(), "", provScheme.GetServiceOffering(), util.ServiceOfferingResourceType, ""); err == nil {
				provScheme.SetServiceOffering(serviceOfferingObject.GetId())
				catalog.SetProvisioningScheme(provScheme)
			} else {
				diagnostics.AddError(
					"Error updating Machine Catalog "+catalog.GetName(),
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to resolve AWS service offering %s, error: %s", provScheme.GetServiceOffering(), err.Error()),
				)
			}
		}
		awsMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.AwsMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, awsMachineConfig)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "Zones" {
				availability_zones := strings.Split(stringPair.GetValue(), ",")
				provSchemeModel.AvailabilityZones = util.StringArrayToStringList(ctx, diagnostics, availability_zones)
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, diagnostics, provSchemeModel.GcpMachineConfig)
		gcpMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.GcpMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, gcpMachineConfig)
		for _, stringPair := range customProperties {
			if stringPair.GetName() == "CatalogZones" && stringPair.GetValue() != "" {
				availability_zones := strings.Split(stringPair.GetValue(), ",")
				provSchemeModel.AvailabilityZones = util.StringArrayToStringList(ctx, diagnostics, availability_zones)
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		vSphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, diagnostics, provSchemeModel.VsphereMachineConfig)
		vSphereMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.VsphereMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, vSphereMachineConfig)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		xenserverMachineConfig := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, diagnostics, provSchemeModel.XenserverMachineConfig)
		xenserverMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.XenserverMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, xenserverMachineConfig)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
		scvmmMachineConfig := util.ObjectValueToTypedObject[SCVMMMachineConfigModel](ctx, diagnostics, provSchemeModel.SCVMMMachineConfigModel)
		scvmmMachineConfig.RefreshProperties(ctx, diagnostics, *catalog)
		provSchemeModel.SCVMMMachineConfigModel = util.TypedObjectToObjectValue(ctx, diagnostics, scvmmMachineConfig)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		if pluginId == util.NUTANIX_PLUGIN_ID {
			nutanixMachineConfig := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, diagnostics, provSchemeModel.NutanixMachineConfig)
			nutanixMachineConfig.RefreshProperties(*catalog)
			provSchemeModel.NutanixMachineConfig = util.TypedObjectToObjectValue(ctx, diagnostics, nutanixMachineConfig)
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
		provSchemeModel.CustomProperties = util.TypedArrayToObjectList[CustomPropertyModel](ctx, diagnostics, nil)
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
		provSchemeModel.NetworkMapping = util.RefreshListValueProperties[util.NetworkMappingModel, citrixorchestration.NetworkMapResponseModel](ctx, diagnostics, provSchemeModel.NetworkMapping, networkMaps, util.GetOrchestrationNetworkMappingKey)
	} else {
		provSchemeModel.NetworkMapping = util.TypedArrayToObjectList[util.NetworkMappingModel](ctx, diagnostics, nil)
	}

	// Identity Pool Properties
	if provScheme.MachineAccountCreationRules != nil {
		machineAccountCreationRulesModel := MachineAccountCreationRulesModel{}
		machineAccountCreationRulesModel.NamingScheme = types.StringValue(machineAccountCreateRules.GetNamingScheme())
		namingSchemeType := machineAccountCreateRules.GetNamingSchemeType()
		machineAccountCreationRulesModel.NamingSchemeType = types.StringValue(string(namingSchemeType))
		provSchemeModel.MachineAccountCreationRules = util.TypedObjectToObjectValue(ctx, diagnostics, machineAccountCreationRulesModel)
	}

	// Refresh machine AD Accounts
	machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, diagnostics, provSchemeModel.MachineADAccounts)
	refreshedMachineAccounts := []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel{}
	for _, account := range machineAdAccounts {
		if slices.ContainsFunc(machineAccountsInPlan, func(accountInPlan MachineADAccountModel) bool {
			return strings.EqualFold(accountInPlan.ADAccountName.ValueString(), account.GetSamName())
		}) {
			refreshedMachineAccounts = append(refreshedMachineAccounts, account)
		}
	}
	if len(refreshedMachineAccounts) > 0 {
		provSchemeModel.MachineADAccounts = util.RefreshListValueProperties[MachineADAccountModel, citrixorchestration.ProvisioningSchemeMachineAccountResponseModel](ctx, diagnostics, provSchemeModel.MachineADAccounts, refreshedMachineAccounts, util.GetMachineAdAccountKey)
	} else {
		attrMap, err := util.ResourceAttributeMapFromObject(MachineADAccountModel{})
		if err != nil {
			diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
			return r
		}
		provSchemeModel.MachineADAccounts = types.ListNull(types.ObjectType{AttrTypes: attrMap})
	}

	// Domain Identity Properties
	if provScheme.GetIdentityType() == citrixorchestration.IDENTITYTYPE_AZURE_AD ||
		provScheme.GetIdentityType() == citrixorchestration.IDENTITYTYPE_WORKGROUP {
		r.ProvisioningScheme = util.TypedObjectToObjectValue(ctx, diagnostics, provSchemeModel)
		return r
	}

	machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, diagnostics, provSchemeModel.MachineDomainIdentity)

	if domain.GetName() != "" && !strings.EqualFold(domain.GetName(), machineDomainIdentityModel.Domain.ValueString()) {
		machineDomainIdentityModel.Domain = types.StringValue(domain.GetName())
	}
	if machineAccountCreateRules.GetOU() != "" {
		machineDomainIdentityModel.Ou = types.StringValue(machineAccountCreateRules.GetOU())
	} else {
		machineDomainIdentityModel.Ou = types.StringNull()
	}

	provSchemeModel.MachineDomainIdentity = util.TypedObjectToObjectValue(ctx, diagnostics, machineDomainIdentityModel)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, provSchemeModel.Metadata), provScheme.GetMetadata())

	if len(effectiveMetadata) > 0 {
		provSchemeModel.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, provSchemeModel.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		provSchemeModel.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.ProvisioningScheme = util.TypedObjectToObjectValue(ctx, diagnostics, provSchemeModel)
	return r
}

func parseCustomPropertiesToClientModel(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType, provisioningType *citrixorchestration.ProvisioningType, isUpdateOperation bool) ([]citrixorchestration.NameValueStringPairModel, error) {
	var res = &[]citrixorchestration.NameValueStringPairModel{}
	var isPvsStreamingCatalog = *provisioningType == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING
	switch connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diagnostics, provisioningScheme.AzureMachineConfig)
		if !provisioningScheme.AvailabilityZones.IsNull() {
			availability_zones := util.StringListToStringArray(ctx, diagnostics, provisioningScheme.AvailabilityZones)
			util.AppendNameValueStringPair(res, "Zones", strings.Join(availability_zones, ","))
		} else {
			util.AppendNameValueStringPair(res, "Zones", "")
		}
		if !isUpdateOperation || !isPvsStreamingCatalog {
			if !azureMachineConfigModel.StorageType.IsNull() {
				if azureMachineConfigModel.StorageType.ValueString() == util.AzureEphemeralOSDisk {
					util.AppendNameValueStringPair(res, "UseEphemeralOsDisk", "true")
				} else {
					util.AppendNameValueStringPair(res, "StorageType", azureMachineConfigModel.StorageType.ValueString())
				}
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
					if !isPvsStreamingCatalog && azureWbcModel.StorageCostSaving.ValueBool() {
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
			if !isPvsStreamingCatalog {
				preparedImageUseSharedGallery := false
				if !azureMachineConfigModel.AzurePreparedImage.IsNull() {
					// Handle prepared image
					preparedImageModel := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, azureMachineConfigModel.AzurePreparedImage)
					imageDefinition := preparedImageModel.ImageDefinition.ValueString()

					imageDefinitionResp, err := image_definition.GetImageDefinition(ctx, client, diagnostics, imageDefinition)
					if err != nil {
						return nil, err
					}
					preparedImageUseSharedGallery = IsAzureImageDefinitionUsingSharedImageGallery(imageDefinitionResp)
				}

				*res = appendUseAzureComputeGalleryCustomProperties(ctx, diagnostics, *res, azureMachineConfigModel, preparedImageUseSharedGallery)
			}
		}

		if !azureMachineConfigModel.VdaResourceGroup.IsNull() {
			util.AppendNameValueStringPair(res, "ResourceGroups", azureMachineConfigModel.VdaResourceGroup.ValueString())
		}

		licenseType := azureMachineConfigModel.LicenseType.ValueString()
		util.AppendNameValueStringPair(res, "LicenseType", licenseType)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if !provisioningScheme.AvailabilityZones.IsNull() {
			availability_zones := util.StringListToStringArray(ctx, diagnostics, provisioningScheme.AvailabilityZones)
			util.AppendNameValueStringPair(res, "Zones", strings.Join(availability_zones, ","))
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		gcpMachineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](context.Background(), nil, provisioningScheme.GcpMachineConfig)
		if !provisioningScheme.AvailabilityZones.IsNull() {
			availability_zones := util.StringListToStringArray(ctx, diagnostics, provisioningScheme.AvailabilityZones)
			util.AppendNameValueStringPair(res, "CatalogZones", strings.Join(availability_zones, ","))
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
		return nil, nil
	}

	if len(*res) == 0 {
		return nil, nil
	}

	return *res, nil
}

func getOnPremImage(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisorName, resourcePoolName, image, snapshot, resourcePoolPath, action string) (*citrixorchestration.HypervisorResourceResponseModel, error) {
	queryPath := ""
	if resourcePoolPath != "" {
		resourcePoolSegments := strings.Split(resourcePoolPath, "/")
		for _, resourcePool := range resourcePoolSegments {
			queryPath = queryPath + resourcePool + ".resourcepool" + "\\"
		}
	}
	resourceType := util.VirtualMachineResourceType
	resourceName := image
	errTemplate := fmt.Sprintf("Failed to locate master image machine %s", image)
	if snapshot != "" {
		queryPath = queryPath + fmt.Sprintf("%s.vm", image)
		snapshotSegments := strings.Split(snapshot, "/")
		snapshotName := snapshotSegments[len(snapshotSegments)-1]
		for i := 0; i < len(snapshotSegments)-1; i++ {
			queryPath = queryPath + "\\" + snapshotSegments[i] + ".snapshot"
		}

		resourceType = util.SnapshotResourceType
		resourceName = snapshotName
		errTemplate = fmt.Sprintf("Failed to locate snapshot %s of master image VM %s", snapshotName, image)
	}

	imageResource, httpResp, err := util.GetSingleResourceFromHypervisor(ctx, client, diags, hypervisorName, resourcePoolName, queryPath, resourceName, resourceType, "")
	if err != nil {
		diags.AddError(
			fmt.Sprintf("Error %s Machine Catalog", action),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\n%s, error: %s", errTemplate, err.Error()),
		)
		return nil, err
	}

	return imageResource, nil
}

func getOnPremImagePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisorName, resourcePoolName, image, snapshot, resourcePoolPath, action string) (string, error) {
	imageResource, err := getOnPremImage(ctx, client, diags, hypervisorName, resourcePoolName, image, snapshot, resourcePoolPath, action)
	if err != nil {
		return "", err
	}

	return imageResource.GetXDPath(), nil
}

func getNetworkMappingForSCVMMCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, hypervisorName, hypervisorResourcePoolName, imageVmName string, provisioningSchemePlan ProvisioningSchemeModel) ([]citrixorchestration.NetworkMapRequestModel, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolResources(ctx, hypervisorName, hypervisorResourcePoolName).Children(0).Path(imageVmName).Detail(true)

	result, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return nil, err
	}
	vmNetworkMappings := result.GetNetworkMappings()

	networkMappingsRequest := []citrixorchestration.NetworkMapRequestModel{}
	if !provisioningSchemePlan.NetworkMapping.IsNull() {
		networkMappings := util.ObjectListToTypedArray[util.NetworkMappingModel](ctx, diag, provisioningSchemePlan.NetworkMapping)
		for _, networkMapping := range networkMappings {
			found := false
			for _, vmNetworkMapping := range vmNetworkMappings {
				vmNetwork := vmNetworkMapping.GetNetwork()

				if strings.EqualFold(vmNetwork.GetName(), networkMapping.Network.ValueString()) &&
					strings.EqualFold(vmNetworkMapping.GetDeviceId(), networkMapping.NetworkDevice.ValueString()) {
					found = true
					networkMappingsRequestModel := citrixorchestration.NetworkMapRequestModel{}
					networkMappingsRequestModel.SetNetworkPath(vmNetwork.GetXDPath())
					networkMappingsRequestModel.SetDeviceNameOrId(vmNetwork.GetName())
					networkMappingsRequestModel.SetNetworkDeviceNameOrId(vmNetworkMapping.GetDeviceId())
					networkMappingsRequest = append(networkMappingsRequest, networkMappingsRequestModel)
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("network %s not found", networkMapping.Network.ValueString())
			}
		}
	} else {
		vmNetwork := vmNetworkMappings[0].GetNetwork()
		networkMappingsRequestModel := citrixorchestration.NetworkMapRequestModel{}
		networkMappingsRequestModel.SetNetworkPath(vmNetwork.GetXDPath())
		networkMappingsRequestModel.SetDeviceNameOrId(vmNetwork.GetName())
		networkMappingsRequestModel.SetNetworkDeviceNameOrId(vmNetworkMappings[0].GetDeviceId())
		networkMappingsRequest = append(networkMappingsRequest, networkMappingsRequestModel)
	}

	return networkMappingsRequest, nil
}

func constructAvailableMachineAccountsRequestModel(diagnostics *diag.Diagnostics, machineADAccounts []MachineADAccountModel, machineCatalogAction string) ([]citrixorchestration.MachineAccountRequestModel, error) {
	availableMachineAccounts := []citrixorchestration.MachineAccountRequestModel{}
	for _, account := range machineADAccounts {
		resetPassword := account.ResetPassword.ValueBool()
		availableAccounts := citrixorchestration.MachineAccountRequestModel{}
		availableAccounts.SetADAccountName(account.ADAccountName.ValueString())
		passwordFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(account.PasswordFormat.ValueString())
		if err != nil {
			diagnostics.AddError(
				fmt.Sprintf("Error %s Machine Catalog", machineCatalogAction),
				fmt.Sprintf("Unsupported machine account password format type. Error: %s", err.Error()),
			)
			return nil, err
		}
		availableAccounts.SetPasswordFormat(*passwordFormat)
		availableAccounts.SetResetPassword(resetPassword)
		if !resetPassword {
			availableAccounts.SetPassword(account.Password.ValueString())
		}
		availableMachineAccounts = append(availableMachineAccounts, availableAccounts)
	}
	return availableMachineAccounts, nil
}

func constructAddMachineWithADAccountRequestBatchItem(diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, relativeUrl string, batchApiHeaders []citrixorchestration.NameValueStringPairModel, catalogName string, index int, availableMachineAccounts []string) (citrixorchestration.BatchRequestItemModel, error) {
	addAdMachineRequestBody := citrixorchestration.MachineAccountRequestModel{}
	addAdMachineRequestBody.SetADAccountName(availableMachineAccounts[index])
	addAdMachineBody := citrixorchestration.AddMachineToMachineCatalogDetailRequestModel{}
	addAdMachineBody.SetAddAvailableMachineAccount(addAdMachineRequestBody)
	var batchRequestItem citrixorchestration.BatchRequestItemModel

	addAdMachineStringBody, err := util.ConvertToString(addAdMachineBody)
	if err != nil {
		diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"An unexpected error occurred: "+err.Error(),
		)
		return batchRequestItem, err
	}
	batchRequestItem.SetMethod(http.MethodPost)
	batchRequestItem.SetReference("adMachine" + strconv.Itoa(index))
	batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
	batchRequestItem.SetBody(addAdMachineStringBody)
	batchRequestItem.SetHeaders(batchApiHeaders)
	return batchRequestItem, nil
}

func setMachinesToMaintenanceMode(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, catalogId string, provSchemeModel ProvisioningSchemeModel, machines []citrixorchestration.MachineResponseModel) error {
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, diagnostics, client, provSchemeModel, false)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		diagnostics.AddError(
			"Error setting Machine(s) to maintenance mode for Machine Catalog "+catalogId,
			"TransactionId: "+txId+
				"\nCould not put machine(s) into maintenance mode, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}

	for index, machineAccountInCatalog := range machines {
		if machineAccountInCatalog.DeliveryGroup == nil {
			// if machine has no delivery group, there is no need to put it in maintenance mode
			continue
		}
		isMachineInMaintenanceMode := machineAccountInCatalog.GetInMaintenanceMode()
		if !isMachineInMaintenanceMode {
			// machine is not in maintenance mode. Put machine in maintenance mode first before deleting
			var updateMachineModel citrixorchestration.UpdateMachineRequestModel
			updateMachineModel.SetInMaintenanceMode(true)
			updateMachineStringBody, err := util.ConvertToString(updateMachineModel)
			if err != nil {
				diagnostics.AddError(
					"Error setting Machine(s) to maintenance mode for Machine Catalog "+catalogId,
					"An unexpected error occurred: "+err.Error(),
				)
				return err
			}
			relativeUrl := fmt.Sprintf("/Machines/%s", machineAccountInCatalog.GetId())

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(strconv.Itoa(index))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetBody(updateMachineStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
	}

	if len(batchRequestItems) > 0 {
		// If there are any machines that need to be put in maintenance mode
		var batchRequestModel citrixorchestration.BatchRequestModel
		batchRequestModel.SetItems(batchRequestItems)
		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
		if err != nil {
			diagnostics.AddError(
				"Error setting Machine(s) to maintenance mode for Machine Catalog "+catalogId,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return err
		}

		if successfulJobs < len(batchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while putting machine(s) into maintenance mode. %d of %d machines were put in the maintenance mode.", successfulJobs, len(batchRequestItems))
			err = fmt.Errorf(errMsg)
			diagnostics.AddError(
				"Error setting Machine(s) to maintenance mode for Machine Catalog "+catalogId,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return err
		}
	}
	return nil
}

func validateMcsPvsMachineCatalogDeleteOptions(diagnostics *diag.Diagnostics, data MachineCatalogResourceModel) error {
	deleteVirtualMachines := data.DeleteVirtualMachines.ValueBool()
	if data.DeleteVirtualMachines.IsNull() {
		deleteVirtualMachines = true
	}
	deleteMachineAccounts := data.DeleteMachineAccounts.ValueString()
	if data.DeleteMachineAccounts.IsNull() {
		deleteMachineAccounts = string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE)
	}
	persistUserChanges := data.PersistUserChanges.ValueString() == "OnLocal"
	provisioningType := data.ProvisioningType.ValueString()

	if provisioningType != string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING) &&
		data.SessionSupport.ValueString() == string(citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION) &&
		data.AllocationType.ValueString() == string(citrixorchestration.ALLOCATIONTYPE_STATIC) {
		persistUserChanges = true
	}

	if !persistUserChanges && !deleteVirtualMachines {
		err := fmt.Errorf("`delete_virtual_machines` cannot be set to `false` when `persist_user_changes` is set to `Discard`")
		diagnostics.AddAttributeError(
			path.Root("delete_virtual_machines"),
			"Incorrect Attribute Configuration",
			err.Error(),
		)
		return err
	}

	if !deleteVirtualMachines && deleteMachineAccounts != string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE) {
		err := fmt.Errorf("`delete_machine_accounts` can only be set to `%s` when `delete_virtual_machines` is set to `false`", string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE))
		diagnostics.AddAttributeError(
			path.Root("delete_machine_accounts"),
			"Incorrect Attribute Configuration",
			err.Error(),
		)
		return err
	}

	return nil
}

func appendUseAzureComputeGalleryCustomProperties(ctx context.Context, diagnostics *diag.Diagnostics, updateCustomProperties []citrixorchestration.NameValueStringPairModel, azureMachineConfigModel AzureMachineConfigModel, preparedImageUseSharedGallery bool) []citrixorchestration.NameValueStringPairModel {
	if !azureMachineConfigModel.UseAzureComputeGallery.IsNull() && !preparedImageUseSharedGallery {
		azureComputeGalleryModel := util.ObjectValueToTypedObject[AzureComputeGallerySettings](ctx, diagnostics, azureMachineConfigModel.UseAzureComputeGallery)
		util.AppendNameValueStringPair(&updateCustomProperties, "UseSharedImageGallery", "true")
		util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaRatio", strconv.Itoa(int(azureComputeGalleryModel.ReplicaRatio.ValueInt64())))
		util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaMaximum", strconv.Itoa(int(azureComputeGalleryModel.ReplicaMaximum.ValueInt64())))
	} else if preparedImageUseSharedGallery {
		util.AppendNameValueStringPair(&updateCustomProperties, "UseSharedImageGallery", "true")
	} else {
		util.AppendNameValueStringPair(&updateCustomProperties, "UseSharedImageGallery", "false")
		util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaRatio", "")
		util.AppendNameValueStringPair(&updateCustomProperties, "SharedImageGalleryReplicaMaximum", "")
	}

	return updateCustomProperties
}
