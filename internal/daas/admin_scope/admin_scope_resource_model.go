// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminScopeModel maps the resource schema data.
type AdminScopeModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	IsTenantScope types.Bool   `tfsdk:"is_tenant_scope"`
	IsBuiltIn     types.Bool   `tfsdk:"is_built_in"`
	IsAllScope    types.Bool   `tfsdk:"is_all_scope"`
	TenantId      types.String `tfsdk:"tenant_id"`
	TenantName    types.String `tfsdk:"tenant_name"`
}

func (r AdminScopeModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminScope *citrixorchestration.ScopeResponseModel) AdminScopeModel {

	// Overwrite admin scope with refreshed state
	r.Id = types.StringValue(adminScope.GetId())
	r.Name = types.StringValue(adminScope.GetName())
	r.Description = types.StringValue(adminScope.GetDescription())
	r.IsTenantScope = types.BoolValue(adminScope.GetIsTenantScope())

	r.IsBuiltIn = types.BoolValue(adminScope.GetIsBuiltIn())
	r.IsAllScope = types.BoolValue(adminScope.GetIsAllScope())
	r.TenantId = types.StringValue(adminScope.GetTenantId())
	r.TenantName = types.StringValue(adminScope.GetTenantName())

	return r
}

func (AdminScopeModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Manages an administrator scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin scope.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin scope.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin scope.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"is_tenant_scope": schema.BoolAttribute{
				Description: "Indicates whether the admin scope is a tenant scope. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Indicates whether the Admin Scope is built-in or not.",
				Computed:    true,
			},
			"is_all_scope": schema.BoolAttribute{
				Description: "Indicates whether the Admin Scope is all scope or not.",
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

func (AdminScopeModel) GetAttributes() map[string]schema.Attribute {
	return AdminScopeModel{}.GetSchema().Attributes
}
