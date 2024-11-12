// Copyright © 2024. Citrix Systems, Inc.

package storefront_server

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (StoreFrontServerResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a StoreFront server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the StoreFront server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the StoreFront server.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the StoreFront server.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL for connecting to the StoreFront server.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates if the StoreFront server is enabled. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
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
