// Copyright © 2023. Citrix Systems, Inc.

package admin_scope

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminScopeDataSourceModel defines the single VDA data model implementation.
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

func (r AdminScopeDataSourceModel) RefreshPropertyValues(adminScope *citrixorchestration.ScopeResponseModel) AdminScopeDataSourceModel {

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
