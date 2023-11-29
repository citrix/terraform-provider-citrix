// Copyright © 2023. Citrix Systems, Inc.

package daas

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"time"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &hypervisorResource{}
	_ resource.ResourceWithConfigure   = &hypervisorResource{}
	_ resource.ResourceWithImportState = &hypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewHypervisorResource() resource.Resource {
	return &hypervisorResource{}
}

// hypervisorResource is the resource implementation.
type hypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *hypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_hypervisor"
}

// Schema defines the schema for the data source.
func (r *hypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
				Required:    true,
			},
			"connection_type": schema.StringAttribute{
				Description: "Connection Type of the hypervisor (AzureRM, AWS, GCP).",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedHypervisorConnectionTypeEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the hypervisor is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "**[Azure: Required]** Application ID of the service principal used to access the Azure APIs.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_secret"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("subscription_id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("active_directory_id"),
					}...),
				},
			},
			"application_secret": schema.StringAttribute{
				Description: "**[Azure: Required]** The Application Secret of the service principal used to access the Azure APIs.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("subscription_id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("active_directory_id"),
					}...),
				},
			},
			"application_secret_expiration_date": schema.StringAttribute{
				Description: "**[Azure: Optional]** The expiration date of the application secret of the service principal used to access the Azure APIs. Format is YYYY-MM-DD.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^((?:19|20|21)\d\d)[-](0[1-9]|1[012])[-](0[1-9]|[12][0-9]|3[01])$`), "Ensure date is valid and is in the format YYYY-MM-DD"),
				},
			},
			"subscription_id": schema.StringAttribute{
				Description: "**[Azure: Required]** Azure Subscription ID.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_secret"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("active_directory_id"),
					}...),
				},
			},
			"active_directory_id": schema.StringAttribute{
				Description: "**[Azure: Required]** Azure Active Directory ID.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_id"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("application_secret"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("subscription_id"),
					}...),
				},
			},
			"aws_region": schema.StringAttribute{
				Description: "**[AWS: Required]** AWS region to connect to.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("secret_key"),
					}...),
				},
			},
			"api_key": schema.StringAttribute{
				Description: "**[AWS: Required]** The API key used to authenticate with the AWS APIs.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("aws_region"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("secret_key"),
					}...),
				},
			},
			"secret_key": schema.StringAttribute{
				Description: "**[AWS: Required]** The secret key used to authenticate with the AWS APIs.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("aws_region"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("api_key"),
					}...),
				},
			},
			"service_account_id": schema.StringAttribute{
				Description: "**[GCP: Required]** The service account ID used to access the Google Cloud APIs.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_credentials"),
					}...),
				},
			},
			"service_account_credentials": schema.StringAttribute{
				Description: "**[GCP: Required]** The JSON-encoded service account credentials used to access the Google Cloud APIs.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_id"),
					}...),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *hypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *hypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan models.HypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionType, err := citrixorchestration.NewHypervisorConnectionTypeFromValue(plan.ConnectionType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}
	connectionDetails.SetConnectionType(*connectionType)

	switch *connectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		if plan.ApplicationId.IsNull() || plan.ApplicationSecret.IsNull() || plan.SubscriptionId.IsNull() || plan.ActiveDirectoryId.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor for AzureRM",
				"ApplicationId/ApplicationSecret/SubscriptionId/ActiveDirectoryId is missing.",
			)
			return
		}
		connectionDetails.SetApplicationId(plan.ApplicationId.ValueString())
		connectionDetails.SetApplicationSecret(plan.ApplicationSecret.ValueString())
		metadata := getMetadataForAzureRmHypervisor(plan)
		connectionDetails.SetMetadata(metadata)
		connectionDetails.SetSubscriptionId(plan.SubscriptionId.ValueString())
		connectionDetails.SetActiveDirectoryId(plan.ActiveDirectoryId.ValueString())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if plan.AwsRegion.IsNull() || plan.ApiKey.IsNull() || plan.SecretKey.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor for AWS",
				"ApiKey/SecretKey is missing.",
			)
			return
		}
		connectionDetails.SetRegion(plan.AwsRegion.ValueString())
		connectionDetails.SetApiKey(plan.ApiKey.ValueString())
		connectionDetails.SetSecretKey(plan.SecretKey.ValueString())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if plan.ServiceAccountId.IsNull() || plan.ServiceAccountCredentials.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor for GCP",
				"ServiceAccountId/ServiceAccountCredential is missing.",
			)
			return
		}
		connectionDetails.SetServiceAccountId(plan.ServiceAccountId.ValueString())
		connectionDetails.SetServiceAccountCredentials(plan.ServiceAccountCredentials.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Error creating Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	// Generate post body for resolving resource paths
	var detail citrixorchestration.HypervisorConnectionDetailRequestModel
	detail.SetName(connectionDetails.GetName())
	detail.SetZone(connectionDetails.GetZone())
	detail.SetConnectionType(connectionDetails.GetConnectionType())

	switch connectionDetails.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		detail.SetApplicationId(connectionDetails.GetApplicationId())
		detail.SetApplicationSecret(connectionDetails.GetApplicationSecret())
		detail.SetSubscriptionId(connectionDetails.GetSubscriptionId())
		detail.SetActiveDirectoryId(connectionDetails.GetActiveDirectoryId())
		detail.SetApplicationSecretFormat(citrixorchestration.IDENTITYPASSWORDFORMAT_PLAIN_TEXT)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		detail.SetRegion(connectionDetails.GetRegion())
		detail.SetApiKey(connectionDetails.GetApiKey())
		detail.SetSecretKey(connectionDetails.GetSecretKey())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		detail.SetServiceAccountId(connectionDetails.GetServiceAccountId())
		detail.SetServiceAccountCredentials(connectionDetails.GetServiceAccountCredentials())
	default:
		resp.Diagnostics.AddError(
			"Error creating hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	createHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsCreateHypervisor(ctx)
	createHypervisorRequest = createHypervisorRequest.CreateHypervisorRequestModel(body).Async(true)

	// Create new hypervisor
	_, httpResp, err := citrixdaasclient.AddRequestData(createHypervisorRequest, r.client).Execute()
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor",
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	jobId := citrixdaasclient.GetJobIdFromHttpResponse(*httpResp)
	jobResponseModel, err := r.client.WaitForJob(ctx, jobId, 10)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor",
			"TransactionId: "+txId+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorDetail := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		resp.Diagnostics.AddError(
			"Error creating Hypervisor",
			errorDetail,
		)
	}

	hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(hypervisor)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *hypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state models.HypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := state.Id.ValueString()
	hypervisor, err := readHypervisor(ctx, r.client, resp, hypervisorId)
	if err != nil {
		return
	}

	// Overwrite hypervisor with refreshed state
	state = state.RefreshPropertyValues(hypervisor)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *hypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan models.HypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisorName := plan.Name.ValueString()
	hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	connectionType, err := citrixorchestration.NewHypervisorConnectionTypeFromValue(plan.ConnectionType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor "+hypervisorName,
			"Unsupported hypervisor connection type.",
		)
		return
	}
	editHypervisorRequestBody.SetConnectionType(*connectionType)
	switch hypervisor.GetConnectionType() {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		editHypervisorRequestBody.SetApplicationId(plan.ApplicationId.ValueString())
		editHypervisorRequestBody.SetApplicationSecret(plan.ApplicationSecret.ValueString())
		metadata := getMetadataForAzureRmHypervisor(plan)
		editHypervisorRequestBody.SetMetadata(metadata)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		editHypervisorRequestBody.SetApiKey(plan.ApiKey.ValueString())
		editHypervisorRequestBody.SetSecretKey(plan.SecretKey.ValueString())
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		editHypervisorRequestBody.SetServiceAccountId(plan.ServiceAccountId.ValueString())
		editHypervisorRequestBody.SetServiceAccountCredential(plan.ServiceAccountCredentials.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Error updating Hypervisor "+hypervisorName,
			"Unsupported hypervisor connection type.",
		)
		return
	}

	// Patch hypervisor
	patchHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsPatchHypervisor(ctx, hypervisorId)
	patchHypervisorRequest = patchHypervisorRequest.EditHypervisorConnectionRequestModel(editHypervisorRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(patchHypervisorRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated hypervisor from GetHypervisor
	updatedHypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(updatedHypervisor)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *hypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state models.HypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing hypervisor
	hypervisorId := state.Id.ValueString()
	hypervisorName := state.Name.ValueString()
	deleteHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisor(ctx, hypervisorId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *hypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Gets the hypervisor and logs any errors
func GetHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId string) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	// Resolve resource path for service offering and master image
	getHypervisorReq := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorId)
	hypervisor, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.HypervisorDetailResponseModel](getHypervisorReq, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return hypervisor, err
}

func readHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, hypervisorId string) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	getHypervisorReq := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorId)
	hypervisor, _, err := util.ReadResource[*citrixorchestration.HypervisorDetailResponseModel](getHypervisorReq, ctx, client, resp, "Hypervisor", hypervisorId)
	return hypervisor, err
}

func getMetadataForAzureRmHypervisor(plan models.HypervisorResourceModel) []citrixorchestration.NameValueStringPairModel {
	secretExpirationDate := "2099-12-31 23:59:59"
	if !plan.ApplicationSecretExpirationDate.IsNull() {
		secretExpirationDate = plan.ApplicationSecretExpirationDate.ValueString()
		secretExpirationDate = secretExpirationDate + " 23:59:59"
	}

	parsedTime, _ := time.Parse(time.DateTime, secretExpirationDate)
	secretExpirationDateInUnix := parsedTime.UnixMilli()
	secretExpirationDateMetada := citrixorchestration.NameValueStringPairModel{}
	secretExpirationDateMetada.SetName("Citrix_Orchestration_Hypervisor_Secret_Expiration_Date")
	secretExpirationDateMetada.SetValue(strconv.Itoa(int(secretExpirationDateInUnix)))
	metadata := []citrixorchestration.NameValueStringPairModel{
		secretExpirationDateMetada,
	}

	return metadata
}
