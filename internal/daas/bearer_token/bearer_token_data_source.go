// Copyright © 2024. Citrix Systems, Inc.

package bearer_token

import (
	"context"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &BearerTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &BearerTokenDataSource{}
)

func NewBearerTokenDataSource() datasource.DataSource {
	return &BearerTokenDataSource{}
}

type BearerTokenDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *BearerTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bearer_token"
}

func (d *BearerTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = BearerTokenDataSourceModel{}.GetSchema()
}

func (d *BearerTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *BearerTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data BearerTokenDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cwsAuthToken, httpResp, err := d.client.SignIn()
	var token string
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Bearer Token",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+err.Error(),
		)
		return
	}

	if cwsAuthToken == "" {
		resp.Diagnostics.AddError(
			"Error fetching Bearer Token",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Bearer token is empty.",
		)
		return
	}
	token = strings.Split(cwsAuthToken, "=")[1]

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, token)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
