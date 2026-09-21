// Copyright © 2026. Citrix Systems, Inc.

package resource_locations

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	ccconnectors "github.com/citrix/citrix-daas-rest-go/ccconnectors"
	resourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetResourceLocation(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, resourceLocationId string) (*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel, error) {
	// Get resource location
	getResourceLocationRequest := client.ResourceLocationsClient.LocationsDAAS.LocationsGet(ctx, resourceLocationId)
	resourceLocation, httpResp, err := citrixdaasclient.ExecuteWithRetry[*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel](getResourceLocationRequest, client)
	if httpResp.StatusCode == http.StatusForbidden {
		diagnostics.AddError(
			"Error reading resource location with id: "+resourceLocationId,
			"Terraform user does not have the Citrix Cloud Resource Location permission. This is required to manage DaaS Zones.",
		)
		return nil, err
	}
	if httpResp.StatusCode == http.StatusNotFound || resourceLocation == nil {
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

// getConnectorsInResourceLocation returns the Connectors that report the given resource location. The API is
// filtered by location, and the results are filtered again client side so a broader response cannot be misread.
func getConnectorsInResourceLocation(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, resourceLocationId string) ([]ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel, error) {
	getConnectorsRequest := client.ConnectorsClient.ConnectorsDAAS.ConnectorsGetAll(ctx).Location(resourceLocationId)
	connectors, httpResp, err := citrixdaasclient.ExecuteWithRetry[[]ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel](getConnectorsRequest, client)
	if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
		diagnostics.AddError(
			"Error reading Connectors for resource location with id: "+resourceLocationId,
			"Terraform user does not have the Citrix Cloud Connector permission. This is required to delete a Citrix Cloud Resource Location, so that Connectors are not left orphaned.",
		)
		if err == nil {
			err = fmt.Errorf("terraform user cannot read Connectors for resource location %s", resourceLocationId)
		}
		return nil, err
	}
	if err != nil {
		diagnostics.AddError(
			"Error reading Connectors for resource location with id: "+resourceLocationId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	connectorsInLocation := []ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel{}
	for _, connector := range connectors {
		if strings.EqualFold(connector.GetLocation(), resourceLocationId) {
			connectorsInLocation = append(connectorsInLocation, connector)
		}
	}

	return connectorsInLocation, nil
}

// deleteConnector removes a single Connector from Citrix Cloud so its resource location can be deleted.
func deleteConnector(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, connector ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel) error {
	deleteConnectorRequest := client.ConnectorsClient.ConnectorsDAAS.ConnectorsDelete(ctx, connector.GetId())
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[bool](deleteConnectorRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error removing Connector "+connector.GetFqdn()+" from Citrix Cloud",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	return nil
}

// connectorsBlockingDeleteError describes why a resource location cannot be deleted while it still contains
// Connectors. Shared so the plan time and delete time checks cannot drift apart.
func connectorsBlockingDeleteError(resourceLocationId string, connectors []ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel) (string, string) {
	return "Cannot delete resource location with id: " + resourceLocationId,
		"Resource location still contains the following Connectors: " + getConnectorFqdns(connectors) +
			".\nRemove these Connectors, or set `force_delete` to `true` and apply before deleting to have the provider remove them from Citrix Cloud."
}

// getConnectorFqdns joins the fqdns of the given Connectors for use in diagnostics.
func getConnectorFqdns(connectors []ccconnectors.CitrixCloudServicesAgentHubApiEdgeServersGetResultModel) string {
	fqdns := []string{}
	for _, connector := range connectors {
		fqdns = append(fqdns, connector.GetFqdn())
	}

	return strings.Join(fqdns, ", ")
}
