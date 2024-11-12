// Copyright © 2024. Citrix Systems, Inc.

package vda

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VdaDataSourceModel defines the VDA data source implementation.
type VdaDataSourceModel struct {
	MachineCatalog types.String `tfsdk:"machine_catalog"`
	DeliveryGroup  types.String `tfsdk:"delivery_group"`
	Vdas           []VdaModel   `tfsdk:"vdas"`
}

func (VdaDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CVAD --- Data source for the list of VDAs that belong to either a machine catalog or a delivery group. Machine catalog and delivery group cannot be specified at the same time.",

		Attributes: map[string]schema.Attribute{
			"machine_catalog": schema.StringAttribute{
				MarkdownDescription: "The machine catalog which the VDAs are associated with.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("machine_catalog"), path.MatchRoot("delivery_group")), // Ensures that only one of either machine_catalog or delivery_group is provided. It will also cause a validation error if none are specified.,
				},
			},
			"delivery_group": schema.StringAttribute{
				MarkdownDescription: "The delivery group which the VDAs are associated with.",
				Optional:            true,
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the specified machine catalog or delivery group.",
				Computed:     true,
				NestedObject: VdaModel{}.GetSchema(),
			},
		},
	}
}

// VdaModel defines the single VDA data model implementation.
type VdaModel struct {
	Id                       types.String `tfsdk:"id"`
	MachineName              types.String `tfsdk:"machine_name"`
	HostedMachineId          types.String `tfsdk:"hosted_machine_id"`
	AssociatedMachineCatalog types.String `tfsdk:"associated_machine_catalog"`
	AssociatedDeliveryGroup  types.String `tfsdk:"associated_delivery_group"`
}

func (VdaModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the VDA.",
				Computed:    true,
			},
			"machine_name": schema.StringAttribute{
				Description: "Machine name of the VDA.",
				Computed:    true,
			},
			"hosted_machine_id": schema.StringAttribute{
				Description: "Machine ID within the hypervisor hosting unit.",
				Computed:    true,
			},
			"associated_machine_catalog": schema.StringAttribute{
				Description: "Machine catalog which the VDA is associated with.",
				Computed:    true,
			},
			"associated_delivery_group": schema.StringAttribute{
				Description: "Delivery group which the VDA is associated with.",
				Computed:    true,
			},
		},
	}
}

func (r VdaDataSourceModel) RefreshPropertyValues(vdas *citrixorchestration.MachineResponseModelCollection) VdaDataSourceModel {

	res := []VdaModel{}
	for _, model := range vdas.GetItems() {
		machineName := model.GetName()
		hosting := model.GetHosting()
		hostedMachineId := hosting.GetHostedMachineId()
		machineCatalog := model.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		deliveryGroup := model.GetDeliveryGroup()
		deliveryGroupId := deliveryGroup.GetId()

		res = append(res, VdaModel{
			Id:                       types.StringValue(model.GetId()),
			MachineName:              types.StringValue(machineName),
			HostedMachineId:          types.StringValue(hostedMachineId),
			AssociatedMachineCatalog: types.StringValue(machineCatalogId),
			AssociatedDeliveryGroup:  types.StringValue(deliveryGroupId),
		})
	}

	r.Vdas = res

	return r
}
