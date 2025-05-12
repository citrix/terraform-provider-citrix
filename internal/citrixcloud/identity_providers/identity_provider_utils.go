// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Create Identity Provider Utility Functions
func createIdentityProvider(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpType string, idpNickname string) (*citrixcws.IdpStatusModel, error) {
	var idpCreateModel citrixcws.IdpCreateModel
	idpCreateModel.SetIdentityProviderId(idpType)
	idpCreateModel.SetIdentityProviderNickname(idpNickname)

	createIdpRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersMultiIdentityProvidersPost(ctx, client.ClientConfig.CustomerId)
	createIdpRequest = createIdpRequest.IdpCreateModel(idpCreateModel)

	idpStatus, httpResp, err := citrixdaasclient.AddRequestData(createIdpRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error creating %s Identity Provider %s", idpCreateModel.GetIdentityProviderId(), idpCreateModel.GetIdentityProviderNickname()),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return idpStatus, nil
}

func createIdentityProviderConnection(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpType string, idpInstanceId string, idpConnectBody citrixcws.IdpInstanceConnectModel) (*citrixcws.IdpStatusModel, error) {
	// Configure Identity Provider Post Request
	connectIdpRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersIdPost(ctx, idpInstanceId, client.ClientConfig.CustomerId)
	connectIdpRequest = connectIdpRequest.IdpInstanceConnectModel(idpConnectBody)

	idpStatus, httpResp, err := citrixdaasclient.AddRequestData(connectIdpRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error configuring %s Identity Provider %s", idpType, idpInstanceId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return idpStatus, nil
}

// Read Identity Provider Utility Functions
func readIdentityProvider(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, idpType string, idpInstanceId string) (*citrixcws.IdpStatusModel, error) {
	getIdpsRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersIdpTypeGet(ctx, idpType, client.ClientConfig.CustomerId)
	getIdpsResult, _, err := util.ReadResource[*citrixcws.IdpStatusesModel](getIdpsRequest, ctx, client, resp, fmt.Sprintf("%s Identity Provider", idpType), idpInstanceId)
	if err != nil {
		return nil, err
	}
	return getIdpWithInstanceId(&resp.Diagnostics, getIdpsResult, idpType, idpInstanceId)
}

func getIdentityProvidersWithType(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, idpType string) (*citrixcws.IdpStatusesModel, error) {
	getIdpsRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersIdpTypeGet(ctx, idpType, client.ClientConfig.CustomerId)
	getIdpsResult, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixcws.IdpStatusesModel](getIdpsRequest, client)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error fetching Identity Provider instances with type: %s", idpType),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return getIdpsResult, nil
}

func getIdentityProviderById(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, idpType string, idpInstanceId string) (*citrixcws.IdpStatusModel, error) {
	getIdpsResult, err := getIdentityProvidersWithType(ctx, client, diagnostics, idpType)
	if err != nil {
		return nil, err
	}
	return getIdpWithInstanceId(diagnostics, getIdpsResult, idpType, idpInstanceId)
}

func getIdentityProviderByName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, idpType string, idpInstanceName string) (*citrixcws.IdpStatusModel, error) {
	getIdpsResult, err := getIdentityProvidersWithType(ctx, client, diagnostics, idpType)
	if err != nil {
		return nil, err
	}
	err = fmt.Errorf("no %s Identity Provider found with name: %s", idpType, idpInstanceName)
	if getIdpsResult == nil || len(getIdpsResult.GetItems()) == 0 {
		diagnostics.AddError(
			fmt.Sprintf("Error fetching %s Identity Provider with name: %s", idpType, idpInstanceName),
			"Error message: "+err.Error(),
		)
		return nil, err
	}

	for _, idp := range getIdpsResult.GetItems() {
		if strings.EqualFold(idp.GetIdpNickname(), idpInstanceName) {
			return &idp, nil
		}
	}
	diagnostics.AddError(
		fmt.Sprintf("Error fetching %s Identity Provider with name: %s", idpType, idpInstanceName),
		"Error message: "+err.Error(),
	)
	return nil, err
}

func getIdpWithInstanceId(diagnostics *diag.Diagnostics, idpsResult *citrixcws.IdpStatusesModel, idpType string, idpInstanceId string) (*citrixcws.IdpStatusModel, error) {
	err := fmt.Errorf("no %s Identity Provider found with id: %s", idpType, idpInstanceId)
	if idpsResult == nil || len(idpsResult.GetItems()) == 0 {
		diagnostics.AddError(
			fmt.Sprintf("Error fetching %s Identity Provider with id: %s", idpType, idpInstanceId),
			"Error message: "+err.Error(),
		)
		return nil, err
	}
	for _, idp := range idpsResult.GetItems() {
		if strings.EqualFold(idp.GetIdpInstanceId(), idpInstanceId) {
			return &idp, nil
		}
	}

	diagnostics.AddError(
		fmt.Sprintf("Error fetching %s Identity Provider with id: %s", idpType, idpInstanceId),
		"Error message: "+err.Error(),
	)
	return nil, err
}

// Update Identity Provider Utility Functions
func updateIdentityProviderNickname(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpType string, idpInstanceId string, newNickname string) {
	var idpUpdateBody citrixcws.IdpUpdateModel
	idpUpdateBody.SetNickname(newNickname)

	updateIdpRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersIdentityProviderIdPut(ctx, idpType, idpInstanceId, client.ClientConfig.CustomerId)
	updateIdpRequest = updateIdpRequest.IdpUpdateModel(idpUpdateBody)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateIdpRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error updating name of %s Identity Provider with id: %s", idpType, idpInstanceId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
}

func updateIdentityProviderAuthDomain(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpType string, idpInstanceId string, oldAuthDomainName string, newAuthDomainName string) {
	updateAuthDomainRequest := client.CwsClient.AuthDomainsDAAS.CustomerAuthDomainsPut(ctx, client.ClientConfig.CustomerId)
	updateAuthDomainRequest = updateAuthDomainRequest.OldName(oldAuthDomainName)
	updateAuthDomainRequest = updateAuthDomainRequest.NewName(newAuthDomainName)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateAuthDomainRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error updating auth domain of %s Identity Provider with id: %s", idpType, idpInstanceId),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}
}

// Delete Identity Provider Utility Functions
func deleteIdentityProvider(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, idpType string, idpInstanceId string) {
	deleteIdpRequest := client.CwsClient.IdentityProvidersDAAS.CustomerIdentityProvidersTypeIdDelete(ctx, idpType, idpInstanceId, client.ClientConfig.CustomerId)
	_, httpResp, err := citrixdaasclient.AddRequestData(deleteIdpRequest, client).Execute()
	if err != nil {
		errBody := ""
		if httpResp != nil && httpResp.Body != nil {
			body, _ := io.ReadAll(httpResp.Body)
			defer httpResp.Body.Close()
			errBody = fmt.Sprintf("\nError body: %s", string(body))
		}

		if httpResp.StatusCode == http.StatusBadRequest && strings.Contains(errBody, "it is the currently selected authentication method") {
			diagnostics.AddError(
				fmt.Sprintf("Error deleting %s Identity Provider with id: %s", idpType, idpInstanceId),
				"\n\nError message: "+util.ReadClientError(err)+errBody+
					"\n\n\nCannot remove an Identity Provider that is in use. Follow these steps to select a different Identity Provider:\n"+
					"\n1. Go to Citrix Cloud Console.\n"+
					"\n2. Navigate to Workspace configuration > Authentication.\n"+
					"\n3. The Workspace Authentication will consist of Connected identity providers.\n"+
					"\n4. If the SAML IDP is selected, that means you cannot delete the SAML IDP via terraform.\n"+
					"\n5. Select some other identity provider in the Connected Identity Providers list.\n"+
					"\n6. After changing the identity provider from SAML to some other identity provider, you can delete the SAML IDP via the Citrix Cloud Console or via terraform.\n",
			)
		} else {
			diagnostics.AddError(
				fmt.Sprintf("Error deleting %s Identity Provider with id: %s", idpType, idpInstanceId),
				"\n\nError message: "+util.ReadClientError(err)+errBody+
					"\n\nTransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp),
			)
		}
		return
	}
}

// Check Auth Domain Name availability
func checkAuthDomainNameAvailability(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, authDomainName string) error {
	checkAuthDomainRequest := client.CwsClient.AuthDomainsDAAS.CustomerAuthDomainsCheckGet(ctx, client.ClientConfig.CustomerId)
	checkAuthDomainRequest = checkAuthDomainRequest.Name(authDomainName)
	isAuthDomainAvailable, httpResp, err := citrixdaasclient.ExecuteWithRetry[bool](checkAuthDomainRequest, client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			return nil
		}
		diagnostics.AddError(
			fmt.Sprintf("Error checking availability for Auth Domain Name %s", authDomainName),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if !isAuthDomainAvailable {
		err := fmt.Errorf("the Auth Domain Name %s is already in use", authDomainName)
		diagnostics.AddError(
			"Error creating Saml Identity Provider",
			err.Error(),
		)
		return err
	}

	return nil
}
