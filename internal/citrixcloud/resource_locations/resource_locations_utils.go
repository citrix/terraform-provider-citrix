// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	"context"
	"fmt"

	resourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetResourceLocation(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, resourceLocationId string) (*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel, error) {
	// Get resource location
	getResourceLocationRequest := client.ResourceLocationsClient.LocationsDAAS.LocationsGet(ctx, resourceLocationId)
	resourceLocation, httpResp, err := citrixdaasclient.ExecuteWithRetry[*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel](getResourceLocationRequest, client)
	if httpResp.StatusCode == 403 {
		diagnostics.AddError(
			"Error reading resource location with id: "+resourceLocationId,
			"Terraform user does not have the Citrix Cloud Resource Location permission. This is required to manage DaaS Zones.",
		)
		return nil, err
	}
	if httpResp.StatusCode == 404 || resourceLocation == nil {
		diagnostics.AddError(
			"Error reading resource location with id: "+resourceLocationId,
			"Resource Location "+resourceLocationId+" not found. Ensure the resource location has been created manually or via terraform, then try again.",
		)
		if err == nil {
			err = fmt.Errorf("resource Location %s not found", resourceLocationId)
		}
		return nil, err
	}
	if err != nil {
		diagnostics.AddError(
			"Error reading resource location with id: "+resourceLocationId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return resourceLocation, err
}
