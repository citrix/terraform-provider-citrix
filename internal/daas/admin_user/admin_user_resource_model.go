// Copyright © 2023. Citrix Systems, Inc.

package admin_user

import (
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminUserResourceModel maps the resource schema data.
type AdminUserResourceModel struct {
	Id         types.String  `tfsdk:"id"`
	Name       types.String  `tfsdk:"name"`
	DomainName types.String  `tfsdk:"domain_name"`
	Rights     []RightsModel `tfsdk:"rights"`
	IsEnabled  types.Bool    `tfsdk:"is_enabled"`
}

type RightsModel struct {
	Role  types.String `tfsdk:"role"`
	Scope types.String `tfsdk:"scope"`
}

func (r AdminUserResourceModel) RefreshPropertyValues(adminUser *citrixorchestration.AdministratorResponseModel) AdminUserResourceModel {
	// Overwrite admin user data with refreshed state
	userDetails := adminUser.GetUser()
	userFQDN := strings.Split(userDetails.GetSamName(), "\\")

	r.Id = types.StringValue(userDetails.GetSid())
	r.Name = types.StringValue(userFQDN[len(userFQDN)-1])
	r.DomainName = types.StringValue(userDetails.GetDomain())
	r.Rights = r.refreshRights(adminUser.GetScopesAndRoles())
	r.IsEnabled = types.BoolValue(adminUser.GetEnabled())

	return r
}

func (r AdminUserResourceModel) refreshRights(rightsFromRemote []citrixorchestration.AdministratorRightResponseModel) []RightsModel {

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
	for _, right := range r.Rights {
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
