// Copyright © 2024. Citrix Systems, Inc.

package autoscale_plugin_template

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AutoscalePluginTemplateResourceModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Dates types.Set    `tfsdk:"dates"` // Set[string]
}

func (AutoscalePluginTemplateResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an autoscale plugin template.",
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
					validators.AlsoRequiresOnStringValues(
						[]string{
							string(citrixorchestration.AUTOSCALEPLUGINTYPE_HOLIDAY),
						},
						path.MatchRelative().AtParent().AtName("dates"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dates": schema.SetAttribute{
				Description: "Date range for the autoscale holiday plugin template.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (AutoscalePluginTemplateResourceModel) GetAttributes() map[string]schema.Attribute {
	return AutoscalePluginTemplateResourceModel{}.GetSchema().Attributes
}

func (AutoscalePluginTemplateResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r AutoscalePluginTemplateResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, autoscalePluginTemplate *citrixorchestration.AutoscalePluginTemplateResponseModel) AutoscalePluginTemplateResourceModel {
	r.Name = types.StringValue(autoscalePluginTemplate.GetName())
	r.Type = types.StringValue(string(autoscalePluginTemplate.GetType()))
	r.Dates = util.StringArrayToStringSet(ctx, diagnostics, autoscalePluginTemplate.GetDates())

	return r
}
