// Copyright © 2024. Citrix Systems, Inc.

package zone

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ZoneResourceModel maps the resource schema data.
type ZoneResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ResourceLocationId types.String `tfsdk:"resource_location_id"`
	Description        types.String `tfsdk:"description"`
	Metadata           types.List   `tfsdk:"metadata"` // []utils.NameValueStringPairModel
}

func (ZoneResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the zone.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the zone. " +
					"\n\n-> **Note** For Citrix Cloud Customer, `name` is not allowed to be used for creating zone and is computed only. Use `resource_location_id` to create zone Instead.",
				Optional: true,
				Computed: true,
			},
			"resource_location_id": schema.StringAttribute{
				Description: "GUID identifier off the resource location the zone belongs to. Only applies to Citrix Cloud customers. " +
					"\n\n-> **Note** When using `resource_location_id`, ensure that the resource location is already created, or the value must be a reference to a [`citrix_cloud_resource_location`](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/cloud_resource_location)'s `id` property.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name"), path.MatchRoot("resource_location_id")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the zone. " +
					"\n\n-> **Note** For Citrix Cloud customer, ensure this matches the description of the existing zone behind the `resource_location_id` that needs to be used.",
				Optional: true,
				Computed: true,
			},
			"metadata": schema.ListNestedAttribute{
				Description:  "Metadata of the zone. Cannot be modified for Citrix Cloud customer.",
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

	if zone.ResourceLocation != nil {
		resourceLocation := zone.GetResourceLocation()
		r.ResourceLocationId = types.StringValue(resourceLocation.GetId())
	} else {
		r.ResourceLocationId = types.StringNull()
	}

	// Set optional values
	metadata := zone.GetMetadata()
	if onpremises && (!r.Metadata.IsNull() || len(metadata) > 0) {
		// Cloud customers cannot modify Zone metadata because of CC zone syncing
		// On-Premises customers can have either nil value for metadata, or provide an empty array
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diags, util.ParseNameValueStringPairToPluginModel(metadata))
	}

	return r
}
