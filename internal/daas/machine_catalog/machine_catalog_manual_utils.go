// Copyright © 2023. Citrix Systems, Inc.

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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getMachinesForManualCatalogs(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machineAccounts []MachineAccountsModel) ([]citrixorchestration.AddMachineToMachineCatalogRequestModel, error) {
	if machineAccounts == nil {
		return nil, nil
	}

	addMachineRequestList := []citrixorchestration.AddMachineToMachineCatalogRequestModel{}
	for _, machineAccount := range machineAccounts {
		hypervisorId := machineAccount.Hypervisor.ValueString()
		var hypervisor *citrixorchestration.HypervisorDetailResponseModel
		var err error
		if hypervisorId != "" {
			hypervisor, err = util.GetHypervisor(ctx, client, nil, hypervisorId)

			if err != nil {
				return nil, err
			}
		}

		for _, machine := range machineAccount.Machines {
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
					return nil, fmt.Errorf("region and resource_group_name are required for Azure")
				}
				region, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, "", machine.Region.ValueString(), "Region", "", hypervisor)
				if err != nil {
					return nil, err
				}
				regionPath := region.GetXDPath()
				vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, fmt.Sprintf("%s\\vm.folder", regionPath), machineName, util.VirtualMachineResourceType, machine.ResourceGroupName.ValueString(), hypervisor)
				if err != nil {
					return nil, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
				if machine.AvailabilityZone.IsNull() {
					return nil, fmt.Errorf("availability_zone is required for AWS")
				}
				availabilityZone, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, "", machine.AvailabilityZone.ValueString(), "", "", hypervisor)
				if err != nil {
					return nil, err
				}
				availabilityZonePath := availabilityZone.GetXDPath()
				vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, availabilityZonePath, machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
				if machine.Region.IsNull() || machine.ProjectName.IsNull() {
					return nil, fmt.Errorf("region and project_name are required for GCP")
				}
				projectName, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, "", machine.ProjectName.ValueString(), "", "", hypervisor)
				if err != nil {
					return nil, err
				}
				projectNamePath := projectName.GetXDPath()
				vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, fmt.Sprintf("%s\\%s.region", projectNamePath, machine.Region.ValueString()), machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
				if machine.Datacenter.IsNull() || machine.Host.IsNull() {
					return nil, fmt.Errorf("datacenter and host are required for Vsphere")
				}

				folderPath := hypervisor.GetXDPath()
				datacenter, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, folderPath, machine.Datacenter.ValueString(), "datacenter", "", hypervisor)
				if err != nil {
					return nil, err
				}

				folderPath = datacenter.GetXDPath()

				if !machine.Cluster.IsNull() {
					cluster, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, folderPath, machine.Cluster.ValueString(), "cluster", "", hypervisor)
					if err != nil {
						return nil, err
					}
					folderPath = cluster.GetXDPath()
				}

				host, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, folderPath, machine.Host.ValueString(), "computeresource", "", hypervisor)
				if err != nil {
					return nil, err
				}
				hostPath := host.GetXDPath()
				vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, hostPath, machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
				vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, "", machineName, util.VirtualMachineResourceType, "", hypervisor)
				if err != nil {
					return nil, err
				}
				vmId = vm.GetId()
			case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
				if hypervisor.GetPluginId() == util.NUTANIX_PLUGIN_ID {
					hypervisorXdPath := hypervisor.GetXDPath()
					vm, err := util.GetSingleHypervisorResource(ctx, client, hypervisorId, fmt.Sprintf("%s\\VirtualMachines.folder", hypervisorXdPath), machineName, util.VirtualMachineResourceType, "", hypervisor)
					if err != nil {
						return nil, err
					}
					vmId = vm.GetId()
				}
			}

			addMachineRequest.SetHostedMachineId(vmId)
			addMachineRequest.SetHypervisorConnection(hypervisorId)

			addMachineRequestList = append(addMachineRequestList, addMachineRequest)
		}
	}

	return addMachineRequestList, nil
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
	for _, machine := range getMachinesResponse.Items {
		if deleteMachinesList[strings.ToLower(machine.GetName())] {
			machinesToDelete = append(machinesToDelete, machine)
		}
	}

	return deleteMachinesFromCatalog(ctx, client, resp, MachineCatalogResourceModel{}, machinesToDelete, catalogNameOrId, false)
}

func addMachinesToManualCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, addMachinesList []MachineAccountsModel, catalogIdOrName string) error {

	if len(addMachinesList) < 1 {
		// no machines to add
		return nil
	}

	addMachinesRequest, err := getMachinesForManualCatalogs(ctx, client, addMachinesList)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding machines(s) to Machine Catalog "+catalogIdOrName,
			fmt.Sprintf("Failed to resolve machines, error: %s", err.Error()),
		)

		return err
	}

	batchApiHeaders, httpResp, err := generateBatchApiHeaders(client, MachineCatalogResourceModel{}, false)
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
	relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/Machines?async=true", catalogIdOrName)
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

func createAddAndRemoveMachinesListForManualCatalogs(state, plan MachineCatalogResourceModel) ([]MachineAccountsModel, map[string]bool) {
	addMachinesList := []MachineAccountsModel{}
	existingMachineAccounts := map[string]map[string]bool{}

	// create map for existing machines marking all machines for deletion
	if state.MachineAccounts != nil {
		for _, machineAccount := range state.MachineAccounts {
			for _, machine := range machineAccount.Machines {
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
	if plan.MachineAccounts != nil {
		for _, machineAccount := range plan.MachineAccounts {
			machineAccountMachines := []MachineCatalogMachineModel{}
			for _, machine := range machineAccount.Machines {
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
				addMachineAccount.Machines = machineAccountMachines
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

func (r MachineCatalogResourceModel) updateCatalogWithMachines(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machines *citrixorchestration.MachineResponseModelCollection) MachineCatalogResourceModel {
	if machines == nil {
		r.MachineAccounts = nil
		return r
	}

	machineMapFromRemote := map[string]citrixorchestration.MachineResponseModel{}
	for _, machine := range machines.GetItems() {
		machineMapFromRemote[strings.ToLower(machine.GetName())] = machine
	}

	if r.MachineAccounts != nil {
		machinesNotPresetInRemote := map[string]bool{}
		for _, machineAccount := range r.MachineAccounts {
			for _, machineFromPlan := range machineAccount.Machines {
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

		machineAccounts := []MachineAccountsModel{}
		for _, machineAccount := range r.MachineAccounts {
			machines := []MachineCatalogMachineModel{}
			for _, machine := range machineAccount.Machines {
				if machinesNotPresetInRemote[strings.ToLower(machine.MachineAccount.ValueString())] {
					continue
				}
				machines = append(machines, machine)
			}
			machineAccount.Machines = machines
			machineAccounts = append(machineAccounts, machineAccount)
		}

		r.MachineAccounts = machineAccounts
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

	if len(newMachines) > 0 && r.MachineAccounts == nil {
		r.MachineAccounts = []MachineAccountsModel{}
	}

	machineAccountMap := map[string]int{}
	for index, machineAccount := range r.MachineAccounts {
		machineAccountMap[machineAccount.Hypervisor.ValueString()] = index
	}

	for hypId, machines := range newMachines {
		machineAccIndex, exists := machineAccountMap[hypId]
		if exists {
			machAccounts := r.MachineAccounts
			machineAccount := machAccounts[machineAccIndex]
			if machineAccount.Machines == nil {
				machineAccount.Machines = []MachineCatalogMachineModel{}
			}
			machineAccountMachines := machineAccount.Machines
			machineAccountMachines = append(machineAccountMachines, machines...)
			machineAccount.Machines = machineAccountMachines
			machAccounts[machineAccIndex] = machineAccount
			r.MachineAccounts = machAccounts
			continue
		}
		var machineAccount MachineAccountsModel
		machineAccount.Hypervisor = types.StringValue(hypId)
		machineAccount.Machines = machines
		machAccounts := r.MachineAccounts
		machAccounts = append(machAccounts, machineAccount)
		machineAccountMap[hypId] = len(machAccounts) - 1
		r.MachineAccounts = machAccounts
	}

	return r
}
