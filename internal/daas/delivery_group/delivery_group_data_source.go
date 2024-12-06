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

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data DeliveryGroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed delivery group state from Orchestration
	var deliveryGroupPathOrId string
	var deliveryGroupNameOrId string
	if !data.Id.IsNull() {
		deliveryGroupPathOrId = data.Id.ValueString()
		deliveryGroupNameOrId = data.Id.ValueString()
	} else {
		deliveryGroupNameOrId = data.Name.ValueString()
		deliveryGroupPathOrId = util.BuildResourcePathForGetRequest(data.DeliveryGroupFolderPath.ValueString(), deliveryGroupNameOrId)
	}

	getDeliveryGroupRequest := d.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupPathOrId)
	deliveryGroup, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Delivery Group "+deliveryGroupNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Get VDAs associated with the delivery group
	deliveryGroupVdas, err := util.GetDeliveryGroupMachines(ctx, d.client, &resp.Diagnostics, deliveryGroup.GetId())

	if err != nil {
		return
	}

	tags := getDeliveryGroupTags(ctx, &resp.Diagnostics, d.client, deliveryGroupPathOrId)

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, deliveryGroup, deliveryGroupVdas, tags)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
