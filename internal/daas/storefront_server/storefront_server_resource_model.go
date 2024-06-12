// Copyright © 2024. Citrix Systems, Inc.

package storefront_server

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StoreFrontServerResourceModel maps the resource schema data.
type StoreFrontServerResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Url         types.String `tfsdk:"url"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func (r StoreFrontServerResourceModel) RefreshPropertyValues(sfServer *citrixorchestration.StoreFrontServerResponseModel) StoreFrontServerResourceModel {
	// Overwrite StoreFront server with refreshed state
	r.Id = types.StringValue(sfServer.GetId())
	r.Name = types.StringValue(sfServer.GetName())
	r.Description = types.StringValue(sfServer.GetDescription())
	r.Enabled = types.BoolValue(sfServer.GetEnabled())

	remoteUrl := sfServer.GetUrl()
	planUrl := r.Url.ValueString()
	if remoteUrl[len(remoteUrl)-1] == '/' && (len(planUrl) == 0 || planUrl[len(planUrl)-1] != '/') {
		remoteUrl = remoteUrl[:len(remoteUrl)-1]
	}
	r.Url = types.StringValue(remoteUrl)

	return r
}
