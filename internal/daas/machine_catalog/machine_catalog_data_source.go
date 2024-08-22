// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &MachineCatalogDataSource{}
)

func NewMachineCatalogDataSource() datasource.DataSource {
	return &MachineCatalogDataSource{}
}

// MachineCatalogDataSource defines the data source implementation.
type MachineCatalogDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *MachineCatalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_catalog"
}

func (d *MachineCatalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MachineCatalogDataSourceModel{}.GetSchema()
}

func (d *MachineCatalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *MachineCatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data MachineCatalogDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalog state from Orchestration
	machineCatalogPath := strings.ReplaceAll(data.MachineCatalogFolderPath.ValueString(), "\\", "|") + data.Name.ValueString()
	getMachineCatalogRequest := d.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogPath).Fields("Id,Name,Description,ProvisioningType,Zone,AllocationType,SessionSupport,TotalCount,HypervisorConnection,ProvisioningScheme,RemotePCEnrollmentScopes,IsPowerManaged,MinimumFunctionalLevel,IsRemotePC")
	machineCatalog, httpResp, err := citrixdaasclient.AddRequestData(getMachineCatalogRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Machine Catalog "+data.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Get VDAs associated with the machine catalog
	getMachineCatalogMachinesRequest := d.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalogMachines(ctx, machineCatalogPath)
	machineCatalogVdas, httpResp, err := citrixdaasclient.AddRequestData(getMachineCatalogMachinesRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VDAs in Machine Catalog "+data.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(machineCatalog, machineCatalogVdas)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
