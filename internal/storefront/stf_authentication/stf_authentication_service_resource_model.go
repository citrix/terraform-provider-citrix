// Copyright © 2024. Citrix Systems, Inc.

package stf_authentication

import (
	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"strconv"
)

// SFAuthenticationServiceResourceModel maps the resource schema data.

type STFAuthenticationServiceResourceModel struct {
	SiteId       types.String `tfsdk:"site_id"`
	VirtualPath  types.String `tfsdk:"virtual_path"`
	FriendlyName types.String `tfsdk:"friendly_name"`
}

func (r *STFAuthenticationServiceResourceModel) RefreshPropertyValues(authService *citrixstorefront.STFAuthenticationServiceResponseModel) {
	// Overwrite SFDeploymentResourceModel with refreshed state
	r.SiteId = types.StringValue(strconv.Itoa(*authService.SiteId.Get()))
	r.VirtualPath = types.StringValue(*authService.VirtualPath.Get())
	r.FriendlyName = types.StringValue(*authService.FriendlyName.Get())
}
