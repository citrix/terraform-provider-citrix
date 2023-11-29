// Copyright © 2023. Citrix Systems, Inc.

package daas

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &deliveryGroupResource{}
	_ resource.ResourceWithConfigure   = &deliveryGroupResource{}
	_ resource.ResourceWithImportState = &deliveryGroupResource{}
	_ resource.ResourceWithModifyPlan  = &deliveryGroupResource{}
)

// NewDeliveryGroupResource is a helper function to simplify the provider implementation.
func NewDeliveryGroupResource() resource.Resource {
	return &deliveryGroupResource{}
}

// deliveryGroupResource is the resource implementation.
type deliveryGroupResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *deliveryGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_delivery_group"
}

// Schema defines the schema for the data source.
func (r *deliveryGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a delivery group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the delivery group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the delivery group.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the delivery group.",
				Optional:    true,
			},
			"associated_machine_catalogs": schema.ListNestedAttribute{
				Description: "Machine catalogs from which to assign machines to the newly created delivery group.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"machine_catalog": schema.StringAttribute{
							Description: "Id of the machine catalog from which to add machines.",
							Required:    true,
						},
						"machine_count": schema.Int64Attribute{
							Description: "The number of machines to assign from the machine catalog to the delivery group.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
			"users": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The users and groups who are explicitly granted access to the published desktop.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(`^[^@]+@\b(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}`), "Users must be in UPN format."),
						),
					),
				},
			},
			"autoscale_enabled": schema.BoolAttribute{
				Description: "Whether auto-scale is enabled for the delivery group.",
				Optional:    true,
			},
			"autoscale_settings": schema.SingleNestedAttribute{
				Description: "The power management settings governing the machine(s) in the delivery group.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"timezone": schema.StringAttribute{
						Description: "The time zone in which this delivery group's machines reside.",
						Optional:    true,
					},
					"peak_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the configured action should be performed after a user session disconnects in peak hours.",
						Optional:    true,
					},
					"peak_log_off_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session ending in peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session disconnecting in peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_extended_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a second configurable period of a user session disconnecting in peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_extended_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the second configured action should be performed after a user session disconnects in peak hours.",
						Optional:    true,
					},
					"off_peak_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the configured action should be performed after a user session disconnectts outside peak hours.",
						Optional:    true,
					},
					"off_peak_log_off_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session ending outside peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session disconnecting outside peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_extended_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a second configurable period of a user session disconnecting outside peak hours.",
						Optional:    true,
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_extended_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the second configured action should be performed after a user session disconnects outside peak hours.",
						Optional:    true,
					},
					"peak_buffer_size_percent": schema.Int64Attribute{
						Description: "The percentage of machines in the delivery group that should be kept available in an idle state in peak hours.",
						Optional:    true,
					},
					"off_peak_buffer_size_percent": schema.Int64Attribute{
						Description: "The percentage of machines in the delivery group that should be kept available in an idle state outside peak hours.",
						Optional:    true,
					},
					"power_off_delay_minutes": schema.Int64Attribute{
						Description: "Delay before machines are powered off, when scaling down. Specified in minutes. Applies only to multi-session machines.",
						Optional:    true,
					},
					"disconnect_peak_idle_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which an idle session belonging to the delivery group is disconnected during peak time.",
						Optional:    true,
					},
					"disconnect_off_peak_idle_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which an idle session belonging to the delivery group is disconnected during off-peak time.",
						Optional:    true,
					},
					"log_off_peak_disconnected_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which a disconnected session belonging to the delivery group is terminated during peak time.",
						Optional:    true,
					},
					"log_off_off_peak_disconnected_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which a disconnected session belonging to the delivery group is terminated during off peak time.",
						Optional:    true,
					},
					"power_time_schemes": schema.ListNestedAttribute{
						Description: "Power management time schemes.  No two schemes for the same delivery group may cover the same day of the week.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"days_of_week": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "The pattern of days of the week that the power time scheme covers.",
									Required:    true,
									Validators: []validator.List{
										listvalidator.ValueStringsAre(
											stringvalidator.OneOf(
												"Unknown",
												"Sunday",
												"Monday",
												"Tuesday",
												"Wednesday",
												"Thursday",
												"Friday",
												"Saturday",
												"Weekdays",
												"Weekend",
											),
										),
									},
								},
								"display_name": schema.StringAttribute{
									Description: "The name of the power time scheme as displayed in the console.",
									Required:    true,
								},
								"peak_time_ranges": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "List of peak time ranges during the day. e.g. 09:00-17:00",
									Required:    true,
								},
								"pool_size_schedules": schema.ListNestedAttribute{
									Description: "List of pool size schedules during the day. Each is specified as a time range and an indicator of the number of machines that should be powered on during that time range.",
									Required:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"time_range": schema.StringAttribute{
												Description: "Time range during which the pool size applies.",
												Required:    true,
											},
											"pool_size": schema.Int64Attribute{
												Description: "The number of machines (either as an absolute number or a percentage of the machines in the delivery group, depending on the value of PoolUsingPercentage) that are to be maintained in a running state, whether they are in use or not.",
												Required:    true,
											},
										},
									},
								},
								"pool_using_percentage": schema.BoolAttribute{
									Description: "Indicates whether the integer values in the pool size array are to be treated as absolute values (if this value is `false`) or as percentages of the number of machines in the delivery group (if this value is `true`).",
									Required:    true,
								},
							},
						},
					},
				},
			},
			"total_machines": schema.Int64Attribute{
				Description: "The total number of machines in the delivery group.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *deliveryGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *deliveryGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get machine catalogs and verify all of them have the same session support
	catalogSessionSupport, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, resp.Diagnostics, plan.AssociatedMachineCatalogs, true)

	if err != nil {
		return
	}

	deliveryGroupMachineCatalogsArray := getDeliveryGroupAddMachinesRequest(plan.AssociatedMachineCatalogs)

	var desktop citrixorchestration.DesktopRequestModel
	desktop.SetPublishedName(plan.Name.ValueString())
	desktop.SetMaxDesktops(1)
	desktop.SetEnabled(true)
	desktop.SetSessionReconnection(citrixorchestration.SESSIONRECONNECTION_ALWAYS)
	users := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Users)
	desktop.SetIncludedUsers(users)
	deliveryGroupDesktopsArray := []citrixorchestration.DesktopRequestModel{desktop}

	var simpleAccessPolicy citrixorchestration.SimplifiedAccessPolicyRequestModel
	simpleAccessPolicy.SetAllowAnonymous(false)
	simpleAccessPolicy.SetIncludedUserFilterEnabled(false)

	var body citrixorchestration.CreateDeliveryGroupRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetMachineCatalogs(deliveryGroupMachineCatalogsArray)
	body.SetMinimumFunctionalLevel(citrixorchestration.FUNCTIONALLEVEL_L7_20)
	deliveryKind := citrixorchestration.DELIVERYKIND_DESKTOPS_AND_APPS
	if catalogSessionSupport != citrixorchestration.SESSIONSUPPORT_MULTI_SESSION {
		deliveryKind = citrixorchestration.DELIVERYKIND_DESKTOPS_ONLY
	}
	body.SetDeliveryType(deliveryKind)
	body.SetDesktops(deliveryGroupDesktopsArray)
	body.SetDefaultDesktopPublishedName(plan.Name.ValueString())
	body.SetSimpleAccessPolicy(simpleAccessPolicy)

	body.SetAutoScaleEnabled(plan.AutoscaleEnabled.ValueBool())

	if plan.AutoscaleSettings != nil {
		body.SetTimeZone(plan.AutoscaleSettings.Timezone.ValueString())
		body.SetPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakDisconnectTimeoutMinutes.ValueInt64()))

		if plan.AutoscaleSettings.PeakLogOffAction.ValueString() != "" {
			body.SetPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakLogOffAction.ValueString()))
		}

		if plan.AutoscaleSettings.PeakDisconnectAction.ValueString() != "" {
			body.SetPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakDisconnectAction.ValueString()))
		}

		if plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString() != "" {
			body.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString()))
		}

		body.SetPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetOffPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.ValueInt64()))

		if plan.AutoscaleSettings.OffPeakLogOffAction.ValueString() != "" {
			body.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakLogOffAction.ValueString()))
		}

		if plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString() != "" {
			body.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString()))
		}

		if plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString() != "" {
			body.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString()))
		}

		body.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetPeakBufferSizePercent(int32(plan.AutoscaleSettings.PeakBufferSizePercent.ValueInt64()))
		body.SetOffPeakBufferSizePercent(int32(plan.AutoscaleSettings.OffPeakBufferSizePercent.ValueInt64()))
		body.SetPowerOffDelayMinutes(int32(plan.AutoscaleSettings.PowerOffDelayMinutes.ValueInt64()))
		body.SetDisconnectPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		body.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := models.ParsePowerTimeSchemesPluginToClientModel(plan.AutoscaleSettings.PowerTimeSchemes)
		body.SetPowerTimeSchemes(powerTimeSchemes)
	}

	createDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsCreateDeliveryGroup(ctx)
	createDeliveryGroupRequest = createDeliveryGroupRequest.CreateDeliveryGroupRequestModel(body)

	// Create new delivery group
	deliveryGroup, httpResp, err := citrixdaasclient.AddRequestData(createDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get desktops
	deliveryGroupId := deliveryGroup.GetId()
	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get power time schemes
	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get machines
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deliveryGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.DeliveryGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deliveryGroupId := state.Id.ValueString()
	deliveryGroup, err := readDeliveryGroup(ctx, r.client, resp, deliveryGroupId)
	if err != nil {
		return
	}

	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed delivery group properties from Orchestration
	deliveryGroupId := plan.Id.ValueString()
	deliveryGroupName := plan.Name.ValueString()
	_, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	var desktop citrixorchestration.DesktopRequestModel
	desktop.SetPublishedName(plan.Name.ValueString())
	desktop.SetMaxDesktops(1)
	desktop.SetEnabled(true)
	desktop.SetSessionReconnection(citrixorchestration.SESSIONRECONNECTION_ALWAYS)
	users := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Users)
	desktop.SetIncludedUsers(users)
	deliveryGroupDesktopsArray := []citrixorchestration.DesktopRequestModel{desktop}

	// Construct the update model
	var editDeliveryGroupRequestBody citrixorchestration.EditDeliveryGroupRequestModel
	editDeliveryGroupRequestBody.SetName(plan.Name.ValueString())
	editDeliveryGroupRequestBody.SetDescription(plan.Description.ValueString())
	editDeliveryGroupRequestBody.SetDesktops(deliveryGroupDesktopsArray)
	editDeliveryGroupRequestBody.SetAutoScaleEnabled(plan.AutoscaleEnabled.ValueBool())

	if plan.AutoscaleSettings != nil {
		editDeliveryGroupRequestBody.SetPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakDisconnectTimeoutMinutes.ValueInt64()))

		if plan.AutoscaleSettings.Timezone.ValueString() != "" {
			editDeliveryGroupRequestBody.SetTimeZone(plan.AutoscaleSettings.Timezone.ValueString())
		}

		if plan.AutoscaleSettings.PeakLogOffAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakLogOffAction.ValueString()))
		}

		if plan.AutoscaleSettings.PeakDisconnectAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakDisconnectAction.ValueString()))
		}

		if plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString()))
		}

		editDeliveryGroupRequestBody.SetPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.ValueInt64()))

		if plan.AutoscaleSettings.OffPeakLogOffAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakLogOffAction.ValueString()))
		}

		if plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString()))
		}

		if plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString() != "" {
			editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString()))
		}

		editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetPeakBufferSizePercent(int32(plan.AutoscaleSettings.PeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakBufferSizePercent(int32(plan.AutoscaleSettings.OffPeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetPowerOffDelayMinutes(int32(plan.AutoscaleSettings.PowerOffDelayMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := models.ParsePowerTimeSchemesPluginToClientModel(plan.AutoscaleSettings.PowerTimeSchemes)
		editDeliveryGroupRequestBody.SetPowerTimeSchemes(powerTimeSchemes)
	}

	updateDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsPatchDeliveryGroup(ctx, deliveryGroupId)
	updateDeliveryGroupRequest = updateDeliveryGroupRequest.EditDeliveryGroupRequestModel(editDeliveryGroupRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get desktops
	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get power time schemes
	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Add or remove machines
	err = addRemoveMachinesFromDeliveryGroup(ctx, r.client, resp.Diagnostics, deliveryGroupId, plan)

	if err != nil {
		return
	}

	// Get machines
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Fetch updated delivery group from GetDeliveryGroup.
	updatedDeliveryGroup, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(updatedDeliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Deletes the resource and removes the Terraform state on success.
func (r *deliveryGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state models.DeliveryGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing delivery group
	deliveryGroupId := state.Id.ValueString()
	deliveryGroupName := state.Name.ValueString()
	deleteDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDeleteDeliveryGroup(ctx, deliveryGroupId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteDeliveryGroupRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error Deleting Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *deliveryGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *deliveryGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan models.DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sessionSupport, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, resp.Diagnostics, plan.AssociatedMachineCatalogs, !create)
	if err != nil {
		return
	}

	isValid, errMsg := validatePowerManagementSettings(plan, sessionSupport)
	operation := "updating"
	if create {
		operation = "creating"
	}
	if !isValid {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Error message: "+errMsg,
		)

		return
	}
}

func getSessionChangeHostingActionValue(v string) citrixorchestration.SessionChangeHostingAction {
	hostingAction, err := citrixorchestration.NewSessionChangeHostingActionFromValue(v)

	if err != nil {
		return citrixorchestration.SESSIONCHANGEHOSTINGACTION_UNKNOWN
	}

	return *hostingAction
}

func sessionHostingActionEnumValidator() validator.String {
	return util.GetValidatorFromEnum(citrixorchestration.AllowedSessionChangeHostingActionEnumValues)
}

func validatePowerManagementSettings(plan models.DeliveryGroupResourceModel, sessionSupport citrixorchestration.SessionSupport) (bool, string) {
	if plan.AutoscaleSettings == nil || sessionSupport == citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION {
		return true, ""
	}

	errStringSuffix := "cannot be set for a Multisession catalog"

	if plan.AutoscaleSettings.PeakLogOffAction.ValueString() != "" && plan.AutoscaleSettings.PeakLogOffAction.ValueString() != "Nothing" {
		return false, "PeakLogOffAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakLogOffAction.ValueString() != "" && plan.AutoscaleSettings.OffPeakLogOffAction.ValueString() != "Nothing" {
		return false, "OffPeakLogOffAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakDisconnectAction.ValueString() != "" && plan.AutoscaleSettings.PeakDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString() != "" && plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString() != "" && plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString() != "Nothing" {
		return false, "OffPeakDisconnectAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString() != "" && plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString() != "Nothing" {
		return false, "OffPeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "PeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "PeakExtendedDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "OffPeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "OffPeakExtendedDisconnectTimeoutMinutes " + errStringSuffix
	}

	return true, ""
}

func getDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.DeliveryGroupDetailResponseModel, error) {
	getDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupId)
	deliveryGroup, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroup, err
}

func readDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, deliveryGroupId string) (*citrixorchestration.DeliveryGroupDetailResponseModel, error) {
	getDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupId)
	deliveryGroup, _, err := util.ReadResource[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, ctx, client, resp, "Delivery Group", deliveryGroupId)
	return deliveryGroup, err
}

func getDeliveryGroupDesktops(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.DesktopResponseModelCollection, error) {
	getDeliveryGroupDesktopsRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupsDesktops(ctx, deliveryGroupId)
	deliveryGroupDesktops, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DesktopResponseModelCollection](getDeliveryGroupDesktopsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Desktops for Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroupDesktops, err
}

func getDeliveryGroupPowerTimeSchemes(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.PowerTimeSchemeResponseModelCollection, error) {
	getDeliveryGroupPowerTimeSchemesRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupPowerTimeSchemes(ctx, deliveryGroupId)
	deliveryGroupPowerTimeSchemes, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.PowerTimeSchemeResponseModelCollection](getDeliveryGroupPowerTimeSchemesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Power Time Schemes for Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroupPowerTimeSchemes, err
}

func getDeliveryGroupMachines(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.MachineResponseModelCollection, error) {
	getDeliveryGroupMachineCatalogsRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupMachines(ctx, deliveryGroupId)
	deliveryGroupMachines, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineResponseModelCollection](getDeliveryGroupMachineCatalogsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Machines for Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroupMachines, err
}

func validateAndReturnMachineCatalogSessionSupport(ctx context.Context, client citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, dgMachineCatalogs []models.DeliveryGroupMachineCatalogModel, addErrorIfCatalogNotFound bool) (citrixorchestration.SessionSupport, error) {
	var sessionSupport *citrixorchestration.SessionSupport

	for _, dgMachineCatalog := range dgMachineCatalogs {
		catalog, err := GetMachineCatalog(ctx, &client, &diagnostics, dgMachineCatalog.MachineCatalog.ValueString(), addErrorIfCatalogNotFound)

		if err != nil {
			return "", err
		}

		if sessionSupport != nil && *sessionSupport != catalog.GetSessionSupport() {
			err := fmt.Errorf("all associated machine catalogs must have the same session support")
			diagnostics.AddError("Error validating associated Machine Catalogs", "Ensure all associated Machine Catalogs have the same Session Support.")
			return "", err
		}

		if sessionSupport == nil {
			sessionSupportValue := catalog.GetSessionSupport()
			sessionSupport = &sessionSupportValue
		}
	}

	return *sessionSupport, nil
}

func getDeliveryGroupAddMachinesRequest(associatedMachineCatalogs []models.DeliveryGroupMachineCatalogModel) []citrixorchestration.DeliveryGroupAddMachinesRequestModel {
	var deliveryGroupMachineCatalogsArray []citrixorchestration.DeliveryGroupAddMachinesRequestModel
	for _, associatedMachineCatalog := range associatedMachineCatalogs {
		var deliveryGroupMachineCatalogs citrixorchestration.DeliveryGroupAddMachinesRequestModel
		deliveryGroupMachineCatalogs.SetMachineCatalog(associatedMachineCatalog.MachineCatalog.ValueString())
		deliveryGroupMachineCatalogs.SetCount(int32(associatedMachineCatalog.MachineCount.ValueInt64()))
		deliveryGroupMachineCatalogs.SetAssignMachinesToUsers([]citrixorchestration.AssignMachineToUserRequestModel{})
		deliveryGroupMachineCatalogsArray = append(deliveryGroupMachineCatalogsArray, deliveryGroupMachineCatalogs)
	}

	return deliveryGroupMachineCatalogsArray
}

func createExistingCatalogsAndMachinesMap(deliveryGroupMachines *citrixorchestration.MachineResponseModelCollection) map[string][]string {
	catalogAndMachinesMap := map[string][]string{}
	for _, dgMachine := range deliveryGroupMachines.GetItems() {
		machineCatalog := dgMachine.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		machineCatalogMachines := catalogAndMachinesMap[machineCatalogId]
		if machineCatalogMachines == nil {
			catalogAndMachinesMap[machineCatalogId] = []string{}
		}
		catalogAndMachinesMap[machineCatalogId] = append(catalogAndMachinesMap[machineCatalogId], dgMachine.GetId())
	}

	return catalogAndMachinesMap
}

func addMachinesToDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, deliveryGroupId string, catalogId string, numOfMachines int) (*citrixorchestration.DeliveryGroupDetailResponseModel, error) {
	var deliveryGroupMachineCatalogs citrixorchestration.DeliveryGroupAddMachinesRequestModel
	var deliveryGroupAssignMachinesToUsers []citrixorchestration.AssignMachineToUserRequestModel
	deliveryGroupMachineCatalogs.SetMachineCatalog(catalogId)
	deliveryGroupMachineCatalogs.SetCount(int32(numOfMachines))
	deliveryGroupMachineCatalogs.SetAssignMachinesToUsers(deliveryGroupAssignMachinesToUsers)

	updateDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDoAddMachines(ctx, deliveryGroupId)
	updateDeliveryGroupRequest = updateDeliveryGroupRequest.DeliveryGroupAddMachinesRequestModel(deliveryGroupMachineCatalogs)
	updatedDeliveryGroup, httpResp, err := citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error adding machine(s) to Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return updatedDeliveryGroup, err
}

func removeMachinesFromDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, deliveryGroupId string, machinesToRemove []string) error {
	for _, machineToRemove := range machinesToRemove {
		updateDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDoRemoveMachines(ctx, deliveryGroupId, machineToRemove)
		httpResp, err := citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, client).Execute()
		if err != nil {
			diagnostics.AddError(
				"Error removing machine from Delivery Group "+deliveryGroupId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)

			return err
		}
	}

	return nil
}

func addRemoveMachinesFromDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, deliveryGroupId string, plan models.DeliveryGroupResourceModel) error {
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, client, &diagnostics, deliveryGroupId)

	if err != nil {
		return err
	}

	existingAssociatedMachineCatalogsMap := createExistingCatalogsAndMachinesMap(deliveryGroupMachines)

	requestedAssociatedMachineCatalogsMap := map[string]bool{}
	for _, associatedMachineCatalog := range plan.AssociatedMachineCatalogs {

		requestedAssociatedMachineCatalogsMap[associatedMachineCatalog.MachineCatalog.ValueString()] = true

		associatedMachineCatalogId := associatedMachineCatalog.MachineCatalog.ValueString()
		requestedCount := int(associatedMachineCatalog.MachineCount.ValueInt64())
		machineCatalogMachines := existingAssociatedMachineCatalogsMap[associatedMachineCatalogId]
		existingCount := len(machineCatalogMachines)

		if requestedCount > existingCount {
			// add machines
			machineCount := (requestedCount - existingCount)
			_, err := addMachinesToDeliveryGroup(ctx, client, diagnostics, deliveryGroupId, associatedMachineCatalogId, machineCount)
			if err != nil {
				return err
			}
		}

		if requestedCount < existingCount {
			// remove machines
			machinesToRemoveCount := existingCount - requestedCount
			machineCatalogMachines := existingAssociatedMachineCatalogsMap[associatedMachineCatalogId]
			machinesToRemove := machineCatalogMachines[0:machinesToRemoveCount]

			err := removeMachinesFromDeliveryGroup(ctx, client, diagnostics, deliveryGroupId, machinesToRemove)

			if err != nil {
				return err
			}
		}
	}

	for key := range existingAssociatedMachineCatalogsMap {
		if !requestedAssociatedMachineCatalogsMap[key] {
			// remove all machines from this catalog
			machinesToRemove := existingAssociatedMachineCatalogsMap[key]

			err := removeMachinesFromDeliveryGroup(ctx, client, diagnostics, deliveryGroupId, machinesToRemove)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
