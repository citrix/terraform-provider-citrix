// Copyright © 2024. Citrix Systems, Inc.
package qcs_image

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (AwsWorkspacesImageModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Data Source of an AWS WorkSpaces image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the image.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"account_id": schema.StringAttribute{
				Description: "GUID identifier of the account.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the image.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"aws_image_id": schema.StringAttribute{
				Description: "Id of the image to be imported in AWS.",
				Computed:    true,
			},
			"aws_imported_image_id": schema.StringAttribute{
				Description: "The Id of the image imported in AWS WorkSpaces.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the image.",
				Computed:    true,
			},
			"session_support": schema.StringAttribute{
				Description: "The supported session type of the image. Possible values are `SingleSession` and `MultiSession`.",
				Computed:    true,
			},
			"operating_system": schema.StringAttribute{
				Description: "The type of operating system of the image. Possible values are `WINDOWS` and `LINUX`.",
				Computed:    true,
			},
			"ingestion_process": schema.StringAttribute{
				Description: "The type of ingestion process of the image. Possible values are `BYOL_REGULAR_BYOP` and `BYOL_GRAPHICS_G4DN_BYOP`.",
				Computed:    true,
			},
			"tenancy": schema.StringAttribute{
				Description: "The type of tenancy of the image.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The state of ingestion process of the image.",
				Computed:    true,
			},
		},
	}
}

func (AwsWorkspacesImageModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AwsWorkspacesImageModel{}.GetDataSourceSchema().Attributes
}
