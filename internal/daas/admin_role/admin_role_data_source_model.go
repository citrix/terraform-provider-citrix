// Copyright © 2024. Citrix Systems, Inc.
package admin_role

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (AdminRoleModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source of an administrator role.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin role.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Path is provided. It will also cause a validation error if none are specified.
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin role.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin role.",
				Computed:    true,
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Flag to determine if the role was built-in or user defined",
				Computed:    true,
			},
			"can_launch_manage": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Manage tab on the console. Defaults to `true`. " +
					"\n\n~> **Please Note** This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`.",
				Computed: true,
			},
			"can_launch_monitor": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Monitor tab on the console. Defaults to `true`. " +
					"\n\n~> **Please Note** This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`.",
				Computed: true,
			},
			"permissions": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Permissions to be associated with the admin role. " +
					"\n\n-> **Note** To get a list of supported permissions, please refer to [Admin Predefined Permissions for Cloud](https://developer-docs.citrix.com/en-us/citrix-daas-service-apis/citrix-daas-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions) and [Admin Predefined Permissions for On-Premise](https://developer-docs.citrix.com/en-us/citrix-virtual-apps-desktops/citrix-cvad-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions).",
				Computed: true,
			},
		},
	}
}

func (AdminRoleModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AdminRoleModel{}.GetDataSourceSchema().Attributes
}
