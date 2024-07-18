// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeliveryGroupMachineCatalogModel struct {
	MachineCatalog types.String `tfsdk:"machine_catalog"`
	MachineCount   types.Int64  `tfsdk:"machine_count"`
}

func (DeliveryGroupMachineCatalogModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

func (DeliveryGroupMachineCatalogModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupMachineCatalogModel{}.GetSchema().Attributes
}

type PowerTimeSchemePoolSizeScheduleRequestModel struct {
	TimeRange types.String `tfsdk:"time_range"`
	PoolSize  types.Int64  `tfsdk:"pool_size"`
}

func (PowerTimeSchemePoolSizeScheduleRequestModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

func (PowerTimeSchemePoolSizeScheduleRequestModel) GetAttributes() map[string]schema.Attribute {
	return PowerTimeSchemePoolSizeScheduleRequestModel{}.GetSchema().Attributes
}

type DeliveryGroupPowerTimeScheme struct {
	DaysOfWeek          types.Set    `tfsdk:"days_of_week"` //Set[string]
	DisplayName         types.String `tfsdk:"display_name"`
	PeakTimeRanges      types.Set    `tfsdk:"peak_time_ranges"`    //Set[string]
	PoolSizeSchedule    types.List   `tfsdk:"pool_size_schedules"` //List[PowerTimeSchemePoolSizeScheduleRequestModel]
	PoolUsingPercentage types.Bool   `tfsdk:"pool_using_percentage"`
}

func (DeliveryGroupPowerTimeScheme) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"days_of_week": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The pattern of days of the week that the power time scheme covers.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
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
			"peak_time_ranges": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Peak time ranges during the day. e.g. 09:00-17:00",
				Required:    true,
			},
			"pool_size_schedules": schema.ListNestedAttribute{
				Description:  "Pool size schedules during the day. Each is specified as a time range and an indicator of the number of machines that should be powered on during that time range. Do not specify schedules when no machines should be powered on.",
				Optional:     true,
				NestedObject: PowerTimeSchemePoolSizeScheduleRequestModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"pool_using_percentage": schema.BoolAttribute{
				Description: "Indicates whether the integer values in the pool size array are to be treated as absolute values (if this value is `false`) or as percentages of the number of machines in the delivery group (if this value is `true`).",
				Required:    true,
			},
		},
	}
}

func (DeliveryGroupPowerTimeScheme) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupPowerTimeScheme{}.GetSchema().Attributes
}

type DeliveryGroupRebootNotificationToUsers struct {
	NotificationDurationMinutes     types.Int64  `tfsdk:"notification_duration_minutes"`
	NotificationMessage             types.String `tfsdk:"notification_message"`
	NotificationRepeatEvery5Minutes types.Bool   `tfsdk:"notification_repeat_every_5_minutes"`
	NotificationTitle               types.String `tfsdk:"notification_title"`
}

func (DeliveryGroupRebootNotificationToUsers) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
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
	}
}

func (DeliveryGroupRebootNotificationToUsers) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupRebootNotificationToUsers{}.GetSchema().Attributes
}

// ensure DeliveryGroupRebootSchedule implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixorchestration.RebootScheduleResponseModel] = DeliveryGroupRebootSchedule{}

type DeliveryGroupRebootSchedule struct {
	Name                                   types.String `tfsdk:"name"`
	Description                            types.String `tfsdk:"description"`
	RebootScheduleEnabled                  types.Bool   `tfsdk:"reboot_schedule_enabled"`
	RestrictToTag                          types.String `tfsdk:"restrict_to_tag"`
	IgnoreMaintenanceMode                  types.Bool   `tfsdk:"ignore_maintenance_mode"`
	Frequency                              types.String `tfsdk:"frequency"`
	FrequencyFactor                        types.Int64  `tfsdk:"frequency_factor"`
	StartDate                              types.String `tfsdk:"start_date"`
	StartTime                              types.String `tfsdk:"start_time"`
	RebootDurationMinutes                  types.Int64  `tfsdk:"reboot_duration_minutes"`
	UseNaturalRebootSchedule               types.Bool   `tfsdk:"natural_reboot_schedule"`
	DaysInWeek                             types.Set    `tfsdk:"days_in_week"` //Set[string]
	WeekInMonth                            types.String `tfsdk:"week_in_month"`
	DayInMonth                             types.String `tfsdk:"day_in_month"`
	DeliveryGroupRebootNotificationToUsers types.Object `tfsdk:"reboot_notification_to_users"` //DeliveryGroupRebootNotificationToUsers
}

func (r DeliveryGroupRebootSchedule) GetKey() string {
	return r.Name.ValueString()
}

func (DeliveryGroupRebootSchedule) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the reboot schedule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the reboot schedule.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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
			"days_in_week": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The days of the week on which the reboot schedule runs weekly. Can only be set to `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, or `Saturday`.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
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
			"reboot_notification_to_users": DeliveryGroupRebootNotificationToUsers{}.GetSchema(),
		},
	}
}

func (DeliveryGroupRebootSchedule) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupRebootSchedule{}.GetSchema().Attributes
}

type DeliveryGroupPowerManagementSettings struct {
	AutoscaleEnabled                             types.Bool   `tfsdk:"autoscale_enabled"`
	Timezone                                     types.String `tfsdk:"timezone"`
	PeakDisconnectTimeoutMinutes                 types.Int64  `tfsdk:"peak_disconnect_timeout_minutes"`
	PeakLogOffAction                             types.String `tfsdk:"peak_log_off_action"`
	PeakDisconnectAction                         types.String `tfsdk:"peak_disconnect_action"`
	PeakExtendedDisconnectAction                 types.String `tfsdk:"peak_extended_disconnect_action"`
	PeakExtendedDisconnectTimeoutMinutes         types.Int64  `tfsdk:"peak_extended_disconnect_timeout_minutes"`
	OffPeakDisconnectTimeoutMinutes              types.Int64  `tfsdk:"off_peak_disconnect_timeout_minutes"`
	OffPeakLogOffAction                          types.String `tfsdk:"off_peak_log_off_action"`
	OffPeakDisconnectAction                      types.String `tfsdk:"off_peak_disconnect_action"`
	OffPeakExtendedDisconnectAction              types.String `tfsdk:"off_peak_extended_disconnect_action"`
	OffPeakExtendedDisconnectTimeoutMinutes      types.Int64  `tfsdk:"off_peak_extended_disconnect_timeout_minutes"`
	PeakBufferSizePercent                        types.Int64  `tfsdk:"peak_buffer_size_percent"`
	OffPeakBufferSizePercent                     types.Int64  `tfsdk:"off_peak_buffer_size_percent"`
	PowerOffDelayMinutes                         types.Int64  `tfsdk:"power_off_delay_minutes"`
	DisconnectPeakIdleSessionAfterSeconds        types.Int64  `tfsdk:"disconnect_peak_idle_session_after_seconds"`
	DisconnectOffPeakIdleSessionAfterSeconds     types.Int64  `tfsdk:"disconnect_off_peak_idle_session_after_seconds"`
	LogoffPeakDisconnectedSessionAfterSeconds    types.Int64  `tfsdk:"log_off_peak_disconnected_session_after_seconds"`
	LogoffOffPeakDisconnectedSessionAfterSeconds types.Int64  `tfsdk:"log_off_off_peak_disconnected_session_after_seconds"`
	PowerTimeSchemes                             types.List   `tfsdk:"power_time_schemes"` //List[DeliveryGroupPowerTimeScheme]
}

func (DeliveryGroupPowerManagementSettings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
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
				Description:  "Power management time schemes.  No two schemes for the same delivery group may cover the same day of the week.",
				Required:     true,
				NestedObject: DeliveryGroupPowerTimeScheme{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (DeliveryGroupPowerManagementSettings) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupPowerManagementSettings{}.GetSchema().Attributes
}

type RestrictedAccessUsers struct {
	AllowList types.Set `tfsdk:"allow_list"` //Set[string]
	BlockList types.Set `tfsdk:"block_list"` //Set[string]
}

func (RestrictedAccessUsers) GetSchema() schema.SingleNestedAttribute {
	return RestrictedAccessUsers{}.getSchemaInternal(false)
}

func (RestrictedAccessUsers) GetSchemaForDeliveryGroup() schema.SingleNestedAttribute {
	return RestrictedAccessUsers{}.getSchemaInternal(true)
}

func (RestrictedAccessUsers) getSchemaInternal(forDeliveryGroup bool) schema.SingleNestedAttribute {
	resource := "Delivery Group"
	description := "Restrict access to this Delivery Group by specifying users and groups in the allow and block list. If no value is specified, all authenticated users will have access to this Delivery Group. To give access to unauthenticated users, use the `allow_anonymous_access` property."
	if !forDeliveryGroup {
		resource = "Desktop"
		description = "Restrict access to this Desktop by specifying users and groups in the allow and block list. If no value is specified, all users that have access to this Delivery Group will have access to the Desktop. Required for Remote PC Delivery Groups."
	}

	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"allow_list": schema.SetAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who can use this %s. Must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
					setvalidator.SizeAtLeast(1),
				},
			},
			"block_list": schema.SetAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who cannot use this %s. A block list is meaningful only when used to block users in the allow list. Must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (RestrictedAccessUsers) GetAttributes() map[string]schema.Attribute {
	return RestrictedAccessUsers{}.GetSchema().Attributes
}

// ensure DeliveryGroupDesktop implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixorchestration.DesktopResponseModel] = DeliveryGroupDesktop{}

type DeliveryGroupDesktop struct {
	PublishedName         types.String `tfsdk:"published_name"`
	DesktopDescription    types.String `tfsdk:"description"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	EnableSessionRoaming  types.Bool   `tfsdk:"enable_session_roaming"`
	RestrictedAccessUsers types.Object `tfsdk:"restricted_access_users"` //RestrictedAccessUsers
}

func (r DeliveryGroupDesktop) GetKey() string {
	return r.PublishedName.ValueString()
}

func (DeliveryGroupDesktop) GetSchema() schema.NestedAttributeObject {
	var restrictedAccessUsers RestrictedAccessUsers
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"published_name": schema.StringAttribute{
				Description: "A display name for the desktop.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the published desktop. The name and description are shown in Citrix Workspace app.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Description: "Specify whether to enable the delivery of this desktop.",
				Required:    true,
			},
			"enable_session_roaming": schema.BoolAttribute{
				Description: "When enabled, if the user launches this desktop and then moves to another device, the same session is used, and applications are available on both devices. When disabled, the session no longer roams between devices. Should be set to false for Remote PC Delivery Group.",
				Required:    true,
			},
			"restricted_access_users": restrictedAccessUsers.GetSchema(),
		},
	}
}

func (DeliveryGroupDesktop) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupDesktop{}.GetSchema().Attributes
}

// DeliveryGroupResourceModel maps the resource schema data.
type DeliveryGroupResourceModel struct {
	Id                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	SessionSupport              types.String `tfsdk:"session_support"`
	SharingKind                 types.String `tfsdk:"sharing_kind"`
	RestrictedAccessUsers       types.Object `tfsdk:"restricted_access_users"`
	AllowAnonymousAccess        types.Bool   `tfsdk:"allow_anonymous_access"`
	Desktops                    types.List   `tfsdk:"desktops"`                    //List[DeliveryGroupDesktop]
	AssociatedMachineCatalogs   types.List   `tfsdk:"associated_machine_catalogs"` //List[DeliveryGroupMachineCatalogModel]
	AutoscaleSettings           types.Object `tfsdk:"autoscale_settings"`          //DeliveryGroupPowerManagementSettings
	RebootSchedules             types.List   `tfsdk:"reboot_schedules"`            //List[DeliveryGroupRebootSchedule]
	TotalMachines               types.Int64  `tfsdk:"total_machines"`
	PolicySetId                 types.String `tfsdk:"policy_set_id"`
	MinimumFunctionalLevel      types.String `tfsdk:"minimum_functional_level"`
	StoreFrontServers           types.Set    `tfsdk:"storefront_servers"` //Set[string]
	Scopes                      types.Set    `tfsdk:"scopes"`             //Set[String]
	MakeResourcesAvailableInLHC types.Bool   `tfsdk:"make_resources_available_in_lhc"`
}

func (DeliveryGroupResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
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
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"session_support": schema.StringAttribute{
				Description: "The session support for the delivery group. Can only be set to `SingleSession` or `MultiSession`. Specify only if you want to create a Delivery Group wthout any `associated_machine_catalogs`. Ensure session support is same as that of the prospective Machine Catalogs you will associate this Delivery Group with.",
				Optional:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionSupportEnumValues),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("sharing_kind"),
					}...),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRelative().AtParent().AtName("associated_machine_catalogs"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.ConfigValue.IsNull() && req.StateValue != req.ConfigValue
					},
						"Force replacement when session_support is changed",
						"Force replacement when session_support is changed",
					),
				},
			},
			"sharing_kind": schema.StringAttribute{
				Description: "The sharing kind for the delivery group. Can only be set to `Shared` or `Private`. Specify only if you want to create a Delivery Group wthout any `associated_machine_catalogs`.",
				Optional:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSharingKindEnumValues),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("session_support"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.ConfigValue.IsNull() && req.StateValue != req.ConfigValue
					},
						"Force replacement when sharing_kind is changed",
						"Force replacement when sharing_kind is changed",
					),
				},
			},
			"restricted_access_users": RestrictedAccessUsers{}.GetSchemaForDeliveryGroup(),
			"allow_anonymous_access": schema.BoolAttribute{
				Description: "Give access to unauthenticated (anonymous) users; no credentials are required to access StoreFront. This feature requires a StoreFront store for unauthenticated users.",
				Optional:    true,
			},
			"desktops": schema.ListNestedAttribute{
				Description:  "A list of Desktop resources to publish on the delivery group. Only 1 desktop can be added to a Remote PC Delivery Group.",
				Optional:     true,
				NestedObject: DeliveryGroupDesktop{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"associated_machine_catalogs": schema.ListNestedAttribute{
				Description:  "Machine catalogs from which to assign machines to the newly created delivery group.",
				Optional:     true,
				NestedObject: DeliveryGroupMachineCatalogModel{}.GetSchema(),
			},
			"autoscale_settings": DeliveryGroupPowerManagementSettings{}.GetSchema(),
			"reboot_schedules": schema.ListNestedAttribute{
				Description:  "The reboot schedule for the delivery group.",
				Optional:     true,
				NestedObject: DeliveryGroupRebootSchedule{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"total_machines": schema.Int64Attribute{
				Description: "The total number of machines in the delivery group.",
				Computed:    true,
			},
			"policy_set_id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"minimum_functional_level": schema.StringAttribute{
				Description: "Specifies the minimum functional level for the VDA machines in the delivery group. Defaults to `L7_20`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("L7_20"),
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(util.GetAllowedFunctionalLevelValues()...),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the delivery group to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"storefront_servers": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A list of GUID identifiers of StoreFront Servers to associate with the delivery group.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.StoreFrontServerIdRegex), "must be specified with StoreFront Server ID format"),
						),
					),
					setvalidator.SizeAtLeast(1),
				},
			},
			"make_resources_available_in_lhc": schema.BoolAttribute{
				Description: "In the event of a service disruption or loss of connectivity, select if you want Local Host Cache to keep resources in the delivery group available to launch new sessions. Existing sessions are not impacted. " +
					"This setting only impacts Single Session OS Random (pooled) desktops which are power managed. LHC is always enabled for Single Session OS static and Multi Session OS desktops." +
					"When set to `true`, machines will remain available and allow new connections and changes to the machine caused by a user might be present in subsequent sessions. " +
					"When set to `false`, machines in the delivery group will be unavailable for new connections during a Local Host Cache event. ",
				Optional: true,
			},
		},
	}
}

func (DeliveryGroupResourceModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupResourceModel{}.GetSchema().Attributes
}

func (r DeliveryGroupResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgDesktops *citrixorchestration.DesktopResponseModelCollection, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection, dgMachines *citrixorchestration.MachineResponseModelCollection, dgRebootSchedule *citrixorchestration.RebootScheduleResponseModelCollection) DeliveryGroupResourceModel {

	// Set required values
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())
	r.TotalMachines = types.Int64Value(int64(deliveryGroup.GetTotalMachines()))
	r.Description = types.StringValue(deliveryGroup.GetDescription())

	// Set optional values
	if deliveryGroup.GetPolicySetGuid() != "" {
		r.PolicySetId = types.StringValue(deliveryGroup.GetPolicySetGuid())
	} else {
		r.PolicySetId = types.StringNull()
	}

	minimumFunctionalLevel := deliveryGroup.GetMinimumFunctionalLevel()
	r.MinimumFunctionalLevel = types.StringValue(string(minimumFunctionalLevel))
	scopeIds := util.GetIdsForScopeObjects(deliveryGroup.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	if deliveryGroup.GetReuseMachinesWithoutShutdownInOutage() {
		r.MakeResourcesAvailableInLHC = types.BoolValue(true)
	} else if !r.MakeResourcesAvailableInLHC.IsNull() {
		r.MakeResourcesAvailableInLHC = types.BoolValue(false)
	}

	if len(dgMachines.GetItems()) < 1 {
		r.SessionSupport = types.StringValue(string(deliveryGroup.GetSessionSupport()))
		r.SharingKind = types.StringValue(string(deliveryGroup.GetSharingKind()))
	}

	r = r.updatePlanWithRestrictedAccessUsers(ctx, diagnostics, deliveryGroup)
	r = r.updatePlanWithDesktops(ctx, diagnostics, dgDesktops)
	r = r.updatePlanWithAssociatedCatalogs(ctx, diagnostics, dgMachines)
	r = r.updatePlanWithAutoscaleSettings(ctx, diagnostics, deliveryGroup, dgPowerTimeSchemes)
	r = r.updatePlanWithRebootSchedule(ctx, diagnostics, dgRebootSchedule)

	if len(deliveryGroup.GetStoreFrontServersForHostedReceiver()) > 0 || !r.StoreFrontServers.IsNull() {
		var remoteAssociatedStoreFrontServers []string
		for _, server := range deliveryGroup.GetStoreFrontServersForHostedReceiver() {
			remoteAssociatedStoreFrontServers = append(remoteAssociatedStoreFrontServers, server.GetId())
		}
		r.StoreFrontServers = util.StringArrayToStringSet(ctx, diagnostics, remoteAssociatedStoreFrontServers)
	} else {
		r.StoreFrontServers = types.SetNull(types.StringType)
	}

	return r
}
