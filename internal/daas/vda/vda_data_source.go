// Copyright © 2024. Citrix Systems, Inc.

package vda

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &VdaDataSource{}
	_ datasource.DataSourceWithConfigure = &VdaDataSource{}
)

func NewVdaDataSource() datasource.DataSource {
	return &VdaDataSource{}
}

// VdaDataSource defines the data source implementation.
type VdaDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *VdaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vda"
}

func (d *VdaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VdaDataSourceModel{}.GetSchema()
}

func (d *VdaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *VdaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data VdaDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalog state from Orchestration
	machineCatalogId := data.MachineCatalog.ValueString()
	if machineCatalogId != "" {
		machineCatalogVdas, err := util.GetMachineCatalogMachines(ctx, d.client, &resp.Diagnostics, machineCatalogId)

		if err != nil {
			return
		}

		data = data.RefreshPropertyValues(machineCatalogVdas)
	}

	deliveryGroupId := data.DeliveryGroup.ValueString()
	if deliveryGroupId != "" {
		deliveryGroupVdas, err := util.GetDeliveryGroupMachines(ctx, d.client, &resp.Diagnostics, deliveryGroupId)
		if err != nil {
			return
		}

		data = data.RefreshPropertyValues(deliveryGroupVdas)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
