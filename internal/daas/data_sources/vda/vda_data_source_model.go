// Copyright © 2023. Citrix Systems, Inc.

package vda

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VdaDataSourceModel defines the VDA data source implementation.
type VdaDataSourceModel struct {
	MachineCatalog types.String `tfsdk:"machine_catalog"`
	DeliveryGroup  types.String `tfsdk:"delivery_group"`
	Vdas           []VdaModel   `tfsdk:"vdas"`
}

// VdaModel defines the single VDA data model implementation.
type VdaModel struct {
	MachineName              types.String `tfsdk:"machine_name"`
	HostedMachineId          types.String `tfsdk:"hosted_machine_id"`
	AssociatedMachineCatalog types.String `tfsdk:"associated_machine_catalog"`
	AssociatedDeliveryGroup  types.String `tfsdk:"associated_delivery_group"`
}

func (r VdaDataSourceModel) RefreshPropertyValues(vdas *citrixorchestration.MachineResponseModelCollection) VdaDataSourceModel {

	var res []VdaModel
	for _, model := range vdas.GetItems() {
		machineName := model.GetName()
		hosting := model.GetHosting()
		hostedMachineId := hosting.GetHostedMachineId()
		machineCatalog := model.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		deliveryGroup := model.GetDeliveryGroup()
		deliveryGroupId := deliveryGroup.GetId()

		res = append(res, VdaModel{
			MachineName:              types.StringValue(machineName),
			HostedMachineId:          types.StringValue(hostedMachineId),
			AssociatedMachineCatalog: types.StringValue(machineCatalogId),
			AssociatedDeliveryGroup:  types.StringValue(deliveryGroupId),
		})
	}

	r.Vdas = res

	return r
}
