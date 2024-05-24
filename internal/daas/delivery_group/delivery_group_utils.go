// Copyright © 2023. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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

func getDeliveryGroupRebootSchedules(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.RebootScheduleResponseModelCollection, error) {
	getDeliveryGroupRebootScheduleRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupRebootSchedules(ctx, deliveryGroupId)
	deliveryGroupRebootSchedule, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.RebootScheduleResponseModelCollection](getDeliveryGroupRebootScheduleRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Reboot Schedule for Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return deliveryGroupRebootSchedule, err
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

func validatePowerManagementSettings(plan DeliveryGroupResourceModel, sessionSupport citrixorchestration.SessionSupport) (bool, string) {
	if plan.AutoscaleSettings == nil || sessionSupport == citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION {
		return true, ""
	}

	errStringSuffix := "cannot be set for a Multisession catalog"

	if plan.AutoscaleSettings.PeakLogOffAction.ValueString() != "Nothing" {
		return false, "PeakLogOffAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakLogOffAction.ValueString() != "Nothing" {
		return false, "OffPeakLogOffAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString() != "Nothing" {
		return false, "OffPeakDisconnectAction " + errStringSuffix
	}

	if plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString() != "Nothing" {
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

func validateAndReturnMachineCatalogSessionSupport(ctx context.Context, client citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, dgMachineCatalogs []DeliveryGroupMachineCatalogModel, addErrorIfCatalogNotFound bool) (catalogSessionSupport *citrixorchestration.SessionSupport, isPowerManagedCatalog bool, isRemotePcCatalog bool, catalogIdentityType citrixorchestration.IdentityType, err error) {
	var sessionSupport *citrixorchestration.SessionSupport
	var provisioningType *citrixorchestration.ProvisioningType
	var identityType citrixorchestration.IdentityType
	isPowerManaged := false
	isRemotePc := false
	for _, dgMachineCatalog := range dgMachineCatalogs {
		catalogId := dgMachineCatalog.MachineCatalog.ValueString()
		if catalogId == "" {
			continue
		}

		catalog, err := util.GetMachineCatalog(ctx, &client, diagnostics, catalogId, addErrorIfCatalogNotFound)

		if err != nil {
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		if provisioningType == nil {
			provisioningType = &catalog.ProvisioningType
			isPowerManaged = catalog.GetIsPowerManaged()
			isRemotePc = catalog.GetIsRemotePC()
			provScheme := catalog.GetProvisioningScheme()
			identityType = provScheme.GetIdentityType()
		}

		if *provisioningType != catalog.GetProvisioningType() {
			err := fmt.Errorf("associated_machine_catalogs must have catalogs with the same provsioning type")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"Ensure all associated Machine Catalogs have the same provisioning type.",
			)
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		provScheme := catalog.GetProvisioningScheme()

		if identityType != provScheme.GetIdentityType() {
			err := fmt.Errorf("associated_machine_catalogs must have catalogs with the same identity type in provisioning scheme")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"Ensure all associated Machine Catalogs have the same identity type in provisioning scheme.",
			)
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		if isPowerManaged != catalog.GetIsPowerManaged() {
			err := fmt.Errorf("all associated_machine_catalogs must either be power managed or non power managed")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"All associated Machine Catalogs must either be power managed or non power managed.",
			)
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		if isRemotePc != catalog.GetIsRemotePC() {
			err := fmt.Errorf("all associated_machine_catalogs must either be Remote PC or non Remote PC")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"All associated Machine Catalogs must either be Remote PC or non Remote PC.",
			)
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		if sessionSupport != nil && *sessionSupport != catalog.GetSessionSupport() {
			err := fmt.Errorf("all associated machine catalogs must have the same session support")
			diagnostics.AddError("Error validating associated Machine Catalogs", "Ensure all associated Machine Catalogs have the same Session Support.")
			return sessionSupport, false, false, citrixorchestration.IDENTITYTYPE_UNKNOWN, err
		}

		if sessionSupport == nil {
			sessionSupportValue := catalog.GetSessionSupport()
			sessionSupport = &sessionSupportValue
		}
	}

	return sessionSupport, isPowerManaged, isRemotePc, identityType, err
}

func getDeliveryGroupAddMachinesRequest(associatedMachineCatalogs []DeliveryGroupMachineCatalogModel) []citrixorchestration.DeliveryGroupAddMachinesRequestModel {
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

func addMachinesToDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string, catalogId string, numOfMachines int) (*citrixorchestration.DeliveryGroupDetailResponseModel, error) {
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

func removeMachinesFromDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string, machinesToRemove []string) error {
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

func addRemoveMachinesFromDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string, plan DeliveryGroupResourceModel) error {
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, client, diagnostics, deliveryGroupId)

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

func getSchemaForRestrictedAccessUsers(forDeliveryGroup bool) schema.NestedAttribute {
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
			"allow_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who can use this %s. Must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(fmt.Sprintf("%s|%s", util.SamRegex, util.UpnRegex)), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
					listvalidator.SizeAtLeast(1),
				},
			},
			"block_list": schema.ListAttribute{
				ElementType: types.StringType,
				Description: fmt.Sprintf("Users who cannot use this %s. A block list is meaningful only when used to block users in the allow list. Must be in `Domain\\UserOrGroupName` or `user@domain.com` format", resource),
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(fmt.Sprintf("%s|%s", util.SamRegex, util.UpnRegex)), "must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func validatePowerTimeSchemes(diagnostics *diag.Diagnostics, powerTimeSchemes []DeliveryGroupPowerTimeScheme) {
	for _, powerTimeScheme := range powerTimeSchemes {
		if powerTimeScheme.PoolSizeSchedule == nil {
			continue
		}

		hoursArray := make([]bool, 24)
		minutesArray := make([]bool, 24)

		hoursPoolSizeArray := make([]int, 24)
		minutesPoolSizeArray := make([]int, 24)

		for _, schedule := range powerTimeScheme.PoolSizeSchedule {
			timeRanges := strings.Split(schedule.TimeRange.ValueString(), "-")

			startTimes := strings.Split(timeRanges[0], ":")
			startHour, _ := strconv.Atoi(startTimes[0])
			startMinutes, _ := strconv.Atoi(startTimes[1])

			endTimes := strings.Split(timeRanges[1], ":")
			endHour, _ := strconv.Atoi(endTimes[0])
			endMinutes, _ := strconv.Atoi(endTimes[1])

			if endHour == 0 && endMinutes == 0 {
				endHour = 24
			}

			if startHour > endHour || (startHour == endHour && startMinutes > endMinutes) {
				diagnostics.AddAttributeError(
					path.Root("time_range"),
					"Incorrect Attribute Value",
					fmt.Sprintf("Unexpected time_range value %s. Start time cannot be greater than end time.", schedule.TimeRange.ValueString()),
				)
			}

			if startHour == endHour && startMinutes == endMinutes {
				diagnostics.AddAttributeError(
					path.Root("time_range"),
					"Incorrect Attribute Value",
					fmt.Sprintf("Unexpected time_range value %s. Start time and end time cannot be the same. Use 00:00-00:00 if you wish to cover all hours of the day.", schedule.TimeRange.ValueString()),
				)
			}

			for i := startHour; i < endHour; i++ {

				if i == startHour && hoursArray[i] && hoursPoolSizeArray[i] == int(schedule.PoolSize.ValueInt64()) {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Contiguous time ranges with the same pool size should be combined.", schedule.TimeRange.ValueString()),
					)
				}

				if i != 23 && hoursArray[i+1] && hoursPoolSizeArray[i+1] == int(schedule.PoolSize.ValueInt64()) {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Contiguous time ranges with the same pool size should be combined.", schedule.TimeRange.ValueString()),
					)
				}

				if minutesArray[i] {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Time ranges cannot overlap.", schedule.TimeRange.ValueString()),
					)
				}

				minutesArray[i] = true
				minutesPoolSizeArray[i] = int(schedule.PoolSize.ValueInt64())
				if startMinutes == 30 {
					continue
				}

				if i == startHour && startHour != 0 && minutesArray[i-1] && minutesPoolSizeArray[i-1] == int(schedule.PoolSize.ValueInt64()) {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Contiguous time ranges with the same pool size should be combined.", schedule.TimeRange.ValueString()),
					)
				}

				if hoursArray[i] {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Time ranges cannot overlap.", schedule.TimeRange.ValueString()),
					)
				}
				hoursArray[i] = true
				hoursPoolSizeArray[i] = int(schedule.PoolSize.ValueInt64())
			}

			if endMinutes == 30 {
				if minutesArray[endHour] && minutesPoolSizeArray[endHour] == int(schedule.PoolSize.ValueInt64()) {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Contiguous time ranges with the same pool size should be combined.", schedule.TimeRange.ValueString()),
					)
				}

				if hoursArray[endHour] {
					diagnostics.AddAttributeError(
						path.Root("time_range"),
						"Incorrect Attribute Value",
						fmt.Sprintf("Unexpected time_range value %s. Time ranges cannot overlap.", schedule.TimeRange.ValueString()),
					)
				}

				hoursArray[endHour] = true
				hoursPoolSizeArray[endHour] = int(schedule.PoolSize.ValueInt64())
			}
		}
	}
}

func validateRebootSchedules(diagnostics *diag.Diagnostics, rebootSchedules []DeliveryGroupRebootSchedule) {
	for _, rebootSchedule := range rebootSchedules {
		switch freq := rebootSchedule.Frequency.ValueString(); freq {
		case "Weekly":
			if len(rebootSchedule.DaysInWeek) == 0 {
				diagnostics.AddAttributeError(
					path.Root("days_in_week"),
					"Missing Attribute Configuration",
					"Days in week must be specified for weekly reboot schedule.",
				)
			}
		case "Monthly":
			if rebootSchedule.WeekInMonth.IsNull() || rebootSchedule.WeekInMonth.ValueString() == "" {
				diagnostics.AddAttributeError(
					path.Root("week_in_month"),
					"Missing Attribute Configuration",
					"Week in month must be specified for monthly reboot schedule.",
				)
			}
			if rebootSchedule.DayInMonth.IsNull() || rebootSchedule.DayInMonth.ValueString() == "" {
				diagnostics.AddAttributeError(
					path.Root("day_in_month"),
					"Missing Attribute Configuration",
					"Day in month must be specified for monthly reboot schedule.",
				)
			}

		}

		if rebootSchedule.UseNaturalRebootSchedule.ValueBool() && rebootSchedule.DeliveryGroupRebootNotificationToUsers != nil {
			diagnostics.AddAttributeError(
				path.Root("reboot_notification_to_users"),
				"Incorrect Attribute Configuration",
				"Reboot notification to users can not be set for using natural reboot",
			)
		}

		if rebootSchedule.DeliveryGroupRebootNotificationToUsers != nil && !rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationDurationMinutes.IsNull() && !rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationRepeatEvery5Minutes.IsNull() &&
			rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationDurationMinutes.ValueInt64() != 15 {
			diagnostics.AddAttributeError(
				path.Root("notification_repeat_every_5_minutes"),
				"Incorrect Attribute Configuration",
				"NotificationRepeatEvery5Minutes can only be set to true when NotificationDurationMinutes is 15 minutes",
			)
		}

	}
}

func getRequestModelForDeliveryGroupCreate(diagnostics *diag.Diagnostics, plan DeliveryGroupResourceModel, catalogSessionSupport *citrixorchestration.SessionSupport, identityType citrixorchestration.IdentityType) (citrixorchestration.CreateDeliveryGroupRequestModel, error) {
	deliveryGroupMachineCatalogsArray := getDeliveryGroupAddMachinesRequest(plan.AssociatedMachineCatalogs)
	deliveryGroupDesktopsArray := parseDeliveryGroupDesktopsToClientModel(plan.Desktops)
	deliveryGroupRebootScheduleArray := parseDeliveryGroupRebootScheduleToClientModel(plan.RebootSchedules)

	var includedUsers []string
	var excludedUsers []string
	includedUsersFilterEnabled := false
	excludedUsersFilterEnabled := false
	if plan.RestrictedAccessUsers != nil {
		includedUsersFilterEnabled = true
		includedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(plan.RestrictedAccessUsers.AllowList)

		if plan.RestrictedAccessUsers.BlockList != nil {
			excludedUsersFilterEnabled = true
			excludedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(plan.RestrictedAccessUsers.BlockList)
		}
	}

	var simpleAccessPolicy citrixorchestration.SimplifiedAccessPolicyRequestModel
	simpleAccessPolicy.SetAllowAnonymous(plan.AllowAnonymousAccess.ValueBool())
	simpleAccessPolicy.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
	simpleAccessPolicy.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
	simpleAccessPolicy.SetIncludedUsers(includedUsers)
	simpleAccessPolicy.SetExcludedUsers(excludedUsers)

	var body citrixorchestration.CreateDeliveryGroupRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetMachineCatalogs(deliveryGroupMachineCatalogsArray)
	body.SetRebootSchedules(deliveryGroupRebootScheduleArray)

	functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating Delivery Group",
			fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
		)
		return body, err
	}
	body.SetMinimumFunctionalLevel(*functionalLevel)

	deliveryKind := citrixorchestration.DELIVERYKIND_DESKTOPS_AND_APPS
	if *catalogSessionSupport != citrixorchestration.SESSIONSUPPORT_MULTI_SESSION {
		deliveryKind = citrixorchestration.DELIVERYKIND_DESKTOPS_ONLY
	}
	body.SetDeliveryType(deliveryKind)
	body.SetDesktops(deliveryGroupDesktopsArray)
	body.SetDefaultDesktopPublishedName(plan.Name.ValueString())
	body.SetSimpleAccessPolicy(simpleAccessPolicy)
	body.SetPolicySetGuid(plan.PolicySetId.ValueString())
	if identityType == citrixorchestration.IDENTITYTYPE_AZURE_AD {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_AZURE_AD)
	} else if identityType == citrixorchestration.IDENTITYTYPE_WORKGROUP {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_LOCAL_MAPPED_ACCOUNT)
	} else {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_ACTIVE_DIRECTORY)
	}

	if plan.AutoscaleSettings != nil {
		body.SetAutoScaleEnabled(plan.AutoscaleSettings.AutoscaleEnabled.ValueBool())
		body.SetTimeZone(plan.AutoscaleSettings.Timezone.ValueString())
		body.SetPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakDisconnectTimeoutMinutes.ValueInt64()))
		body.SetPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakLogOffAction.ValueString()))
		body.SetPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakDisconnectAction.ValueString()))
		body.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString()))
		body.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakLogOffAction.ValueString()))
		body.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString()))
		body.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString()))
		body.SetPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetOffPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.ValueInt64()))
		body.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetPeakBufferSizePercent(int32(plan.AutoscaleSettings.PeakBufferSizePercent.ValueInt64()))
		body.SetOffPeakBufferSizePercent(int32(plan.AutoscaleSettings.OffPeakBufferSizePercent.ValueInt64()))
		body.SetPowerOffDelayMinutes(int32(plan.AutoscaleSettings.PowerOffDelayMinutes.ValueInt64()))
		body.SetDisconnectPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		body.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := parsePowerTimeSchemesPluginToClientModel(plan.AutoscaleSettings.PowerTimeSchemes)
		body.SetPowerTimeSchemes(powerTimeSchemes)
	}

	return body, nil
}

func getRequestModelForDeliveryGroupUpdate(diagnostics *diag.Diagnostics, plan DeliveryGroupResourceModel, currentDeliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel) (citrixorchestration.EditDeliveryGroupRequestModel, error) {
	deliveryGroupDesktopsArray := parseDeliveryGroupDesktopsToClientModel(plan.Desktops)
	deliveryGroupRebootScheduleArray := parseDeliveryGroupRebootScheduleToClientModel(plan.RebootSchedules)

	includedUsers := []string{}
	excludedUsers := []string{}
	includedUsersFilterEnabled := false
	excludedUsersFilterEnabled := false
	advancedAccessPolicies := []citrixorchestration.AdvancedAccessPolicyRequestModel{}

	allowedUser := citrixorchestration.ALLOWEDUSER_ANY_AUTHENTICATED

	if plan.AllowAnonymousAccess.ValueBool() {
		allowedUser = citrixorchestration.ALLOWEDUSER_ANY
	}

	if plan.RestrictedAccessUsers != nil {
		allowedUser = citrixorchestration.ALLOWEDUSER_FILTERED

		if plan.AllowAnonymousAccess.ValueBool() {
			allowedUser = citrixorchestration.ALLOWEDUSER_FILTERED_OR_ANONYMOUS
		}

		includedUsersFilterEnabled = true
		includedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(plan.RestrictedAccessUsers.AllowList)

		if plan.RestrictedAccessUsers.BlockList != nil {
			excludedUsersFilterEnabled = true
			excludedUsers = util.ConvertBaseStringArrayToPrimitiveStringArray(plan.RestrictedAccessUsers.BlockList)
		}
	}

	existingAdvancedAccessPolicies := currentDeliveryGroup.GetAdvancedAccessPolicy()
	for _, existingAdvancedAccessPolicy := range existingAdvancedAccessPolicies {
		var advancedAccessPolicyRequest citrixorchestration.AdvancedAccessPolicyRequestModel
		advancedAccessPolicyRequest.SetId(existingAdvancedAccessPolicy.GetId())
		advancedAccessPolicyRequest.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
		advancedAccessPolicyRequest.SetIncludedUsers(includedUsers)
		advancedAccessPolicyRequest.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
		advancedAccessPolicyRequest.SetExcludedUsers(excludedUsers)
		advancedAccessPolicyRequest.SetAllowedUsers(allowedUser)
		advancedAccessPolicies = append(advancedAccessPolicies, advancedAccessPolicyRequest)
	}

	// Construct the update model
	var editDeliveryGroupRequestBody citrixorchestration.EditDeliveryGroupRequestModel
	editDeliveryGroupRequestBody.SetName(plan.Name.ValueString())
	editDeliveryGroupRequestBody.SetDescription(plan.Description.ValueString())
	editDeliveryGroupRequestBody.SetDesktops(deliveryGroupDesktopsArray)
	editDeliveryGroupRequestBody.SetRebootSchedules(deliveryGroupRebootScheduleArray)
	editDeliveryGroupRequestBody.SetAdvancedAccessPolicy(advancedAccessPolicies)

	if plan.PolicySetId.ValueString() != "" {
		editDeliveryGroupRequestBody.SetPolicySetGuid(plan.PolicySetId.ValueString())
	} else {
		editDeliveryGroupRequestBody.SetPolicySetGuid(util.DefaultSitePolicySetId)
	}

	if plan.AutoscaleSettings != nil {

		if plan.AutoscaleSettings.Timezone.ValueString() != "" {
			editDeliveryGroupRequestBody.SetTimeZone(plan.AutoscaleSettings.Timezone.ValueString())
		}

		editDeliveryGroupRequestBody.SetAutoScaleEnabled(plan.AutoscaleSettings.AutoscaleEnabled.ValueBool())
		editDeliveryGroupRequestBody.SetPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakLogOffAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.PeakExtendedDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakLogOffAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(plan.AutoscaleSettings.OffPeakExtendedDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(plan.AutoscaleSettings.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetPeakBufferSizePercent(int32(plan.AutoscaleSettings.PeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakBufferSizePercent(int32(plan.AutoscaleSettings.OffPeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetPowerOffDelayMinutes(int32(plan.AutoscaleSettings.PowerOffDelayMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(plan.AutoscaleSettings.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(plan.AutoscaleSettings.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := parsePowerTimeSchemesPluginToClientModel(plan.AutoscaleSettings.PowerTimeSchemes)
		editDeliveryGroupRequestBody.SetPowerTimeSchemes(powerTimeSchemes)
	}

	functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error updating Delivery Group",
			fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
		)
		return editDeliveryGroupRequestBody, err
	}
	editDeliveryGroupRequestBody.SetMinimumFunctionalLevel(*functionalLevel)

	return editDeliveryGroupRequestBody, nil
}

func parsePowerTimeSchemesPluginToClientModel(powerTimeSchemes []DeliveryGroupPowerTimeScheme) []citrixorchestration.PowerTimeSchemeRequestModel {
	res := []citrixorchestration.PowerTimeSchemeRequestModel{}
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

func parsePowerTimeSchemesClientToPluginModel(powerTimeSchemesResponse []citrixorchestration.PowerTimeSchemeResponseModel) []DeliveryGroupPowerTimeScheme {
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
			if poolSizeSchedule.GetPoolSize() == 0 {
				continue
			}

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

func parseDeliveryGroupRebootScheduleToClientModel(rebootSchedules []DeliveryGroupRebootSchedule) []citrixorchestration.RebootScheduleRequestModel {
	res := []citrixorchestration.RebootScheduleRequestModel{}
	if rebootSchedules == nil {
		return res
	}

	for _, rebootSchedule := range rebootSchedules {
		var rebootScheduleRequest citrixorchestration.RebootScheduleRequestModel
		rebootScheduleRequest.SetName(rebootSchedule.Name.ValueString())
		if !rebootSchedule.Description.IsNull() {
			rebootScheduleRequest.SetDescription(rebootSchedule.Description.ValueString())
		}
		if !rebootSchedule.RestrictToTag.IsNull() {
			rebootScheduleRequest.SetRestrictToTag(rebootSchedule.RestrictToTag.ValueString())
		}

		rebootScheduleRequest.SetIgnoreMaintenanceMode(true)
		rebootScheduleRequest.SetEnabled(rebootSchedule.RebootScheduleEnabled.ValueBool())
		rebootScheduleRequest.SetFrequency(getFrequencyActionValue(rebootSchedule.Frequency.ValueString()))
		rebootScheduleRequest.SetFrequencyFactor(int32(rebootSchedule.FrequencyFactor.ValueInt64()))
		rebootScheduleRequest.SetStartDate(rebootSchedule.StartDate.ValueString())
		rebootScheduleRequest.SetStartTime(rebootSchedule.StartTime.ValueString() + ":00")
		rebootScheduleRequest.SetRebootDurationMinutes(int32(rebootSchedule.RebootDurationMinutes.ValueInt64()))
		rebootScheduleRequest.SetUseNaturalReboot(rebootSchedule.UseNaturalRebootSchedule.ValueBool())
		if rebootSchedule.Frequency.ValueString() == "Weekly" {
			rebootScheduleRequest.SetDaysInWeek(getRebootScheduleDaysInWeekActionValue(util.ConvertBaseStringArrayToPrimitiveStringArray(rebootSchedule.DaysInWeek)))
		}
		if rebootSchedule.Frequency.ValueString() == "Monthly" {
			rebootScheduleRequest.SetWeekInMonth(getRebootScheduleWeekActionValue(rebootSchedule.WeekInMonth.ValueString()))
			rebootScheduleRequest.SetDayInMonth(getRebootScheduleDaysActionValue(rebootSchedule.DayInMonth.ValueString()))
		}

		if rebootSchedule.DeliveryGroupRebootNotificationToUsers != nil && !rebootSchedule.UseNaturalRebootSchedule.ValueBool() {
			rebootScheduleRequest.SetWarningDurationMinutes(int32(rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationDurationMinutes.ValueInt64())) //can only be 1 5 15, or 0 means no warning
			rebootScheduleRequest.SetWarningTitle(rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationTitle.ValueString())
			rebootScheduleRequest.SetWarningMessage(rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationMessage.ValueString())
			if rebootSchedule.DeliveryGroupRebootNotificationToUsers.NotificationRepeatEvery5Minutes.ValueBool() {
				rebootScheduleRequest.SetWarningRepeatIntervalMinutes(5)
			} else {
				rebootScheduleRequest.SetWarningRepeatIntervalMinutes(0)
			}
		} else {
			rebootScheduleRequest.SetRebootDurationMinutes(0)
			rebootScheduleRequest.SetWarningDurationMinutes(0) //can only be 1 5 15, or 0 means no warning
		}

		res = append(res, rebootScheduleRequest)
	}

	return res

}

func (schedule DeliveryGroupRebootSchedule) RefreshListItem(rebootSchedule citrixorchestration.RebootScheduleResponseModel) DeliveryGroupRebootSchedule {
	schedule.Name = types.StringValue(rebootSchedule.GetName())
	if rebootSchedule.GetDescription() != "" {
		schedule.Description = types.StringValue(rebootSchedule.GetDescription())
	}

	schedule.RebootScheduleEnabled = types.BoolValue(rebootSchedule.GetEnabled())
	if rebootSchedule.GetRestrictToTag().Id.Get() != nil {
		schedule.RestrictToTag = types.StringValue(*rebootSchedule.GetRestrictToTag().Name.Get())
	}
	schedule.IgnoreMaintenanceMode = types.BoolValue(rebootSchedule.GetIgnoreMaintenanceMode()) //bug in orchestration side
	schedule.Frequency = types.StringValue(reflect.ValueOf(rebootSchedule.GetFrequency()).String())

	startDate := schedule.StartDate.ValueString()
	if reflect.ValueOf(rebootSchedule.GetFrequency()).String() == "Weekly" {
		res := []types.String{}
		for _, scheduleDay := range rebootSchedule.GetDaysInWeek() {
			res = append(res, types.StringValue(reflect.ValueOf(scheduleDay).String()))
		}
		schedule.DaysInWeek = res
		if startDate != "" && startDate[5:6] == rebootSchedule.GetStartDate()[5:6] { //same month for plan and remote schedule
			schedule.StartDate = types.StringValue(startDate)
		}
	} else if reflect.ValueOf(rebootSchedule.GetFrequency()).String() == "Monthly" {
		schedule.WeekInMonth = types.StringValue(reflect.ValueOf(rebootSchedule.GetWeekInMonth()).String())
		schedule.DayInMonth = types.StringValue(reflect.ValueOf(rebootSchedule.GetDayInMonth()).String())
		schedule.StartDate = types.StringValue(startDate)
	} else {
		if startDate != "" {
			schedule.StartDate = types.StringValue(rebootSchedule.GetStartDate())
		}
	}
	schedule.FrequencyFactor = types.Int64Value(int64(rebootSchedule.GetFrequencyFactor()))

	schedule.StartTime = types.StringValue(rebootSchedule.GetStartTime()[:5])

	if rebootSchedule.GetUseNaturalReboot() {
		schedule.UseNaturalRebootSchedule = types.BoolValue(true)
	} else {
		schedule.UseNaturalRebootSchedule = types.BoolValue(false)
		schedule.RebootDurationMinutes = types.Int64Value(int64(rebootSchedule.GetRebootDurationMinutes()))
		if rebootSchedule.GetWarningDurationMinutes() != 0 {
			notif := DeliveryGroupRebootNotificationToUsers{
				NotificationDurationMinutes: types.Int64Value(int64(rebootSchedule.GetWarningDurationMinutes())),
				NotificationTitle:           types.StringValue(rebootSchedule.GetWarningTitle()),
				NotificationMessage:         types.StringValue(rebootSchedule.GetWarningMessage()),
			}
			if rebootSchedule.GetWarningRepeatIntervalMinutes() == 5 {
				notif.NotificationRepeatEvery5Minutes = types.BoolValue(true)
			}
			schedule.DeliveryGroupRebootNotificationToUsers = &notif
		}
	}
	return schedule
}

func (dgDesktop DeliveryGroupDesktop) RefreshListItem(desktop citrixorchestration.DesktopResponseModel) DeliveryGroupDesktop {
	dgDesktop.PublishedName = types.StringValue(desktop.GetPublishedName())
	if desktop.GetDescription() != "" {
		dgDesktop.DesktopDescription = types.StringValue(desktop.GetDescription())
	} else {
		dgDesktop.DesktopDescription = types.StringNull()
	}
	dgDesktop.Enabled = types.BoolValue(desktop.GetEnabled())
	sessionReconnection := desktop.GetSessionReconnection()
	if sessionReconnection == citrixorchestration.SESSIONRECONNECTION_ALWAYS {
		dgDesktop.EnableSessionRoaming = types.BoolValue(true)
	} else {
		dgDesktop.EnableSessionRoaming = types.BoolValue(false)
	}

	if !desktop.GetIncludedUserFilterEnabled() {
		dgDesktop.RestrictedAccessUsers = nil
		return dgDesktop
	}

	if dgDesktop.RestrictedAccessUsers == nil {
		dgDesktop.RestrictedAccessUsers = &RestrictedAccessUsers{}
	}

	includedUsers := desktop.GetIncludedUsers()
	excludedUsers := desktop.GetExcludedUsers()

	if len(includedUsers) == 0 {
		dgDesktop.RestrictedAccessUsers.AllowList = nil
	} else {
		updatedAllowedUsers := refreshUsersList(dgDesktop.RestrictedAccessUsers.AllowList, includedUsers)
		dgDesktop.RestrictedAccessUsers.AllowList = updatedAllowedUsers
	}

	if len(excludedUsers) == 0 {
		dgDesktop.RestrictedAccessUsers.BlockList = nil
	} else {
		updatedBlockedUsers := refreshUsersList(dgDesktop.RestrictedAccessUsers.BlockList, excludedUsers)
		dgDesktop.RestrictedAccessUsers.BlockList = updatedBlockedUsers
	}

	return dgDesktop
}

func getFrequencyActionValue(v string) citrixorchestration.RebootScheduleFrequency {
	frequency, err := citrixorchestration.NewRebootScheduleFrequencyFromValue(v)

	if err != nil {
		return citrixorchestration.REBOOTSCHEDULEFREQUENCY_UNKNOWN
	}

	return *frequency
}

func getRebootScheduleWeekActionValue(v string) citrixorchestration.RebootScheduleWeeks {
	week, err := citrixorchestration.NewRebootScheduleWeeksFromValue(v)

	if err != nil {
		return citrixorchestration.REBOOTSCHEDULEWEEKS_UNKNOWN
	}

	return *week
}

func getRebootScheduleDaysActionValue(v string) citrixorchestration.RebootScheduleDays {
	days, err := citrixorchestration.NewRebootScheduleDaysFromValue(v)

	if err != nil {
		return citrixorchestration.REBOOTSCHEDULEDAYS_UNKNOWN
	}

	return *days
}

func getRebootScheduleDaysInWeekActionValue(v []string) []citrixorchestration.RebootScheduleDays {
	var res []citrixorchestration.RebootScheduleDays
	for _, day := range v {
		days, err := citrixorchestration.NewRebootScheduleDaysFromValue(day)

		if err != nil {
			res = append(res, citrixorchestration.REBOOTSCHEDULEDAYS_UNKNOWN)
		} else {
			res = append(res, *days)
		}
	}

	return res
}

func parseDeliveryGroupDesktopsToClientModel(deliveryGroupDesktops []DeliveryGroupDesktop) []citrixorchestration.DesktopRequestModel {
	desktopRequests := []citrixorchestration.DesktopRequestModel{}

	if deliveryGroupDesktops == nil {
		return desktopRequests
	}

	for _, deliveryGroupDesktop := range deliveryGroupDesktops {
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

func (r DeliveryGroupResourceModel) updatePlanWithRebootSchedule(rebootSchedules *citrixorchestration.RebootScheduleResponseModelCollection) DeliveryGroupResourceModel {
	schedules := util.RefreshListProperties[DeliveryGroupRebootSchedule, citrixorchestration.RebootScheduleResponseModel](r.RebootSchedules, "Name", rebootSchedules.GetItems(), "Name", "RefreshListItem")
	r.RebootSchedules = schedules
	return r
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
	desktops := util.RefreshListProperties[DeliveryGroupDesktop, citrixorchestration.DesktopResponseModel](r.Desktops, "PublishedName", deliveryGroupDesktops.GetItems(), "PublishedName", "RefreshListItem")
	r.Desktops = desktops
	return r
}

func updateDeliveryGroupUserAccessDetails(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, deliveryGroupDesktops *citrixorchestration.DesktopResponseModelCollection) (*citrixorchestration.DeliveryGroupDetailResponseModel, *citrixorchestration.DesktopResponseModelCollection, error) {
	simpleAccessPolicy := deliveryGroup.GetSimpleAccessPolicy()
	updatedIncludedUsers, updatedExcludedUsers, err := updateIdentityUserDetails(ctx, client, diagnostics, simpleAccessPolicy.GetIncludedUsers(), simpleAccessPolicy.GetExcludedUsers())
	if err != nil {
		return deliveryGroup, deliveryGroupDesktops, err
	}

	simpleAccessPolicy.SetIncludedUsers(updatedIncludedUsers)
	simpleAccessPolicy.SetExcludedUsers(updatedExcludedUsers)
	deliveryGroup.SetSimpleAccessPolicy(simpleAccessPolicy)

	updatedDeliveryGroupDesktops := []citrixorchestration.DesktopResponseModel{}
	for _, desktop := range deliveryGroupDesktops.GetItems() {
		updatedIncludedUsers, updatedExcludedUsers, err := updateIdentityUserDetails(ctx, client, diagnostics, desktop.GetIncludedUsers(), desktop.GetExcludedUsers())
		if err != nil {
			return deliveryGroup, deliveryGroupDesktops, err
		}
		desktop.SetIncludedUsers(updatedIncludedUsers)
		desktop.SetExcludedUsers(updatedExcludedUsers)
		updatedDeliveryGroupDesktops = append(updatedDeliveryGroupDesktops, desktop)
	}
	deliveryGroupDesktops.SetItems(updatedDeliveryGroupDesktops)

	return deliveryGroup, deliveryGroupDesktops, nil
}

func verifyIdentityUserListCompleteness(inputUserNames []string, remoteUsers []citrixorchestration.IdentityUserResponseModel) error {
	if len(remoteUsers) < len(inputUserNames) {
		missingUsers := []string{}
		for _, includedUser := range inputUserNames {
			userIndex := slices.IndexFunc(remoteUsers, func(i citrixorchestration.IdentityUserResponseModel) bool {
				return includedUser == i.GetSamName() || includedUser == i.GetPrincipalName()
			})
			if userIndex == -1 {
				missingUsers = append(missingUsers, includedUser)
			}
		}

		return fmt.Errorf("The following users could not be found: " + strings.Join(missingUsers, ", "))
	}
	return nil
}

func updateIdentityUserDetails(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, includedUsers []citrixorchestration.IdentityUserResponseModel, excludedUsers []citrixorchestration.IdentityUserResponseModel) ([]citrixorchestration.IdentityUserResponseModel, []citrixorchestration.IdentityUserResponseModel, error) {
	includedUserNames := []string{}
	for _, includedUser := range includedUsers {
		if includedUser.GetSamName() != "" {
			includedUserNames = append(includedUserNames, includedUser.GetSamName())
		} else if includedUser.GetPrincipalName() != "" {
			includedUserNames = append(includedUserNames, includedUser.GetPrincipalName())
		}
	}

	if len(includedUserNames) > 0 {
		getIncludedUsersRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetUsers(ctx)
		getIncludedUsersRequest = getIncludedUsersRequest.User(includedUserNames)
		getIncludedUsersRequest = getIncludedUsersRequest.Provider(citrixorchestration.IDENTITYPROVIDERTYPE_ALL)
		includedUsersResponse, httpResp, err := citrixdaasclient.AddRequestData(getIncludedUsersRequest, client).Execute()
		if err != nil {
			diagnostics.AddError(
				"Error fetching delivery group user details",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, nil, err
		}

		err = verifyIdentityUserListCompleteness(includedUserNames, includedUsersResponse.GetItems())
		if err != nil {
			diagnostics.AddError(
				"Error fetching delivery group user details",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+err.Error(),
			)
			return nil, nil, err
		}

		includedUsers = includedUsersResponse.GetItems()
	}

	excludedUserNames := []string{}
	for _, excludedUser := range excludedUsers {
		if excludedUser.GetSamName() != "" {
			excludedUserNames = append(excludedUserNames, excludedUser.GetSamName())
		} else if excludedUser.GetPrincipalName() != "" {
			excludedUserNames = append(excludedUserNames, excludedUser.GetPrincipalName())
		}
	}

	if len(excludedUserNames) > 0 {
		getExcludedUsersRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetUsers(ctx)
		getExcludedUsersRequest = getExcludedUsersRequest.User(excludedUserNames)
		getExcludedUsersRequest = getExcludedUsersRequest.Provider(citrixorchestration.IDENTITYPROVIDERTYPE_ALL)
		excludedUsersResponse, httpResp, err := citrixdaasclient.AddRequestData(getExcludedUsersRequest, client).Execute()
		if err != nil {
			diagnostics.AddError(
				"Error fetching delivery group user details",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, nil, err
		}

		err = verifyIdentityUserListCompleteness(excludedUserNames, excludedUsersResponse.GetItems())
		if err != nil {
			diagnostics.AddError(
				"Error fetching delivery group user details",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+err.Error(),
			)
			return nil, nil, err
		}

		excludedUsers = excludedUsersResponse.GetItems()
	}

	return includedUsers, excludedUsers, nil
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

	updatedAllowList := refreshUsersList(r.RestrictedAccessUsers.AllowList, simpleAccessPolicy.GetIncludedUsers())
	r.RestrictedAccessUsers.AllowList = updatedAllowList

	if simpleAccessPolicy.GetExcludedUserFilterEnabled() {
		updatedBlockList := refreshUsersList(r.RestrictedAccessUsers.BlockList, simpleAccessPolicy.GetExcludedUsers())
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
	parsedPowerTimeSchemes := parsePowerTimeSchemesClientToPluginModel(dgPowerTimeSchemes.GetItems())
	r.AutoscaleSettings.PowerTimeSchemes = preserveOrderInPowerTimeSchemes(r.AutoscaleSettings.PowerTimeSchemes, parsedPowerTimeSchemes)

	return r
}

func preserveOrderInPowerTimeSchemes(powerTimeSchemeInPlan, powerTimeSchemesInRemote []DeliveryGroupPowerTimeScheme) []DeliveryGroupPowerTimeScheme {
	planPowerTimeSchemesMap := map[string]int{}

	for index, powerTimeScheme := range powerTimeSchemeInPlan {
		planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()] = index
	}

	for _, powerTimeScheme := range powerTimeSchemesInRemote {
		index, exists := planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()]
		if !exists {
			powerTimeSchemeInPlan = append(powerTimeSchemeInPlan, powerTimeScheme)
		} else {
			powerTimeSchemeInPlan[index].PoolSizeSchedule = preserveOrderInPoolSizeSchedule(powerTimeSchemeInPlan[index].PoolSizeSchedule, powerTimeScheme.PoolSizeSchedule)
		}
		planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()] = -1
	}

	powerTimeSchemes := []DeliveryGroupPowerTimeScheme{}
	for _, powerTimeScheme := range powerTimeSchemeInPlan {
		if planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()] == -1 {
			powerTimeSchemes = append(powerTimeSchemes, powerTimeScheme)
		}
	}

	return powerTimeSchemes
}

func preserveOrderInPoolSizeSchedule(poolSizeScheduleInPlan, poolSizeScheduleInRemote []PowerTimeSchemePoolSizeScheduleRequestModel) []PowerTimeSchemePoolSizeScheduleRequestModel {
	if len(poolSizeScheduleInRemote) == 0 {
		return nil
	}

	planPoolSizeScheduleMap := map[string]int{}
	for index, poolSizeSchedule := range poolSizeScheduleInPlan {
		planPoolSizeScheduleMap[poolSizeSchedule.TimeRange.ValueString()] = index
	}

	for _, poolSizeSchedule := range poolSizeScheduleInRemote {
		_, exists := planPoolSizeScheduleMap[poolSizeSchedule.TimeRange.ValueString()]
		if !exists {
			poolSizeScheduleInPlan = append(poolSizeScheduleInPlan, poolSizeSchedule)
		}
		planPoolSizeScheduleMap[poolSizeSchedule.TimeRange.ValueString()] = -1
	}

	poolSizeSchedules := []PowerTimeSchemePoolSizeScheduleRequestModel{}
	for _, poolSizeSchedule := range poolSizeScheduleInPlan {
		if planPoolSizeScheduleMap[poolSizeSchedule.TimeRange.ValueString()] == -1 {
			poolSizeSchedules = append(poolSizeSchedules, poolSizeSchedule)
		}
	}

	return poolSizeSchedules
}

func refreshUsersList(users []basetypes.StringValue, usersInRemote []citrixorchestration.IdentityUserResponseModel) []basetypes.StringValue {
	samNamesMap := map[string]int{}
	upnMap := map[string]int{}

	for index, userInRemote := range usersInRemote {
		userSamName := userInRemote.GetSamName()
		userPrincipalName := userInRemote.GetPrincipalName()
		if userSamName != "" {
			samNamesMap[userSamName] = index
		}
		if userPrincipalName != "" {
			upnMap[userPrincipalName] = index
		}
	}

	res := []basetypes.StringValue{}
	for _, user := range users {
		userStringValue := user.ValueString()
		samRegex, _ := regexp.Compile(util.SamRegex)
		if samRegex.MatchString(userStringValue) {
			index, exists := samNamesMap[userStringValue]
			if !exists {
				continue
			}
			res = append(res, user)
			samNamesMap[userStringValue] = -1
			userPrincipalName := usersInRemote[index].GetPrincipalName()
			_, exists = upnMap[userPrincipalName]
			if exists {
				upnMap[userPrincipalName] = -1
			}

			continue
		}

		upnRegex, _ := regexp.Compile(util.UpnRegex)
		if upnRegex.MatchString(userStringValue) {
			index, exists := upnMap[userStringValue]
			if !exists {
				continue
			}
			res = append(res, user)
			upnMap[userStringValue] = -1
			samName := usersInRemote[index].GetSamName()
			_, exists = samNamesMap[samName]
			if exists {
				samNamesMap[samName] = -1
			}
		}
	}

	for samName, index := range samNamesMap {
		if index != -1 { // Users that are only in remote
			res = append(res, types.StringValue(samName))
		}
	}

	return res
}
