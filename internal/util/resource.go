// Copyright © 2023. Citrix Systems, Inc.

package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const NUTANIX_PLUGIN_ID string = "AcropolisFactory"

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
			"Error reading ResourcePool for Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return resourcePool, err
}

func GetMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogId string, addErrorToDiagnostics bool) (*citrixorchestration.MachineCatalogDetailResponseModel, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogId).Fields("Id,Name,HypervisorConnection,ProvisioningScheme,RemotePCEnrollmentScopes")
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

func GetMachineCatalogMachines(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineCatalogId string) (*citrixorchestration.MachineResponseModelCollection, error) {
	getMachineCatalogMachinesRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalogMachines(ctx, machineCatalogId).Fields("Id,Name,Hosting,DeliveryGroup")
	machines, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineResponseModelCollection](getMachineCatalogMachinesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading Machines for Machine Catalog "+machineCatalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
	}

	return machines, err
}

func GetSingleResourcePathFromHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string) (string, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolResources(ctx, hypervisorName, hypervisorPoolName)
	req = req.Children(1)

	if folderPath != "" {
		req = req.Path(folderPath)
	}

	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	resources, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return "", err
	}

	for _, child := range resources.Children {
		if child.GetName() == resourceName {
			if resourceType == "VirtualPrivateCloud" {
				// For vnet, ID is resourceGroup/vnetName. Match the resourceGroup as well
				resourceGroupAndVnetName := strings.Split(child.GetId(), "/")
				if resourceGroupAndVnetName[0] == resourceGroupName {
					return child.GetXDPath(), nil
				} else {
					continue
				}
			}

			return child.GetXDPath(), nil
		}
	}

	return "", fmt.Errorf("could not find resource")
}

func GetSingleHypervisorResource(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName string, hypervisor *citrixorchestration.HypervisorDetailResponseModel) (*citrixorchestration.HypervisorResourceResponseModel, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	if folderPath != "" {
		req = req.Path(folderPath)
	}
	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	resources, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return nil, err
	}

	for _, child := range resources.Children {
		switch hypervisor.GetConnectionType() {
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
			if (resourceType == "Vm" || resourceType == "VirtualPrivateCloud") && strings.EqualFold(child.GetName(), resourceName) {
				resourceGroupAndVmName := strings.Split(child.GetId(), "/")
				if strings.EqualFold(resourceGroupAndVmName[0], resourceGroupName) {
					return &child, nil
				}
			}

			if resourceType == "Region" && (strings.EqualFold(child.GetName(), resourceName) || strings.EqualFold(child.GetId(), resourceName)) { // support both Azure region name or id ("East US" and "eastus")
				return &child, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
			if resourceType == "Vm" && strings.EqualFold(strings.Split(child.GetName(), " ")[0], resourceName) {
				return &child, nil
			}
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_XEN_SERVER:
			if strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil
			}
		case citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM:
			if hypervisor.GetPluginId() == NUTANIX_PLUGIN_ID && strings.EqualFold(child.GetName(), resourceName) {
				return &child, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find resource")
}

func GetAllResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorId, folderPath, resourceType string) []string {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	req = req.Path(folderPath)
	req = req.Type_([]string{resourceType})

	resources, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}
	}

	result := []string{}
	for _, child := range resources.Children {
		result = append(result, child.GetXDPath())
	}

	return result
}

func GetFilteredResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorId, folderPath, resourceType string, filter []string, connectionType citrixorchestration.HypervisorConnectionType) ([]string, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResources(ctx, hypervisorId)
	req = req.Children(1)
	req = req.Path(folderPath)
	req = req.Type_([]string{resourceType})

	resources, _, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	if filter != nil {
		for _, child := range resources.Children {
			name := child.GetName()
			if connectionType == citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS {
				name = strings.Split(name, " ")[0]
			}
			if Contains(filter, name) {
				result = append(result, child.GetXDPath())
			}
		}
	} else {
		//when the filter is empty
		for _, child := range resources.Children {
			result = append(result, child.GetXDPath())
		}
	}

	return result, nil
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
