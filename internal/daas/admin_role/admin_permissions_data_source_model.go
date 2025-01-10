// Copyright © 2024. Citrix Systems, Inc.
package admin_role

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AdminPermissionModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	GroupId     types.String `tfsdk:"group_id"`
	GroupName   types.String `tfsdk:"group_name"`
}

func (AdminPermissionModel) GetAdminPermissionModelSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin permission.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin permission.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin permission.",
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "ID of the group to which the permission belongs.",
				Computed:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "Name of the group to which the permission belongs.",
				Computed:    true,
			},
		},
	}
}

type AdminPermissionsDataSourceModel struct {
	Permissions []AdminPermissionModel `tfsdk:"permissions"`
}

func (AdminPermissionsDataSourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source for a list of administrator permissions.",

		Attributes: map[string]schema.Attribute{
			"permissions": schema.ListNestedAttribute{
				Description:  "The list of administrator permissions.",
				Computed:     true,
				NestedObject: AdminPermissionModel{}.GetAdminPermissionModelSchema(),
			},
		},
	}
}

func (r AdminPermissionsDataSourceModel) RefreshPropertyValues(adminPermissions []citrixorchestration.PredefinedPermissionResponseModel) AdminPermissionsDataSourceModel {
	res := []AdminPermissionModel{}
	for _, model := range adminPermissions {
		res = append(res, AdminPermissionModel{
			Id:          types.StringValue(model.GetId()),
			Name:        types.StringValue(model.GetName()),
			Description: types.StringValue(model.GetDescription()),
			GroupId:     types.StringValue(model.GetGroupId()),
			GroupName:   types.StringValue(model.GetGroupName()),
		})
	}

	r.Permissions = res
	return r
}
