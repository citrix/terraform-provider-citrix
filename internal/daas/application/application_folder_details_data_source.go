// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ApplicationDataSource{}
	_ datasource.DataSourceWithConfigure = &ApplicationDataSource{}
)

func NewApplicationDataSourceSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource defines the data source implementation.
type ApplicationDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_folder_details"
}

// Schema defines the data source schema.
func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ApplicationFolderDetailsDataSourceModel{}.GetDataSourceSchema()
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data ApplicationFolderDetailsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the list of applications using the path
	if !data.Path.IsNull() {
		applicationFolderPath := util.BuildResourcePathForGetRequest(data.Path.ValueString(), "")
		getApplicationsRequest := d.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolderApplications(ctx, applicationFolderPath)
		apps, httpResp, err := citrixdaasclient.AddRequestData(getApplicationsRequest, d.client).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting Applications from folder "+data.Path.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return // Stop processing
		}

		data = data.RefreshPropertyValues(ctx, d.client, &resp.Diagnostics, apps)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
