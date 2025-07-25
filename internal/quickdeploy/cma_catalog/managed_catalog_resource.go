// Copyright © 2024. Citrix Systems, Inc.
package cma_catalog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &citrixManagedCatalogResource{}
	_ resource.ResourceWithConfigure      = &citrixManagedCatalogResource{}
	_ resource.ResourceWithImportState    = &citrixManagedCatalogResource{}
	_ resource.ResourceWithValidateConfig = &citrixManagedCatalogResource{}
	_ resource.ResourceWithModifyPlan     = &citrixManagedCatalogResource{}
)

func NewCitrixManagedCatalogResource() resource.Resource {
	return &citrixManagedCatalogResource{}
}

// citrixManagedCatalogResource is the resource implementation.
type citrixManagedCatalogResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *citrixManagedCatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickdeploy_catalog"
}

// Schema defines the schema for the resource.
func (r *citrixManagedCatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CitrixManagedCatalogResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *citrixManagedCatalogResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create is the implementation of the Create method in the resource.ResourceWithValidateConfig interface.
func (r *citrixManagedCatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan CitrixManagedCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var managedCatalogConfigBody citrixquickdeploy.CitrixManagedCatalogConfigDeployModel

	// Configure managed catalog config model
	var addCatalog citrixquickdeploy.AddCitrixManagedCatalogModel
	addCatalog.SetName(plan.Name.ValueString())
	// Validate region and set region ID
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, plan.Region.ValueString())
	if region == nil {
		// Region not supported, abort creation
		return
	}
	addCatalog.SetRegion(region.GetId())
	addCatalog.SetType(citrixquickdeploy.AddCatalogType(plan.CatalogType.ValueString()))

	// Set static catalog configuration
	addCatalog.SetIsDomainJoined(false)
	addCatalog.SetIsAzureAdJoined(false)
	addCatalog.SetPersistStaticAllocatedVmDisks(true)

	if !plan.MachineNamingScheme.IsNull() {
		var namingScheme citrixquickdeploy.MachineNamingSchemeModel
		machineNamingScheme := util.ObjectValueToTypedObject[MachineNamingSchemeModel](ctx, &resp.Diagnostics, plan.MachineNamingScheme)
		namingScheme.SetNamingScheme(machineNamingScheme.NamingScheme.ValueString())
		namingScheme.SetIsSchemeTypeNumeric(citrixorchestration.AccountNamingSchemeType(machineNamingScheme.NamingSchemeType.ValueString()) == citrixorchestration.ACCOUNTNAMINGSCHEMETYPE_NUMERIC)
		addCatalog.SetMachineNamingScheme(namingScheme)
	}

	managedCatalogConfigBody.SetAddCatalog(addCatalog)

	// Configure managed catalog capacity model
	var catalogCapacity citrixquickdeploy.CatalogCapacitySettingsModel

	// Configure computer worker model
	computerWorker := getComputerWorkerRequestModel(plan)
	catalogCapacity.SetComputeWorker(computerWorker)

	// Configure scale settings model
	catalogScaleSettings := util.ObjectValueToTypedObject[PowerScheduleModel](ctx, &resp.Diagnostics, plan.PowerSchedule)
	scaleSettings := getScaleSettingsRequestModel(ctx, &resp.Diagnostics, catalogScaleSettings)
	// Set max instance to be the same as number of machines
	scaleSettings.SetMaxInstances(int32(plan.NumberOfMachines.ValueInt64()))
	catalogCapacity.SetScaleSettings(scaleSettings)

	// Set static catalog capacity settings
	catalogCapacity.SetSessionTimeout(60)
	catalogCapacity.SetMultiSessionDisconnectedSessionTimeout(15)

	managedCatalogConfigBody.SetAddCatalogCapacity(catalogCapacity)

	// Set subscription ID based on the subscription name
	subscription := util.GetCitrixManagedSubscriptionWithName(ctx, r.client, &resp.Diagnostics, plan.SubscriptionName.ValueString())
	if subscription == nil {
		return
	}
	managedCatalogConfigBody.SetManagedSubscriptionId(subscription.GetSubscriptionId())

	// Set template image
	var templateImageModel citrixquickdeploy.CatalogTemplateImageModel
	templateImage, _, err := util.GetTemplateImageWithId(ctx, r.client, &resp.Diagnostics, plan.TemplateImageId.ValueString(), false)
	if err != nil {
		return
	}
	templateImageModel.SetTemplateId(templateImage.GetId())
	templateImageModel.SetCitrixPrepared(templateImage.CitrixPrepared)
	managedCatalogConfigBody.SetAddCatalogImage(templateImageModel)

	createManagedCatalogRequest := r.client.QuickDeployClient.CatalogCMD.ConfigureAndDeployCitrixManagedCatalogApi(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	createManagedCatalogRequest = createManagedCatalogRequest.Body(managedCatalogConfigBody)

	// Create new Citrix Managed Catalog
	_, httpResp, err := citrixdaasclient.AddRequestData(createManagedCatalogRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Citrix Managed Catalog: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}

	// Get Catalog ID from name
	catalogId := ""
	getManagedCatalogsRequest := r.client.QuickDeployClient.CatalogCMD.GetCustomerManagedCatalogs(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId)
	catalogs, httpResp, err := citrixdaasclient.AddRequestData(getManagedCatalogsRequest, r.client).Execute()
	for _, catalog := range catalogs.GetItems() {
		if catalog.GetName() == plan.Name.ValueString() {
			catalogId = catalog.GetId()
			break
		}
	}
	if catalogId == "" {
		resp.Diagnostics.AddError(
			"Error getting Citrix Managed Catalog ID for catalog: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Catalog not found",
		)
		return
	}

	// Try getting the new Citrix Managed Catalog
	catalog, httpResp, err := waitForCatalogDeployCompletion(ctx, r.client, &resp.Diagnostics, catalogId)
	if err != nil {
		return
	}

	// Verify catalog state
	if catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_INPUT_REQUIRED && catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_ACTIVE {
		resp.Diagnostics.AddError(
			"Error Creating Citrix Managed Catalog: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Catalog state is "+string(catalog.GetState())+
				"\nError details: "+catalog.GetStatusMessage(),
		)
	}

	// Get catalog capacity
	catalogCapacitySettings, err := getManagedCatalogCapacityWithId(ctx, r.client, &resp.Diagnostics, catalog.GetId(), true)
	if err != nil {
		return
	}

	// Get catalog region
	catalogRegion := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, catalog.GetRegion())

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, catalog, catalogCapacitySettings, catalogRegion)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is the implementation of the Read method in the resource.ResourceWithRead interface.
func (r *citrixManagedCatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CitrixManagedCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the Citrix Managed Azure Template Image
	catalog, _, err := getManagedCatalogWithId(ctx, r.client, &resp.Diagnostics, state.Id.ValueString(), true)
	if err != nil {
		// Remove from state
		resp.State.RemoveResource(ctx)
		return
	}
	// Check capacity settings
	catalogCapacitySettings, err := getManagedCatalogCapacityWithId(ctx, r.client, &resp.Diagnostics, catalog.GetId(), true)
	if err != nil {
		return
	}
	// Check image region
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, catalog.GetRegion())

	// Map response body to schema and populate computed attribute values
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, catalog, catalogCapacitySettings, region)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is the implementation of the Update method in the resource.Resource interface.
func (r *citrixManagedCatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan CitrixManagedCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state CitrixManagedCatalogResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := plan.Id.ValueString()
	// Try getting the existing Citrix Managed Catalog
	catalog, httpResp, err := getManagedCatalogWithId(ctx, r.client, &resp.Diagnostics, catalogId, true)
	if err != nil {
		return
	}

	templateImageId := plan.TemplateImageId.ValueString()
	if !strings.EqualFold(catalog.GetImageId(), templateImageId) {
		// If the template image ID has changed, we need to update the catalog image.
		templateImage, _, err := util.GetTemplateImageWithId(ctx, r.client, &resp.Diagnostics, templateImageId, true)
		if err != nil {
			return
		}

		var templateImageUpdateModel citrixquickdeploy.UpdateCatalogTemplateImageModel
		templateImageUpdateModel.SetTemplateId(templateImage.GetId())
		templateImageUpdateModel.SetCitrixPrepared(templateImage.GetCitrixPrepared())

		updateCatalogImageRequest := r.client.QuickDeployClient.CatalogCMD.UpdateCatalogImage(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId, catalogId)
		updateCatalogImageRequest = updateCatalogImageRequest.Body(templateImageUpdateModel)

		// Update Citrix Managed Azure Template Image
		_, httpResp, err = citrixdaasclient.AddRequestData(updateCatalogImageRequest, r.client).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Citrix Managed Catalog Image: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadCatalogServiceClientError(err),
			)
			return
		}

		// Try getting the new Citrix Managed Catalog
		catalog, httpResp, err := waitForCatalogDeployCompletion(ctx, r.client, &resp.Diagnostics, catalogId)
		if err != nil {
			return
		}

		// Verify catalog state
		if catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_INPUT_REQUIRED && catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_ACTIVE {
			resp.Diagnostics.AddError(
				"Error Creating Citrix Managed Catalog: "+plan.Name.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: Catalog state is "+string(catalog.GetState())+
					"\nError details: "+catalog.GetStatusMessage(),
			)
		}
	}

	// Configure managed catalog capacity model
	var catalogCapacity citrixquickdeploy.CatalogCapacitySettingsModel

	// Configure computer worker model
	computerWorker := getComputerWorkerRequestModel(plan)
	catalogCapacity.SetComputeWorker(computerWorker)

	// Configure scale settings model
	catalogScaleSettings := util.ObjectValueToTypedObject[PowerScheduleModel](ctx, &resp.Diagnostics, plan.PowerSchedule)
	scaleSettings := getScaleSettingsRequestModel(ctx, &resp.Diagnostics, catalogScaleSettings)
	// Set pending max instances to be the same as number of machines for update instead of max instances
	scaleSettings.SetPendingMaxInstances(int32(plan.NumberOfMachines.ValueInt64()))
	scaleSettings.SetMaxInstances(int32(state.NumberOfMachines.ValueInt64()))
	catalogCapacity.SetScaleSettings(scaleSettings)

	// Set static catalog capacity settings
	catalogCapacity.SetSessionTimeout(60)
	catalogCapacity.SetMultiSessionDisconnectedSessionTimeout(15)

	updateCatalogCapacityRequest := r.client.QuickDeployClient.CatalogCMD.UpdateCatalogScaleConfiguration(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId, catalogId)
	updateCatalogCapacityRequest = updateCatalogCapacityRequest.Body(catalogCapacity)

	// Update Citrix Managed Catalog Capacity Settings
	httpResp, err = citrixdaasclient.AddRequestData(updateCatalogCapacityRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Citrix Managed Catalog Capacity Settings: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}

	// Wait for the catalog to be updated
	catalog, httpResp, err = waitForCatalogDeployCompletion(ctx, r.client, &resp.Diagnostics, catalogId)
	if err != nil {
		return
	}

	// Verify catalog state
	if catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_INPUT_REQUIRED && catalog.GetState() != citrixquickdeploy.CATALOGOVERALLSTATE_ACTIVE {
		resp.Diagnostics.AddError(
			"Error Creating Citrix Managed Catalog: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: Catalog state is "+string(catalog.GetState())+
				"\nError details: "+catalog.GetStatusMessage(),
		)
	}

	// Try getting the new Citrix Managed Catalog
	catalog, httpResp, err = getManagedCatalogWithId(ctx, r.client, &resp.Diagnostics, catalogId, true)
	if err != nil {
		return
	}
	// Get catalog capacity
	catalogCapacitySettings, err := getManagedCatalogCapacityWithId(ctx, r.client, &resp.Diagnostics, catalog.GetId(), true)
	if err != nil {
		return
	}
	// Check image region
	region := util.GetCmaRegion(ctx, r.client, &resp.Diagnostics, catalog.GetRegion())

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, catalog, catalogCapacitySettings, region)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete is the implementation of the Delete method in the resource.Resource interface.
func (r *citrixManagedCatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CitrixManagedCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var deleteModel citrixquickdeploy.DeleteCatalogModel
	deleteModel.SetDeleteResourceLocationIfUnused(true)

	// Delete Citrix Managed Catalog
	deleteCatalogRequest := r.client.QuickDeployClient.CatalogCMD.DeleteCustomerCatalog(ctx, r.client.ClientConfig.CustomerId, r.client.ClientConfig.SiteId, state.Id.ValueString())
	deleteCatalogRequest = deleteCatalogRequest.Body(deleteModel)
	httpResp, err := citrixdaasclient.AddRequestData(deleteCatalogRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing Citrix Managed Catalog: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return
	}

	// Wait for the catalog to be deleted
	httpResp, err = waitForCatalogDeleteCompletion(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil && httpResp != nil {
		resp.Diagnostics.AddError(
			"Error removing Citrix Managed Catalog: "+state.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for Citrix Managed Catalog deletion: "+state.Name.ValueString(),
			"Error message: "+util.ReadCatalogServiceClientError(err),
		)
	}
}

func (r *citrixManagedCatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *citrixManagedCatalogResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data CitrixManagedCatalogResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate Catalog Type
	if citrixquickdeploy.AddCatalogType(data.CatalogType.ValueString()) != citrixquickdeploy.ADDCATALOGTYPE_MULTI_SESSION {
		// Configure scale settings model
		catalogScaleSettings := util.ObjectValueToTypedObject[PowerScheduleModel](ctx, &resp.Diagnostics, data.PowerSchedule)
		if catalogScaleSettings.PeakMinInstances.ValueInt64() > 0 {
			resp.Diagnostics.AddError(
				"Invalid peak minimum instances",
				fmt.Sprintf("Catalog Type is set to %s, but the Power Schedule has peak minimum instances set to %d. "+
					"Peak minimum instances are only applicable for Multi-Session Catalogs. Please set the catalog type to MultiSession or remove the peak_min_instances from power_schedule.",
					data.CatalogType.ValueString(), catalogScaleSettings.PeakMinInstances.ValueInt64()),
			)
		}

		if catalogScaleSettings.OffPeakMinInstances.ValueInt64() > 0 {
			resp.Diagnostics.AddError(
				"Invalid off peak minimum instances.",
				fmt.Sprintf("Catalog Type is set to %s, but the Power Schedule has off peak minimum instances set to %d. "+
					"Minimum instances are only applicable for Multi-Session Catalogs. Please set the catalog type to MultiSession or remove the off_peak_min_instances from power_schedule.",
					data.CatalogType.ValueString(), catalogScaleSettings.OffPeakMinInstances.ValueInt64()),
			)
		}

		if data.MaxUsersPerVm.ValueInt64() > 1 {
			resp.Diagnostics.AddError(
				"Invalid maximum users per VM.",
				fmt.Sprintf("Catalog Type is set to %s, but the maximum users per VM was set to %d. "+
					"Maximum users per VM are only applicable for Multi-Session Catalogs. Please set the catalog type to MultiSession or remove max_users_per_vm.",
					data.CatalogType.ValueString(), data.MaxUsersPerVm.ValueInt64()),
			)
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *citrixManagedCatalogResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.QuickDeployClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan CitrixManagedCatalogResourceModel
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

func getManagedCatalogWithId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, catalogId string, addWarningIfNotFound bool) (*citrixquickdeploy.CatalogOverview, *http.Response, error) {
	getCatalogRequest := client.QuickDeployClient.CatalogCMD.GetCustomerCatalog(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId, catalogId)
	catalog, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickdeploy.CatalogOverview](getCatalogRequest, client)

	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			if addWarningIfNotFound {
				diagnostics.AddWarning(
					fmt.Sprintf("Managed Catalog with ID: %s not found", catalogId),
					fmt.Sprintf("Managed Catalog with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", catalogId),
				)
			}
			return nil, httpResp, err
		}
		diagnostics.AddError(
			"Error getting Managed Catalog: "+catalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return nil, httpResp, err
	}

	return catalog, httpResp, nil
}

func getManagedCatalogCapacityWithId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, catalogId string, addWarningIfNotFound bool) (*citrixquickdeploy.CatalogCapacitySettingsModel, error) {
	getCatalogCapacityRequest := client.QuickDeployClient.CatalogCMD.GetCatalogCapacityConfiguration(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId, catalogId)
	capacitySettings, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickdeploy.CatalogCapacitySettingsModel](getCatalogCapacityRequest, client)

	if err != nil {
		if addWarningIfNotFound && httpResp.StatusCode == http.StatusNotFound {
			diagnostics.AddWarning(
				fmt.Sprintf("Managed Catalog with ID: %s not found", catalogId),
				fmt.Sprintf("Managed Catalog with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", catalogId),
			)
			return nil, err
		}
		diagnostics.AddError(
			"Error getting Managed Catalog Capacity Settings for catalog: "+catalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadCatalogServiceClientError(err),
		)
		return nil, err
	}

	return capacitySettings, nil
}

func waitForCatalogDeployCompletion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, catalogId string) (*citrixquickdeploy.CatalogOverview, *http.Response, error) {
	// default polling to every 30 seconds
	startTime := time.Now()

	var catalog *citrixquickdeploy.CatalogOverview
	for {
		// Set create timeout to 60 minutes
		if time.Since(startTime) > time.Minute*time.Duration(60) {
			break
		}

		// Sleep ahead of getting the image to account for the time of resource group creation
		time.Sleep(time.Second * time.Duration(30))

		catalog, httpResp, err := getManagedCatalogWithId(ctx, client, diagnostics, catalogId, false)
		if err != nil {
			return nil, httpResp, err
		}

		if catalog.GetState() == citrixquickdeploy.CATALOGOVERALLSTATE_PRE_DEPLOYMENT || catalog.GetState() == citrixquickdeploy.CATALOGOVERALLSTATE_PROCESSING {
			continue
		}

		return catalog, httpResp, err
	}

	return catalog, nil, nil
}

func waitForCatalogDeleteCompletion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, catalogId string) (*http.Response, error) {
	// default polling to every 30 seconds
	startTime := time.Now()

	for {
		// Set deletion timeout to 30 minutes
		if time.Since(startTime) > time.Minute*time.Duration(30) {
			break
		}

		// Sleep ahead of getting the image to account for the time of resource group creation
		time.Sleep(time.Second * time.Duration(30))

		catalog, httpResp, err := getManagedCatalogWithId(ctx, client, diagnostics, catalogId, false)
		if err != nil {
			if httpResp.StatusCode == http.StatusNotFound {
				// If the catalog is not found, it means the deletion is complete
				return httpResp, nil
			}

			return httpResp, err
		}

		if catalog.GetState() == citrixquickdeploy.CATALOGOVERALLSTATE_DELETING {
			continue
		}

		return httpResp, fmt.Errorf("Catalog deletion is no longer in progress for catalog ID: %s, current state: %s", catalogId, catalog.GetState())
	}

	return nil, fmt.Errorf("Timed out waiting for catalog deletion to complete for catalog ID: %s", catalogId)
}

func getScaleSettingsRequestModel(ctx context.Context, diagnostics *diag.Diagnostics, plan PowerScheduleModel) citrixquickdeploy.CatalogScaleSettingsModel {
	var scaleSettings citrixquickdeploy.CatalogScaleSettingsModel
	scaleSettings.SetPeakStartTime(int32(plan.PeakStartTime.ValueInt64()))
	scaleSettings.SetPeakEndTime(int32(plan.PeakEndTime.ValueInt64()))
	scaleSettings.SetPeakTimeZoneId(plan.PeakTimeZoneId.ValueString())
	scaleSettings.SetPeakDisconnectedSessionTimeout(int32(plan.PeakDisconnectedSessionTimeout.ValueInt64()))
	scaleSettings.SetOffPeakDisconnectedSessionTimeout(int32(plan.OffPeakDisconnectedSessionTimeout.ValueInt64()))
	scaleSettings.SetPeakExtendedDisconnectTimeoutMinutes(int32(plan.PeakExtendedDisconnectTimeout.ValueInt64()))
	scaleSettings.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(plan.OffPeakExtendedDisconnectTimeout.ValueInt64()))
	scaleSettings.SetBufferCapacity(int32(plan.PeakBufferCapacity.ValueInt64()))
	scaleSettings.SetOffPeakBufferCapacity(int32(plan.OffPeakBufferCapacity.ValueInt64()))
	scaleSettings.SetPeakMinInstances(int32(plan.PeakMinInstances.ValueInt64()))
	scaleSettings.SetMinInstances(int32(plan.OffPeakMinInstances.ValueInt64()))
	scaleSettings.SetPeakDisconnectedSessionAction(citrixquickdeploy.SessionChangeHostingAction(plan.PeakDisconnectedSessionAction.ValueString()))
	scaleSettings.SetOffPeakDisconnectedSessionAction(citrixquickdeploy.SessionChangeHostingAction(plan.OffPeakDisconnectedSessionAction.ValueString()))

	weekdaysMap := make(map[string]bool)
	weekdays := util.StringSetToStringArray(ctx, diagnostics, plan.Weekdays)
	for _, weekday := range weekdays {
		weekdaysMap[weekday] = true
	}
	scaleSettings.SetWeekdays(weekdaysMap)

	return scaleSettings
}

func getComputerWorkerRequestModel(plan CitrixManagedCatalogResourceModel) citrixquickdeploy.CatalogComputeWorkerModel {
	// Configure computer worker model
	var computerWorker citrixquickdeploy.CatalogComputeWorkerModel
	computerWorker.SetInstanceTypeId(plan.MachineSize.ValueString())
	computerWorker.SetStorageType(citrixquickdeploy.CatalogCapacityStorageType(plan.StorageType.ValueString()))
	computerWorker.SetUseManagedDisks(plan.UseManagedDisks.ValueBool())
	computerWorker.SetMaxUsersPerVM(int32(plan.MaxUsersPerVm.ValueInt64()))

	return computerWorker
}
