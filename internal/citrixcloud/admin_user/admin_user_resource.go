// Copyright © 2024. Citrix Systems, Inc.

package cc_admin_user

import (
	"context"
	"fmt"
	"net/http"

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
	var body ccadmins.CitrixCloudServicesAdministratorsApiModelsCreateAdministratorInputModel
	body.SetType(plan.Type.ValueString())
	body.SetAccessType(plan.AccessType.ValueString())
	body.SetProviderType(plan.ProviderType.ValueString())
	body.SetEmail(plan.Email.ValueString())
	body.SetFirstName(plan.FirstName.ValueString())
	body.SetLastName(plan.LastName.ValueString())
	body.SetDisplayName(plan.DisplayName.ValueString())

	createAdminUserRequest := r.client.CCAdminsClient.AdministratorsAPI.CustomerAdministratorsCreatePost(ctx, r.client.ClientConfig.CustomerId)
	createAdminUserRequest = createAdminUserRequest.CitrixCloudServicesAdministratorsApiModelsCreateAdministratorInputModel(body)

	// Send invitation to the admin user
	_, httpResp, err := citrixdaasclient.AddRequestData(createAdminUserRequest, r.client).Execute()

	//In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating admin with email: "+plan.Email.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new admin user from remote
	adminUser, err := getAdminUser(ctx, r.client, plan.Email.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching admin user",
			util.ReadClientError(err),
		)
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, adminUser)

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
	var adminUser *ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorResult
	var err error

	adminUser, err = getAdminUser(ctx, r.client, state.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("Admin user with email: %s not found", state.Email.ValueString()),
			fmt.Sprintf("Admin user: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", state.Email.ValueString()),
		)
		// Remove from state
		resp.State.RemoveResource(ctx)
	}

	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, adminUser)

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

	resp.Diagnostics.AddError(
		"Error updating admin user",
		"Admin Users with access_type set to Full cannot be updated.",
	)
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

	if state.UserId.IsNull() || state.UserId.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Error deleting admin user",
			"Admin User needs to accept the invitation before they can be deleted.",
		)
		return
	}

	deleteAdminUserRequest := r.client.CCAdminsClient.AdministratorsAPI.CustomerAdministratorsAdminIdDelete(ctx, state.UserId.ValueString(), r.client.ClientConfig.CustomerId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteAdminUserRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting admin user with email: "+state.Email.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *ccAdminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("email"), req, resp)
}

func getAdminUser(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, adminUserEmail string) (*ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorResult, error) {

	// Get admin user
	getAdminUsersRequest := client.CCAdminsClient.AdministratorsAPI.CustomerAdministratorsGet(ctx, client.ClientConfig.CustomerId)
	var adminUser *ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorResult
	adminUsersResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorsResult](getAdminUsersRequest, client)

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return adminUser, err
	}

	for _, adminUser := range adminUsersResponse.GetItems() {
		if adminUser.GetEmail() == adminUserEmail {
			return &adminUser, nil
		}
	}

	for adminUsersResponse.GetContinuationToken() != "" {
		getAdminUsersRequest = getAdminUsersRequest.RequestContinuation(adminUsersResponse.GetContinuationToken())
		adminUsersResponse, httpResp, err = citrixdaasclient.ExecuteWithRetry[*ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorsResult](getAdminUsersRequest, client)
		if err != nil {
			err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
			return adminUser, err
		}

		for _, adminUser := range adminUsersResponse.GetItems() {
			if adminUser.GetEmail() == adminUserEmail {
				return &adminUser, nil
			}
		}
	}

	err = fmt.Errorf("could not find admin user with email: %s", adminUserEmail)

	return adminUser, err
}

func (r *ccAdminUserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data CCAdminUserResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
