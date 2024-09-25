// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeliveryGroupDataSourceModel defines the Delivery Group data source implementation.
type DeliveryGroupDataSourceModel struct {
	Id                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	DeliveryGroupFolderPath types.String   `tfsdk:"delivery_group_folder_path"`
	Vdas                    []vda.VdaModel `tfsdk:"vdas"`    // List[VdaModel]
	Tenants                 types.Set      `tfsdk:"tenants"` // Set[string]
	Tags                    types.Set      `tfsdk:"tags"`    // Set[string]
}

func (DeliveryGroupDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing delivery group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the delivery group.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the delivery group.",
				Required:    true,
			},
			"delivery_group_folder_path": schema.StringAttribute{
				Description: "The path to the folder in which the delivery group is located.",
				Optional:    true,
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the delivery group.",
				Computed:     true,
				NestedObject: vda.VdaModel{}.GetSchema(),
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the delivery group.",
				Computed:    true,
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the delivery group.",
				Computed:    true,
			},
		},
	}
}

func (r DeliveryGroupDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, vdas *citrixorchestration.MachineResponseModelCollection, tags []string) DeliveryGroupDataSourceModel {
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())

	adminFolder := deliveryGroup.GetAdminFolder()
	adminFolderPath := adminFolder.GetName()
	if adminFolderPath != "" {
		r.DeliveryGroupFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.DeliveryGroupFolderPath = types.StringNull()
	}

	res := []vda.VdaModel{}
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

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, deliveryGroup.GetTenants())
	r.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

	return r
}
