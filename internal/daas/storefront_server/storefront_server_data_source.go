// Copyright © 2024. Citrix Systems, Inc.

package storefront_server

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &StoreFrontServerDataSource{}
)

func NewStoreFrontServerDataSource() datasource.DataSource {
	return &StoreFrontServerDataSource{}
}

type StoreFrontServerDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *StoreFrontServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storefront_server"
}

func (d *StoreFrontServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = StoreFrontServerResourceModel{}.GetDataSourceSchema()
}

func (d *StoreFrontServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *StoreFrontServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data StoreFrontServerResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var storeFrontServerNameOrId string

	if data.Id.ValueString() != "" {
		storeFrontServerNameOrId = data.Id.ValueString()
	} else if data.Name.ValueString() != "" {
		storeFrontServerNameOrId = data.Name.ValueString()
	}

	// Try getting the new StoreFront server with StoreFront server name
	storeFrontServer, _, err := getStoreFrontServer(ctx, d.client, storeFrontServerNameOrId)
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(storeFrontServer)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
