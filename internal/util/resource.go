package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
)

func GetSingleResourcePathFromHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorName, hypervisorPoolName, folderPath, resourceName, resourceType, resourceGroupName string) string {
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
		return ""
	}

	for _, child := range resources.Children {
		if child.GetName() == resourceName {
			if resourceType == "VirtualPrivateCloud" {
				// For vnet, ID is resourceGroup/vnetName. Match the resourceGroup as well
				resourceGroupAndVnetName := strings.Split(child.GetId(), "/")
				if resourceGroupAndVnetName[0] == resourceGroupName {
					return child.GetXDPath()
				} else {
					continue
				}
			}

			return child.GetXDPath()
		}
	}

	return ""
}

func GetSingleResourcePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorId, folderPath, resourceName, resourceType, resourceGroupName string) (string, error) {
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
		return "", err
	}

	for _, child := range resources.Children {
		if strings.EqualFold(child.GetName(), resourceName) ||
			(child.GetResourceType() == "Region" && strings.EqualFold(child.GetId(), resourceName)) { // support both Azure region name or id ("East US" and "eastus")
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

	return "", fmt.Errorf("Could not find resource.")
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
