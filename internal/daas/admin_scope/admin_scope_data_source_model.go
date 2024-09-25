// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AdminScopeDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	IsBuiltIn     types.Bool   `tfsdk:"is_built_in"`
	IsAllScope    types.Bool   `tfsdk:"is_all_scope"`
	IsTenantScope types.Bool   `tfsdk:"is_tenant_scope"`
	TenantId      types.String `tfsdk:"tenant_id"`
	TenantName    types.String `tfsdk:"tenant_name"`
}

func (r AdminScopeDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminScope *citrixorchestration.ScopeResponseModel) AdminScopeDataSourceModel {

	r.Id = types.StringValue(adminScope.GetId())
	r.Name = types.StringValue(adminScope.GetName())
	r.Description = types.StringValue(adminScope.GetDescription())
	r.IsBuiltIn = types.BoolValue(adminScope.GetIsBuiltIn())
	r.IsAllScope = types.BoolValue(adminScope.GetIsAllScope())
	r.IsTenantScope = types.BoolValue(adminScope.GetIsTenantScope())
	r.TenantId = types.StringValue(adminScope.GetTenantId())
	r.TenantName = types.StringValue(adminScope.GetTenantName())

	return r
}

func GetAdminScopeDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get details regarding a specific Administrator scope.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Admin Scope.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Admin Scope.",
				Optional:    true,
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
