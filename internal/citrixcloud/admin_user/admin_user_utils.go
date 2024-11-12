// Copyright © 2024. Citrix Systems, Inc.

package cc_admin_user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ccadmins "github.com/citrix/citrix-daas-rest-go/ccadmins"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// List of AdministratorServiceNames
const (
	ADMINISTRATORSERVICENAME_XENDESKTOP string = "XenDesktop"
	ADMINISTRATORSERVICENAME_PLATFORM   string = "Platform"
	ADMINISTRATORACCESSTYPE_WEM         string = "WEM"
	ADMINISTRATORACCESSTYPE_CAS         string = "CAS"
)

func getAdminProviderType(providerType string) (ccadmins.AdministratorProviderType, error) {
	switch providerType {
	case string(ccadmins.ADMINISTRATORPROVIDERTYPE_CITRIX_STS):
		return ccadmins.ADMINISTRATORPROVIDERTYPE_CITRIX_STS, nil
	case string(ccadmins.ADMINISTRATORPROVIDERTYPE_AZURE_AD):
		return ccadmins.ADMINISTRATORPROVIDERTYPE_AZURE_AD, nil
	case string(ccadmins.ADMINISTRATORPROVIDERTYPE_AD):
		return ccadmins.ADMINISTRATORPROVIDERTYPE_AD, nil
	case string(ccadmins.ADMINISTRATORPROVIDERTYPE_GOOGLE):
		return ccadmins.ADMINISTRATORPROVIDERTYPE_GOOGLE, nil
	default:
		return "", fmt.Errorf("unable to parse admin provider type %s", providerType)
	}
}

func getAdminAccessType(accessType string) (ccadmins.AdministratorAccessType, error) {
	switch accessType {
	case string(ccadmins.ADMINISTRATORACCESSTYPE_FULL):
		return ccadmins.ADMINISTRATORACCESSTYPE_FULL, nil
	case string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM):
		return ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM, nil
	default:
		return "", fmt.Errorf("unable to parse admin access type %s", accessType)
	}
}

func isInvitationAccepted(adminUserResource CCAdminUserResourceModel) bool {
	return !adminUserResource.AdminId.IsNull() && adminUserResource.AdminId.ValueString() != ""
}

func isCustomAccessTypeWithPolicies(adminUserResource CCAdminUserResourceModel) bool {
	return adminUserResource.AccessType.ValueString() == string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM) && !adminUserResource.Policies.IsNull()
}

func getAdminUser(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, adminUserResource CCAdminUserResourceModel) (ccadmins.AdministratorResult, error) {
	adminUserEmail := adminUserResource.Email.ValueString()
	adminId := adminUserResource.AdminId.ValueString()
	externalUserId := adminUserResource.ExternalUserId.ValueString()

	// Initialize the request to fetch admin users
	fetchAdminUsersRequest := client.CCAdminsClient.AdministratorsAPI.FetchAdministrators(ctx)
	fetchAdminUsersRequest = fetchAdminUsersRequest.CitrixCustomerId(client.ClientConfig.CustomerId)
	var adminUser ccadmins.AdministratorResult

	for {
		// Execute the request with retry logic
		adminUsersResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*ccadmins.AdministratorsResult](fetchAdminUsersRequest, client)
		if err != nil {
			err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
			return adminUser, err
		}

		// Check if the user is already present
		for _, adminUser := range adminUsersResponse.GetItems() {
			if (adminId != "" && (adminUser.GetUserId() == adminId || adminUser.GetUcOid() == adminId)) ||
				(externalUserId != "" && strings.EqualFold(getExternalUserId(adminUser.GetExternalOid()), externalUserId)) ||
				(adminUserEmail != "" && strings.EqualFold(adminUser.GetEmail(), adminUserEmail)) {
				return adminUser, nil
			}
		}

		// Check if there is a continuation token for more results
		if adminUsersResponse.GetContinuationToken() == "" {
			break
		}
		fetchAdminUsersRequest = fetchAdminUsersRequest.RequestContinuation(adminUsersResponse.GetContinuationToken())
	}

	var identifier string
	if adminUserEmail != "" {
		identifier = fmt.Sprintf("email: %s", adminUserEmail)
	} else if adminId != "" {
		identifier = fmt.Sprintf("id: %s", adminId)
	} else if externalUserId != "" {
		identifier = fmt.Sprintf("external user id: %s", externalUserId)
	}
	return adminUser, fmt.Errorf("could not find admin user %s", identifier)
}

func getAdminUserPolicies(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, adminUserResourceModel CCAdminUserResourceModel) ([]ccadmins.AdministratorAccessPolicyModel, error) {

	// If access type is Full or policies are not set, return nil
	if strings.EqualFold(adminUserResourceModel.AccessType.ValueString(), string(ccadmins.ADMINISTRATORACCESSTYPE_FULL)) && adminUserResourceModel.Policies.IsNull() {
		return nil, nil
	}

	// If access type is Custom, retrieve and add policies
	if strings.EqualFold(adminUserResourceModel.AccessType.ValueString(), string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM)) {
		return fetchAdminPoliciesWithRetry(ctx, diagnostics, client, adminUserResourceModel)
	}
	return nil, fmt.Errorf("invalid access type")
}

func fetchAdminPoliciesWithRetry(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, adminUserResourceModel CCAdminUserResourceModel) ([]ccadmins.AdministratorAccessPolicyModel, error) {
	adminId, err := getAdminIdFromAuthToken(client)
	if err != nil {
		err = fmt.Errorf("Unable to verify access of the admin user\n" + err.Error())
		return nil, err
	}

	if adminId == "" {
		err = fmt.Errorf("admin user not found")
		return nil, err
	}

	// Adding a retry logic to fetch the admin policies when a role name or scope name is not found. This is necessary because when a new role or scope is created using the orchestration API, the CC Admin API may take up to 3 minutes
	// to reflect the updated list. Therefore, we need to wait before throwing an error indicating that the value is not found.
	const (
		maxRetries    = 7
		retryInterval = 30 * time.Second
	)

	var (
		accessPolicies          *ccadmins.AdministratorAccessModel
		adminPolicyAccessModels []ccadmins.AdministratorAccessPolicyModel
		itemFound               bool
	)

	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		accessPolicies, err = getAccessPolicies(ctx, client, adminId)
		if err != nil {
			return nil, err
		}

		adminPolicyAccessModels, itemFound, err = fetchAdminPolicyAccessModels(ctx, diagnostics, adminUserResourceModel, accessPolicies)
		// If no error or item found, break the loop
		if err == nil || itemFound {
			break
		}
		// Sleep for the retry interval before trying again
		time.Sleep(retryInterval)
	}

	if err != nil {
		err = fmt.Errorf("error fetching admin policy access models\n" + err.Error())
		return nil, err
	}

	return adminPolicyAccessModels, nil
}

func fetchAdminPolicyAccessModels(ctx context.Context, diagnostics *diag.Diagnostics, adminUserResourceModel CCAdminUserResourceModel, accessPolicies *ccadmins.AdministratorAccessModel) ([]ccadmins.AdministratorAccessPolicyModel, bool, error) {
	policies := util.ObjectListToTypedArray[CCAdminPolicyResourceModel](ctx, diagnostics, adminUserResourceModel.Policies)
	adminPolicyAccessModels := []ccadmins.AdministratorAccessPolicyModel{}
	for _, policy := range policies {
		adminAccessPolicyModel, itemFound, err := getAdminAccessPolicy(ctx, diagnostics, policy, accessPolicies.GetPolicies())
		if err != nil {
			return nil, itemFound, err
		}
		adminPolicyAccessModels = append(adminPolicyAccessModels, adminAccessPolicyModel)
	}
	return adminPolicyAccessModels, false, nil
}

func getAdminAccessPolicy(ctx context.Context, diagnostics *diag.Diagnostics, adminPolicyResourceModel CCAdminPolicyResourceModel, remoteAdminPolicies []ccadmins.AdministratorAccessPolicyModel) (ccadmins.AdministratorAccessPolicyModel, bool, error) {
	policyDisplayName := adminPolicyResourceModel.Name.ValueString()
	serviceName := adminPolicyResourceModel.ServiceName.ValueString()
	scopes := util.StringSetToStringArray(ctx, diagnostics, adminPolicyResourceModel.Scopes)

	checkable := *ccadmins.NewBooleanPolicyProperty()
	checkable.SetValue(true)
	checkable.SetCanChangeValue(true)

	policyNameExists := false
	createAdminPolicyModel := ccadmins.AdministratorAccessPolicyModel{}
	var serviceNameList []string
	for _, remotePolicy := range remoteAdminPolicies {
		trimmedRemotePolicyDisplayName := strings.TrimSuffix(remotePolicy.GetDisplayName(), util.AdminUserMonitorAccessPolicySuffix)
		if strings.EqualFold(remotePolicy.GetDisplayName(), policyDisplayName) || strings.EqualFold(trimmedRemotePolicyDisplayName, policyDisplayName) {
			// If service name is specified, check if the policy is associated with the service
			if serviceName != "" && !strings.EqualFold(remotePolicy.GetServiceName(), serviceName) {
				continue
			}
			policyNameExists = true
			serviceNameList = append(serviceNameList, remotePolicy.GetServiceName())
			createAdminPolicyModel.SetName(remotePolicy.GetName())
			createAdminPolicyModel.SetServiceName(remotePolicy.GetServiceName())
			createAdminPolicyModel.SetDisplayName(remotePolicy.GetDisplayName())
			createAdminPolicyModel.SetCheckable(checkable)
			createScopeChoices := ccadmins.AdministratorAccessScopeChoices{}
			createScopeChoices.SetAllScopesSelected(false)
			remotePolicyScopeChoices := remotePolicy.GetScopeChoices()
			if remotePolicyScopeChoices.GetChoices() != nil {
				for _, scope := range scopes {
					scopeNameExists := false
					for _, remoteScopeChoice := range remotePolicyScopeChoices.GetChoices() {
						if strings.EqualFold(remoteScopeChoice.GetDisplayName(), scope) {
							scopeNameExists = true
							var createScopeChoiceModel ccadmins.AdministratorAccessScopeChoicesModel
							createScopeChoiceModel.SetName(remoteScopeChoice.GetName())
							createScopeChoiceModel.SetDisplayName(remoteScopeChoice.GetDisplayName())
							createScopeChoiceModel.SetCheckable(checkable)
							createScopeChoices.Choices = append(createScopeChoices.Choices, createScopeChoiceModel)
						}
					}
					if !scopeNameExists {
						err := fmt.Errorf("scope with name: %s not found", scope)
						return createAdminPolicyModel, false, err
					}
					createAdminPolicyModel.SetScopeChoices(createScopeChoices)
				}
			} else if len(scopes) > 0 {
				err := fmt.Errorf("policy with name: %s does not contain any scopes", policyDisplayName)
				return createAdminPolicyModel, true, err
			} else {
				createAdminPolicyModel.SetScopeChoices(remotePolicyScopeChoices)
			}
		}
	}
	if !policyNameExists {
		err := fmt.Errorf("policy with name: %s not found", policyDisplayName)
		return createAdminPolicyModel, false, err
	}
	if len(serviceNameList) > 1 {
		err := fmt.Errorf("policy with name: %s is associated with multiple services %v. Please specify one of the services in the 'service_name' attribute", policyDisplayName, strings.Join(serviceNameList, ", "))
		return createAdminPolicyModel, true, err
	}

	if createAdminPolicyModel.GetServiceName() == ADMINISTRATORSERVICENAME_XENDESKTOP && len(scopes) == 0 {
		err := fmt.Errorf("policy '%s' with service name '%s' has no scopes; please add scope values", policyDisplayName, ADMINISTRATORSERVICENAME_XENDESKTOP)
		return createAdminPolicyModel, true, err
	}
	return createAdminPolicyModel, true, nil
}

func getAccessPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, adminId string) (*ccadmins.AdministratorAccessModel, error) {
	getAccessPoliciesRequest := client.CCAdminsClient.AdministratorsAPI.GetAdministratorAccess(ctx, adminId)
	accessPoliciesResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*ccadmins.AdministratorAccessModel](getAccessPoliciesRequest, client)
	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return accessPoliciesResponse, err
	}
	return accessPoliciesResponse, nil
}

func getAdminIdFromAuthToken(client *citrixdaasclient.CitrixDaasClient) (string, error) {
	authToken, _, err := client.SignIn()
	if err != nil {
		return "", fmt.Errorf("failed to sign in: %v", err)
	}

	if authToken == "" {
		return "", fmt.Errorf("received empty auth token")
	}

	// Split the token into its parts
	parts := strings.Split(authToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid auth token format")
	}

	// Decode the payload part (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode auth token payload: %v", err)
	}

	// Parse the JSON payload
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse auth token JSON payload: %v", err)
	}

	// Fetch the user_id claim
	userId, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in the auth token")
	}
	return userId, nil
}

func fetchAndUpdateAdminUser(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, plan CCAdminUserResourceModel, diagnostics *diag.Diagnostics) (CCAdminUserResourceModel, error) {
	// Fetch the admin user from the remote source
	adminUser, err := getAdminUser(ctx, client, plan)
	if err != nil {
		diagnostics.AddError(
			"Error fetching admin user",
			util.ReadClientError(err),
		)
		return plan, err
	}

	// Update the plan with the fetched admin user details
	plan = plan.RefreshPropertyValues(ctx, diagnostics, adminUser)

	// Check if custom access type with policies is required
	if isCustomAccessTypeWithPolicies(plan) {
		if isInvitationAccepted(plan) {
			adminId := plan.AdminId.ValueString()
			// Fetch access policies for the admin user
			accessPolicies, err := getAccessPolicies(ctx, client, adminId)
			if err != nil {
				diagnostics.AddError(
					"Error getting access policies for user "+plan.Email.ValueString(),
					"Error message: "+util.ReadClientError(err),
				)
				return plan, err
			}
			// Update the plan with the fetched access policies
			plan = plan.RefreshPropertyValuesForPolicies(ctx, diagnostics, accessPolicies)
		} else {
			// If the invitation is not accepted
			var filteredPolicies []CCAdminPolicyResourceModel
			policies := util.ObjectListToTypedArray[CCAdminPolicyResourceModel](ctx, diagnostics, plan.Policies)
			for _, policy := range policies {
				// Set the service name to empty string if it is not set as its a computed field
				if policy.ServiceName.IsNull() || policy.ServiceName.ValueString() == "" {
					policy.ServiceName = types.StringValue("")
				}
				filteredPolicies = append(filteredPolicies, policy)
			}
			plan.Policies = util.TypedArrayToObjectList[CCAdminPolicyResourceModel](ctx, diagnostics, filteredPolicies)
		}
	}
	return plan, nil
}

// Checks if the provider type exists in the list of legacy providers
func providerTypeExists(remoteProviderTypes []string, providerType string) bool {
	if providerType != "" {
		for _, pt := range remoteProviderTypes {
			if strings.EqualFold(pt, providerType) {
				return true
			}
		}
	}
	return false
}

func getExternalUserId(externalOid string) string {
	externalOidParts := strings.Split(externalOid, "/")
	return externalOidParts[len(externalOidParts)-1]
}
