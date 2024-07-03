// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MachineCatalogDataSourceModel defines the Machine Catalog data source implementation.
type MachineCatalogDataSourceModel struct {
	Id   types.String   `tfsdk:"id"`
	Name types.String   `tfsdk:"name"`
	Vdas []vda.VdaModel `tfsdk:"vdas"` // List[VdaModel]
}

func (MachineCatalogDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Read data of an existing machine catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the machine catalog.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine catalog.",
				Required:    true,
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the machine catalog.",
				Computed:     true,
				NestedObject: vda.VdaModel{}.GetSchema(),
			},
		},
	}
}

func (r MachineCatalogDataSourceModel) RefreshPropertyValues(catalog *citrixorchestration.MachineCatalogDetailResponseModel, vdas *citrixorchestration.MachineResponseModelCollection) MachineCatalogDataSourceModel {
	r.Id = types.StringValue(catalog.GetId())
	r.Name = types.StringValue(catalog.GetName())

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
