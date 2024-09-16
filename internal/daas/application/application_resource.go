// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"net/http"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &applicationResource{}
	_ resource.ResourceWithConfigure      = &applicationResource{}
	_ resource.ResourceWithImportState    = &applicationResource{}
	_ resource.ResourceWithValidateConfig = &applicationResource{}
	_ resource.ResourceWithModifyPlan     = &applicationResource{}
)

// NewApplicationResource is a helper function to simplify the provider implementation.
func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

// applicationResource is the resource implementation.
type applicationResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Configure adds the provider configured client to the data source.
func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ApplicationResourceModel{}.GetSchema()
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var createInstalledAppRequest citrixorchestration.CreateInstalledAppRequestModel
	var installedAppProperties = util.ObjectValueToTypedObject[InstalledAppResponseModel](ctx, &resp.Diagnostics, plan.InstalledAppProperties)
	createInstalledAppRequest.SetCommandLineArguments(installedAppProperties.CommandLineArguments.ValueString())
	createInstalledAppRequest.SetCommandLineExecutable(installedAppProperties.CommandLineExecutable.ValueString())
	createInstalledAppRequest.SetWorkingDirectory(installedAppProperties.WorkingDirectory.ValueString())

	var createApplicationRequest citrixorchestration.CreateApplicationRequestModel
	createApplicationRequest.SetName(plan.Name.ValueString())
	createApplicationRequest.SetDescription(plan.Description.ValueString())
	createApplicationRequest.SetPublishedName(plan.PublishedName.ValueString())
	createApplicationRequest.SetInstalledAppProperties(createInstalledAppRequest)
	createApplicationRequest.SetApplicationFolder(plan.ApplicationFolderPath.ValueString())
	createApplicationRequest.SetIcon(plan.Icon.ValueString())
	createApplicationRequest.SetClientFolder(plan.ApplicationCategoryPath.ValueString())

	if plan.LimitVisibilityToUsers.IsNull() {
		createApplicationRequest.SetIncludedUserFilterEnabled(false)
		createApplicationRequest.SetIncludedUsers([]string{})
	} else {
		limitVisibilityToUsers := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.LimitVisibilityToUsers)
		limitVisibilityToUserIds, httpResponse, err := util.GetUserIdsUsingIdentity(ctx, r.client, limitVisibilityToUsers)
		if err != nil {
			diags.AddError(
				"Error fetching user details for application resource",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResponse)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
		createApplicationRequest.SetIncludedUsers(limitVisibilityToUserIds)
		createApplicationRequest.SetIncludedUserFilterEnabled(true)
	}

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	createApplicationRequest.SetMetadata(metadata)

	var newApplicationRequest []citrixorchestration.CreateApplicationRequestModel
	newApplicationRequest = append(newApplicationRequest, createApplicationRequest)

	deliveryGroups := buildDeliveryGroupsPriorityRequestModel(ctx, &resp.Diagnostics, plan)

	var body citrixorchestration.AddApplicationsRequestModel
	body.SetNewApplications(newApplicationRequest)
	body.SetDeliveryGroups(deliveryGroups)

	addApplicationsRequest := r.client.ApiClient.ApplicationsAPIsDAAS.ApplicationsAddApplications(ctx)
	addApplicationsRequest = addApplicationsRequest.AddApplicationsRequestModel(body)

	folderPathExists := checkIfApplicationFolderPathExist(ctx, r.client, &resp.Diagnostics, plan.ApplicationFolderPath.ValueString())
	if !folderPathExists {
		return
	}

	// Create new application
	httpResp, err := citrixdaasclient.AddRequestData(addApplicationsRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Application",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new application with application name

	applicationName := plan.Name.ValueString()

	// If the application is present in an application folder, we specify the name in this format: {application folder path plus application name}.For example, FolderName1|FolderName2|ApplicationName.
	if plan.ApplicationFolderPath.ValueString() != "" {
		applicationName = strings.ReplaceAll(plan.ApplicationFolderPath.ValueString(), "\\", "|") + applicationName
	}

	application, err := getApplication(ctx, r.client, &resp.Diagnostics, applicationName)
	if err != nil {
		return
	}

	applicationDeliveryGroups, err := getApplicationDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationName)
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, application, applicationDeliveryGroups)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed application properties from Orchestration
	application, err := readApplication(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	applicationDeliveryGroups, err := getApplicationDeliveryGroups(ctx, r.client, &resp.Diagnostics, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, application, applicationDeliveryGroups)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan ApplicationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationId := plan.Id.ValueString()
	applicationName := plan.Name.ValueString()

	// Construct the update model
	var editApplicationRequestBody = &citrixorchestration.EditApplicationRequestModel{}
	editApplicationRequestBody.SetName(plan.Name.ValueString())
	editApplicationRequestBody.SetDescription(plan.Description.ValueString())
	editApplicationRequestBody.SetPublishedName(plan.PublishedName.ValueString())
	editApplicationRequestBody.SetApplicationFolder(plan.ApplicationFolderPath.ValueString())
	editApplicationRequestBody.SetIcon(plan.Icon.ValueString())
	editApplicationRequestBody.SetClientFolder(plan.ApplicationCategoryPath.ValueString())

	if plan.LimitVisibilityToUsers.IsNull() {
		editApplicationRequestBody.SetIncludedUserFilterEnabled(false)
		editApplicationRequestBody.SetIncludedUsers([]string{})
	} else {
		limitVisibilityToUsers := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.LimitVisibilityToUsers)
		limitVisibilityToUserIds, httpResponse, err := util.GetUserIdsUsingIdentity(ctx, r.client, limitVisibilityToUsers)
		if err != nil {
			diags.AddError(
				"Error fetching user details for application resource",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResponse)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
		editApplicationRequestBody.SetIncludedUsers(limitVisibilityToUserIds)
		editApplicationRequestBody.SetIncludedUserFilterEnabled(true)
	}
	var editInstalledAppRequest citrixorchestration.EditInstalledAppRequestModel
	var installedAppProperties = util.ObjectValueToTypedObject[InstalledAppResponseModel](ctx, &resp.Diagnostics, plan.InstalledAppProperties)
	editInstalledAppRequest.SetCommandLineArguments(installedAppProperties.CommandLineArguments.ValueString())
	editInstalledAppRequest.SetCommandLineExecutable(installedAppProperties.CommandLineExecutable.ValueString())
	editInstalledAppRequest.SetWorkingDirectory(installedAppProperties.WorkingDirectory.ValueString())

	editApplicationRequestBody.SetInstalledAppProperties(editInstalledAppRequest)

	deliveryGroups := buildDeliveryGroupsPriorityRequestModel(ctx, &resp.Diagnostics, plan)
	editApplicationRequestBody.SetDeliveryGroups(deliveryGroups)

	folderPathExists := checkIfApplicationFolderPathExist(ctx, r.client, &resp.Diagnostics, plan.ApplicationFolderPath.ValueString())
	if !folderPathExists {
		return
	}

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editApplicationRequestBody.SetMetadata(metadata)

	// Update Application
	editApplicationRequest := r.client.ApiClient.ApplicationsAPIsDAAS.ApplicationsPatchApplication(ctx, applicationId)
	editApplicationRequest = editApplicationRequest.EditApplicationRequestModel(*editApplicationRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(editApplicationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Application "+applicationName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Get updated application from GetApplication
	application, err := getApplication(ctx, r.client, &resp.Diagnostics, applicationId)
	if err != nil {
		return
	}

	applicationDeliveryGroups, err := getApplicationDeliveryGroups(ctx, r.client, &resp.Diagnostics, applicationId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, application, applicationDeliveryGroups)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state ApplicationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing delivery group
	applicationId := state.Id.ValueString()
	applicationName := state.Name.ValueString()
	deleteApplicationRequest := r.client.ApiClient.ApplicationsAPIsDAAS.ApplicationsDeleteApplication(ctx, applicationId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteApplicationRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Application "+applicationName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *applicationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ApplicationResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

	validateDeliveryGroupsPriority(ctx, &resp.Diagnostics, data)
}

func (r *applicationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func readApplication(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, applicationId string) (*citrixorchestration.ApplicationDetailResponseModel, error) {
	getApplicationRequest := client.ApiClient.ApplicationsAPIsDAAS.ApplicationsGetApplication(ctx, applicationId)
	applicationResource, _, err := util.ReadResource[*citrixorchestration.ApplicationDetailResponseModel](getApplicationRequest, ctx, client, resp, "Application", applicationId)
	return applicationResource, err
}

func getApplication(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, applicationId string) (*citrixorchestration.ApplicationDetailResponseModel, error) {
	getApplicationRequest := client.ApiClient.ApplicationsAPIsDAAS.ApplicationsGetApplication(ctx, applicationId)
	application, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ApplicationDetailResponseModel](getApplicationRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Application "+applicationId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return application, err
}

func getApplicationDeliveryGroups(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, applicationNameOrId string) (*citrixorchestration.ApplicationDeliveryGroupResponseModelCollection, error) {
	getApplicationDeliveryGroupsRequest := client.ApiClient.ApplicationsAPIsDAAS.ApplicationsGetApplicationDeliveryGroups(ctx, applicationNameOrId)
	applicationDeliveryGroups, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ApplicationDeliveryGroupResponseModelCollection](getApplicationDeliveryGroupsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Delivery Groups associated with Application "+applicationNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return applicationDeliveryGroups, err
}

// checkIfApplicationFolderPathExist checks if the application folder path exists.
func checkIfApplicationFolderPathExist(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, applicationFolderPath string) bool {
	if applicationFolderPath == "" {
		return true
	}

	tempFolderPath := strings.ReplaceAll(applicationFolderPath, "\\", "|")
	appFolderExistRequest := client.ApiClient.ApplicationFoldersAPIsDAAS.ApplicationFoldersCheckApplicationFolderPathExists(ctx, tempFolderPath)
	httpResp, err := citrixdaasclient.AddRequestData(appFolderExistRequest, client).Execute()
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			diagnostics.AddError("Application Folder Path \""+applicationFolderPath+"\" does not exist. Create the folder first and then create the application in it", "")
		} else {
			diagnostics.AddError(
				"Error while checking Application Folder Path "+applicationFolderPath,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}
		return false
	}
	return true
}

func buildDeliveryGroupsPriorityRequestModel(ctx context.Context, diagnostics *diag.Diagnostics, plan ApplicationResourceModel) []citrixorchestration.PriorityRefRequestModel {
	var deliveryGroups []citrixorchestration.PriorityRefRequestModel
	if !plan.DeliveryGroups.IsNull() {
		for index, deliveryGroup := range util.StringListToStringArray(ctx, diagnostics, plan.DeliveryGroups) {
			var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
			deliveryGroupRequestModel.SetItem(deliveryGroup)
			deliveryGroupRequestModel.SetPriority(int32(index))
			deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
		}
	} else if !plan.DeliveryGroupsPriority.IsNull() {
		for _, deliveryGroup := range util.ObjectSetToTypedArray[DeliveryGroupPriorityModel](ctx, diagnostics, plan.DeliveryGroupsPriority) {
			var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
			deliveryGroupRequestModel.SetItem(deliveryGroup.Id.ValueString())
			deliveryGroupRequestModel.SetPriority(deliveryGroup.Priority.ValueInt32())
			deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
		}
	}
	return deliveryGroups
}

func validateDeliveryGroupsPriority(ctx context.Context, diagnostics *diag.Diagnostics, data ApplicationResourceModel) {
	// Make sure the delivery_groups_priority does not have duplicated priority values
	if !data.DeliveryGroupsPriority.IsNull() && !data.DeliveryGroupsPriority.IsUnknown() {
		dgPriority := util.ObjectSetToTypedArray[DeliveryGroupPriorityModel](ctx, diagnostics, data.DeliveryGroupsPriority)
		priorityMap := map[int32]bool{}
		deliveryGroupMap := map[string]bool{}
		isValid := true
		for _, dg := range dgPriority {
			if !dg.Priority.IsNull() {
				if priorityMap[dg.Priority.ValueInt32()] {
					isValid = false
				} else {
					priorityMap[dg.Priority.ValueInt32()] = true
				}
			}

			if !dg.Id.IsNull() {
				if deliveryGroupMap[dg.Id.ValueString()] {
					isValid = false
				} else {
					deliveryGroupMap[dg.Id.ValueString()] = true
				}
			}

			if !isValid {
				diagnostics.AddError(
					"Invalid configuration in delivery_groups_priority",
					"The value of priority and delivery group id should be unique in delivery_groups_priority",
				)
				return
			}
		}
	}
}
