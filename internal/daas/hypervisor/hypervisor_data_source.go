// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &HypervisorDataSource{}
)

func NewHypervisorDataSource() datasource.DataSource {
	return &HypervisorDataSource{}
}

// HypervisorDataSource defines the data source implementation.
type HypervisorDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *HypervisorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hypervisor"
}

func (d *HypervisorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = HypervisorDataSourceModel{}.GetSchema()
}

func (d *HypervisorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *HypervisorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data HypervisorDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor state from Orchestration
	hypervisorName := data.Name.ValueString()
	getHypervisorRequest := d.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorName)
	hypervisor, httpResp, err := citrixdaasclient.AddRequestData(getHypervisorRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, hypervisor)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
