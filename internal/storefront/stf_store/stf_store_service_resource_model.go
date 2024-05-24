// Copyright © 2023. Citrix Systems, Inc.

package stf_store

import (
	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SFStoreServiceResourceModel maps the resource schema data.
type STFStoreServiceResourceModel struct {
	VirtualPath           types.String `tfsdk:"virtual_path"`
	SiteId                types.String `tfsdk:"site_id"`
	FriendlyName          types.String `tfsdk:"friendly_name"`
	AuthenticationService types.String `tfsdk:"authentication_service"`
	Anonymous             types.Bool   `tfsdk:"anonymous"`
	LoadBalance           types.Bool   `tfsdk:"load_balance"`
	FarmConfig            *FarmConfig  `tfsdk:"farm_config"`
}

type FarmConfig struct {
	FarmName types.String   `tfsdk:"farm_name"`
	FarmType types.String   `tfsdk:"farm_type"`
	Servers  []types.String `tfsdk:"servers"`
}

func (r *STFStoreServiceResourceModel) RefreshPropertyValues(storeservice *citrixstorefront.STFStoreDetailModel) {
	// Overwrite SFStoreServiceResourceModel with refreshed state
	r.VirtualPath = types.StringValue(*storeservice.VirtualPath.Get())
	r.SiteId = types.StringValue(strconv.Itoa(*storeservice.SiteId.Get()))
	r.FriendlyName = types.StringValue(*storeservice.FriendlyName.Get())
}
