// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &DeliveryGroupDataSource{}
)

func NewDeliveryGroupDataSource() datasource.DataSource {
	return &DeliveryGroupDataSource{}
}

// DeliveryGroupDataSource defines the data source implementation.
type DeliveryGroupDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *DeliveryGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_group"
}

func (d *DeliveryGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DeliveryGroupDataSourceModel{}.GetSchema()
}

func (d *DeliveryGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *DeliveryGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data DeliveryGroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed delivery group state from Orchestration
	deliveryGroupName := data.Name.ValueString()
	getDeliveryGroupRequest := d.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupName)
	deliveryGroup, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Get VDAs associated with the delivery group
	getDeliveryGroupMachinesRequest := d.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupMachines(ctx, deliveryGroupName)
	deliveryGroupVdas, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupMachinesRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VDAs in Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(deliveryGroup, deliveryGroupVdas)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
