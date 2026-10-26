// Copyright © 2026. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"strings"
	"testing"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newIdentityUser builds a remote user entry as returned in an access policy's user filters.
func newIdentityUser(samName string) citrixorchestration.IdentityUserResponseModel {
	user := citrixorchestration.IdentityUserResponseModel{}
	user.SetSamName(samName)
	return user
}

// newIdentityUsers maps SAM names to remote user entries.
func newIdentityUsers(samNames ...string) []citrixorchestration.IdentityUserResponseModel {
	users := []citrixorchestration.IdentityUserResponseModel{}
	for _, samName := range samNames {
		users = append(users, newIdentityUser(samName))
	}
	return users
}

// withStaleUsers leaves a policy's filter-enabled flags untouched while populating the user lists,
// reproducing a policy whose filter has been turned off but whose users were never cleared because
// IncludedUsers and ExcludedUsers are omitted from the request model when empty.
func withStaleUsers(policy citrixorchestration.AdvancedAccessPolicyResponseModel, includedUsers, excludedUsers []string) citrixorchestration.AdvancedAccessPolicyResponseModel {
	if includedUsers != nil {
		policy.SetIncludedUsers(newIdentityUsers(includedUsers...))
	}
	if excludedUsers != nil {
		policy.SetExcludedUsers(newIdentityUsers(excludedUsers...))
	}
	return policy
}

// newBuiltInAccessPolicy builds a built-in advanced access policy carrying the given user filters.
// A nil list means the corresponding filter is reported as disabled by the API, which is what
// happens once the last user is removed from the list.
func newBuiltInAccessPolicy(allowedConnection string, includedUsers, excludedUsers []string) citrixorchestration.AdvancedAccessPolicyResponseModel {
	policy := citrixorchestration.AdvancedAccessPolicyResponseModel{}
	policy.SetIsBuiltIn(true)
	policy.SetAllowedConnection(citrixorchestration.AllowedConnection(allowedConnection))
	policy.SetAllowedUsers(citrixorchestration.ALLOWEDUSER_FILTERED)
	policy.SetIncludedUserFilterEnabled(len(includedUsers) > 0)
	policy.SetIncludedUsers(newIdentityUsers(includedUsers...))
	policy.SetExcludedUserFilterEnabled(len(excludedUsers) > 0)
	policy.SetExcludedUsers(newIdentityUsers(excludedUsers...))
	return policy
}

// userSet builds the Terraform set value for an allow_list or block_list.
func userSet(ctx context.Context, t *testing.T, users ...string) types.Set {
	t.Helper()
	var diags diag.Diagnostics
	set := util.StringArrayToStringSet(ctx, &diags, users)
	if diags.HasError() {
		t.Fatalf("failed to construct user set: %s", diags)
	}
	return set
}

// restrictedAccessUsersObject builds the DG-level restricted_access_users state value.
func restrictedAccessUsersObject(ctx context.Context, t *testing.T, allowList, blockList types.Set) types.Object {
	t.Helper()
	var diags diag.Diagnostics
	object := util.TypedObjectToObjectValue(ctx, &diags, RestrictedAccessUsers{
		AllowList: allowList,
		BlockList: blockList,
	})
	if diags.HasError() {
		t.Fatalf("failed to construct restricted_access_users object: %s", diags)
	}
	return object
}

// defaultAccessPoliciesList builds the state value for default_access_policies from the
// allowed_connection values the configuration manages.
func defaultAccessPoliciesList(ctx context.Context, t *testing.T, allowedConnections ...string) types.List {
	t.Helper()
	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}
	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}
	criteriaFilters := types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes})

	policies := []DeliveryGroupAccessPolicyModel{}
	for _, allowedConnection := range allowedConnections {
		policies = append(policies, DeliveryGroupAccessPolicyModel{
			AllowedConnection:      types.StringValue(allowedConnection),
			RestrictedAccessUsers:  types.ObjectNull(usersAttributes),
			IncludeCriteriaFilters: criteriaFilters,
			ExcludeCriteriaFilters: criteriaFilters,
		})
	}

	var diags diag.Diagnostics
	list := util.TypedArrayToObjectList(ctx, &diags, policies)
	if diags.HasError() {
		t.Fatalf("failed to construct default_access_policies list: %s", diags)
	}
	return list
}

// unrestrictedAccessPolicy builds a built-in policy that does not restrict access to a set of
// users, which is how the API reports a delivery group every authenticated user can reach.
func unrestrictedAccessPolicy(allowedConnection string) citrixorchestration.AdvancedAccessPolicyResponseModel {
	policy := newBuiltInAccessPolicy(allowedConnection, nil, nil)
	policy.SetAllowedUsers(citrixorchestration.ALLOWEDUSER_ANY_AUTHENTICATED)
	return policy
}

// newCustomAccessPolicy builds a non built-in advanced access policy restricted to the given
// included users. Custom policies are identified by name rather than by allowed connection.
func newCustomAccessPolicy(name string, includedUsers []string) citrixorchestration.AdvancedAccessPolicyResponseModel {
	policy := newBuiltInAccessPolicy("ViaAG", includedUsers, nil)
	policy.SetIsBuiltIn(false)
	policy.SetName(name)
	return policy
}

// accessPolicyAttributes returns the attribute map of an access policy list element.
func accessPolicyAttributes(t *testing.T) map[string]attr.Type {
	t.Helper()
	attributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyModel attribute map: %s", err)
	}
	return attributes
}

// customAccessPoliciesList builds the state value for custom_access_policies from policy names.
func customAccessPoliciesList(ctx context.Context, t *testing.T, names ...string) types.List {
	t.Helper()
	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}
	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}
	criteriaFilters := types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes})

	policies := []DeliveryGroupAccessPolicyModel{}
	for _, name := range names {
		policies = append(policies, DeliveryGroupAccessPolicyModel{
			Name:                   types.StringValue(name),
			AllowedConnection:      types.StringValue("ViaAG"),
			RestrictedAccessUsers:  types.ObjectNull(usersAttributes),
			IncludeCriteriaFilters: criteriaFilters,
			ExcludeCriteriaFilters: criteriaFilters,
		})
	}

	var diags diag.Diagnostics
	list := util.TypedArrayToObjectList(ctx, &diags, policies)
	if diags.HasError() {
		t.Fatalf("failed to construct custom_access_policies list: %s", diags)
	}
	return list
}

// newDeliveryGroupWithPolicies wraps advanced access policies in a delivery group response.
func newDeliveryGroupWithPolicies(policies ...citrixorchestration.AdvancedAccessPolicyResponseModel) *citrixorchestration.DeliveryGroupDetailResponseModel {
	deliveryGroup := &citrixorchestration.DeliveryGroupDetailResponseModel{}
	deliveryGroup.SetAdvancedAccessPolicy(policies)
	return deliveryGroup
}

// TestAdvancedAccessPolicyRequestClearsUserLists pins the serialization behaviour the fix relies
// on. ToMap guards IncludedUsers and ExcludedUsers with a nil check, so a nil slice is dropped from
// the request and leaves the users a previous apply wrote on the policy in place, while a non-nil
// empty slice is sent as [] and clears them.
func TestAdvancedAccessPolicyRequestClearsUserLists(t *testing.T) {
	t.Parallel()

	nilRequest := citrixorchestration.AdvancedAccessPolicyRequestModel{}
	nilRequest.SetExcludedUserFilterEnabled(false)
	nilRequest.SetExcludedUsers(nil)
	nilBody, err := nilRequest.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal request with nil ExcludedUsers: %s", err)
	}
	if strings.Contains(string(nilBody), "ExcludedUsers") {
		t.Fatalf("expected a nil ExcludedUsers to be omitted, got %s", nilBody)
	}

	emptyRequest := citrixorchestration.AdvancedAccessPolicyRequestModel{}
	emptyRequest.SetExcludedUserFilterEnabled(false)
	emptyRequest.SetExcludedUsers(emptySliceIfNil(nil))
	emptyBody, err := emptyRequest.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal request with empty ExcludedUsers: %s", err)
	}
	if !strings.Contains(string(emptyBody), `"ExcludedUsers":[]`) {
		t.Fatalf(`expected "ExcludedUsers":[] in the request body, got %s`, emptyBody)
	}
}

// TestAccessPolicyRefreshListItemRestrictedAccessUsers covers the per-policy
// restricted_access_users on a default_access_policies / custom_access_policies entry.
// Nulling the whole object out when the API reports the include filter as disabled discarded a
// populated block_list and contradicted a plan that set the attribute, which surfaced as
// "Provider produced inconsistent result after apply" when moving restricted_access_users from
// the delivery group level down into a single access policy.
func TestAccessPolicyRefreshListItemRestrictedAccessUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}
	nullCriteria := types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes})

	type testCase struct {
		stateAllowList  []string
		stateBlockList  []string
		allowListNull   bool
		blockListNull   bool
		remoteIncluded  []string
		remoteExcluded  []string
		staleIncluded   []string
		staleExcluded   []string
		expectAllowList []string
		expectBlockList []string
		expectBlockNull bool
	}

	tests := map[string]testCase{
		// Moving restricted_access_users from the delivery group level down into a single access
		// policy leaves the users written by the previous apply on the policy with the filter
		// turned off. Appending them back grew an empty block_list to length 1.
		"stale users behind a disabled exclude filter are ignored": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{},
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			staleExcluded:   []string{"aneeshk@ctxlabs17.onmicrosoft.com"},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"stale users behind a disabled include filter are ignored": {
			stateAllowList:  []string{},
			stateBlockList:  []string{"domain\\user1"},
			remoteIncluded:  nil,
			remoteExcluded:  []string{"domain\\user1"},
			staleIncluded:   []string{"aneeshk@ctxlabs17.onmicrosoft.com"},
			expectAllowList: []string{},
			expectBlockList: []string{"domain\\user1"},
		},
		"block_list only keeps the object and the block_list": {
			stateAllowList:  []string{},
			stateBlockList:  []string{"domain\\user1"},
			remoteIncluded:  nil,
			remoteExcluded:  []string{"domain\\user1"},
			expectAllowList: []string{},
			expectBlockList: []string{"domain\\user1"},
		},
		"populated lists refresh in place": {
			stateAllowList:  []string{"domain\\user1", "domain\\user2"},
			stateBlockList:  []string{"domain\\user2"},
			remoteIncluded:  []string{"domain\\user1", "domain\\user2"},
			remoteExcluded:  []string{"domain\\user2"},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{"domain\\user2"},
		},
		"allow_list entry removed out of band is detected": {
			stateAllowList:  []string{"domain\\user1", "domain\\user2"},
			stateBlockList:  []string{},
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"every user removed out of band leaves empty lists": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{"domain\\user2"},
			remoteIncluded:  nil,
			remoteExcluded:  nil,
			expectAllowList: []string{},
			expectBlockList: []string{},
		},
		"null block_list stays null": {
			stateAllowList:  []string{"domain\\user1"},
			blockListNull:   true,
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectAllowList: []string{"domain\\user1"},
			expectBlockNull: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			allowList := userSet(ctx, t, test.stateAllowList...)
			if test.allowListNull {
				allowList = types.SetNull(types.StringType)
			}
			blockList := userSet(ctx, t, test.stateBlockList...)
			if test.blockListNull {
				blockList = types.SetNull(types.StringType)
			}

			policy := DeliveryGroupAccessPolicyModel{
				Name:                   types.StringValue(util.CitrixGatewayConnections),
				Enabled:                types.BoolValue(true),
				AllowedConnection:      types.StringValue("ViaAG"),
				IncludeCriteriaFilters: nullCriteria,
				ExcludeCriteriaFilters: nullCriteria,
				RestrictedAccessUsers:  restrictedAccessUsersObject(ctx, t, allowList, blockList),
			}

			remote := newBuiltInAccessPolicy("ViaAG", test.remoteIncluded, test.remoteExcluded)
			remote = withStaleUsers(remote, test.staleIncluded, test.staleExcluded)
			remote.SetName(util.CitrixGatewayConnections)

			var diags diag.Diagnostics
			refreshed := policy.RefreshListItem(ctx, &diags, remote)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}
			result, ok := refreshed.(DeliveryGroupAccessPolicyModel)
			if !ok {
				t.Fatalf("expected DeliveryGroupAccessPolicyModel, got %T", refreshed)
			}

			if result.RestrictedAccessUsers.IsNull() {
				t.Fatalf("expected restricted_access_users to stay non-null, got null")
			}

			users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, result.RestrictedAccessUsers)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics reading result: %s", diags)
			}

			assertSameElements(t, "allow_list", stringSetValues(ctx, t, users.AllowList), test.expectAllowList)

			if test.expectBlockNull {
				if !users.BlockList.IsNull() {
					t.Fatalf("expected block_list to be null, got %v", users.BlockList)
				}
			} else {
				assertSameElements(t, "block_list", stringSetValues(ctx, t, users.BlockList), test.expectBlockList)
			}
		})
	}
}

// TestUpdatePlanWithRestrictedAccessUsers is the regression test for GitHub issue #376.
// Before the fix, finding a managed built-in policy caused an early return that skipped the
// refresh entirely. Because RefreshPropertyValues is shared by Create, Update and Read, that also
// skipped the refresh during Read, so a user removed from allow_list or block_list outside of
// Terraform was never detected and terraform plan reported no drift.
func TestUpdatePlanWithRestrictedAccessUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type testCase struct {
		stateAllowList    []string
		stateBlockList    []string
		allowListNull     bool
		blockListNull     bool
		remoteIncluded    []string
		remoteExcluded    []string
		staleIncluded     []string
		staleExcluded     []string
		policies          []citrixorchestration.AdvancedAccessPolicyResponseModel
		customPolicyNames []string
		stateUsersNull    bool
		isRead            bool
		expectAllowList   []string
		expectBlockList   []string
		expectAllowNull   bool
		expectBlockNull   bool
		expectUsersNull   bool
	}

	const viaAG = "ViaAG"
	const notViaAG = "NotViaAG"

	tests := map[string]testCase{
		"allow_list entry removed out of band is detected": {
			stateAllowList:  []string{"domain\\user1", "domain\\user2"},
			stateBlockList:  []string{"domain\\user3"},
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  []string{"domain\\user3"},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{"domain\\user3"},
		},
		"block_list entry removed out of band is detected": {
			stateAllowList:  []string{"domain\\user1", "domain\\user2"},
			stateBlockList:  []string{"domain\\user1", "domain\\user2"},
			remoteIncluded:  []string{"domain\\user1", "domain\\user2"},
			remoteExcluded:  []string{"domain\\user1"},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{"domain\\user1"},
		},
		"last block_list entry removed out of band is detected": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{"domain\\user2"},
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"stale users behind a disabled exclude filter are ignored": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{},
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			staleExcluded:   []string{"aneeshk@ctxlabs17.onmicrosoft.com"},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"user added out of band is detected": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{"domain\\user2"},
			remoteIncluded:  []string{"domain\\user1", "domain\\user3"},
			remoteExcluded:  []string{"domain\\user2"},
			expectAllowList: []string{"domain\\user1", "domain\\user3"},
			expectBlockList: []string{"domain\\user2"},
		},
		"matching lists are left unchanged": {
			stateAllowList:  []string{"domain\\user1", "domain\\user2"},
			stateBlockList:  []string{"domain\\user2"},
			remoteIncluded:  []string{"domain\\user1", "domain\\user2"},
			remoteExcluded:  []string{"domain\\user2"},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{"domain\\user2"},
		},
		"null block_list stays null": {
			stateAllowList:  []string{"domain\\user1"},
			blockListNull:   true,
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectAllowList: []string{"domain\\user1"},
			expectBlockNull: true,
		},
		// Access opened up to every authenticated user out of band. AllowedUsers is what says so;
		// empty user lists on their own would be ambiguous with restricted_access_users = {}.
		"restriction lifted out of band nulls out restricted_access_users": {
			stateAllowList: []string{"domain\\user1"},
			stateBlockList: []string{"domain\\user2"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				unrestrictedAccessPolicy(viaAG),
				unrestrictedAccessPolicy(notViaAG),
			},
			expectUsersNull: true,
		},
		// restricted_access_users = {} set out of band: the policy still filters on users, but no
		// user matches, so nobody has access. That must not be reported as null, which means the
		// opposite.
		"no access set out of band is detected": {
			stateAllowList:  []string{"domain\\user1"},
			stateBlockList:  []string{},
			remoteIncluded:  nil,
			remoteExcluded:  nil,
			expectAllowList: []string{},
			expectBlockList: []string{},
		},
		// The configuration leaves restricted_access_users unset, so every authenticated user has
		// access. Restricting the delivery group out of band has to surface on a read.
		"restriction added out of band when the attribute is unset is detected": {
			stateUsersNull:  true,
			isRead:          true,
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"no access set out of band when the attribute is unset is detected": {
			stateUsersNull:  true,
			isRead:          true,
			remoteIncluded:  nil,
			remoteExcluded:  nil,
			expectAllowList: []string{},
			expectBlockList: []string{},
		},
		// The same remote state during an apply must leave a null configuration null, otherwise
		// Terraform rejects the result as inconsistent.
		"restriction added out of band is not applied to a null configuration during apply": {
			stateUsersNull:  true,
			isRead:          false,
			remoteIncluded:  []string{"domain\\user1"},
			remoteExcluded:  nil,
			expectUsersNull: true,
		},
		// The delivery-group level attribute is written to every managed built-in policy, so drift
		// on any one of them has to surface. Reading only the policy matching the first
		// default_access_policies entry made detection depend on the declared order.
		"drift on the second managed policy is detected": {
			stateAllowList: []string{"domain\\user1", "domain\\user2"},
			stateBlockList: []string{},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1"}, nil),
			},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"drift on the first managed policy is detected": {
			stateAllowList: []string{"domain\\user1", "domain\\user2"},
			stateBlockList: []string{},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1", "domain\\user2"}, nil),
			},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"block_list drift on the second managed policy is detected": {
			stateAllowList: []string{"domain\\user1"},
			stateBlockList: []string{"domain\\user2"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, []string{"domain\\user2"}),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1"}, nil),
			},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"policies agreeing with state are left unchanged": {
			stateAllowList: []string{"domain\\user1", "domain\\user2"},
			stateBlockList: []string{},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1", "domain\\user2"}, nil),
			},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{},
		},
		// A custom access policy receives the delivery-group level users too, so drift on one has
		// to surface even when every built-in policy still agrees with state.
		"drift on a custom policy is detected": {
			stateAllowList:    []string{"domain\\user1", "domain\\user2"},
			stateBlockList:    []string{},
			customPolicyNames: []string{"MyCustomPolicy"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newCustomAccessPolicy("MyCustomPolicy", []string{"domain\\user1"}),
			},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{},
		},
		"custom policy agreeing with state is left unchanged": {
			stateAllowList:    []string{"domain\\user1", "domain\\user2"},
			stateBlockList:    []string{},
			customPolicyNames: []string{"MyCustomPolicy"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newCustomAccessPolicy("MyCustomPolicy", []string{"domain\\user1", "domain\\user2"}),
			},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{},
		},
		"undeclared custom policy with stale users is not read": {
			stateAllowList: []string{"domain\\user1", "domain\\user2"},
			stateBlockList: []string{},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1", "domain\\user2"}, nil),
				newCustomAccessPolicy("NotInConfig", []string{"domain\\stale"}),
			},
			expectAllowList: []string{"domain\\user1", "domain\\user2"},
			expectBlockList: []string{},
		},
		"unmanaged policy with stale users is not read": {
			stateAllowList: []string{"domain\\user1"},
			stateBlockList: []string{"domain\\user2"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy("Unfiltered", []string{"domain\\stale"}, []string{"domain\\stale"}),
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, []string{"domain\\user2"}),
			},
			expectAllowList: []string{"domain\\user1"},
			expectBlockList: []string{"domain\\user2"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			allowList := userSet(ctx, t, test.stateAllowList...)
			if test.allowListNull {
				allowList = types.SetNull(types.StringType)
			}
			blockList := userSet(ctx, t, test.stateBlockList...)
			if test.blockListNull {
				blockList = types.SetNull(types.StringType)
			}

			policies := test.policies
			if policies == nil {
				policies = []citrixorchestration.AdvancedAccessPolicyResponseModel{
					withStaleUsers(newBuiltInAccessPolicy(viaAG, test.remoteIncluded, test.remoteExcluded), test.staleIncluded, test.staleExcluded),
					withStaleUsers(newBuiltInAccessPolicy(notViaAG, test.remoteIncluded, test.remoteExcluded), test.staleIncluded, test.staleExcluded),
				}
			}

			restrictedAccessUsers := restrictedAccessUsersObject(ctx, t, allowList, blockList)
			if test.stateUsersNull {
				usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
				if err != nil {
					t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
				}
				restrictedAccessUsers = types.ObjectNull(usersAttributes)
			}

			state := DeliveryGroupResourceModel{
				RestrictedAccessUsers: restrictedAccessUsers,
				DefaultAccessPolicies: defaultAccessPoliciesList(ctx, t, viaAG, notViaAG),
				CustomAccessPolicies:  types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)}),
			}
			if len(test.customPolicyNames) > 0 {
				state.CustomAccessPolicies = customAccessPoliciesList(ctx, t, test.customPolicyNames...)
			}

			var diags diag.Diagnostics
			result := state.updatePlanWithRestrictedAccessUsers(ctx, &diags, newDeliveryGroupWithPolicies(policies...), test.isRead)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}

			if test.expectUsersNull {
				if !result.RestrictedAccessUsers.IsNull() {
					t.Fatalf("expected restricted_access_users to be null, got %v", result.RestrictedAccessUsers)
				}
				return
			}

			if result.RestrictedAccessUsers.IsNull() {
				t.Fatalf("expected restricted_access_users to be set, got null")
			}

			users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, result.RestrictedAccessUsers)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics reading result: %s", diags)
			}

			if test.expectAllowNull {
				if !users.AllowList.IsNull() {
					t.Fatalf("expected allow_list to be null, got %v", users.AllowList)
				}
			} else {
				assertSameElements(t, "allow_list", stringSetValues(ctx, t, users.AllowList), test.expectAllowList)
			}

			if test.expectBlockNull {
				if !users.BlockList.IsNull() {
					t.Fatalf("expected block_list to be null, got %v", users.BlockList)
				}
			} else {
				assertSameElements(t, "block_list", stringSetValues(ctx, t, users.BlockList), test.expectBlockList)
			}
		})
	}
}

// TestAlignRestrictedAccessUsersWithPolicies covers where the users live after a read.
//
// Access policies that all restrict access to the same users are a delivery-group-wide
// restriction, so the users belong on restricted_access_users. Policies that differ are contextual
// access, so the users belong on the individual policies.
func TestAlignRestrictedAccessUsersWithPolicies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const viaAG = "ViaAG"
	const notViaAG = "NotViaAG"

	type testCase struct {
		stateUsersNull     bool
		stateAllowList     []string
		policies           []citrixorchestration.AdvancedAccessPolicyResponseModel
		expectUsersNull    bool
		expectAllowList    []string
		expectPolicyUsers  map[string][]string
		expectPoliciesNull bool
	}

	tests := map[string]testCase{
		"policies sharing users put them on the delivery group": {
			stateUsersNull: true,
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1"}, nil),
			},
			expectAllowList:    []string{"domain\\user1"},
			expectPoliciesNull: true,
		},
		"policies differing put the users on each policy": {
			stateUsersNull: true,
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user2"}, nil),
			},
			expectUsersNull: true,
			expectPolicyUsers: map[string][]string{
				viaAG:    {"domain\\user1"},
				notViaAG: {"domain\\user2"},
			},
		},
		"contextual access set out of band moves users off the delivery group": {
			stateAllowList: []string{"domain\\user1"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, nil),
				newBuiltInAccessPolicy(notViaAG, []string{"domain\\user1", "domain\\user2"}, nil),
			},
			expectUsersNull: true,
			expectPolicyUsers: map[string][]string{
				viaAG:    {"domain\\user1"},
				notViaAG: {"domain\\user1", "domain\\user2"},
			},
		},
		"no access on every policy stays on the delivery group": {
			stateAllowList: []string{"domain\\user1"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, nil, nil),
				newBuiltInAccessPolicy(notViaAG, nil, nil),
			},
			expectAllowList:    []string{},
			expectPoliciesNull: true,
		},
		"unrestricted policies null out the delivery group attribute": {
			stateAllowList: []string{"domain\\user1"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				unrestrictedAccessPolicy(viaAG),
				unrestrictedAccessPolicy(notViaAG),
			},
			expectUsersNull:    true,
			expectPoliciesNull: true,
		},
		"one restricted policy and one open policy is contextual": {
			stateAllowList: []string{"domain\\user1"},
			policies: []citrixorchestration.AdvancedAccessPolicyResponseModel{
				newBuiltInAccessPolicy(viaAG, []string{"domain\\user1"}, nil),
				unrestrictedAccessPolicy(notViaAG),
			},
			expectUsersNull: true,
			expectPolicyUsers: map[string][]string{
				viaAG: {"domain\\user1"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
			if err != nil {
				t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
			}

			restrictedAccessUsers := types.ObjectNull(usersAttributes)
			if !test.stateUsersNull {
				restrictedAccessUsers = restrictedAccessUsersObject(ctx, t, userSet(ctx, t, test.stateAllowList...), userSet(ctx, t))
			}

			state := DeliveryGroupResourceModel{
				RestrictedAccessUsers: restrictedAccessUsers,
				DefaultAccessPolicies: defaultAccessPoliciesList(ctx, t, viaAG, notViaAG),
				CustomAccessPolicies:  types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)}),
			}

			var diags diag.Diagnostics
			result := state.alignRestrictedAccessUsersWithPolicies(ctx, &diags, newDeliveryGroupWithPolicies(test.policies...))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}

			if test.expectUsersNull {
				if !result.RestrictedAccessUsers.IsNull() {
					t.Fatalf("expected restricted_access_users to be null, got %v", result.RestrictedAccessUsers)
				}
			} else {
				if result.RestrictedAccessUsers.IsNull() {
					t.Fatalf("expected restricted_access_users to be set, got null")
				}
				users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, result.RestrictedAccessUsers)
				assertSameElements(t, "allow_list", stringSetValues(ctx, t, users.AllowList), test.expectAllowList)
			}

			policies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &diags, result.DefaultAccessPolicies)
			for _, policy := range policies {
				connection := policy.AllowedConnection.ValueString()
				if test.expectPoliciesNull {
					if !policy.RestrictedAccessUsers.IsNull() {
						t.Fatalf("expected %s restricted_access_users to be null, got %v", connection, policy.RestrictedAccessUsers)
					}
					continue
				}

				expected, ok := test.expectPolicyUsers[connection]
				if !ok {
					if !policy.RestrictedAccessUsers.IsNull() {
						t.Fatalf("expected %s restricted_access_users to be null, got %v", connection, policy.RestrictedAccessUsers)
					}
					continue
				}

				if policy.RestrictedAccessUsers.IsNull() {
					t.Fatalf("expected %s restricted_access_users to be set, got null", connection)
				}
				users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, policy.RestrictedAccessUsers)
				assertSameElements(t, connection+" allow_list", stringSetValues(ctx, t, users.AllowList), expected)
			}
		})
	}
}

// TestAlignRestrictedAccessUsersWithSingleCustomPolicy is the regression test for a single
// restricted custom policy alongside unrestricted built-in policies.
//
// Comparing only the policies the configuration declares made this look delivery-group-wide, since
// one policy trivially agrees with itself. The read then moved the users to the delivery group
// level and cleared the custom policy, contradicting the configuration on every plan.
func TestAlignRestrictedAccessUsersWithSingleCustomPolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics

	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}

	state := DeliveryGroupResourceModel{
		RestrictedAccessUsers: types.ObjectNull(usersAttributes),
		DefaultAccessPolicies: types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)}),
		CustomAccessPolicies:  customAccessPoliciesList(ctx, t, "MyCustomPolicy"),
	}

	policies := []citrixorchestration.AdvancedAccessPolicyResponseModel{
		unrestrictedAccessPolicy("ViaAG"),
		unrestrictedAccessPolicy("NotViaAG"),
		newCustomAccessPolicy("MyCustomPolicy", []string{"domain\\user1"}),
	}

	result := state.alignRestrictedAccessUsersWithPolicies(ctx, &diags, newDeliveryGroupWithPolicies(policies...))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	if !result.RestrictedAccessUsers.IsNull() {
		t.Fatalf("expected delivery group restricted_access_users to stay null, got %v", result.RestrictedAccessUsers)
	}

	customPolicies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &diags, result.CustomAccessPolicies)
	if len(customPolicies) != 1 {
		t.Fatalf("expected 1 custom policy, got %d", len(customPolicies))
	}
	if customPolicies[0].RestrictedAccessUsers.IsNull() {
		t.Fatalf("expected the custom policy to keep its restricted_access_users, got null")
	}
	users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, customPolicies[0].RestrictedAccessUsers)
	assertSameElements(t, "custom policy allow_list", stringSetValues(ctx, t, users.AllowList), []string{"domain\\user1"})
}

// TestUpdatePlanWithRestrictedAccessUsersNoDeclaredPolicies covers a read where no access policy
// lists are declared. Falling back to the first remote policy hid an out-of-band restriction that
// landed on any later policy.
func TestUpdatePlanWithRestrictedAccessUsersNoDeclaredPolicies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics

	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}

	state := DeliveryGroupResourceModel{
		RestrictedAccessUsers: types.ObjectNull(usersAttributes),
		DefaultAccessPolicies: types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)}),
		CustomAccessPolicies:  types.ListNull(types.ObjectType{AttrTypes: accessPolicyAttributes(t)}),
	}

	policies := []citrixorchestration.AdvancedAccessPolicyResponseModel{
		unrestrictedAccessPolicy("ViaAG"),
		newBuiltInAccessPolicy("NotViaAG", []string{"domain\\user1"}, nil),
	}

	result := state.updatePlanWithRestrictedAccessUsers(ctx, &diags, newDeliveryGroupWithPolicies(policies...), true)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	if result.RestrictedAccessUsers.IsNull() {
		t.Fatalf("expected the out-of-band restriction on the second policy to surface, got null")
	}
	users := util.ObjectValueToTypedObject[RestrictedAccessUsers](ctx, &diags, result.RestrictedAccessUsers)
	assertSameElements(t, "allow_list", stringSetValues(ctx, t, users.AllowList), []string{"domain\\user1"})
}
