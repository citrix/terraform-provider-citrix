// Copyright © 2023. Citrix Systems, Inc.

package admin_scope

import (
	"context"
	"reflect"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AdminScopeResourceModel maps the resource schema data.
type AdminScopeResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ScopedObjects types.List   `tfsdk:"scoped_objects"` //List[ScopedObjectsModel]
}

type ScopedObjectsModel struct {
	ObjectType types.String `tfsdk:"object_type"`
	Object     types.String `tfsdk:"object"`
}

func (ScopedObjectsModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"object_type": schema.StringAttribute{
				Description: "Type of the scoped object. Allowed values are: `HypervisorConnection`, `MachineCatalog`, `DeliveryGroup`, `ApplicationGroup`, `Tag`, `PolicySet` and `Unknown`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Unknown",
						"HypervisorConnection",
						"MachineCatalog",
						"DeliveryGroup",
						"ApplicationGroup",
						"Tag",
						"PolicySet"),
				},
			},
			"object": schema.StringAttribute{
				Description: "Name of an existing object under the object type to be added to the scope.",
				Required:    true,
			},
		},
	}
}

func (ScopedObjectsModel) GetAttributes() map[string]schema.Attribute {
	return ScopedObjectsModel{}.GetSchema().Attributes
}

func (r AdminScopeResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminScope *citrixorchestration.ScopeResponseModel, scopedObjects []citrixorchestration.ScopedObjectResponseModel) AdminScopeResourceModel {

	// Overwrite admin scope with refreshed state
	r.Id = types.StringValue(adminScope.GetId())
	r.Name = types.StringValue(adminScope.GetName())
	r.ScopedObjects = util.TypedArrayToObjectList[ScopedObjectsModel](ctx, diagnostics, r.refreshScopedObjects(ctx, diagnostics, scopedObjects))

	if adminScope.GetDescription() != "" || !r.Description.IsNull() {
		r.Description = types.StringValue(adminScope.GetDescription())
	}

	return r
}

func (r AdminScopeResourceModel) refreshScopedObjects(ctx context.Context, diagnostics *diag.Diagnostics, scopedObjectsFromRemote []citrixorchestration.ScopedObjectResponseModel) []ScopedObjectsModel {

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
	for _, scopedObject := range util.ObjectListToTypedArray[ScopedObjectsModel](ctx, diagnostics, r.ScopedObjects) {
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

func GetAdminScopeSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Manages an administrator scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin scope.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin scope.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin scope.",
				Optional:    true,
			},
			"scoped_objects": schema.ListNestedAttribute{
				Description: "List of scoped objects to be associated with the admin scope.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: ScopedObjectsModel{}.GetSchema(),
			},
		},
	}
}
