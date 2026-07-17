// Copyright © 2026. Citrix Systems, Inc.

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MarkUnknownIfNoPriorState() planmodifier.String {
	return markUnknownIfNoPriorState{}
}

type markUnknownIfNoPriorState struct{}

func (m markUnknownIfNoPriorState) Description(_ context.Context) string {
	return "Sets value to (known after apply) when the attribute has no prior state, such as when a new nested list element is added."
}

func (m markUnknownIfNoPriorState) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString marks the planned value as unknown when both plan and state are null.
// TF does not automatically mark Computed attributes as unknown for new nested list elements
// (elements with no prior state entry), so without this modifier the provider produces an
// inconsistency error when the API returns a real value after apply.
func (m markUnknownIfNoPriorState) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() && req.StateValue.IsNull() {
		resp.PlanValue = types.StringUnknown()
	}
}
