// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"fmt"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policy_resource"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ PolicyFilterInterface = AccessControlFilterModel{}
	_ PolicyFilterInterface = BranchRepeaterFilterModel{}
	_ PolicyFilterInterface = ClientIPFilterModel{}
	_ PolicyFilterInterface = ClientNameFilterModel{}
	_ PolicyFilterInterface = ClientPlatformFilterModel{}
	_ PolicyFilterInterface = DeliveryGroupFilterModel{}
	_ PolicyFilterInterface = DeliveryGroupTypeFilterModel{}
	_ PolicyFilterInterface = OuFilterModel{}
	_ PolicyFilterInterface = UserFilterModel{}
	_ PolicyFilterInterface = TagFilterModel{}
)

type PolicyFilterInterface interface {
	GetSchema() schema.Schema
	GetAttributes() map[string]schema.Attribute
	GetDataSourceSchema() dataSourceSchema.Schema
	GetDataSourceNestedAttributeObjectSchema() dataSourceSchema.NestedAttributeObject
	GetDataSourceAttributes() map[string]dataSourceSchema.Attribute
	GetId() string
	GetPolicyId() string
	GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error)
}

type PolicyFilterUuidDataClientModel struct {
	Server string `json:"server,omitempty"`
	Uuid   string `json:"uuid,omitempty"`
}

type PolicyFilterGatewayDataClientModel struct {
	Connection string `json:"connection,omitempty"`
	Condition  string `json:"condition,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
}

func getServerValue(client *citrixdaasclient.CitrixDaasClient) string {
	if client.AuthConfig.OnPremises || !client.AuthConfig.ApiGateway {
		return client.ApiClient.GetConfig().Host
	} else {
		switch client.AuthConfig.Environment {
		case "Japan":
			return fmt.Sprintf("%s.xendesktop.jp", client.ClientConfig.CustomerId)
		case "Gov":
			return fmt.Sprintf("%s.xendesktop.us", client.ClientConfig.CustomerId)
		case "GovStaging":
			return fmt.Sprintf("%s.xdstaging.us", client.ClientConfig.CustomerId)
		default:
			return fmt.Sprintf("%s.xendesktop.net", client.ClientConfig.CustomerId)
		}
	}
}

func createPolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilter PolicyFilterInterface) (*citrixorchestration.FilterResponse, error) {
	_, err := policy_resource.GetPolicy(ctx, client, diagnostics, policyFilter.GetPolicyId(), true, false)
	if err != nil {
		return nil, err
	}

	serverValue := getServerValue(client)
	createFilterRequestBody, err := policyFilter.GetFilterRequest(diagnostics, serverValue)
	if err != nil {
		return nil, err
	}

	createPolicyFilterRequest := client.ApiClient.GpoDAAS.GpoCreateGpoFilter(ctx)
	createPolicyFilterRequest = createPolicyFilterRequest.FilterRequest(createFilterRequestBody)
	createPolicyFilterRequest = createPolicyFilterRequest.PolicyGuid(policyFilter.GetPolicyId())

	policyFilterCreated, httpResp, err := citrixdaasclient.AddRequestData(createPolicyFilterRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Creating Policy Filter",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policyFilterCreated, nil
}

func getPolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilterId string) (*citrixorchestration.FilterResponse, error) {
	getPolicyFilterRequest := client.ApiClient.GpoDAAS.GpoReadGpoFilter(ctx, policyFilterId)
	policyFilter, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.FilterResponse](getPolicyFilterRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Filter "+policyFilterId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policyFilter, nil
}

func updatePolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilter PolicyFilterInterface) error {
	serverValue := getServerValue(client)
	filterRequest, err := policyFilter.GetFilterRequest(diagnostics, serverValue)
	if err != nil {
		return err
	}
	editPolicyFilterRequest := client.ApiClient.GpoDAAS.GpoUpdateGpoFilter(ctx, policyFilter.GetId())
	editPolicyFilterRequest = editPolicyFilterRequest.FilterRequest(filterRequest)

	// Update policy setting
	httpResp, err := citrixdaasclient.AddRequestData(editPolicyFilterRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Updating Policy Filter "+policyFilter.GetId(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}

func deletePolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilterId string) error {
	deletePolicyFilterRequest := client.ApiClient.GpoDAAS.GpoDeleteGpoFilter(ctx, policyFilterId)
	httpResp, err := citrixdaasclient.AddRequestData(deletePolicyFilterRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error Deleting Policy Filter "+policyFilterId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	return nil
}

func getPolicyFilterResourceDescription(policyFilterName string) string {
	return fmt.Sprintf("CVAD --- Manages an instance of the %s Policy Filter.", policyFilterName) +
		"\n\n -> **Please Note** For detailed information about policy filters, please refer to [this document](https://github.com/citrix/terraform-provider-citrix/blob/main/internal/daas/policies/policy_set_resource.md)."
}

func getPolicyFilterDataSourceDescription(policyFilterName string) string {
	return fmt.Sprintf("CVAD --- Data source of an instance of the %s Policy Filter.", policyFilterName)
}
