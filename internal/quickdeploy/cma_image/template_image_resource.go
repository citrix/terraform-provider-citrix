// Copyright © 2024. Citrix Systems, Inc.
package cma_image

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &citrixManagedAzureImageResource{}
	_ resource.ResourceWithConfigure      = &citrixManagedAzureImageResource{}
	_ resource.ResourceWithImportState    = &citrixManagedAzureImageResource{}
	_ resource.ResourceWithValidateConfig = &citrixManagedAzureImageResource{}
	_ resource.ResourceWithModifyPlan     = &citrixManagedAzureImageResource{}
)

func NewCitrixManagedAzureImageResource() resource.Resource {
	return &citrixManagedAzureImageResource{}
}

// citrixManagedAzureImageResource is the resource implementation.
type citrixManagedAzureImageResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *citrixManagedAzureImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickdeploy_template_image"
}

// Schema defines the schema for the resource.
func (r *citrixManagedAzureImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CitrixManagedAzureImageResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *citrixManagedAzureImageResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create is the implementation of the Create method in the resource.ResourceWithValidateConfig interface.
func (r *citrixManagedAzureImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan CitrixManagedAzureImageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var importImageBody citrixquickdeploy.ImportTemplateImageModel
	importImageBody.SetName(plan.Name.ValueString())
	importImageBody.SetNotes(plan.Notes.ValueString())
	importImageBody.SetVhdUri(plan.VhdUri.ValueString())
	// Validate region and set region ID
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, plan.Region.ValueString())
	if region == nil {
		// Region not supported, abort creation
		return
	}
	importImageBody.SetRegion(region.GetId())
	importImageBody.SetHyperVGen(plan.MachineGeneration.ValueString())
	osPlatform, err := citrixquickdeploy.NewSupportedOperatingSystemTypeFromValue(plan.OsPlatform.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing Template Image: "+plan.Name.ValueString(),
			"Unsupported OS Platform: "+plan.OsPlatform.ValueString(),
		)

		return
	}

	importImageBody.SetOsPlatform(*osPlatform)
	importImageBody.SetVtpmEnabled(plan.VtpmEnabled.ValueBool())
	importImageBody.SetSecureBootEnabled(plan.SecureBootEnabled.ValueBool())
	if !plan.GuestDiskUri.IsNull() {
		importImageBody.SetVhdEncryptionUri(plan.GuestDiskUri.ValueString())
	}

	// Set subscription ID based on the subscription name
	subscription := util.GetCitrixManagedSubscriptionWithName(ctx, r.client, &resp.Diagnostics, plan.SubscriptionName.ValueString())
	if subscription == nil {
		return
	}
	importImageBody.SetAzureSubscriptionId(subscription.GetSubscriptionId())

	importTemplateImageRequest := r.client.QuickDeployClient.MasterImageCMD.ImportTemplateImage(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	importTemplateImageRequest = importTemplateImageRequest.Body(importImageBody)

	// Import new Citrix Managed Azure Template Image
	importImageResponse, httpResp, err := citrixdaasclient.AddRequestData(importTemplateImageRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing Template Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}

	// Try getting the new Citrix Managed Azure Template Image
	image, httpResp, err := waitForImageImportCompletion(ctx, r.client, &resp.Diagnostics, importImageResponse)
	if err != nil {
		return
	}

	// Verify image state
	if image.GetState() != citrixquickdeploy.TEMPLATEIMAGESTATE_READY {
		resp.Diagnostics.AddError(
			"Error importing Template Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Image state is "+string(image.GetState())+
				"\nError details: "+image.GetStatus(),
		)
	}
	// Check image region
	region = util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, image.GetRegion())

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, image, region)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is the implementation of the Read method in the resource.ResourceWithRead interface.
func (r *citrixManagedAzureImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CitrixManagedAzureImageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the Citrix Managed Azure Template Image
	image, _, err := util.GetTemplateImageWithId(ctx, r.client, &resp.Diagnostics, state.Id.ValueString(), true)
	if err != nil {
		// Remove from state
		resp.State.RemoveResource(ctx)
		return
	}
	// Check image region
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, image.GetRegion())

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, image, region)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is the implementation of the Update method in the resource.Resource interface.
func (r *citrixManagedAzureImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan CitrixManagedAzureImageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageId := plan.Id.ValueString()

	// Generate API request body from plan
	var templateImageUpdateBody citrixquickdeploy.UpdateTemplateImageModel
	templateImageUpdateBody.SetNewName(plan.Name.ValueString())
	templateImageUpdateBody.SetNewNotes(plan.Notes.ValueString())

	updateTemplateImageRequest := r.client.QuickDeployClient.MasterImageCMD.UpdateTemplateImage(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId, imageId)
	updateTemplateImageRequest = updateTemplateImageRequest.Body(templateImageUpdateBody)

	// Update Citrix Managed Azure Template Image
	httpResp, err := citrixdaasclient.AddRequestData(updateTemplateImageRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Template Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}

	// Try getting the updated Citrix Managed Azure Template Image
	image, httpResp, err := util.GetTemplateImageWithId(ctx, r.client, &resp.Diagnostics, imageId, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Template Image: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}
	// Check image region
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, image.GetRegion())

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, image, region)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete is the implementation of the Delete method in the resource.Resource interface.
func (r *citrixManagedAzureImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CitrixManagedAzureImageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Citrix Managed Azure Template Image
	deleteImageRequest := r.client.QuickDeployClient.MasterImageCMD.DeleteTemplateImage(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteImageRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing Template Image: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}
}

func (r *citrixManagedAzureImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Error importing Template Image",
		"Import operation is not supported for Template Image. To use an existing Template Image, use data source instead.",
	)
}

func (r *citrixManagedAzureImageResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data CitrixManagedAzureImageResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.VtpmEnabled.ValueBool() && strings.EqualFold(data.MachineGeneration.ValueString(), util.HypervGen1) {
		resp.Diagnostics.AddError(
			"Error validating Template Image configuration",
			"vTPM is only supported for V2 generation images",
		)
	}

	if data.SecureBootEnabled.ValueBool() {
		if strings.EqualFold(data.MachineGeneration.ValueString(), util.HypervGen1) {
			resp.Diagnostics.AddError(
				"Error validating Template Image configuration",
				"Secure Boot is only supported for V2 generation images",
			)
		}
		if !data.VtpmEnabled.ValueBool() {
			resp.Diagnostics.AddError(
				"Error validating Template Image configuration",
				"vTPM must be enabled when Secure Boot is enabled",
			)
		}
		if data.GuestDiskUri.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating Template Image configuration",
				"Guest Disk URI must be specified when Secure Boot is enabled",
			)
		}
	}
	if !data.GuestDiskUri.IsNull() && !data.SecureBootEnabled.ValueBool() {
		resp.Diagnostics.AddError(
			"Error validating Template Image configuration",
			"Guest Disk URI is only applicable when Secure Boot is enabled",
		)
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *citrixManagedAzureImageResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickDeployClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan CitrixManagedAzureImageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate region
	if !plan.Region.IsUnknown() && r.client != nil {
		util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, plan.Region.ValueString())
	}

	// Validate subscription
	if !plan.SubscriptionName.IsUnknown() && r.client != nil {
		util.GetCitrixManagedSubscriptionWithName(ctx, r.client, &resp.Diagnostics, plan.SubscriptionName.ValueString())
	}
}

func waitForImageImportCompletion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, imageResp *citrixquickdeploy.TemplateImageOverview) (*citrixquickdeploy.TemplateImageDetails, *http.Response, error) {
	// default polling to every 30 seconds
	startTime := time.Now()
	imageId := imageResp.GetId()

	var image *citrixquickdeploy.TemplateImageDetails
	for {
		if time.Since(startTime) > time.Minute*time.Duration(120) {
			break
		}

		// Sleep ahead of getting the image to account for the time of resource group creation
		time.Sleep(time.Second * time.Duration(30))

		image, httpResp, err := util.GetTemplateImageWithId(ctx, client, diagnostics, imageId, false)
		if err != nil {
			return nil, httpResp, err
		}

		if image.GetState() == citrixquickdeploy.TEMPLATEIMAGESTATE_PENDING || image.GetState() == citrixquickdeploy.TEMPLATEIMAGESTATE_IMPORTING {
			continue
		}

		return image, httpResp, err
	}

	return image, nil, nil
}
