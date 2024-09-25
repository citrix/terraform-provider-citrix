// Copyright © 2024. Citrix Systems, Inc.
package tags

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagDataSourceModel struct {
	Id                              types.String `tfsdk:"id"`
	Name                            types.String `tfsdk:"name"`
	Description                     types.String `tfsdk:"description"`
	Scopes                          types.Set    `tfsdk:"scopes"` //Set[String]
	AssociatedMachineCount          types.Int32  `tfsdk:"associated_machine_count"`
	AssociatedApplicationCount      types.Int32  `tfsdk:"associated_application_count"`
	AssociatedApplicationGroupCount types.Int32  `tfsdk:"associated_application_group_count"`
	AssociatedMachineCatalogCount   types.Int32  `tfsdk:"associated_machine_catalog_count"`
	AssociatedDeliveryGroupCount    types.Int32  `tfsdk:"associated_delivery_group_count"`
}

func (TagDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of a tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the tag.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the tag.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the tag.",
				Computed:    true,
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"associated_machine_count": schema.Int32Attribute{
				Description: "Number of machines associated with the tag.",
				Computed:    true,
			},
			"associated_application_count": schema.Int32Attribute{
				Description: "Number of applications associated with the tag.",
				Computed:    true,
			},
			"associated_application_group_count": schema.Int32Attribute{
				Description: "Number of application groups associated with the tag.",
				Computed:    true,
			},
			"associated_machine_catalog_count": schema.Int32Attribute{
				Description: "Number of machine catalogs associated with the tag.",
				Computed:    true,
			},
			"associated_delivery_group_count": schema.Int32Attribute{
				Description: "Number of delivery groups associated with the tag.",
				Computed:    true,
			},
		},
	}
}

func (TagDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return TagDataSourceModel{}.GetSchema().Attributes
}

func (r TagDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, tag *citrixorchestration.TagDetailResponseModel) TagDataSourceModel {

	r.Id = types.StringValue(tag.GetId())
	r.Name = types.StringValue(tag.GetName())
	r.Description = types.StringValue(tag.GetDescription())

	remoteScopeIds := []string{}
	for _, scope := range tag.GetScopeReferences() {
		remoteScopeIds = append(remoteScopeIds, scope.GetScopeId())
	}
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, remoteScopeIds)

	// Refresh computed attributes
	r.AssociatedMachineCount = types.Int32Value(tag.GetNumMachines())
	r.AssociatedApplicationCount = types.Int32Value(tag.GetNumApplications())
	r.AssociatedApplicationGroupCount = types.Int32Value(tag.GetNumApplicationGroups())
	r.AssociatedMachineCatalogCount = types.Int32Value(tag.GetNumMachineCatalogs())
	r.AssociatedDeliveryGroupCount = types.Int32Value(tag.GetNumDeliveryGroups())

	return r
}
