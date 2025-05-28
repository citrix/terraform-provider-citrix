// Copyright © 2024. Citrix Systems, Inc.
package util

import (
	"context"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// <summary>
// Helper function to find Citrix Managed Azure region
// </summary>
// <param name="ctx">Context from caller</param>
// <param name="client">Citrix DaaS client from provider context</param>
// <param name="diagnostics">Terraform diagnostics from context</param>
// <param name="region">Region name</param>
// <returns>DeploymentRegionModel object of the queried region</returns>
func GetCmaRegion(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, regionValue string) *citrixquickdeploy.DeploymentRegionModel {
	regionsRequest := client.QuickDeployClient.ManagedCapacityCMD.GetDeploymentRegions(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId)
	regions, httpResp, err := citrixdaasclient.AddRequestData(regionsRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error getting Citrix Managed Azure regions",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadCatalogServiceClientError(err),
		)
		return nil
	}
	regionFound := false
	for _, region := range regions.GetItems() {
		if strings.EqualFold(region.GetId(), regionValue) || strings.EqualFold(region.GetName(), regionValue) {
			return &region
		}
	}
	if !regionFound {
		supportedRegions := make([]string, 0)
		for _, region := range regions.GetItems() {
			supportedRegions = append(supportedRegions, region.GetName())
		}
		diagnostics.AddError(
			"Error validating Template Image configuration",
			"Region "+regionValue+" is not a supported Citrix Managed Azure regions"+
				"\nRegion should be one of the following: "+strings.Join(supportedRegions, ", ")+
				"\nRegion format should be either region name (East US) or region ID (eastus)",
		)
	}

	return nil
}
