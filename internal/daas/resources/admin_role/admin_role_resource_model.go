// Copyright © 2023. Citrix Systems, Inc.

package admin_role

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminRoleResourceModel maps the resource schema data.
type AdminRoleResourceModel struct {
	Id               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	IsBuiltIn        types.Bool     `tfsdk:"is_built_in"`
	Description      types.String   `tfsdk:"description"`
	CanLaunchManage  types.Bool     `tfsdk:"can_launch_manage"`
	CanLaunchMonitor types.Bool     `tfsdk:"can_launch_monitor"`
	Permissions      []types.String `tfsdk:"permissions"`
}

func (r AdminRoleResourceModel) RefreshPropertyValues(adminRole *citrixorchestration.RoleResponseModel) AdminRoleResourceModel {

	// Overwrite admin role with refreshed state
	r.Id = types.StringValue(adminRole.GetId())
	r.Name = types.StringValue(adminRole.GetName())
	if adminRole.GetDescription() != "" || !r.Description.IsNull() {
		r.Description = types.StringValue(adminRole.GetDescription())
	}
	r.IsBuiltIn = types.BoolValue(adminRole.GetIsBuiltIn())
	r.CanLaunchManage = types.BoolValue(adminRole.GetCanLaunchManage())
	r.CanLaunchMonitor = types.BoolValue(adminRole.GetCanLaunchMonitor())

	var permissionListFromRemote []string
	for _, permission := range adminRole.GetPermissions() {
		permissionListFromRemote = append(permissionListFromRemote, permission.GetId())
	}
	r.Permissions = util.RefreshList(r.Permissions, permissionListFromRemote)

	return r
}
