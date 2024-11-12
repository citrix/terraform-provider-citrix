// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &OktaIdentityProviderResource{}
	_ resource.ResourceWithConfigure      = &OktaIdentityProviderResource{}
	_ resource.ResourceWithImportState    = &OktaIdentityProviderResource{}
	_ resource.ResourceWithValidateConfig = &OktaIdentityProviderResource{}
	_ resource.ResourceWithModifyPlan     = &OktaIdentityProviderResource{}
)

// NewOktaIdentityProviderResource is a helper function to simplify the provider implementation.
func NewOktaIdentityProviderResource() resource.Resource {
	return &OktaIdentityProviderResource{}
}

// OktaIdentityProviderResource is the resource implementation.
type OktaIdentityProviderResource struct {
	client  *citrixdaasclient.CitrixDaasClient
	idpType string
}

// Metadata returns the resource type name.
func (r *OktaIdentityProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_okta_identity_provider"
}

// Schema defines the schema for the resource.
func (r *OktaIdentityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = OktaIdentityProviderModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *OktaIdentityProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	r.idpType = string(citrixcws.CWSIDENTITYPROVIDERTYPE_OKTA)
}

// Create creates the resource and sets the initial Terraform state.
func (r *OktaIdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan OktaIdentityProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create Identity Provider
	idpStatus, err := createIdentityProvider(ctx, &resp.Diagnostics, r.client, r.idpType, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Refresh state with created Identity Provider before Identity Provider configuration
	plan = plan.RefreshPropertyValues(true, idpStatus)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := idpStatus.GetIdpInstanceId()

	// Configure Identity Provider
	// Validate Identity Provider Credentials
	var oktaIdpConnectBody citrixcws.OktaConnectionModel
	oktaIdpConnectBody.SetOktaDomain(plan.OktaDomain.ValueString())
	oktaIdpConnectBody.SetOktaClientId(plan.OktaClientId.ValueString())
	oktaIdpConnectBody.SetOktaClientSecret(plan.OktaClientSecret.ValueString())
	oktaIdpConnectBody.SetOktaApiToken(plan.OktaApiToken.ValueString())

	idpValidationRequest := r.client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersConfigureOktaPost(ctx, r.client.ClientConfig.CustomerId)
	idpValidationRequest = idpValidationRequest.OktaConnectionModel(oktaIdpConnectBody)
	idpValidationResult, httpResp, err := citrixdaasclient.AddRequestData(idpValidationRequest, r.client).Execute()
	if !isOktaConfigurationValid(&resp.Diagnostics, plan.Name.ValueString(), idpValidationResult, httpResp, err) {
		return
	}

	// Configure Identity Provider Body
	var idpConnectBody citrixcws.IdpInstanceConnectModel
	idpConnectBody.SetIdentityProviderType(r.idpType)
	idpConnectBody.SetOktaConnectionModel(oktaIdpConnectBody)

	idpStatus, err = createIdentityProviderConnection(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, idpConnectBody)
	if err != nil {
		return
	}

	// Refresh plan
	plan = plan.RefreshPropertyValues(true, idpStatus)

	// Set state with fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *OktaIdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state OktaIdentityProviderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Okta Identity Provider
	idpStatus, err := readIdentityProvider(ctx, r.client, resp, r.idpType, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(true, idpStatus)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *OktaIdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan OktaIdentityProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := plan.Id.ValueString()

	// Update Idp Nickname
	updateIdentityProviderNickname(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, plan.Name.ValueString())

	// Get the updated Okta Identity Provider
	idpStatus, err := getIdentityProviderById(ctx, r.client, &resp.Diagnostics, r.idpType, idpInstanceId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(true, idpStatus)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *OktaIdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state OktaIdentityProviderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteIdentityProvider(ctx, &resp.Diagnostics, r.client, r.idpType, state.Id.ValueString())
}

func (r *OktaIdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *OktaIdentityProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data OktaIdentityProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Resource Location is a cloud concept which is not supported for on-prem environment
func (r *OktaIdentityProviderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ResourceLocationsClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing Okta Identity Provider resource", "Okta Identity Provider resource is only supported for Cloud customers.")
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan OktaIdentityProviderModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func isOktaConfigurationValid(diagnostics *diag.Diagnostics, oktaIdentityProviderName string, idpValidationResult *citrixcws.OktaResultModel, httpResp *http.Response, err error) bool {
	isOktaConfigurationValid := true

	errorTitle := fmt.Sprintf("Error validating credentials for Okta Identity Provider %s", oktaIdentityProviderName)
	errorTransactionId := "TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	errorMessageFormat := "\nError message: Okta %s is not valid"

	if err != nil {
		diagnostics.AddError(
			errorTitle,
			errorTransactionId+"\nError message: "+util.ReadClientError(err),
		)
		isOktaConfigurationValid = false
	}

	if !idpValidationResult.GetValidDomain() {
		diagnostics.AddError(
			errorTitle,
			errorTransactionId+fmt.Sprintf(errorMessageFormat, "Domain"),
		)
		isOktaConfigurationValid = false
	}

	if !idpValidationResult.GetValidClientId() {
		diagnostics.AddError(
			errorTitle,
			errorTransactionId+fmt.Sprintf(errorMessageFormat, "Client ID"),
		)
		isOktaConfigurationValid = false
	}

	if !idpValidationResult.GetValidClientSecret() {
		diagnostics.AddError(
			errorTitle,
			errorTransactionId+fmt.Sprintf(errorMessageFormat, "Client Secret"),
		)
		isOktaConfigurationValid = false
	}

	if !idpValidationResult.GetValidApiToken() {
		diagnostics.AddError(
			errorTitle,
			errorTransactionId+fmt.Sprintf(errorMessageFormat, "API Token"),
		)
		isOktaConfigurationValid = false
	}
	return isOktaConfigurationValid
}
