// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	ccresourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResourceLocationModel maps the resource schema data.
type ResourceLocationModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	InternalOnly types.Bool   `tfsdk:"internal_only"`
	TimeZone     types.String `tfsdk:"time_zone"`
}

func (ResourceLocationModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Manages a Citrix Cloud resource location." +
			"\n\n~> **Please Note** For Citrix Cloud Customer, DaaS Zone permissions are required to manage Citrix Cloud Resource Location.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the resource location.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource location.",
				Required:    true,
			},
			"internal_only": schema.BoolAttribute{
				Description: "Flag to determine if the resource location can only be used internally. Defaults to `false`.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"time_zone": schema.StringAttribute{
				Description: "Timezone associated with the resource location. Please refer to the `Timezone` column in the following [table](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/default-time-zones?view=windows-11#time-zones) for allowed values.",
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString("GMT Standard Time"),
			},
		},
	}
}

func (ResourceLocationModel) GetAttributes() map[string]schema.Attribute {
	return ResourceLocationModel{}.GetSchema().Attributes
}

func (ResourceLocationModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r ResourceLocationModel) RefreshPropertyValues(ccResourceLocation *ccresourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel) ResourceLocationModel {

	// Overwrite resource location with refreshed state
	r.Id = types.StringValue(ccResourceLocation.GetId())
	r.Name = types.StringValue(ccResourceLocation.GetName())
	r.InternalOnly = types.BoolValue(ccResourceLocation.GetInternalOnly())
	r.TimeZone = types.StringValue(ccResourceLocation.GetTimeZone())

	return r
}
