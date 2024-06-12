// Copyright © 2024. Citrix Systems, Inc.

package admin_user

import (
	"context"
	"fmt"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &adminUserResource{}
	_ resource.ResourceWithConfigure   = &adminUserResource{}
	_ resource.ResourceWithImportState = &adminUserResource{}
)

// NewAdminUserResource is a helper function to simplify the provider implementation.
func NewAdminUserResource() resource.Resource {
	return &adminUserResource{}
}

// adminUserResource is the resource implementation.
type adminUserResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *adminUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_user"
}

// Schema defines the schema for the resource.
func (r *adminUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Manages an administrator user for on-premise environment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of an existing user in the active directory.",
				Required:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "Name of the domain that the user is a part of. For example, if the domain is `example.com`, then provide the value `example` for this field.",
				Required:    true,
			},
			"rights": schema.ListNestedAttribute{
				Description: "Rights to be associated with the admin user.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.StringAttribute{
							Description: "Name of the role to be associated with the admin user.",
							Required:    true,
						},
						"scope": schema.StringAttribute{
							Description: "Name of the scope to be associated with the admin user.",
							Required:    true,
						},
					},
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Flag to determine if the administrator is to be enabled or not.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *adminUserResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *adminUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var adminRights []citrixorchestration.AdminRightRequestModel
	for _, right := range util.ObjectListToTypedArray[RightsModel](ctx, &resp.Diagnostics, plan.Rights) {
		adminRights = append(adminRights, citrixorchestration.AdminRightRequestModel{
			Role:  right.Role.ValueString(),
			Scope: right.Scope.ValueString(),
		})
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateAdminAdministratorRequestModel
	body.SetUser(plan.DomainName.ValueString() + "\\" + plan.Name.ValueString())
	body.SetRights(adminRights)
	body.SetEnabled(plan.IsEnabled.ValueBool())

	createAdminUserRequest := r.client.ApiClient.AdminAPIsDAAS.AdminCreateAdminAdministrator(ctx)
	createAdminUserRequest = createAdminUserRequest.CreateAdminAdministratorRequestModel(body)

	// Create new admin user
	httpResp, err := citrixdaasclient.AddRequestData(createAdminUserRequest, r.client).Execute()

	//In case of error, add it to diagnostics so that the resource gets marked as tainted
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Admin: "+plan.DomainName.ValueString()+"\\"+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Try getting the new admin user from remote
	adminUser, err := getAdminIfExists(ctx, r.client, &resp.Diagnostics, plan.DomainName.ValueString(), plan.Name.ValueString())
	if err != nil {
		return
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
func (r *adminUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AdminUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the admin user from remote
	adminUser, err := readAdminUser(ctx, r.client, resp, state.Id.ValueString())
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
func (r *adminUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var adminUserId = plan.Id.ValueString()
	var adminUserName = plan.Name.ValueString()

	var adminRights []citrixorchestration.AdminRightRequestModel
	for _, right := range util.ObjectListToTypedArray[RightsModel](ctx, &resp.Diagnostics, plan.Rights) {
		adminRights = append(adminRights, citrixorchestration.AdminRightRequestModel{
			Role:  right.Role.ValueString(),
			Scope: right.Scope.ValueString(),
		})
	}

	// Generate Update API request body from plan
	var body citrixorchestration.UpdateAdminAdministratorRequestModel

	body.SetRights(adminRights)
	body.SetEnabled(plan.IsEnabled.ValueBool())

	// Update admin user using orchestration call
	updateAdminUserRequest := r.client.ApiClient.AdminAPIsDAAS.AdminUpdateAdminAdministrator(ctx, adminUserId)
	updateAdminUserRequest = updateAdminUserRequest.UpdateAdminAdministratorRequestModel(body)

	httpResp, err := citrixdaasclient.AddRequestData(updateAdminUserRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Admin User: "+adminUserName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated admin user using orchestration.
	updatedAdminUser, err := getAdminUser(ctx, r.client, &resp.Diagnostics, adminUserId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedAdminUser)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *adminUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AdminUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing admin user
	adminUserId := state.Id.ValueString()
	adminUserName := state.Name.ValueString()
	deleteAdminUserRequest := r.client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminAdministrator(ctx, adminUserId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteAdminUserRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Admin User: "+adminUserName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *adminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getAdminUser(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, adminUserFqdnOrId string) (*citrixorchestration.AdministratorResponseModel, error) {

	// Get admin user
	getAdminUserRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminAdministrator(ctx, adminUserFqdnOrId)
	adminUser, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.AdministratorResponseModel](getAdminUserRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Admin User: "+adminUserFqdnOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return adminUser, err
}

func getAdminIfExists(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, adminUserDomainName string, adminUserName string) (*citrixorchestration.AdministratorResponseModel, error) {
	adminUsers, errorMsg := getAllAdminUsers(ctx, client, diagnostics)
	if errorMsg != nil {
		return nil, errorMsg
	}

	for _, adminUser := range adminUsers {
		userDetails := adminUser.GetUser()
		if userDetails.GetSamName() == adminUserDomainName+"\\"+adminUserName {
			return &adminUser, nil
		}
	}

	errMsg := fmt.Sprintf("Could not find Admin User: %s\\%s", adminUserDomainName, adminUserName)
	err := fmt.Errorf(errMsg)
	diagnostics.AddError(
		"Error fetching Admin User: "+adminUserDomainName+"\\"+adminUserName,
		errMsg,
	)

	return nil, err

}

func getAllAdminUsers(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.AdministratorResponseModel, error) {
	// Get admin users
	getAdminUsersRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminAdministrators(ctx)

	var adminUsers []citrixorchestration.AdministratorResponseModel
	getAdminUsersResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.AdministratorResponseModelCollection](getAdminUsersRequest, client)

	if err != nil {
		diagnostics.AddError(
			"Error fetching Admin Users",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	adminUsers = getAdminUsersResponse.GetItems()

	for getAdminUsersResponse.GetContinuationToken() != "" {
		getAdminUsersRequest = getAdminUsersRequest.ContinuationToken(getAdminUsersResponse.GetContinuationToken())
		getAdminUsersResponse, httpResp, err = citrixdaasclient.ExecuteWithRetry[*citrixorchestration.AdministratorResponseModelCollection](getAdminUsersRequest, client)
		if err != nil {
			diagnostics.AddError(
				"Error fetching Admin Users",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}
		adminUsers = append(adminUsers, getAdminUsersResponse.GetItems()...)
	}

	return adminUsers, err
}

func readAdminUser(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, adminUserFqdnOrId string) (*citrixorchestration.AdministratorResponseModel, error) {
	getAdminUserRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminAdministrator(ctx, adminUserFqdnOrId)
	adminUser, _, err := util.ReadResource[*citrixorchestration.AdministratorResponseModel](getAdminUserRequest, ctx, client, resp, "Admin User", adminUserFqdnOrId)
	return adminUser, err
}

func (r *adminUserResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if !r.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Environment Not Supported", "This terraform resource is only supported for on-premise deployments")
	}

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan AdminUserResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
