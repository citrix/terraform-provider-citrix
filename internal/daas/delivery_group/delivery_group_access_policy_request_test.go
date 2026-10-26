// Copyright © 2026. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"strings"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestGetAdvancedAccessPolicyRequestOmitsUnknownId covers creating a custom access policy.
//
// id is Computed and marked unknown when the access policy has no prior state, and an unknown
// value is not null. Guarding only on null sent an empty id for a brand new policy, which is
// rejected rather than treated as a create.
func TestGetAdvancedAccessPolicyRequestOmitsUnknownId(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}
	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}
	nullCriteria := types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes})

	newPolicy := func(id types.String) DeliveryGroupAccessPolicyModel {
		return DeliveryGroupAccessPolicyModel{
			Id:                     id,
			Name:                   types.StringValue("MyCustomPolicy"),
			Enabled:                types.BoolValue(true),
			AllowedConnection:      types.StringValue("ViaAG"),
			IncludeCriteriaFilters: nullCriteria,
			ExcludeCriteriaFilters: nullCriteria,
			RestrictedAccessUsers:  types.ObjectNull(usersAttributes),
		}
	}

	tests := map[string]struct {
		id          types.String
		expectIdSet bool
	}{
		"unknown id is omitted on create": {id: types.StringUnknown()},
		"null id is omitted":              {id: types.StringNull()},
		"known id is sent on update":      {id: types.StringValue("policy-guid-1"), expectIdSet: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			request, err := getAdvancedAccessPolicyRequest(ctx, &diags, newPolicy(test.id))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}

			body, err := request.MarshalJSON()
			if err != nil {
				t.Fatalf("failed to marshal request: %s", err)
			}

			if test.expectIdSet {
				if !strings.Contains(string(body), `"Id":"policy-guid-1"`) {
					t.Fatalf(`expected "Id":"policy-guid-1" in the request body, got %s`, body)
				}
				return
			}

			if strings.Contains(string(body), `"Id"`) {
				t.Fatalf("expected no Id in the request body, got %s", body)
			}
		})
	}
}

// TestApplyAccessPolicyUserFiltersPrefersPerPolicyUsers covers the delivery group create path.
//
// Create builds its access policy requests separately from update and applied the delivery group
// level simplified access policy to every policy unconditionally. A policy declaring its own
// restricted_access_users was therefore created with no users, and the refresh that followed read
// the empty policy back and contradicted the plan.
func TestApplyAccessPolicyUserFiltersPrefersPerPolicyUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	criteriaAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupAccessPolicyCriteriaTagsModel{})
	if err != nil {
		t.Fatalf("failed to build DeliveryGroupAccessPolicyCriteriaTagsModel attribute map: %s", err)
	}
	usersAttributes, err := util.ResourceAttributeMapFromObject(RestrictedAccessUsers{})
	if err != nil {
		t.Fatalf("failed to build RestrictedAccessUsers attribute map: %s", err)
	}
	nullCriteria := types.ListNull(types.ObjectType{AttrTypes: criteriaAttributes})

	// The delivery group level attribute is unset, so the simplified policy restricts nothing.
	var simpleAccessPolicy citrixorchestration.SimplifiedAccessPolicyRequestModel
	simpleAccessPolicy.SetIncludedUserFilterEnabled(false)
	simpleAccessPolicy.SetIncludedUsers([]string{})
	simpleAccessPolicy.SetExcludedUserFilterEnabled(false)
	simpleAccessPolicy.SetExcludedUsers([]string{})

	policy := DeliveryGroupAccessPolicyModel{
		Id:                     types.StringUnknown(),
		Name:                   types.StringValue("MyCustomPolicy"),
		Enabled:                types.BoolValue(true),
		AllowedConnection:      types.StringValue("ViaAG"),
		IncludeCriteriaFilters: nullCriteria,
		ExcludeCriteriaFilters: nullCriteria,
		RestrictedAccessUsers:  types.ObjectNull(usersAttributes),
	}

	t.Run("policy without its own users inherits the delivery group values", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics
		var request citrixorchestration.AdvancedAccessPolicyRequestModel
		if err := applyAccessPolicyUserFilters(ctx, &diags, nil, &request, policy, simpleAccessPolicy, false, citrixorchestration.ALLOWEDUSER_ANY_AUTHENTICATED); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags)
		}

		if request.GetIncludedUserFilterEnabled() {
			t.Fatalf("expected the include filter to stay disabled, got enabled")
		}
		if request.GetAllowedUsers() != citrixorchestration.ALLOWEDUSER_ANY_AUTHENTICATED {
			t.Fatalf("expected AllowedUsers to be inherited, got %v", request.GetAllowedUsers())
		}
	})
}
