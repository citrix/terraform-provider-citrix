package image_definition

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func buildImageScheme(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, imageScheme *citrixorchestration.CreateImageSchemeRequestModel, plan ImageVersionModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel, hypervisorResourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) (string, error) {
	hypervisorId := hypervisor.GetId()
	var masterImagePath string
	var err error

	switch hypervisor.GetPluginId() {
	case util.AZURERM_FACTORY_NAME:
		azureImageSpecs := util.ObjectValueToTypedObject[util.AzureImageSpecsModel](ctx, diagnostics, plan.AzureImageSpecs)
		sharedSubscription := azureImageSpecs.SharedSubscription.ValueString()
		resourceGroup := azureImageSpecs.ResourceGroup.ValueString()
		masterImage := azureImageSpecs.MasterImage.ValueString()

		masterImagePath, err = util.BuildAzureMasterImagePath(ctx, client, diagnostics, azureImageSpecs.GalleryImage, sharedSubscription, resourceGroup, "", "", masterImage, hypervisorId, plan.ResourcePool.ValueString(), "Error creating Image Version")
		if err != nil {
			return masterImagePath, err
		}

		machineProfile := azureImageSpecs.MachineProfile
		if !machineProfile.IsNull() {
			machineProfilePath, err := util.HandleMachineProfileForAzureMcsPvsCatalog(ctx, client, diagnostics, hypervisorId, hypervisorResourcePool.GetName(), util.ObjectValueToTypedObject[util.AzureMachineProfileModel](ctx, diagnostics, machineProfile), "Error creating Image Version")
			if err != nil {
				return masterImagePath, err
			}
			imageScheme.SetMachineProfile(machineProfilePath)
		}

		// Set ServiceOfferingPath
		serviceOffering := azureImageSpecs.ServiceOffering.ValueString()
		serviceOfferingQueryPath := "serviceoffering.folder"
		serviceOfferingPath, httpResp, err := util.GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorId, hypervisorResourcePool.GetName(), serviceOfferingQueryPath, serviceOffering, util.ServiceOfferingResourceType, "")
		if err != nil {
			diagnostics.AddError(
				"Error creating Image Version",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on Azure, error: %s", serviceOffering, err.Error()),
			)
			return masterImagePath, err
		}
		imageScheme.SetServiceOfferingPath(serviceOfferingPath)

		// Set CustomProperties
		customProperties := &[]citrixorchestration.NameValueStringPairModel{}
		// Set Storage Type
		util.AppendNameValueStringPair(customProperties, "StorageType", azureImageSpecs.StorageType.ValueString())
		licenseType := azureImageSpecs.LicenseType.ValueString()
		// Set License Type
		util.AppendNameValueStringPair(customProperties, "LicenseType", licenseType)
		// Set Disk Encryption Set
		if !azureImageSpecs.DiskEncryptionSet.IsNull() {
			diskEncryptionSetModel := util.ObjectValueToTypedObject[util.AzureDiskEncryptionSetModel](ctx, diagnostics, azureImageSpecs.DiskEncryptionSet)
			diskEncryptionSet := diskEncryptionSetModel.DiskEncryptionSetName.ValueString()
			diskEncryptionSetRg := diskEncryptionSetModel.DiskEncryptionSetResourceGroup.ValueString()
			des, httpResp, err := util.GetSingleResourceFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorId, plan.ResourcePool.ValueString(), fmt.Sprintf("%s\\diskencryptionset.folder", hypervisorResourcePool.GetXDPath()), diskEncryptionSet, "", diskEncryptionSetRg)
			if err != nil {
				diagnostics.AddError(
					"Error creating Image Version",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate disk encryption set %s in resource group %s, error: %s", diskEncryptionSet, diskEncryptionSetRg, err.Error()),
				)
			}
			util.AppendNameValueStringPair(customProperties, "DiskEncryptionSetId", des.GetId())
		}
		imageScheme.SetCustomProperties(*customProperties)
	case util.VMWARE_FACTORY_NAME:
		vsphereImageSpecs := util.ObjectValueToTypedObject[VsphereImageSpecsModel](ctx, diagnostics, plan.VsphereImageSpecs)
		imageScheme.SetMemoryMB(vsphereImageSpecs.MemoryMB.ValueInt32())

		imageVm := vsphereImageSpecs.MasterImageVm.ValueString()
		snapshotPath := vsphereImageSpecs.ImageSnapshot.ValueString()
		masterImagePath = fmt.Sprintf("XDHyp:\\HostingUnits\\%s\\%s.vm", hypervisorResourcePool.GetName(), imageVm)
		for _, snapshot := range strings.Split(snapshotPath, "/") {
			masterImagePath += fmt.Sprintf("\\%s.snapshot", snapshot)
		}

		imageScheme.SetCpuCount(vsphereImageSpecs.CpuCount.ValueInt32())
		if !vsphereImageSpecs.MachineProfile.IsNull() {
			// Set Machine Profile
			machineProfileName := vsphereImageSpecs.MachineProfile.ValueString()
			machineProfile, httpResp, err := util.GetSingleResourceFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", machineProfileName, util.TemplateResourceType, "")
			if err != nil {
				diagnostics.AddError(
					"Error creating Machine Catalog",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						fmt.Sprintf("\nFailed to locate machine profile %s on vSphere, error: %s", machineProfileName, err.Error()),
				)
				return masterImagePath, err
			}

			imageScheme.SetMachineProfile(machineProfile.GetXDPath())

			// CPU count will inherit from machine profile
			machineProfileAdditionalData := machineProfile.GetAdditionalData()
			machineProfileCpuCountSpecified := false
			for _, entry := range machineProfileAdditionalData {
				if strings.EqualFold(entry.GetName(), util.CPU_COUNT_PROPERTY_NAME) {
					machineProfileCpuCountSpecified = true
					machineProfileCpuCount, err := strconv.ParseInt(entry.GetValue(), 10, 32)
					if err != nil {
						diagnostics.AddError(
							"Error creating vSphere Image Version",
							"Unable to parse the number of CPU(s) from the configuration of machine profile "+machineProfileName,
						)
						return masterImagePath, err
					}
					if int32(machineProfileCpuCount) != vsphereImageSpecs.CpuCount.ValueInt32() {
						diagnostics.AddError(
							"Error creating vSphere Image Version",
							fmt.Sprintf("The specified `cpu_count` value %d does not match the number of CPU(s) %d of machine profile %s.", vsphereImageSpecs.CpuCount.ValueInt32(), machineProfileCpuCount, machineProfileName),
						)
						return masterImagePath, err
					}
				}
			}
			if !machineProfileCpuCountSpecified {
				err = fmt.Errorf("machine profile %s does not have the number of cpu(s) specified", machineProfileName)
				diagnostics.AddError(
					"Error creating vSphere Image Version",
					err.Error(),
				)
				return masterImagePath, err
			}
		}
	case util.AWS_FACTORY_NAME:
		awsEc2ImageSpecs := util.ObjectValueToTypedObject[AwsEc2ImageSpecsModel](ctx, diagnostics, plan.AwsEc2ImageSpecs)
		inputServiceOffering := awsEc2ImageSpecs.ServiceOffering.ValueString()
		serviceOffering, httpResp, err := util.GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorId, hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diagnostics.AddError(
				"Error creating Image Version",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return masterImagePath, err
		}
		imageScheme.SetServiceOfferingPath(serviceOffering)

		masterImage := awsEc2ImageSpecs.AmiName.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, awsEc2ImageSpecs.AmiId.ValueString())
		masterImagePath, httpResp, err = util.GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			diagnostics.AddError(
				"Error creating Image Version",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate AWS image %s with AMI %s, error: %s", masterImage, awsEc2ImageSpecs.AmiId.ValueString(), err.Error()),
			)
			return masterImagePath, err
		}

		machineProfile := util.ObjectValueToTypedObject[util.AwsMachineProfileModel](ctx, diagnostics, awsEc2ImageSpecs.MachineProfile)
		machineProfilePath, err := util.HandleMachineProfileForAwsMcsPvsCatalog(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), machineProfile, "Error creating AWS EC2 Image Version")
		if err != nil {
			return masterImagePath, err
		}
		imageScheme.SetMachineProfile(machineProfilePath)
	case util.AMAZON_WORKSPACES_CORE_FACTORY_NAME:
		amazonWorkspacesCoreImageSpecs := util.ObjectValueToTypedObject[AmazonWorkspacesCoreImageSpecsModel](ctx, diagnostics, plan.AmazonWorkspacesCoreImageSpecs)
		inputServiceOffering := amazonWorkspacesCoreImageSpecs.ServiceOffering.ValueString()
		serviceOffering, httpResp, err := util.GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorId, hypervisorResourcePool.GetId(), "", inputServiceOffering, util.ServiceOfferingResourceType, "")

		if err != nil {
			diagnostics.AddError(
				"Error creating Image Version",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to resolve service offering %s on AWS, error: %s", serviceOffering, err.Error()),
			)
			return masterImagePath, err
		}
		imageScheme.SetServiceOfferingPath(serviceOffering)

		masterImage := amazonWorkspacesCoreImageSpecs.MasterImage.ValueString()
		imageId := fmt.Sprintf("%s (%s)", masterImage, amazonWorkspacesCoreImageSpecs.ImageAmi.ValueString())
		masterImagePath, httpResp, err = util.GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), "", imageId, util.TemplateResourceType, "")
		if err != nil {
			diagnostics.AddError(
				"Error creating Image Version",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					fmt.Sprintf("\nFailed to locate AWS image %s with AMI %s, error: %s", masterImage, amazonWorkspacesCoreImageSpecs.ImageAmi.ValueString(), err.Error()),
			)
			return masterImagePath, err
		}

		if !amazonWorkspacesCoreImageSpecs.MachineProfile.IsNull() {
			machineProfile := util.ObjectValueToTypedObject[util.AwsMachineProfileModel](ctx, diagnostics, amazonWorkspacesCoreImageSpecs.MachineProfile)
			machineProfilePath, err := util.HandleMachineProfileForAwsMcsPvsCatalog(ctx, client, diagnostics, hypervisor.GetName(), hypervisorResourcePool.GetName(), machineProfile, "Error creating AWS WorkSpaces Core Image Version")
			if err != nil {
				return masterImagePath, err
			}
			imageScheme.SetMachineProfile(machineProfilePath)
		}

	default:
		err = fmt.Errorf("unsupported hypervisor connection: %s", hypervisorId)
		diagnostics.AddError(
			"Error creating Image Version",
			err.Error(),
		)
		return masterImagePath, err
	}

	return masterImagePath, nil
}
