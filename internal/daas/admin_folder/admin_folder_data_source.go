// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &AdminFolderDataSource{}
)

func NewAdminFolderDataSource() datasource.DataSource {
	return &AdminFolderDataSource{}
}

type AdminFolderDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AdminFolderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_folder"
}

func (d *AdminFolderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AdminFolderModel{}.GetDataSourceSchema()
}

func (d *AdminFolderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AdminFolderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AdminFolderModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var adminFolderIdOrPath string

	if !data.Id.IsNull() {
		adminFolderIdOrPath = data.Id.ValueString()
	} else {
		adminFolderIdOrPath = util.BuildResourcePathForGetRequest(data.Path.ValueString(), "")
	}

	adminFolder, err := getAdminFolder(ctx, d.client, &resp.Diagnostics, adminFolderIdOrPath)

	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, adminFolder)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
