// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"
	"strings"

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

	accountId := data.AccountId.ValueString()
	getServiceAccounts := d.client.ApiClient.IdentityAPIsDAAS.IdentityGetServiceAccounts(ctx)
	serviceAccounts, httpResp, err := citrixdaasclient.AddRequestData(getServiceAccounts, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Service Accounts",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	for _, serviceAccount := range serviceAccounts.GetItems() {
		if strings.EqualFold(serviceAccount.GetAccountId(), accountId) {
			data = data.RefreshPropertyValues(ctx, serviceAccount)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Error reading Service Account "+accountId,
		"Service Account not found.",
	)

}
