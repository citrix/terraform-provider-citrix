// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"

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

	var machineCatalogPathOrId string
	var machineCatalogNameOrId string
	if !data.Id.IsNull() {
		machineCatalogPathOrId = data.Id.ValueString()
		machineCatalogNameOrId = data.Id.ValueString()
	} else {
		// Get refreshed machine catalog state from Orchestration
		machineCatalogNameOrId = data.Name.ValueString()
		machineCatalogPathOrId = util.BuildResourcePathForGetRequest(data.MachineCatalogFolderPath.ValueString(), data.Name.ValueString())
	}
	getMachineCatalogRequest := d.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogPathOrId).Fields("Id,Name,Description,ProvisioningType,PersistChanges,Zone,AllocationType,SessionSupport,TotalCount,HypervisorConnection,ProvisioningScheme,RemotePCEnrollmentScopes,IsPowerManaged,MinimumFunctionalLevel,IsRemotePC")
	machineCatalog, httpResp, err := citrixdaasclient.AddRequestData(getMachineCatalogRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Machine Catalog "+machineCatalogNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	machineCatalogId := machineCatalog.GetId()

	// Get VDAs associated with the machine catalog
	machineCatalogVdas, err := util.GetMachineCatalogMachines(ctx, d.client, &resp.Diagnostics, machineCatalogId)
	if err != nil {
		return
	}

	tags := getMachineCatalogTags(ctx, &resp.Diagnostics, d.client, machineCatalogId)

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, machineCatalog, machineCatalogVdas, tags)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
