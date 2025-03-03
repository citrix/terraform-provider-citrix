// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/image_definition"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func validateImageVersion(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan MachineCatalogResourceModel) error {
	if !plan.ProvisioningScheme.IsNull() {
		sessionSupport := plan.SessionSupport.ValueString()
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, diagnostics, plan.ProvisioningScheme)
		if !provSchemePlan.AzureMachineConfig.IsNull() {
			azureMachineConfig := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, diagnostics, provSchemePlan.AzureMachineConfig)
			if !azureMachineConfig.AzurePreparedImage.IsNull() {
				azurePreparedImage := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, azureMachineConfig.AzurePreparedImage)
				imageDefinition, imageScheme, imageSpecs, err := validateImageVersionHelper(ctx, diagnostics, client, sessionSupport, azurePreparedImage, "azure_machine_config")
				if err != nil {
					return err
				}

				if !azureMachineConfig.UseManagedDisks.IsNull() && !azureMachineConfig.UseManagedDisks.ValueBool() {
					err := fmt.Errorf("use_managed_disks must be set to true when using prepared image")
					diagnostics.AddError(
						"Error validating `azure_machine_config`",
						err.Error(),
					)
					return err
				}

				// Validate machine profile
				if imageScheme.MachineProfile != nil && azureMachineConfig.MachineProfile.IsNull() {
					err := fmt.Errorf("machine_profile needs to be specified when using prepared image configured with machine profile")
					diagnostics.AddError(
						"Error validating `azure_machine_config`",
						err.Error(),
					)
					return err
				}
				if imageScheme.MachineProfile == nil && !azureMachineConfig.MachineProfile.IsNull() {
					err := fmt.Errorf("machine_profile cannot be specified when using prepared image without a machine profile")
					diagnostics.AddError(
						"Error validating `azure_machine_config`",
						err.Error(),
					)
					return err
				}

				// Validate the usage of shared gallery image
				imgDefinitionConn := imageDefinition.GetHypervisorConnections()
				if len(imgDefinitionConn) > 0 {
					preparedImageUseSharedGallery := false
					customProperties := imgDefinitionConn[0].GetCustomProperties()
					for _, property := range customProperties {
						if property.GetName() == "UseSharedImageGallery" {
							preparedImageUseSharedGallery, _ = strconv.ParseBool(property.GetValue())
						}
					}
					if preparedImageUseSharedGallery && !azureMachineConfig.UseAzureComputeGallery.IsNull() {
						err := fmt.Errorf("`use_azure_compute_gallery` cannot be specified when the prepared image is using a shared image gallery. The machine catalog will inherit the azure compute gallery settings of the prepared image")
						diagnostics.AddAttributeError(
							path.Root("use_azure_compute_gallery"),
							"Incorrect Attribute Value",
							err.Error(),
						)
						return err
					}
				}

				// Validate disk encryption set
				for _, customerProperty := range imageScheme.GetCustomProperties() {
					if strings.EqualFold(customerProperty.GetName(), "DiskEncryptionSetId") && customerProperty.GetValue() != "" {
						desName, desResourceGroup := util.ParseDiskEncryptionSetIdToNameAndResourceGroup(customerProperty.GetValue())
						if azureMachineConfig.DiskEncryptionSet.IsNull() {
							err := fmt.Errorf("disk_encryption_set needs to be specified with the same disk encryption set configuration used for the prepared image")
							diagnostics.AddError(
								"Error validating `azure_machine_config`",
								err.Error(),
							)
							return err
						} else {
							diskEncryptionSet := util.ObjectValueToTypedObject[util.AzureDiskEncryptionSetModel](ctx, diagnostics, azureMachineConfig.DiskEncryptionSet)
							if diskEncryptionSet.DiskEncryptionSetName.ValueString() != strings.ToLower(desName) || diskEncryptionSet.DiskEncryptionSetResourceGroup.ValueString() != strings.ToLower(desResourceGroup) { // already lower in plan
								err := fmt.Errorf("disk_encryption_set " + diskEncryptionSet.DiskEncryptionSetResourceGroup.ValueString() + "/" + diskEncryptionSet.DiskEncryptionSetName.ValueString() +
									" does not match the disk encryption set configured in the prepared image: " + strings.ToLower(desResourceGroup+"/"+desName))
								diagnostics.AddError(
									"Error validating `azure_machine_config`",
									err.Error(),
								)
								return err
							}
						}
					}
				}

				if !provSchemePlan.HypervisorResourcePool.IsNull() {
					imageVersionResourcePool := imageSpecs.GetResourcePool()
					if imageVersionResourcePool.GetId() != provSchemePlan.HypervisorResourcePool.ValueString() {
						err := fmt.Errorf("resource pool specified in the prepared image does not match the resource pool configured in the provisioning scheme")
						diagnostics.AddError(
							"Error validating `azure_machine_config`",
							err.Error(),
						)
						return err
					}
				}
			}
		} else if !provSchemePlan.VsphereMachineConfig.IsNull() {
			vsphereMachineConfig := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, diagnostics, provSchemePlan.VsphereMachineConfig)
			if !vsphereMachineConfig.VspherePreparedImage.IsNull() {
				vspherePreparedImage := util.ObjectValueToTypedObject[PreparedImageConfigModel](ctx, diagnostics, vsphereMachineConfig.VspherePreparedImage)

				_, imageScheme, imageSpecs, err := validateImageVersionHelper(ctx, diagnostics, client, sessionSupport, vspherePreparedImage, "vsphere_machine_config")
				if err != nil {
					return err
				}

				// Validate machine profile
				if imageScheme.MachineProfile != nil && vsphereMachineConfig.MachineProfile.IsNull() {
					err := fmt.Errorf("machine_profile needs to be specified when using prepared image configured with machine profile")
					diagnostics.AddError(
						"Error validating `vsphere_machine_config`",
						err.Error(),
					)
					return err
				}
				if imageScheme.MachineProfile == nil && !vsphereMachineConfig.MachineProfile.IsNull() {
					err := fmt.Errorf("machine_profile cannot be specified when using prepared image without a machine profile")
					diagnostics.AddError(
						"Error validating `vsphere_machine_config`",
						err.Error(),
					)
					return err
				}

				if !provSchemePlan.HypervisorResourcePool.IsNull() {
					imageVersionResourcePool := imageSpecs.GetResourcePool()
					if imageVersionResourcePool.GetId() != provSchemePlan.HypervisorResourcePool.ValueString() {
						err := fmt.Errorf("resource pool specified in the prepared image does not match the resource pool configured in the provisioning scheme")
						diagnostics.AddError(
							"Error validating `vsphere_machine_config`",
							err.Error(),
						)
						return err
					}
				}
			}
		}
	}
	return nil
}

func validateImageVersionHelper(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, sessionSupport string, preparedImageConfig PreparedImageConfigModel, machineConfigAttributeName string) (*citrixorchestration.ImageDefinitionResponseModel, citrixorchestration.ImageSchemeResponseModel, citrixorchestration.ImageVersionSpecResponseModel, error) {
	isPreparedImageSupported := util.CheckProductVersion(client, diagnostics, 121, 118, 7, 41, "Error using Prepared Image in citrix_machine_catalog resource", "Prepared Image")
	var imageScheme citrixorchestration.ImageSchemeResponseModel
	var imageSpecs citrixorchestration.ImageVersionSpecResponseModel
	var imageDefinition *citrixorchestration.ImageDefinitionResponseModel
	if !isPreparedImageSupported {
		err := fmt.Errorf("prepared Image is not supported in this version of Citrix Virtual Apps and Desktops service")
		return nil, imageScheme, imageSpecs, err
	}
	imageDefinition, err := image_definition.GetImageDefinition(ctx, client, diagnostics, preparedImageConfig.ImageDefinition.ValueString())
	if err != nil {
		return imageDefinition, imageScheme, imageSpecs, err
	}

	imageVersion, err := image_definition.GetImageVersion(ctx, client, diagnostics, imageDefinition.GetId(), preparedImageConfig.ImageVersion.ValueString())
	if err != nil {
		return imageDefinition, imageScheme, imageSpecs, err
	}

	imageVersionStatus := imageVersion.GetImageVersionStatus()
	if imageVersionStatus != citrixorchestration.IMAGEVERSIONSTATUS_SUCCESS {
		err := fmt.Errorf("image version in state `%s` cannot be used to create machine catalog", string(imageVersionStatus))
		diagnostics.AddError(
			"Error validating azure_machine_config",
			err.Error(),
		)
		return imageDefinition, imageScheme, imageSpecs, err
	}

	imageContextConfigured := false
	for _, spec := range imageVersion.GetImageVersionSpecs() {
		if spec.Context != nil {
			context := spec.GetContext()
			if context.ImageScheme == nil {
				continue
			}
			imageContextConfigured = true
			imageScheme = context.GetImageScheme()
			imageSpecs = spec
		}
	}
	imageRuntimeEnvironment := imageSpecs.GetImageRuntimeEnvironment()
	if imageContextConfigured {
		// Validate session support
		if imageRuntimeEnvironment.GetVDASessionSupport() != "" && !strings.EqualFold(sessionSupport, imageRuntimeEnvironment.GetVDASessionSupport()) {
			err := fmt.Errorf("session_support specified does not match the session support configured in the prepared image")
			diagnostics.AddError(
				fmt.Sprintf("Error validating `%s`", machineConfigAttributeName),
				err.Error(),
			)
			return imageDefinition, imageScheme, imageSpecs, err
		}
	}

	return imageDefinition, imageScheme, imageSpecs, nil
}
