// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &ServiceAccountDataSource{}
)

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

type ServiceAccountDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ServiceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ServiceAccountDataSourceModel{}.GetDataSourceSchema()
}

func (d *ServiceAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data ServiceAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the service account using the account ID
	serviceAccount, err := GetServiceAccountUsingAccountId(ctx, d.client, &resp.Diagnostics, data.AccountId.ValueString())
	if err != nil {
		return
	}
	data = data.RefreshPropertyValues(ctx, *serviceAccount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
