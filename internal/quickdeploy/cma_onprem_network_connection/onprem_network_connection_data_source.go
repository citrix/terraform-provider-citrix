// Copyright © 2024. Citrix Systems, Inc.
package cma_vnet_peering

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &CitrixManagedAzureOnPremNetworkConnectionDataSource{}
	_ datasource.DataSourceWithConfigure = &CitrixManagedAzureOnPremNetworkConnectionDataSource{}
)

func NewCitrixManagedAzureOnPremNetworkConnectionDataSource() datasource.DataSource {
	return &CitrixManagedAzureOnPremNetworkConnectionDataSource{}
}

// CitrixManagedAzureOnPremNetworkConnectionDataSource is the data source implementation.
type CitrixManagedAzureOnPremNetworkConnectionDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *CitrixManagedAzureOnPremNetworkConnectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickdeploy_onprem_network_connection"
}

func (d *CitrixManagedAzureOnPremNetworkConnectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CitrixManagedAzureOnPremNetworkConnectionDataSourceModel{}.GetDataSourceSchema()
}

func (d *CitrixManagedAzureOnPremNetworkConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *CitrixManagedAzureOnPremNetworkConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data CitrixManagedAzureOnPremNetworkConnectionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get VNet Peering Id from Name
	onPremConnection, err := util.GetCitrixManagedOnPremConnectionWithName(ctx, d.client, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting Citrix Managed Azure VNet Peering", err.Error())
		return
	}
	if onPremConnection == nil {
		resp.Diagnostics.AddError("Error getting Citrix Managed Azure VNet Peering", "VNet Peering with name "+data.Name.ValueString()+" not found")
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, onPremConnection)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
