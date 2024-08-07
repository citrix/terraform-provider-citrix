// Copyright © 2024. Citrix Systems, Inc.
package qcs_deployment

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &awsWorkspacesDeploymentResource{}
	_ resource.ResourceWithConfigure      = &awsWorkspacesDeploymentResource{}
	_ resource.ResourceWithImportState    = &awsWorkspacesDeploymentResource{}
	_ resource.ResourceWithValidateConfig = &awsWorkspacesDeploymentResource{}
	_ resource.ResourceWithModifyPlan     = &awsWorkspacesDeploymentResource{}
)

func NewAwsWorkspacesDeploymentResource() resource.Resource {
	return &awsWorkspacesDeploymentResource{}
}

// awsWorkspacesDeploymentResource is the resource implementation.
type awsWorkspacesDeploymentResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *awsWorkspacesDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_deployment"
}

// Schema defines the schema for the resource.
func (r *awsWorkspacesDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AwsWorkspacesDeploymentResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *awsWorkspacesDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *awsWorkspacesDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	initiateDeploymentRequestBody, err := constructAddAwsWorkspacesDeploymentRequestBody(ctx, &resp.Diagnostics, plan)
	if err != nil {
		return
	}
	// Generate API request body from plan
	initateDeploymentRequest := r.client.QuickCreateClient.DeploymentQCS.InitiateDeploymentAsync(ctx, r.client.ClientConfig.CustomerId)
	initateDeploymentRequest = initateDeploymentRequest.Body(initiateDeploymentRequestBody)

	// Create new AWS Workspaces Deployment
	initiateDeploymentResponse, httpResp, err := citrixdaasclient.AddRequestData(initateDeploymentRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error initiating AWS Workspaces Deployment: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}

	deploymentId := initiateDeploymentResponse.GetDeploymentId()
	deploymentResult, err := WaitForQcsDeployment(ctx, &resp.Diagnostics, r.client, 120, deploymentId)
	if err != nil {
		if deploymentResult != nil &&
			(deploymentResult.GetDeploymentState() == citrixquickcreate.DEPLOYMENTSTATE_ERROR ||
				deploymentResult.GetDeploymentState() == citrixquickcreate.DEPLOYMENTSTATE_ERROR_INVALID_ACCOUNT) {
			resp.Diagnostics.AddError(
				"Error creating AWS Workspaces Deployment: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nDeployment is in state "+string(deploymentResult.GetDeploymentState()),
			)
			return
		}
		return
	}

	// Call Orchestration Service to update workspace maintenance modes
	workspaces := util.ObjectListToTypedArray[AwsWorkspacesDeploymentWorkspaceModel](ctx, &resp.Diagnostics, plan.Workspaces)
	if len(workspaces) != len(deploymentResult.GetWorkspaces()) {
		resp.Diagnostics.AddError(
			"Error creating AWS Workspaces Deployment: "+plan.Name.ValueString(),
			"Number of workspaces created does not match the number of workspaces in the plan",
		)
		return
	}
	if len(workspaces) > 0 {
		if plan.UserDecoupledWorkspaces.ValueBool() {
			deploymentWorkspaces := deploymentResult.GetWorkspaces()
			brokerMachineIdMaintenanceModeMap := map[string]bool{}
			for index, workspace := range workspaces {
				brokerMachineIdMaintenanceModeMap[deploymentWorkspaces[index].GetBrokerMachineId()] = workspace.MaintenanceMode.ValueBool()
			}
			err = updateMachinesMaintenceMode(ctx, &resp.Diagnostics, r.client, deploymentId, brokerMachineIdMaintenanceModeMap)
		} else {
			err = updateMachinesMaintenanceModeWithUsername(ctx, &resp.Diagnostics, r.client, deploymentResult, workspaces)
		}
		if err != nil {
			return
		}
	}

	// Map response body to schema and populate computed attribute values
	// Try getting the AWS Workspaces Deployment
	deployment, _, err := getAwsWorkspacesDeploymentUsingId(ctx, r.client, &resp.Diagnostics, deploymentResult.GetDeploymentId(), true)
	if err != nil {
		return
	}

	if deployment.GetDeploymentState() == citrixquickcreate.DEPLOYMENTSTATE_ERROR ||
		deployment.GetDeploymentState() == citrixquickcreate.DEPLOYMENTSTATE_ERROR_INVALID_ACCOUNT {
		resp.Diagnostics.AddError(
			"Error creating AWS Workspaces Deployment: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nDeployment is in state "+string(deployment.GetDeploymentState()),
		)
	}
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, *deployment)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *awsWorkspacesDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS Workspaces Deployment
	deployment, httpResp, err := getAwsWorkspacesDeploymentUsingId(ctx, r.client, &resp.Diagnostics, state.Id.ValueString(), false)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("AWS Workspaces Deployment with ID: %s not found", state.Id.ValueString()),
				fmt.Sprintf("AWS Workspaces Deployment with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.Id.ValueString()),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, *deployment)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *awsWorkspacesDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	// Retrieve values from plan
	var plan AwsWorkspacesDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state AwsWorkspacesDeploymentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 1. Update Image
	if !strings.EqualFold(plan.ImageId.ValueString(), state.ImageId.ValueString()) {
		var imageUpdateBody citrixquickcreate.ImageUpdateBody
		imageUpdateBody.SetImageId(plan.ImageId.ValueString())
		imageUpdateRequest := r.client.QuickCreateClient.DeploymentQCS.UpdateDeploymentImageAsync(ctx, r.client.ClientConfig.CustomerId, plan.Id.ValueString())
		imageUpdateRequest = imageUpdateRequest.ImageUpdateBody(imageUpdateBody)
		updateImageTask, httpResp, err := imageUpdateRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating image for AWS Workspaces Deployment: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return
		}
		err = util.WaitForQcsDeploymentTaskWithDiags(ctx, &resp.Diagnostics, r.client, 600, updateImageTask.GetTaskId(), "Update image task", plan.Name.ValueString(), "updating image for")
		if err != nil {
			return
		}
	}

	// 2. Update Running Mode and Scale Settings
	if !strings.EqualFold(plan.RunningMode.ValueString(), state.RunningMode.ValueString()) {
		newRunningMode, err := citrixquickcreate.NewAwsEdcWorkspaceRunningModeFromValue(plan.RunningMode.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating AWS Workspaces Deployment: "+plan.Name.ValueString(),
				"Error message: "+err.Error(),
			)
			return
		}
		var deploymentPropertiesUpdateBody citrixquickcreate.UpdateAwsEdcDeploymentProperties
		deploymentPropertiesUpdateBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
		deploymentPropertiesUpdateBody.SetRunningMode(*newRunningMode)
		if *newRunningMode == citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_MANUAL {
			scaleSettingRequestModel := createScaleSettingRequestModelFromConfig(ctx, &resp.Diagnostics, plan)
			deploymentPropertiesUpdateBody.SetScaleSettings(scaleSettingRequestModel)
		}

		deploymentPropertiesUpdateRequest := r.client.QuickCreateClient.DeploymentQCS.UpdateDeploymentPropertiesAsync(ctx, r.client.ClientConfig.CustomerId, plan.Id.ValueString())
		deploymentPropertiesUpdateRequest = deploymentPropertiesUpdateRequest.Body(deploymentPropertiesUpdateBody)
		updatePropertiesTask, httpResp, err := deploymentPropertiesUpdateRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating properties for AWS Workspaces Deployment: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return
		}
		err = util.WaitForQcsDeploymentTaskWithDiags(ctx, &resp.Diagnostics, r.client, 600, updatePropertiesTask.GetTaskId(), "Update deployment scale settings task", plan.Name.ValueString(), "updating scale settings for")
		if err != nil {
			return
		}
	}

	// 3. Update Workspaces
	deployment, _, err := getAwsWorkspacesDeploymentUsingId(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString(), true)
	if err != nil {
		return
	}

	if !deployment.GetUserDecoupledWorkspaces() {
		err = updateUserCoupledWorkspaces(ctx, &resp.Diagnostics, r.client, deployment, plan.Workspaces)
	} else {
		err = updateUserDecoupledWorkspaces(ctx, &resp.Diagnostics, r.client, deployment, plan.Workspaces)
	}
	if err != nil {
		return
	}

	// Fetch the latest remote configruation and update state
	deployment, _, err = getAwsWorkspacesDeploymentUsingId(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString(), true)
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, *deployment)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *awsWorkspacesDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete AWS Workspaces Deployment
	deleteDeploymentRequest := r.client.QuickCreateClient.DeploymentQCS.InitiateDeleteDeploymentAsync(ctx, r.client.ClientConfig.CustomerId, state.Id.ValueString())
	deleteDeploymentTask, httpResp, err := r.client.QuickCreateClient.DeploymentQCS.InitiateDeleteDeploymentAsyncExecute(deleteDeploymentRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error initiating deletion of AWS Workspaces Deployment: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}

	err = util.WaitForQcsDeploymentTaskWithDiags(ctx, &resp.Diagnostics, r.client, 3600, deleteDeploymentTask.GetTaskId(), "Delete deployment task", state.Name.ValueString(), "deleting")
	if err != nil {
		return
	}
}

func (r *awsWorkspacesDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *awsWorkspacesDeploymentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var config AwsWorkspacesDeploymentResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &config)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

	if !config.RootVolumeSize.IsUnknown() && !config.UserVolumeSize.IsUnknown() && !validateVolumeSize(config.RootVolumeSize.ValueInt64(), config.UserVolumeSize.ValueInt64()) {
		resp.Diagnostics.AddError(
			"Error validating AWS Workspaces Deployment: "+config.Name.ValueString(),
			"Root volume size and user volume size should be within the allowed range",
		)
	}

	if !config.RunningMode.IsUnknown() && !config.ScaleSettings.IsUnknown() {
		if strings.EqualFold(config.RunningMode.ValueString(), string(citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_ALWAYS_ON)) &&
			!config.ScaleSettings.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating AWS Workspaces Deployment: "+config.Name.ValueString(),
				fmt.Sprintf("Scale settings should not be provided when running mode is set to `%s`", string(citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_ALWAYS_ON)),
			)
		}
	}

	if !config.Workspaces.IsUnknown() {
		workspaces := util.ObjectListToTypedArray[AwsWorkspacesDeploymentWorkspaceModel](ctx, &resp.Diagnostics, config.Workspaces)
		for _, workspace := range workspaces {
			if !config.UserDecoupledWorkspaces.IsUnknown() {
				if config.UserDecoupledWorkspaces.ValueBool() {
					if !workspace.Username.IsUnknown() && !workspace.Username.IsNull() {
						resp.Diagnostics.AddError(
							"Error validating AWS Workspaces Deployment: "+config.Name.ValueString(),
							"When `user_decoupled_workspaces` is set to `true`, `username` should not be provided",
						)
						break
					}
				} else {
					if !workspace.Username.IsUnknown() && workspace.Username.IsNull() {
						resp.Diagnostics.AddError(
							"Error validating AWS Workspaces Deployment: "+config.Name.ValueString(),
							"When `user_decoupled_workspaces` is set to `false`, `username` should be provided",
						)
						break
					}
				}
			}
			if !workspace.RootVolumeSize.IsUnknown() && !workspace.UserVolumeSize.IsUnknown() && !validateVolumeSize(workspace.RootVolumeSize.ValueInt64(), workspace.UserVolumeSize.ValueInt64()) {
				resp.Diagnostics.AddError(
					"Error validating AWS Workspaces Deployment "+config.Name.ValueString(),
					"Root volume size and user volume size should be within the allowed range for workspace user: "+workspace.Username.ValueString(),
				)
				break
			}
		}
	}
}

func getAwsWorkspacesDeploymentUsingId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deploymentId string, addErrorIfNotFound bool) (*citrixquickcreate.AwsEdcDeployment, *http.Response, error) {
	getDeploymentRequest := client.QuickCreateClient.DeploymentQCS.GetDeploymentAsync(ctx, client.ClientConfig.CustomerId, deploymentId)
	deployment, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AwsEdcDeployment](getDeploymentRequest, client)

	if err != nil {
		if !addErrorIfNotFound {
			return nil, httpResp, err
		}
		diagnostics.AddError(
			"Error getting AWS Workspaces Deployment: "+deploymentId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return deployment, httpResp, nil
}

func constructAddAwsWorkspacesDeploymentRequestBody(ctx context.Context, diagnostics *diag.Diagnostics, config AwsWorkspacesDeploymentResourceModel) (citrixquickcreate.InitiateAwsEdcDeployment, error) {
	var initiateDeploymentBody citrixquickcreate.InitiateAwsEdcDeployment
	initiateDeploymentBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	initiateDeploymentBody.SetDeploymentName(config.Name.ValueString())
	initiateDeploymentBody.SetConnectionId(config.DirectoryId.ValueString())
	initiateDeploymentBody.SetImageId(config.ImageId.ValueString())
	computeType, err := citrixquickcreate.NewAwsEdcWorkspaceComputeFromValue(config.Performance.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating AWS Workspaces Deployment: "+config.Name.ValueString(),
			"Error message: "+err.Error(),
		)
		return initiateDeploymentBody, err
	}
	initiateDeploymentBody.SetComputeType(*computeType)
	initiateDeploymentBody.SetRootVolumeSize(int32(config.RootVolumeSize.ValueInt64()))
	initiateDeploymentBody.SetUserVolumeSize(int32(config.UserVolumeSize.ValueInt64()))
	initiateDeploymentBody.SetVolumesEncrypted(config.VolumesEncrypted.ValueBool())
	if !config.VolumesEncryptionKey.IsNull() {
		initiateDeploymentBody.SetVolumesEncryptionKey(config.VolumesEncryptionKey.ValueString())
	}
	runningMode, err := citrixquickcreate.NewAwsEdcWorkspaceRunningModeFromValue(config.RunningMode.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating AWS Workspaces Deployment: "+config.Name.ValueString(),
			"Error message: "+err.Error(),
		)
		return initiateDeploymentBody, err
	}
	initiateDeploymentBody.SetRunningMode(*runningMode)
	if *runningMode == citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_MANUAL {
		scaleSettingRequestModel := createScaleSettingRequestModelFromConfig(ctx, diagnostics, config)
		initiateDeploymentBody.SetScaleSettings(scaleSettingRequestModel)
	}

	initiateDeploymentBody.SetUserDecoupledWorkspaces(config.UserDecoupledWorkspaces.ValueBool())
	if !config.Workspaces.IsNull() {
		workspaces := util.ObjectListToTypedArray[AwsWorkspacesDeploymentWorkspaceModel](ctx, diagnostics, config.Workspaces)
		workspaceRequestModels := []citrixquickcreate.AddAwsEdcWorkspace{}
		for _, workspace := range workspaces {
			var workspaceRequestModel citrixquickcreate.AddAwsEdcWorkspace
			if config.UserDecoupledWorkspaces.ValueBool() {
				workspaceRequestModel.SetUsername(util.UsernameForDecoupledWorkspaces)
			} else {
				workspaceRequestModel.SetUsername(workspace.Username.ValueString())
			}
			workspaceRequestModel.SetUserVolumeSize(int32(workspace.UserVolumeSize.ValueInt64()))
			workspaceRequestModel.SetRootVolumeSize(int32(workspace.RootVolumeSize.ValueInt64()))
			workspaceRequestModels = append(workspaceRequestModels, workspaceRequestModel)
		}

		initiateDeploymentBody.SetWorkspaces(workspaceRequestModels)
	}
	return initiateDeploymentBody, nil
}

func createScaleSettingRequestModelFromConfig(ctx context.Context, diagnostics *diag.Diagnostics, config AwsWorkspacesDeploymentResourceModel) citrixquickcreate.ScaleSettings {
	var scaleSettingRequestModel citrixquickcreate.ScaleSettings
	scaleSettingRequestModel.SetAutoScaleEnabled(true)

	if !config.ScaleSettings.IsNull() {
		scaleSetting := util.ObjectValueToTypedObject[AwsWorkspacesScaleSettingsModel](ctx, diagnostics, config.ScaleSettings)
		scaleSettingRequestModel.SetSessionIdleTimeoutMinutes(int32(scaleSetting.SessionIdleTimeoutMinutes.ValueInt64()))
		scaleSettingRequestModel.SetOffPeakDisconnectTimeoutMinutes(int32(scaleSetting.OffPeakDisconnectTimeoutMinutes.ValueInt64()))
		scaleSettingRequestModel.SetOffPeakLogOffTimeoutMinutes(int32(scaleSetting.OffPeakLogOffTimeoutMinutes.ValueInt64()))
		scaleSettingRequestModel.SetOffPeakBufferSizePercent(int32(scaleSetting.OffPeakBufferSizePercentage.ValueInt64()))
	} else {
		scaleSettingRequestModel.SetSessionIdleTimeoutMinutes(int32(util.DefaultQcsAwsWorkspacesSessionIdleTimeoutMinutes))
		scaleSettingRequestModel.SetOffPeakDisconnectTimeoutMinutes(int32(util.DefaultQcsAwsWorkspacesOffPeakDisconnectTimeoutMinutes))
		scaleSettingRequestModel.SetOffPeakLogOffTimeoutMinutes(int32(util.DefaultQcsAwsWorkspacesOffPeakLogOffTimeoutMinutes))
		scaleSettingRequestModel.SetOffPeakBufferSizePercent(int32(util.DefaultQcsAwsWorkspacesOffPeakBufferSizePercent))
	}
	return scaleSettingRequestModel
}

func validateVolumeSize(rootVolumeSize int64, userVolumeSize int64) bool {
	if rootVolumeSize == 80 {
		if userVolumeSize == 10 || userVolumeSize == 50 || userVolumeSize == 100 {
			return true
		}
	} else if rootVolumeSize >= 175 && rootVolumeSize <= 2000 {
		if userVolumeSize >= 100 && userVolumeSize <= 2000 {
			return true
		}
	}
	return false
}

func WaitForQcsDeployment(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, maxWaitTimeInMinutes int, deploymentId string) (*citrixquickcreate.AwsEdcDeployment, error) {
	startTime := time.Now()
	sleepDuration := time.Second * time.Duration(30)
	var deploymentResponseModel *citrixquickcreate.AwsEdcDeployment
	var err error
	deploymentNotFoundCounter := 0

	for {
		if time.Since(startTime) > time.Minute*time.Duration(maxWaitTimeInMinutes) {
			diagnostics.AddError(
				"Error waiting for AWS Workspaces Deployment: "+deploymentId,
				fmt.Sprintf("Error message: wait time exceeded the maximum allowed time of %d minutes", maxWaitTimeInMinutes),
			)
			break
		}

		deploymentResponseModel, httpResp, err := getAwsWorkspacesDeploymentUsingId(ctx, client, diagnostics, deploymentId, false)
		if err != nil {
			if deploymentNotFoundCounter < 5 && httpResp.StatusCode == http.StatusNotFound {
				deploymentNotFoundCounter++
				time.Sleep(sleepDuration)
				continue
			}
			return deploymentResponseModel, err
		}

		deploymentState := deploymentResponseModel.GetDeploymentState()

		if deploymentState != citrixquickcreate.DEPLOYMENTSTATE_ACTIVE &&
			deploymentState != citrixquickcreate.DEPLOYMENTSTATE_DELETED &&
			deploymentState != citrixquickcreate.DEPLOYMENTSTATE_ERROR_INVALID_ACCOUNT &&
			deploymentState != citrixquickcreate.DEPLOYMENTSTATE_ERROR {
			time.Sleep(sleepDuration)
			continue
		}

		return deploymentResponseModel, err
	}

	return deploymentResponseModel, err
}

func generateBatchApiHeaders(ctx context.Context, client *citrixdaasclient.CitrixDaasClient) (context.Context, []citrixorchestration.NameValueStringPairModel, *http.Response, error) {
	headers := []citrixorchestration.NameValueStringPairModel{}

	cwsAuthToken, httpResp, err := client.SignIn()
	ctx = tflog.SetField(ctx, "cws_auth_token", cwsAuthToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "cws_auth_token")
	if err != nil {
		return ctx, headers, httpResp, err
	}

	if cwsAuthToken != "" {
		token := strings.Split(cwsAuthToken, "=")[1]
		ctx = tflog.SetField(ctx, "cws_auth_token_value", token)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "cws_auth_token_value")
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("Authorization")
		header.SetValue("Bearer " + token)
		headers = append(headers, header)
	}

	return ctx, headers, httpResp, err
}

func updateMachinesMaintenanceModeWithUsername(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deployment *citrixquickcreate.AwsEdcDeployment, plannedWorkspaces []AwsWorkspacesDeploymentWorkspaceModel) error {
	usernameBrokerMachineIdMap := getUsernameMachineIdMap(deployment)
	brokerMachineIdMaintenanceModeMap := map[string]bool{}
	for _, workspace := range plannedWorkspaces {
		brokerMachineIdMaintenanceModeMap[usernameBrokerMachineIdMap[workspace.Username.ValueString()]] = workspace.MaintenanceMode.ValueBool()
	}

	return updateMachinesMaintenceMode(ctx, diagnostics, client, deployment.GetDeploymentId(), brokerMachineIdMaintenanceModeMap)
}

func updateMachinesMaintenceMode(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deploymentId string, brokerMachineIdMaintenanceModeMap map[string]bool) error {
	ctx, batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, client)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		diagnostics.AddError(
			"Error updating Maintenance Mode status for machine(s)",
			"TransactionId: "+txId+
				"\nCould not update maintenance mode status for machine(s), unexpected error: "+util.ReadClientError(err), // Use ReadClientError as this is orchestration service call
		)
		return err
	}
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}

	index := 0
	for brokerMachineId, maintenanceMode := range brokerMachineIdMaintenanceModeMap {
		var updateMachineModel citrixorchestration.UpdateMachineRequestModel
		updateMachineModel.SetInMaintenanceMode(maintenanceMode)
		updateMachineStringBody, err := util.ConvertToString(updateMachineModel)
		if err != nil {
			diagnostics.AddError(
				"Error updating Maintenance Mode status for machine "+brokerMachineId,
				"An unexpected error occurred: "+err.Error(),
			)
			return err
		}
		relativeUrl := fmt.Sprintf("/Machines/%s?async=true", brokerMachineId)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("maintenanceMode%s", strconv.Itoa(index)))
		batchRequestItem.SetMethod(http.MethodPatch)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetBody(updateMachineStringBody)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
		index++
	}

	if len(batchRequestItems) > 0 {
		// If there are any machines that need to be put in maintenance mode
		var batchRequestModel citrixorchestration.BatchRequestModel
		batchRequestModel.SetItems(batchRequestItems)
		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
		if err != nil {
			diagnostics.AddError(
				"Error updating maintenance mode status for machine(s) in deployment "+deploymentId,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err), // Use ReadClientError as this is orchestration service call
			)
			return err
		}

		if successfulJobs < len(batchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while updating maintenance mode status for machine(s). %d of %d machines were updated successfully.", successfulJobs, len(batchRequestItems))
			err = fmt.Errorf(errMsg)
			diagnostics.AddError(
				"Error updating maintenance mode status for machine(s) in deployment "+deploymentId,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return err
		}
	}
	return nil
}

func getUsernameMachineIdMap(deployment *citrixquickcreate.AwsEdcDeployment) map[string]string {
	usernameMachineIdMap := map[string]string{}
	if len(deployment.GetWorkspaces()) > 0 {
		for _, workspace := range deployment.GetWorkspaces() {
			usernameMachineIdMap[workspace.GetUsername()] = workspace.GetBrokerMachineId()
		}
	}

	return usernameMachineIdMap
}

func updateUserCoupledWorkspaces(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deployment *citrixquickcreate.AwsEdcDeployment, planWorkspaces types.List) error {
	// 1. Find usernames in plan
	workspaceUsernamesInPlan := []string{}
	workspaceUsernameConfigMapInPlan := map[string]AwsWorkspacesDeploymentWorkspaceModel{}
	workspacesInPlan := util.ObjectListToTypedArray[AwsWorkspacesDeploymentWorkspaceModel](ctx, diagnostics, planWorkspaces)
	for _, workspace := range workspacesInPlan {
		workspaceUsernamesInPlan = append(workspaceUsernamesInPlan, workspace.Username.ValueString())
		workspaceUsernameConfigMapInPlan[workspace.Username.ValueString()] = workspace
	}

	// 2. Delete removed workspaces and update existing ones
	workspaceIdsForDeletion := []string{}
	machinesForMaintenanceModeBeforeDeletion := []AwsWorkspacesDeploymentWorkspaceModel{}
	updateWorkspaceMachineRequests := []citrixquickcreate.DeploymentQCSUpdateMachineAsyncRequest{}
	for _, workspace := range deployment.GetWorkspaces() {
		if !slices.Contains(workspaceUsernamesInPlan, workspace.GetUsername()) {
			workspaceIdsForDeletion = append(workspaceIdsForDeletion, workspace.GetWorkspaceId())
			machinesForMaintenanceModeBeforeDeletion = append(machinesForMaintenanceModeBeforeDeletion, AwsWorkspacesDeploymentWorkspaceModel{
				Username:        types.StringValue(workspace.GetUsername()),
				BrokerMachineId: types.StringValue(workspace.GetBrokerMachineId()),
				MaintenanceMode: types.BoolValue(true),
			})
		} else {
			// Update existing workspace
			plannedWorkspace := workspaceUsernameConfigMapInPlan[workspace.GetUsername()]
			if plannedWorkspace.RootVolumeSize.ValueInt64() != int64(workspace.GetRootVolumeSize()) ||
				plannedWorkspace.UserVolumeSize.ValueInt64() != int64(workspace.GetUserVolumeSize()) {
				// Update existing workspace
				var updateWorkspaceRequestDetail citrixquickcreate.UpdateAwsEdcDeploymentMachine
				updateWorkspaceRequestDetail.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
				updateWorkspaceRequestDetail.SetRootVolumeSize(int32(plannedWorkspace.RootVolumeSize.ValueInt64()))
				updateWorkspaceRequestDetail.SetUserVolumeSize(int32(plannedWorkspace.UserVolumeSize.ValueInt64()))
				updateWorkspaceRequest := client.QuickCreateClient.DeploymentQCS.UpdateMachineAsync(ctx, client.ClientConfig.CustomerId, deployment.GetDeploymentId(), workspace.GetMachineId())
				updateWorkspaceRequest = updateWorkspaceRequest.Body(updateWorkspaceRequestDetail)
				updateWorkspaceMachineRequests = append(updateWorkspaceMachineRequests, updateWorkspaceRequest)
			}
		}
	}

	// 2.1 Delete removed workspaces
	if len(workspaceIdsForDeletion) > 0 {
		err := updateMachinesMaintenanceModeWithUsername(ctx, diagnostics, client, deployment, machinesForMaintenanceModeBeforeDeletion)
		if err != nil {
			return err
		}
		err = deleteAwsWorkspaceMachines(ctx, diagnostics, client, deployment, workspaceIdsForDeletion)
		if err != nil {
			return err
		}
	}

	// 2.2 Update existing workspaces
	for _, updateWorkspaceRequest := range updateWorkspaceMachineRequests {
		deploymentTask, httpResp, err := updateWorkspaceRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error updating existing workspaces in AWS Workspaces Deployment: "+deployment.GetDeploymentName(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return err
		}

		err = util.WaitForQcsDeploymentTaskWithDiags(ctx, diagnostics, client, 3600, deploymentTask.GetTaskId(), "Update existing workspaces in deployment task", deployment.GetDeploymentName(), "updating existing workspaces in")
		if err != nil {
			return err
		}
	}

	// 3. Find workspaces to be created
	deployment, _, err := getAwsWorkspacesDeploymentUsingId(ctx, client, diagnostics, deployment.GetDeploymentId(), true)
	if err != nil {
		return err
	}
	existingWorkspaceUsernames := []string{}
	for _, workspace := range deployment.GetWorkspaces() {
		existingWorkspaceUsernames = append(existingWorkspaceUsernames, workspace.GetUsername())
	}

	addMachinesRequestDetails := []citrixquickcreate.AddAwsEdcWorkspace{}
	for _, username := range workspaceUsernamesInPlan {
		if !slices.Contains(existingWorkspaceUsernames, username) {
			// Create new workspace
			workspaceConfig := workspaceUsernameConfigMapInPlan[username]
			var addWorkspaceRequestDetail citrixquickcreate.AddAwsEdcWorkspace
			addWorkspaceRequestDetail.SetUsername(username)
			addWorkspaceRequestDetail.SetRootVolumeSize(int32(workspaceConfig.RootVolumeSize.ValueInt64()))
			addWorkspaceRequestDetail.SetUserVolumeSize(int32(workspaceConfig.UserVolumeSize.ValueInt64()))
			addMachinesRequestDetails = append(addMachinesRequestDetails, addWorkspaceRequestDetail)
		}
	}

	if len(addMachinesRequestDetails) > 0 {
		addMachinesRequestBody := citrixquickcreate.AddAwsEdcDeploymentMachines{}
		addMachinesRequestBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
		addMachinesRequestBody.SetWorkspaces(addMachinesRequestDetails)
		addMachinesRequest := client.QuickCreateClient.DeploymentQCS.AddMachineAsync(ctx, client.ClientConfig.CustomerId, deployment.GetDeploymentId())
		addMachinesRequest = addMachinesRequest.Body(addMachinesRequestBody)
		deploymentTask, httpResp, err := addMachinesRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error adding new workspaces to AWS Workspaces Deployment: "+deployment.GetDeploymentName(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return err
		}
		err = util.WaitForQcsDeploymentTaskWithDiags(ctx, diagnostics, client, 3600, deploymentTask.GetTaskId(), "Add workspaces to deployment task", deployment.GetDeploymentName(), "adding new workspaces to")
		if err != nil {
			return err
		}
	}

	// 4. Call Orchestration Service to update workspace maintenance modes
	deployment, _, err = getAwsWorkspacesDeploymentUsingId(ctx, client, diagnostics, deployment.GetDeploymentId(), true)
	if err != nil {
		return err
	}

	err = updateMachinesMaintenanceModeWithUsername(ctx, diagnostics, client, deployment, workspacesInPlan)
	if err != nil {
		return err
	}

	return nil
}

func updateUserDecoupledWorkspaces(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deployment *citrixquickcreate.AwsEdcDeployment, plannedWorkspacesList types.List) error {
	brokerIdMapForMaintenace := map[string]bool{}
	workspaceIdsForDeletion := []string{}
	for _, workspace := range deployment.GetWorkspaces() {
		brokerIdMapForMaintenace[workspace.GetBrokerMachineId()] = true
		workspaceIdsForDeletion = append(workspaceIdsForDeletion, workspace.GetMachineId())
	}

	// 1. Set existing workspaces to maintenance mode
	err := updateMachinesMaintenceMode(ctx, diagnostics, client, deployment.GetDeploymentId(), brokerIdMapForMaintenace)
	if err != nil {
		return err
	}

	// 2. Delete existing workspaces
	err = deleteAwsWorkspaceMachines(ctx, diagnostics, client, deployment, workspaceIdsForDeletion)
	if err != nil {
		return err
	}

	// 3. Add new workspaces
	addMachinesRequestDetails := []citrixquickcreate.AddAwsEdcWorkspace{}
	workspaces := util.ObjectListToTypedArray[AwsWorkspacesDeploymentWorkspaceModel](ctx, diagnostics, plannedWorkspacesList)
	for _, workspace := range workspaces {
		var addWorkspaceRequestDetail citrixquickcreate.AddAwsEdcWorkspace
		addWorkspaceRequestDetail.SetUsername(util.UsernameForDecoupledWorkspaces)
		addWorkspaceRequestDetail.SetRootVolumeSize(int32(workspace.RootVolumeSize.ValueInt64()))
		addWorkspaceRequestDetail.SetUserVolumeSize(int32(workspace.UserVolumeSize.ValueInt64()))
		addMachinesRequestDetails = append(addMachinesRequestDetails, addWorkspaceRequestDetail)
	}

	if len(addMachinesRequestDetails) > 0 {
		addMachinesRequestBody := citrixquickcreate.AddAwsEdcDeploymentMachines{}
		addMachinesRequestBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
		addMachinesRequestBody.SetWorkspaces(addMachinesRequestDetails)
		addMachinesRequest := client.QuickCreateClient.DeploymentQCS.AddMachineAsync(ctx, client.ClientConfig.CustomerId, deployment.GetDeploymentId())
		addMachinesRequest = addMachinesRequest.Body(addMachinesRequestBody)
		deploymentTask, httpResp, err := addMachinesRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error adding new workspaces to AWS Workspaces Deployment: "+deployment.GetDeploymentName(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadQcsClientError(err),
			)
			return err
		}
		err = util.WaitForQcsDeploymentTaskWithDiags(ctx, diagnostics, client, 3600, deploymentTask.GetTaskId(), "Add workspaces to deployment task", deployment.GetDeploymentName(), "adding new workspaces to")
		if err != nil {
			return err
		}
	}

	/// 4. Update maintenace mode
	deployment, _, err = getAwsWorkspacesDeploymentUsingId(ctx, client, diagnostics, deployment.GetDeploymentId(), true)
	if err != nil {
		return err
	}
	deploymentWorkspaces := deployment.GetWorkspaces()
	if len(deploymentWorkspaces) != len(workspaces) {
		diagnostics.AddError(
			"Error updating AWS Workspaces Deployment: "+deployment.GetDeploymentName(),
			"Number of workspaces in updated deployment does not match the number of workspaces in the plan",
		)
	}
	brokerMachineIdMaintenanceModeMap := map[string]bool{}
	for index, workspace := range workspaces {
		brokerMachineIdMaintenanceModeMap[deploymentWorkspaces[index].GetBrokerMachineId()] = workspace.MaintenanceMode.ValueBool()
	}
	err = updateMachinesMaintenceMode(ctx, diagnostics, client, deployment.GetDeploymentId(), brokerMachineIdMaintenanceModeMap)
	if err != nil {
		return err
	}
	return nil
}

func deleteAwsWorkspaceMachines(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deployment *citrixquickcreate.AwsEdcDeployment, workspaceIdsForDeletion []string) error {
	var removeMachinesBody citrixquickcreate.MachinesDeleteBody
	removeMachinesBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	removeMachinesBody.SetMachineIds(workspaceIdsForDeletion)
	removeMachinesRequest := client.QuickCreateClient.DeploymentQCS.RemoveMachinesAsync(ctx, client.ClientConfig.CustomerId, deployment.GetDeploymentId())
	removeMachinesRequest = removeMachinesRequest.MachinesDeleteBody(removeMachinesBody)
	deploymentTask, httpResp, err := removeMachinesRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating AWS Workspaces Deployment: "+deployment.GetDeploymentName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return err
	}
	err = util.WaitForQcsDeploymentTaskWithDiags(ctx, diagnostics, client, 3600, deploymentTask.GetTaskId(), "Update deployment task", deployment.GetDeploymentName(), "updating")
	if err != nil {
		return err
	}
	return nil
}

func (r *awsWorkspacesDeploymentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
