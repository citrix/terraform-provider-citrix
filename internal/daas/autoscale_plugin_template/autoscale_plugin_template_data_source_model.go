// Copyright © 2024. Citrix Systems, Inc.

package autoscale_plugin_template

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (AutoscalePluginTemplateResourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing autoscale plugin template.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the autoscale plugin template.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the autoscale plugin template. Only template type `Holiday` is supported.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(citrixorchestration.AUTOSCALEPLUGINTYPE_HOLIDAY)),
				},
			},
			"dates": schema.SetAttribute{
				Description: "Date range for the autoscale holiday plugin template.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}
