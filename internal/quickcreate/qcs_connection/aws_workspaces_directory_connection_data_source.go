// Copyright © 2024. Citrix Systems, Inc.
package qcs_connection

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

var (
	_ datasource.DataSource = &AwsWorkspacesDirectoryConnectionDataSource{}
)

func NewAwsWorkspacesDirectoryConnectionDataSource() datasource.DataSource {
	return &AwsWorkspacesDirectoryConnectionDataSource{}
}

// AwsWorkspacesDirectoryConnectionDataSource is the data source implementation.
type AwsWorkspacesDirectoryConnectionDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AwsWorkspacesDirectoryConnectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_directory_connection"
}

func (d *AwsWorkspacesDirectoryConnectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AwsWorkspacesDirectoryConnectionDataSourceModel{}.GetSchema()
}

func (d *AwsWorkspacesDirectoryConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AwsWorkspacesDirectoryConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AwsWorkspacesDirectoryConnectionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Image
	var directoryConnection *citrixquickcreate.AwsEdcDirectoryConnection
	var directoryConnectionIdentifier string
	var err error
	if data.DirectoryConnectionId.ValueString() != "" {
		directoryConnectionIdentifier = data.DirectoryConnectionId.ValueString()
		directoryConnection, _, err = getAwsWorkspacesDirectoryConnection(ctx, d.client, &resp.Diagnostics, data.AccountId.ValueString(), data.DirectoryConnectionId.ValueString(), false)
	} else if data.Name.ValueString() != "" {
		directoryConnectionIdentifier = data.Name.ValueString()
		directoryConnection, _, err = getAwsWorkspacesDirectoryConnectionWithName(ctx, d.client, &resp.Diagnostics, data.AccountId.ValueString(), data.Name.ValueString())
	}

	if err != nil {
		return
	}

	if directoryConnection == nil {
		resp.Diagnostics.AddError(
			"Error getting AWS WorkSpaces Directory Connection",
			fmt.Sprintf("Unable to read AWS WorkSpaces Directory Connection: %s", directoryConnectionIdentifier),
		)
		return
	}
	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, directoryConnection)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAwsWorkspacesDirectoryConnectionWithName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, directoryConnectionName string) (*citrixquickcreate.AwsEdcDirectoryConnection, *http.Response, error) {
	getDirectoryConnectionRequest := client.QuickCreateClient.ConnectionQCS.GetResourceConnectionsAsync(ctx, client.ClientConfig.CustomerId, accountId)
	directoryConnections, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.ResourceConnections](getDirectoryConnectionRequest, client)

	if err != nil {
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Directory Connection: "+directoryConnectionName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	awsDirectoryConnections, ok := directoryConnections.GetItemsOk()

	if !ok {
		diagnostics.AddError(
			"Error getting directory connection with name: "+directoryConnectionName,
			"No directory connection found with account: "+accountId,
		)
		return nil, httpResp, fmt.Errorf("no directory connection found with account: %s", accountId)
	}

	for _, directoryConnection := range awsDirectoryConnections {
		if strings.EqualFold(directoryConnectionName, directoryConnection.GetName()) {
			return &directoryConnection, httpResp, nil
		}
	}

	diagnostics.AddError(
		"Error getting directory connection with name: "+directoryConnectionName,
		fmt.Sprintf("No directory connection with name %s found with account %s", directoryConnectionName, accountId),
	)
	return nil, httpResp, fmt.Errorf("no directory connection with name %s found with account %s", directoryConnectionName, accountId)
}
