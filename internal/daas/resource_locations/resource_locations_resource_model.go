// Copyright © 2023. Citrix Systems, Inc.

package resource_locations

import (
	ccresourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResourceLocationResourceModel maps the resource schema data.
type ResourceLocationResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	InternalOnly types.Bool   `tfsdk:"internal_only"`
	TimeZone     types.String `tfsdk:"time_zone"`
}

func (r ResourceLocationResourceModel) RefreshPropertyValues(ccResourceLocation *ccresourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel) ResourceLocationResourceModel {

	// Overwrite resource location with refreshed state
	r.Id = types.StringValue(ccResourceLocation.GetId())
	r.Name = types.StringValue(ccResourceLocation.GetName())
	r.InternalOnly = types.BoolValue(ccResourceLocation.GetInternalOnly())
	r.TimeZone = types.StringValue(ccResourceLocation.GetTimeZone())

	return r
}
