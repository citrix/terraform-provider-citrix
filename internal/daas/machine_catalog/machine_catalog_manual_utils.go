// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getMachinesForManualCatalogs(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, machineAccounts []MachineAccountsModel) ([]citrixorchestration.AddMachineToMachineCatalogRequestModel, *http.Response, error) {
	if machineAccounts == nil {
		return nil, nil, nil
	}

	addMachineRequestList := []citrixorchestration.AddMachineToMachineCatalogRequestModel{}
	for _, machineAccount := range machineAccounts {
		hypervisorId := machineAccount.Hypervisor.ValueString()
		var hypervisor *citrixorchestration.HypervisorDetailResponseModel
		var err error
		if hypervisorId != "" {
			hypervisor, err = util.GetHypervisor(ctx, client, nil, hypervisorId)

			if err != nil {
				return nil, nil, err
			}
		}

		machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)

		// Verify machine accounts using Identity API
		httpResp, err := verifyMachinesUsingIdentity(ctx, client, machines)
		if err != nil {
			return nil, httpResp, err
		}

		for _, machine := range machines {
			addMachineRequest := citrixorchestration.AddMachineToMachineCatalogRequestModel{}
			addMachineRequest.SetMachineName(machine.MachineAccount.ValueString())

			if hypervisorId == "" {
				// no hypervisor, non-power managed manual catalog
				addMachineRequestList = append(addMachineRequestList, addMachineRequest)
				continue
			}

			machineName := machine.MachineName.ValueString()
			var vmId string
			connectionType := hypervisor.GetConnectionType()
			switch connectionType {
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
				if machine.Region.IsNull() || machine.ResourceGroupName.IsNull() {
					return nil, nil, fmt.Errorf("region and resource_group_name are required for Azure")
				}
				region, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, "", machine.Region.ValueString(), "Region", "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				regionPath := region.GetXDPath()
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, fmt.Sprintf("%s\\vm.folder", regionPath), machineName, util.VirtualMachineResourceType, machine.ResourceGroupName.ValueString(), hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
				if machine.AvailabilityZone.IsNull() {
					return nil, nil, fmt.Errorf("availability_zone is required for AWS")
				}
				availabilityZone, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, "", machine.AvailabilityZone.ValueString(), "", "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				availabilityZonePath := availabilityZone.GetXDPath()
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, availabilityZonePath, machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
				if machine.Region.IsNull() || machine.ProjectName.IsNull() {
					return nil, nil, fmt.Errorf("region and project_name are required for GCP")
				}
				projectName, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, "", machine.ProjectName.ValueString(), "", "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				projectNamePath := projectName.GetXDPath()
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, fmt.Sprintf("%s\\%s.region", projectNamePath, machine.Region.ValueString()), machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
				if machine.Datacenter.IsNull() || machine.Host.IsNull() {
					return nil, nil, fmt.Errorf("datacenter and host are required for vSphere")
				}

				folderPath := hypervisor.GetXDPath()
				datacenter, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, folderPath, machine.Datacenter.ValueString(), "datacenter", "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}

				folderPath = datacenter.GetXDPath()

				if !machine.ClusterFolderPath.IsNull() {
					folders := strings.Split(machine.ClusterFolderPath.ValueString(), "\\")
					for _, folder := range folders {
						folderPath = fmt.Sprintf("%s\\%s.folder", folderPath, folder)
					}
				}

				if !machine.Cluster.IsNull() {
					cluster, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, folderPath, machine.Cluster.ValueString(), "cluster", "", hypervisor)
					if err != nil {
						return nil, httpResp, err
					}
					folderPath = cluster.GetXDPath()
				}

				host, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, folderPath, machine.Host.ValueString(), "computeresource", "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				hostPath := host.GetXDPath()
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, hostPath, machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, "", machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
				host, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, "", machine.Host.ValueString(), util.HostResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, host.GetFullName(), machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, httpResp, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
				if hypervisor.GetPluginId() == util.NUTANIX_PLUGIN_ID {
					hypervisorXdPath := hypervisor.GetXDPath()
					vm, httpResp, err := util.GetSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, fmt.Sprintf("%s\\VirtualMachines.folder", hypervisorXdPath), machineName, util.VirtualMachineResourceType, "", hypervisor)
					if err != nil {
						return nil, httpResp, err
					}
					vmId = vm.GetId()
				}
			}

			addMachineRequest.SetHostedMachineId(vmId)
			addMachineRequest.SetHypervisorConnection(hypervisorId)

			addMachineRequestList = append(addMachineRequestList, addMachineRequest)
		}
	}

	return addMachineRequestList, nil, nil
}

func deleteMachinesFromManualCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, deleteMachinesList map[string]bool, catalogNameOrId string) error {

	if len(deleteMachinesList) < 1 {
		// nothing to delete
		return nil
	}

	getMachinesResponse, err := util.GetMachineCatalogMachines(ctx, client, &resp.Diagnostics, catalogNameOrId)
	if err != nil {
		return err
	}

	machinesToDelete := []citrixorchestration.MachineResponseModel{}
	for _, machine := range getMachinesResponse {
		if deleteMachinesList[strings.ToLower(machine.GetName())] {
			machinesToDelete = append(machinesToDelete, machine)
		}
	}

	return deleteMachinesFromCatalog(ctx, client, resp, ProvisioningSchemeModel{}, machinesToDelete, catalogNameOrId, false, []MachineADAccountModel{})
}

func addMachinesToManualCatalog(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, addMachinesList []MachineAccountsModel, catalogIdOrName string) error {

	if len(addMachinesList) < 1 {
		// no machines to add
		return nil
	}

	addMachinesRequest, httpResp, err := getMachinesForManualCatalogs(ctx, diagnostics, client, addMachinesList)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding machines(s) to Machine Catalog "+catalogIdOrName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nFailed to resolve machines, err: "+err.Error(),
		)

		return err
	}

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, client, ProvisioningSchemeModel{}, false)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogIdOrName,
			"TransactionId: "+txId+
				"\nCould not add machine to Machine Catalog, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/Machines", catalogIdOrName)
	for i := 0; i < len(addMachinesRequest); i++ {
		addMachineRequestStringBody, err := util.ConvertToString(addMachinesRequest[i])
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding Machine to Machine Catalog "+catalogIdOrName,
				"An unexpected error occurred: "+err.Error(),
			)
			return err
		}
		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetReference(strconv.Itoa(i))
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetBody(addMachineRequestStringBody)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	var batchRequestModel citrixorchestration.BatchRequestModel
	batchRequestModel.SetItems(batchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding machine(s) to Machine Catalog "+catalogIdOrName,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(addMachinesList) {
		errMsg := fmt.Sprintf("An error occurred while adding machine(s) to the Machine Catalog. %d of %d machines were added to the Machine Catalog.", successfulJobs, len(addMachinesList))
		err = fmt.Errorf(errMsg)
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogIdOrName,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}

	return nil
}

func createAddAndRemoveMachinesListForManualCatalogs(ctx context.Context, diagnostics *diag.Diagnostics, state, plan MachineCatalogResourceModel) ([]MachineAccountsModel, map[string]bool) {
	addMachinesList := []MachineAccountsModel{}
	existingMachineAccounts := map[string]map[string]bool{}

	// create map for existing machines marking all machines for deletion
	if !state.MachineAccounts.IsNull() {
		machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, state.MachineAccounts)
		for _, machineAccount := range machineAccounts {
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			for _, machine := range machines {
				machineMap, exists := existingMachineAccounts[machineAccount.Hypervisor.ValueString()]
				if !exists {
					existingMachineAccounts[machineAccount.Hypervisor.ValueString()] = map[string]bool{}
					machineMap = existingMachineAccounts[machineAccount.Hypervisor.ValueString()]
				}
				machineMap[strings.ToLower(machine.MachineAccount.ValueString())] = true
			}
		}
	}

	// iterate over plan and if machine already exists, mark false for deletion. If not, add it to the addMachineList
	if !plan.MachineAccounts.IsNull() {
		machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, plan.MachineAccounts)
		for _, machineAccount := range machineAccounts {
			machineAccountMachines := []MachineCatalogMachineModel{}
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			for _, machine := range machines {
				if existingMachineAccounts[machineAccount.Hypervisor.ValueString()][strings.ToLower(machine.MachineAccount.ValueString())] {
					// Machine exists. Mark false for deletion
					existingMachineAccounts[machineAccount.Hypervisor.ValueString()][strings.ToLower(machine.MachineAccount.ValueString())] = false
				} else {
					// Machine does not exist and needs to be added
					machineAccountMachines = append(machineAccountMachines, machine)
				}
			}

			if len(machineAccountMachines) > 0 {
				var addMachineAccount MachineAccountsModel
				addMachineAccount.Hypervisor = machineAccount.Hypervisor
				addMachineAccount.Machines = util.TypedArrayToObjectList[MachineCatalogMachineModel](ctx, diagnostics, machineAccountMachines)
				addMachinesList = append(addMachinesList, addMachineAccount)
			}
		}
	}

	deleteMachinesMap := map[string]bool{}

	for _, machineMap := range existingMachineAccounts {
		for machineName, canBeDeleted := range machineMap {
			if canBeDeleted {
				deleteMachinesMap[machineName] = true
			}
		}
	}

	return addMachinesList, deleteMachinesMap
}

func (r MachineCatalogResourceModel) updateCatalogWithMachines(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, machines []citrixorchestration.MachineResponseModel) MachineCatalogResourceModel {
	if len(machines) == 0 {
		r.MachineAccounts = util.TypedArrayToObjectList[MachineAccountsModel](ctx, diagnostics, nil)
		return r
	}

	machineMapFromRemote := map[string]citrixorchestration.MachineResponseModel{}
	for _, machine := range machines {
		machineMapFromRemote[strings.ToLower(machine.GetName())] = machine
	}

	if !r.MachineAccounts.IsNull() {
		machinesNotPresetInRemote := map[string]bool{}
		machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, r.MachineAccounts)
		for _, machineAccount := range machineAccounts {
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			for _, machineFromPlan := range machines {
				machineFromPlanName := machineFromPlan.MachineAccount.ValueString()
				machineFromRemote, exists := machineMapFromRemote[strings.ToLower(machineFromPlanName)]
				if !exists {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				hosting := machineFromRemote.GetHosting()
				hypervisor := hosting.GetHypervisorConnection()
				hypervisorId := hypervisor.GetId()
				hostingServerName := hosting.GetHostingServerName()
				hostedMachineName := hosting.GetHostedMachineName()

				if !strings.EqualFold(hypervisorId, machineAccount.Hypervisor.ValueString()) {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				if hypervisorId == "" {
					delete(machineMapFromRemote, strings.ToLower(machineFromPlanName))
					continue
				}

				hyp, err := util.GetHypervisor(ctx, client, nil, hypervisorId)
				if err != nil {
					machinesNotPresetInRemote[strings.ToLower(machineFromPlanName)] = true
					continue
				}

				connectionType := hyp.GetConnectionType()
				hostedMachineId := hosting.GetHostedMachineId()
				switch connectionType {
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
					if hostedMachineId != "" {
						hostedMachineIdArray := strings.Split(hostedMachineId, "/") // hosted machine id is resourcegroupname/vmname
						if !strings.EqualFold(machineFromPlan.ResourceGroupName.ValueString(), hostedMachineIdArray[0]) {
							machineFromPlan.ResourceGroupName = types.StringValue(hostedMachineIdArray[0])
						}
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), hostedMachineIdArray[1]) {
							machineFromPlan.MachineName = types.StringValue(hostedMachineIdArray[1])
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
					if hostedMachineId != "" {
						machineIdArray := strings.Split(hostedMachineId, ":") // hosted machine id is projectname:region:vmname
						if !strings.EqualFold(machineFromPlan.Region.ValueString(), machineIdArray[1]) {
							machineFromPlan.Region = types.StringValue(machineIdArray[1])
						}
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), machineIdArray[2]) {
							machineFromPlan.MachineName = types.StringValue(machineIdArray[2])
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
					if hostingServerName != "" {
						if !strings.EqualFold(machineFromPlan.Host.ValueString(), hostingServerName) {
							machineFromPlan.Host = types.StringValue(hostingServerName)
						}
					}
					if hostedMachineName != "" {
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), hostedMachineName) {
							machineFromPlan.MachineName = types.StringValue(hostedMachineName)
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
					if hostedMachineName != "" {
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), hostedMachineName) {
							machineFromPlan.MachineName = types.StringValue(hostedMachineName)
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
					if hostedMachineName != "" {
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), hostedMachineName) {
							machineFromPlan.MachineName = types.StringValue(hostedMachineName)
						}
					}
					if hostingServerName != "" {
						host := strings.Split(hostingServerName, ".")[0] // hosting server name is hosting-name.domain.com
						if !strings.EqualFold(machineFromPlan.Host.ValueString(), host) {
							machineFromPlan.Host = types.StringValue(host)
						}
					}
				case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
					if hyp.GetPluginId() == util.NUTANIX_PLUGIN_ID && hostedMachineName != "" {
						if !strings.EqualFold(machineFromPlan.MachineName.ValueString(), hostedMachineName) {
							machineFromPlan.MachineName = types.StringValue(hostedMachineName)
						}
					}
					// case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS: AvailabilityZone is not available from remote
				}

				delete(machineMapFromRemote, strings.ToLower(machineFromPlanName))
			}
		}

		machineAccountsArray := []MachineAccountsModel{}
		machineAccountsModel := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, r.MachineAccounts)
		for _, machineAccount := range machineAccountsModel {
			machineAccountMachines := []MachineCatalogMachineModel{}
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			for _, machine := range machines {
				if machinesNotPresetInRemote[strings.ToLower(machine.MachineAccount.ValueString())] {
					continue
				}
				machine.MachineAccount = types.StringValue(strings.ToLower(machine.MachineAccount.ValueString()))
				machineAccountMachines = append(machineAccountMachines, machine)
			}
			machineAccount.Machines = util.TypedArrayToObjectList[MachineCatalogMachineModel](ctx, diagnostics, machineAccountMachines)
			machineAccountsArray = append(machineAccountsArray, machineAccount)
		}

		r.MachineAccounts = util.TypedArrayToObjectList[MachineAccountsModel](ctx, diagnostics, machineAccountsArray)
	}

	// go over any machines that are in remote but were not in plan
	newMachines := map[string][]MachineCatalogMachineModel{}
	for machineName, machineFromRemote := range machineMapFromRemote {
		hosting := machineFromRemote.GetHosting()
		hypConnection := hosting.GetHypervisorConnection()
		hypId := hypConnection.GetId()

		var machineModel MachineCatalogMachineModel
		machineModel.MachineAccount = types.StringValue(machineName)

		if hypId != "" {
			hyp, err := util.GetHypervisor(ctx, client, nil, hypId)
			if err != nil {
				continue
			}

			connectionType := hyp.GetConnectionType()
			hostedMachineId := hosting.GetHostedMachineId()
			hostingServerName := hosting.GetHostingServerName()
			hostedMachineName := hosting.GetHostedMachineName()

			switch connectionType {
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
				if hostedMachineId != "" {
					hostedMachineIdArray := strings.Split(hostedMachineId, "/") // hosted machine id is resourcegroupname/vmname
					machineModel.ResourceGroupName = types.StringValue(hostedMachineIdArray[0])
					machineModel.MachineName = types.StringValue(hostedMachineIdArray[1])
					// region is not available from remote
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
				if hostedMachineId != "" {
					machineIdArray := strings.Split(hostedMachineId, ":") // hosted machine id is projectname:region:vmname
					machineModel.ProjectName = types.StringValue(machineIdArray[0])
					machineModel.Region = types.StringValue(machineIdArray[1])
					machineModel.MachineName = types.StringValue(machineIdArray[2])
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
				if hostingServerName != "" {
					machineModel.Host = types.StringValue(hostingServerName)
				}
				if hostedMachineName != "" {
					machineModel.MachineName = types.StringValue(hostedMachineName)
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
				if hostedMachineName != "" {
					machineModel.MachineName = types.StringValue(hostedMachineName)
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
				if hostedMachineName != "" {
					machineModel.MachineName = types.StringValue(hostedMachineName)
				}
				if hostingServerName != "" {
					host := strings.Split(hostingServerName, ".")[0] // hosting server name is hosting-name.domain.com
					machineModel.Host = types.StringValue(host)
				}
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
				if hyp.GetPluginId() == util.NUTANIX_PLUGIN_ID && hostedMachineName != "" {
					machineModel.MachineName = types.StringValue(hostedMachineName)
				}
				// case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS: AvailabilityZone is not available from remote
			}
		}

		_, exists := newMachines[hypId]
		if !exists {
			newMachines[hypId] = []MachineCatalogMachineModel{}
		}

		newMachines[hypId] = append(newMachines[hypId], machineModel)
	}

	if len(newMachines) > 0 && r.MachineAccounts.IsNull() {
		r.MachineAccounts = util.TypedArrayToObjectList[MachineAccountsModel](ctx, diagnostics, []MachineAccountsModel{})
	}

	machineAccountMap := map[string]int{}
	machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, r.MachineAccounts)
	for index, machineAccount := range machineAccounts {
		machineAccountMap[machineAccount.Hypervisor.ValueString()] = index
	}

	for hypId, machines := range newMachines {
		machineAccIndex, exists := machineAccountMap[hypId]
		if exists {
			machAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, r.MachineAccounts)
			machineAccount := machAccounts[machineAccIndex]
			if machineAccount.Machines.IsNull() {
				machineAccount.Machines = util.TypedArrayToObjectList[MachineCatalogMachineModel](ctx, diagnostics, []MachineCatalogMachineModel{})
			}
			machineAccountMachines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			machineAccountMachines = append(machineAccountMachines, machines...)
			machineAccount.Machines = util.TypedArrayToObjectList[MachineCatalogMachineModel](ctx, diagnostics, machineAccountMachines)
			machAccounts[machineAccIndex] = machineAccount
			r.MachineAccounts = util.TypedArrayToObjectList[MachineAccountsModel](ctx, diagnostics, machAccounts)
			continue
		}
		var machineAccount MachineAccountsModel
		machineAccount.Hypervisor = types.StringValue(hypId)
		machineAccount.Machines = util.TypedArrayToObjectList[MachineCatalogMachineModel](ctx, diagnostics, machines)
		machAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, r.MachineAccounts)
		machAccounts = append(machAccounts, machineAccount)
		machineAccountMap[hypId] = len(machAccounts) - 1
		r.MachineAccounts = util.TypedArrayToObjectList[MachineAccountsModel](ctx, diagnostics, machAccounts)
	}

	return r
}
