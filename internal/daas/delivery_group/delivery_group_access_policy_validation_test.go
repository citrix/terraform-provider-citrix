// Copyright © 2026. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// policyWithUsers builds a declared access policy, optionally setting restricted_access_users.
// A nil allowList means the attribute is left unset.
func policyWithUsers(ctx context.Context, t *testing.T, name string, allowedConnection string, allowList []string, blockList []string, setUsers bool) DeliveryGroupAccessPolicyModel {
	t.Helper()

	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}
	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}

	restrictedAccessUsers := types.ObjectNull(usersAttributes)
	if setUsers {
		restrictedAccessUsers = restrictedAccessUsersObject(ctx, t, userSet(ctx, t, allowList...), userSet(ctx, t, blockList...))
	}

	return DeliveryGroupAccessPolicyModel{
		Name:                   types.StringValue(name),
		AllowedConnection:      types.StringValue(allowedConnection),
		RestrictedAccessUsers:  restrictedAccessUsers,
		IncludeCriteriaFilters: types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes}),
		ExcludeCriteriaFilters: types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes}),
	}
}

func policyListOf(ctx context.Context, t *testing.T, policies ...DeliveryGroupAccessPolicyModel) types.List {
	t.Helper()
	var diags diag.Diagnostics
	list := util.TypedArrayToObjectList(ctx, &diags, policies)
	if diags.HasError() {
		t.Fatalf("failed to construct access policy list: %s", diags)
	}
	return list
}

// TestAllAccessPoliciesShareConfiguredUsers covers the validation that rejects a delivery-group-wide
// restriction written one access policy at a time. Such a configuration applies, but the next
// refresh reports the users back at the delivery group level and the plan never converges.
func TestAllAccessPoliciesShareConfiguredUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	nullList := types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)})

	type testCase struct {
		defaultPolicies []DeliveryGroupAccessPolicyModel
		customPolicies  []DeliveryGroupAccessPolicyModel
		expectRejected  bool
	}

	tests := map[string]testCase{
		"identical users on every policy is rejected": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"domain\\user1"}, nil, true),
			},
			expectRejected: true,
		},
		"identical users differing only in case is rejected": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"DOMAIN\\User1"}, nil, true),
			},
			expectRejected: true,
		},
		"identical users listed in a different order is rejected": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1", "domain\\user2"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"domain\\user2", "domain\\user1"}, nil, true),
			},
			expectRejected: true,
		},
		"identical users spanning default and custom policies is rejected": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"domain\\user1"}, nil, true),
			},
			customPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "MyCustomPolicy", "ViaAG", []string{"domain\\user1"}, nil, true),
			},
			expectRejected: true,
		},
		// The built-in policies are left unrestricted when default_access_policies is omitted, so
		// custom policies sharing users are still contextual access for the delivery group.
		"identical users on custom policies alone is contextual access": {
			customPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "FirstCustomPolicy", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "SecondCustomPolicy", "ViaAG", []string{"domain\\user1"}, nil, true),
			},
		},
		// Only one built-in policy is declared, so the other stays unrestricted.
		"identical users with one built-in policy left undeclared is contextual access": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
			},
			customPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "MyCustomPolicy", "ViaAG", []string{"domain\\user1"}, nil, true),
			},
		},
		"differing allow lists are contextual access": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"domain\\user2"}, nil, true),
			},
		},
		"differing block lists are contextual access": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, []string{"domain\\user2"}, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", []string{"domain\\user1"}, nil, true),
			},
		},
		"one policy leaving the attribute unset is contextual access": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", nil, nil, false),
			},
		},
		"a single declared policy is always allowed": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", []string{"domain\\user1"}, nil, true),
			},
		},
		"no policy setting the attribute is allowed": {
			defaultPolicies: []DeliveryGroupAccessPolicyModel{
				policyWithUsers(ctx, t, "Citrix Gateway connections", "ViaAG", nil, nil, false),
				policyWithUsers(ctx, t, "Non-Citrix Gateway Connections", "NotViaAG", nil, nil, false),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defaultPolicies := nullList
			if len(test.defaultPolicies) > 0 {
				defaultPolicies = policyListOf(ctx, t, test.defaultPolicies...)
			}
			customPolicies := nullList
			if len(test.customPolicies) > 0 {
				customPolicies = policyListOf(ctx, t, test.customPolicies...)
			}

			var diags diag.Diagnostics
			rejected := allAccessPoliciesShareConfiguredUsers(ctx, &diags, defaultPolicies, customPolicies)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}
			if rejected != test.expectRejected {
				t.Fatalf("expected rejected=%v, got %v", test.expectRejected, rejected)
			}
		})
	}
}
