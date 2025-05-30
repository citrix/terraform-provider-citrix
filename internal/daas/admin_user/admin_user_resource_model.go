// Copyright © 2024. Citrix Systems, Inc.

package admin_user

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminUserResourceModel maps the resource schema data.
type AdminUserResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DomainName types.String `tfsdk:"domain_name"`
	Rights     types.List   `tfsdk:"rights"` //List[RightsModel]
	IsEnabled  types.Bool   `tfsdk:"is_enabled"`
}

type RightsModel struct {
	Role  types.String `tfsdk:"role"`
	Scope types.String `tfsdk:"scope"`
}

func (r AdminUserResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, adminUser *citrixorchestration.AdministratorResponseModel) AdminUserResourceModel {
	// Overwrite admin user data with refreshed state
	userDetails := adminUser.GetUser()
	userFQDN := strings.Split(userDetails.GetSamName(), "\\")

	r.Id = types.StringValue(userDetails.GetSid())
	r.Name = types.StringValue(userFQDN[len(userFQDN)-1])
	r.DomainName = types.StringValue(userDetails.GetDomain())
	if isResource {
		r.Rights = util.TypedArrayToObjectList[RightsModel](ctx, diagnostics, r.refreshRights(ctx, diagnostics, adminUser.GetScopesAndRoles()))
	} else {
		r.Rights = util.DataSourceTypedArrayToObjectList[RightsModel](ctx, diagnostics, r.refreshRights(ctx, diagnostics, adminUser.GetScopesAndRoles()))
	}
	r.IsEnabled = types.BoolValue(adminUser.GetEnabled())

	return r
}

func (RightsModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				Description: "Name of the role to be associated with the admin user.",
				Required:    true,
			},
			"scope": schema.StringAttribute{
				Description: "Name of the scope to be associated with the admin user.",
				Required:    true,
			},
		},
	}
}

func (RightsModel) GetAttributes() map[string]schema.Attribute {
	return RightsModel{}.GetSchema().Attributes
}

func (r AdminUserResourceModel) refreshRights(ctx context.Context, diagnostics *diag.Diagnostics, rightsFromRemote []citrixorchestration.AdministratorRightResponseModel) []RightsModel {

	type RemoteRightsTracker struct {
		RoleName  types.String
		ScopeName types.String
		IsVisited bool
	}

	//Create a map of RoleName+ScopeName -> RemoteRightsTracker from the rights returned from remote
	rightsMapFromRemote := map[string]*RemoteRightsTracker{}
	for _, rightFromRemote := range rightsFromRemote {
		role := rightFromRemote.GetRole()
		scope := rightFromRemote.GetScope()

		rightsMapFromRemote[strings.ToLower(role.GetName()+scope.GetName())] = &RemoteRightsTracker{
			RoleName:  types.StringValue(role.GetName()),
			ScopeName: types.StringValue(scope.GetName()),
			IsVisited: false,
		}
	}

	// Prepare the rights list to be stored in the state
	var rightsForState []RightsModel
	for _, right := range util.ObjectListToTypedArray[RightsModel](ctx, diagnostics, r.Rights) {
		rightFromRemote, exists := rightsMapFromRemote[strings.ToLower(right.Role.ValueString()+right.Scope.ValueString())]
		if !exists {
			// If right is not present in the remote, then don't add it to the state
			continue
		}

		rightsForState = append(rightsForState, RightsModel{
			Role:  right.Role,
			Scope: right.Scope,
		})
		rightFromRemote.IsVisited = true
	}

	// Add the rights from remote which are not present in the state
	for _, right := range rightsMapFromRemote {
		if !right.IsVisited {
			rightsForState = append(rightsForState, RightsModel{
				Role:  right.RoleName,
				Scope: right.ScopeName,
			})
		}
	}

	return rightsForState
}

func (AdminUserResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Manages an administrator user for on-premise environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of an existing user in the active directory.",
				Required:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "Name of the domain that the user is a part of. For example, if the domain is `example.com`, then provide the value `example` for this field.",
				Required:    true,
			},
			"rights": schema.ListNestedAttribute{
				Description: "Rights to be associated with the admin user.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: RightsModel{}.GetSchema(),
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Flag to determine if the administrator is to be enabled or not.",
				Optional:    true,
			},
		},
	}
}

func (AdminUserResourceModel) GetAttributes() map[string]schema.Attribute {
	return AdminUserResourceModel{}.GetSchema().Attributes
}

func (AdminUserResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{
		"name": true,
	}
}
