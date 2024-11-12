// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (AdminFolderModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get details regarding a specific admin folder.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the admin folder.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("path")), // Ensures that only one of either Id or Path is provided. It will also cause a validation error if none are specified.
					stringvalidator.LengthAtLeast(1),
				},
			},
			"path": schema.StringAttribute{
				Description: "Path to the admin folder.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin folder.",
				Computed:    true,
			},
			"parent_path": schema.StringAttribute{
				Description: "Path of the parent admin folder.",
				Computed:    true,
			},
			"type": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Set of types of the admin folder.",
				Computed:    true,
			},
			"total_applications": schema.Int64Attribute{
				Description: "Number of applications contained in the admin folder.",
				Computed:    true,
			},
			"total_machine_catalogs": schema.Int64Attribute{
				Description: "Number of machine catalogs contained in the admin folder.",
				Computed:    true,
			},
			"total_application_groups": schema.Int64Attribute{
				Description: "Number of application groups contained in the admin folder.",
				Computed:    true,
			},
			"total_delivery_groups": schema.Int64Attribute{
				Description: "Number of delivery groups contained in the admin folder.",
				Computed:    true,
			},
		},
	}
}
