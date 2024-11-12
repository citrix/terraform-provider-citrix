// Copyright © 2024. Citrix Systems, Inc.
package qcs_image

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	_ resource.Resource                   = &awsWorkspacesImageResource{}
	_ resource.ResourceWithConfigure      = &awsWorkspacesImageResource{}
	_ resource.ResourceWithImportState    = &awsWorkspacesImageResource{}
	_ resource.ResourceWithValidateConfig = &awsWorkspacesImageResource{}
	_ resource.ResourceWithModifyPlan     = &awsWorkspacesImageResource{}
)

func NewAwsEdcImageResource() resource.Resource {
	return &awsWorkspacesImageResource{}
}

// awsWorkspacesImageResource is the resource implementation.
type awsWorkspacesImageResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *awsWorkspacesImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_image"
}

// Schema defines the schema for the resource.
func (r *awsWorkspacesImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AwsWorkspacesImageModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *awsWorkspacesImageResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create is the implementation of the Create method in the resource.ResourceWithValidateConfig interface.
func (r *awsWorkspacesImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AwsWorkspacesImageModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var importImageBody citrixquickcreate.ImportAwsEdcImage
	importImageBody.SetAccountType(citrixquickcreate.ACCOUNTTYPE_AWSEDC)
	importImageBody.SetName(plan.Name.ValueString())
	importImageBody.SetDescription(plan.Description.ValueString())
	importImageBody.SetEc2ImageId(plan.AwsImageId.ValueString())
	ingestionProcessEnum, err := citrixquickcreate.NewAwsEdcWorkspaceImageIngestionProcessFromValue(plan.IngestionProcess.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting AWS EDC Workspace Image Ingestion Process",
			"Error Message: "+err.Error(),
		)
		return
	}
	importImageBody.SetIngestionProcess(*ingestionProcessEnum)
	sessionSupportEnum, err := citrixquickcreate.NewSessionSupportFromValue(plan.SessionSupport.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting AWS EDC Workspace Image Session Support",
			"Error Message: "+err.Error(),
		)
		return
	}
	importImageBody.SetSessionSupport(*sessionSupportEnum)

	// Generate API request from plan
	importImageRequest := r.client.QuickCreateClient.ImageQCS.ImportImageAsync(ctx, r.client.ClientConfig.CustomerId, plan.AccountId.ValueString())
	importImageRequest = importImageRequest.Body(importImageBody)

	// Import new AWS WorkSpaces Image
	importImageResponse, httpResp, err := citrixdaasclient.AddRequestData(importImageRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing AWS WorkSpaces Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}

	// Try getting the new AWS WorkSpaces Image
	image, httpResp, err := waitForImageImportCompletion(ctx, r.client, &resp.Diagnostics, importImageResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting AWS WorkSpaces Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}
	if image.GetWorkspaceImageState() != citrixquickcreate.AWSEDCWORKSPACEIMAGESTATE_AVAILABLE {
		resp.Diagnostics.AddError(
			"Error importing AWS WorkSpaces Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Image state is "+util.AwsEdcWorkspaceImageStateEnumToString(image.GetWorkspaceImageState()),
		)
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, image)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is the implementation of the Read method in the resource.ResourceWithRead interface.
func (r *awsWorkspacesImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesImageModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Image
	image, _, err := getAwsWorkspacesImageWithId(ctx, r.client, &resp.Diagnostics, state.AccountId.ValueString(), state.Id.ValueString(), true)
	if err != nil {
		// Remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, true, image)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is the implementation of the Update method in the resource.Resource interface.
func (r *awsWorkspacesImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	resp.Diagnostics.AddError(
		"Error updating AWS WorkSpaces Image",
		"Update operation is not supported for AWS WorkSpaces Image",
	)
}

// Delete is the implementation of the Delete method in the resource.Resource interface.
func (r *awsWorkspacesImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AwsWorkspacesImageModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete AWS WorkSpaces Image
	deleteImageRequest := r.client.QuickCreateClient.ImageQCS.RemoveImageAsync(ctx, r.client.ClientConfig.CustomerId, state.AccountId.ValueString(), state.Id.ValueString())
	httpResp, err := r.client.QuickCreateClient.ImageQCS.RemoveImageAsyncExecute(deleteImageRequest)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing AWS WorkSpaces Image: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return
	}
}

func (r *awsWorkspacesImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: accountId,imageId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (r *awsWorkspacesImageResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AwsWorkspacesImageModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *awsWorkspacesImageResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickCreateClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func getAwsWorkspacesImageWithId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, imageId string, addWarningIfNotFound bool) (*citrixquickcreate.AwsEdcImage, *http.Response, error) {
	getImageRequest := client.QuickCreateClient.ImageQCS.GetImageAsync(ctx, client.ClientConfig.CustomerId, accountId, imageId)
	image, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.AwsEdcImage](getImageRequest, client)

	if err != nil {
		if addWarningIfNotFound && httpResp.StatusCode == http.StatusNotFound {
			diagnostics.AddWarning(
				fmt.Sprintf("AWS WorkSpaces Image with ID: %s not found", imageId),
				fmt.Sprintf("AWS WorkSpaces Image with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", imageId),
			)
			return nil, httpResp, err
		}
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Image: "+imageId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	return image, httpResp, nil
}

func waitForImageImportCompletion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, image *citrixquickcreate.AwsEdcImage) (*citrixquickcreate.AwsEdcImage, *http.Response, error) {
	// default polling to every 10 seconds
	startTime := time.Now()
	imageId := image.GetImageId()

	for {
		if time.Since(startTime) > time.Minute*time.Duration(120) {
			break
		}

		image, httpResp, err := getAwsWorkspacesImageWithId(ctx, client, diagnostics, image.GetAccountId(), imageId, false)
		if err != nil {
			return nil, httpResp, err
		}

		if image.GetWorkspaceImageState() == citrixquickcreate.AWSEDCWORKSPACEIMAGESTATE_PENDING {
			time.Sleep(time.Second * time.Duration(30))
			continue
		}

		return image, httpResp, err
	}

	return image, nil, nil
}
