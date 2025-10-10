package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func UseStateIfUnknownAndStateValueExists() planmodifier.String {
	return useStateIfUnknownAndStateValueExists{}
}

// useStateForUnknownModifier implements the plan modifier.
type useStateIfUnknownAndStateValueExists struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateIfUnknownAndStateValueExists) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateIfUnknownAndStateValueExists) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyString implements the plan modification logic.
func (m useStateIfUnknownAndStateValueExists) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state (resource is being created).
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
