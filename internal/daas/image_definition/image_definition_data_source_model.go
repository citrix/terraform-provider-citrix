// Copyright © 2024. Citrix Systems, Inc.
package image_definition

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (AzureImageDefinitionModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Details of the Azure Image Definition.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"resource_group": schema.StringAttribute{
				Description: "Existing resource group to store the image definition. If not specified, a new resource group will be created.",
				Computed:    true,
			},
			"use_image_gallery": schema.BoolAttribute{
				Description: "Whether image gallery is used to store the image definition. Defaults to `false`.",
				Computed:    true,
			},
			"image_gallery_name": schema.StringAttribute{
				Description: "Name of the existing image gallery. If not specified and `use_image_gallery` is `true`, a new image gallery will be created.",
				Computed:    true,
			},
		},
	}
}

func (AzureImageDefinitionModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AzureImageDefinitionModel{}.GetDataSourceSchema().Attributes
}

func (ImageDefinitionModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of an image definition. **Note that this feature is in Tech Preview.**",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The GUID identifier of the image definition.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("name"),
					),
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the image definition.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the image definition.",
				Computed:    true,
			},
			"os_type": schema.StringAttribute{
				Description: "Operating system type of the image definition.",
				Computed:    true,
			},
			"session_support": schema.StringAttribute{
				Description: "Session support of the image definition.",
				Computed:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "ID of the hypervisor connection to be used for image definition.",
				Computed:    true,
			},
			"azure_image_definition": AzureImageDefinitionModel{}.GetDataSourceSchema(),
			"latest_version": schema.Int64Attribute{
				Description: "Latest version of the image definition.",
				Computed:    true,
			},
		},
	}
}

func (ImageDefinitionModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ImageDefinitionModel{}.GetDataSourceSchema().Attributes
}
