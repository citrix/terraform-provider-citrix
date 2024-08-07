// Copyright © 2024. Citrix Systems, Inc.

package stf_deployment

import (
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (STFDeploymentResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront Deployment.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront deployment. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_base_url": schema.StringAttribute{
				Description: "Url used to access the StoreFront server group.",
				Required:    true,
			},
		},
	}
}

func (STFDeploymentResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFDeploymentResourceModel{}.GetSchema().Attributes
}
