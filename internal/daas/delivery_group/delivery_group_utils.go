// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AssociatedMachineCatalogProperties struct {
	SessionSupport    citrixorchestration.SessionSupport
	IsPowerManaged    bool
	IsRemotePcCatalog bool
	IdentityType      citrixorchestration.IdentityType
	AllocationType    citrixorchestration.AllocationType
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

func validatePowerManagementSettings(ctx context.Context, diags *diag.Diagnostics, plan DeliveryGroupResourceModel, sessionSupport citrixorchestration.SessionSupport) (bool, string) {
	if plan.AutoscaleSettings.IsNull() || sessionSupport == citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION {
		return true, ""
	}
	autoscale := util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](ctx, diags, plan.AutoscaleSettings)

	errStringSuffix := "cannot be set for a Multisession catalog"

	if autoscale.PeakLogOffAction.ValueString() != "Nothing" {
		return false, "PeakLogOffAction " + errStringSuffix
	}

	if autoscale.OffPeakLogOffAction.ValueString() != "Nothing" {
		return false, "OffPeakLogOffAction " + errStringSuffix
	}

	if autoscale.PeakDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectAction " + errStringSuffix
	}

	if autoscale.PeakExtendedDisconnectAction.ValueString() != "Nothing" {
		return false, "PeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if autoscale.OffPeakDisconnectAction.ValueString() != "Nothing" {
		return false, "OffPeakDisconnectAction " + errStringSuffix
	}

	if autoscale.OffPeakExtendedDisconnectAction.ValueString() != "Nothing" {
		return false, "OffPeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if autoscale.PeakDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "PeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if autoscale.PeakExtendedDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "PeakExtendedDisconnectTimeoutMinutes " + errStringSuffix
	}

	if autoscale.OffPeakDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "OffPeakDisconnectTimeoutMinutes " + errStringSuffix
	}

	if autoscale.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64() != 0 {
		return false, "OffPeakExtendedDisconnectTimeoutMinutes " + errStringSuffix
	}

	return true, ""
}

func validateAndReturnMachineCatalogSessionSupport(ctx context.Context, client citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, dgMachineCatalogs []DeliveryGroupMachineCatalogModel, addErrorIfCatalogNotFound bool) (machineCatalogProperties AssociatedMachineCatalogProperties, err error) {
	var provisioningType *citrixorchestration.ProvisioningType
	var sessionSupport citrixorchestration.SessionSupport
	var allocationType citrixorchestration.AllocationType
	var identityType citrixorchestration.IdentityType
	var associatedMachineCatalogProperties AssociatedMachineCatalogProperties
	isPowerManaged := false
	isRemotePc := false
	for _, dgMachineCatalog := range dgMachineCatalogs {
		catalogId := dgMachineCatalog.MachineCatalog.ValueString()
		if catalogId == "" {
			continue
		}

		catalog, err := util.GetMachineCatalog(ctx, &client, diagnostics, catalogId, addErrorIfCatalogNotFound)

		if err != nil {
			return associatedMachineCatalogProperties, err
		}

		if provisioningType == nil {
			provisioningType = catalog.ProvisioningType
			isPowerManaged = catalog.GetIsPowerManaged()
			isRemotePc = catalog.GetIsRemotePC()
			provScheme := catalog.GetProvisioningScheme()
			identityType = provScheme.GetIdentityType()
			sessionSupport = catalog.GetSessionSupport()
			allocationType = catalog.GetAllocationType()
		}

		if *provisioningType != catalog.GetProvisioningType() {
			err := fmt.Errorf("associated_machine_catalogs must have catalogs with the same provsioning type")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"Ensure all associated Machine Catalogs have the same provisioning type.",
			)
			return associatedMachineCatalogProperties, err
		}

		provScheme := catalog.GetProvisioningScheme()

		if identityType != provScheme.GetIdentityType() {
			err := fmt.Errorf("associated_machine_catalogs must have catalogs with the same identity type in provisioning scheme")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"Ensure all associated Machine Catalogs have the same identity type in provisioning scheme.",
			)
			return associatedMachineCatalogProperties, err
		}

		if isPowerManaged != catalog.GetIsPowerManaged() {
			err := fmt.Errorf("all associated_machine_catalogs must either be power managed or non power managed")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"All associated Machine Catalogs must either be power managed or non power managed.",
			)
			return associatedMachineCatalogProperties, err
		}

		if isRemotePc != catalog.GetIsRemotePC() {
			err := fmt.Errorf("all associated_machine_catalogs must either be Remote PC or non Remote PC")
			diagnostics.AddError("Error validating associated Machine Catalogs",
				"All associated Machine Catalogs must either be Remote PC or non Remote PC.",
			)
			return associatedMachineCatalogProperties, err
		}

		if sessionSupport != "" && sessionSupport != catalog.GetSessionSupport() {
			err := fmt.Errorf("all associated machine catalogs must have the same session support")
			diagnostics.AddError("Error validating associated Machine Catalogs", "Ensure all associated Machine Catalogs have the same Session Support.")
			return associatedMachineCatalogProperties, err
		}

		if allocationType != "" && allocationType != catalog.GetAllocationType() {
			err := fmt.Errorf("all associated machine catalogs must have the same allocation type")
			diagnostics.AddError("Error validating associated Machine Catalogs", "Ensure all associated Machine Catalogs have the same Allocation Type.")
			return associatedMachineCatalogProperties, err
		}
	}

	associatedMachineCatalogProperties.SessionSupport = sessionSupport
	associatedMachineCatalogProperties.AllocationType = allocationType
	associatedMachineCatalogProperties.IdentityType = identityType
	associatedMachineCatalogProperties.IsPowerManaged = isPowerManaged
	associatedMachineCatalogProperties.IsRemotePcCatalog = isRemotePc

	return associatedMachineCatalogProperties, err
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
	for _, associatedMachineCatalog := range util.ObjectListToTypedArray[DeliveryGroupMachineCatalogModel](ctx, diagnostics, plan.AssociatedMachineCatalogs) {

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

func validatePowerTimeSchemes(ctx context.Context, diagnostics *diag.Diagnostics, powerTimeSchemes []DeliveryGroupPowerTimeScheme) {
	for _, powerTimeScheme := range powerTimeSchemes {
		if powerTimeScheme.PoolSizeSchedule.IsNull() {
			continue
		}

		hoursArray := make([]bool, 24)
		minutesArray := make([]bool, 24)

		hoursPoolSizeArray := make([]int, 24)
		minutesPoolSizeArray := make([]int, 24)

		for _, schedule := range util.ObjectListToTypedArray[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diagnostics, powerTimeScheme.PoolSizeSchedule) {
			if schedule.TimeRange.IsNull() {
				continue
			}
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

func validateRebootSchedules(ctx context.Context, diagnostics *diag.Diagnostics, rebootSchedules []DeliveryGroupRebootSchedule) {
	for _, rebootSchedule := range rebootSchedules {
		switch freq := rebootSchedule.Frequency.ValueString(); freq {
		case "Weekly":
			if len(rebootSchedule.DaysInWeek.Elements()) == 0 {
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

		if rebootSchedule.UseNaturalRebootSchedule.ValueBool() && !rebootSchedule.DeliveryGroupRebootNotificationToUsers.IsNull() {
			diagnostics.AddAttributeError(
				path.Root("reboot_notification_to_users"),
				"Incorrect Attribute Configuration",
				"Reboot notification to users can not be set for using natural reboot",
			)
		}

		if !rebootSchedule.DeliveryGroupRebootNotificationToUsers.IsNull() {
			notification := util.ObjectValueToTypedObject[DeliveryGroupRebootNotificationToUsers](ctx, diagnostics, rebootSchedule.DeliveryGroupRebootNotificationToUsers)
			if !notification.NotificationDurationMinutes.IsNull() && !notification.NotificationRepeatEvery5Minutes.IsNull() &&
				notification.NotificationDurationMinutes.ValueInt64() != 15 {
				diagnostics.AddAttributeError(
					path.Root("notification_repeat_every_5_minutes"),
					"Incorrect Attribute Configuration",
					"NotificationRepeatEvery5Minutes can only be set to true when NotificationDurationMinutes is 15 minutes",
				)
			}
		}

	}
}

func getRequestModelForDeliveryGroupCreate(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan DeliveryGroupResourceModel, associatedMachineCatalogProperties AssociatedMachineCatalogProperties) (citrixorchestration.CreateDeliveryGroupRequestModel, error) {
	desktops := util.ObjectListToTypedArray[DeliveryGroupDesktop](ctx, diagnostics, plan.Desktops)
	deliveryGroupDesktopsArray, err := verifyUsersAndParseDeliveryGroupDesktopsToClientModel(ctx, diagnostics, client, desktops)

	if err != nil {
		return citrixorchestration.CreateDeliveryGroupRequestModel{}, err
	}

	rebootSchedules := util.ObjectListToTypedArray[DeliveryGroupRebootSchedule](ctx, diagnostics, plan.RebootSchedules)
	deliveryGroupRebootScheduleArray := parseDeliveryGroupRebootScheduleToClientModel(ctx, diagnostics, rebootSchedules)

	var httpResp *http.Response
	var includedUserIds []string
	var excludedUserIds []string
	includedUsersFilterEnabled := false
	excludedUsersFilterEnabled := false
	if !plan.RestrictedAccessUsers.IsNull() {
		includedUsersFilterEnabled = true
		users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, diagnostics, plan.RestrictedAccessUsers)
		includedUsers := util.StringSetToStringArray(ctx, diagnostics, users.AllowList)
		includedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, includedUsers)

		if err != nil {
			diagnostics.AddError(
				"Error fetching user details for delivery group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return citrixorchestration.CreateDeliveryGroupRequestModel{}, err
		}

		if !users.BlockList.IsNull() {
			excludedUsersFilterEnabled = true
			excludedUsers := util.StringSetToStringArray(ctx, diagnostics, users.BlockList)
			excludedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, excludedUsers)

			if err != nil {
				diagnostics.AddError(
					"Error fetching user details for delivery group",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return citrixorchestration.CreateDeliveryGroupRequestModel{}, err
			}
		}
	}

	var simpleAccessPolicy citrixorchestration.SimplifiedAccessPolicyRequestModel
	simpleAccessPolicy.SetAllowAnonymous(plan.AllowAnonymousAccess.ValueBool())
	simpleAccessPolicy.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
	simpleAccessPolicy.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
	simpleAccessPolicy.SetIncludedUsers(includedUserIds)
	simpleAccessPolicy.SetExcludedUsers(excludedUserIds)

	var body citrixorchestration.CreateDeliveryGroupRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetRebootSchedules(deliveryGroupRebootScheduleArray)

	if !plan.AssociatedMachineCatalogs.IsNull() && len(plan.AssociatedMachineCatalogs.Elements()) > 0 {
		deliveryGroupMachineCatalogsArray := getDeliveryGroupAddMachinesRequest(util.ObjectListToTypedArray[DeliveryGroupMachineCatalogModel](ctx, diagnostics, plan.AssociatedMachineCatalogs))
		body.SetMachineCatalogs(deliveryGroupMachineCatalogsArray)
	}

	if !plan.SessionSupport.IsNull() {
		sessionSupport, err := citrixorchestration.NewSessionSupportFromValue(plan.SessionSupport.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Error creating Delivery Group",
				"Unsupported session support.",
			)
			return body, err
		}
		body.SetSessionSupport(*sessionSupport)
	}

	if !plan.SharingKind.IsNull() {
		sharingKind, err := citrixorchestration.NewSharingKindFromValue(plan.SharingKind.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Error creating Delivery Group",
				"Unsupported sharing kind.",
			)
			return body, err
		}
		body.SetSharingKind(*sharingKind)
	}

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
	if associatedMachineCatalogProperties.SessionSupport != "" && associatedMachineCatalogProperties.SessionSupport != citrixorchestration.SESSIONSUPPORT_MULTI_SESSION {
		deliveryKind = citrixorchestration.DELIVERYKIND_DESKTOPS_ONLY
	}
	body.SetDeliveryType(deliveryKind)
	body.SetDesktops(deliveryGroupDesktopsArray)
	body.SetDefaultDesktopPublishedName(plan.Name.ValueString())
	body.SetSimpleAccessPolicy(simpleAccessPolicy)
	body.SetPolicySetGuid(plan.PolicySetId.ValueString())
	if associatedMachineCatalogProperties.IdentityType == citrixorchestration.IDENTITYTYPE_AZURE_AD {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_AZURE_AD)
	} else if associatedMachineCatalogProperties.IdentityType == citrixorchestration.IDENTITYTYPE_WORKGROUP {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_LOCAL_MAPPED_ACCOUNT)
	} else {
		body.SetMachineLogOnType(citrixorchestration.MACHINELOGONTYPE_ACTIVE_DIRECTORY)
	}

	if !plan.MakeResourcesAvailableInLHC.IsNull() {
		body.SetReuseMachinesWithoutShutdownInOutage(plan.MakeResourcesAvailableInLHC.ValueBool())
	}

	if !plan.AutoscaleSettings.IsNull() {
		autoscale := util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](ctx, diagnostics, plan.AutoscaleSettings)
		body.SetAutoScaleEnabled(autoscale.AutoscaleEnabled.ValueBool())
		body.SetTimeZone(autoscale.Timezone.ValueString())
		body.SetPeakDisconnectTimeoutMinutes(int32(autoscale.PeakDisconnectTimeoutMinutes.ValueInt64()))
		body.SetPeakLogOffAction(getSessionChangeHostingActionValue(autoscale.PeakLogOffAction.ValueString()))
		body.SetPeakDisconnectAction(getSessionChangeHostingActionValue(autoscale.PeakDisconnectAction.ValueString()))
		body.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(autoscale.PeakExtendedDisconnectAction.ValueString()))
		body.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(autoscale.OffPeakLogOffAction.ValueString()))
		body.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(autoscale.OffPeakDisconnectAction.ValueString()))
		body.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(autoscale.OffPeakExtendedDisconnectAction.ValueString()))
		body.SetPeakExtendedDisconnectTimeoutMinutes(int32(autoscale.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetOffPeakDisconnectTimeoutMinutes(int32(autoscale.OffPeakDisconnectTimeoutMinutes.ValueInt64()))
		body.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(autoscale.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		body.SetPeakBufferSizePercent(int32(autoscale.PeakBufferSizePercent.ValueInt64()))
		body.SetOffPeakBufferSizePercent(int32(autoscale.OffPeakBufferSizePercent.ValueInt64()))
		body.SetPowerOffDelayMinutes(int32(autoscale.PowerOffDelayMinutes.ValueInt64()))
		body.SetDisconnectPeakIdleSessionAfterSeconds(int32(autoscale.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(autoscale.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		body.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(autoscale.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		body.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(autoscale.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := parsePowerTimeSchemesPluginToClientModel(ctx, diagnostics, util.ObjectListToTypedArray[DeliveryGroupPowerTimeScheme](ctx, diagnostics, autoscale.PowerTimeSchemes))
		body.SetPowerTimeSchemes(powerTimeSchemes)
	}

	if !plan.Scopes.IsNull() {
		plannedScopes := util.StringSetToStringArray(ctx, diagnostics, plan.Scopes)
		body.SetScopes(plannedScopes)
	}
	if !plan.StoreFrontServers.IsNull() {
		associatedStoreFrontServers := util.StringSetToStringArray(ctx, diagnostics, plan.StoreFrontServers)
		var storeFrontServersList []citrixorchestration.StoreFrontServerRequestModel
		for _, storeFrontServer := range associatedStoreFrontServers {
			storeFrontServerModel := citrixorchestration.StoreFrontServerRequestModel{}
			storeFrontServerModel.SetId(storeFrontServer)
			storeFrontServersList = append(storeFrontServersList, storeFrontServerModel)
		}
		body.SetStoreFrontServersForHostedReceiver(storeFrontServersList)
	}

	return body, nil
}

func getRequestModelForDeliveryGroupUpdate(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan DeliveryGroupResourceModel, currentDeliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel) (citrixorchestration.EditDeliveryGroupRequestModel, error) {
	desktops := util.ObjectListToTypedArray[DeliveryGroupDesktop](ctx, diagnostics, plan.Desktops)
	deliveryGroupDesktopsArray, err := verifyUsersAndParseDeliveryGroupDesktopsToClientModel(ctx, diagnostics, client, desktops)
	if err != nil {
		return citrixorchestration.EditDeliveryGroupRequestModel{}, err
	}
	rebootSchedules := util.ObjectListToTypedArray[DeliveryGroupRebootSchedule](ctx, diagnostics, plan.RebootSchedules)
	deliveryGroupRebootScheduleArray := parseDeliveryGroupRebootScheduleToClientModel(ctx, diagnostics, rebootSchedules)

	var httpResp *http.Response
	includedUserIds := []string{}
	excludedUserIds := []string{}
	includedUsersFilterEnabled := false
	excludedUsersFilterEnabled := false
	advancedAccessPolicies := []citrixorchestration.AdvancedAccessPolicyRequestModel{}

	allowedUser := citrixorchestration.ALLOWEDUSER_ANY_AUTHENTICATED

	if plan.AllowAnonymousAccess.ValueBool() {
		allowedUser = citrixorchestration.ALLOWEDUSER_ANY
	}

	if !plan.RestrictedAccessUsers.IsNull() {
		users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, diagnostics, plan.RestrictedAccessUsers)

		allowedUser = citrixorchestration.ALLOWEDUSER_FILTERED

		if plan.AllowAnonymousAccess.ValueBool() {
			allowedUser = citrixorchestration.ALLOWEDUSER_FILTERED_OR_ANONYMOUS
		}

		includedUsersFilterEnabled = true
		includedUsers := util.StringSetToStringArray(ctx, diagnostics, users.AllowList)
		includedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, includedUsers)
		if err != nil {
			diagnostics.AddError(
				"Error fetching user details for delivery group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return citrixorchestration.EditDeliveryGroupRequestModel{}, err
		}

		if !users.BlockList.IsNull() {
			excludedUsersFilterEnabled = true
			excludedUsers := util.StringSetToStringArray(ctx, diagnostics, users.BlockList)
			excludedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, excludedUsers)

			if err != nil {
				diagnostics.AddError(
					"Error fetching user details for delivery group",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return citrixorchestration.EditDeliveryGroupRequestModel{}, err
			}
		}
	}

	existingAdvancedAccessPolicies := currentDeliveryGroup.GetAdvancedAccessPolicy()
	for _, existingAdvancedAccessPolicy := range existingAdvancedAccessPolicies {
		var advancedAccessPolicyRequest citrixorchestration.AdvancedAccessPolicyRequestModel
		advancedAccessPolicyRequest.SetId(existingAdvancedAccessPolicy.GetId())
		advancedAccessPolicyRequest.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
		advancedAccessPolicyRequest.SetIncludedUsers(includedUserIds)
		advancedAccessPolicyRequest.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
		advancedAccessPolicyRequest.SetExcludedUsers(excludedUserIds)
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

	if !plan.Scopes.IsNull() {
		plannedScopes := util.StringSetToStringArray(ctx, diagnostics, plan.Scopes)
		editDeliveryGroupRequestBody.SetScopes(plannedScopes)
	}

	if plan.PolicySetId.ValueString() != "" {
		editDeliveryGroupRequestBody.SetPolicySetGuid(plan.PolicySetId.ValueString())
	} else {
		editDeliveryGroupRequestBody.SetPolicySetGuid(util.DefaultSitePolicySetId)
	}

	editDeliveryGroupRequestBody.SetReuseMachinesWithoutShutdownInOutage(plan.MakeResourcesAvailableInLHC.ValueBool())

	if !plan.AutoscaleSettings.IsNull() {
		autoscale := util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](ctx, diagnostics, plan.AutoscaleSettings)

		if autoscale.Timezone.ValueString() != "" {
			editDeliveryGroupRequestBody.SetTimeZone(autoscale.Timezone.ValueString())
		}

		editDeliveryGroupRequestBody.SetAutoScaleEnabled(autoscale.AutoscaleEnabled.ValueBool())
		editDeliveryGroupRequestBody.SetPeakDisconnectTimeoutMinutes(int32(autoscale.PeakDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetPeakLogOffAction(getSessionChangeHostingActionValue(autoscale.PeakLogOffAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakDisconnectAction(getSessionChangeHostingActionValue(autoscale.PeakDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(autoscale.PeakExtendedDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetPeakExtendedDisconnectTimeoutMinutes(int32(autoscale.PeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakDisconnectTimeoutMinutes(int32(autoscale.OffPeakDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakLogOffAction(getSessionChangeHostingActionValue(autoscale.OffPeakLogOffAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakDisconnectAction(getSessionChangeHostingActionValue(autoscale.OffPeakDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectAction(getSessionChangeHostingActionValue(autoscale.OffPeakExtendedDisconnectAction.ValueString()))
		editDeliveryGroupRequestBody.SetOffPeakExtendedDisconnectTimeoutMinutes(int32(autoscale.OffPeakExtendedDisconnectTimeoutMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetPeakBufferSizePercent(int32(autoscale.PeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetOffPeakBufferSizePercent(int32(autoscale.OffPeakBufferSizePercent.ValueInt64()))
		editDeliveryGroupRequestBody.SetPowerOffDelayMinutes(int32(autoscale.PowerOffDelayMinutes.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectPeakIdleSessionAfterSeconds(int32(autoscale.DisconnectPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetDisconnectOffPeakIdleSessionAfterSeconds(int32(autoscale.DisconnectOffPeakIdleSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffPeakDisconnectedSessionAfterSeconds(int32(autoscale.LogoffPeakDisconnectedSessionAfterSeconds.ValueInt64()))
		editDeliveryGroupRequestBody.SetLogoffOffPeakDisconnectedSessionAfterSeconds(int32(autoscale.LogoffOffPeakDisconnectedSessionAfterSeconds.ValueInt64()))

		powerTimeSchemes := parsePowerTimeSchemesPluginToClientModel(ctx, diagnostics, util.ObjectListToTypedArray[DeliveryGroupPowerTimeScheme](ctx, diagnostics, autoscale.PowerTimeSchemes))
		editDeliveryGroupRequestBody.SetPowerTimeSchemes(powerTimeSchemes)
	}

	storeFrontServersList := []citrixorchestration.StoreFrontServerRequestModel{}
	if !plan.StoreFrontServers.IsNull() {
		associatedStoreFrontServers := util.StringSetToStringArray(ctx, diagnostics, plan.StoreFrontServers)
		for _, storeFrontServer := range associatedStoreFrontServers {
			storeFrontServerModel := citrixorchestration.StoreFrontServerRequestModel{}
			storeFrontServerModel.SetId(storeFrontServer)
			storeFrontServersList = append(storeFrontServersList, storeFrontServerModel)
		}
	}
	editDeliveryGroupRequestBody.SetStoreFrontServersForHostedReceiver(storeFrontServersList)

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

func parsePowerTimeSchemesPluginToClientModel(ctx context.Context, diags *diag.Diagnostics, powerTimeSchemes []DeliveryGroupPowerTimeScheme) []citrixorchestration.PowerTimeSchemeRequestModel {
	res := []citrixorchestration.PowerTimeSchemeRequestModel{}
	for _, powerTimeScheme := range powerTimeSchemes {
		var powerTimeSchemeRequest citrixorchestration.PowerTimeSchemeRequestModel

		var daysOfWeek []citrixorchestration.TimeSchemeDays
		for _, dayOfWeek := range util.StringSetToStringArray(ctx, diags, powerTimeScheme.DaysOfWeek) {
			timeSchemeDay := getTimeSchemeDayValue(dayOfWeek)
			daysOfWeek = append(daysOfWeek, timeSchemeDay)
		}

		var poolSizeScheduleRequests []citrixorchestration.PoolSizeScheduleRequestModel
		for _, poolSizeSchedule := range util.ObjectListToTypedArray[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diags, powerTimeScheme.PoolSizeSchedule) {
			var poolSizeScheduleRequest citrixorchestration.PoolSizeScheduleRequestModel
			poolSizeScheduleRequest.SetTimeRange(poolSizeSchedule.TimeRange.ValueString())
			poolSizeScheduleRequest.SetPoolSize(int32(poolSizeSchedule.PoolSize.ValueInt64()))
			poolSizeScheduleRequests = append(poolSizeScheduleRequests, poolSizeScheduleRequest)
		}

		peakTimeRanges := util.StringSetToStringArray(ctx, diags, powerTimeScheme.PeakTimeRanges)

		powerTimeSchemeRequest.SetDisplayName(powerTimeScheme.DisplayName.ValueString())
		powerTimeSchemeRequest.SetPeakTimeRanges(peakTimeRanges)
		powerTimeSchemeRequest.SetPoolUsingPercentage(powerTimeScheme.PoolUsingPercentage.ValueBool())
		powerTimeSchemeRequest.SetDaysOfWeek(daysOfWeek)
		powerTimeSchemeRequest.SetPoolSizeSchedule(poolSizeScheduleRequests)
		res = append(res, powerTimeSchemeRequest)
	}

	return res
}

func parsePowerTimeSchemesClientToPluginModel(ctx context.Context, diags *diag.Diagnostics, powerTimeSchemesResponse []citrixorchestration.PowerTimeSchemeResponseModel) []DeliveryGroupPowerTimeScheme {
	var res []DeliveryGroupPowerTimeScheme
	for _, powerTimeSchemeResponse := range powerTimeSchemesResponse {
		var deliveryGroupPowerTimeScheme DeliveryGroupPowerTimeScheme

		var daysOfWeek []string
		for _, dayOfWeek := range powerTimeSchemeResponse.GetDaysOfWeek() {
			timeSchemeDay := string(dayOfWeek)
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
		deliveryGroupPowerTimeScheme.PeakTimeRanges = util.StringArrayToStringSet(ctx, diags, powerTimeSchemeResponse.GetPeakTimeRanges())
		deliveryGroupPowerTimeScheme.PoolUsingPercentage = types.BoolValue(powerTimeSchemeResponse.GetPoolUsingPercentage())
		deliveryGroupPowerTimeScheme.DaysOfWeek = util.StringArrayToStringSet(ctx, diags, daysOfWeek)
		deliveryGroupPowerTimeScheme.PoolSizeSchedule = util.TypedArrayToObjectList[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diags, poolSizeScheduleRequests)

		res = append(res, deliveryGroupPowerTimeScheme)
	}

	return res
}

func parseDeliveryGroupRebootScheduleToClientModel(ctx context.Context, diags *diag.Diagnostics, rebootSchedules []DeliveryGroupRebootSchedule) []citrixorchestration.RebootScheduleRequestModel {
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
			rebootScheduleRequest.SetDaysInWeek(getRebootScheduleDaysInWeekActionValue(util.StringSetToStringArray(ctx, diags, rebootSchedule.DaysInWeek)))
		}
		if rebootSchedule.Frequency.ValueString() == "Monthly" {
			rebootScheduleRequest.SetWeekInMonth(getRebootScheduleWeekActionValue(rebootSchedule.WeekInMonth.ValueString()))
			rebootScheduleRequest.SetDayInMonth(getRebootScheduleDaysActionValue(rebootSchedule.DayInMonth.ValueString()))
		}

		if !rebootSchedule.DeliveryGroupRebootNotificationToUsers.IsNull() && !rebootSchedule.UseNaturalRebootSchedule.ValueBool() {
			notification := util.ObjectValueToTypedObject[DeliveryGroupRebootNotificationToUsers](ctx, diags, rebootSchedule.DeliveryGroupRebootNotificationToUsers)

			rebootScheduleRequest.SetWarningDurationMinutes(int32(notification.NotificationDurationMinutes.ValueInt64())) //can only be 1 5 15, or 0 means no warning
			rebootScheduleRequest.SetWarningTitle(notification.NotificationTitle.ValueString())
			rebootScheduleRequest.SetWarningMessage(notification.NotificationMessage.ValueString())
			if notification.NotificationRepeatEvery5Minutes.ValueBool() {
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

func (schedule DeliveryGroupRebootSchedule) RefreshListItem(ctx context.Context, diags *diag.Diagnostics, rebootSchedule citrixorchestration.RebootScheduleResponseModel) util.ModelWithAttributes {
	schedule.Name = types.StringValue(rebootSchedule.GetName())
	if rebootSchedule.GetDescription() != "" {
		schedule.Description = types.StringValue(rebootSchedule.GetDescription())
	}

	schedule.RebootScheduleEnabled = types.BoolValue(rebootSchedule.GetEnabled())
	if rebootSchedule.GetRestrictToTag().Id.Get() != nil {
		schedule.RestrictToTag = types.StringValue(*rebootSchedule.GetRestrictToTag().Name.Get())
	}
	schedule.IgnoreMaintenanceMode = types.BoolValue(rebootSchedule.GetIgnoreMaintenanceMode()) //bug in orchestration side
	schedule.Frequency = types.StringValue(string(rebootSchedule.GetFrequency()))

	if rebootSchedule.GetFrequency() == citrixorchestration.REBOOTSCHEDULEFREQUENCY_WEEKLY {
		res := []string{}
		for _, scheduleDay := range rebootSchedule.GetDaysInWeek() {
			res = append(res, string(scheduleDay))
		}
		schedule.DaysInWeek = util.StringArrayToStringSet(ctx, diags, res)
	} else if rebootSchedule.GetFrequency() == citrixorchestration.REBOOTSCHEDULEFREQUENCY_MONTHLY {
		schedule.WeekInMonth = types.StringValue(string(rebootSchedule.GetWeekInMonth()))
		schedule.DayInMonth = types.StringValue(string(rebootSchedule.GetDayInMonth()))
	} else {
		schedule.StartDate = types.StringValue(rebootSchedule.GetStartDate())
	}
	schedule.FrequencyFactor = types.Int64Value(int64(rebootSchedule.GetFrequencyFactor()))
	schedule.StartTime = types.StringValue(rebootSchedule.GetStartTime()[:5])

	if schedule.StartDate.IsNull() {
		schedule.StartDate = types.StringValue(rebootSchedule.GetStartDate())
	}

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
			schedule.DeliveryGroupRebootNotificationToUsers = util.TypedObjectToObjectValue(ctx, diags, notif)
		}
	}
	return schedule
}

func (dgDesktop DeliveryGroupDesktop) RefreshListItem(ctx context.Context, diagnostics *diag.Diagnostics, desktop citrixorchestration.DesktopResponseModel) util.ModelWithAttributes {
	dgDesktop.PublishedName = types.StringValue(desktop.GetPublishedName())
	dgDesktop.DesktopDescription = types.StringValue(desktop.GetDescription())

	dgDesktop.Enabled = types.BoolValue(desktop.GetEnabled())
	sessionReconnection := desktop.GetSessionReconnection()
	if sessionReconnection == citrixorchestration.SESSIONRECONNECTION_ALWAYS {
		dgDesktop.EnableSessionRoaming = types.BoolValue(true)
	} else {
		dgDesktop.EnableSessionRoaming = types.BoolValue(false)
	}

	var users RestrictedAccessUsers
	if !desktop.GetIncludedUserFilterEnabled() {
		if attributes, err := util.AttributeMapFromObject(users); err == nil {
			dgDesktop.RestrictedAccessUsers = types.ObjectNull(attributes)
		} else {
			diagnostics.AddWarning("Error when creating null RestrictedAccessUsers", err.Error())
		}
		return dgDesktop
	}

	users = util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, diagnostics, dgDesktop.RestrictedAccessUsers)

	includedUsers := desktop.GetIncludedUsers()
	excludedUsers := desktop.GetExcludedUsers()

	if len(includedUsers) == 0 {
		users.AllowList = types.SetNull(types.StringType)
	} else {
		users.AllowList = util.RefreshUsersList(ctx, diagnostics, users.AllowList, includedUsers)
	}

	if len(excludedUsers) == 0 {
		users.BlockList = types.SetNull(types.StringType)
	} else {
		users.BlockList = util.RefreshUsersList(ctx, diagnostics, users.BlockList, excludedUsers)
	}
	usersObj := util.TypedObjectToObjectValue(ctx, diagnostics, users)
	dgDesktop.RestrictedAccessUsers = usersObj

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

func verifyUsersAndParseDeliveryGroupDesktopsToClientModel(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, deliveryGroupDesktops []DeliveryGroupDesktop) ([]citrixorchestration.DesktopRequestModel, error) {
	desktopRequests := []citrixorchestration.DesktopRequestModel{}

	if deliveryGroupDesktops == nil {
		return desktopRequests, nil
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

		includedUserIds := []string{}
		excludedUserIds := []string{}
		var err error
		var httpResp *http.Response
		includedUsersFilterEnabled := false
		excludedUsersFilterEnabled := false
		if !deliveryGroupDesktop.RestrictedAccessUsers.IsNull() {
			users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, diagnostics, deliveryGroupDesktop.RestrictedAccessUsers)

			includedUsersFilterEnabled = true
			includedUsers := util.StringSetToStringArray(ctx, diagnostics, users.AllowList)

			// Call identity to make sure users exist. Extract the Ids from the reponse
			includedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, includedUsers)
			if err != nil {
				diagnostics.AddError(
					"Error fetching user details for delivery group",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return desktopRequests, err
			}

			if !users.BlockList.IsNull() {
				excludedUsersFilterEnabled = true
				excludedUsers := util.StringSetToStringArray(ctx, diagnostics, users.BlockList)
				excludedUserIds, httpResp, err = util.GetUserIdsUsingIdentity(ctx, client, excludedUsers)

				if err != nil {
					diagnostics.AddError(
						"Error fetching user details for delivery group",
						"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
							"\nError message: "+util.ReadClientError(err),
					)
					return desktopRequests, err
				}
			}
		}

		desktopRequest.SetIncludedUserFilterEnabled(includedUsersFilterEnabled)
		desktopRequest.SetExcludedUserFilterEnabled(excludedUsersFilterEnabled)
		desktopRequest.SetIncludedUsers(includedUserIds)
		desktopRequest.SetExcludedUsers(excludedUserIds)

		desktopRequests = append(desktopRequests, desktopRequest)
	}

	return desktopRequests, nil
}

func getTimeSchemeDayValue(v string) citrixorchestration.TimeSchemeDays {
	timeSchemeDay, err := citrixorchestration.NewTimeSchemeDaysFromValue(v)
	if err != nil {
		return citrixorchestration.TIMESCHEMEDAYS_UNKNOWN
	}

	return *timeSchemeDay
}

func (r DeliveryGroupResourceModel) updatePlanWithRebootSchedule(ctx context.Context, diagnostics *diag.Diagnostics, rebootSchedules *citrixorchestration.RebootScheduleResponseModelCollection) DeliveryGroupResourceModel {
	schedules := util.RefreshListValueProperties[DeliveryGroupRebootSchedule, citrixorchestration.RebootScheduleResponseModel](ctx, diagnostics, r.RebootSchedules, rebootSchedules.GetItems(), util.GetOrchestrationRebootScheduleKey)
	r.RebootSchedules = schedules
	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithAssociatedCatalogs(ctx context.Context, diags *diag.Diagnostics, machines *citrixorchestration.MachineResponseModelCollection) DeliveryGroupResourceModel {
	machineCatalogMap := map[string]int{}

	for _, machine := range machines.GetItems() {
		machineCatalog := machine.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		machineCatalogMap[machineCatalogId] += 1
	}

	var associatedMachineCatalogs []DeliveryGroupMachineCatalogModel
	if !r.AssociatedMachineCatalogs.IsNull() {
		associatedMachineCatalogs = []DeliveryGroupMachineCatalogModel{}
	}
	for key, val := range machineCatalogMap {
		var deliveryGroupMachineCatalogModel DeliveryGroupMachineCatalogModel
		deliveryGroupMachineCatalogModel.MachineCatalog = types.StringValue(key)
		deliveryGroupMachineCatalogModel.MachineCount = types.Int64Value(int64(val))
		associatedMachineCatalogs = append(associatedMachineCatalogs, deliveryGroupMachineCatalogModel)
	}

	r.AssociatedMachineCatalogs = util.TypedArrayToObjectList[DeliveryGroupMachineCatalogModel](ctx, diags, associatedMachineCatalogs)

	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithDesktops(ctx context.Context, diagnostics *diag.Diagnostics, deliveryGroupDesktops *citrixorchestration.DesktopResponseModelCollection) DeliveryGroupResourceModel {
	desktops := util.RefreshListValueProperties[DeliveryGroupDesktop, citrixorchestration.DesktopResponseModel](ctx, diagnostics, r.Desktops, deliveryGroupDesktops.GetItems(), util.GetOrchestrationDesktopKey)
	r.Desktops = desktops
	return r
}

func updateDeliveryGroupAndDesktopUsers(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, deliveryGroupDesktops *citrixorchestration.DesktopResponseModelCollection) (*citrixorchestration.DeliveryGroupDetailResponseModel, *citrixorchestration.DesktopResponseModelCollection, error) {
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

func updateIdentityUserDetails(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, includedUsers []citrixorchestration.IdentityUserResponseModel, excludedUsers []citrixorchestration.IdentityUserResponseModel) ([]citrixorchestration.IdentityUserResponseModel, []citrixorchestration.IdentityUserResponseModel, error) {
	includedUserNames := []string{}
	var err error
	var httpResp *http.Response
	for _, includedUser := range includedUsers {
		if includedUser.GetPrincipalName() != "" {
			includedUserNames = append(includedUserNames, includedUser.GetPrincipalName())
		} else if includedUser.GetSamName() != "" {
			includedUserNames = append(includedUserNames, includedUser.GetSamName())
		}
	}

	if len(includedUserNames) > 0 {
		includedUsers, httpResp, err = util.GetUsersUsingIdentity(ctx, client, includedUserNames)
		if err != nil {
			diagnostics.AddError(
				"Error fetching user details for delivery group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, nil, err
		}
	}

	excludedUserNames := []string{}
	for _, excludedUser := range excludedUsers {
		if excludedUser.GetPrincipalName() != "" {
			excludedUserNames = append(excludedUserNames, excludedUser.GetPrincipalName())
		} else if excludedUser.GetSamName() != "" {
			excludedUserNames = append(excludedUserNames, excludedUser.GetSamName())
		}
	}

	if len(excludedUserNames) > 0 {
		excludedUsers, httpResp, err = util.GetUsersUsingIdentity(ctx, client, excludedUserNames)
		if err != nil {
			diagnostics.AddError(
				"Error fetching user details for delivery group",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return nil, nil, err
		}
	}

	return includedUsers, excludedUsers, nil
}

func (r DeliveryGroupResourceModel) updatePlanWithRestrictedAccessUsers(ctx context.Context, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel) DeliveryGroupResourceModel {
	simpleAccessPolicy := deliveryGroup.GetSimpleAccessPolicy()

	if !r.AllowAnonymousAccess.IsNull() {
		r.AllowAnonymousAccess = types.BoolValue(simpleAccessPolicy.GetAllowAnonymous())
	}

	if !simpleAccessPolicy.GetIncludedUserFilterEnabled() || r.RestrictedAccessUsers.IsNull() {
		if attributes, err := util.AttributeMapFromObject(RestrictedAccessUsers{}); err == nil {
			r.RestrictedAccessUsers = types.ObjectNull(attributes)
		} else {
			diagnostics.AddWarning("Error when creating null RestrictedAccessUsers", err.Error())
		}
		return r
	}

	users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, diagnostics, r.RestrictedAccessUsers)

	remoteIncludedUsers := simpleAccessPolicy.GetIncludedUsers()
	if len(remoteIncludedUsers) == 0 {
		users.AllowList = types.SetNull(types.StringType)
	} else {
		users.AllowList = util.RefreshUsersList(ctx, diagnostics, users.AllowList, simpleAccessPolicy.GetIncludedUsers())
	}

	if simpleAccessPolicy.GetExcludedUserFilterEnabled() {
		if len(simpleAccessPolicy.GetExcludedUsers()) == 0 {
			users.BlockList = types.SetNull(types.StringType)
		} else {
			users.BlockList = util.RefreshUsersList(ctx, diagnostics, users.BlockList, simpleAccessPolicy.GetExcludedUsers())
		}
	}

	r.RestrictedAccessUsers = util.TypedObjectToObjectValue(ctx, diagnostics, users)

	return r
}

func (r DeliveryGroupResourceModel) updatePlanWithAutoscaleSettings(ctx context.Context, diags *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, dgPowerTimeSchemes *citrixorchestration.PowerTimeSchemeResponseModelCollection) DeliveryGroupResourceModel {
	if r.AutoscaleSettings.IsNull() {
		return r
	}

	autoscale := util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](context.Background(), nil, r.AutoscaleSettings)
	autoscale.AutoscaleEnabled = types.BoolValue(deliveryGroup.GetAutoScaleEnabled())

	if !autoscale.Timezone.IsNull() {
		autoscale.Timezone = types.StringValue(deliveryGroup.GetTimeZone())
	}

	autoscale.PeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakDisconnectTimeoutMinutes()))
	autoscale.PeakLogOffAction = types.StringValue(string(deliveryGroup.GetPeakLogOffAction()))
	autoscale.PeakDisconnectAction = types.StringValue(string(deliveryGroup.GetPeakDisconnectAction()))
	autoscale.PeakExtendedDisconnectAction = types.StringValue(string(deliveryGroup.GetPeakExtendedDisconnectAction()))
	autoscale.PeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetPeakExtendedDisconnectTimeoutMinutes()))
	autoscale.OffPeakDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakDisconnectTimeoutMinutes()))
	autoscale.OffPeakLogOffAction = types.StringValue(string(deliveryGroup.GetOffPeakLogOffAction()))
	autoscale.OffPeakDisconnectAction = types.StringValue(string(deliveryGroup.GetOffPeakExtendedDisconnectAction()))
	autoscale.OffPeakExtendedDisconnectAction = types.StringValue(string(deliveryGroup.GetOffPeakExtendedDisconnectAction()))
	autoscale.OffPeakExtendedDisconnectTimeoutMinutes = types.Int64Value(int64(deliveryGroup.GetOffPeakExtendedDisconnectTimeoutMinutes()))
	autoscale.PeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetPeakBufferSizePercent()))
	autoscale.OffPeakBufferSizePercent = types.Int64Value(int64(deliveryGroup.GetOffPeakBufferSizePercent()))
	autoscale.PowerOffDelayMinutes = types.Int64Value(int64(deliveryGroup.GetPowerOffDelayMinutes()))
	autoscale.DisconnectPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectPeakIdleSessionAfterSeconds()))
	autoscale.DisconnectOffPeakIdleSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetDisconnectOffPeakIdleSessionAfterSeconds()))
	autoscale.LogoffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffPeakDisconnectedSessionAfterSeconds()))
	autoscale.LogoffOffPeakDisconnectedSessionAfterSeconds = types.Int64Value(int64(deliveryGroup.GetLogoffOffPeakDisconnectedSessionAfterSeconds()))

	parsedPowerTimeSchemes := parsePowerTimeSchemesClientToPluginModel(ctx, diags, dgPowerTimeSchemes.GetItems())
	autoscalePowerTimeSchemes := util.ObjectListToTypedArray[DeliveryGroupPowerTimeScheme](ctx, diags, autoscale.PowerTimeSchemes)
	parsedPowerTimeSchemes = preserveOrderInPowerTimeSchemes(ctx, diags, autoscalePowerTimeSchemes, parsedPowerTimeSchemes)
	autoscale.PowerTimeSchemes = util.TypedArrayToObjectList[DeliveryGroupPowerTimeScheme](ctx, diags, parsedPowerTimeSchemes)

	r.AutoscaleSettings = util.TypedObjectToObjectValue(ctx, diags, autoscale)
	return r
}

func preserveOrderInPowerTimeSchemes(ctx context.Context, diags *diag.Diagnostics, powerTimeSchemeInPlan, powerTimeSchemesInRemote []DeliveryGroupPowerTimeScheme) []DeliveryGroupPowerTimeScheme {
	planPowerTimeSchemesMap := map[string]int{}

	for index, powerTimeScheme := range powerTimeSchemeInPlan {
		planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()] = index
	}

	for _, powerTimeScheme := range powerTimeSchemesInRemote {
		index, exists := planPowerTimeSchemesMap[powerTimeScheme.DisplayName.ValueString()]
		if !exists {
			powerTimeSchemeInPlan = append(powerTimeSchemeInPlan, powerTimeScheme)
		} else {
			updatedPoolSizeSchedule := preserveOrderInPoolSizeSchedule(
				util.ObjectListToTypedArray[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diags, powerTimeSchemeInPlan[index].PoolSizeSchedule),
				util.ObjectListToTypedArray[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diags, powerTimeScheme.PoolSizeSchedule))
			powerTimeSchemeInPlan[index].PoolSizeSchedule = util.TypedArrayToObjectList[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, diags, updatedPoolSizeSchedule)
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
