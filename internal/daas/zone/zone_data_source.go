// Copyright © 2024. Citrix Systems, Inc.

package zone

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &ZoneDataSource{}
)

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

// ZoneDataSource defines the data source implementation.
type ZoneDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ZoneDataSourceModel{}.GetSchema()
}

func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data ZoneDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed zone state from Orchestration
	zoneName := data.Name.ValueString()
	getZoneRequest := d.client.ApiClient.ZonesAPIsDAAS.ZonesGetZone(ctx, zoneName)
	zone, httpResp, err := citrixdaasclient.AddRequestData(getZoneRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Zone "+zoneName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(zone)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
