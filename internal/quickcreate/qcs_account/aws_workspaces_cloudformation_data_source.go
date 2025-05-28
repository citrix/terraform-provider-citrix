// Copyright © 2024. Citrix Systems, Inc.
package qcs_account

import (
	"context"
	"fmt"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &AwsWorkspacesCloudFormationDataSource{}
	_ datasource.DataSourceWithConfigure = &AwsWorkspacesCloudFormationDataSource{}
)

func NewAwsWorkspacesCloudFormationDataSource() datasource.DataSource {
	return &AwsWorkspacesCloudFormationDataSource{}
}

type AwsWorkspacesCloudFormationDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the datasource type name.
func (r *AwsWorkspacesCloudFormationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_cloudformation_template"
}

// Schema defines the schema for the datasource.
func (r *AwsWorkspacesCloudFormationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AwsWorkspacesCloudFormationDataSourceModel{}.GetSchema()
}

func (r *AwsWorkspacesCloudFormationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Read refreshes the Terraform state with the latest data.
func (r *AwsWorkspacesCloudFormationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AwsWorkspacesCloudFormationDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cloudFormationFile, err := getAwsWorkspacesCloudFormationTemplate(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, cloudFormationFile)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAwsWorkspacesCloudFormationTemplate(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) (*citrixquickcreate.AwsEdcAccountResourceFile, error) {
	getTemplateRequestBody := citrixquickcreate.SearchAwsEdcAccountResourceRequest{}
	getTemplateRequestBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	getTemplateRequestBody.SetResourceType(citrixquickcreate.AWSACCOUNTRESOURCETYPE_AWS_CLOUDFORMATION)
	getTemplateRequestBody.SetFilterProperty("CreateType")
	getTemplateRequestBody.SetFilterValue("AccountRoleCreation")

	getTemplateRequest := client.QuickCreateClient.AccountQCS.GetCustomerAccountResourcesAsync(ctx, client.ClientConfig.CustomerId)
	getTemplateRequest = getTemplateRequest.Body(getTemplateRequestBody)
	getTemplateRequest.Execute()
	accountResource, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AccountResources](getTemplateRequest, client)
	if err != nil {
		return nil, err
	}
	resources := accountResource.GetItems()
	if len(resources) == 0 {
		err = fmt.Errorf("AWS WorkSpaces CloudFormation template not exists.")
		diagnostics.AddError(
			"Error getting AWS WorkSpaces CloudFormation template",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+err.Error(),
		)
		return nil, err
	}
	resource := resources[0]
	fileResource := resource.AwsEdcAccountResourceFile
	if fileResource != nil && fileResource.GetResourceType() == citrixquickcreate.AWSACCOUNTRESOURCETYPE_AWS_CLOUDFORMATION {
		return fileResource, nil
	}

	err = fmt.Errorf("Error getting AWS WorkSpaces CloudFormation template")
	diagnostics.AddError(
		"Error getting AWS WorkSpaces CloudFormation template",
		"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp),
	)
	return nil, err
}
