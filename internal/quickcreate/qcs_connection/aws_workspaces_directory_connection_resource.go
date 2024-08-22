package qcs_connection

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	citrixquickcreate "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &awsWorkspacesDirectoryConnectionResource{}
	_ resource.ResourceWithConfigure      = &awsWorkspacesDirectoryConnectionResource{}
	_ resource.ResourceWithImportState    = &awsWorkspacesDirectoryConnectionResource{}
	_ resource.ResourceWithValidateConfig = &awsWorkspacesDirectoryConnectionResource{}
	_ resource.ResourceWithModifyPlan     = &awsWorkspacesDirectoryConnectionResource{}
)

func NewAwsWorkspacesDirectoryConnectionResource() resource.Resource {
	return &awsWorkspacesDirectoryConnectionResource{}
}

// directoryConnectionResource is the resource implementation.
type awsWorkspacesDirectoryConnectionResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *awsWorkspacesDirectoryConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_directory_connection"
}

// Schema defines the schema for the resource.
func (r *awsWorkspacesDirectoryConnectionResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AwsWorkspacesDirectoryConnectionResourceModel{}.GetSchema()
}

// Configure adds the proider configured client to the resource.
func (r *awsWorkspacesDirectoryConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *awsWorkspacesDirectoryConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesDirectoryConnectionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Connection API request body from plan
	var directoryDetails citrixquickcreate.AddAwsEdcDirectoryConnection
	directoryDetails.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	directoryDetails.SetName(plan.Name.ValueString())
	directoryDetails.SetZoneId(plan.ZoneId.ValueString())
	directoryDetails.SetResourceLocationId(plan.ResourceLocationId.ValueString())
	directoryDetails.SetDirectoryId(plan.Directory.ValueString())
	subnets := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Subnets)
	directoryDetails.SetSubnet1Id(subnets[0])
	directoryDetails.SetSubnet2Id(subnets[1])
	directoryDetails.SetTenancy(citrixquickcreate.AwsEdcDirectoryTenancy(plan.Tenancy.ValueString()))
	directoryDetails.SetEnableWorkDocs(false)
	directoryDetails.SetUserEnabledAsLocalAdministrator(plan.UserEnabledAsLocalAdministrator.ValueBool())
	directoryDetails.SetSecurityGroupId(plan.SecurityGroup.ValueString())
	directoryDetails.SetDefaultOu(plan.DefaultOu.ValueString())
	directoryDetails.SetEnableMaintenanceMode(false)

	// Add the Directory Connection
	addDirectoryConnectionResponse, _, err := addAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, plan.AccountId.ValueString(), directoryDetails)
	if err != nil {
		// Error was logged in addAwsWorkspacesDirectoryConnection. Just return
		return
	}

	// Try getting the new AWS WorkSpaces Directory Connection
	directoryConnection, _, err := getAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, plan.AccountId.ValueString(), addDirectoryConnectionResponse.GetResourceConnectionId(), false)
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, directoryConnection)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *awsWorkspacesDirectoryConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesDirectoryConnectionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Directory Connection
	directoryConnection, httpResp, err := getAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, state.AccountId.ValueString(), state.DirectoryConnectionId.ValueString(), true)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Remove from state
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("AWS WorkSpaces Directory Connection with ID: %s not found", state.DirectoryConnectionId.ValueString()),
				fmt.Sprintf("AWS WorkSpaces Directory Connection with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.DirectoryConnectionId.ValueString()),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, directoryConnection)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Updates the resource and sets the updated Terraform state on success.
func (r *awsWorkspacesDirectoryConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesDirectoryConnectionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepary request body
	// Note that this is missing "EnableInternetAccess" because the property is not available on the create model, and we decided
	// that we should just omit it for now. If we change our mind in the future, we can update the model and this code.
	updateDirectoryConnectionRequestBody := citrixquickcreate.UpdateAwsEdcDirectoryConnection{}
	updateDirectoryConnectionRequestBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	updateDirectoryConnectionRequestBody.SetDefaultOU(plan.DefaultOu.ValueString())
	updateDirectoryConnectionRequestBody.SetEnableWorkDocs(false)
	updateDirectoryConnectionRequestBody.SetUserEnabledAsLocalAdministrator(plan.UserEnabledAsLocalAdministrator.ValueBool())
	updateDirectoryConnectionRequestBody.SetEnableMaintananceMode(false)
	updateDirectoryConnectionRequestBody.SetSecurityGroupId(plan.SecurityGroup.ValueString())

	// Update the Directory Connection
	accountId := plan.AccountId.ValueString()
	connectionId := plan.DirectoryConnectionId.ValueString()
	_, _, err := updateAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, accountId, connectionId, updateDirectoryConnectionRequestBody)
	if err != nil {
		return
	}

	// Try getting the AWS WorkSpaces Directory Connection
	directoryConnection, httpResp, err := getAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, plan.AccountId.ValueString(), connectionId, true)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			// Remove from state
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("AWS WorkSpaces Directory Connection with ID: %s not found", connectionId),
				fmt.Sprintf("AWS WorkSpaces Directory Connection with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", connectionId),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	// Update resource state with new connection details
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, directoryConnection)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete removes the resource and removes the Terraform state on success.
func (r *awsWorkspacesDirectoryConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesDirectoryConnectionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Directory Connection
	_, httpResp, err := getAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, state.AccountId.ValueString(), state.DirectoryConnectionId.ValueString(), true)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			return
		}
		return
	}

	// Remove the AWS WorkSpaces Directory Connection
	removeAwsWorkspacesDirectoryConnection(ctx, r.client, &resp.Diagnostics, state.AccountId.ValueString(), state.DirectoryConnectionId.ValueString())
}

// ImportState imports the resource state from the given ID
func (r *awsWorkspacesDirectoryConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: accountId,directoryConnectionId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

// ValidateConfig validates the configuration of the resource.
func (r *awsWorkspacesDirectoryConnectionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from config
	var data AwsWorkspacesDirectoryConnectionResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func addAwsWorkspacesDirectoryConnection(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, requestBody citrixquickcreate.AddAwsEdcDirectoryConnection) (*citrixquickcreate.ResourceConnectionTask, *http.Response, error) {
	addDirectoryConnectionRequest := client.QuickCreateClient.ConnectionQCS.AddResourceConnectionAsync(ctx, client.ClientConfig.CustomerId, accountId)
	addDirectoryConnectionRequest = addDirectoryConnectionRequest.Body(requestBody)
	// Initiate the addition of the Directory Connection
	directoryConnectionTask, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.ResourceConnectionTask](addDirectoryConnectionRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error adding AWS WorkSpaces Directory Connection: "+requestBody.GetName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	// Wait for the task
	pollTaskResponse, httpResp, err := util.PollQcsTask(ctx, client, diagnostics, directoryConnectionTask.GetTaskId(), 10, 300)
	if err != nil {
		// Error messages logged in pollQcsTask. Just return
		return nil, httpResp, err
	}

	return pollTaskResponse.ResourceConnectionTask, httpResp, nil

}

func getAwsWorkspacesDirectoryConnection(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, directoryConnectionId string, returnIfNotFound bool) (*citrixquickcreate.AwsEdcDirectoryConnection, *http.Response, error) {
	getDirectoryConnectionRequest := client.QuickCreateClient.ConnectionQCS.GetResourceConnectionAsync(ctx, client.ClientConfig.CustomerId, accountId, directoryConnectionId)
	directoryConnection, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AwsEdcDirectoryConnection](getDirectoryConnectionRequest, client)

	if err != nil {
		if returnIfNotFound && httpResp.StatusCode == http.StatusNotFound {
			return nil, httpResp, err
		}
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Directory Connection: "+directoryConnectionId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return directoryConnection, httpResp, nil
}

func updateAwsWorkspacesDirectoryConnection(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, connectionId string, requestBody citrixquickcreate.UpdateAwsEdcDirectoryConnection) (*citrixquickcreate.AwsEdcDirectoryConnection, *http.Response, error) {
	updateDirectoryConnectionRequest := client.QuickCreateClient.ConnectionQCS.ModifyResourceConnectionAsync(ctx, client.ClientConfig.CustomerId, accountId, connectionId)
	updateDirectoryConnectionRequest = updateDirectoryConnectionRequest.Body(requestBody)
	directoryConnection, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AwsEdcDirectoryConnection](updateDirectoryConnectionRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error updating AWS WorkSpaces Directory Connection: "+connectionId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return directoryConnection, httpResp, nil
}

func removeAwsWorkspacesDirectoryConnection(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, connectionId string) (*citrixquickcreate.ResourceConnectionTask, *http.Response, error) {
	removeDirectoryConnectionRequest := client.QuickCreateClient.ConnectionQCS.RemoveResourceConnectionAsync(ctx, client.ClientConfig.CustomerId, accountId, connectionId)
	taskResp, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.ResourceConnectionTask](removeDirectoryConnectionRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error removing AWS WorkSpaces Directory Connection: "+connectionId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return taskResp, httpResp, nil
}

func (r *awsWorkspacesDirectoryConnectionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
