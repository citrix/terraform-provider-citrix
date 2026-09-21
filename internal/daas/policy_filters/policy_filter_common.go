// Copyright © 2026. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
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

// policyLocks holds one mutex per policy, keyed on the lowercased policy GUID because DaaS
// does not guarantee GUID casing.
var policyLocks sync.Map

// lockPolicy acquires the lock for a policy and returns its unlock function.
func lockPolicy(policyId string) func() {
	value, _ := policyLocks.LoadOrStore(strings.ToLower(policyId), &sync.Mutex{})
	mutex := value.(*sync.Mutex) //nolint:forcetypeassert // only *sync.Mutex is ever stored
	mutex.Lock()
	return mutex.Unlock
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

	return postPolicyFilter(ctx, client, diagnostics, policyFilter)
}

// postPolicyFilter issues the create without the policy existence check, so callers holding
// the policy lock can run that check outside it.
func postPolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilter PolicyFilterInterface) (*citrixorchestration.FilterResponse, error) {
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

func readPolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilterId string) (*citrixorchestration.FilterResponse, error) {
	getPolicyFilterRequest := client.ApiClient.GpoDAAS.GpoReadGpoFilter(ctx, policyFilterId)
	policyFilter, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.FilterResponse](getPolicyFilterRequest, client)
	if err != nil {
		// Check if this is a 404 Not Found error - return a specific error that can be handled by the caller
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %w", util.ErrPolicyFilterNotFound, err)
		}

		diagnostics.AddError(
			"Error Reading Policy Filter "+policyFilterId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policyFilter, nil
}

// getPolicyFilter reads a filter back after a create or update, retrying on 404 because DaaS
// can briefly not find one it just accepted. Read and the data sources use readPolicyFilter
// instead, which must keep treating 404 as "gone".
func getPolicyFilter(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyFilterId string) (*citrixorchestration.FilterResponse, error) {
	getPolicyFilterRequest := client.ApiClient.GpoDAAS.GpoReadGpoFilter(ctx, policyFilterId)
	policyFilter, httpResp, err := citrixdaasclient.ExecuteWithRetryOnNotFound[*citrixorchestration.FilterResponse](getPolicyFilterRequest, client)
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

func getPolicyFilters(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policyGuid string) ([]citrixorchestration.FilterResponse, error) {
	getPolicyFiltersRequest := client.ApiClient.GpoDAAS.GpoReadGpoFilters(ctx)
	getPolicyFiltersRequest = getPolicyFiltersRequest.PolicyGuid(policyGuid)
	policyFilters, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfFilterResponse](getPolicyFiltersRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Filters",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return policyFilters.GetItems(), nil
}

var ErrDuplicateDeliveryGroupFilter = errors.New("duplicate delivery group policy filter")

// findConflictingDeliveryGroupFilter returns the existing DesktopGroup filter on policyId that
// already targets deliveryGroupId. Matching ignores `allowed`, per XAC-71644.
func findConflictingDeliveryGroupFilter(policyFilters []citrixorchestration.FilterResponse, policyId string, deliveryGroupId string) (citrixorchestration.FilterResponse, bool) {
	for _, policyFilter := range policyFilters {
		if !strings.EqualFold(policyFilter.GetPolicyGuid(), policyId) || policyFilter.GetFilterType() != "DesktopGroup" {
			continue
		}

		var uuidFilterData util.PolicyFilterUuidDataClientModel
		if err := json.Unmarshal([]byte(policyFilter.GetFilterData()), &uuidFilterData); err != nil {
			continue
		}

		if strings.EqualFold(uuidFilterData.Uuid, deliveryGroupId) {
			return policyFilter, true
		}
	}

	return citrixorchestration.FilterResponse{}, false
}

func buildDuplicateDeliveryGroupFilterError(existingFilterId string, deliveryGroupId string, policyId string) string {
	return fmt.Sprintf(
		"A Delivery Group Policy Filter (ID: %s) for Delivery Group %s already exists on policy %s.\n\n"+
			"If this filter is not managed by Terraform, import it instead of creating a new one:\n"+
			"  terraform import <resource_address> %s\n\n"+
			"If a previous apply created the filter without recording it, importing it will reconcile the state.",
		existingFilterId, deliveryGroupId, policyId, existingFilterId,
	)
}

// createDeliveryGroupFilterChecked runs the duplicate check and the create under the policy
// lock, so no other Create on the same policy can list between the two. The policy existence
// check is kept outside the lock because it retries 404 for ~105s and would otherwise
// serialise across every filter on the policy.
func createDeliveryGroupFilterChecked(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan DeliveryGroupFilterModel) (*citrixorchestration.FilterResponse, error) {
	if _, err := policy_resource.GetPolicy(ctx, client, diagnostics, plan.GetPolicyId(), true, false); err != nil {
		return nil, err
	}

	defer lockPolicy(plan.GetPolicyId())()

	policyFilters, err := getPolicyFilters(ctx, client, diagnostics, plan.GetPolicyId())
	if err != nil {
		return nil, err
	}

	if conflict, found := findConflictingDeliveryGroupFilter(policyFilters, plan.GetPolicyId(), plan.DeliveryGroupId.ValueString()); found {
		diagnostics.AddError(
			"Error creating Delivery Group Policy Filter",
			buildDuplicateDeliveryGroupFilterError(conflict.GetFilterGuid(), plan.DeliveryGroupId.ValueString(), plan.GetPolicyId()),
		)
		return nil, ErrDuplicateDeliveryGroupFilter
	}

	return postPolicyFilter(ctx, client, diagnostics, plan)
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
