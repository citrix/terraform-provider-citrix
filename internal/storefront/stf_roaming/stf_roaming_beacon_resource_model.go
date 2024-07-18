// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type STFRoamingBeaconResourceModel struct {
	Internal types.String `tfsdk:"internal_ip"`
	External types.List   `tfsdk:"external_ips"`
	SiteId   types.Int64  `tfsdk:"site_id"`
}

func (*stfRoamingBeaconResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource is used to manage the StoreFront Roaming Beacon.",
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
