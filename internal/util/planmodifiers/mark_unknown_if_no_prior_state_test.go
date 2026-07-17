// Copyright © 2026. Citrix Systems, Inc.

package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/util/planmodifiers"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMarkUnknownIfNoPriorStatePlanModifyString(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		request  planmodifier.StringRequest
		expected *planmodifier.StringResponse
	}{
		"both-null": {
			// new nested list element with no prior state — mark unknown so TF
			// accepts the real value returned by the API after apply
			request: planmodifier.StringRequest{
				StateValue: types.StringNull(),
				PlanValue:  types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"non-null-state": {
			// existing element already has a state value — leave the plan alone
			request: planmodifier.StringRequest{
				StateValue: types.StringValue("existing-id"),
				PlanValue:  types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringNull(),
			},
		},
		"known-plan": {
			// plan already has a concrete value — leave it unchanged
			request: planmodifier.StringRequest{
				StateValue: types.StringNull(),
				PlanValue:  types.StringValue("known"),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringValue("known"),
			},
		},
		"unknown-plan-non-null-state": {
			// plan is already unknown and state has a value — modifier should
			// not interfere (both-null condition not met)
			request: planmodifier.StringRequest{
				StateValue: types.StringValue("existing-id"),
				PlanValue:  types.StringUnknown(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown(),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &planmodifier.StringResponse{
				PlanValue: testCase.request.PlanValue,
			}
			planmodifiers.MarkUnknownIfNoPriorState().PlanModifyString(context.Background(), testCase.request, resp)
			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMarkUnknownIfNoPriorStateDescription(t *testing.T) {
	t.Parallel()
	m := planmodifiers.MarkUnknownIfNoPriorState()
	ctx := context.Background()
	if m.Description(ctx) == "" {
		t.Error("Description should not be empty")
	}
	if m.MarkdownDescription(ctx) != m.Description(ctx) {
		t.Error("MarkdownDescription should equal Description")
	}
}
