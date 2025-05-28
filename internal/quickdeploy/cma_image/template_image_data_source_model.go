// Copyright © 2024. Citrix Systems, Inc.
package cma_image

import (
	"context"

	catalogservice "github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CitrixManagedAzureImageDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Notes             types.String `tfsdk:"notes"`
	State             types.String `tfsdk:"state"`
	SubscriptionName  types.String `tfsdk:"subscription_name"`
	Region            types.String `tfsdk:"region"`
	MachineGeneration types.String `tfsdk:"machine_generation"`
	OsPlatform        types.String `tfsdk:"os_platform"`
	VtpmEnabled       types.Bool   `tfsdk:"vtpm_enabled"`
	SecureBootEnabled types.Bool   `tfsdk:"secure_boot_enabled"`
}

func (CitrixManagedAzureImageDataSourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - Citrix Managed Azure --- Data Source of an Citrix Managed Azure image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the image.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the image.",
				Required:    true,
			},
			"notes": schema.StringAttribute{
				Description: "Notes of the image.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the image.",
				Computed:    true,
			},
			"subscription_name": schema.StringAttribute{
				Description: "The name of the Citrix Managed Azure subscription the image was imported to.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "The Azure region the image was imported to.",
				Computed:    true,
			},
			"machine_generation": schema.StringAttribute{
				Description: "The generation of the virtual machine this image will be used to create.",
				Computed:    true,
			},
			"os_platform": schema.StringAttribute{
				Description: "The OS platform of the image.",
				Computed:    true,
			},
			"vtpm_enabled": schema.BoolAttribute{
				Description: "Indicates whether the image supports vTPM TrustedLaunch.",
				Computed:    true,
			},
			"secure_boot_enabled": schema.BoolAttribute{
				Description: "Indicates whether the image supports Secure Boot.",
				Computed:    true,
			},
		},
	}
}

func (CitrixManagedAzureImageDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return CitrixManagedAzureImageDataSourceModel{}.GetDataSourceSchema().Attributes
}

func (r CitrixManagedAzureImageDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, image *catalogservice.TemplateImageDetails) CitrixManagedAzureImageDataSourceModel {
	r.Id = types.StringValue(image.GetId())
	r.Name = types.StringValue(image.GetName())
	r.Notes = types.StringValue(image.GetNotes())
	r.State = types.StringValue(string(image.GetState()))
	r.SubscriptionName = types.StringValue(image.GetSubscriptionName())
	r.Region = types.StringValue(image.GetRegion())
	r.MachineGeneration = types.StringValue(image.GetHyperVGen())
	r.OsPlatform = types.StringValue(string(image.GetOsPlatform()))
	r.VtpmEnabled = types.BoolValue(image.GetVtpmEnabled())
	r.SecureBootEnabled = types.BoolValue(image.GetSecureBootEnabled())

	return r
}
