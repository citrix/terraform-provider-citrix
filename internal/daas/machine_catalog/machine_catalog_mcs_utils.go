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

	provisioningScheme, errorMsg := buildProvSchemeForMcsCatalog(ctx, client, plan, hypervisor, hypervisorResourcePool)
	if errorMsg != "" || provisioningScheme == nil {
		diagnostics.AddError(
			"Error creating Machine Catalog",
			errorMsg,
		)

		return nil, fmt.Errorf(errorMsg)
	}

	return provisioningScheme, nil
}

func buildProvSchemeForMcsCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan MachineCatalogResourceModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) (*citrixorchestration.CreateMachineCatalogProvisioningSchemeRequestModel, string) {

	var machineAccountCreationRules citrixorchestration.MachineAccountCreationRulesRequestModel
	machineAccountCreationRules.SetNamingScheme(plan.ProvisioningScheme.MachineAccountCreationRules.NamingScheme.ValueString())
	namingScheme, err := citrixorchestration.NewNamingSchemeTypeFromValue(plan.ProvisioningScheme.MachineAccountCreationRules.NamingSchemeType.ValueString())
	if err != nil {
		return nil, "Unsupported machine account naming scheme type."
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
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, "serviceoffering", "")
		if err != nil {
			return nil, fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error())
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
					return nil, fmt.Sprintf("Failed to resolve master image VHD %s in container %s of storage account %s, error: %s", masterImage, container, storageAccount, err.Error())
				}
			} else {
				queryPath = fmt.Sprintf(
					"image.folder\\%s.resourcegroup",
					resourceGroup)
				imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, masterImage, "", "")
				if err != nil {
					return nil, fmt.Sprintf("Failed to resolve master image Managed Disk or Snapshot %s, error: %s", masterImage, err.Error())
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
					return nil, fmt.Sprintf("Failed to locate Azure Image Gallery image %s of version %s in gallery %s, error: %s", masterImage, version, gallery, err.Error())
				}
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		machineProfile := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile
		if machineProfile != nil {
			machine := machineProfile.MachineProfileVmName.ValueString()
			machineProfileResourceGroup := machineProfile.MachineProfileResourceGroup.ValueString()
			queryPath = fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, machine, "vm", "")
			if err != nil {
				return nil, fmt.Sprintf("Failed to locate machine profile %s on Azure, error: %s", plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileVmName.ValueString(), err.Error())
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if plan.ProvisioningScheme.AzureMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.PersistVm.ValueBool() && !plan.ProvisioningScheme.AzureMachineConfig.WritebackCache.PersistOsDisk.ValueBool() {
				return nil, "Could not set persist_vm attribute, which can only be set when persist_os_disk = true"
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		serviceOffering := plan.ProvisioningScheme.AwsMachineConfig.ServiceOffering.ValueString()
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", serviceOffering, "serviceoffering", "")
		if err != nil {
			return nil, fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error())
		}
		provisioningScheme.SetServiceOfferingPath(serviceOfferingPath)

		masterImage := plan.ProvisioningScheme.AwsMachineConfig.MasterImage.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString())
		imagePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, "template", "")
		if err != nil {
			return nil, fmt.Sprintf("Failed to locate AWS image %s with AMI %s, error: %s", masterImage, plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString(), err.Error())
		}
		provisioningScheme.SetMasterImagePath(imagePath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		imagePath := ""
		snapshot := plan.ProvisioningScheme.GcpMachineConfig.MachineSnapshot.ValueString()
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", plan.ProvisioningScheme.GcpMachineConfig.MasterImage.ValueString())
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, plan.ProvisioningScheme.GcpMachineConfig.MachineSnapshot.ValueString(), "snapshot", "")
			if err != nil {
				return nil, fmt.Sprintf("Failed to locate master image snapshot %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error())
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.GcpMachineConfig.MasterImage.ValueString(), "vm", "")
			if err != nil {
				return nil, fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error())
			}
		}

		provisioningScheme.SetMasterImagePath(imagePath)

		machineProfile := plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfile, "vm", "")
			if err != nil {
				return nil, fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error())
			}
			provisioningScheme.SetMachineProfilePath(machineProfilePath)
		}

		if plan.ProvisioningScheme.GcpMachineConfig.WritebackCache != nil {
			provisioningScheme.SetUseWriteBackCache(true)
			provisioningScheme.SetWriteBackCacheDiskSizeGB(int32(plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheDiskSizeGB.ValueInt64()))
			if !plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.IsNull() {
				provisioningScheme.SetWriteBackCacheMemorySizeMB(int32(plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.WriteBackCacheMemorySizeMB.ValueInt64()))
			}
			if plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.PersistVm.ValueBool() && !plan.ProvisioningScheme.GcpMachineConfig.WritebackCache.PersistOsDisk.ValueBool() {
				return nil, "Could not set persist_vm attribute, which can only be set when persist_os_disk = true"
			}

		}
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping, err := parseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool)
		if err != nil {
			return nil, err.Error()
		}
		provisioningScheme.SetNetworkMapping(networkMapping)
	}

	customProperties := parseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType, false)
	provisioningScheme.SetCustomProperties(customProperties)

	return &provisioningScheme, ""
}

func setProvSchemePropertiesForUpdateCatalog(plan MachineCatalogResourceModel, body citrixorchestration.UpdateMachineCatalogRequestModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, connectionType *citrixorchestration.HypervisorConnectionType) (citrixorchestration.UpdateMachineCatalogRequestModel, error) {
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
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, serviceOffering, "serviceoffering", "")
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
		if machineProfile := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile; machineProfile != nil {
			machineProfileName := machineProfile.MachineProfileVmName.ValueString()
			if machineProfileName != "" {
				machineProfileResourceGroup := plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileResourceGroup.ValueString()
				queryPath = fmt.Sprintf("machineprofile.folder\\%s.resourcegroup", machineProfileResourceGroup)
				machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, machineProfileName, "vm", "")
				if err != nil {
					diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Failed to locate machine profile %s on Azure, error: %s", plan.ProvisioningScheme.AzureMachineConfig.MachineProfile.MachineProfileVmName.ValueString(), err.Error()),
					)
					return body, err
				}
				body.SetMachineProfilePath(machineProfilePath)
			}
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		serviceOffering := plan.ProvisioningScheme.AwsMachineConfig.ServiceOffering.ValueString()
		serviceOfferingPath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", serviceOffering, "serviceoffering", "")
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return body, err
		}
		body.SetServiceOfferingPath(serviceOfferingPath)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		machineProfile := plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString()
		if machineProfile != "" {
			machineProfilePath, err := util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), "vm", "")
			if err != nil {
				diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate machine profile %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return body, err
			}
			body.SetMachineProfilePath(machineProfilePath)
		}
	}

	if plan.ProvisioningScheme.NetworkMapping != nil {
		networkMapping, err := parseNetworkMappingToClientModel(*plan.ProvisioningScheme.NetworkMapping, hypervisorResourcePool)
		if err != nil {
			diagnostics.AddError(
				"Error updating Machine Catalog",
				fmt.Sprintf("Failed to parse network mapping, error: %s", err.Error()),
			)
			return body, err
		}
		body.SetNetworkMapping(networkMapping)
	}

	customProperties := parseCustomPropertiesToClientModel(*plan.ProvisioningScheme, hypervisor.ConnectionType, true)
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

func updateCatalogImage(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, catalog *citrixorchestration.MachineCatalogDetailResponseModel, plan MachineCatalogResourceModel) error {

	catalogName := catalog.GetName()
	catalogId := catalog.GetId()

	provScheme := catalog.GetProvisioningScheme()
	masterImage := provScheme.GetMasterImage()

	hypervisor, errResp := util.GetHypervisor(ctx, client, &resp.Diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString())
	if errResp != nil {
		return errResp
	}

	hypervisorResourcePool, errResp := util.GetHypervisorResourcePool(ctx, client, &resp.Diagnostics, plan.ProvisioningScheme.Hypervisor.ValueString(), plan.ProvisioningScheme.HypervisorResourcePool.ValueString())
	if errResp != nil {
		return errResp
	}

	var customProperties = &[]citrixorchestration.NameValueStringPairModel{}

	// Check if XDPath has changed for the image
	imagePath := ""
	var err error
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		newImage := plan.ProvisioningScheme.AzureMachineConfig.MasterImage.ValueString()
		resourceGroup := plan.ProvisioningScheme.AzureMachineConfig.ResourceGroup.ValueString()
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
		var provAzureUseSharedImageGallery bool
		for _, stringPair := range provScheme.GetCustomProperties() {
			if stringPair.GetName() == "UseSharedImageGallery" {
				provAzureUseSharedImageGallery, _ = strconv.ParseBool(stringPair.GetValue())
				break
			}
		}
		planAzureUseSharedImageGallery := plan.ProvisioningScheme.AzureMachineConfig.PlaceImageInGallery != nil
		if planAzureUseSharedImageGallery != provAzureUseSharedImageGallery {
			util.AppendNameValueStringPair(customProperties, "UseSharedImageGallery", strconv.FormatBool(planAzureUseSharedImageGallery))
		}
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		imageId := fmt.Sprintf("%s (%s)", plan.ProvisioningScheme.AwsMachineConfig.MasterImage.ValueString(), plan.ProvisioningScheme.AwsMachineConfig.ImageAmi.ValueString())
		imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, "template", "")
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
		if snapshot != "" {
			queryPath := fmt.Sprintf("%s.vm", newImage)
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), queryPath, plan.ProvisioningScheme.GcpMachineConfig.MachineSnapshot.ValueString(), "snapshot", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image snapshot %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		} else {
			imagePath, err = util.GetSingleResourcePathFromHypervisor(ctx, client, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", newImage, "vm", "")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog",
					fmt.Sprintf("Failed to locate master image machine %s on GCP, error: %s", plan.ProvisioningScheme.GcpMachineConfig.MachineProfile.ValueString(), err.Error()),
				)
				return err
			}
		}
	}

	if len(*customProperties) == 0 && masterImage.GetXDPath() == imagePath {
		return nil
	}

	// Update Master Image for Machine Catalog
	var updateProvisioningSchemeModel citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel
	var rebootOption citrixorchestration.RebootMachinesRequestModel

	// Update the image immediately
	rebootOption.SetRebootDuration(60)
	rebootOption.SetWarningDuration(15)
	rebootOption.SetWarningMessage("Warning: An important update is about to be installed. To ensure that no loss of data occurs, save any outstanding work and close all applications.")
	updateProvisioningSchemeModel.SetRebootOptions(rebootOption)
	updateProvisioningSchemeModel.SetMasterImagePath(imagePath)
	updateProvisioningSchemeModel.SetStoreOldImage(true)
	updateProvisioningSchemeModel.SetMinimumFunctionalLevel("L7_20")
	if len(*customProperties) > 0 {
		/*
			util.AppendNameValueStringPair(customProperties, "SharedImageGalleryReplicaRatio", util.TypeInt64ToString(plan.ProvisioningScheme.AzureMachineConfig.PlaceImageInGallery.ReplicaRatio))
			util.AppendNameValueStringPair(customProperties, "SharedImageGalleryReplicaMaximum", util.TypeInt64ToString(plan.ProvisioningScheme.AzureMachineConfig.PlaceImageInGallery.ReplicaMaximum))
		*/
		updateProvisioningSchemeModel.SetCustomProperties(*customProperties)
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

func (r MachineCatalogResourceModel) updateCatalogWithProvScheme(catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType) MachineCatalogResourceModel {
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
		/* For AWS Network, the XDPath looks like:
		 * XDHyp:\\HostingUnits\\{resource pool}\\{availability zone}.availabilityzone\\{network ip}`/{prefix length} (vpc-{vpc-id}).network
		 * The Network property should be set to {network ip}/{prefix length}
		 */
		r.ProvisioningScheme.NetworkMapping.Network = types.StringValue(strings.ReplaceAll(strings.Split((strings.Split(segments[lastIndex-1], ".network"))[0], " ")[0], "`/", "/"))
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

func parseCustomPropertiesToClientModel(provisioningScheme ProvisioningSchemeModel, connectionType citrixorchestration.HypervisorConnectionType, update bool) []citrixorchestration.NameValueStringPairModel {
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
		if !provisioningScheme.AzureMachineConfig.UseEphemeralOsDisk.IsNull() {
			if provisioningScheme.AzureMachineConfig.UseEphemeralOsDisk.ValueBool() {
				util.AppendNameValueStringPair(res, "UseEphemeralOsDisk", "true")
			} else {
				util.AppendNameValueStringPair(res, "UseEphemeralOsDisk", "false")
			}
		}
		if !provisioningScheme.AzureMachineConfig.LicenseType.IsNull() {
			util.AppendNameValueStringPair(res, "LicenseType", provisioningScheme.AzureMachineConfig.LicenseType.ValueString())
		} else {
			util.AppendNameValueStringPair(res, "LicenseType", "")
		}
		if provisioningScheme.AzureMachineConfig.PlaceImageInGallery != nil {
			if !update {
				util.AppendNameValueStringPair(res, "UseSharedImageGallery", "true")
			}
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaRatio", util.TypeInt64ToString(provisioningScheme.AzureMachineConfig.PlaceImageInGallery.ReplicaRatio))
			util.AppendNameValueStringPair(res, "SharedImageGalleryReplicaMaximum", util.TypeInt64ToString(provisioningScheme.AzureMachineConfig.PlaceImageInGallery.ReplicaMaximum))
		}
		if !provisioningScheme.AzureMachineConfig.DiskEncryptionSetId.IsNull() {
			util.AppendNameValueStringPair(res, "DiskEncryptionSetId", provisioningScheme.AzureMachineConfig.DiskEncryptionSetId.ValueString())
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

	if len(*res) == 0 {
		return nil
	}

	return *res
}

func parseNetworkMappingToClientModel(networkMapping NetworkMappingModel, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) ([]citrixorchestration.NetworkMapRequestModel, error) {
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