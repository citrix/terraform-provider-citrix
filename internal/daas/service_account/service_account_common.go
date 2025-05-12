// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func GetServiceAccountUsingAccountId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string) (*citrixorchestration.ServiceAccountResponseModel, error) {
	getServiceAccountsRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetServiceAccounts(ctx)

	// Get all service accounts
	serviceAccounts, httpResp, err := citrixdaasclient.AddRequestData(getServiceAccountsRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error reading Service Accounts",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	// Find the service account in the collection response
	for _, serviceAccount := range serviceAccounts.GetItems() {
		if strings.EqualFold(serviceAccount.GetAccountId(), accountId) {
			return &serviceAccount, nil
		}
	}

	// If not found, return an error
	diagnostics.AddError(
		"Error reading Service Account",
		"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
			"\nError message: Service Account with accountId "+accountId+" not found",
	)

	return nil, nil
}
