// Copyright © 2023. Citrix Systems, Inc.

package admin_scope

import (
	"reflect"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminScopeResourceModel maps the resource schema data.
type AdminScopeResourceModel struct {
	Id            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	ScopedObjects []ScopedObjectsModel `tfsdk:"scoped_objects"`
}

type ScopedObjectsModel struct {
	ObjectType types.String `tfsdk:"object_type"`
	Object     types.String `tfsdk:"object"`
}

func (r AdminScopeResourceModel) RefreshPropertyValues(adminScope *citrixorchestration.ScopeResponseModel, scopedObjects []citrixorchestration.ScopedObjectResponseModel) AdminScopeResourceModel {

	// Overwrite admin scope with refreshed state
	r.Id = types.StringValue(adminScope.GetId())
	r.Name = types.StringValue(adminScope.GetName())
	r.ScopedObjects = r.refreshScopedObjects(scopedObjects)

	if adminScope.GetDescription() != "" || !r.Description.IsNull() {
		r.Description = types.StringValue(adminScope.GetDescription())
	}

	return r
}

func (r AdminScopeResourceModel) refreshScopedObjects(scopedObjectsFromRemote []citrixorchestration.ScopedObjectResponseModel) []ScopedObjectsModel {

	type RemoteScopedObjectTracker struct {
		ObjectType types.String
		IsVisited  bool
	}

	// Create a map of ObjectName -> RemoteScopedObjectTracker from the scoped objects returned from remote
	scopedObjectMapFromRemote := map[string]*RemoteScopedObjectTracker{}
	for _, scopedObjectFromRemote := range scopedObjectsFromRemote {
		scopedObjectMapFromRemote[scopedObjectFromRemote.Object.GetName()] = &RemoteScopedObjectTracker{
			ObjectType: types.StringValue(reflect.ValueOf(scopedObjectFromRemote.GetObjectType()).String()),
			IsVisited:  false,
		}
	}

	// Prepare the scoped objects list to be stored in the state
	var scopedObjectsForState []ScopedObjectsModel
	for _, scopedObject := range r.ScopedObjects {
		scopedObjectFromRemote, exists := scopedObjectMapFromRemote[scopedObject.Object.ValueString()]
		if !exists {
			// If scoped object is not present in the remote, then don't add it to the state
			continue
		}

		scopedObjectsForState = append(scopedObjectsForState, ScopedObjectsModel{
			ObjectType: scopedObjectFromRemote.ObjectType,
			Object:     scopedObject.Object,
		})

		scopedObjectMapFromRemote[scopedObject.Object.ValueString()].IsVisited = true
	}

	// Add all the scoped objects from remote which are not present in the state
	for scopedObjectName, scopedObjectType := range scopedObjectMapFromRemote {
		if !scopedObjectType.IsVisited {
			scopedObjectsForState = append(scopedObjectsForState, ScopedObjectsModel{
				ObjectType: scopedObjectType.ObjectType,
				Object:     types.StringValue(scopedObjectName),
			})
		}
	}

	return scopedObjectsForState
}
