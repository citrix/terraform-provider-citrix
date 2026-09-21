// Copyright © 2026. Citrix Systems, Inc.

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
	ForceDelete  types.Bool   `tfsdk:"force_delete"`
}

func (ResourceLocationModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Manages a Citrix Cloud resource location." +
			"\n\n~> **Please Note** For Citrix Cloud Customer, DaaS Zone permissions are required to manage Citrix Cloud Resource Location." +
			"\n\n~> **Please Note** A resource location cannot be deleted while it still contains Connectors. Set `force_delete` to `true` to have the provider remove those Connectors from Citrix Cloud when the resource location is deleted.",

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
			"force_delete": schema.BoolAttribute{
				Description: "Boolean that indicates any Connectors still in the resource location should be removed from Citrix Cloud on `terraform destroy` action. When `false`, the destroy is blocked while the resource location still contains Connectors. Defaults to `false`." +
					"\n\n~> **Please Note** Only the Citrix Cloud Connector records are removed. The Connector VMs themselves still need to be cleaned up separately." +
					"\n\n~> **Please Note** The force deletion only happens when the resource location is actually deleted, not when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run before a `destroy` to update this value in the resource state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. If setting this field in the same operation that would destroy the resource location, this flag will not work. Additionally when importing a resource location, a successful `terraform apply` is required to set this value in state before it will take effect on a destroy operation.",
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
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

	// force_delete is a provider-only flag with no server representation, so default it when unset (for example on import).
	if r.ForceDelete.IsNull() || r.ForceDelete.IsUnknown() {
		r.ForceDelete = types.BoolValue(false)
	}

	return r
}
