// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (AdminScopeModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get details regarding a specific Administrator scope.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Admin Scope.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Admin Scope. For `tenant` scope please use `tenant customer Id` for this field.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the Admin Scope.",
				Computed:    true,
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Indicates whether the Admin Scope is built-in or not.",
				Computed:    true,
			},
			"is_all_scope": schema.BoolAttribute{
				Description: "Indicates whether the Admin Scope is all scope or not.",
				Computed:    true,
			},
			"is_tenant_scope": schema.BoolAttribute{
				Description: "Indicates whether the Admin Scope is tenant scope or not.",
				Computed:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "ID of the tenant to which the Admin Scope belongs.",
				Computed:    true,
			},
			"tenant_name": schema.StringAttribute{
				Description: "Name of the tenant to which the Admin Scope belongs.",
				Computed:    true,
			},
		},
	}
}

func (AdminScopeModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AdminScopeModel{}.GetDataSourceSchema().Attributes
}
