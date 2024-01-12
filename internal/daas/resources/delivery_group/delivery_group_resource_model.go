// Copyright © 2023. Citrix Systems, Inc.

package delivery_group

import (
	"reflect"
	"strings"

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
	Desktops                  *[]DeliveryGroupDesktop               `tfsdk:"desktops"`
	AssociatedMachineCatalogs []DeliveryGroupMachineCatalogModel    `tfsdk:"associated_machine_catalogs"`
	AutoscaleSettings         *DeliveryGroupPowerManagementSettings `tfsdk:"autoscale_settings"`
	TotalMachines             types.Int64                           `tfsdk:"total_machines"`
}

func (r DeliveryGroupResourceModel) RefreshPropertyValues(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgDesktops *citrixorchestration.DesktopResponseModelCollection, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection, dgMachines *citrixorchestration.MachineResponseModelCollection) DeliveryGroupResourceModel {

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

	r = r.updatePlanWithRestrictedAccessUsers(deliveryGroup)
	r = r.updatePlanWithDesktops(dgDesktops)
	r = r.updatePlanWithAssociatedCatalogs(dgMachines)
	r = r.updatePlanWithAutoscaleSettings(deliveryGroup, dgPowerTimeSchemes)

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

func ParseDeliveryGroupDesktopsToClientModel(deliveryGroupDesktops *[]DeliveryGroupDesktop) []citrixorchestration.DesktopRequestModel {
	desktopRequests := []citrixorchestration.DesktopRequestModel{}

	if deliveryGroupDesktops == nil {
		return desktopRequests
	}

	for _, deliveryGroupDesktop := range *deliveryGroupDesktops {
		var desktopRequest citrixorchestration.DesktopRequestModel
		desktopRequest.SetPublishedName(deliveryGroupDesktop.PublishedName.ValueString())
		desktopRequest.SetDescription(deliveryGroupDesktop.DesktopDescription.ValueString())
		sessionReconnection := citrixorchestration.SESSIONRECONNECTION_ALWAYS
		if !deliveryGroupDesktop.EnableSessionRoaming.ValueBool() {
			sessionReconnection = citrixorchestration.SESSIONRECONNECTION_SAME_ENDPOINT_ONLY
		}
		desktopRequest.SetEnabled(deliveryGroupDesktop.Enabled.ValueBool())
		desktopRequest.SetSessionReconnection(sessionReconnection)

		includedUsers := []string{}
		excludedUsers := []string{}
		includedUsersFilterEnabled := false
		excludedUsersFilterEnabled := false
		if deliveryGroupDesktop.RestrictedAccessUsers != nil {
			includedUsersFilterEnabled = true
			includedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(deliveryGroupDesktop.RestrictedAccessUsers.AllowList)

			if deliveryGroupDesktop.RestrictedAccessUsers.BlockList != nil {
				excludedUsersFilterEnabled = true
				excludedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(deliveryGroupDesktop.RestrictedAccessUsers.BlockList)
			}
		}

		desktopRequest.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
		desktopRequest.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
		desktopRequest.SetIncludedUsers(includedUsers)
		desktopRequest.SetExcludedUsers(excludedUsers)

		desktopRequests = append(desktopRequests, desktopRequest)
	}

	return desktopRequests
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
		if r.Desktops != nil && len(*r.Desktops) > 0 {
			r.Desktops = nil
		}
		return r
	}

	if r.Desktops == nil {
		r.Desktops = &[]DeliveryGroupDesktop{}
	}

	existingDesktopsMap := map[string]int{}

	for index, desktop := range *r.Desktops {
		existingDesktopsMap[strings.ToLower(desktop.PublishedName.ValueString())] = index
	}

	existingDesktops := *r.Desktops
	for _, desktop := range dgDesktops {
		index, exists := existingDesktopsMap[strings.ToLower(desktop.GetPublishedName())]
		existingDesktop := DeliveryGroupDesktop{}
		if exists {
			existingDesktop = existingDesktops[index]
		}

		existingDesktop.PublishedName = types.StringValue(desktop.GetPublishedName())
		if desktop.GetDescription() != "" {
			existingDesktop.DesktopDescription = types.StringValue(desktop.GetDescription())
		} else {
			existingDesktop.DesktopDescription = types.StringNull()
		}
		existingDesktop.Enabled = types.BoolValue(desktop.GetEnabled())
		sessionReconnection := desktop.GetSessionReconnection()
		if sessionReconnection == citrixorchestration.SESSIONRECONNECTION_ALWAYS {
			existingDesktop.EnableSessionRoaming = types.BoolValue(true)
		} else {
			existingDesktop.EnableSessionRoaming = types.BoolValue(false)
		}

		if !desktop.GetIncludedUserFilterEnabled() {
			existingDesktop.RestrictedAccessUsers = nil
			if exists {
				existingDesktops[index] = existingDesktop
			} else {
				existingDesktops = append(existingDesktops, existingDesktop)
			}
			existingDesktopsMap[strings.ToLower(desktop.GetPublishedName())] = -1 // visited
			continue
		}

		if existingDesktop.RestrictedAccessUsers == nil {
			existingDesktop.RestrictedAccessUsers = &RestrictedAccessUsers{}
		}

		includedUsers := desktop.GetIncludedUsers()
		excludedUsers := desktop.GetExcludedUsers()

		if len(includedUsers) == 0 {
			if existingDesktop.RestrictedAccessUsers.AllowList != nil && len(existingDesktop.RestrictedAccessUsers.AllowList) > 0 {
				existingDesktop.RestrictedAccessUsers.AllowList = nil
			}
		} else {
			allowedUsers := getPrincipalNameForUserFromIdentityResponseModel(includedUsers)
			updatedAllowedUsers := getUpdatedUsersList(existingDesktop.RestrictedAccessUsers.AllowList, allowedUsers)
			existingDesktop.RestrictedAccessUsers.AllowList = updatedAllowedUsers
		}

		if len(excludedUsers) == 0 {
			if existingDesktop.RestrictedAccessUsers.BlockList != nil && len(existingDesktop.RestrictedAccessUsers.BlockList) > 0 {
				existingDesktop.RestrictedAccessUsers.BlockList = nil
			}
		} else {
			blockedUsers := getPrincipalNameForUserFromIdentityResponseModel(excludedUsers)
			updatedBlockedUsers := getUpdatedUsersList(existingDesktop.RestrictedAccessUsers.BlockList, blockedUsers)
			existingDesktop.RestrictedAccessUsers.BlockList = updatedBlockedUsers
		}

		if exists {
			existingDesktops[index] = existingDesktop
		} else {
			existingDesktops = append(existingDesktops, existingDesktop)
		}

		existingDesktopsMap[strings.ToLower(desktop.GetPublishedName())] = -1 // visited
	}

	updatedDesktopsList := []DeliveryGroupDesktop{}
	for _, desktop := range existingDesktops {
		if existingDesktopsMap[strings.ToLower(desktop.PublishedName.ValueString())] == -1 {
			updatedDesktopsList = append(updatedDesktopsList, desktop)
		}
	}

	r.Desktops = &updatedDesktopsList
	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithRestrictedAccessUsers(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel) DeliveryGroupResourceModel {
	simpleAccessPolicy := deliveryGroup.GetSimpleAccessPolicy()

	if !r.AllowAnonymousAccess.IsNull() {
		r.AllowAnonymousAccess = types.BoolValue(simpleAccessPolicy.GetAllowAnonymous())
	}

	if !simpleAccessPolicy.GetIncludedUserFilterEnabled() {
		r.RestrictedAccessUsers = nil
		return r
	}

	if r.RestrictedAccessUsers == nil {
		r.RestrictedAccessUsers = &RestrictedAccessUsers{}
	}

	includedUsers := getPrincipalNameForUserFromIdentityResponseModel(simpleAccessPolicy.GetIncludedUsers())
	updatedAllowList := getUpdatedUsersList(r.RestrictedAccessUsers.AllowList, includedUsers)
	r.RestrictedAccessUsers.AllowList = updatedAllowList

	if simpleAccessPolicy.GetExcludedUserFilterEnabled() {
		excludedUsers := getPrincipalNameForUserFromIdentityResponseModel(simpleAccessPolicy.GetExcludedUsers())
		updatedBlockList := getUpdatedUsersList(r.RestrictedAccessUsers.BlockList, excludedUsers)
		r.RestrictedAccessUsers.BlockList = updatedBlockList
	}

	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithAutoscaleSettings(deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection) DeliveryGroupResourceModel {
	if r.AutoscaleSettings == nil {
		return r
	}

	r.AutoscaleSettings.AutoscaleEnabled = types.BoolValue(deliveryGroup.GetAutoScaleEnabled())

	if !r.AutoscaleSettings.Timezone.IsNull() {
		r.AutoscaleSettings.Timezone = types.StringValue(deliveryGroup.GetTimeZone())
	}

	if !r.AutoscaleSettings.PeakDisconnectTimeoutMinutes.IsNull() {
		r.AutoscaleSettings.PeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakDisconnectTimeoutMinutes()))
	}

	if !r.AutoscaleSettings.PeakLogOffAction.IsNull() {
		r.AutoscaleSettings.PeakLogOffAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakLogOffAction()).String())
	}

	if !r.AutoscaleSettings.PeakDisconnectAction.IsNull() {
		r.AutoscaleSettings.PeakDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakDisconnectAction()).String())
	}

	if !r.AutoscaleSettings.PeakExtendedDisconnectAction.IsNull() {
		r.AutoscaleSettings.PeakExtendedDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetPeakExtendedDisconnectAction()).String())
	}

	if !r.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.IsNull() {
		r.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakExtendedDisconnectTimeoutMinutes()))
	}

	if !r.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.IsNull() {
		r.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakDisconnectTimeoutMinutes()))
	}

	if !r.AutoscaleSettings.OffPeakLogOffAction.IsNull() {
		r.AutoscaleSettings.OffPeakLogOffAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakLogOffAction()).String())
	}

	if !r.AutoscaleSettings.OffPeakDisconnectAction.IsNull() {
		r.AutoscaleSettings.OffPeakDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakExtendedDisconnectAction()).String())
	}

	if !r.AutoscaleSettings.OffPeakExtendedDisconnectAction.IsNull() {
		r.AutoscaleSettings.OffPeakExtendedDisconnectAction = types.StringValue(reflect.ValueOf(deliveryGroup.GetOffPeakExtendedDisconnectAction()).String())
	}

	if !r.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.IsNull() {
		r.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakExtendedDisconnectTimeoutMinutes()))
	}

	if !r.AutoscaleSettings.PeakBufferSizePercent.IsNull() {
		r.AutoscaleSettings.PeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetPeakBufferSizePercent()))
	}

	if !r.AutoscaleSettings.OffPeakBufferSizePercent.IsNull() {
		r.AutoscaleSettings.OffPeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetOffPeakBufferSizePercent()))
	}

	if !r.AutoscaleSettings.PowerOffDelayMinutes.IsNull() {
		r.AutoscaleSettings.PowerOffDelayMinutes = types.Int64Value(int64(deliveryGroup.GetPowerOffDelayMinutes()))
	}

	if !r.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds.IsNull() {
		r.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectPeakIdleSessionAfterSeconds()))
	}

	if !r.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds.IsNull() {
		r.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectOffPeakIdleSessionAfterSeconds()))
	}

	if !r.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds.IsNull() {
		r.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffPeakDisconnectedSessionAfterSeconds()))
	}

	if !r.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds.IsNull() {
		r.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffOffPeakDisconnectedSessionAfterSeconds()))
	}

	parsedPowerTimeSchemes := ParsePowerTimeSchemesClientToPluginModel(dgPowerTimeSchemes.GetItems())
	r.AutoscaleSettings.PowerTimeSchemes = parsedPowerTimeSchemes

	return r
}

func getPrincipalNameForUserFromIdentityResponseModel(users []citrixorchestration.IdentityUserResponseModel) []string {
	var res []string
	for _, user := range users {
		res = append(res, user.GetPrincipalName())
	}
	return res
}

func getUpdatedUsersList(existingUsers []types.String, usersFromRemote []string) []types.String {
	existingUsersMap := map[string]bool{}
	for _, user := range existingUsers {
		existingUsersMap[strings.ToLower(user.ValueString())] = false // not visited
	}

	for _, user := range usersFromRemote {
		userInLowerCase := strings.ToLower(user)
		_, exists := existingUsersMap[userInLowerCase]
		if !exists {
			existingUsers = append(existingUsers, types.StringValue(user))
		}
		existingUsersMap[userInLowerCase] = true
	}

	updatedUsersList := []types.String{}
	for _, user := range existingUsers {
		if existingUsersMap[strings.ToLower(user.ValueString())] {
			updatedUsersList = append(updatedUsersList, user)
		}
	}

	return updatedUsersList
}
