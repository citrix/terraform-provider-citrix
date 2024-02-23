// Copyright © 2023. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"net/http"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &vsphereHypervisorResource{}
	_ resource.ResourceWithConfigure   = &vsphereHypervisorResource{}
	_ resource.ResourceWithImportState = &vsphereHypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewVsphereHypervisorResource() resource.Resource {
	return &vsphereHypervisorResource{}
}

type vsphereHypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata implements resource.Resource.
func (*vsphereHypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vsphere_hypervisor"
}

// Configure implements resource.ResourceWithConfigure.
func (r *vsphereHypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (r *vsphereHypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vsphere hypervisor.",
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
			"zone": schema.StringAttribute{
				Description: "Id of the zone the hypervisor is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username of the hypervisor.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password of the hypervisor.",
				Required:    true,
			},
			"password_format": schema.StringAttribute{
				Description: "Password format of the hypervisor. Choose between Base64 and PlainText.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_BASE64),
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_PLAIN_TEXT),
					),
				},
			},
			"addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Hypervisor address(es).  At least one is required.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.IPv4RegexWithProtocol), "must be a valid IPv4 address prefixed with protoccol (http:// or https://)"),
					),
				},
			},
			"ssl_thumbprints": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "SSL certificate thumbprints to consider acceptable for this connection.  If not specified, and the hypervisor uses SSL for its connection, the SSL certificate's root certification authority and any intermediate certificates must be trusted.",
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.SslThumbprintRegex), "must be specified with SSL thumbprint without colons"),
					),
				},
			},
			"max_absolute_active_actions": schema.Int64Attribute{
				Description: "Maximum number of actions that can execute in parallel on the hypervisor. Default is 40.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(40),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_absolute_new_actions_per_minute": schema.Int64Attribute{
				Description: "Maximum number of actions that can be started on the hypervisor per-minute. Default is 10.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_power_actions_percentage_of_machines": schema.Int64Attribute{
				Description: "Maximum percentage of machines on the hypervisor which can have their power state changed simultaneously. Default is 20.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
	}
}

// ImportState implements resource.ResourceWithImportState.
func (*vsphereHypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create implements resource.Resource.
func (r *vsphereHypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan VsphereHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER)
	connectionDetails.SetUserName(plan.Username.ValueString())
	connectionDetails.SetPassword(plan.Password.ValueString())
	pwdFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.PasswordFormat.ValueString())
	if err != nil || pwdFormat == nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor for Vsphere",
			"Unsupported password format: "+plan.PasswordFormat.ValueString(),
		)
	}
	connectionDetails.SetPasswordFormat(*pwdFormat)

	addresses := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Addresses)
	connectionDetails.SetAddresses(addresses)

	if plan.SslThumbprints != nil {
		sslThumbprints := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.SslThumbprints)
		connectionDetails.SetSslThumbprints(sslThumbprints)
	}

	connectionDetails.SetMaxAbsoluteActiveActions(int32(plan.MaxAbsoluteActiveActions.ValueInt64()))
	connectionDetails.SetMaxAbsoluteNewActionsPerMinute(int32(plan.MaxAbsoluteNewActionsPerMinute.ValueInt64()))
	connectionDetails.SetMaxPowerActionsPercentageOfMachines(int32(plan.MaxPowerActionsPercentageOfMachines.ValueInt64()))

	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	hypervisor, err := CreateHypervisor(ctx, r.client, &resp.Diagnostics, body)
	if err != nil {
		// Directly return. Error logs have been populated in common function.
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

// Read implements resource.Resource.
func (r *vsphereHypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state VsphereHypervisorResourceModel
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

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not a Vsphere connection type hypervisor.",
		)
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

// Update implements resource.Resource.
func (r *vsphereHypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan VsphereHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER)
	editHypervisorRequestBody.SetUserName(plan.Username.ValueString())
	editHypervisorRequestBody.SetPassword(plan.Password.ValueString())
	pwdFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(plan.PasswordFormat.ValueString())
	if err != nil || pwdFormat == nil {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor for Vsphere",
			"Unsupported password format: "+plan.PasswordFormat.ValueString(),
		)
	}
	editHypervisorRequestBody.SetPasswordFormat(*pwdFormat)

	addresses := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Addresses)
	editHypervisorRequestBody.SetAddresses(addresses)

	sslThumbprints := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.SslThumbprints)
	editHypervisorRequestBody.SetSslThumbprints(sslThumbprints)

	editHypervisorRequestBody.SetMaxAbsoluteActiveActions(int32(plan.MaxAbsoluteActiveActions.ValueInt64()))
	editHypervisorRequestBody.SetMaxAbsoluteNewActionsPerMinute(int32(plan.MaxAbsoluteNewActionsPerMinute.ValueInt64()))
	editHypervisorRequestBody.SetMaxPowerActionsPercentageOfMachines(int32(plan.MaxPowerActionsPercentageOfMachines.ValueInt64()))

	// Patch hypervisor
	updatedHypervisor, err := UpdateHypervisor(ctx, r.client, &resp.Diagnostics, hypervisor, editHypervisorRequestBody)
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

// Delete implements resource.Resource.
func (r *vsphereHypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state VsphereHypervisorResourceModel
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
