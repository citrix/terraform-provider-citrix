// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &HypervisorResourcePoolDataSource{}
)

func NewHypervisorResourcePoolDataSource() datasource.DataSource {
	return &HypervisorResourcePoolDataSource{}
}

// HypervisorResourcePoolDataSource defines the data source implementation.
type HypervisorResourcePoolDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *HypervisorResourcePoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hypervisor_resource_pool"
}

func (d *HypervisorResourcePoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = HypervisorResourcePoolDataSourceModel{}.GetSchema()
}

func (d *HypervisorResourcePoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *HypervisorResourcePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data HypervisorResourcePoolDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor resource pool state from Orchestration
	hypervisorName := data.HypervisorName.ValueString()
	resourcePoolName := data.Name.ValueString()
	getHypervisorResourcePoolRequest := d.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePool(ctx, hypervisorName, resourcePoolName)
	resourcePool, httpResp, err := citrixdaasclient.AddRequestData(getHypervisorResourcePoolRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Resource Pool "+resourcePoolName+" of Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, resourcePool)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
