// Copyright © 2026. Citrix Systems, Inc.

package cc_admin_user // whitebox testing in same package

import (
	"context"
	"regexp"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/ccadmins"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// newRemotePolicy creates a selected AdministratorAccessPolicyModel without scopes.
func newRemotePolicy(displayName, serviceName string) ccadmins.AdministratorAccessPolicyModel {
	checkable := *ccadmins.NewBooleanPolicyProperty()
	checkable.SetValue(true)

	policy := ccadmins.AdministratorAccessPolicyModel{}
	policy.SetDisplayName(displayName)
	policy.SetServiceName(serviceName)
	policy.SetCheckable(checkable)
	return policy
}

// newStatePolicy creates a CCAdminPolicyResourceModel, leaving service_name unknown when empty to mimic a plan that omits it.
func newStatePolicy(name, serviceName string) CCAdminPolicyResourceModel {
	policy := CCAdminPolicyResourceModel{
		Name:        types.StringValue(name),
		ServiceName: types.StringUnknown(),
		Scopes:      types.SetNull(types.StringType),
	}
	if serviceName != "" {
		policy.ServiceName = types.StringValue(serviceName)
	}
	return policy
}

func newAdminUserWithPolicies(ctx context.Context, diagnostics *diag.Diagnostics, policies []CCAdminPolicyResourceModel) CCAdminUserResourceModel {
	adminUser := CCAdminUserResourceModel{}
	if policies == nil {
		attributeMap, err := util.ResourceAttributeMapFromObject(CCAdminPolicyResourceModel{})
		if err != nil {
			diagnostics.AddError("Error creating policy attribute map", err.Error())
		}
		adminUser.Policies = types.ListNull(types.ObjectType{AttrTypes: attributeMap})
	} else {
		adminUser.Policies = util.TypedArrayToObjectList[CCAdminPolicyResourceModel](ctx, diagnostics, policies)
	}
	return adminUser
}

type policyNameAndService struct {
	name        string
	serviceName string
}

func getPolicyNamesAndServices(ctx context.Context, diagnostics *diag.Diagnostics, adminUser CCAdminUserResourceModel) []policyNameAndService {
	result := []policyNameAndService{}
	for _, policy := range util.ObjectListToTypedArray[CCAdminPolicyResourceModel](ctx, diagnostics, adminUser.Policies) {
		result = append(result, policyNameAndService{policy.Name.ValueString(), policy.ServiceName.ValueString()})
	}
	return result
}

// ========== Tests ==========

func TestRefreshPropertyValuesForPolicies(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name           string
		statePolicies  []CCAdminPolicyResourceModel
		remotePolicies []ccadmins.AdministratorAccessPolicyModel
		expected       []policyNameAndService
	}{
		{
			name:          "import with same name under two services keeps both",
			statePolicies: nil,
			remotePolicies: []ccadmins.AdministratorAccessPolicyModel{
				newRemotePolicy("Help Desk", "XenDesktop"),
				newRemotePolicy("Help Desk", "Platform"),
			},
			expected: []policyNameAndService{{"Help Desk", "XenDesktop"}, {"Help Desk", "Platform"}},
		},
		{
			name: "same name under two services matches by service",
			statePolicies: []CCAdminPolicyResourceModel{
				newStatePolicy("Help Desk", "Platform"),
				newStatePolicy("Help Desk", "XenDesktop"),
			},
			remotePolicies: []ccadmins.AdministratorAccessPolicyModel{
				newRemotePolicy("Help Desk", "XenDesktop"),
				newRemotePolicy("Help Desk", "Platform"),
			},
			expected: []policyNameAndService{{"Help Desk", "Platform"}, {"Help Desk", "XenDesktop"}},
		},
		{
			name: "omitted service name resolves and keeps config order",
			statePolicies: []CCAdminPolicyResourceModel{
				newStatePolicy("Read Only", ""),
				newStatePolicy("Help Desk", ""),
			},
			remotePolicies: []ccadmins.AdministratorAccessPolicyModel{
				newRemotePolicy("Help Desk", "XenDesktop"),
				newRemotePolicy("Read Only", "Platform"),
			},
			expected: []policyNameAndService{{"Read Only", "Platform"}, {"Help Desk", "XenDesktop"}},
		},
		{
			name: "omitted service name resolves when the name matches several remote policies",
			statePolicies: []CCAdminPolicyResourceModel{
				newStatePolicy("Help Desk", ""),
			},
			remotePolicies: []ccadmins.AdministratorAccessPolicyModel{
				newRemotePolicy("Help Desk", "Platform"),
				newRemotePolicy("Help Desk"+util.AdminUserMonitorAccessPolicySuffix, "XenDesktop"),
			},
			expected: []policyNameAndService{{"Help Desk", "Platform"}, {"Help Desk", "XenDesktop"}},
		},
		{
			name: "service name matches regardless of case",
			statePolicies: []CCAdminPolicyResourceModel{
				newStatePolicy("Help Desk", "xendesktop"),
			},
			remotePolicies: []ccadmins.AdministratorAccessPolicyModel{
				newRemotePolicy("Help Desk", "XenDesktop"),
			},
			expected: []policyNameAndService{{"Help Desk", "XenDesktop"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			diagnostics := &diag.Diagnostics{}
			adminUser := newAdminUserWithPolicies(ctx, diagnostics, tt.statePolicies)
			accessModel := &ccadmins.AdministratorAccessModel{}
			accessModel.SetPolicies(tt.remotePolicies)

			result := adminUser.RefreshPropertyValuesForPolicies(ctx, diagnostics, accessModel)

			require.False(t, diagnostics.HasError(), "unexpected diagnostics: %v", diagnostics)
			assert.Equal(t, tt.expected, getPolicyNamesAndServices(ctx, diagnostics, result))
		})
	}
}

// The onboarding script reads the backticked names in the description, so it must list exactly what the validator accepts.
func TestServiceNameDescriptionListsSupportedServiceNames(t *testing.T) {
	matches := regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(getServiceNameDescription(), -1)
	described := []string{}
	for _, match := range matches {
		described = append(described, match[1])
	}

	assert.Equal(t, supportedServiceNames, described)
	assert.Subset(t, supportedServiceNames, []string{"MAS", "NetscalerGateway", "SDWANOrchestrator"})
}

func TestRefreshPropertyValuesForPolicies_RepeatedRefreshIsStable(t *testing.T) {
	ctx := context.Background()
	diagnostics := &diag.Diagnostics{}
	adminUser := newAdminUserWithPolicies(ctx, diagnostics, nil)
	accessModel := &ccadmins.AdministratorAccessModel{}
	accessModel.SetPolicies([]ccadmins.AdministratorAccessPolicyModel{
		newRemotePolicy("Help Desk", "XenDesktop"),
		newRemotePolicy("Help Desk", "Platform"),
	})
	expected := []policyNameAndService{{"Help Desk", "XenDesktop"}, {"Help Desk", "Platform"}}

	for range 3 {
		adminUser = adminUser.RefreshPropertyValuesForPolicies(ctx, diagnostics, accessModel)

		require.False(t, diagnostics.HasError(), "unexpected diagnostics: %v", diagnostics)
		assert.Equal(t, expected, getPolicyNamesAndServices(ctx, diagnostics, adminUser))
	}
}
