// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureImageSpecsDataSourceModel struct {
	// Required Attributes
	ServiceOffering types.String `tfsdk:"service_offering"`
	LicenseType     types.String `tfsdk:"license_type"`
	StorageType     types.String `tfsdk:"storage_type"`

	// Optional Attributes
	MachineProfile    types.Object `tfsdk:"machine_profile"`
	DiskEncryptionSet types.Object `tfsdk:"disk_encryption_set"`
}

func (AzureImageSpecsDataSourceModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Image configuration for Azure image version.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"service_offering": schema.StringAttribute{
				Description: "The Azure VM Sku to use when creating machines.",
				Computed:    true,
			},
			"license_type": schema.StringAttribute{
				Description: "Windows license type used to provision virtual machines in Azure at the base compute rate. License types include: `Windows_Client` and `Windows_Server`.",
				Computed:    true,
			},
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Computed:    true,
			},
			"machine_profile":     util.AzureMachineProfileModel{}.GetDataSourceSchema(),
			"disk_encryption_set": util.AzureDiskEncryptionSetModel{}.GetDataSourceSchema(),
		},
	}
}

func (AzureImageSpecsDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AzureImageSpecsDataSourceModel{}.GetDataSourceSchema().Attributes
}

func (ImageVersionModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source an image version. **Note that this feature is in Tech Preview.**",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the image version.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("version_number")),
				},
			},
			"image_definition": schema.StringAttribute{
				Description: "Id of the image definition to associate this image version with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"version_number": schema.Int32Attribute{
				Description: "The version number of the image version.",
				Optional:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor to use for creating this image version.",
				Computed:    true,
			},
			"hypervisor_resource_pool": schema.StringAttribute{
				Description: "Id of the hypervisor resource pool to use for creating this image version.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the image version.",
				Computed:    true,
			},
			"azure_image_specs": AzureImageSpecsDataSourceModel{}.GetDataSourceSchema(),
			"session_support": schema.StringAttribute{
				Description: "Session support for the image version.",
				Computed:    true,
			},
			"os_type": schema.StringAttribute{
				Description: "The OS type of the image version.",
				Computed:    true,
			},
		},
	}
}

func (ImageVersionModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ImageVersionModel{}.GetDataSourceSchema().Attributes
}

func (r ImageVersionModel) RefreshDataSourcePropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, imageVersion *citrixorchestration.ImageVersionResponseModel) ImageVersionModel {
	r, imageSpecs, specConfigured := r.RefreshImageVersionBaseProperties(ctx, diagnostics, imageVersion)
	if specConfigured {
		return r
	}

	imageContext := imageSpecs.GetContext()
	switch imageContext.GetPluginFactoryName() {
	case util.AZURERM_FACTORY_NAME:
		imageScheme := imageContext.GetImageScheme()
		azureImageSpecs := AzureImageSpecsDataSourceModel{}

		azureImageSpecs.ServiceOffering = parseAzureImageVersionServiceOffering(imageScheme.GetServiceOffering())

		licenseType, storageType, des, err := parseAzureImageCustomProperties(ctx, diagnostics, false, imageScheme.GetCustomProperties(), azureImageSpecs.DiskEncryptionSet)
		if err != nil {
			return r
		}
		azureImageSpecs.LicenseType = licenseType
		azureImageSpecs.StorageType = storageType
		azureImageSpecs.DiskEncryptionSet = des

		updatedMachineProfile, err := refreshAzureImageVersionMachineProfile(ctx, diagnostics, false, imageScheme)
		if err == nil {
			azureImageSpecs.MachineProfile = updatedMachineProfile
		}

		r.AzureImageSpecs = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, azureImageSpecs)
	default:
		diagnostics.AddError(
			"Error refreshing Image Version data source",
			fmt.Sprintf("Hypervisor connection type %s is not supported", imageContext.GetPluginFactoryName()),
		)
	}

	return r
}
