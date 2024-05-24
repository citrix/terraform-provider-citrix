// Copyright © 2023. Citrix Systems, Inc.

package admin_role

import (
	"context"
	"net/http"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &adminRoleResource{}
	_ resource.ResourceWithConfigure      = &adminRoleResource{}
	_ resource.ResourceWithImportState    = &adminRoleResource{}
	_ resource.ResourceWithValidateConfig = &adminRoleResource{}
	_ resource.ResourceWithModifyPlan     = &adminRoleResource{}
)

// NewAdminRoleResource is a helper function to simplify the provider implementation.
func NewAdminRoleResource() resource.Resource {
	return &adminRoleResource{}
}

// adminRoleResource is the resource implementation.
type adminRoleResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *adminRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_role"
}

// Schema defines the schema for the resource.
func (r *adminRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Manages an administrator role.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the admin role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin role.",
				Optional:    true,
			},
			"is_built_in": schema.BoolAttribute{
				Description: "Flag to determine if the role was built-in or user defined",
				Computed:    true,
			},
			"can_launch_manage": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Manage tab on the console. This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true), // Default value gets set for an attribute after Validation and before applying configuration changes
			},
			"can_launch_monitor": schema.BoolAttribute{
				Description: "Flag to determine if the user will have access to the Monitor tab on the console. This field is only applicable for cloud admins. For on-premise admins, the only acceptable value is `true`. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"permissions": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of permissions to be associated with the admin role. To get a list of supported permissions, please refer to [Admin Predefined Permissions for Cloud](https://developer-docs.citrix.com/en-us/citrix-daas-service-apis/citrix-daas-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions) and [Admin Predefined Permissions for On-Premise](https://developer-docs.citrix.com/en-us/citrix-virtual-apps-desktops/citrix-cvad-rest-apis/apis/#/Admin-APIs/Admin-GetPredefinedPermissions).",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *adminRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *adminRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixorchestration.CreateAdminRoleRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetCanLaunchManage(plan.CanLaunchManage.ValueBool())
	body.SetCanLaunchMonitor(plan.CanLaunchMonitor.ValueBool())
	body.SetPermissions(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Permissions))

	createAdminRoleRequest := r.client.ApiClient.AdminAPIsDAAS.AdminCreateAdminRole(ctx)
	createAdminRoleRequest = createAdminRoleRequest.CreateAdminRoleRequestModel(body)

	// Create new admin role
	httpResp, err := citrixdaasclient.AddRequestData(createAdminRoleRequest, r.client).Execute()
	if err != nil {

		// If httpresponse is forbidden, then check if the role exists and delete it
		if httpResp.StatusCode == http.StatusForbidden {
			_, getRoleError := getAdminRole(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
			if getRoleError == nil {
				deleteAdminRoleRequest := r.client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminRole(ctx, plan.Name.ValueString())
				citrixdaasclient.AddRequestData(deleteAdminRoleRequest, r.client).Execute()
			}
		}

		resp.Diagnostics.AddError(
			"Error creating Admin Role: "+plan.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new admin role with role name
	adminRole, err := getAdminRole(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	plan = plan.RefreshPropertyValues(adminRole)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *adminRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AdminRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the admin role from remote
	adminRole, err := readAdminRole(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(adminRole)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *adminRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AdminRoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var adminRoleId = plan.Id.ValueString()
	var adminRoleName = plan.Name.ValueString()

	// Generate Update API request body from plan
	var body citrixorchestration.EditAdminRoleRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetCanLaunchManage(plan.CanLaunchManage.ValueBool())
	body.SetCanLaunchMonitor(plan.CanLaunchMonitor.ValueBool())
	body.SetPermissions(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Permissions))

	// Update admin role using orchestration call
	updateAdminRoleRequest := r.client.ApiClient.AdminAPIsDAAS.AdminUpdateAdminRole(ctx, adminRoleId)
	updateAdminRoleRequest = updateAdminRoleRequest.EditAdminRoleRequestModel(body)

	httpResp, err := citrixdaasclient.AddRequestData(updateAdminRoleRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Admin Role: "+adminRoleName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	// Fetch updated admin role using orchestration.
	updatedAdminRole, err := getAdminRole(ctx, r.client, &resp.Diagnostics, adminRoleId)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(updatedAdminRole)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *adminRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AdminRoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing admin role
	adminRoleId := state.Id.ValueString()
	adminRoleName := state.Name.ValueString()
	deleteAdminRoleRequest := r.client.ApiClient.AdminAPIsDAAS.AdminDeleteAdminRole(ctx, adminRoleId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteAdminRoleRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Admin Role: "+adminRoleName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *adminRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getAdminRole(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, adminRoleName string) (*citrixorchestration.RoleResponseModel, error) {
	// Get admin role
	getAdminRoleRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminRole(ctx, adminRoleName)
	adminRole, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.RoleResponseModel](getAdminRoleRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Admin Role: "+adminRoleName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return adminRole, err
}

func readAdminRole(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, adminRoleName string) (*citrixorchestration.RoleResponseModel, error) {
	getAdminRoleRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminRole(ctx, adminRoleName)
	adminRole, _, err := util.ReadResource[*citrixorchestration.RoleResponseModel](getAdminRoleRequest, ctx, client, resp, "Admin Role", adminRoleName)
	return adminRole, err
}

func (r *adminRoleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AdminRoleResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminRoleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	if !req.Plan.Raw.IsNull() {
		var plan AdminRoleResourceModel
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if r.client.AuthConfig.OnPremises {
			if !plan.CanLaunchManage.ValueBool() {
				resp.Diagnostics.AddError("CanLaunchManage", "CanLaunchManage can only be set to true for On-Premise deployments. Please either set the attribute to true or remove it from the configuration and try again.")
			}
			if !plan.CanLaunchMonitor.ValueBool() {
				resp.Diagnostics.AddError("CanLaunchMonitor", "CanLaunchMonitor can only be set to true for On-Premise deployments. Please either set the attribute to true or remove it from the configuration and try again.")
			}
		}
	}
}
