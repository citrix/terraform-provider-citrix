// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
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
	_ resource.Resource                   = &ImageVersionResource{}
	_ resource.ResourceWithConfigure      = &ImageVersionResource{}
	_ resource.ResourceWithImportState    = &ImageVersionResource{}
	_ resource.ResourceWithValidateConfig = &ImageVersionResource{}
	_ resource.ResourceWithModifyPlan     = &ImageVersionResource{}
)

// NewImageVersionResource is a helper function to simplify the provider implementation.
func NewImageVersionResource() resource.Resource {
	return &ImageVersionResource{}
}

// ImageDefinitionResource is the resource implementation.
type ImageVersionResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *ImageVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_version"
}

// Configure adds the provider configured client to the data source.
func (r *ImageVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *ImageVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ImageVersionModel{}.GetSchema()
}

// Create implements resource.Resource.
func (r *ImageVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var plan ImageVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := validateImageVersionConfigs(ctx, r.client, &resp.Diagnostics, plan, "creating")
	if err != nil {
		return
	}

	// Construct image scheme
	var imageScheme citrixorchestration.CreateImageSchemeRequestModel
	// Fetch master image path
	var masterImagePath string
	var httpResp *http.Response

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}
	hypervisorResourcePool, err := util.GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, hypervisorId, plan.ResourcePool.ValueString())
	if err != nil {
		return
	}

	// Set NetworkMappings
	if !plan.NetworkMapping.IsNull() {
		networkMappingModel := util.ObjectListToTypedArray[util.NetworkMappingModel](ctx, &resp.Diagnostics, plan.NetworkMapping)
		networkMapping, err := util.ParseNetworkMappingToClientModel(networkMappingModel, hypervisorResourcePool, hypervisor.GetPluginId())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Image Version",
				fmt.Sprintf("Failed to find hypervisor network, error: %s", err.Error()),
			)
			return
		}
		imageScheme.SetNetworkMapping(networkMapping)
	}

	masterImagePath, err = buildImageScheme(ctx, r.client, &resp.Diagnostics, &imageScheme, plan, hypervisor, hypervisorResourcePool)
	if err != nil {
		return
	}

	createImageVersionRequestBody := citrixorchestration.CreateImageVersionRequestModel{}
	createImageVersionRequestBody.SetDescription(plan.Description.ValueString())
	createImageVersionRequestBody.SetResourcePool(plan.ResourcePool.ValueString())
	createImageVersionRequestBody.SetMasterImagePath(masterImagePath)
	createImageVersionRequestBody.SetImageScheme(imageScheme)

	createImageVersionRequest := r.client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsCreateImageVersion(ctx, plan.ImageDefinition.ValueString())
	createImageVersionRequest = createImageVersionRequest.CreateImageVersionRequestModel(createImageVersionRequestBody).Async(true)

	// Create new Image Version
	imageVersion, httpResp, err := citrixdaasclient.AddRequestData(createImageVersionRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Image Version",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		if imageVersion != nil {
			imageDefinition := imageVersion.GetImageDefinition()
			if imageDefinition.GetId() != "" {
				plan.ImageDefinition = types.StringValue(imageDefinition.GetId())
			}

			if imageVersion.GetId() != "" {
				plan.Id = types.StringValue(imageVersion.GetId())
			}

			// Set refreshed state
			diags = resp.State.Set(ctx, &plan)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		return
	}

	timeoutConfigs := util.ObjectValueToTypedObject[ImageVersionTimeout](ctx, &resp.Diagnostics, plan.Timeout)
	createTimeout := timeoutConfigs.Create.ValueInt32()
	if createTimeout == 0 {
		createTimeout = getImageVersionTimeoutConfigs().CreateDefault
	}
	imageVersion, err = util.GetAsyncJobResult[*citrixorchestration.ImageVersionResponseModel](ctx, r.client, httpResp, "Error creating Image Version", &resp.Diagnostics, createTimeout)
	if err != nil {
		return
	}

	imageVersion, err = GetImageVersion(ctx, r.client, &resp.Diagnostics, plan.ImageDefinition.ValueString(), imageVersion.GetId())
	if err != nil {
		return
	}

	if imageVersion.GetImageVersionStatus() != citrixorchestration.IMAGEVERSIONSTATUS_SUCCESS {
		resp.Diagnostics.AddError(
			"Error creating Image Version",
			"Image Version creation finished with status: "+string(imageVersion.GetImageVersionStatus())+
				"\nTransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp),
		)
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, imageVersion)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *ImageVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ImageVersionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed image version from Orchestration
	imageVersion, err := readImageVersion(ctx, r.client, resp, state.ImageDefinition.ValueString(), state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, imageVersion)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *ImageVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var plan ImageVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ImageVersionModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := validateImageVersionConfigs(ctx, r.client, &resp.Diagnostics, plan, "updating")
	if err != nil {
		return
	}

	updateImageVersionRequestModel := citrixorchestration.UpdateImageVersionRequestModel{}
	updateImageVersionRequestModel.SetDescription(plan.Description.ValueString())
	updateImageVersionRequest := r.client.ApiClient.ImageVersionsAPIsDAAS.ImageVersionsUpdateImageVersion(ctx, state.Id.ValueString())
	updateImageVersionRequest = updateImageVersionRequest.UpdateImageVersionRequestModel(updateImageVersionRequestModel)

	imageVersion, httpResp, err := citrixdaasclient.AddRequestData(updateImageVersionRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Image Version "+state.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	imageVersion, err = GetImageVersion(ctx, r.client, &resp.Diagnostics, plan.ImageDefinition.ValueString(), imageVersion.GetId())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, imageVersion)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *ImageVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ImageVersionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageDefinitionId := state.ImageDefinition.ValueString()
	imageVersionId := state.Id.ValueString()

	deleteImageVersionRequest := r.client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsDeleteImageDefinitionImageVersion(ctx, imageDefinitionId, imageVersionId)
	deleteImageVersionRequest = deleteImageVersionRequest.Async(true)
	httpResp, err := citrixdaasclient.AddRequestData(deleteImageVersionRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Image Version "+imageVersionId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	timeoutConfigs := util.ObjectValueToTypedObject[ImageVersionTimeout](ctx, &resp.Diagnostics, state.Timeout)
	deleteTimeout := timeoutConfigs.Delete.ValueInt32()
	if deleteTimeout == 0 {
		deleteTimeout = getImageVersionTimeoutConfigs().DeleteDefault
	}
	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Image Version", &resp.Diagnostics, deleteTimeout)
	if err != nil {
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ImageVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: imageDefinitionId,imageVersionId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image_definition"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (r *ImageVersionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var config ImageVersionModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &config)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *ImageVersionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	util.CheckProductVersion(r.client, &resp.Diagnostics, 121, 118, 7, 41, "Error managing Image Version resource", "Image Version resource")

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ImageVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func readImageVersion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, imageDefinitionId string, imageVersionId string) (*citrixorchestration.ImageVersionResponseModel, error) {
	getImageVersionRequest := client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsGetImageDefinitionImageVersion(ctx, imageDefinitionId, imageVersionId)
	imageVersionResource, _, err := util.ReadResource[*citrixorchestration.ImageVersionResponseModel](getImageVersionRequest, ctx, client, resp, "Image Version", imageVersionId)
	return imageVersionResource, err
}

func GetImageVersion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, imageDefinitionId string, imageVersionId string) (*citrixorchestration.ImageVersionResponseModel, error) {
	getImageVersionRequest := client.ApiClient.ImageDefinitionsAPIsDAAS.ImageDefinitionsGetImageDefinitionImageVersion(ctx, imageDefinitionId, imageVersionId)
	imageVersionResource, httpResp, err := citrixdaasclient.AddRequestData(getImageVersionRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error reading Image Version "+imageVersionId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return imageVersionResource, nil
}

func validateImageVersionConfigs(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan ImageVersionModel, operation string) error {
	imageDefinition, err := GetImageDefinition(ctx, client, diagnostics, plan.ImageDefinition.ValueString())
	if err != nil {
		return err
	}
	if len(imageDefinition.GetHypervisorConnections()) > 0 {
		imgDefHypervisor := imageDefinition.GetHypervisorConnections()[0]
		if !strings.EqualFold(imgDefHypervisor.GetId(), plan.Hypervisor.ValueString()) {
			err := fmt.Errorf("Image Definition is associated with a different hypervisor connection")
			diagnostics.AddError(
				"Error "+operation+" Image Version",
				err.Error(),
			)
			return err
		}
	}

	// Validate image definition
	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, plan.Hypervisor.ValueString())
	if err != nil {
		return err
	}

	switch hypervisor.GetPluginId() {
	case util.AZURERM_FACTORY_NAME:
		// validate azure image specs
		if plan.AzureImageSpecs.IsNull() {
			err := fmt.Errorf("azure_image_specs is required when creating image version with an Azure hypervisor connection")
			diagnostics.AddError(
				"Error "+operation+" Image Version",
				err.Error(),
			)
			return err
		}
		azureImageSpecs := util.ObjectValueToTypedObject[util.AzureImageSpecsModel](ctx, diagnostics, plan.AzureImageSpecs)
		// Validate image version machine profile usage consistency within the image definition
		imageVersionsInDefinition, err := getImageVersions(ctx, diagnostics, client, plan.ImageDefinition.ValueString())
		if err != nil {
			return err
		}

		machineProfileSpecified := !azureImageSpecs.MachineProfile.IsNull()
		err = validateImageVersionMachineProfileConfigs(diagnostics, plan.Id.ValueString(), imageVersionsInDefinition, machineProfileSpecified, operation)
		if err != nil {
			return err
		}
	case util.VMWARE_FACTORY_NAME:
		if plan.VsphereImageSpecs.IsNull() {
			err := fmt.Errorf("vsphere_image_specs is required when creating image version with an vSphere hypervisor connection")
			diagnostics.AddError(
				"Error "+operation+" Image Version",
				err.Error(),
			)
			return err
		}

		vsphereImageSpecs := util.ObjectValueToTypedObject[VsphereImageSpecsModel](ctx, diagnostics, plan.VsphereImageSpecs)
		// Validate image version machine profile usage consistency within the image definition
		imageVersionsInDefinition, err := getImageVersions(ctx, diagnostics, client, plan.ImageDefinition.ValueString())
		if err != nil {
			return err
		}

		if vsphereImageSpecs.MemoryMB.ValueInt32()%4 != 0 {
			err := fmt.Errorf("Attribute `vsphere_image_specs.memory_mb` must be a multiple of 4")
			diagnostics.AddError(
				"Error "+operation+" Image Version",
				err.Error(),
			)
			return err
		}

		machineProfileSpecified := !vsphereImageSpecs.MachineProfile.IsNull()
		err = validateImageVersionMachineProfileConfigs(diagnostics, plan.Id.ValueString(), imageVersionsInDefinition, machineProfileSpecified, operation)
		if err != nil {
			return err
		}
	case util.AMAZON_WORKSPACES_CORE_FACTORY_NAME:
		if plan.AmazonWorkspacesCoreImageSpecs.IsNull() {
			err := fmt.Errorf("amazon_workspaces_core_image_specs is required when creating image version with an Amazon Workspaces Core hypervisor connection")
			diagnostics.AddError(
				"Error "+operation+" Image Version",
				err.Error(),
			)
			return err
		}
	default:
		err := fmt.Errorf("Unsupported hypervisor connection type: %s", hypervisor.GetPluginId())
		diagnostics.AddError(
			"Error "+operation+" Image Version",
			err.Error(),
		)
	}
	return nil
}

func validateImageVersionMachineProfileConfigs(diagnostics *diag.Diagnostics, id string, imageVersionsInDefinition []citrixorchestration.ImageVersionResponseModel, machineProfileSpecified bool, operation string) error {
	for _, imageVersion := range imageVersionsInDefinition {
		for _, spec := range imageVersion.GetImageVersionSpecs() {
			if spec.Context != nil {
				imageContext := spec.GetContext()
				if imageContext.ImageScheme == nil {
					continue
				}
				if (imageContext.MachineProfileMetadata != nil) != machineProfileSpecified &&
					!strings.EqualFold(imageVersion.GetId(), id) {
					err := fmt.Errorf("all image versions within an image definition must consistently use or not use a machine profile")
					diagnostics.AddError(
						"Error "+operation+" Image Version",
						err.Error(),
					)
					return err
				}
			}
		}
	}
	return nil
}
