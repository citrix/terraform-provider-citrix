// Copyright © 2024. Citrix Systems, Inc.

package qcs_account

import (
	"context"
	"net/http"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &awsWorkspacesAccountResource{}
	_ resource.ResourceWithConfigure      = &awsWorkspacesAccountResource{}
	_ resource.ResourceWithImportState    = &awsWorkspacesAccountResource{}
	_ resource.ResourceWithValidateConfig = &awsWorkspacesAccountResource{}
	_ resource.ResourceWithModifyPlan     = &awsWorkspacesAccountResource{}
)

func NewAwsWorkspacesAccountResource() resource.Resource {
	return &awsWorkspacesAccountResource{}
}

// awsWorkspaceAccountResource is the resource implementation.
type awsWorkspacesAccountResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *awsWorkspacesAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_account"
}

// Schema defines the schema for the resource.
func (r *awsWorkspacesAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AwsWorkspacesAccountResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *awsWorkspacesAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *awsWorkspacesAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Account API request body from plan
	var accountDetails citrixquickcreate.AddAwsEdcAccount
	accountDetails.AccountType = citrixquickcreate.ACCOUNTTYPE_AWSEDC
	accountDetails.Name = plan.Name.ValueString()
	accountDetails.AwsRegion = plan.AwsRegion.ValueStringPointer()
	// Always set AWS External ID to Customer ID
	accountDetails.SetAwsExternalId(r.client.ClientConfig.CustomerId)
	// Validate that plan has either AWS Access Key ID and Secret Access Key or Role ARN
	if !plan.AwsAccessKeyId.IsNull() && !plan.AwsSecretAccessKey.IsNull() {
		accountDetails.SetAwsAccessKeyId(plan.AwsAccessKeyId.ValueString())
		accountDetails.AwsSecretAccessKey = *citrixquickcreate.NewNullableString(plan.AwsSecretAccessKey.ValueStringPointer())
	} else if !plan.AwsRoleArn.IsNull() {
		accountDetails.AwsRoleArn = *citrixquickcreate.NewNullableString(plan.AwsRoleArn.ValueStringPointer())
	} else {
		// Return error if both AWS Access Key ID and Secret Access Key are empty
		resp.Diagnostics.AddError("Error adding AWS WorkSpaces Account: "+plan.Name.ValueString(), "Error message: You must provide either AWS Access Key ID and Secret Access Key or Role ARN")
		return
	}

	// Generate API request body from plan
	createAccountRequest := r.client.QuickCreateClient.AccountQCS.AddAccountAsync(ctx, r.client.ClientConfig.CustomerId)
	createAccountRequest = createAccountRequest.Body(accountDetails)

	// Create new AWS WorkSpaces Account
	addAccountResponse, httpResp, err := citrixdaasclient.AddRequestData(createAccountRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding AWS WorkSpaces Account: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}

	// Try getting the new AWS WorkSpaces Account
	account, _, err := getAwsWorkspacesAccountUsingId(ctx, r.client, &resp.Diagnostics, *addAccountResponse.AccountId.Get())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(account)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *awsWorkspacesAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesAccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Account
	account, err := readAwsWorkspacesAccountUsingId(ctx, r.client, resp, state.AccountId.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(account)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *awsWorkspacesAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state AwsWorkspacesAccountResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Two possible options for WorkSpaces accounts
	// 1. Update account name
	// 2. Update account credentials

	// Get refreshed account properties from QCS
	accountId := plan.AccountId.ValueString()
	account, httpResp, err := getAwsWorkspacesAccountUsingId(ctx, r.client, &resp.Diagnostics, accountId)
	if err != nil {
		return
	}
	if account.AccountType != citrixquickcreate.ACCOUNTTYPE_AWSEDC {
		resp.Diagnostics.AddError(
			"Error updating AWS WorkSpaces Account: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Account is not an AWS WorkSpaces account",
		)
		return
	}

	// Check if the account name is being updated
	if plan.Name.ValueString() != *account.Name.Get() {
		// Update account name
		updateAccountNameRequestBody := citrixquickcreate.UpdateAccountName{}
		updateAccountNameRequestBody.SetName(plan.Name.ValueString())
		updateAccountNameRequestBody.SetAccountOperationType(citrixquickcreate.UPDATEACCOUNTOPERATIONTYPE_RENAME_ACCOUNT)

		httpResp, err := updateAwsWorkspacesAccount(ctx, r.client, &resp.Diagnostics, accountId, citrixquickcreate.UpdateCustomerAccountAsyncRequest{UpdateAccountName: &updateAccountNameRequestBody})

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating AWS WorkSpaces Account Name: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return
		}
	}

	// Throw an error if only Access Key ID or only Secret Access Key is changed
	if (plan.AwsAccessKeyId.ValueString() != state.AwsAccessKeyId.ValueString() && plan.AwsSecretAccessKey.ValueString() == state.AwsSecretAccessKey.ValueString()) ||
		(plan.AwsAccessKeyId.ValueString() == state.AwsAccessKeyId.ValueString() && plan.AwsSecretAccessKey.ValueString() != state.AwsSecretAccessKey.ValueString()) {
		resp.Diagnostics.AddError(
			"Error updating AWS WorkSpaces Account Credentials: "+plan.Name.ValueString(),
			"Error message: You must update both AWS Access Key ID and Secret Access Key",
		)
		return
	}

	// Check if the account credentials are being updated
	if plan.AwsRoleArn.ValueString() != state.AwsRoleArn.ValueString() ||
		(plan.AwsAccessKeyId.ValueString() != state.AwsAccessKeyId.ValueString() &&
			plan.AwsSecretAccessKey.ValueString() != state.AwsSecretAccessKey.ValueString()) {
		// Update account credentials
		updateAccountCredentialsRequestBody := citrixquickcreate.UpdateAwsEdcAccountCredentials{}
		updateAccountCredentialsRequestBody.SetAccountOperationType(citrixquickcreate.UPDATEACCOUNTOPERATIONTYPE_UPDATE_AWS_EDC_ACCOUNT_CREDENTIALS)
		if !plan.AwsRoleArn.IsNull() {
			updateAccountCredentialsRequestBody.SetAwsRoleArn(plan.AwsRoleArn.ValueString())
		} else {
			updateAccountCredentialsRequestBody.SetAwsAccessKeyId(plan.AwsAccessKeyId.ValueString())
			updateAccountCredentialsRequestBody.SetAwsSecretAccessKey(plan.AwsSecretAccessKey.ValueString())
		}

		httpResp, err := updateAwsWorkspacesAccount(ctx, r.client, &resp.Diagnostics, accountId, citrixquickcreate.UpdateCustomerAccountAsyncRequest{UpdateAwsEdcAccountCredentials: &updateAccountCredentialsRequestBody})

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating AWS WorkSpaces Account Credentials: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return
		}
	}

	// Get updated account details
	account, _, getAcctErr := getAwsWorkspacesAccountUsingId(ctx, r.client, &resp.Diagnostics, accountId)
	if getAcctErr != nil {
		return
	}

	// Update resource state with new account details
	plan = plan.RefreshPropertyValues(account)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *awsWorkspacesAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesAccountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete AWS WorkSpaces Account
	deleteAccountRequest := r.client.QuickCreateClient.AccountQCS.DeleteCustomerAccountAsync(ctx, r.client.ClientConfig.CustomerId, state.AccountId.ValueString())
	httpResp, err := r.client.QuickCreateClient.AccountQCS.DeleteCustomerAccountAsyncExecute(deleteAccountRequest)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing AWS WorkSpaces Account: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}
}

func (r *awsWorkspacesAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *awsWorkspacesAccountResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AwsWorkspacesAccountResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *awsWorkspacesAccountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func readAwsWorkspacesAccountUsingId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, accountId string) (*citrixquickcreate.AwsEdcAccount, error) {
	getAccountRequest := client.QuickCreateClient.AccountQCS.GetCustomerAccountAsync(ctx, client.ClientConfig.CustomerId, accountId)
	account, _, err := util.ReadResource[*citrixquickcreate.AwsEdcAccount](getAccountRequest, ctx, client, resp, "AWS WorkSpaces Account", accountId)

	return account, err
}

func getAwsWorkspacesAccountUsingId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string) (*citrixquickcreate.AwsEdcAccount, *http.Response, error) {
	getAccountRequest := client.QuickCreateClient.AccountQCS.GetCustomerAccountAsync(ctx, client.ClientConfig.CustomerId, accountId)
	account, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AwsEdcAccount](getAccountRequest, client)

	if err != nil {
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Account: "+accountId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return account, httpResp, nil
}

func updateAwsWorkspacesAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, requestBody citrixquickcreate.UpdateCustomerAccountAsyncRequest) (*http.Response, error) {
	updateAccountRequest := client.QuickCreateClient.AccountQCS.UpdateCustomerAccountAsync(ctx, client.ClientConfig.CustomerId, accountId)
	updateAccountRequest = updateAccountRequest.UpdateCustomerAccountAsyncRequest(requestBody)
	// Had to use [any] as the ResponseBodyType because this API returns a 204 No Content status, without a response body, and go won't let us
	// omit the ResponseBodyType parameter
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[any](updateAccountRequest, client)

	if err != nil {
		diagnostics.AddError(
			"Error performing "+accountOperationTypeEnumToString(requestBody.UpdateAccount.GetAccountOperationType())+" on AWS WorkSpaces Account: "+accountId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return httpResp, err
	}

	return httpResp, nil
}

func accountOperationTypeEnumToString(operationType citrixquickcreate.UpdateAccountOperationType) string {
	switch operationType {
	case citrixquickcreate.UPDATEACCOUNTOPERATIONTYPE_RENAME_ACCOUNT:
		return "RenameAccount"
	case citrixquickcreate.UPDATEACCOUNTOPERATIONTYPE_UPDATE_AWS_EDC_ACCOUNT_CREDENTIALS:
		return "UpdateAwsEdcAccountCredentials"
	default:
		return ""
	}
}
