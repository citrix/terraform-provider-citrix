// Copyright © 2023. Citrix Systems, Inc.

package delivery_group

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &deliveryGroupResource{}
	_ resource.ResourceWithConfigure      = &deliveryGroupResource{}
	_ resource.ResourceWithImportState    = &deliveryGroupResource{}
	_ resource.ResourceWithValidateConfig = &deliveryGroupResource{}
	_ resource.ResourceWithModifyPlan     = &deliveryGroupResource{}
	_ resource.ResourceWithValidateConfig = &deliveryGroupResource{}
)

// NewDeliveryGroupResource is a helper function to simplify the provider implementation.
func NewDeliveryGroupResource() resource.Resource {
	return &deliveryGroupResource{}
}

// deliveryGroupResource is the resource implementation.
type deliveryGroupResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *deliveryGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_delivery_group"
}

// Schema defines the schema for the resource.
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
			"restricted_access_users": getSchemaForRestrictedAccessUsers(true),
			"allow_anonymous_access": schema.BoolAttribute{
				Description: "Give access to unauthenticated (anonymous) users; no credentials are required to access StoreFront. This feature requires a StoreFront store for unauthenticated users.",
				Optional:    true,
			},
			"desktops": schema.ListNestedAttribute{
				Description: "A list of Desktop resources to publish on the delivery group. Only 1 desktop can be added to a Remote PC Delivery Group.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"published_name": schema.StringAttribute{
							Description: "A display name for the desktop.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description for the published desktop. The name and description are shown in Citrix Workspace app.",
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Specify whether to enable the delivery of this desktop.",
							Required:    true,
						},
						"enable_session_roaming": schema.BoolAttribute{
							Description: "When enabled, if the user launches this desktop and then moves to another device, the same session is used, and applications are available on both devices. When disabled, the session no longer roams between devices. Should be set to false for Remote PC Delivery Group.",
							Required:    true,
						},
						"restricted_access_users": getSchemaForRestrictedAccessUsers(false),
					},
				},
			},
			"associated_machine_catalogs": schema.ListNestedAttribute{
				Description: "Machine catalogs from which to assign machines to the newly created delivery group.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"machine_catalog": schema.StringAttribute{
							Description: "Id of the machine catalog from which to add machines.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
							},
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
			"autoscale_settings": schema.SingleNestedAttribute{
				Description: "The power management settings governing the machine(s) in the delivery group.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"autoscale_enabled": schema.BoolAttribute{
						Description: "Whether auto-scale is enabled for the delivery group.",
						Required:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "The time zone in which this delivery group's machines reside.",
						Optional:    true,
					},
					"peak_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the configured action should be performed after a user session disconnects in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"peak_log_off_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session ending in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session disconnecting in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_extended_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a second configurable period of a user session disconnecting in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"peak_extended_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the second configured action should be performed after a user session disconnects in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"off_peak_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the configured action should be performed after a user session disconnectts outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"off_peak_log_off_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session ending outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a configurable period of a user session disconnecting outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_extended_disconnect_action": schema.StringAttribute{
						Description: "The action to be performed after a second configurable period of a user session disconnecting outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
						Validators: []validator.String{
							sessionHostingActionEnumValidator(),
						},
					},
					"off_peak_extended_disconnect_timeout_minutes": schema.Int64Attribute{
						Description: "The number of minutes before the second configured action should be performed after a user session disconnects outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"peak_buffer_size_percent": schema.Int64Attribute{
						Description: "The percentage of machines in the delivery group that should be kept available in an idle state in peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"off_peak_buffer_size_percent": schema.Int64Attribute{
						Description: "The percentage of machines in the delivery group that should be kept available in an idle state outside peak hours.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"power_off_delay_minutes": schema.Int64Attribute{
						Description: "Delay before machines are powered off, when scaling down. Specified in minutes. By default, the power-off delay is 30 minutes. You can set it in a range of 0 to 60 minutes. Applies only to multi-session machines.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(30),
						Validators: []validator.Int64{
							int64validator.Between(0, 60),
						},
					},
					"disconnect_peak_idle_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which an idle session belonging to the delivery group is disconnected during peak time.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"disconnect_off_peak_idle_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which an idle session belonging to the delivery group is disconnected during off-peak time.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"log_off_peak_disconnected_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which a disconnected session belonging to the delivery group is terminated during peak time.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"log_off_off_peak_disconnected_session_after_seconds": schema.Int64Attribute{
						Description: "Specifies the time in seconds after which a disconnected session belonging to the delivery group is terminated during off peak time.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
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
									Description: "List of pool size schedules during the day. Each is specified as a time range and an indicator of the number of machines that should be powered on during that time range. Do not specify schedules when no machines should be powered on.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"time_range": schema.StringAttribute{
												Description: "Time range during which the pool size applies. Format is HH:mm-HH:mm. e.g. 09:00-17:00",
												Required:    true,
												Validators: []validator.String{
													stringvalidator.RegexMatches(regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0|3]0-([0-1][0-9]|2[0-3]):[0|3]0$`), "must be specified in format HH:mm-HH:mm and range between 00:00-00:00 with minutes being 00 or 30."),
												},
											},
											"pool_size": schema.Int64Attribute{
												Description: "The number of machines (either as an absolute number or a percentage of the machines in the delivery group, depending on the value of PoolUsingPercentage) that are to be maintained in a running state, whether they are in use or not.",
												Required:    true,
												Validators: []validator.Int64{
													int64validator.AtLeast(1),
												},
											},
										},
									},
									Validators: []validator.List{
										listvalidator.SizeAtLeast(1),
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
			"reboot_schedules": schema.ListNestedAttribute{
				Description: "The reboot schedule for the delivery group.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the reboot schedule.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the reboot schedule.",
							Optional:    true,
						},
						"reboot_schedule_enabled": schema.BoolAttribute{
							Description: "Whether the reboot schedule is enabled.",
							Required:    true,
						},
						"restrict_to_tag": schema.StringAttribute{
							Description: "The tag to which the reboot schedule is restricted.",
							Optional:    true,
						},
						"ignore_maintenance_mode": schema.BoolAttribute{
							Description: "Whether the reboot schedule ignores machines in the maintenance mode.",
							Required:    true,
						},
						"frequency": schema.StringAttribute{
							Description: "The frequency of the reboot schedule. Can only be set to `Daily`, `Weekly`, `Monthly`, or `Once`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"Daily",
									"Weekly",
									"Monthly",
									"Once",
								),
							},
						},
						"frequency_factor": schema.Int64Attribute{
							Description: "Repeats every X days/weeks/months. Minimum value is 1.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"start_date": schema.StringAttribute{
							Description: "The date on which the reboot schedule starts. The date format is `YYYY-MM-DD`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(util.DateRegex), "Date must be in the format YYYY-MM-DD"),
							},
						},
						"start_time": schema.StringAttribute{
							Description: "The time at which the reboot schedule starts. The time format is `HH:MM`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(util.TimeRegex), "Time must be in the format HH:MM"),
							},
						},
						"reboot_duration_minutes": schema.Int64Attribute{
							Description: "Restart all machines within x minutes. 0 means restarting all machines at the same time. To restart machines after draining sessions, set natural_reboot_schedule to true instead. ",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
						"natural_reboot_schedule": schema.BoolAttribute{
							Description: "Indicates whether the reboot will be a natural reboot, where the machines will be rebooted when they have no sessions. This should set to false for reboot_duration_minutes to work. Once UseNaturalReboot is set to true, RebootDurationMinutes won't have any effect.",
							Required:    true,
						},
						"days_in_week": schema.ListAttribute{
							ElementType: types.StringType,
							Description: "The days of the week on which the reboot schedule runs weekly. Can only be set to `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, or `Saturday`.",
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf(
										"Sunday",
										"Monday",
										"Tuesday",
										"Wednesday",
										"Thursday",
										"Friday",
										"Saturday",
									),
								),
							},
						},
						"week_in_month": schema.StringAttribute{
							Description: "The week in the month on which the reboot schedule runs monthly. Can only be set to `First`, `Second`, `Third`, `Fourth`, or `Last`.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"First",
									"Second",
									"Third",
									"Fourth",
									"Last",
								),
							},
						},
						"day_in_month": schema.StringAttribute{
							Description: "The day in the month on which the reboot schedule runs monthly. Can only be set to `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, or `Saturday`.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"Sunday",
									"Monday",
									"Tuesday",
									"Wednesday",
									"Thursday",
									"Friday",
									"Saturday",
								),
							},
						},
						"reboot_notification_to_users": schema.SingleNestedAttribute{
							Description: "The reboot notification for the reboot schedule. Not available for natural reboot.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"notification_duration_minutes": schema.Int64Attribute{
									Description: "Send notification to users X minutes before user is logged off. Can only be 0, 1, 5 or 15. 0 means no notification.",
									Required:    true,
									Validators: []validator.Int64{
										int64validator.OneOf(0, 1, 5, 15),
									},
								},
								"notification_title": schema.StringAttribute{
									Description: "The title to be displayed to users before they are logged off.",
									Required:    true,
								},
								"notification_message": schema.StringAttribute{
									Description: "The message to be displayed to users before they are logged off.",
									Required:    true,
								},
								"notification_repeat_every_5_minutes": schema.BoolAttribute{
									Description: "Repeat notification every 5 minutes, only available for 15 minutes notification duration. ",
									Optional:    true,
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

// Configure adds the provider configured client to the resource.
func (r *deliveryGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *deliveryGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get machine catalogs and verify all of them have the same session support
	catalogSessionSupport, areCatalogsPowerManaged, isRemotePcCatalog, identityType, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, &resp.Diagnostics, plan.AssociatedMachineCatalogs, true)

	if err != nil {
		return
	}

	if plan.AutoscaleSettings != nil && !areCatalogsPowerManaged {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group "+plan.Name.ValueString(),
			"Autoscale settings can only be configured if associated machine catalogs are power managed.",
		)
		return
	}

	if isRemotePcCatalog && plan.Desktops != nil && len(plan.Desktops) > 1 {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group "+plan.Name.ValueString(),
			"Only one assignment policy rule can be added to a Remote PC Delivery Group.",
		)
		return
	}

	if isRemotePcCatalog && plan.Desktops != nil && len(plan.Desktops) > 0 {
		desktops := plan.Desktops
		if desktops[0].EnableSessionRoaming.ValueBool() {
			resp.Diagnostics.AddError(
				"Error creating Delivery Group "+plan.Name.ValueString(),
				"enable_session_roaming cannot be set to true for Remote PC Delivery Group.",
			)
			return
		}

		if desktops[0].RestrictedAccessUsers != nil {
			resp.Diagnostics.AddError(
				"Error creating Delivery Group "+plan.Name.ValueString(),
				"restricted_access_users needs to be set for Remote PC Delivery Group.",
			)
			return
		}
	}

	body := getRequestModelForDeliveryGroupCreate(plan, catalogSessionSupport, identityType)

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

	deliveryGroupId := deliveryGroup.GetId()

	//Create Reboot Schedule after delivery group is created
	var editbody citrixorchestration.EditDeliveryGroupRequestModel
	editbody.SetRebootSchedules(body.GetRebootSchedules())
	updateDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsPatchDeliveryGroup(ctx, deliveryGroupId)
	updateDeliveryGroupRequest = updateDeliveryGroupRequest.EditDeliveryGroupRequestModel(editbody)
	httpResp, err = citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating reboot schedule for Delivery Group",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
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

	// Get machines
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deliveryGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state DeliveryGroupResourceModel
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

	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed delivery group properties from Orchestration
	deliveryGroupId := plan.Id.ValueString()
	deliveryGroupName := plan.Name.ValueString()
	currentDeliveryGroup, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	editDeliveryGroupRequestBody := getRequestModelForDeliveryGroupUpdate(plan, currentDeliveryGroup)

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

	// Add or remove machines
	err = addRemoveMachinesFromDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId, plan)

	if err != nil {
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

	// Get machines
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	// Fetch updated delivery group from GetDeliveryGroup.
	updatedDeliveryGroup, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(updatedDeliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Deletes the resource and removes the Terraform state on success.
func (r *deliveryGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DeliveryGroupResourceModel
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
			"Error deleting Delivery Group "+deliveryGroupName,
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

func (r *deliveryGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DeliveryGroupResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AutoscaleSettings == nil {
		return
	}

	validatePowerTimeSchemes(&resp.Diagnostics, data.AutoscaleSettings.PowerTimeSchemes)

	if data.RebootSchedules == nil {
		return
	}

	validateRebootSchedules(&resp.Diagnostics, data.RebootSchedules)
}

func (r *deliveryGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sessionSupport, areCatalogsPowerManaged, isRemotePcCatalog, _, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, &resp.Diagnostics, plan.AssociatedMachineCatalogs, !create)
	if err != nil || sessionSupport == nil {
		return
	}

	isValid, errMsg := validatePowerManagementSettings(plan, *sessionSupport)
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

	if plan.AutoscaleSettings != nil && !areCatalogsPowerManaged {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Autoscale settings can only be configured if associated machine catalogs are power managed.",
		)

		return
	}

	if isRemotePcCatalog && plan.Desktops != nil && len(plan.Desktops) > 1 {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Only one assignment policy rule can be added to a Remote PC Delivery Group",
		)
		return
	}

	if isRemotePcCatalog && plan.Desktops != nil && len(plan.Desktops) > 0 {
		desktops := plan.Desktops
		if desktops[0].EnableSessionRoaming.ValueBool() {
			resp.Diagnostics.AddError(
				"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
				"enable_session_roaming cannot be set to true for Remote PC Delivery Group.",
			)
			return
		}

		if desktops[0].RestrictedAccessUsers != nil {
			resp.Diagnostics.AddError(
				"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
				"restricted_access_users needs to be set for Remote PC Delivery Group.",
			)
			return
		}
	}
}
