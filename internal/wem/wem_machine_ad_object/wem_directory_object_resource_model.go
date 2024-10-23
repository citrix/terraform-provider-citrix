package wem_machine_ad_object

import (
	"context"
	"regexp"
	"strconv"

	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WemDirectoryResourceModel struct {
	Id        types.String `tfsdk:"id"`
	CatalogId types.String `tfsdk:"machine_catalog_id"`
	SiteId    types.Int64  `tfsdk:"configuration_set_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

func (WemDirectoryResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "WEM --- Manages machine-level AD objects within a WEM deployment." +
			"\n\n~> **Disclaimer** This feature is supported for Citrix Cloud customers, and will be made available for On-Premises soon." +
			"\n\n~> **Warning**  Having more than one Directory Object with the same Catalog ID is not allowed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the directory object.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"machine_catalog_id": schema.StringAttribute{
				Description: "GUID identifier of the machine catalog.",
				Required:    true,
				Validators: []validator.String{
					validator.String(
						stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configuration_set_id": schema.Int64Attribute{
				Description: "Identifier of the site to which the machine-level AD object belongs.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates whether the machine-level AD object is enabled.",
				Required:    true,
			},
		},
	}
}

func (WemDirectoryResourceModel) GetAttributes() map[string]schema.Attribute {
	return WemDirectoryResourceModel{}.GetSchema().Attributes
}

func (r WemDirectoryResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, wemSite *citrixwemservice.MachineModel) WemDirectoryResourceModel {
	r.Id = types.StringValue(strconv.FormatInt(wemSite.GetId(), 10))
	r.CatalogId = types.StringValue(wemSite.GetSid())
	r.SiteId = types.Int64Value(wemSite.GetSiteId())
	r.Enabled = types.BoolValue(wemSite.GetEnabled())
	return r
}
