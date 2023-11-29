// Copyright © 2023. Citrix Systems, Inc.

package models

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ZoneResourceModel maps the resource schema data.
type ZoneResourceModel struct {
	Id          types.String                     `tfsdk:"id"`
	Name        types.String                     `tfsdk:"name"`
	Description types.String                     `tfsdk:"description"`
	Metadata    *[]util.NameValueStringPairModel `tfsdk:"metadata"`
}

func (r ZoneResourceModel) RefreshPropertyValues(zone *citrixorchestration.ZoneDetailResponseModel, onpremises bool) ZoneResourceModel {
	// Overwrite zone with refreshed state
	r.Id = types.StringValue(zone.GetId())
	r.Name = types.StringValue(zone.GetName())

	// Set optional values
	if zone.GetDescription() != "" {
		r.Description = types.StringValue(zone.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	metadata := zone.GetMetadata()
	if onpremises && (r.Metadata != nil || len(metadata) > 0) {
		// Cloud customers cannot modify Zone metadata because of CC zone syncing
		// On-Prem customers can have either nil value for metadata, or provide an empty array
		r.Metadata = util.ParseNameValueStringPairToPluginModel(metadata)
	}

	return r
}
