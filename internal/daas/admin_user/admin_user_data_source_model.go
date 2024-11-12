// Copyright © 2024. Citrix Systems, Inc.

package admin_user

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (RightsModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				Description: "Name of the role to be associated with the admin user.",
				Computed:    true,
			},
			"scope": schema.StringAttribute{
				Description: "Name of the scope to be associated with the admin user.",
				Computed:    true,
			},
		},
	}
}

func (RightsModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return RightsModel{}.GetDataSourceSchema().Attributes
}

func (AdminUserResourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source of an administrator user for on-premise environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin user.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of an existing user in the active directory.",
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "Name of the domain that the user is a part of. For example, if the domain is `example.com`, then provide the value `example` for this field.",
				Computed:    true,
			},
			"rights": schema.ListNestedAttribute{
				Description:  "Rights to be associated with the admin user.",
				Computed:     true,
				NestedObject: RightsModel{}.GetDataSourceSchema(),
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Flag to determine if the administrator is to be enabled or not.",
				Computed:    true,
			},
		},
	}
}

func (AdminUserResourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AdminUserResourceModel{}.GetDataSourceSchema().Attributes
}
