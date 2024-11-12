// Copyright © 2024. Citrix Systems, Inc.
package qcs_deployment

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
	_ datasource.DataSource              = &awsWorkspacesDeploymentDataSource{}
	_ datasource.DataSourceWithConfigure = &awsWorkspacesDeploymentDataSource{}
)

func NewAwsWorkspacesDeploymentDataSource() datasource.DataSource {
	return &awsWorkspacesDeploymentDataSource{}
}

// awsWorkspacesDeploymentDataSource is the datasource implementation.
type awsWorkspacesDeploymentDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the datasource type name.
func (r *awsWorkspacesDeploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_deployment"
}

// Schema defines the schema for the datasource.
func (r *awsWorkspacesDeploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AwsWorkspacesDeploymentDataSourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the datasource.
func (r *awsWorkspacesDeploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Read refreshes the Terraform state with the latest data.
func (r *awsWorkspacesDeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Retrieve values from state
	var data AwsWorkspacesDeploymentDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deployment *citrixquickcreate.AwsEdcDeployment
	var err error
	// Try getting the AWS WorkSpaces Deployment
	if !data.Id.IsNull() {
		deployment, _, err = getAwsWorkspacesDeploymentUsingId(ctx, r.client, &resp.Diagnostics, data.Id.ValueString(), true)
	} else {
		deployment, _, err = getAwsWorkspacesDeploymentByName(ctx, r.client, &resp.Diagnostics, data.Name.ValueString(), data.AccountId.ValueString())
	}
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, *deployment)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAwsWorkspacesDeploymentByName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deploymentName string, accountId string) (*citrixquickcreate.AwsEdcDeployment, *http.Response, error) {
	getDeploymentsRequest := client.QuickCreateClient.DeploymentQCS.GetDeploymentsAsync(ctx, client.ClientConfig.CustomerId)
	deployments, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.Deployments](getDeploymentsRequest, client)

	if err != nil {
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Deployment: "+deploymentName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	for _, deployment := range deployments.GetItems() {
		if strings.EqualFold(deployment.GetDeploymentName(), deploymentName) && strings.EqualFold(deployment.GetAccountId(), accountId) {
			return &deployment, httpResp, nil
		}
	}

	err = fmt.Errorf("AWS WorkSpaces Deployment not found: " + deploymentName)
	diagnostics.AddError(
		"Error getting AWS WorkSpaces Deployment",
		util.ReadQcsClientError(err),
	)

	return nil, httpResp, err
}
