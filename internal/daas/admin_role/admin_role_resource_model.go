// Copyright © 2024. Citrix Systems, Inc.

package admin_role

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminRoleResourceModel maps the resource schema data.
type AdminRoleResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	IsBuiltIn        types.Bool   `tfsdk:"is_built_in"`
	Description      types.String `tfsdk:"description"`
	CanLaunchManage  types.Bool   `tfsdk:"can_launch_manage"`
	CanLaunchMonitor types.Bool   `tfsdk:"can_launch_monitor"`
	Permissions      types.Set    `tfsdk:"permissions"` //Set[string]
}

func (r AdminRoleResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminRole *citrixorchestration.RoleResponseModel) AdminRoleResourceModel {

	// Overwrite admin role with refreshed state
	r.Id = types.StringValue(adminRole.GetId())
	r.Name = types.StringValue(adminRole.GetName())
	r.Description = types.StringValue(adminRole.GetDescription())
	r.IsBuiltIn = types.BoolValue(adminRole.GetIsBuiltIn())
	r.CanLaunchManage = types.BoolValue(adminRole.GetCanLaunchManage())
	r.CanLaunchMonitor = types.BoolValue(adminRole.GetCanLaunchMonitor())

	var permissionListFromRemote []string
	for _, permission := range adminRole.GetPermissions() {
		permissionListFromRemote = append(permissionListFromRemote, permission.GetId())
	}
	r.Permissions = util.StringArrayToStringSet(ctx, diagnostics, permissionListFromRemote)

	return r
}

func (AdminRoleResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Manages an administrator role.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin role.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Flag to determine if the role was built-in or user defined",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"can_launch_manage": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Manage tab on the console. Defaults to `true`. " +
					"\n\n~> **Please Note** This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true), // Default value gets set for an attribute after Validation and before applying configuration changes
			},
			"can_launch_monitor": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Monitor tab on the console. Defaults to `true`. " +
					"\n\n~> **Please Note** This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"permissions": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Permissions to be associated with the admin role. " +
					"\n\n-> **Note** To get a list of supported permissions, please refer to [Admin Predefined Permissions for Cloud](https://developer-docs.citrix.com/en-us/citrix-daas-service-apis/citrix-daas-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions) and [Admin Predefined Permissions for On-Premise](https://developer-docs.citrix.com/en-us/citrix-virtual-apps-desktops/citrix-cvad-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions).",
				Required: true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (AdminRoleResourceModel) GetAttributes() map[string]schema.Attribute {
	return AdminRoleResourceModel{}.GetSchema().Attributes
}
