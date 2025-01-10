// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
				Description: "Time range during which the pool size applies. " +
					"\n\n-> **Note** Time range format is `HH:mm-HH:mm`, e.g. `09:00-17:00`",
				Required: true,
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
				Description: "Pool size schedules during the day. Each is specified as a time range and an indicator of the number of machines that should be powered on during that time range. " +
					"\n\n~> **Please Note** Do not specify schedules when no machines should be powered on.",
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
		Description: "The reboot notification for the reboot schedule. " +
			"\n\n~> **Please Note** Not available for natural reboot.",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"notification_duration_minutes": schema.Int64Attribute{
				Description: "Send notification to users X minutes before user is logged off. Can only be `0`, `1`, `5` or `15`. `0` means no notification.",
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
				Description: "Repeat notification every 5 minutes. " +
					"\n\n~> **Please Note** notification repeat is available only when `notification_duration_minutes` is set to `15`.",
				Optional: true,
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
				Description: "Restrict reboot schedule to machines with tag specified in Guid.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
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
				Description: "Repeats every X days/weeks/months. Minimum value is `1`.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"start_date": schema.StringAttribute{
				Description: "The date on which the reboot schedule starts. " +
					"\n\n-> **Note** The date format is `YYYY-MM-DD`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.DateRegex), "Date must be in the format YYYY-MM-DD"),
				},
			},
			"start_time": schema.StringAttribute{
				Description: "The time at which the reboot schedule starts. " +
					"\n\n-> **Note** The time format is `HH:MM`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeRegex), "Time must be in the format HH:MM"),
				},
			},
			"reboot_duration_minutes": schema.Int64Attribute{
				Description: "Restart all machines within x minutes. 0 means restarting all machines at the same time. To restart machines after draining sessions, set natural_reboot_schedule to true instead. ",
				Optional:    true,
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
	AutoscaleEnabled                                     types.Bool   `tfsdk:"autoscale_enabled"`
	RestrictAutoscaleTag                                 types.String `tfsdk:"restrict_autoscale_tag"`
	RestrictAutoscaleMinIdleUntaggedPercentDuringPeak    types.Int32  `tfsdk:"peak_restrict_min_idle_untagged_percent"`
	RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak types.Int32  `tfsdk:"off_peak_restrict_min_idle_untagged_percent"`
	Timezone                                             types.String `tfsdk:"timezone"`
	PeakDisconnectTimeoutMinutes                         types.Int64  `tfsdk:"peak_disconnect_timeout_minutes"`
	PeakLogOffAction                                     types.String `tfsdk:"peak_log_off_action"`
	PeakLogOffTimeoutMinutes                             types.Int64  `tfsdk:"peak_log_off_timeout_minutes"`
	PeakDisconnectAction                                 types.String `tfsdk:"peak_disconnect_action"`
	PeakExtendedDisconnectAction                         types.String `tfsdk:"peak_extended_disconnect_action"`
	PeakExtendedDisconnectTimeoutMinutes                 types.Int64  `tfsdk:"peak_extended_disconnect_timeout_minutes"`
	OffPeakDisconnectTimeoutMinutes                      types.Int64  `tfsdk:"off_peak_disconnect_timeout_minutes"`
	OffPeakLogOffAction                                  types.String `tfsdk:"off_peak_log_off_action"`
	OffPeakLogOffTimeoutMinutes                          types.Int64  `tfsdk:"off_peak_log_off_timeout_minutes"`
	OffPeakDisconnectAction                              types.String `tfsdk:"off_peak_disconnect_action"`
	OffPeakExtendedDisconnectAction                      types.String `tfsdk:"off_peak_extended_disconnect_action"`
	OffPeakExtendedDisconnectTimeoutMinutes              types.Int64  `tfsdk:"off_peak_extended_disconnect_timeout_minutes"`
	PeakBufferSizePercent                                types.Int64  `tfsdk:"peak_buffer_size_percent"`
	OffPeakBufferSizePercent                             types.Int64  `tfsdk:"off_peak_buffer_size_percent"`
	PowerOffDelayMinutes                                 types.Int64  `tfsdk:"power_off_delay_minutes"`
	PeakAutoscaleAssignedPowerOnIdleAction               types.String `tfsdk:"peak_autoscale_assigned_power_on_idle_action"`
	PeakAutoscaleAssignedPowerOnIdleTimeoutMinutes       types.Int64  `tfsdk:"peak_autoscale_assigned_power_on_idle_timeout_minutes"`
	DisconnectPeakIdleSessionAfterSeconds                types.Int64  `tfsdk:"disconnect_peak_idle_session_after_seconds"`
	DisconnectOffPeakIdleSessionAfterSeconds             types.Int64  `tfsdk:"disconnect_off_peak_idle_session_after_seconds"`
	LogoffPeakDisconnectedSessionAfterSeconds            types.Int64  `tfsdk:"log_off_peak_disconnected_session_after_seconds"`
	LogoffOffPeakDisconnectedSessionAfterSeconds         types.Int64  `tfsdk:"log_off_off_peak_disconnected_session_after_seconds"`
	LimitSecondsToForceLogOffUserDuringOffPeak           types.Int32  `tfsdk:"off_peak_limit_seconds_to_force_log_off_user"`
	LimitSecondsToForceLogOffUserDuringPeak              types.Int32  `tfsdk:"peak_limit_seconds_to_force_log_off_user"`
	LogOffWarningMessage                                 types.String `tfsdk:"log_off_warning_message"`
	LogOffWarningTitle                                   types.String `tfsdk:"log_off_warning_title"`
	PowerTimeSchemes                                     types.List   `tfsdk:"power_time_schemes"` //List[DeliveryGroupPowerTimeScheme]
	AutoscaleLogOffReminderEnabled                       types.Bool   `tfsdk:"log_off_reminder_enabled"`
	AutoscaleLogOffReminderIntervalSecondsOffPeak        types.Int32  `tfsdk:"off_peak_log_off_reminder_interval"`
	AutoscaleLogOffReminderIntervalSecondsPeak           types.Int32  `tfsdk:"peak_log_off_reminder_interval"`
	AutoscaleLogOffReminderMessage                       types.String `tfsdk:"log_off_reminder_message"`
	AutoscaleLogOffReminderTitle                         types.String `tfsdk:"log_off_reminder_title"`
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
			"restrict_autoscale_tag": schema.StringAttribute{
				Description: "Name of the tag on the machines that autoscale will apply on.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"peak_restrict_min_idle_untagged_percent": schema.Int32Attribute{
				Description: "Specifies the percentage of remaining untagged capacity to fall below to start powering on tagged machines during peak hours. " +
					"\n\n~> **Please Note** This setting is only applicable when the `restrict_autoscale_tag` is set.",
				Optional: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 100),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("restrict_autoscale_tag"), path.MatchRelative().AtParent().AtName("off_peak_restrict_min_idle_untagged_percent")),
				},
			},
			"off_peak_restrict_min_idle_untagged_percent": schema.Int32Attribute{
				Description: "Specifies the percentage of remaining untagged capacity to fall below to start powering on tagged machines during off peak hours. " +
					"\n\n~> **Please Note** This setting is only applicable when the `restrict_autoscale_tag` is set.",
				Optional: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 100),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("restrict_autoscale_tag"), path.MatchRelative().AtParent().AtName("peak_restrict_min_idle_untagged_percent")),
				},
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
				Description: "The action to be performed after a configurable period of a user session ending in peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					sessionHostingActionEnumValidator(),
				},
			},
			"peak_log_off_timeout_minutes": schema.Int64Attribute{
				Description: "The number of minutes before the configured action should be performed after a user session ends in peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"peak_disconnect_action": schema.StringAttribute{
				Description: "The action to be performed after a configurable period of a user session disconnecting in peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					sessionHostingActionEnumValidator(),
				},
			},
			"peak_extended_disconnect_action": schema.StringAttribute{
				Description: "The action to be performed after a second configurable period of a user session disconnecting in peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
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
				Description: "The action to be performed after a configurable period of a user session ending outside peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					sessionHostingActionEnumValidator(),
				},
			},
			"off_peak_log_off_timeout_minutes": schema.Int64Attribute{
				Description: "The number of minutes before the configured action should be performed after a user session ends outside peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"off_peak_disconnect_action": schema.StringAttribute{
				Description: "The action to be performed after a configurable period of a user session disconnecting outside peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					sessionHostingActionEnumValidator(),
				},
			},
			"off_peak_extended_disconnect_action": schema.StringAttribute{
				Description: "The action to be performed after a second configurable period of a user session disconnecting outside peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
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
				Description: "Delay before machines are powered off, when scaling down. Specified in minutes. " +
					"\n\n~> **Please Note** Applies only to multi-session machines. " +
					"\n\n-> **Note** By default, the power-off delay is 30 minutes. You can set it in a range of 0 to 60 minutes. ",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
				Validators: []validator.Int64{
					int64validator.Between(0, 60),
				},
			},
			"peak_autoscale_assigned_power_on_idle_action": schema.StringAttribute{
				Description: "The action to be performed on an assigned machine previously started by autoscale that subsequently remains unused. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					sessionHostingActionEnumValidator(),
				},
			},
			"peak_autoscale_assigned_power_on_idle_timeout_minutes": schema.Int64Attribute{
				Description: "The number of minutes before the configured action is performed on an assigned machine previously started by autoscale that subsequently remains unused.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
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
			"off_peak_limit_seconds_to_force_log_off_user": schema.Int32Attribute{
				Description: "Limit in seconds to force log off user after user logs off from their sessions during off-peak hours. Defaults to `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
			},
			"peak_limit_seconds_to_force_log_off_user": schema.Int32Attribute{
				Description: "Limit in seconds to force log off user after user logs off from their sessions during peak hours. Defaults to `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
			},
			"log_off_warning_message": schema.StringAttribute{
				Description: "The message to be displayed in the log off warning.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"log_off_warning_title": schema.StringAttribute{
				Description: "The title of the log off warning.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"power_time_schemes": schema.ListNestedAttribute{
				Description: "Power management time schemes." +
					"\n\n~> **Please Note** It is not allowed to have more than one power time scheme that cover the same day of the week for the same delivery group.",
				Optional:     true,
				NestedObject: DeliveryGroupPowerTimeScheme{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"log_off_reminder_enabled": schema.BoolAttribute{
				Description: "Indicates whether log off reminder is enabled. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"off_peak_log_off_reminder_interval": schema.Int32Attribute{
				Description: "The interval in seconds at which the log off reminder is sent during off-peak hours. Defaults to `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
			"peak_log_off_reminder_interval": schema.Int32Attribute{
				Description: "The interval in seconds at which the log off reminder is sent during peak hours. Defaults to `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
			"log_off_reminder_message": schema.StringAttribute{
				Description: "The message to be displayed in the log off reminder.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"log_off_reminder_title": schema.StringAttribute{
				Description: "The title of the log off reminder.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
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
	description := "Restrict access to this Delivery Group by specifying users and groups in the allow and block list. To give access to unauthenticated users, use the `allow_anonymous_access` property." +
		"\n\n~> **Please Note** If `restricted_access_users` attribute is omitted or set to `null`, all authenticated users will have access to this Delivery Group. If attribute is specified as an empty object i.e. `{}`, then no user will have access to the delivery group because `allow_list` and `block_list` will be set as empty sets by default."
	if !forDeliveryGroup {
		resource = "Desktop"
		description = "Restrict access to this Desktop by specifying users and groups in the allow and block list. " +
			"\n\n~> **Please Note** If `restricted_access_users` attribute is omitted or set to `null`, all authenticated users will have access to this Desktop. If attribute is specified as an empty object i.e. `{}`, then no user will have access to the desktop because `allow_list` and `block_list` will be set as empty sets by default." +
			"\n\n~> **Please Note** For Remote PC Delivery Groups desktops, `restricted_access_users` has to be set."
	}

	return schema.SingleNestedAttribute{Description: description,
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"allow_list": schema.SetAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who can use this %s. \n\n-> **Note** Users must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
				},
			},
			"block_list": schema.SetAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who cannot use this %s. A block list is meaningful only when used to block users in the allow list. \n\n-> **Note** Users must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
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
	RestrictToTag         types.String `tfsdk:"restrict_to_tag"`
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
			"restrict_to_tag": schema.StringAttribute{
				Description: "Restrict session launch to machines with tag specified in GUID.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Specify whether to enable the delivery of this desktop. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"enable_session_roaming": schema.BoolAttribute{
				Description: "When enabled, if the user launches this desktop and then moves to another device, the same session is used, and applications are available on both devices. When disabled, the session no longer roams between devices. " +
					"\n\n~> **Please Note** Session roaming should be set to `false` for Remote PC Delivery Group.",
				Optional: true,
			},
			"restricted_access_users": restrictedAccessUsers.GetSchema(),
		},
	}
}

func (DeliveryGroupDesktop) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupDesktop{}.GetSchema().Attributes
}

type DeliveryGroupAppProtection struct {
	ApplyContextually       types.List `tfsdk:"apply_contextually"` //DeliveryGroupAppProtectionApplyContextually
	EnableAntiKeyLogging    types.Bool `tfsdk:"enable_anti_key_logging"`
	EnableAntiScreenCapture types.Bool `tfsdk:"enable_anti_screen_capture"`
}

func (DeliveryGroupAppProtection) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "App Protection, an add-on feature for the Citrix Workspace app, provides enhanced security for Citrix published apps and desktops. The feature provides anti-keylogging and anti-screen capture capabilities for client sessions, helping protect data from keyloggers and screen scrapers." +
			"\n\n~> **Please Note** Before using the feature, make sure that these [requirements](https://docs.citrix.com/en-us/citrix-workspace-app/app-protection.html#system-requirements) are met.",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"enable_anti_key_logging": schema.BoolAttribute{
				Description: "When enabled, anti-keylogging is applied when a protected window is in focus.",
				Optional:    true,
				Validators: []validator.Bool{
					boolvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("enable_anti_screen_capture"),
					),
				},
			},
			"enable_anti_screen_capture": schema.BoolAttribute{
				Description: "Specify whether to use anti-screen capture." +
					"\n\n-> **Note** For Windows and macOS, only the window with protected content is blank. Anti-screen capture is only applied when the window is open. For Linux, the entire screen will appear blank. Anti-screen capture is only applied when the window is open or minimized.",
				Optional: true,
			},
			"apply_contextually": schema.ListNestedAttribute{
				Description:  "Implement contextual App Protection using the connection filters defined in the Access Policy rule.",
				Optional:     true,
				NestedObject: DeliveryGroupAppProtectionApplyContextuallyModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("enable_anti_key_logging"),
					),
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (DeliveryGroupAppProtection) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupAppProtection{}.GetSchema().Attributes
}

var _ util.RefreshableListItemWithAttributes[citrixorchestration.AdvancedAccessPolicyResponseModel] = DeliveryGroupAppProtectionApplyContextuallyModel{}

type DeliveryGroupAppProtectionApplyContextuallyModel struct {
	PolicyName              types.String `tfsdk:"policy_name"`
	EnableAntiKeyLogging    types.Bool   `tfsdk:"enable_anti_key_logging"`
	EnableAntiScreenCapture types.Bool   `tfsdk:"enable_anti_screen_capture"`
}

// GetKey implements util.RefreshableListItemWithAttributes.
func (r DeliveryGroupAppProtectionApplyContextuallyModel) GetKey() string {
	if strings.EqualFold(r.PolicyName.ValueString(), util.CitrixGatewayConnections) {
		return util.CitrixGatewayConnections
	}
	if strings.EqualFold(r.PolicyName.ValueString(), util.NonCitrixGatewayConnections) {
		return util.NonCitrixGatewayConnections
	}
	return r.PolicyName.ValueString()
}

func (DeliveryGroupAppProtectionApplyContextuallyModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"policy_name": schema.StringAttribute{
				Description: "The name of the policy." +
					"\n\n-> **Note** To refer to default policies, use `Citrix Gateway connections` as the name for the default policy that is Via Access Gateway and `Non-Citrix Gateway connections` as the name for the default policy that is Not Via Access Gateway.",
				Required: true,
			},
			"enable_anti_key_logging": schema.BoolAttribute{
				Description: "When enabled, anti-keylogging is applied when a protected window is in focus.",
				Required:    true,
			},
			"enable_anti_screen_capture": schema.BoolAttribute{
				Description: "Specify whether to use anti-screen capture." +
					"\n\n-> **Note** For Windows and macOS, only the window with protected content is blank. Anti-screen capture is only applied when the window is open. For Linux, the entire screen will appear blank. Anti-screen capture is only applied when the window is open or minimized.",
				Required: true,
			},
		},
	}
}

func (DeliveryGroupAppProtectionApplyContextuallyModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupAppProtectionApplyContextuallyModel{}.GetSchema().Attributes
}

var _ util.RefreshableListItemWithAttributes[citrixorchestration.AdvancedAccessPolicyResponseModel] = DeliveryGroupAccessPolicyModel{}

type DeliveryGroupAccessPolicyModel struct {
	Id                                  types.String `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	AllowedConnection                   types.String `tfsdk:"allowed_connection"`
	EnableCriteriaForIncludeConnections types.Bool   `tfsdk:"enable_criteria_for_include_connections"`
	IncludeConnectionsCriteriaType      types.String `tfsdk:"include_connections_criteria_type"`
	EnableCriteriaForExcludeConnections types.Bool   `tfsdk:"enable_criteria_for_exclude_connections"`
	IncludeCriteriaFilters              types.List   `tfsdk:"include_criteria_filters"` //List[DeliveryGroupAccessPolicyCriteriaTagsModel]
	ExcludeCriteriaFilters              types.List   `tfsdk:"exclude_criteria_filters"` //List[DeliveryGroupAccessPolicyCriteriaTagsModel]
}

func (r DeliveryGroupAccessPolicyModel) GetKey() string {
	if strings.EqualFold(r.Name.ValueString(), util.CitrixGatewayConnections) {
		return util.CitrixGatewayConnections
	}
	if strings.EqualFold(r.Name.ValueString(), util.NonCitrixGatewayConnections) {
		return util.NonCitrixGatewayConnections
	}
	return r.Name.ValueString()
}

func (DeliveryGroupAccessPolicyModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the resource location.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the access policy." +
					"\n\n-> **Note** For default_access_policies, use `Citrix Gateway connections` as the name for the policy that is Via Access Gateway and `Non-Citrix Gateway connections` as the name for the policy that is Not Via Access Gateway.",
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the access policy is enabled. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"allowed_connection": schema.StringAttribute{
				Description: "The behavior of the include filter. Choose between `Filtered`, `ViaAG`, and `NotViaAG`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Filtered",
						"ViaAG",
						"NotViaAG",
					),
				},
			},
			"enable_criteria_for_include_connections": schema.BoolAttribute{
				Description: "Whether to enable criteria for include connections.",
				Required:    true,
				Validators: []validator.Bool{
					validators.AlsoRequiresOnBoolValues(
						[]bool{true},
						path.MatchRelative().AtParent().AtName("include_connections_criteria_type"),
					),
				},
			},
			"include_connections_criteria_type": schema.StringAttribute{
				Description: "The type of criteria for include connections. Choose between `MatchAny` and `MatchAll`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"MatchAny",
						"MatchAll",
					),
				},
			},
			"include_criteria_filters": schema.ListNestedAttribute{
				Description:  "The list of filters that meet the criteria for include connections.",
				Optional:     true,
				NestedObject: DeliveryGroupAccessPolicyCriteriaTagsModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"enable_criteria_for_exclude_connections": schema.BoolAttribute{
				Description: "Whether to enable criteria for exclude connections.",
				Required:    true,
			},
			"exclude_criteria_filters": schema.ListNestedAttribute{
				Description:  "The list of filters that meet the criteria for exclude connections.",
				Optional:     true,
				NestedObject: DeliveryGroupAccessPolicyCriteriaTagsModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (DeliveryGroupAccessPolicyModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupAccessPolicyModel{}.GetSchema().Attributes
}

var _ util.RefreshableListItemWithAttributes[citrixorchestration.SmartAccessTagResponseModel] = DeliveryGroupAccessPolicyCriteriaTagsModel{}

type DeliveryGroupAccessPolicyCriteriaTagsModel struct {
	FilterName  types.String `tfsdk:"filter_name"`
	FilterValue types.String `tfsdk:"filter_value"`
}

func (r DeliveryGroupAccessPolicyCriteriaTagsModel) GetKey() string {
	return r.FilterName.ValueString() + r.FilterValue.ValueString()
}

func (DeliveryGroupAccessPolicyCriteriaTagsModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"filter_name": schema.StringAttribute{
				Description: "The name of the filter.",
				Required:    true,
			},
			"filter_value": schema.StringAttribute{
				Description: "The value of the filter.",
				Required:    true,
			},
		},
	}
}

func (DeliveryGroupAccessPolicyCriteriaTagsModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupAccessPolicyCriteriaTagsModel{}.GetSchema().Attributes
}

// DeliveryGroupResourceModel maps the resource schema data.
type DeliveryGroupResourceModel struct {
	Id                          types.String `tfsdk:"id"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	Name                        types.String `tfsdk:"name"`
	Description                 types.String `tfsdk:"description"`
	DeliveryType                types.String `tfsdk:"delivery_type"`
	SessionSupport              types.String `tfsdk:"session_support"`
	SharingKind                 types.String `tfsdk:"sharing_kind"`
	RestrictedAccessUsers       types.Object `tfsdk:"restricted_access_users"`
	AllowAnonymousAccess        types.Bool   `tfsdk:"allow_anonymous_access"`
	Desktops                    types.List   `tfsdk:"desktops"`                    // List[DeliveryGroupDesktop]
	AssociatedMachineCatalogs   types.Set    `tfsdk:"associated_machine_catalogs"` // List[DeliveryGroupMachineCatalogModel]
	AutoscaleSettings           types.Object `tfsdk:"autoscale_settings"`          // DeliveryGroupPowerManagementSettings
	RebootSchedules             types.List   `tfsdk:"reboot_schedules"`            // List[DeliveryGroupRebootSchedule]
	TotalMachines               types.Int64  `tfsdk:"total_machines"`
	MinimumFunctionalLevel      types.String `tfsdk:"minimum_functional_level"`
	StoreFrontServers           types.Set    `tfsdk:"storefront_servers"` //Set[string]
	Scopes                      types.Set    `tfsdk:"scopes"`             //Set[String]
	BuiltInScopes               types.Set    `tfsdk:"built_in_scopes"`    //Set[String]
	InheritedScopes             types.Set    `tfsdk:"inherited_scopes"`   //Set[String]
	MakeResourcesAvailableInLHC types.Bool   `tfsdk:"make_resources_available_in_lhc"`
	AppProtection               types.Object `tfsdk:"app_protection"`          // DeliveryGroupAppProtection
	DefaultAccessPolicies       types.List   `tfsdk:"default_access_policies"` // List[DeliveryGroupAccessPolicyModel]
	CustomAccessPolicies        types.List   `tfsdk:"custom_access_policies"`  // List[DeliveryGroupAccessPolicyModel]
	DeliveryGroupFolderPath     types.String `tfsdk:"delivery_group_folder_path"`
	Tenants                     types.Set    `tfsdk:"tenants"`  // Set[String]
	Metadata                    types.List   `tfsdk:"metadata"` // List[NameValueStringPairmodel]
	Tags                        types.Set    `tfsdk:"tags"`     // Set[string]
	DefaultDesktopIcon          types.String `tfsdk:"default_desktop_icon"`
}

func (DeliveryGroupResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a delivery group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the delivery group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the delivery group is enabled or not. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
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
			"delivery_type": schema.StringAttribute{
				Description: "Delivery type of the delivery group. Available values are `DesktopsOnly`, `AppsOnly`, and `DesktopsAndApps`. Defaults to `DesktopsOnly` for Delivery Groups with associated Machine Catalogs that have `allocation_type` set to `Static` and for Delivery Groups that have `sharing_kind` set to `private`. Otherwise defaults to `DesktopsAndApps",
				Optional:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedDeliveryKindEnumValues),
				},
			},
			"session_support": schema.StringAttribute{
				Description: "The session support for the delivery group. Can only be set to `SingleSession` or `MultiSession`. Specify only if you want to create a Delivery Group without any `associated_machine_catalogs`. Ensure session support is same as that of the prospective Machine Catalogs you will associate this Delivery Group with.",
				Optional:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionSupportEnumValues),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("sharing_kind"),
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
				Description: "Give access to unauthenticated (anonymous) users. When set to `True`, no credentials are required to access StoreFront. " +
					"\n\n~> **Please Note** This feature requires a StoreFront store for unauthenticated users.",
				Optional: true,
			},
			"desktops": schema.ListNestedAttribute{
				Description:  "A list of Desktop resources to publish on the delivery group. Only 1 desktop can be added to a Remote PC Delivery Group.",
				Optional:     true,
				NestedObject: DeliveryGroupDesktop{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"associated_machine_catalogs": schema.SetNestedAttribute{
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
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"built_in_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the built-in scopes of the delivery group.",
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"inherited_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the inherited scopes of the delivery group.",
				Computed:    true,
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
					"\n\n~> **Please Note** This setting only impacts Single Session OS Random (pooled) desktops which are power managed. LHC is always enabled for Single Session OS static and Multi Session OS desktops." +
					"\n\n-> **Note** When set to `true`, machines will remain available and allow new connections and changes to the machine caused by a user might be present in subsequent sessions. " +
					"When set to `false`, machines in the delivery group will be unavailable for new connections during a Local Host Cache event. ",
				Optional: true,
			},
			"app_protection": DeliveryGroupAppProtection{}.GetSchema(),
			"default_access_policies": schema.ListNestedAttribute{
				Description: "Manage built-in Access Policies for the delivery group. These are the Citrix Gateway Connections (via Access Gateway) and Non-Citrix Gateway Connections (not via Access Gateway) access policies." +
					"\n\n~> **Please Note** Default Access Policies can only be modified; they cannot be deleted. If using this property, both default policies have to be specified." +
					"\n\n-> **Note** Use `Citrix Gateway connections` as the name for the default policy that is Via Access Gateway and `Non-Citrix Gateway connections` as the name for the default policy that is Not Via Access Gateway.",
				Optional:     true,
				NestedObject: DeliveryGroupAccessPolicyModel{}.GetSchema(),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.StateValue.IsNull() && req.ConfigValue.IsNull()
					},
						"Force replacement when discarding changes in default access policies",
						"Force replacement when discarding changes in default access policies"),
				},
			},
			"custom_access_policies": schema.ListNestedAttribute{
				Description:  "Custom Access Policies for the delivery group. To manage built-in access policies use the `default_access_policies` instead.",
				Optional:     true,
				NestedObject: DeliveryGroupAccessPolicyModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"delivery_group_folder_path": schema.StringAttribute{
				Description: "The path of the folder in which the delivery group is located.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the delivery group.",
				Computed:    true,
			},
			"metadata": util.GetMetadataListSchema("Delivery Group"),
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the delivery group.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"default_desktop_icon": schema.StringAttribute{
				Description: "The id of the icon to be used as the default icon for the desktops in the delivery group." +
					"\n\n~> **Please Note** This option is only supported for Citrix Cloud Customer",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("1"),
			},
		},
	}
}

func (DeliveryGroupResourceModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupResourceModel{}.GetSchema().Attributes
}

func (r DeliveryGroupResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixclient.CitrixDaasClient, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgDesktops *citrixorchestration.DesktopResponseModelCollection, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection, dgMachines []citrixorchestration.MachineResponseModel, dgRebootSchedule *citrixorchestration.RebootScheduleResponseModelCollection, tags []string) DeliveryGroupResourceModel {

	// Set required values
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())
	r.TotalMachines = types.Int64Value(int64(deliveryGroup.GetTotalMachines()))
	r.Description = types.StringValue(deliveryGroup.GetDescription())

	// Set optional values
	r.Enabled = types.BoolValue(deliveryGroup.GetEnabled())
	if !r.DeliveryType.IsNull() {
		r.DeliveryType = types.StringValue(string(deliveryGroup.GetDeliveryType()))
	} else {
		r.DeliveryType = types.StringNull()
	}

	minimumFunctionalLevel := deliveryGroup.GetMinimumFunctionalLevel()
	r.MinimumFunctionalLevel = types.StringValue(string(minimumFunctionalLevel))

	parentList := []string{}
	associatedCatalogs := util.ObjectSetToTypedArray[DeliveryGroupMachineCatalogModel](ctx, diagnostics, r.AssociatedMachineCatalogs)
	for _, machineCatalog := range associatedCatalogs {
		parentList = append(parentList, machineCatalog.MachineCatalog.ValueString())
	}
	scopeIdsInPlan := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds, builtInScopes, inheritedScopeIds, err := util.CategorizeScopes(ctx, client, diagnostics, deliveryGroup.GetScopes(), citrixorchestration.SCOPEDOBJECTTYPE_MACHINE_CATALOG, parentList, scopeIdsInPlan)
	if err != nil {
		return r
	}
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)
	r.BuiltInScopes = util.StringArrayToStringSet(ctx, diagnostics, builtInScopes)
	r.InheritedScopes = util.StringArrayToStringSet(ctx, diagnostics, inheritedScopeIds)

	if deliveryGroup.GetReuseMachinesWithoutShutdownInOutage() {
		r.MakeResourcesAvailableInLHC = types.BoolValue(true)
	} else if !r.MakeResourcesAvailableInLHC.IsNull() {
		r.MakeResourcesAvailableInLHC = types.BoolValue(false)
	}

	if len(dgMachines) < 1 || !r.SessionSupport.IsNull() {
		r.SessionSupport = types.StringValue(string(deliveryGroup.GetSessionSupport()))
		r.SharingKind = types.StringValue(string(deliveryGroup.GetSharingKind()))
	}

	r = r.updatePlanWithRestrictedAccessUsers(ctx, diagnostics, deliveryGroup)
	r = r.updatePlanWithDesktops(ctx, diagnostics, dgDesktops)
	r = r.updatePlanWithAssociatedCatalogs(ctx, diagnostics, dgMachines)
	r = r.updatePlanWithAutoscaleSettings(ctx, diagnostics, deliveryGroup, dgPowerTimeSchemes)
	r = r.updatePlanWithRebootSchedule(ctx, diagnostics, dgRebootSchedule)
	r = r.updatePlanWithAppProtection(ctx, diagnostics, deliveryGroup)

	var defaultAccessPolicies []citrixorchestration.AdvancedAccessPolicyResponseModel
	var customAccessPolicies []citrixorchestration.AdvancedAccessPolicyResponseModel
	for _, policy := range deliveryGroup.GetAdvancedAccessPolicy() {
		if policy.GetIsBuiltIn() {
			defaultAccessPolicies = append(defaultAccessPolicies, policy)
		} else {
			customAccessPolicies = append(customAccessPolicies, policy)
		}
	}
	r = r.updatePlanWithDefaultAccessPolicies(ctx, diagnostics, defaultAccessPolicies)
	r = r.updatePlanWithCustomAccessPolicies(ctx, diagnostics, customAccessPolicies)

	if len(deliveryGroup.GetStoreFrontServersForHostedReceiver()) > 0 || !r.StoreFrontServers.IsNull() {
		var remoteAssociatedStoreFrontServers []string
		for _, server := range deliveryGroup.GetStoreFrontServersForHostedReceiver() {
			remoteAssociatedStoreFrontServers = append(remoteAssociatedStoreFrontServers, server.GetId())
		}
		r.StoreFrontServers = util.StringArrayToStringSet(ctx, diagnostics, remoteAssociatedStoreFrontServers)
	} else {
		r.StoreFrontServers = types.SetNull(types.StringType)
	}

	adminFolder := deliveryGroup.GetAdminFolder()
	adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
	if adminFolderPath != "" {
		r.DeliveryGroupFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.DeliveryGroupFolderPath = types.StringNull()
	}

	r.DefaultDesktopIcon = types.StringValue(deliveryGroup.GetDefaultDesktopIconId())

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, deliveryGroup.GetTenants())

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), deliveryGroup.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

	return r
}
