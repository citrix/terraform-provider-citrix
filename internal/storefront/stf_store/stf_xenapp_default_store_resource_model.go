// Copyright © 2024. Citrix Systems, Inc.

package stf_store

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// STFXenappDefaultStoreResourceModel maps the resource schema data.
type STFXenappDefaultStoreResourceModel struct {
	StoreVirtualPath types.String `tfsdk:"store_virtual_path"` // The Virtual Path of the StoreFront Default Store for XenApp Service.
	StoreSiteID      types.String `tfsdk:"store_site_id"`      // The Site ID of the StoreFront Default Store for XenApp Service.
}

func (*stfXenappDefaultStoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Default Storefront Store for XenApp Service.",
		Attributes: map[string]schema.Attribute{
			"store_site_id": schema.StringAttribute{
				Description: "The Site ID of the StoreFront Default Store for XenApp Service.",
				Required:    true,
			},
			"store_virtual_path": schema.StringAttribute{
				Description: "The Virtual Path of the StoreFront Default Store for XenApp Service.",
				Required:    true,
			},
		},
	}
}

func (r *STFXenappDefaultStoreResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, defaultStore models.STFPna) {
	if defaultStore.FeatureData.SiteID.IsSet() {
		r.StoreSiteID = types.StringValue(*defaultStore.FeatureData.SiteID.Get())
	}
	if defaultStore.FeatureData.VirtualPath.IsSet() && *defaultStore.DefaultPnaService.Get() {
		r.StoreVirtualPath = types.StringValue(*defaultStore.FeatureData.VirtualPath.Get())
	}
}
