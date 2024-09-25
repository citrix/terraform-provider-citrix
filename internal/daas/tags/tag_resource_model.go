// Copyright © 2024. Citrix Systems, Inc.
package tags

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagResourceModel struct {
	Id                              types.String `tfsdk:"id"`
	Name                            types.String `tfsdk:"name"`
	Description                     types.String `tfsdk:"description"`
	Scopes                          types.Set    `tfsdk:"scopes"` // Set[String]
	AssociatedMachineCount          types.Int32  `tfsdk:"associated_machine_count"`
	AssociatedApplicationCount      types.Int32  `tfsdk:"associated_application_count"`
	AssociatedApplicationGroupCount types.Int32  `tfsdk:"associated_application_group_count"`
	AssociatedMachineCatalogCount   types.Int32  `tfsdk:"associated_machine_catalog_count"`
	AssociatedDeliveryGroupCount    types.Int32  `tfsdk:"associated_delivery_group_count"`
}

func (TagResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the tag.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the tag.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The set of IDs of the scopes applied on the tag.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
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

func (TagResourceModel) GetAttributes() map[string]schema.Attribute {
	return TagResourceModel{}.GetSchema().Attributes
}

func (r TagResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixclient.CitrixDaasClient, tag *citrixorchestration.TagDetailResponseModel) TagResourceModel {

	r.Id = types.StringValue(tag.GetId())
	r.Name = types.StringValue(tag.GetName())
	r.Description = types.StringValue(tag.GetDescription())

	if len(tag.GetScopeReferences()) > 0 {
		remoteScopeIds := []string{}
		for _, scope := range tag.GetScopeReferences() {
			remoteScopeIds = append(remoteScopeIds, scope.GetScopeId())
		}
		r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, remoteScopeIds)
	} else {
		r.Scopes = types.SetNull(types.StringType)
	}

	// Refresh computed attributes
	r.AssociatedMachineCount = types.Int32Value(tag.GetNumMachines())
	r.AssociatedApplicationCount = types.Int32Value(tag.GetNumApplications())
	r.AssociatedApplicationGroupCount = types.Int32Value(tag.GetNumApplicationGroups())
	r.AssociatedMachineCatalogCount = types.Int32Value(tag.GetNumMachineCatalogs())
	r.AssociatedDeliveryGroupCount = types.Int32Value(tag.GetNumDeliveryGroups())

	return r
}
