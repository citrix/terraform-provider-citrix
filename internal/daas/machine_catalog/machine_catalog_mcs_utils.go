// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func getProvSchemeForMcsCatalog(plan MachineCatalogResourceModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, isOnPremises bool) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {
	if plan.ProvisioningScheme.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_AZURE_AD) {
		if isOnPremises {
			diagnostics.AddAttributeError(
				path.Root("identity_type"),
				"Unsupported Machine Catalog Configuration",
				fmt.Sprintf("Identity type %s is not supported in OnPremises environment. ", string(citrixorchestration.IDENTITYTYPE_AZURE_AD)),
			)

			return nil, fmt.Errorf("identity type %s is not supported in OnPremises environment. ", string(citrixorchestration.IDENTITYTYPE_AZURE_AD))
		}
	}

	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString())
	if err != nil {
		return nil, err
	}

	hypervisorResourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString(), plan.ProvisioningScheme.HypervisorResourcePool.ValueString())
	if err != nil {
		return nil, err
	}

	provisioningScheme, err := buildProvSchemeForMcsCatalog(ctx, client, diagnostics, plan, hypervisor, hypervisorResourcePool)
	if err != nil {
		return nil, err
	}

	return provisioningScheme, nil
}

func buildProvSchemeForMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diag *diag.Diagnostics, plan MachineCatalogResourceModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, error) {

	var machineAccountCreationRules citrixorchestration.MachineAccountCreationRulesRequestModel
	machineAccountCreationRules.SetNamingScheme(plan.ProvisioningScheme.MachineAccountCreationRules.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(plan.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType.ValueString())
	if err != nil {
		diag.AddError(
			"Error creating Machine Catalog",
			fmt.Sprintf("Unsupported machine account naming scheme type. Error: %s", err.Error()),
		)
		return nil, err
	}

	machineAccountCreationRules.SetNamingSchemeType(*namingScheme)
	if plan.ProvisioningScheme.MachineDomainIdentity != nil {
		machineAccountCreationRules.SetDomain(plan.ProvisioningScheme.MachineDomainIdentity.Domain.ValueString())
		machineAccountCreationRules.SetOU(plan.ProvisioningScheme.MachineDomainIdentity.Ou.ValueString())
	}

	var provisioningScheme citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel
	provisioningScheme.SetNumTotalMachines(int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()))
	identityType := citrixorchestration.IdentityType(plan.ProvisioningScheme.IdentityType.ValueString())
	provisioningScheme.SetIdentityType(identityType)
	provisioningScheme.SetWorkGroupMachines(false) // Non-Managed setup does not support non-domain joined
	if identityType == citrixorchestration.IDENTITYTYPE_AZURE_AD {
		provisioningScheme.SetWorkGroupMachines(true)
	}
	provisioningScheme.SetMachineAccountCreationRules(machineAccountCreationRules)
	provisioningScheme.SetResourcePool(plan.ProvisioningScheme.HypervisorResourcePool.ValueString())

	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		serviceOffering := plan.ProvisioningScheme.AzureMachineConfig.ServiceOffering.ValueString()
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

		resourceGroup := plan.ProvisioningScheme.AzureMachineConfig.ResourceGroup.ValueString()
		masterImage := plan.ProvisioningScheme.AzureMachineConfig.MasterImage.ValueString()
		imagePath := ""
		if masterImage != "" {
			storageAccount := plan.ProvisioningScheme.AzureMachineConfig.StorageAccount.ValueString()
			container := plan.ProvisioningScheme.AzureMachineConfig.Container.ValueString()
			if storageAccount != "" && container != "" {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.storageaccount\\%s.container",
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
					"image.folder\\%s.resourcegroup",
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
		} else if plan.ProvisioningScheme.AzureMachineConfig.GalleryImage != nil {
			gallery := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Gallery.ValueString()
			definition := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Definition.ValueString()
			version := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Version.ValueString()
			if gallery != "" && definition != "" {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup\\%s.gallery\\%s.imagedefinition",
					resourceGroup,
					gallery,
					definition)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, version, "", "")
				if err != nil {
					diag.AddError(
						"Error creating Machine Catalog",
						fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", masterImage, version, gallery, err.Error()),
					)
					return nil, err
				}
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		machineProfile := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile
		if machineProfile != nil {
			machine := machineProfile.MachineProfileVmName.ValueString()
			machineProfileResourceGroup := machineProfile.MachineProfileResourceGroup.ValueString()
			queryPath = fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, machine, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on Azure, error: %s", plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileVmName.ValueString(), err.Error()),
				)
				return nil, err
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if plan.ProvisioningScheme.AzureMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		inputServiceOffering := plan.ProvisioningScheme.AwsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetServiceOfferingPath(serviceOffering)

		masterImage := plan.ProvisioningScheme.AwsMachineConfig.MasterImage.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString())
		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", masterImage, plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		imagePath := ""
		snapshot := plan.ProvisioningScheme.GcpMachineConfig.MachineSnapshot.ValueString()
		imageVm := plan.ProvisioningScheme.GcpMachineConfig.MasterImage.ValueString()
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

		machineProfile := plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfile, util.VirtualMachineResourceType, "")
			if err != nil {
				diag.AddError(
					"Error creating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return nil, err
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if plan.ProvisioningScheme.GcpMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		provisioningScheme.SetMemoryMB(int32(plan.ProvisioningScheme.VsphereMachineConfig.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(plan.ProvisioningScheme.VsphereMachineConfig.CpuCount.ValueInt64()))

		image := plan.ProvisioningScheme.VsphereMachineConfig.MasterImageVm.ValueString()
		snapshot := plan.ProvisioningScheme.VsphereMachineConfig.ImageSnapshot.ValueString()
		imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		if plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if !plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache.WriteBackCacheDriveLetter.IsNull() {
				provisioningScheme.SetWriteBackCacheDriveLetter(plan.ProvisioningScheme.VsphereMachineConfig.WritebackCache.WriteBackCacheDriveLetter.ValueString())
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		provisioningScheme.SetCpuCount(int32(plan.ProvisioningScheme.XenserverMachineConfig.CpuCount.ValueInt64()))
		provisioningScheme.SetMemoryMB(int32(plan.ProvisioningScheme.XenserverMachineConfig.MemoryMB.ValueInt64()))

		image := plan.ProvisioningScheme.XenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := plan.ProvisioningScheme.XenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err := getOnPremImagePath(ctx, client, diag, hypervisor.GetName(), hypervisorResourcePool.GetName(), image, snapshot, "creating")
		if err != nil {
			return nil, err
		}
		provisioningScheme.SetMasterImagePath(imagePath)

		if plan.ProvisioningScheme.XenserverMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.XenserverMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.XenserverMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.XenserverMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		if hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
			return nil, fmt.Errorf("unsupported hypervisor plugin %s", hypervisor.GetPluginId())
		}

		provisioningScheme.SetMemoryMB(int32(plan.ProvisioningScheme.NutanixMachineConfigModel.MemoryMB.ValueInt64()))
		provisioningScheme.SetCpuCount(int32(plan.ProvisioningScheme.NutanixMachineConfigModel.CpuCount.ValueInt64()))
		provisioningScheme.SetCoresPerCpuCount(int32(plan.ProvisioningScheme.NutanixMachineConfigModel.CoresPerCpuCount.ValueInt64()))

		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.NutanixMachineConfigModel.MasterImage.ValueString(), util.TemplateResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate master image %s on NUTANIX, error: %s", plan.ProvisioningScheme.NutanixMachineConfigModel.MasterImage.ValueString(), err.Error()),
			)
			return nil, err
		}

		containerId, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.NutanixMachineConfigModel.Container.ValueString(), util.StorageResourceType, "")

		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to locate container %s on NUTANIX, error: %s", plan.ProvisioningScheme.NutanixMachineConfigModel.Container.ValueString(), err.Error()),
			)
			return nil, err
		}

		provisioningScheme.SetMasterImagePath(imagePath)
		var customProperty citrixorchestration.NameValueStringPairModel
		customProperty.SetName("NutanixContainerId")
		customProperty.SetValue(containerId)
		customProperties := []citrixorchestration.NameValueStringPairModel{customProperty}
		provisioningScheme.SetCustomProperties(customProperties)
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping, err := parseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diag.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to find hypervisor network, error: %s", err.Error()),
			)
			return nil, err
		}
		provisioningScheme.SetNetworkMapping(networkMapping)
	}

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM || hypervisor.GetPluginId() != util.NUTANIX_PLUGIN_ID {
		customProperties := parseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType)
		provisioningScheme.SetCustomProperties(customProperties)
	}

	return &provisioningScheme, nil
}

func setProvSchemePropertiesForUpdateCatalog(plan MachineCatalogResourceModel, body citrixorchestration.UpdateMachineCatalogRequestModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (citrixorchestration.UpdateMachineCatalogRequestModel, error) {
	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString())
	if err != nil {
		return body, err
	}

	hypervisorResourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString(), plan.ProvisioningScheme.HypervisorResourcePool.ValueString())
	if err != nil {
		return body, err
	}

	// Resolve resource path for service offering and master image
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		serviceOffering := plan.ProvisioningScheme.AzureMachineConfig.ServiceOffering.ValueString()
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
		inputServiceOffering := plan.ProvisioningScheme.AwsMachineConfig.ServiceOffering.ValueString()
		serviceOffering, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetId(), hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", inputServiceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOffering)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		body.SetCpuCount(int32(plan.ProvisioningScheme.XenserverMachineConfig.CpuCount.ValueInt64()))
		body.SetMemoryMB(int32(plan.ProvisioningScheme.XenserverMachineConfig.MemoryMB.ValueInt64()))
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		body.SetCpuCount(int32(plan.ProvisioningScheme.VsphereMachineConfig.CpuCount.ValueInt64()))
		body.SetMemoryMB(int32(plan.ProvisioningScheme.VsphereMachineConfig.MemoryMB.ValueInt64()))
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping, err := parseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to parse network mapping, error: %s", err.Error()),
			)
			return body, err
		}
		body.SetNetworkMapping(networkMapping)
	}

	customProperties := parseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType)
	body.SetCustomProperties(customProperties)

	return body, nil
}

func generateAdminCredentialHeader(plan MachineCatalogResourceModel) string {
	credential := fmt.Sprintf("%s\\%s:%s", plan.ProvisioningScheme.MachineDomainIdentity.Domain.ValueString(), plan.ProvisioningScheme.MachineDomainIdentity.ServiceAccount.ValueString(), plan.ProvisioningScheme.MachineDomainIdentity.ServiceAccountPassword.ValueString())
	encodedData := base64.StdEncoding.EncodeToString([]byte(credential))
	header := fmt.Sprintf("Basic %s", encodedData)

	return header
}

func deleteMachinesFromMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan MachineCatalogResourceModel) error {
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

	machineDeleteRequestCount := int(catalog.GetTotalCount()) - int(plan.ProvisioningScheme.NumTotalMachines.ValueInt64())
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

	return deleteMachinesFromCatalog(ctx, client, resp, plan, machinesToDelete, catalogName, true)
}

func addMachinesToMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan MachineCatalogResourceModel) error {
	catalogId := catalog.GetId()
	catalogName := catalog.GetName()

	addMachinesCount := int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) - catalog.GetTotalCount()

	var updateMachineAccountCreationRule citrixorchestration.UpdateMachineAccountCreationRulesRequestModel
	updateMachineAccountCreationRule.SetNamingScheme(plan.ProvisioningScheme.MachineAccountCreationRules.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(plan.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Machine to Machine Catalog "+catalogName,
			"Unsupported machine account naming scheme type.",
		)
		return err
	}
	updateMachineAccountCreationRule.SetNamingSchemeType(*namingScheme)
	if plan.ProvisioningScheme.MachineDomainIdentity != nil {
		updateMachineAccountCreationRule.SetDomain(plan.ProvisioningScheme.MachineDomainIdentity.Domain.ValueString())
		updateMachineAccountCreationRule.SetOU(plan.ProvisioningScheme.MachineDomainIdentity.Ou.ValueString())
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

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client, plan, true)
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

	hypervisor, errResp := util.GetHypervisor(ctx, client, &resp.Diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString())
	if errResp != nil {
		return errResp
	}

	hypervisorResourcePool, errResp := util.GetHypervisorResourcePool(ctx, client, &resp.Diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString(), plan.ProvisioningScheme.HypervisorResourcePool.ValueString())
	if errResp != nil {
		return errResp
	}

	// Check if XDPath has changed for the image
	imagePath := ""
	machineProfilePath := ""
	var err error
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		newImage := plan.ProvisioningScheme.AzureMachineConfig.MasterImage.ValueString()
		resourceGroup := plan.ProvisioningScheme.AzureMachineConfig.ResourceGroup.ValueString()
		azureMachineProfile := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile
		if newImage != "" {
			storageAccount := plan.ProvisioningScheme.AzureMachineConfig.StorageAccount.ValueString()
			container := plan.ProvisioningScheme.AzureMachineConfig.Container.ValueString()
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
		} else if plan.ProvisioningScheme.AzureMachineConfig.GalleryImage != nil {
			gallery := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Gallery.ValueString()
			definition := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Definition.ValueString()
			version := plan.ProvisioningScheme.AzureMachineConfig.GalleryImage.Version.ValueString()
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

		if azureMachineProfile != nil {
			machineProfileName := azureMachineProfile.MachineProfileVmName.ValueString()
			machineProfileResourceGroup := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileResourceGroup.ValueString()
			queryPath := fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
			machineProfilePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, machineProfileName, util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on Azure, error: %s", plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileVmName.ValueString(), err.Error()),
				)
				return err
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		imageId := fmt.Sprintf("%s (%s)", plan.ProvisioningScheme.AwsMachineConfig.MasterImage.ValueString(), plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString())
		imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", plan.ProvisioningScheme.AwsMachineConfig.MasterImage.ValueString(), plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString(), err.Error()),
			)
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		newImage := plan.ProvisioningScheme.GcpMachineConfig.MasterImage.ValueString()
		snapshot := plan.ProvisioningScheme.GcpMachineConfig.MachineSnapshot.ValueString()
		gcpMachineProfile := plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString()

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
			machineProfilePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), util.VirtualMachineResourceType, "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		newImage := plan.ProvisioningScheme.VsphereMachineConfig.MasterImageVm.ValueString()
		snapshot := plan.ProvisioningScheme.VsphereMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "updating")
		if err != nil {
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		newImage := plan.ProvisioningScheme.XenserverMachineConfig.MasterImageVm.ValueString()
		snapshot := plan.ProvisioningScheme.XenserverMachineConfig.ImageSnapshot.ValueString()
		imagePath, err = getOnPremImagePath(ctx, client, &resp.Diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), newImage, snapshot, "updating")
		if err != nil {
			return err
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		if hypervisor.GetPluginId() == util.NUTANIX_PLUGIN_ID {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.NutanixMachineConfigModel.MasterImage.ValueString(), util.TemplateResourceType, "")

			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image %s on NUTANIX, error: %s", plan.ProvisioningScheme.NutanixMachineConfigModel.MasterImage.ValueString(), err.Error()),
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

func (r MachineCatalogResourceModel) updateCatalogWithProvScheme(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, pluginId string) MachineCatalogResourceModel {
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
		} else {

			serviceOfferingObject, err := util.GetSingleResourceFromHypervisor(ctx, client, hypervisor.GetId(), resourcePool.GetId(), "", provScheme.GetServiceOffering(), util.ServiceOfferingResourceType, "")
			if err == nil {
				provScheme.SetServiceOffering(serviceOfferingObject.GetId())
				catalog.SetProvisioningScheme(provScheme)
			}
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
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		if r.ProvisioningScheme.VsphereMachineConfig == nil {
			r.ProvisioningScheme.VsphereMachineConfig = &VsphereMachineConfigModel{}
		}

		r.ProvisioningScheme.VsphereMachineConfig.RefreshProperties(*catalog)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
		if r.ProvisioningScheme.XenserverMachineConfig == nil {
			r.ProvisioningScheme.XenserverMachineConfig = &XenserverMachineConfigModel{}
		}

		r.ProvisioningScheme.XenserverMachineConfig.RefreshProperties(*catalog)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
		if pluginId == util.NUTANIX_PLUGIN_ID {
			if r.ProvisioningScheme.NutanixMachineConfigModel == nil {
				r.ProvisioningScheme.NutanixMachineConfigModel = &NutanixMachineConfigModel{}
			}

			r.ProvisioningScheme.NutanixMachineConfigModel.RefreshProperties(*catalog)
		}
	}

	// Refresh Total Machine Count
	r.ProvisioningScheme.NumTotalMachines = types.Int64Value(int64(provScheme.GetMachineCount()))

	// Refresh Identity Type
	if identityType := types.StringValue(string(provScheme.GetIdentityType())); identityType.ValueString() != "" {
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

		networkName := (strings.Split(segments[lastIndex-1], "."))[0]
		if *connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
			/* For AWS Network, the XDPath looks like:
			 * XDHyp:\\HostingUnits\\{resource pool}\\{availability zone}.availabilityzone\\{network ip}`/{prefix length} (vpc-{vpc-id}).network
			 * The Network property should be set to {network ip}/{prefix length}
			 */
			networkName = strings.ReplaceAll(strings.Split((strings.Split(segments[lastIndex-1], ".network"))[0], " ")[0], "`/", "/")
		}
		r.ProvisioningScheme.NetworkMapping.Network = types.StringValue(networkName)
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
	if provScheme.GetIdentityType() == citrixorchestration.IDENTITYTYPE_AZURE_AD {
		return r
	}

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

func parseCustomPropertiesToClientModel(provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType) []citrixorchestration.NameValueStringPairModel {
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
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
		return nil
	}

	if len(*res) == 0 {
		return nil
	}

	return *res
}

func parseNetworkMappingToClientModel(networkMapping NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel, hypervisorPluginId string) ([]citrixorchestration.NetworkMapRequestModel, error) {
	var networks []citrixorchestration.HypervisorResourceRefResponseModel
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		networks = resourcePool.Subnets
	} else if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && hypervisorPluginId == util.NUTANIX_PLUGIN_ID {
		networks = resourcePool.Networks
	}

	var res = []citrixorchestration.NetworkMapRequestModel{}
	var networkName string
	if resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM ||
		resourcePool.ConnectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER ||
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
