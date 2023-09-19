package util

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
)

func GetSingleResourcePathFromHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, hypervisorName, folderPath, resourceName, resourceType, resourceGroupName string) string {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolResources(ctx, hypervisorName, hypervisorName)
	req = req.Children(1)

	if folderPath != "" {
		req = req.Path(folderPath)
	}

	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	resources, res, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return ""
	}
	defer res.Body.Close()

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

func GetSingleResourcePath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, connectionDetails *citrixorchestration.HypervisorConnectionDetailRequestModel, folderPath, resourceName, resourceType, resourceGroupName string) string {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResourcesWithoutConnection(ctx)
	req = req.Children(1)
	req = req.HypervisorConnectionDetailRequestModel(*connectionDetails)
	if folderPath != "" {
		req = req.Path(folderPath)
	}
	if resourceType != "" {
		req = req.Type_([]string{resourceType})
	}

	resources, res, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return ""
	}
	defer res.Body.Close()

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

func GetAllResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, connectionDetails *citrixorchestration.HypervisorConnectionDetailRequestModel, folderPath, resourceType string) []string {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResourcesWithoutConnection(ctx)
	req = req.Children(1)
	req = req.HypervisorConnectionDetailRequestModel(*connectionDetails)
	req = req.Path(folderPath)
	req = req.Type_([]string{resourceType})

	resources, res, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}
	}
	defer res.Body.Close()

	result := []string{}
	for _, child := range resources.Children {
		result = append(result, child.GetXDPath())
	}

	return result
}

func GetFilteredResourcePathList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, connectionDetails *citrixorchestration.HypervisorConnectionDetailRequestModel, folderPath, resourceType string, filter []string) ([]string, error) {
	req := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorAllResourcesWithoutConnection(ctx)
	req = req.Children(1)
	req = req.HypervisorConnectionDetailRequestModel(*connectionDetails)
	req = req.Path(folderPath)
	req = req.Type_([]string{resourceType})

	token, _ := client.SignIn()
	req = req.Authorization(token)

	resources, res, err := citrixdaasclient.AddRequestData(req, client).Execute()
	if err != nil {
		return []string{}, err
	}
	defer res.Body.Close()

	result := []string{}
	if filter != nil {
		for _, child := range resources.Children {
			if Contains(filter, child.GetName()) {
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
