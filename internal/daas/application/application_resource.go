// Copyright © 2023. Citrix Systems, Inc.

package application

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &applicationResource{}
	_ resource.ResourceWithConfigure   = &applicationResource{}
	_ resource.ResourceWithImportState = &applicationResource{}
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
	resp.Schema = schema.Schema{
		Description: "Resource for creating and managing applications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the application.",
				Required:    true,
			},
			"published_name": schema.StringAttribute{
				Description: "A display name for the application that is shown to users.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the application.",
				Optional:    true,
			},
			"installed_app_properties": schema.SingleNestedAttribute{
				Description: "The install application properties.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"command_line_arguments": schema.StringAttribute{
						Description: "The command-line arguments to use when launching the executable.",
						Optional:    true,
						Validators: []validator.String{
							validator.String(
								stringvalidator.LengthAtLeast(1),
							),
						},
					},
					"command_line_executable": schema.StringAttribute{
						Description: "The path of the executable file to launch.",
						Required:    true,
					},
					"working_directory": schema.StringAttribute{
						Description: "The working directory which the executable is launched from.",
						Optional:    true,
						Validators: []validator.String{
							validator.String(
								stringvalidator.LengthAtLeast(1),
							),
						},
					},
				},
			},
			"delivery_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The delivery group id's to which the application should be added.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"application_folder_path": schema.StringAttribute{
				Description: "The application folder path in which the application should be created.",
				Optional:    true,
			},
		},
	}
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
	createInstalledAppRequest.SetCommandLineArguments(plan.InstalledAppProperties.CommandLineArguments.ValueString())
	createInstalledAppRequest.SetCommandLineExecutable(plan.InstalledAppProperties.CommandLineExecutable.ValueString())
	createInstalledAppRequest.SetWorkingDirectory(plan.InstalledAppProperties.WorkingDirectory.ValueString())

	var createApplicationRequest citrixorchestration.CreateApplicationRequestModel
	createApplicationRequest.SetName(plan.Name.ValueString())
	createApplicationRequest.SetDescription(plan.Description.ValueString())
	createApplicationRequest.SetPublishedName(plan.PublishedName.ValueString())
	createApplicationRequest.SetInstalledAppProperties(createInstalledAppRequest)
	createApplicationRequest.SetApplicationFolder(plan.ApplicationFolderPath.ValueString())

	var newApplicationRequest []citrixorchestration.CreateApplicationRequestModel
	newApplicationRequest = append(newApplicationRequest, createApplicationRequest)

	var deliveryGroups []citrixorchestration.PriorityRefRequestModel
	for _, value := range plan.DeliveryGroups {
		var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
		deliveryGroupRequestModel.SetItem(value.ValueString())
		deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
	}

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

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(application)

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

	state = state.RefreshPropertyValues(application)

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

	// Get refreshed application properties from Orchestration
	applicationId := plan.Id.ValueString()
	applicationName := plan.Name.ValueString()

	_, err := getApplication(ctx, r.client, &resp.Diagnostics, applicationId)
	if err != nil {
		return
	}

	// Construct the update model
	var editApplicationRequestBody = &citrixorchestration.EditApplicationRequestModel{}
	editApplicationRequestBody.SetName(plan.Name.ValueString())
	editApplicationRequestBody.SetDescription(plan.Description.ValueString())
	editApplicationRequestBody.SetPublishedName(plan.PublishedName.ValueString())
	editApplicationRequestBody.SetApplicationFolder(plan.ApplicationFolderPath.ValueString())

	var editInstalledAppRequest citrixorchestration.EditInstalledAppRequestModel
	editInstalledAppRequest.SetCommandLineArguments(plan.InstalledAppProperties.CommandLineArguments.ValueString())
	editInstalledAppRequest.SetCommandLineExecutable(plan.InstalledAppProperties.CommandLineExecutable.ValueString())
	editInstalledAppRequest.SetWorkingDirectory(plan.InstalledAppProperties.WorkingDirectory.ValueString())

	editApplicationRequestBody.SetInstalledAppProperties(editInstalledAppRequest)

	var deliveryGroups []citrixorchestration.PriorityRefRequestModel
	for _, value := range plan.DeliveryGroups {
		var deliveryGroupRequestModel citrixorchestration.PriorityRefRequestModel
		deliveryGroupRequestModel.SetItem(value.ValueString())
		deliveryGroups = append(deliveryGroups, deliveryGroupRequestModel)
	}

	editApplicationRequestBody.SetDeliveryGroups(deliveryGroups)

	folderPathExists := checkIfApplicationFolderPathExist(ctx, r.client, &resp.Diagnostics, plan.ApplicationFolderPath.ValueString())
	if !folderPathExists {
		return
	}

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

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(application)

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
