// Copyright © 2023. Citrix Systems, Inc.

package delivery_group

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeliveryGroupMachineCatalogModel struct {
	MachineCatalog types.String `tfsdk:"machine_catalog"`
	MachineCount   types.Int64  `tfsdk:"machine_count"`
}

type PowerTimeSchemePoolSizeScheduleRequestModel struct {
	TimeRange types.String `tfsdk:"time_range"`
	PoolSize  types.Int64  `tfsdk:"pool_size"`
}

type DeliveryGroupPowerTimeScheme struct {
	DaysOfWeek          []types.String                                `tfsdk:"days_of_week"`
	DisplayName         types.String                                  `tfsdk:"display_name"`
	PeakTimeRanges      []types.String                                `tfsdk:"peak_time_ranges"`
	PoolSizeSchedule    []PowerTimeSchemePoolSizeScheduleRequestModel `tfsdk:"pool_size_schedules"`
	PoolUsingPercentage types.Bool                                    `tfsdk:"pool_using_percentage"`
}

type DeliveryGroupRebootNotificationToUsers struct {
	NotificationDurationMinutes     types.Int64  `tfsdk:"notification_duration_minutes"`
	NotificationMessage             types.String `tfsdk:"notification_message"`
	NotificationRepeatEvery5Minutes types.Bool   `tfsdk:"notification_repeat_every_5_minutes"`
	NotificationTitle               types.String `tfsdk:"notification_title"`
}

type DeliveryGroupRebootSchedule struct {
	Name                                   types.String                            `tfsdk:"name"`
	Description                            types.String                            `tfsdk:"description"`
	RebootScheduleEnabled                  types.Bool                              `tfsdk:"reboot_schedule_enabled"`
	RestrictToTag                          types.String                            `tfsdk:"restrict_to_tag"`
	IgnoreMaintenanceMode                  types.Bool                              `tfsdk:"ignore_maintenance_mode"`
	Frequency                              types.String                            `tfsdk:"frequency"`
	FrequencyFactor                        types.Int64                             `tfsdk:"frequency_factor"`
	StartDate                              types.String                            `tfsdk:"start_date"`
	StartTime                              types.String                            `tfsdk:"start_time"`
	RebootDurationMinutes                  types.Int64                             `tfsdk:"reboot_duration_minutes"`
	UseNaturalRebootSchedule               types.Bool                              `tfsdk:"natural_reboot_schedule"`
	DaysInWeek                             []types.String                          `tfsdk:"days_in_week"`
	WeekInMonth                            types.String                            `tfsdk:"week_in_month"`
	DayInMonth                             types.String                            `tfsdk:"day_in_month"`
	DeliveryGroupRebootNotificationToUsers *DeliveryGroupRebootNotificationToUsers `tfsdk:"reboot_notification_to_users"`
}

type DeliveryGroupPowerManagementSettings struct {
	AutoscaleEnabled                             types.Bool                     `tfsdk:"autoscale_enabled"`
	Timezone                                     types.String                   `tfsdk:"timezone"`
	PeakDisconnectTimeoutMinutes                 types.Int64                    `tfsdk:"peak_disconnect_timeout_minutes"`
	PeakLogOffAction                             types.String                   `tfsdk:"peak_log_off_action"`
	PeakDisconnectAction                         types.String                   `tfsdk:"peak_disconnect_action"`
	PeakExtendedDisconnectAction                 types.String                   `tfsdk:"peak_extended_disconnect_action"`
	PeakExtendedDisconnectTimeoutMinutes         types.Int64                    `tfsdk:"peak_extended_disconnect_timeout_minutes"`
	OffPeakDisconnectTimeoutMinutes              types.Int64                    `tfsdk:"off_peak_disconnect_timeout_minutes"`
	OffPeakLogOffAction                          types.String                   `tfsdk:"off_peak_log_off_action"`
	OffPeakDisconnectAction                      types.String                   `tfsdk:"off_peak_disconnect_action"`
	OffPeakExtendedDisconnectAction              types.String                   `tfsdk:"off_peak_extended_disconnect_action"`
	OffPeakExtendedDisconnectTimeoutMinutes      types.Int64                    `tfsdk:"off_peak_extended_disconnect_timeout_minutes"`
	PeakBufferSizePercent                        types.Int64                    `tfsdk:"peak_buffer_size_percent"`
	OffPeakBufferSizePercent                     types.Int64                    `tfsdk:"off_peak_buffer_size_percent"`
	PowerOffDelayMinutes                         types.Int64                    `tfsdk:"power_off_delay_minutes"`
	DisconnectPeakIdleSessionAfterSeconds        types.Int64                    `tfsdk:"disconnect_peak_idle_session_after_seconds"`
	DisconnectOffPeakIdleSessionAfterSeconds     types.Int64                    `tfsdk:"disconnect_off_peak_idle_session_after_seconds"`
	LogoffPeakDisconnectedSessionAfterSeconds    types.Int64                    `tfsdk:"log_off_peak_disconnected_session_after_seconds"`
	LogoffOffPeakDisconnectedSessionAfterSeconds types.Int64                    `tfsdk:"log_off_off_peak_disconnected_session_after_seconds"`
	PowerTimeSchemes                             []DeliveryGroupPowerTimeScheme `tfsdk:"power_time_schemes"`
}

type RestrictedAccessUsers struct {
	AllowList []types.String `tfsdk:"allow_list"`
	BlockList []types.String `tfsdk:"block_list"`
}

type DeliveryGroupDesktop struct {
	PublishedName         types.String           `tfsdk:"published_name"`
	DesktopDescription    types.String           `tfsdk:"description"`
	Enabled               types.Bool             `tfsdk:"enabled"`
	EnableSessionRoaming  types.Bool             `tfsdk:"enable_session_roaming"`
	RestrictedAccessUsers *RestrictedAccessUsers `tfsdk:"restricted_access_users"`
}

// DeliveryGroupResourceModel maps the resource schema data.
type DeliveryGroupResourceModel struct {
	Id                        types.String                          `tfsdk:"id"`
	Name                      types.String                          `tfsdk:"name"`
	Description               types.String                          `tfsdk:"description"`
	RestrictedAccessUsers     *RestrictedAccessUsers                `tfsdk:"restricted_access_users"`
	AllowAnonymousAccess      types.Bool                            `tfsdk:"allow_anonymous_access"`
	Desktops                  []DeliveryGroupDesktop                `tfsdk:"desktops"`
	AssociatedMachineCatalogs []DeliveryGroupMachineCatalogModel    `tfsdk:"associated_machine_catalogs"`
	AutoscaleSettings         *DeliveryGroupPowerManagementSettings `tfsdk:"autoscale_settings"`
	RebootSchedules           []DeliveryGroupRebootSchedule         `tfsdk:"reboot_schedules"`
	TotalMachines             types.Int64                           `tfsdk:"total_machines"`
	PolicySetId               types.String                          `tfsdk:"policy_set_id"`
}

func (r DeliveryGroupResourceModel) RefreshPropertyValues(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgDesktops *citrixorchestration.DesktopResponseModelCollection, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection, dgMachines *citrixorchestration.MachineResponseModelCollection, dgRebootSchedule *citrixorchestration.RebootScheduleResponseModelCollection) DeliveryGroupResourceModel {

	// Set required values
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())
	r.TotalMachines = types.Int64Value(int64(deliveryGroup.GetTotalMachines()))

	// Set optional values
	if deliveryGroup.GetDescription() != "" {
		r.Description = types.StringValue(deliveryGroup.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	if deliveryGroup.GetPolicySetGuid() != "" {
		r.PolicySetId = types.StringValue(deliveryGroup.GetPolicySetGuid())
	} else {
		r.PolicySetId = types.StringNull()
	}

	r = r.updatePlanWithRestrictedAccessUsers(deliveryGroup)
	r = r.updatePlanWithDesktops(dgDesktops)
	r = r.updatePlanWithAssociatedCatalogs(dgMachines)
	r = r.updatePlanWithAutoscaleSettings(deliveryGroup, dgPowerTimeSchemes)
	r = r.updatePlanWithRebootSchedule(dgRebootSchedule)
	return r
}
