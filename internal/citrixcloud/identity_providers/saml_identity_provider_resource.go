// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &SamlIdentityProviderResource{}
	_ resource.ResourceWithConfigure      = &SamlIdentityProviderResource{}
	_ resource.ResourceWithImportState    = &SamlIdentityProviderResource{}
	_ resource.ResourceWithValidateConfig = &SamlIdentityProviderResource{}
	_ resource.ResourceWithModifyPlan     = &SamlIdentityProviderResource{}
)

// SamlIdentityProviderResource is a helper function to simplify the provider implementation.
func NewSamlIdentityProviderResource() resource.Resource {
	return &SamlIdentityProviderResource{}
}

// SamlIdentityProviderResource is the resource implementation.
type SamlIdentityProviderResource struct {
	client  *citrixdaasclient.CitrixDaasClient
	idpType string
}

// Metadata returns the resource type name.
func (r *SamlIdentityProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_saml_identity_provider"
}

// Schema defines the schema for the resource.
func (r *SamlIdentityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SamlIdentityProviderModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *SamlIdentityProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	r.idpType = string(citrixcws.CWSIDENTITYPROVIDERTYPE_SAML)
}

// Create creates the resource and sets the initial Terraform state.
func (r *SamlIdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan SamlIdentityProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check auth domain name availability
	err := checkAuthDomainNameAvailability(ctx, r.client, &resp.Diagnostics, plan.AuthDomainName.ValueString())
	if err != nil {
		return
	}

	// Open and read the SAML certificate file
	certFile, err := os.Open(plan.CertFilePath.ValueString())
	if err != nil {
		if os.IsPermission(err) {
			resp.Diagnostics.AddError(
				"Error reading SAML Certificate",
				"Permission denied to read SAML Certificate: "+plan.CertFilePath.ValueString()+
					"\nError message: "+err.Error(),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error open certificate for Saml Identity Provider with path "+plan.CertFilePath.ValueString(),
			"Error message: "+err.Error(),
		)
		return
	}
	// Add defer to close the certFile after the function completes or errors out
	defer certFile.Close()
	certFileName := filepath.Base(plan.CertFilePath.ValueString())

	// Validate Saml Certificate
	samlCertValidationRequest := r.client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersValidateSamlPost(ctx, r.client.ClientConfig.CustomerId)
	samlCertValidationRequest = samlCertValidationRequest.FileName(certFileName)
	samlCertValidationRequest = samlCertValidationRequest.CertFile(certFile)
	samlCertInfo, httpResp, err := citrixdaasclient.AddRequestData(samlCertValidationRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error validating certificate with path %s for Saml Identity Provider %s", plan.CertFilePath.ValueString(), plan.Name.ValueString()),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Create Identity Provider
	idpStatus, err := createIdentityProvider(ctx, &resp.Diagnostics, r.client, r.idpType, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Refresh state with created Identity Provider before Identity Provider configuration
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, idpStatus, nil)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := idpStatus.GetIdpInstanceId()

	// Configure Identity Provider
	// Configure Identity Provider Body
	var samlIdpConnectBody citrixcws.SamlConnectionModel
	samlIdpConnectBody.SetSamlCertFileName(certFileName)
	samlIdpConnectBody.SetSamlCertInfo(*samlCertInfo)
	samlIdpConnectBody.SetSamlAuthDomainName(plan.AuthDomainName.ValueString())

	samlIdpConnectBody.SetSamlEntityId(plan.EntityId.ValueString())
	samlIdpConnectBody.SetSamlSignAuthRequest(plan.SignAuthRequest.ValueString())
	samlIdpConnectBody.SetSamlAuthenticationContext(plan.AuthenticationContext.ValueString())
	samlIdpConnectBody.SetSamlAuthenticationContextComparison(plan.AuthenticationContextComparison.ValueString())
	samlIdpConnectBody.SetSamlSingleSignOnServiceUrl(plan.SingleSignOnServiceUrl.ValueString())
	samlIdpConnectBody.SetSamlSingleSignOnServiceBinding(plan.SingleSignOnServiceBinding.ValueString())
	samlIdpConnectBody.SetSamlResponse(plan.SamlResponse.ValueString())
	samlIdpConnectBody.SetSamlLogoutUrl(plan.LogoutUrl.ValueString())

	if plan.LogoutUrl.ValueString() != "" {
		samlIdpConnectBody.SetSamlSignLogoutRequest(plan.SignLogoutRequest.ValueString())
		samlIdpConnectBody.SetSamlLogoutRequestBinding(plan.LogoutBinding.ValueString())
	}

	if plan.UseScopedEntityId.ValueBool() {
		samlIdpConnectBody.SetSamlSpEntityIdSuffix(uuid.NewString())
	}

	samlAttributeNames := util.ObjectValueToTypedObject[SamlAttributeNameMappings](ctx, &resp.Diagnostics, plan.AttributeNames)
	samlIdpConnectBody.SetSamlAttributeNameForUserDisplayName(samlAttributeNames.UserDisplayName.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForUserGivenName(samlAttributeNames.UserGivenName.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForUserFamilyName(samlAttributeNames.UserFamilyName.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForSid(samlAttributeNames.SecurityIdentifier.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForUpn(samlAttributeNames.UserPrincipalName.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForEmail(samlAttributeNames.Email.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForAdOid(samlAttributeNames.AdObjectIdentifier.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForAdForest(samlAttributeNames.AdForest.ValueString())
	samlIdpConnectBody.SetSamlAttributeNameForAdDomain(samlAttributeNames.AdDomain.ValueString())

	var idpConnectBody citrixcws.IdpInstanceConnectModel
	idpConnectBody.SetIdentityProviderType(r.idpType)
	idpConnectBody.SetAuthDomainName(plan.AuthDomainName.ValueString())
	idpConnectBody.SetSamlConnectionModel(samlIdpConnectBody)

	idpStatus, err = createIdentityProviderConnection(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, idpConnectBody)
	if err != nil {
		return
	}

	// Get SAML Configuration
	samlConfig, err := getSamlConfiguration(ctx, resp.Diagnostics, r.client, idpInstanceId)
	if err != nil {
		return
	}

	// Refresh plan
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, true, idpStatus, samlConfig)

	// Set state with fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *SamlIdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state SamlIdentityProviderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := state.Id.ValueString()
	// Read Saml Identity Provider
	idpStatus, err := readIdentityProvider(ctx, r.client, resp, r.idpType, idpInstanceId)
	if err != nil {
		return
	}

	// Read SAML Configuration
	samlConfig, err := getSamlConfiguration(ctx, resp.Diagnostics, r.client, idpInstanceId)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, true, idpStatus, samlConfig)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SamlIdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan SamlIdentityProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state SamlIdentityProviderModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := state.Id.ValueString()

	// Update Idp Nickname
	updateIdentityProviderNickname(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, plan.Name.ValueString())

	// Update Idp Auth Domain
	if !strings.EqualFold(plan.AuthDomainName.ValueString(), state.AuthDomainName.ValueString()) {
		// Check auth domain name availability
		err := checkAuthDomainNameAvailability(ctx, r.client, &resp.Diagnostics, plan.AuthDomainName.ValueString())
		if err == nil {
			updateIdentityProviderAuthDomain(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, state.AuthDomainName.ValueString(), plan.AuthDomainName.ValueString())
		}
	}

	// Get the updated Saml Identity Provider
	idpStatus, err := getIdentityProviderById(ctx, r.client, &resp.Diagnostics, r.idpType, idpInstanceId)
	if err != nil {
		return
	}

	// Read SAML Configuration
	samlConfig, err := getSamlConfiguration(ctx, resp.Diagnostics, r.client, idpInstanceId)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, true, idpStatus, samlConfig)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SamlIdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state SamlIdentityProviderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteIdentityProvider(ctx, &resp.Diagnostics, r.client, r.idpType, state.Id.ValueString())
}

func (r *SamlIdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SamlIdentityProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data SamlIdentityProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Resource Location is a cloud concept which is not supported for on-prem environment
func (r *SamlIdentityProviderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ResourceLocationsClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing SAML Identity Provider resource", "SAML Identity Provider resource is only supported for Cloud customers.")
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan SamlIdentityProviderModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func getSamlConfiguration(ctx context.Context, diagnostics diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpInstanceId string) (*citrixcws.SamlConfigModel, error) {
	samlConfigRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersConfigurationSamlIdGet(ctx, idpInstanceId, client.ClientConfig.CustomerId)
	samlConfig, httpResp, err := citrixdaasclient.AddRequestData(samlConfigRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error getting SAML Identity Provider configuration id "+idpInstanceId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return samlConfig, nil
}
