// Copyright © 2024. Citrix Systems, Inc.

package qcs_account

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &awsWorkspacesAccountDataSource{}
	_ datasource.DataSourceWithConfigure = &awsWorkspacesAccountDataSource{}
)

func NewAccountDataSource() datasource.DataSource {
	return &awsWorkspacesAccountDataSource{}
}

type awsWorkspacesAccountDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the datasource type name.
func (r *awsWorkspacesAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_account"
}

// Schema defines the schema for the datasource.
func (r *awsWorkspacesAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AwsWorkspacesAccountDataSourceModel{}.GetSchema()
}

func (r *awsWorkspacesAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Read refreshes the Terraform state with the latest data.
func (r *awsWorkspacesAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AwsWorkspacesAccountDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var account *citrixquickcreate.AwsEdcAccount
	var err error
	if data.AccountId.ValueString() != "" {
		account, _, err = getAwsWorkspacesAccountUsingId(ctx, r.client, &resp.Diagnostics, data.AccountId.ValueString())
	} else if data.Name.ValueString() != "" {
		account, _, err = getAwsWorkspacesAccountUsingName(ctx, r.client, &resp.Diagnostics, data.Name.ValueString())
	}
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(account)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAwsWorkspacesAccountUsingName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountName string) (*citrixquickcreate.AwsEdcAccount, *http.Response, error) {
	getAccountsRequest := client.QuickCreateClient.AccountQCS.GetCustomerAccountsAsync(ctx, client.ClientConfig.CustomerId)
	accounts, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.Accounts](getAccountsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Account: "+accountName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	for _, account := range accounts.GetItems() {
		if strings.EqualFold(account.GetName(), accountName) {
			return &account, httpResp, nil
		}
	}

	return nil, httpResp, fmt.Errorf("AWS WorkSpaces Account not found: %s", accountName)
}
