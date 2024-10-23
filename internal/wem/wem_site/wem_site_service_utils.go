// Copyright © 2024. Citrix Systems, Inc.

package wem_site

import (
	"context"
	"fmt"
	"strconv"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func readConfigurationSet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, wemResource WemSiteResourceModel) (*citrixwemservice.SiteModel, error) {
	idInt64, err := strconv.ParseInt(wemResource.Id.ValueString(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %v", err)
	}
	siteGetRequest := client.WemClient.SiteDAAS.SiteQueryById(ctx, idInt64)
	siteGetResponse, _, err := util.ReadResource[*citrixwemservice.SiteModel](siteGetRequest, ctx, client, resp, "Configuration Set", wemResource.Name.ValueString())
	return siteGetResponse, err
}

func getSiteByName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, wemResource WemSiteResourceModel) (citrixwemservice.SiteModel, error) {
	siteName := wemResource.Name.ValueString()
	siteGetRequest := client.WemClient.SiteDAAS.SiteQuery(ctx)
	siteGetRequest = siteGetRequest.Name(siteName)
	siteGetResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixwemservice.SiteQuery200Response](siteGetRequest, client)

	siteConfigList := siteGetResponse.GetItems()
	var siteConfig citrixwemservice.SiteModel

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return siteConfig, err
	}

	if len(siteConfigList) != 0 {
		siteConfig = siteConfigList[0]
	}
	if siteConfig.Id == nil {
		return siteConfig, fmt.Errorf("site with name %s not found", wemResource.Name.ValueString())
	}
	return siteConfig, nil
}

func getSiteById(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, wemResource WemSiteResourceModel) (*citrixwemservice.SiteModel, error) {
	idInt64, err := strconv.ParseInt(wemResource.Id.ValueString(), 10, 64)
	if err != nil {
		return &citrixwemservice.SiteModel{}, fmt.Errorf("invalid id: %v", err)
	}
	siteGetRequest := client.WemClient.SiteDAAS.SiteQueryById(ctx, idInt64)
	siteGetResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixwemservice.SiteModel](siteGetRequest, client)

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return siteGetResponse, err
	}

	if siteGetResponse == nil {
		return nil, fmt.Errorf("site with name %s not found", wemResource.Name.ValueString())
	}
	return siteGetResponse, nil
}
