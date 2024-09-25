// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &GoogleIdentityProviderResource{}
	_ resource.ResourceWithConfigure      = &GoogleIdentityProviderResource{}
	_ resource.ResourceWithImportState    = &GoogleIdentityProviderResource{}
	_ resource.ResourceWithValidateConfig = &GoogleIdentityProviderResource{}
	_ resource.ResourceWithModifyPlan     = &GoogleIdentityProviderResource{}
)

// NewGoogleIdentityProviderResource is a helper function to simplify the provider implementation.
func NewGoogleIdentityProviderResource() resource.Resource {
	return &GoogleIdentityProviderResource{}
}

// GoogleIdentityProviderResource is the resource implementation.
type GoogleIdentityProviderResource struct {
	client  *citrixdaasclient.CitrixDaasClient
	idpType string
}

// Metadata returns the resource type name.
func (r *GoogleIdentityProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_google_identity_provider"
}

// Schema defines the schema for the resource.
func (r *GoogleIdentityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GoogleIdentityProviderResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *GoogleIdentityProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	r.idpType = string(citrixcws.CWSIDENTITYPROVIDERTYPE_GOOGLE)
}

// Create creates the resource and sets the initial Terraform state.
func (r *GoogleIdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan GoogleIdentityProviderResourceModel
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

	idpInstanceId := idpStatus.GetIdpInstanceId()

	// Refresh state with created Identity Provider before Identity Provider configuration
	plan = plan.RefreshIdAndNameValues(idpStatus)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure Identity Provider
	// Validate Identity Provider Credentials
	var googleIdpConnectBody citrixcws.GoogleConnectionModel
	googleIdpConnectBody.SetGoogleClientEmail(plan.ClientEmail.ValueString())
	googleIdpConnectBody.SetGooglePrivateKey(plan.PrivateKey.ValueString())
	googleIdpConnectBody.SetGoogleImpersonatedUser(plan.ImpersonatedUser.ValueString())

	idpValidationRequest := r.client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersConfigureGooglePost(ctx, r.client.ClientConfig.CustomerId)
	idpValidationRequest = idpValidationRequest.GoogleConnectionModel(googleIdpConnectBody)
	idpValidationResult, httpResp, err := citrixdaasclient.AddRequestData(idpValidationRequest, r.client).Execute()
	if err != nil || !idpValidationResult.GetValidCredentials() {
		resp.Diagnostics.AddError(
			"Error validating credentials for Google Identity Provider "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Configure Identity Provider Body
	var idpConnectBody citrixcws.IdpInstanceConnectModel
	idpConnectBody.SetGoogleConnectionModel(googleIdpConnectBody)
	idpConnectBody.SetIdentityProviderType(r.idpType)
	idpConnectBody.SetAuthDomainName(plan.AuthDomainName.ValueString())

	idpStatus, err = createIdentityProviderConnection(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, idpConnectBody)
	if err != nil {
		return
	}

	// Refresh plan
	plan = plan.RefreshPropertyValues(idpStatus)

	// Set state with fully populated data
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *GoogleIdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state GoogleIdentityProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Google Identity Provider
	idpStatus, err := readIdentityProvider(ctx, r.client, resp, r.idpType, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(idpStatus)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *GoogleIdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan GoogleIdentityProviderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state GoogleIdentityProviderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idpInstanceId := plan.Id.ValueString()

	// Update Idp Nickname
	updateIdentityProviderNickname(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, plan.Name.ValueString())

	// Update Idp Auth Domain
	updateIdentityProviderAuthDomain(ctx, &resp.Diagnostics, r.client, r.idpType, idpInstanceId, state.AuthDomainName.ValueString(), plan.AuthDomainName.ValueString())

	// Get the updated Google Identity Provider
	idpStatus, err := getIdentityProviderById(ctx, r.client, &resp.Diagnostics, r.idpType, idpInstanceId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(idpStatus)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *GoogleIdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state GoogleIdentityProviderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteIdentityProvider(ctx, &resp.Diagnostics, r.client, r.idpType, state.Id.ValueString())
}

func (r *GoogleIdentityProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *GoogleIdentityProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data GoogleIdentityProviderResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Resource Location is a cloud concept which is not supported for on-prem environment
func (r *GoogleIdentityProviderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ResourceLocationsClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing Google Identity Provider resource", "Google Identity Provider resource is only supported for Cloud customers.")
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan GoogleIdentityProviderResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
