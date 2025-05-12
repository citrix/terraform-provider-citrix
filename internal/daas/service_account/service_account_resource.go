// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

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
	_ resource.Resource                   = &serviceAccountResource{}
	_ resource.ResourceWithConfigure      = &serviceAccountResource{}
	_ resource.ResourceWithImportState    = &serviceAccountResource{}
	_ resource.ResourceWithValidateConfig = &serviceAccountResource{}
	_ resource.ResourceWithModifyPlan     = &serviceAccountResource{}
)

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

type serviceAccountResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (*serviceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *serviceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *serviceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan ServiceAccountModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorMessage := fmt.Sprintf("Error creating Service Account %s", plan.DisplayName.ValueString())

	serviceAccountRequestModel := citrixorchestration.CreateServiceAccountRequestModel{}
	serviceAccountRequestModel.SetDisplayName(plan.DisplayName.ValueString())
	serviceAccountRequestModel.SetDescription(plan.Description.ValueString())
	serviceAccountRequestModel.SetIdentityProviderType(plan.IdentityProviderType.ValueString())
	serviceAccountRequestModel.SetIdentityProviderIdentifier(plan.IdentityProviderIdentifier.ValueString())
	serviceAccountRequestModel.SetAccountId(plan.AccountId.ValueString())
	serviceAccountRequestModel.SetAccountSecret(plan.AccountSecret.ValueString())

	secretFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.AccountSecretFormat.ValueString())
	if err != nil || secretFormat == nil {
		resp.Diagnostics.AddError(
			"Error creating Service Account",
			"Unsupported password format: "+plan.AccountSecretFormat.ValueString(),
		)
		return
	}

	serviceAccountRequestModel.SetAccountSecretFormat(*secretFormat)

	if !plan.SecretExpiryTime.IsNull() {
		expiryDateTime, err := time.Parse(time.DateOnly, plan.SecretExpiryTime.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				errorMessage,
				"Error parsing secret expiry time: "+plan.SecretExpiryTime.ValueString(),
			)
			return
		}
		serviceAccountRequestModel.SetSecretExpiryTime(expiryDateTime.Format(time.RFC3339))
	}

	serviceAccountRequestModel.SetCapabilities(getCapabilitiesForServiceAccount(plan))
	serviceAccountRequestModel.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))

	createServiceAccountRequest := r.client.ApiClient.IdentityAPIsDAAS.IdentityCreateServiceAccount(ctx)
	serviceAccountResponse := &citrixorchestration.ServiceAccountResponseModel{}
	httpResp := &http.Response{}

	// Check if the Cloud DDC version is supported for async operation
	isDdcVersionSupported := r.client.ClientConfig.OrchestrationApiVersion >= util.DDCVersionToCreateServiceAccountWithAsync && !r.client.AuthConfig.OnPremises
	if isDdcVersionSupported {
		createServiceAccountRequest = createServiceAccountRequest.CreateServiceAccountRequestModel(serviceAccountRequestModel).Async(true)
		_, httpResp, err = citrixdaasclient.AddRequestData(createServiceAccountRequest, r.client).Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				errorMessage,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		// Process the async job response
		err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, errorMessage, &diags, 10)
		if err != nil {
			return
		}

		// Fetch the service account using the account ID
		serviceAccountResponse, err = GetServiceAccountUsingAccountId(ctx, r.client, &resp.Diagnostics, plan.AccountId.ValueString())
		if err != nil {
			return
		}

	} else {
		createServiceAccountRequest = createServiceAccountRequest.CreateServiceAccountRequestModel(serviceAccountRequestModel)
		serviceAccountResponse, httpResp, err = citrixdaasclient.AddRequestData(createServiceAccountRequest, r.client).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				errorMessage,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, serviceAccountResponse)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *serviceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state ServiceAccountModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountResponseModel, err := readServiceAccount(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, serviceAccountResponseModel)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *serviceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan ServiceAccountModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateServiceAccountRequestModel := citrixorchestration.UpdateServiceAccountRequestModel{}

	updateServiceAccountRequestModel.SetDisplayName(plan.DisplayName.ValueString())
	updateServiceAccountRequestModel.SetDescription(plan.Description.ValueString())
	updateServiceAccountRequestModel.SetAccountId(plan.AccountId.ValueString())
	updateServiceAccountRequestModel.SetAccountSecret(plan.AccountSecret.ValueString())

	secretFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.AccountSecretFormat.ValueString())
	if err != nil || secretFormat == nil {
		resp.Diagnostics.AddError(
			"Error creating Service Account",
			"Unsupported password format: "+plan.AccountSecretFormat.ValueString(),
		)
		return
	}

	updateServiceAccountRequestModel.SetAccountSecretFormat(*secretFormat)

	if !plan.SecretExpiryTime.IsNull() {
		expiryDateTime, err := time.Parse(time.DateOnly, plan.SecretExpiryTime.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Service Account",
				"Error parsing secret expiry time: "+plan.SecretExpiryTime.ValueString(),
			)
			return
		}
		updateServiceAccountRequestModel.SetSecretExpiryTime(expiryDateTime.Format(time.RFC3339))
	} else {
		updateServiceAccountRequestModel.SetSecretExpiryTime("")
	}

	updateServiceAccountRequestModel.SetCapabilities(getCapabilitiesForServiceAccount(plan))
	updateServiceAccountRequestModel.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))

	updateServiceAccountRequest := r.client.ApiClient.IdentityAPIsDAAS.IdentitySetServiceAccount(ctx, plan.Id.ValueString())
	updateServiceAccountRequest = updateServiceAccountRequest.UpdateServiceAccountRequestModel(updateServiceAccountRequestModel)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateServiceAccountRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Service Account "+plan.DisplayName.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	serviceAccountResponse, err := getServiceAccount(ctx, r.client, &resp.Diagnostics, plan.Id.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, serviceAccountResponse)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete implements resource.Resource.
func (r *serviceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state ServiceAccountModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.Id.ValueString()
	serviceAccountName := state.DisplayName.ValueString()
	deleteServiceAccountRequest := r.client.ApiClient.IdentityAPIsDAAS.IdentityDeleteServiceAccount(ctx, serviceAccountId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteServiceAccountRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Service Account "+serviceAccountName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *serviceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema implements resource.Resource.
func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ServiceAccountModel{}.GetSchema()
}

func (r *serviceAccountResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data ServiceAccountModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.IdentityProviderType.ValueString() == string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY) && !data.EnableIntuneEnrolledDeviceManagement.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("enable_intune_enrolled_device_management"),
			"Incorrect Attribute Configuration",
			"`enable_intune_enrolled_device_management` can only be set when identity_provider_type is `AzureAD`",
		)
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *serviceAccountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan ServiceAccountModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Scopes.IsUnknown() && !plan.Scopes.IsNull() {
		scopesResponse, httpResp, err := util.FetchScopes(ctx, r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching scopes",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		scopeIds := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes)
		for _, scope := range scopesResponse.GetItems() {
			if slices.ContainsFunc(scopeIds, func(scopeId string) bool {
				return strings.EqualFold(scope.GetId(), scopeId)
			}) {
				// Check if the scope provided in the plan is a built-in scope adn throw an error if it is
				if scope.GetIsBuiltIn() {
					resp.Diagnostics.AddError(
						"Error applying scopes",
						fmt.Sprintf("Scope %s is a built-in scope and cannot be applied to a service account", scope.GetId()),
					)
					return
				}
			}
		}
	}
}

func readServiceAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, serviceAccountId string) (*citrixorchestration.ServiceAccountResponseModel, error) {
	getServiceAccountRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetServiceAccount(ctx, serviceAccountId)
	servcieAccountResponse, _, err := util.ReadResource[*citrixorchestration.ServiceAccountResponseModel](getServiceAccountRequest, ctx, client, resp, "Service Account", serviceAccountId)
	return servcieAccountResponse, err
}

func getServiceAccount(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, serviceAccountId string) (*citrixorchestration.ServiceAccountResponseModel, error) {
	getServiceAccountRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetServiceAccount(ctx, serviceAccountId)
	serviceAccountResponseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ServiceAccountResponseModel](getServiceAccountRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Service Account: "+serviceAccountId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return serviceAccountResponseModel, err
}

func getCapabilitiesForServiceAccount(plan ServiceAccountModel) []string {
	if plan.IdentityProviderType.ValueString() == string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY) {
		return nil
	}

	capabilities := []string{}
	capabilities = append(capabilities, util.ServiceAccountAzureADDeviceManagementCapability)
	if plan.EnableIntuneEnrolledDeviceManagement.ValueBool() {
		capabilities = append(capabilities, util.ServiceAccountIntuneEnrolledDeviceManagementCapability)
	}

	return capabilities
}
