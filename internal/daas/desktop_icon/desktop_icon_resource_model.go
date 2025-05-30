// Copyright © 2024. Citrix Systems, Inc.

package desktop_icon

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DesktopIconResourceModel maps the resource schema data.
type DesktopIconResourceModel struct {
	Id       types.String `tfsdk:"id"`
	RawData  types.String `tfsdk:"raw_data"`
	FilePath types.String `tfsdk:"file_path"`
}

func (DesktopIconResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Resource for managing desktop icons. \n\n-> **Note** Please use just one icon resource per icon. Having multiple icon resources with the same icon data will result in inconsistencies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the desktop icon.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"raw_data": schema.StringAttribute{
				Description: "Prepare an icon in ICO format and convert its binary raw data to base64 encoding. Use the base64 encoded string as the value of this attribute. Exactly one of `raw_data` and `file_path` is required.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("file_path"),
					),
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Description: "Path to the icon file. Exactly one of `raw_data` and `file_path` is required.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (DesktopIconResourceModel) GetAttributes() map[string]schema.Attribute {
	return DesktopIconResourceModel{}.GetSchema().Attributes
}

func (DesktopIconResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r DesktopIconResourceModel) RefreshPropertyValues(desktop *citrixorchestration.IconResponseModel) DesktopIconResourceModel {
	// Overwrite desktop folder with refreshed state
	r.Id = types.StringValue(desktop.GetId())
	if r.FilePath.IsNull() {
		r.RawData = types.StringValue(desktop.GetRawData())
	}
	return r
}
