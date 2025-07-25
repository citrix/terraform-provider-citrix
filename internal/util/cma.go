// Copyright © 2024. Citrix Systems, Inc.
package util

import (
	"context"
	"fmt"
	"net/http"
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

// <summary>
// Helper function to Get Citrix Template Image with ID
// </summary>
// <param name="ctx">Context from caller</param>
// <param name="client">Citrix DaaS client from provider context</param>
// <param name="diagnostics">Terraform diagnostics from context</param>
// <param name="imageId">Image ID</param>
// <param name="addWarningIfNotFound">Write warning message to diagnostics if image cannot be found</param>
// <returns>TemplateImageDetails object of the queried template image</returns>
func GetTemplateImageWithId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, imageId string, addWarningIfNotFound bool) (*citrixquickdeploy.TemplateImageDetails, *http.Response, error) {
	getImageRequest := client.QuickDeployClient.MasterImageCMD.GetTemplateImage(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId, imageId)
	image, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickdeploy.TemplateImageDetails](getImageRequest, client)

	if err != nil {
		if addWarningIfNotFound && httpResp.StatusCode == http.StatusNotFound {
			diagnostics.AddWarning(
				fmt.Sprintf("Template Image with ID: %s not found", imageId),
				fmt.Sprintf("Template Image with ID: %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", imageId),
			)
			return nil, httpResp, err
		}
		diagnostics.AddError(
			"Error getting Template Image: "+imageId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadCatalogServiceClientError(err),
		)
		return nil, httpResp, err
	}

	return image, httpResp, nil
}

// <summary>
// Helper function to Get Citrix Managed Subscription with name
// </summary>
// <param name="ctx">Context from caller</param>
// <param name="client">Citrix DaaS client from provider context</param>
// <param name="diagnostics">Terraform diagnostics from context</param>
// <param name="subscriptionName">Name of the Citrix Managed Azure subscription</param>
// <returns>AzureSubscriptionOverview object of the queried Citrix Managed subscription</returns>
func GetCitrixManagedSubscriptionWithName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, subscriptionName string) *citrixquickdeploy.AzureSubscriptionOverview {
	getSubscriptionReq := client.QuickDeployClient.AzureSubscriptionsCMD.GetSubscriptions(ctx, client.ClientConfig.CustomerId, client.ClientConfig.SiteId)
	subscriptionResp, httpResp, err := citrixdaasclient.AddRequestData(getSubscriptionReq, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error getting Citrix Managed Azure subscription: "+subscriptionName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadCatalogServiceClientError(err),
		)
		return nil
	}

	for _, subscription := range subscriptionResp.GetSubscriptions() {
		if strings.EqualFold(subscription.GetName(), subscriptionName) {
			return &subscription
		}
	}

	diagnostics.AddError(
		"Error getting Citrix Managed Azure subscription: "+subscriptionName,
		"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
			"\nError message: Subscription not found",
	)

	return nil
}
