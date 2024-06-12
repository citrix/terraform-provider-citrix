// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApplicationIconResourceModel maps the resource schema data.
type ApplicationIconResourceModel struct {
	Id      types.String `tfsdk:"id"`
	RawData types.String `tfsdk:"raw_data"`
}

func GetApplicationIconSchema() schema.Schema {
	return schema.Schema{
		Description: "Resource for managing application icons.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application icon.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"raw_data": schema.StringAttribute{
				Description: "Prepare an icon in ICO format and convert its binary raw data to base64 encoding. Use the base64 encoded string as the value of this attribute.",
				Required:    true,
			},
		},
	}
}

func (r ApplicationIconResourceModel) RefreshPropertyValues(application *citrixorchestration.IconResponseModel) ApplicationIconResourceModel {
	// Overwrite application folder with refreshed state
	r.Id = types.StringValue(application.GetId())
	r.RawData = types.StringValue(application.GetRawData())
	return r
}
