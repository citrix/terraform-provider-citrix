// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeliveryGroupDataSourceModel defines the Delivery Group data source implementation.
type DeliveryGroupDataSourceModel struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Vdas []vda.VdaModel `tfsdk:"vdas"` // List[VdaModel]
}

func (DeliveryGroupDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Read data of an existing delivery group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the delivery group.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the delivery group.",
				Required:    true,
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the delivery group.",
				Computed:     true,
				NestedObject: vda.VdaModel{}.GetSchema(),
			},
		},
	}
}

func (r DeliveryGroupDataSourceModel) RefreshPropertyValues(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, vdas *citrixorchestration.MachineResponseModelCollection) DeliveryGroupDataSourceModel {
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())

	var res []vda.VdaModel
	for _, model := range vdas.GetItems() {
		machineName := model.GetName()
		hosting := model.GetHosting()
		hostedMachineId := hosting.GetHostedMachineId()
		machineCatalog := model.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		deliveryGroup := model.GetDeliveryGroup()
		deliveryGroupId := deliveryGroup.GetId()

		res = append(res, vda.VdaModel{
			MachineName:              types.StringValue(machineName),
			HostedMachineId:          types.StringValue(hostedMachineId),
			AssociatedMachineCatalog: types.StringValue(machineCatalogId),
			AssociatedDeliveryGroup:  types.StringValue(deliveryGroupId),
		})
	}

	r.Vdas = res

	return r
}
