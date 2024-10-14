// Copyright © 2024. Citrix Systems, Inc.

package cc_admin_user

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	ccadmins "github.com/citrix/citrix-daas-rest-go/ccadmins"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &ccAdminUserResource{}
	_ resource.ResourceWithConfigure      = &ccAdminUserResource{}
	_ resource.ResourceWithImportState    = &ccAdminUserResource{}
	_ resource.ResourceWithValidateConfig = &ccAdminUserResource{}
	_ resource.ResourceWithModifyPlan     = &ccAdminUserResource{}
)

// NewAdminUserResource is a helper function to simplify the provider implementation.
func NewCCAdminUserResource() resource.Resource {
	return &ccAdminUserResource{}
}

// adminUserResource is the resource implementation.
type ccAdminUserResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *ccAdminUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_admin_user"
}

// Schema defines the schema for the resource.
func (r *ccAdminUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CCAdminUserResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *ccAdminUserResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *ccAdminUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan CCAdminUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body ccadmins.CreateAdministratorInputModel
	body.SetType(plan.Type.ValueString())
	adminAccessType, err := getAdminAccessType(plan.AccessType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid access type",
			"Error message: "+err.Error())
		return
	}
	body.SetAccessType(adminAccessType)

	if !plan.Email.IsNull() {
		body.SetEmail(plan.Email.ValueString())
	}

	adminProviderType, err := getAdminProviderType(plan.ProviderType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting the provider type",
			"Error message: "+err.Error())
		return
	}
	body.SetProviderType(adminProviderType)

	if !plan.FirstName.IsNull() {
		body.SetFirstName(plan.FirstName.ValueString())
	}
	if !plan.LastName.IsNull() {
		body.SetLastName(plan.LastName.ValueString())
	}
	if !plan.DisplayName.IsNull() {
		body.SetDisplayName(plan.DisplayName.ValueString())
	}
	if !plan.ExternalProviderId.IsNull() {
		body.SetExternalProviderId(plan.ExternalProviderId.ValueString())
	}
	if !plan.ExternalUserId.IsNull() {
		body.SetExternalUserId(plan.ExternalUserId.ValueString())
	}

	// Add policies to the admin user
	accessPolicy, err := getAdminUserPolicies(ctx, &resp.Diagnostics, r.client, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error adding policies to the user", "Error message: "+err.Error())
		return
	}
	body.SetPolicies(accessPolicy)

	createAdminUserRequest := r.client.CCAdminsClient.AdministratorsAPI.CreateAdministrator(ctx)
	createAdminUserRequest = createAdminUserRequest.CitrixCustomerId(r.client.ClientConfig.CustomerId).CreateAdministratorInputModel(body)

	// Create a new admin user
	_, httpResp, err := citrixdaasclient.AddRequestData(createAdminUserRequest, r.client).Execute()

	//In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating admin "+plan.Email.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	plan, err = fetchAndUpdateAdminUser(ctx, r.client, plan, &resp.Diagnostics)
	if err != nil {
		return // Error already added to diagnostics
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ccAdminUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state CCAdminUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the admin user from remote
	var adminUser ccadmins.AdministratorResult
	var err error

	adminUser, err = getAdminUser(ctx, r.client, state)
	if err != nil {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("Admin user with id: %s not found", state.AdminId.ValueString()),
			fmt.Sprintf("Admin user: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.AdminId.ValueString()),
		)
		// Remove from state
		resp.State.RemoveResource(ctx)
	}
	if err != nil {
		return
	}
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, adminUser)

	if isInvitationAccepted(state) && state.AccessType.ValueString() == string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM) {
		adminId := state.AdminId.ValueString()
		accessPolicies, err := getAccessPolicies(ctx, r.client, adminId)
		if err != nil {
			resp.Diagnostics.AddError("Error getting access policies for user "+state.AdminId.ValueString(),
				"\nError message: "+util.ReadClientError(err))
			return
		}
		state = state.RefreshPropertyValuesForPolicies(ctx, &resp.Diagnostics, accessPolicies)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ccAdminUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CCAdminUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !isInvitationAccepted(state) {
		resp.Diagnostics.AddError(
			"Error updating admin user with email: "+state.Email.ValueString(),
			"User should first accept the invitation before updating the user.",
		)
		return
	}
	adminId := state.AdminId.ValueString()

	var plan CCAdminUserResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.FirstName.IsNull() && state.FirstName.ValueString() != plan.FirstName.ValueString() {
		resp.Diagnostics.AddError(
			"Error updating first name",
			"First name cannot be updated",
		)
		return
	}

	if !plan.LastName.IsNull() && state.LastName.ValueString() != plan.LastName.ValueString() {
		resp.Diagnostics.AddError(
			"Error updating last name",
			"Last name cannot be updated",
		)
		return
	}

	if !plan.DisplayName.IsNull() && state.DisplayName.ValueString() != plan.DisplayName.ValueString() {
		resp.Diagnostics.AddError(
			"Error updating display name",
			"Display name cannot be updated",
		)
		return
	}

	accessPolicy, err := getAdminUserPolicies(ctx, &resp.Diagnostics, r.client, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policies for the user",
			"Error message: "+err.Error())
		return
	}

	updateAdministratorAccessModel := ccadmins.AdministratorAccessModel{}
	adminAccessType, err := getAdminAccessType(plan.AccessType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid access type",
			"Error message: "+err.Error())
		return
	}
	updateAdministratorAccessModel.SetAccessType(adminAccessType)
	updateAdministratorAccessModel.SetPolicies(accessPolicy)
	updateAdminUserRequest := r.client.CCAdminsClient.AdministratorsAPI.UpdateAdministratorAccess(ctx)
	updateAdminUserRequest = updateAdminUserRequest.CitrixCustomerId(r.client.ClientConfig.CustomerId)
	updateAdminUserRequest = updateAdminUserRequest.Id(adminId)
	updateAdminUserRequest = updateAdminUserRequest.AdministratorAccessModel(updateAdministratorAccessModel)
	httpResp, err := citrixdaasclient.AddRequestData(updateAdminUserRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating policies for "+plan.AdminId.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	plan, err = fetchAndUpdateAdminUser(ctx, r.client, plan, &resp.Diagnostics)
	if err != nil {
		return // Error already added to diagnostics
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ccAdminUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state CCAdminUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the user has accepted the invitation if not delete the invitation
	if !isInvitationAccepted(state) && state.Email.ValueString() != "" {
		deleteAdminInvitationRequest := r.client.CCAdminsClient.AdministratorsAPI.DeleteInvitation(ctx)
		deleteAdminInvitationRequest = deleteAdminInvitationRequest.CitrixCustomerId(r.client.ClientConfig.CustomerId)
		deleteAdminInvitationRequest = deleteAdminInvitationRequest.Email(state.Email.ValueString())
		_, httpResp, err := citrixdaasclient.AddRequestData(deleteAdminInvitationRequest, r.client).Execute()
		if err != nil && httpResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Error deleting admin user invitation with email: "+state.Email.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
	} else {
		deleteAdminUserRequest := r.client.CCAdminsClient.AdministratorsAPI.DeleteAdministrator(ctx, state.AdminId.ValueString())
		deleteAdminUserRequest = deleteAdminUserRequest.CitrixCustomerId(r.client.ClientConfig.CustomerId)
		httpResp, err := citrixdaasclient.AddRequestData(deleteAdminUserRequest, r.client).Execute()
		if err != nil && httpResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Error deleting admin user with id: "+state.AdminId.ValueString(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}
	}
}

func (r *ccAdminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("admin_id"), req, resp)
}

func (r *ccAdminUserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data CCAdminUserResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if strings.EqualFold(data.AccessType.ValueString(), string(ccadmins.ADMINISTRATORACCESSTYPE_FULL)) && !data.Policies.IsNull() {
		resp.Diagnostics.AddError(
			"Error validating policies",
			"Full access type does not require policies",
		)
	}

	if strings.EqualFold(data.AccessType.ValueString(), string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM)) && !data.Policies.IsUnknown() && data.Policies.IsNull() {
		resp.Diagnostics.AddError(
			"Error validating policies",
			"Policies are required to be set for access type Custom",
		)
		return
	}

	if data.Type.ValueString() == string(ccadmins.ADMINISTRATORTYPE_ADMINISTRATOR_USER) {

		if !data.Email.IsUnknown() && data.Email.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating email",
				"Email is required for Administrator type User",
			)
			return
		}

		// TODO: Implement validation for the provider type field when set to 'AzureAd' for users, pending API support.
		if data.ProviderType.ValueString() != string(ccadmins.ADMINISTRATORPROVIDERTYPE_CITRIX_STS) {
			resp.Diagnostics.AddError(
				"Error validating provider type",
				"Provider type should be CitrixSts for Administrator type User",
			)
			return
		}

		if !data.ExternalProviderId.IsNull() || !data.ExternalUserId.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating external provider id and external user id",
				"Administrator type User does not require external provider id and external user id",
			)
			return
		}
	}

	if data.Type.ValueString() == string(ccadmins.ADMINISTRATORTYPE_ADMINISTRATOR_GROUP) {

		if !data.Email.IsNull() {
			resp.Diagnostics.AddError(
				"Error validating email",
				"Email is not supported for Administrator Groups",
			)
			return
		}

		// TODO: Remove this once https://updates.cloud.com/details/cc50882/ is completed
		if data.AccessType.ValueString() != string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM) {
			resp.Diagnostics.AddError(
				"Error validating access type",
				"Access type should be Custom for Administrator type Group",
			)
			return
		}
	}

	if (data.ProviderType.ValueString() == string(ccadmins.ADMINISTRATORPROVIDERTYPE_AZURE_AD) || data.ProviderType.ValueString() == string(ccadmins.ADMINISTRATOREXTERNALPROVIDERTYPE_AD)) && !data.ExternalProviderId.IsUnknown() && !data.ExternalUserId.IsUnknown() && (data.ExternalProviderId.IsNull() || data.ExternalUserId.IsNull()) {
		resp.Diagnostics.AddError(
			"Error validating provider type",
			"External provider id and external user id are required for provider type Azure AD and AD",
		)
		return
	}

	if data.ProviderType.ValueString() == string(ccadmins.ADMINISTRATORPROVIDERTYPE_AZURE_AD) && !regexp.MustCompile(util.GuidRegex).MatchString(data.ExternalProviderId.ValueString()) {
		resp.Diagnostics.AddError(
			"Error validating external provider id",
			"The external provider ID for AzureAd must be a valid GUID",
		)
		return
	}

	if data.ProviderType.ValueString() == string(ccadmins.ADMINISTRATORPROVIDERTYPE_AD) && !regexp.MustCompile(util.DomainFqdnRegex).MatchString(data.ExternalProviderId.ValueString()) {
		resp.Diagnostics.AddError(
			"Error validating external provider id",
			"The external provider ID for AD must be in FQDN format",
		)
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *ccAdminUserResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.CCAdminsClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
