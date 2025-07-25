// Copyright © 2024. Citrix Systems, Inc.

package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Plugin Factory Names
const AZURERM_FACTORY_NAME string = "AzureRmFactory"
const AMAZON_WORKSPACES_CORE_FACTORY_NAME string = "AmazonWorkSpacesCoreMachineManagerFactory"
const VMWARE_FACTORY_NAME string = "VmwareFactory"
const NUTANIX_PLUGIN_ID string = "AcropolisFactory"
const OPENSHIFT_PLUGIN_ID string = "OpenShiftPluginFactory"
const HPE_MOONSHOT_PLUGIN_ID = "HPMoonshotFactory"
const REMOTE_PC_WAKE_ON_LAN_PLUGIN_ID string = "VdaWOLMachineManagerFactory"

// Gets the hypervisor and logs any errors
func GetHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId string) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	// Resolve resource path for service offering and master image
	getHypervisorReq := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorId)
	hypervisor, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.HypervisorDetailResponseModel](getHypervisorReq, client)
	if err != nil && diagnostics != nil {
		diagnostics.AddError(
			"Error reading Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return hypervisor, err
}

func GetHypervisorResourcePool(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, hypervisorResourcePoolId string) (*citrixorchestration.HypervisorResourcePoolDetailResponseModel, error) {
	getResourcePoolsRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePool(ctx, hypervisorId, hypervisorResourcePoolId)
	resourcePool, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.HypervisorResourcePoolDetailResponseModel](getResourcePoolsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading ResourcePool "+hypervisorResourcePoolId+" for Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return resourcePool, err
}

func GetMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogId string, addErrorToDiagnostics bool) (*citrixorchestration.MachineCatalogDetailResponseModel, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogId).Fields("Id,Name,Description,ProvisioningType,PersistChanges,Zone,AllocationType,SessionSupport,TotalCount,HypervisorConnection,ProvisioningScheme,RemotePCEnrollmentScopes,IsPowerManaged,MinimumFunctionalLevel,IsRemotePC,Metadata,Scopes,UpgradeInfo,AdminFolder")
	catalog, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, client)
	if err != nil && addErrorToDiagnostics {
		diagnostics.AddError(
			"Error reading Machine Catalog "+machineCatalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return catalog, err
}

func GetMachineCatalogIdWithPath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogPath string) (string, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogPath).Fields("Id")
	catalog, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Machine Catalog "+machineCatalogPath,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return catalog.GetId(), err
}

func GetDeliveryGroups(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, fields string) ([]citrixorchestration.DeliveryGroupResponseModel, error) {
	req := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroups(ctx)
	req = req.Limit(250)
	if fields != "" {
		req = req.Fields(fields)
	}

	deliveryGroups := []citrixorchestration.DeliveryGroupResponseModel{}
	continuationToken := ""
	for {
		req = req.ContinuationToken(continuationToken)

		responseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupResponseModelCollection](req, client)
		if err != nil {
			diagnostics.AddError(
				"Error reading delivery groups",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadClientError(err),
			)
			return deliveryGroups, err
		}
		deliveryGroups = append(deliveryGroups, responseModel.GetItems()...)

		if responseModel.GetContinuationToken() == "" {
			return deliveryGroups, nil
		}
		continuationToken = responseModel.GetContinuationToken()
	}
}

func GetDeliveryGroup(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) (*citrixorchestration.DeliveryGroupDetailResponseModel, error) {
	getDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupId)
	deliveryGroup, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Delivery Group "+deliveryGroupId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return deliveryGroup, err
}

func GetDeliveryGroupIdWithPath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupPath string) (string, error) {
	getDeliveryGroupRequest := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroup(ctx, deliveryGroupPath).Fields("Id")
	deliveryGroup, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.DeliveryGroupDetailResponseModel](getDeliveryGroupRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Delivery Group "+deliveryGroupPath,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return deliveryGroup.GetId(), err
}

func GetDeliveryGroupMachines(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, deliveryGroupId string) ([]citrixorchestration.MachineResponseModel, error) {
	req := client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupMachines(ctx, deliveryGroupId)
	req = req.Limit(250)

	responses := []citrixorchestration.MachineResponseModel{}
	continuationToken := ""
	for {
		req = req.ContinuationToken(continuationToken)
		responseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineResponseModelCollection](req, client)
		if err != nil {
			diagnostics.AddError(
				"Error reading Machines for Delivery Group "+deliveryGroupId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadClientError(err),
			)
			return responses, err
		}
		responses = append(responses, responseModel.GetItems()...)
		if responseModel.GetContinuationToken() == "" {
			return responses, nil
		}
		continuationToken = responseModel.GetContinuationToken()
	}
}

func GetApplicationGroupIdWithPath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, appGroupPath string) (string, error) {
	getAppGroupRequest := client.ApiClient.ApplicationGroupsAPIsDAAS.ApplicationGroupsGetApplicationGroup(ctx, appGroupPath).Fields("Id")
	appGroup, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.ApplicationGroupDetailResponseModel](getAppGroupRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Application Group "+appGroupPath,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return appGroup.GetId(), err
}

func GetMachineCatalogMachines(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogId string) ([]citrixorchestration.MachineResponseModel, error) {
	req := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalogMachines(ctx, machineCatalogId).Fields("Id,Name,Hosting,DeliveryGroup,InMaintenanceMode,AssignedUsers,AssociatedUsers,AllocationType,Sid")
	req = req.Limit(250)

	responses := []citrixorchestration.MachineResponseModel{}
	continuationToken := ""
	for {
		req = req.ContinuationToken(continuationToken)
		responseModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineResponseModelCollection](req, client)
		if err != nil {
			diagnostics.AddError(
				"Error reading Machines for Machine Catalog "+machineCatalogId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadClientError(err),
			)
			return responses, err
		}
		responses = append(responses, responseModel.GetItems()...)
		if responseModel.GetContinuationToken() == "" {
			return responses, nil
		}
		continuationToken = responseModel.GetContinuationToken()
	}
}

func GetSingleResourceFromHypervisorWithNoCacheRetry(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string) (*citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	resource, httpResp, err := getSingleResourceFromHypervisor(ctx, client, diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName, false, false)
	if err != nil {
		resource, httpResp, err = getSingleResourceFromHypervisor(ctx, client, diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName, true, true)
		if err != nil {
			return nil, httpResp, err
		}
	}
	return resource, httpResp, nil
}

func GetAllChildrenForResourcePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, hypervisorPoolName, folderPath string, resourceType string, addToDiagnostics bool, noCache bool) ([]citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolResources(ctx, hypervisorName, hypervisorPoolName)
	req = req.Children(1)
	req = req.NoCache(noCache)

	if folderPath != "" {
		req = req.Path(folderPath)
	}

	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	req = req.Async(true)

	_, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return nil, httpResp, err
	}

	resources, err := GetAsyncJobResultWithAddToDiagsOption[citrixorchestration.HypervisorResourceResponseModel](ctx, client, httpResp, "Error getting Hypervisor resources", diagnostics, 5, addToDiagnostics)
	if errors.Is(err, &JobPollError{}) {
		return nil, httpResp, err
	}

	return resources.Children, httpResp, nil
}

func getSingleResourceFromHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string, addToDiagnostics bool, noCache bool) (*citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolResources(ctx, hypervisorName, hypervisorPoolName)
	req = req.Children(1)
	req = req.NoCache(noCache)

	if folderPath != "" {
		req = req.Path(folderPath)
	}

	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	req = req.Async(true)

	_, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return nil, httpResp, err
	}

	resources, err := GetAsyncJobResultWithAddToDiagsOption[citrixorchestration.HypervisorResourceResponseModel](ctx, client, httpResp, "Error getting Hypervisor resources", diagnostics, 5, addToDiagnostics)
	if errors.Is(err, &JobPollError{}) {
		return nil, httpResp, err
	} // if the job failed continue processing

	for _, child := range resources.Children {
		if strings.EqualFold(child.GetName(), resourceName) {
			if strings.EqualFold(resourceType, VirtualPrivateCloudResourceType) {
				// For vnet, ID is resourceGroup/vnetName. Match the resourceGroup as wells
				resourceGroupAndVnetName := strings.Split(child.GetId(), "/")
				if strings.EqualFold(resourceGroupAndVnetName[0], resourceGroupName) {
					return &child, nil, nil
				} else {
					continue
				}
			}

			if strings.Contains(folderPath, "diskencryptionset") {
				// For diskencryptionset, The ID is the Azure Resource Id of the format subscriptions/{subId}/resourceGroups/{rgName}/providers/.../
				desIdArray := strings.Split(child.GetId(), "/")
				resourceGroupsIndex := slices.Index(desIdArray, "resourceGroups")
				rgName := desIdArray[resourceGroupsIndex+1]

				if strings.EqualFold(rgName, resourceGroupName) {
					return &child, nil, nil
				}
			}
			return &child, nil, nil
		} else if strings.EqualFold(child.GetId(), resourceName) && strings.EqualFold(resourceType, ServiceOfferingResourceType) {
			return &child, nil, nil
		}
	}

	return nil, httpResp, fmt.Errorf("could not find resource. Resource name: %s, resource type: %s", resourceName, resourceType)
}

func GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string) (string, *http.Response, error) {
	path, httpResp, err := getSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName, false, false)
	if err != nil {
		path, httpResp, err = getSingleResourcePathFromHypervisor(ctx, client, diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName, true, true)
		if err != nil {
			return "", httpResp, err
		}
	}
	return path, httpResp, nil
}

func getSingleResourcePathFromHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string, addToDiagnostics bool, noCache bool) (string, *http.Response, error) {
	resource, httpResp, err := getSingleResourceFromHypervisor(ctx, client, diagnostics, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName, addToDiagnostics, noCache)
	if err != nil {
		return "", httpResp, err
	}

	if strings.EqualFold(resourceType, StorageResourceType) {
		return resource.GetId(), nil, nil
	}

	return resource.GetXDPath(), nil, nil
}

func GetSingleHypervisorResourceWithNoCacheRetry(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName string, hypervisor *citrixorchestration.HypervisorDetailResponseModel) (*citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	resource, httpResp, err := getSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName, hypervisor, false, false)
	if err != nil {
		resource, httpResp, err = getSingleHypervisorResource(ctx, client, diagnostics, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName, hypervisor, true, true)
		if err != nil {
			return nil, httpResp, err
		}
	}

	return resource, httpResp, nil
}

func getSingleHypervisorResource(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName string, hypervisor *citrixorchestration.HypervisorDetailResponseModel, addToDiagnostics bool, noCache bool) (*citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	req = req.NoCache(noCache)
	if folderPath != "" {
		req = req.Path(folderPath)
	}
	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}
	req = req.Async(true)

	_, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return nil, httpResp, err
	}

	resources, err := GetAsyncJobResultWithAddToDiagsOption[citrixorchestration.HypervisorResourceResponseModel](ctx, client, httpResp, "Error getting Hypervisor resources", diagnostics, 5, addToDiagnostics)
	if errors.Is(err, &JobPollError{}) {
		return nil, httpResp, err
	}

	for _, child := range resources.Children {
		switch hypervisor.GetConnectionType() {
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
			if (strings.EqualFold(resourceType, VirtualMachineResourceType) || strings.EqualFold(resourceType, VirtualPrivateCloudResourceType)) && strings.EqualFold(child.GetName(), resourceName) {
				resourceGroupAndVmName := strings.Split(child.GetId(), "/")
				if strings.EqualFold(resourceGroupAndVmName[0], resourceGroupName) {
					return &child, nil, nil
				}
			}

			if strings.EqualFold(resourceType, RegionResourceType) && (strings.EqualFold(child.GetName(), resourceName) || strings.EqualFold(child.GetId(), resourceName)) { // support both Azure region name or id ("East US" and "eastus")
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
			if strings.EqualFold(resourceType, VirtualMachineResourceType) && (strings.EqualFold(strings.Split(child.GetName(), " ")[0], resourceName) || strings.EqualFold(child.GetId(), resourceName)) {
				return &child, nil, nil
			}
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
			if (hypervisor.GetPluginId() == NUTANIX_PLUGIN_ID || hypervisor.GetPluginId() == HPE_MOONSHOT_PLUGIN_ID) && strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil, nil
			}
		}
	}

	return nil, httpResp, fmt.Errorf("could not find resource. Resource name: %s, resource type: %s", resourceName, resourceType)
}

func GetAllResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath, resourceType string, addToDiagnostics bool, noCache bool) []string {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	req = req.Path(folderPath)
	req = req.Type_([]string{resourceType})
	req = req.Async(true)
	req = req.NoCache(noCache)

	_, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}
	}

	resources, err := GetAsyncJobResultWithAddToDiagsOption[citrixorchestration.HypervisorResourceResponseModel](ctx, client, httpResp, "Error getting Hypervisor resources", diagnostics, 5, addToDiagnostics)
	if errors.Is(err, &JobPollError{}) {
		return []string{}
	}

	result := []string{}
	for _, child := range resources.Children {
		result = append(result, child.GetXDPath())
	}

	return result
}

func GetFilteredResourcePathListWithNoCacheRetry(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath, resourceType string, filter []string, connectionType citrixorchestration.HypervisorConnectionType, pluginId string) ([]string, error) {
	resource, err := getFilteredResourcePathList(ctx, client, diagnostics, hypervisorId, folderPath, resourceType, filter, connectionType, pluginId, false, false)
	if err != nil {
		resource, err = getFilteredResourcePathList(ctx, client, diagnostics, hypervisorId, folderPath, resourceType, filter, connectionType, pluginId, true, true)
		if err != nil {
			return nil, err
		}
	}
	return resource, nil
}

func getFilteredResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath, resourceType string, filter []string, connectionType citrixorchestration.HypervisorConnectionType, pluginId string, addToDiagnostics bool, noCache bool) ([]string, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	req = req.Path(folderPath)
	req = req.NoCache(noCache)
	// Skip resource type filter for on-prem hypervisors to avoid server side filtering timeout
	if connectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM &&
		connectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER &&
		connectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER &&
		connectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_SCVMM &&
		connectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT {
		req = req.Type_([]string{resourceType})
	}

	req = req.Async(true)

	_, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}, err
	}

	resources, err := GetAsyncJobResultWithAddToDiagsOption[citrixorchestration.HypervisorResourceResponseModel](ctx, client, httpResp, "Error getting Hypervisor resources", diagnostics, 5, addToDiagnostics)
	if errors.Is(err, &JobPollError{}) {
		return []string{}, err
	}

	result := []string{}
	if filter != nil {

		filterMap := map[string]bool{}
		for _, f := range filter {
			filterMap[strings.ToLower(f)] = false
		}

		for _, child := range resources.Children {
			if strings.EqualFold(child.ResourceType, resourceType) {
				name := child.GetName()
				if connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS || connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AMAZON_WORK_SPACES_CORE {
					name = strings.Split(name, " ")[0]
				}
				if _, exists := filterMap[strings.ToLower(name)]; exists {
					if (connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM || connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER || connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT || connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AMAZON_WORK_SPACES_CORE) && strings.EqualFold(resourceType, NetworkResourceType) {
						result = append(result, child.GetRelativePath())
					} else if (connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM && strings.EqualFold(pluginId, NUTANIX_PLUGIN_ID) && strings.EqualFold(resourceType, NetworkResourceType)) || (connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT && strings.EqualFold(pluginId, OPENSHIFT_PLUGIN_ID) && strings.EqualFold(resourceType, NamespaceResourceType)) {
						result = append(result, child.GetFullName())
					} else {
						result = append(result, child.GetXDPath())
					}

					filterMap[strings.ToLower(name)] = true
				}
			}
		}

		resourcesNotFound := []string{}
		for f, found := range filterMap {
			if !found {
				resourcesNotFound = append(resourcesNotFound, f)
			}
		}

		if len(resourcesNotFound) > 0 {
			return nil, fmt.Errorf("the following resources were not found: %v", resourcesNotFound)
		}

	} else {
		//when the filter is empty
		for _, child := range resources.Children {
			if strings.EqualFold(child.ResourceType, resourceType) {
				result = append(result, child.GetXDPath())
			}
		}
	}

	return result, nil
}

func ValidateHypervisorResource(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorName string, hypervisorPoolName string, resourcePath string) (bool, string) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsValidateHypervisorResourcePoolResource(ctx, hypervisorName, hypervisorPoolName)
	var validationRequestModel citrixorchestration.HypervisorResourceValidationRequestModel
	validationRequestModel.SetPath(resourcePath)
	req = req.HypervisorResourceValidationRequestModel(validationRequestModel)

	responseModel, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return false, ReadClientError(err)
	}

	reports := responseModel.GetReports()
	index := slices.IndexFunc(reports, func(report citrixorchestration.ResourceValidationReportModel) bool {
		return report.GetCategory() == citrixorchestration.RESOURCEVALIDATIONCATEGORY_MACHINE_PROFILE
	})

	report := reports[index]
	if report.GetResult() == citrixorchestration.RESOURCEVALIDATIONRESULT_FAILED {
		violations := report.GetViolations()
		errIndex := slices.IndexFunc(violations, func(violation citrixorchestration.ResourceValidationViolationModel) bool {
			return violation.GetLevel() == citrixorchestration.RESOURCEVIOLATIONLEVEL_ERROR
		})
		violation := violations[errIndex]
		return false, violation.GetMessage()
	}

	return true, ""
}

func CategorizeScopes(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, scopeResponses []citrixorchestration.ScopeResponseModel, parentObjectType citrixorchestration.ScopedObjectType, parentObjectIds []string, scopeIdsInPlan []string) ([]string, []string, []string, error) {
	regularScopeIds := []string{}
	builtInScopeIds := []string{}
	inheritedScopeIds := []string{}
	for _, scope := range scopeResponses {
		scopeId := scope.GetId()

		if slices.Contains(scopeIdsInPlan, scopeId) {
			regularScopeIds = append(regularScopeIds, scopeId)
			continue
		} else if scope.GetIsBuiltIn() {
			builtInScopeIds = append(builtInScopeIds, scopeId)
			continue
		}
		isInheritedScope, err := IsScopeInherited(ctx, client, diagnostics, scopeId, parentObjectType, parentObjectIds)
		if err != nil {
			return regularScopeIds, builtInScopeIds, inheritedScopeIds, err
		}
		if !isInheritedScope {
			regularScopeIds = append(regularScopeIds, scopeId)
		} else {
			inheritedScopeIds = append(inheritedScopeIds, scopeId)
		}
	}
	return regularScopeIds, builtInScopeIds, inheritedScopeIds, nil
}

func IsScopeInherited(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, scopeNameOrId string, parentObjectType citrixorchestration.ScopedObjectType, parentObjectIds []string) (bool, error) {
	responseModels, err := GetAllScopedObjects(ctx, client, diagnostics, scopeNameOrId, "")
	if err != nil {
		return false, err
	}
	for _, scopedObject := range responseModels {
		if parentObjectType == scopedObject.GetObjectType() {
			object := scopedObject.GetObject()
			objectId := object.GetId()
			// For the ScopedObjects API, the id attribute of Machine Catalog, Delivery Group, and Application Group responses use UID instead of GUID
			if parentObjectType == citrixorchestration.SCOPEDOBJECTTYPE_MACHINE_CATALOG {
				objectId, err = GetMachineCatalogIdWithPath(ctx, client, diagnostics, strings.ReplaceAll(object.GetName(), "\\", "|"))
				if err != nil {
					return false, err
				}
			} else if parentObjectType == citrixorchestration.SCOPEDOBJECTTYPE_DELIVERY_GROUP {
				objectId, err = GetDeliveryGroupIdWithPath(ctx, client, diagnostics, strings.ReplaceAll(object.GetName(), "\\", "|"))
				if err != nil {
					return false, err
				}
			} else if parentObjectType == citrixorchestration.SCOPEDOBJECTTYPE_APPLICATION_GROUP {
				objectId, err = GetApplicationGroupIdWithPath(ctx, client, diagnostics, strings.ReplaceAll(object.GetName(), "\\", "|"))
				if err != nil {
					return false, err
				}
			}
			if slices.Contains(parentObjectIds, objectId) {
				return true, nil
			}
		}
	}

	return false, nil
}

func GetAllScopedObjects(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, scopeNameOrId string, continuationToken string) ([]citrixorchestration.ScopedObjectResponseModel, error) {
	req := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScopedObjects(ctx, scopeNameOrId)
	req = req.Limit(250)
	req = req.ContinuationToken(continuationToken)
	responseModel, httpResp, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetching associated objects for admin scope "+scopeNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
		return []citrixorchestration.ScopedObjectResponseModel{}, err
	}
	if responseModel.GetContinuationToken() != "" {
		childResponse, err := GetAllScopedObjects(ctx, client, diagnostics, scopeNameOrId, responseModel.GetContinuationToken())
		return append(responseModel.GetItems(), childResponse...), err
	} else {
		return responseModel.GetItems(), nil
	}
}

func BuildResourcePathForGetRequest(resourcePathInput string, resourceName string) string {
	if resourcePathInput != "" {
		return strings.ReplaceAll(resourcePathInput, "\\", "|") + "|" + resourceName
	} else {
		return resourceName
	}
}
