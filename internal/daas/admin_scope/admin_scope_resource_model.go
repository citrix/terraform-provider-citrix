// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminScopeResourceModel maps the resource schema data.
type AdminScopeResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r AdminScopeResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminScope *citrixorchestration.ScopeResponseModel) AdminScopeResourceModel {

	// Overwrite admin scope with refreshed state
	r.Id = types.StringValue(adminScope.GetId())
	r.Name = types.StringValue(adminScope.GetName())
	r.Description = types.StringValue(adminScope.GetDescription())

	return r
}

func (AdminScopeResourceModel) GetSchema() schema.Schema {
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
		},
	}
}

func (AdminScopeResourceModel) GetAttributes() map[string]schema.Attribute {
	return AdminScopeResourceModel{}.GetSchema().Attributes
}
