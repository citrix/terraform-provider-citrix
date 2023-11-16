package models

import (
	"reflect"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
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

type DeliveryGroupPowerManagementSettings struct {
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

// DeliveryGroupResourceModel maps the resource schema data.
type DeliveryGroupResourceModel struct {
	Id                        types.String                          `tfsdk:"id"`
	Name                      types.String                          `tfsdk:"name"`
	Description               types.String                          `tfsdk:"description"`
	AssociatedMachineCatalogs []DeliveryGroupMachineCatalogModel    `tfsdk:"associated_machine_catalogs"`
	Users                     []types.String                        `tfsdk:"users"`
	AutoscaleEnabled          types.Bool                            `tfsdk:"autoscale_enabled"`
	AutoscaleSettings         *DeliveryGroupPowerManagementSettings `tfsdk:"autoscale_settings"`
	TotalMachines             types.Int64                           `tfsdk:"total_machines"`
}

func (r DeliveryGroupResourceModel) RefreshPropertyValues(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgDesktops *citrixorchestration.DesktopResponseModelCollection, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection, dgMachines *citrixorchestration.MachineResponseModelCollection) DeliveryGroupResourceModel {

	// Set required values
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())
	r.AutoscaleEnabled = types.BoolValue(deliveryGroup.GetAutoScaleEnabled())
	r.TotalMachines = types.Int64Value(int64(deliveryGroup.GetTotalMachines()))

	// Set optional values
	if deliveryGroup.GetDescription() != "" {
		r.Description = types.StringValue(deliveryGroup.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	if r.AutoscaleSettings == nil {
		return r
	}

	r.AutoscaleSettings.Timezone = types.StringValue(deliveryGroup.GetTimeZone())
	r.AutoscaleSettings.PeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakDisconnectTimeoutMinutes()))
	r.AutoscaleSettings.PeakLogOffAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakLogOffAction()).String())
	r.AutoscaleSettings.PeakDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakDisconnectAction()).String())
	r.AutoscaleSettings.PeakExtendedDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakExtendedDisconnectAction()).String())
	r.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakExtendedDisconnectTimeoutMinutes()))
	r.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakDisconnectTimeoutMinutes()))
	r.AutoscaleSettings.OffPeakLogOffAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakLogOffAction()).String())
	r.AutoscaleSettings.OffPeakDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakExtendedDisconnectAction()).String())
	r.AutoscaleSettings.OffPeakExtendedDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakExtendedDisconnectAction()).String())
	r.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakExtendedDisconnectTimeoutMinutes()))
	r.AutoscaleSettings.PeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetPeakBufferSizePercent()))
	r.AutoscaleSettings.OffPeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetOffPeakBufferSizePercent()))
	r.AutoscaleSettings.PowerOffDelayMinutes = types.Int64Value(int64(deliveryGroup.GetPowerOffDelayMinutes()))
	r.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectPeakIdleSessionAfterSeconds()))
	r.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectOffPeakIdleSessionAfterSeconds()))
	r.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffPeakDisconnectedSessionAfterSeconds()))
	r.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffOffPeakDisconnectedSessionAfterSeconds()))

	parsedPowerTimeSchemes := ParsePowerTimeSchemesClientToPluginModel(dgPowerTimeSchemes.GetItems())
	r.AutoscaleSettings.PowerTimeSchemes = parsedPowerTimeSchemes

	r = r.updatePlanWithDesktops(dgDesktops)
	r = r.updatePlanWithAssociatedCatalogs(dgMachines)

	return r
}

func ParsePowerTimeSchemesPluginToClientModel(powerTimeSchemes []DeliveryGroupPowerTimeScheme) []citrixorchestration.PowerTimeSchemeRequestModel {
	var res []citrixorchestration.PowerTimeSchemeRequestModel
	for _, powerTimeScheme := range powerTimeSchemes {
		var powerTimeSchemeRequest citrixorchestration.PowerTimeSchemeRequestModel

		var daysOfWeek []citrixorchestration.TimeSchemeDays
		for _, dayOfWeek := range powerTimeScheme.DaysOfWeek {
			timeSchemeDay := getTimeSchemeDayValue(dayOfWeek.ValueString())
			daysOfWeek = append(daysOfWeek, timeSchemeDay)
		}

		var poolSizeScheduleRequests []citrixorchestration.PoolSizeScheduleRequestModel
		for _, poolSizeSchedule := range powerTimeScheme.PoolSizeSchedule {
			var poolSizeScheduleRequest citrixorchestration.PoolSizeScheduleRequestModel
			poolSizeScheduleRequest.SetTimeRange(poolSizeSchedule.TimeRange.ValueString())
			poolSizeScheduleRequest.SetPoolSize(int32(poolSizeSchedule.PoolSize.ValueInt64()))
			poolSizeScheduleRequests = append(poolSizeScheduleRequests, poolSizeScheduleRequest)
		}

		peakTimeRanges := util.ConvertBaseStringArrayToPrimitiveStringArray(powerTimeScheme.PeakTimeRanges)

		powerTimeSchemeRequest.SetDisplayName(powerTimeScheme.DisplayName.ValueString())
		powerTimeSchemeRequest.SetPeakTimeRanges(peakTimeRanges)
		powerTimeSchemeRequest.SetPoolUsingPercentage(powerTimeScheme.PoolUsingPercentage.ValueBool())
		powerTimeSchemeRequest.SetDaysOfWeek(daysOfWeek)
		powerTimeSchemeRequest.SetPoolSizeSchedule(poolSizeScheduleRequests)
		res = append(res, powerTimeSchemeRequest)
	}

	return res
}

func ParsePowerTimeSchemesClientToPluginModel(powerTimeSchemesResponse []citrixorchestration.PowerTimeSchemeResponseModel) []DeliveryGroupPowerTimeScheme {
	var res []DeliveryGroupPowerTimeScheme
	for _, powerTimeSchemeResponse := range powerTimeSchemesResponse {
		var deliveryGroupPowerTimeScheme DeliveryGroupPowerTimeScheme

		var daysOfWeek []types.String
		for _, dayOfWeek := range powerTimeSchemeResponse.GetDaysOfWeek() {
			timeSchemeDay := types.StringValue(reflect.ValueOf(dayOfWeek).String())
			daysOfWeek = append(daysOfWeek, timeSchemeDay)
		}

		var poolSizeScheduleRequests []PowerTimeSchemePoolSizeScheduleRequestModel
		for _, poolSizeSchedule := range powerTimeSchemeResponse.GetPoolSizeSchedule() {
			var poolSizeScheduleRequest PowerTimeSchemePoolSizeScheduleRequestModel
			poolSizeScheduleRequest.TimeRange = types.StringValue(poolSizeSchedule.GetTimeRange())
			poolSizeScheduleRequest.PoolSize = types.Int64Value(int64(poolSizeSchedule.GetPoolSize()))
			poolSizeScheduleRequests = append(poolSizeScheduleRequests, poolSizeScheduleRequest)
		}

		deliveryGroupPowerTimeScheme.DisplayName = types.StringValue(powerTimeSchemeResponse.GetDisplayName())
		peakTimeRanges := util.ConvertPrimitiveStringArrayToBaseStringArray(powerTimeSchemeResponse.GetPeakTimeRanges())
		deliveryGroupPowerTimeScheme.PeakTimeRanges = peakTimeRanges
		deliveryGroupPowerTimeScheme.PoolUsingPercentage = types.BoolValue(powerTimeSchemeResponse.GetPoolUsingPercentage())
		deliveryGroupPowerTimeScheme.DaysOfWeek = daysOfWeek
		deliveryGroupPowerTimeScheme.PoolSizeSchedule = poolSizeScheduleRequests

		res = append(res, deliveryGroupPowerTimeScheme)
	}

	return res
}

func getTimeSchemeDayValue(v string) citrixorchestration.TimeSchemeDays {
	timeSchemeDay, err := citrixorchestration.NewTimeSchemeDaysFromValue(v)
	if err != nil {
		return citrixorchestration.TIMESCHEMEDAYS_UNKNOWN
	}

	return *timeSchemeDay
}

func (r DeliveryGroupResourceModel) updatePlanWithAssociatedCatalogs(machines *citrixorchestration.MachineResponseModelCollection) DeliveryGroupResourceModel {
	machineCatalogMap := map[string]int{}

	for _, machine := range machines.GetItems() {
		machineCatalog := machine.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		machineCatalogMap[machineCatalogId] += 1
	}

	r.AssociatedMachineCatalogs = []DeliveryGroupMachineCatalogModel{}
	for key, val := range machineCatalogMap {
		var deliveryGroupMachineCatalogModel DeliveryGroupMachineCatalogModel
		deliveryGroupMachineCatalogModel.MachineCatalog = types.StringValue(key)
		deliveryGroupMachineCatalogModel.MachineCount = types.Int64Value(int64(val))
		r.AssociatedMachineCatalogs = append(r.AssociatedMachineCatalogs, deliveryGroupMachineCatalogModel)
	}

	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithDesktops(deliveryGroupDesktops *citrixorchestration.DesktopResponseModelCollection) DeliveryGroupResourceModel {
	dgDesktops := deliveryGroupDesktops.GetItems()
	if len(dgDesktops) == 0 {
		if r.Users != nil {
			// If plan has empty list
			r.Users = []types.String{}
		}
	} else {
		dgDesktop := dgDesktops[0]
		includedUsers := getPrincipalNameForUserFromDeliveryGroupDesktop(dgDesktop)
		if len(includedUsers) == 0 {
			if r.Users != nil {
				// If plan has empty list
				r.Users = []types.String{}
			}
		} else {
			r.Users = util.ConvertPrimitiveStringArrayToBaseStringArray(includedUsers)
		}
	}

	return r
}

func getPrincipalNameForUserFromDeliveryGroupDesktop(desktop citrixorchestration.DesktopResponseModel) []string {
	var res []string
	for _, user := range desktop.GetIncludedUsers() {
		res = append(res, user.GetPrincipalName())
	}
	return res
}
