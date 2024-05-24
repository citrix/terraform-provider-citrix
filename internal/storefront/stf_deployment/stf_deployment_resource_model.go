// Copyright © 2023. Citrix Systems, Inc.

package stf_deployment

import (
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SFDeploymentResourceModel maps the resource schema data.
type STFDeploymentResourceModel struct {
	SiteId      types.String `tfsdk:"site_id"`
	HostBaseUrl types.String `tfsdk:"host_base_url"`
}

func (r *STFDeploymentResourceModel) RefreshPropertyValues(deployment *citrixstorefront.STFDeploymentDetailModel) {
	// Overwrite SFDeploymentResourceModel with refreshed state
	r.SiteId = types.StringValue(strconv.Itoa(int(*deployment.SiteId.Get())))
	r.HostBaseUrl = types.StringValue(strings.TrimRight(*deployment.HostBaseUrl.Get(), "/"))

}
