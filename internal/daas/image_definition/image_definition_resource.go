// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"net/http"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &ImageDefinitionResource{}
	_ resource.ResourceWithConfigure      = &ImageDefinitionResource{}
	_ resource.ResourceWithImportState    = &ImageDefinitionResource{}
	_ resource.ResourceWithValidateConfig = &ImageDefinitionResource{}
	_ resource.ResourceWithModifyPlan     = &ImageDefinitionResource{}
)

// NewImageDefinitionResource is a helper function to simplify the provider implementation.
func NewImageDefinitionResource() resource.Resource {
	return &ImageDefinitionResource{}
}

// ImageDefinitionResource is the resource implementation.
type ImageDefinitionResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *ImageDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_definition"
}

// Configure adds the provider configured client to the data source.
func (r *ImageDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *ImageDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ImageDefinitionModel{}.GetSchema()
}

// Create creates the resource and sets the initial Terraform state.
func (r *ImageDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ImageDefinitionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Generate API request body from plan
	var createImageDefinitionRequestBody citrixorchestration.CreateImageDefinitionRequestModel
	createImageDefinitionRequestBody.SetName(plan.Name.ValueString())
	createImageDefinitionRequestBody.SetDescription(plan.Description.ValueString())

	osTypeEnum, err := citrixorchestration.NewOsTypeFromValue(plan.OsType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Image Definition"+plan.Name.ValueString(),
			"Unsupported OS Type: "+plan.OsType.ValueString(),
		)
		return
	}
	createImageDefinitionRequestBody.SetOsType(*osTypeEnum)
	sessionSupportEnum, err := citrixorchestration.NewSessionSupportFromValue(plan.SessionSupport.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Image Definition"+plan.Name.ValueString(),
			"Unsupported Session Support: "+plan.SessionSupport.ValueString(),
		)
		return
	}
	createImageDefinitionRequestBody.SetVDASessionSupport(*sessionSupportEnum)

	if r.client.ClientConfig.OrchestrationApiVersion >= 121 {
		// Set Assigned Hypervisor Connection
		assignedHypervisorConnection := buildAssignedHypervisorConnectionRequest(ctx, &resp.Diagnostics, plan, hypervisor)
		createImageDefinitionRequestBody.SetAssignedHypervisorConnection(assignedHypervisorConnection)
	}

	createImageDefinitionRequest := r.client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsCreateImageDefinition(ctx)
	createImageDefinitionRequest = createImageDefinitionRequest.CreateImageDefinitionRequestModel(createImageDefinitionRequestBody)

	if r.client.ClientConfig.OrchestrationApiVersion >= 121 {
		createImageDefinitionRequest = createImageDefinitionRequest.Async(true)
	}

	// Create new Image Definition
	_, httpResp, err := citrixdaasclient.AddRequestData(createImageDefinitionRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Image Definition",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	if r.client.ClientConfig.OrchestrationApiVersion >= 121 {
		err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error creating Image Definition", &resp.Diagnostics, 10, true)
		if err != nil {
			return
		}
	}

	imageDefinitionResp, err := getImageDefinition(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, imageDefinitionResp)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ImageDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ImageDefinitionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed image definition from Orchestration
	imageDefinition, err := readImageDefinition(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, true, imageDefinition)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ImageDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ImageDefinitionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ImageDefinitionModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageDefinitionId := state.Id.ValueString()

	// Construct the update model
	var updateImageDefinitionRequestBody = citrixorchestration.UpdateImageDefinitionRequestModel{}
	updateImageDefinitionRequestBody.SetName(plan.Name.ValueString())
	updateImageDefinitionRequestBody.SetDescription(plan.Description.ValueString())

	// Update Image Definition
	updateImageDefinitionRequest := r.client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsUpdateImageDefinition(ctx, imageDefinitionId)
	updateImageDefinitionRequest = updateImageDefinitionRequest.UpdateImageDefinitionRequestModel(updateImageDefinitionRequestBody)

	imageDefinitionResponse, httpResp, err := citrixdaasclient.AddRequestData(updateImageDefinitionRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Image Definition "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, imageDefinitionResponse)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ImageDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ImageDefinitionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageDefinitionId := state.Id.ValueString()
	imageDefinitionName := state.Name.ValueString()

	imageVersions, err := getImageVersions(ctx, &resp.Diagnostics, r.client, imageDefinitionId)
	if err != nil {
		return
	}

	if len(imageVersions) > 0 {
		resp.Diagnostics.AddError(
			"Error deleting Image Definition "+imageDefinitionName,
			"Image Definition has associated Image Versions. Please delete the Image Versions before deleting the Image Definition.",
		)
		return
	}

	deleteImageDefinitionRequest := r.client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsDeleteImageDefinition(ctx, imageDefinitionId)
	deleteImageDefinitionRequest = deleteImageDefinitionRequest.Async(true)
	httpResp, err := citrixdaasclient.AddRequestData(deleteImageDefinitionRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Image Definition "+imageDefinitionName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Image Definition", &resp.Diagnostics, 10, true)
	if err != nil {
		return
	}
}

func (r *ImageDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ImageDefinitionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ImageDefinitionModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *ImageDefinitionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	util.CheckProductVersion(r.client, &resp.Diagnostics, 121, 118, 7, 41, "Error managing Image Definition resource", "Image Definition resource")

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ImageDefinitionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client.ClientConfig.OrchestrationApiVersion >= 121 {
		// At least one type of hypervisor image definition is required for Orchestration API version 121 and above
		if plan.AzureImageDefinitionModel.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating Image Definition configuration",
				"`azure_image_definition` is required.",
			)
			return
		}
	} else if !plan.AzureImageDefinitionModel.IsNull() && !plan.AzureImageDefinitionModel.IsUnknown() {
		// `azure_image_definition` cannot be specified with Orchestration Service version 120 or earlier
		resp.Diagnostics.AddError(
			"Error validating Image Definition configuration",
			fmt.Sprintf("`azure_image_definition` cannot be specified with the current Orchestration Service version %d.", r.client.ClientConfig.OrchestrationApiVersion),
		)
		return
	}

	if !plan.AzureImageDefinitionModel.IsUnknown() && !plan.AzureImageDefinitionModel.IsNull() {
		azureImageDefinition := util.ObjectValueToTypedObject[AzureImageDefinitionModel](ctx, &resp.Diagnostics, plan.AzureImageDefinitionModel)
		if azureImageDefinition.ResourceGroup.IsNull() && !azureImageDefinition.ImageGalleryName.IsNull() && !azureImageDefinition.ImageGalleryName.IsUnknown() {
			resp.Diagnostics.AddError(
				"Error validating Image Definition configuration",
				"`azure_image_definition` is required when `image_gallery_name` is provided.",
			)
			return
		}

		if !azureImageDefinition.UseImageGallery.IsNull() && !azureImageDefinition.UseImageGallery.IsUnknown() &&
			!azureImageDefinition.ImageGalleryName.IsNull() && !azureImageDefinition.ImageGalleryName.IsNull() {
			if !azureImageDefinition.UseImageGallery.ValueBool() {
				resp.Diagnostics.AddError(
					"Error in `azure_image_definition`",
					"`image_gallery_name` cannot be specified when `use_image_gallery` is set to `false`.",
				)
				return
			}
		}
	}
}

func getImageDefinition(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, imageDefinitionNameOrId string) (*citrixorchestration.ImageDefinitionResponseModel, error) {
	getImageDefinitionRequest := client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsGetImageDefinition(ctx, imageDefinitionNameOrId)
	imageDefinitionResource, httpResp, err := citrixdaasclient.AddRequestData(getImageDefinitionRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error reading Image Definition "+imageDefinitionNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return imageDefinitionResource, nil
}

func readImageDefinition(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, imageDefinitionId string) (*citrixorchestration.ImageDefinitionResponseModel, error) {
	getImageDefinitionRequest := client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsGetImageDefinition(ctx, imageDefinitionId)
	imageDefinitionResource, _, err := util.ReadResource[*citrixorchestration.ImageDefinitionResponseModel](getImageDefinitionRequest, ctx, client, resp, "Image Definition", imageDefinitionId)
	return imageDefinitionResource, err
}

func buildAssignedHypervisorConnectionRequest(ctx context.Context, diagnostics *diag.Diagnostics, plan ImageDefinitionModel, hypervisor *citrixorchestration.HypervisorDetailResponseModel) citrixorchestration.AssignHypervisorConnectionToImageDefinitionRequestModel {
	assignedHypervisorConnection := citrixorchestration.AssignHypervisorConnectionToImageDefinitionRequestModel{}
	assignedHypervisorConnection.SetHypervisorConnection(hypervisor.GetName())
	if !plan.AzureImageDefinitionModel.IsNull() {
		azureImageDefinition := util.ObjectValueToTypedObject[AzureImageDefinitionModel](ctx, diagnostics, plan.AzureImageDefinitionModel)
		azureImageDefinitions := []citrixorchestration.NameValueStringPairModel{}

		if !azureImageDefinition.ResourceGroup.IsNull() {
			customProperty := citrixorchestration.NameValueStringPairModel{}
			customProperty.SetName("ResourceGroups")
			customProperty.SetValue(azureImageDefinition.ResourceGroup.ValueString())
			azureImageDefinitions = append(azureImageDefinitions, customProperty)
		}

		if !azureImageDefinition.UseImageGallery.IsNull() {
			customProperty := citrixorchestration.NameValueStringPairModel{}
			customProperty.SetName("UseSharedImageGallery")
			customProperty.SetValue(fmt.Sprintf("%t", azureImageDefinition.UseImageGallery.ValueBool()))
			azureImageDefinitions = append(azureImageDefinitions, customProperty)
		}

		if !azureImageDefinition.ImageGalleryName.IsNull() {
			customProperty := citrixorchestration.NameValueStringPairModel{}
			customProperty.SetName("ImageGallery")
			customProperty.SetValue(azureImageDefinition.ImageGalleryName.ValueString())
			azureImageDefinitions = append(azureImageDefinitions, customProperty)
		}
		assignedHypervisorConnection.SetCustomProperties(azureImageDefinitions)
	}
	return assignedHypervisorConnection
}

func getImageVersions(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, imageDefinitionId string) ([]citrixorchestration.ImageVersionResponseModel, error) {
	req := client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsGetImageDefinitionImageVersions(ctx, imageDefinitionId)
	req.Limit(250)
	responses := []citrixorchestration.ImageVersionResponseModel{}
	continuationToken := ""
	for {
		req = req.ContinuationToken(continuationToken)
		responseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ImageVersionResponseModelCollection](req, client)
		if err != nil {
			diagnostics.AddError(
				"Error reading Image Versions of Image Definition "+imageDefinitionId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return responses, err
		}
		responses = append(responses, responseModel.GetItems()...)
		if responseModel.GetContinuationToken() == "" {
			return responses, nil
		}
		continuationToken = responseModel.GetContinuationToken()
	}

}
