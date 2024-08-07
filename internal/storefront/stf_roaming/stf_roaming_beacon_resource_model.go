// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type STFRoamingBeaconResourceModel struct {
	Internal types.String `tfsdk:"internal_ip"`
	External types.List   `tfsdk:"external_ips"`
	SiteId   types.Int64  `tfsdk:"site_id"`
}

func (r *STFRoamingBeaconResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, roamInt *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel, roamExt *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel) {

	r.Internal = types.StringValue(roamInt.Internal)

	if roamExt != nil && roamExt.External != nil {
		r.External = util.RefreshListValues(ctx, diagnostics, r.External, roamExt.External)
	}
}

func (STFRoamingBeaconResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- This resource is used to manage the StoreFront Roaming Beacon.",
		Attributes: map[string]schema.Attribute{
			"internal_ip": schema.StringAttribute{
				Description: "Internal IP address of the beacon.",
				Required:    true,
			},
			"external_ips": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "External IP addresses of the beacon.",
				Optional:    true,
			},
			"site_id": schema.Int64Attribute{
				Description: "Site Id of the StoreFront Roaming Service instance.",
				Required:    true,
			},
		},
	}
}

func (STFRoamingBeaconResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFRoamingBeaconResourceModel{}.GetSchema().Attributes
}
