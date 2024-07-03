// Copyright © 2024. Citrix Systems, Inc.

package zone

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ZoneResourceModel maps the resource schema data.
type ZoneResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Metadata    types.List   `tfsdk:"metadata"` // []utils.NameValueStringPairModel
}

func (ZoneResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a zone.\nFor cloud DDC, Zones and Cloud Connectors are managed only by Citrix Cloud. Ensure you have a resource location manually created and connectors deployed in it. You may then apply or import the zone using the zone Id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the zone.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the zone.\nFor Cloud DDC, ensure this matches the name of the existing zone that needs to be used.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the zone.\nFor Cloud DDC, ensure this matches the description of the existing zone that needs to be used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"metadata": schema.ListNestedAttribute{
				Description:  "Metadata of the zone. Cannot be modified in DaaS cloud.",
				Optional:     true,
				NestedObject: util.NameValueStringPairModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (ZoneResourceModel) GetAttributes() map[string]schema.Attribute {
	return ZoneResourceModel{}.GetSchema().Attributes
}

func (r ZoneResourceModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, zone *citrixorchestration.ZoneDetailResponseModel, onpremises bool) ZoneResourceModel {
	// Overwrite zone with refreshed state
	r.Id = types.StringValue(zone.GetId())
	r.Name = types.StringValue(zone.GetName())
	r.Description = types.StringValue(zone.GetDescription())

	// Set optional values
	metadata := zone.GetMetadata()
	if onpremises && (!r.Metadata.IsNull() || len(metadata) > 0) {
		// Cloud customers cannot modify Zone metadata because of CC zone syncing
		// On-Premises customers can have either nil value for metadata, or provide an empty array
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diags, util.ParseNameValueStringPairToPluginModel(metadata))
	}

	return r
}
