// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BranchRepeaterFilterModel struct {
	Id       types.String `tfsdk:"id"`
	PolicyId types.String `tfsdk:"policy_id"`
	Allowed  types.Bool   `tfsdk:"allowed"`
}

func (BranchRepeaterFilterModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: getPolicyFilterResourceDescription("Branch Repeater"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the branch repeater policy filter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Id of the policy to which the filter belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
		},
	}
}

func (BranchRepeaterFilterModel) GetAttributes() map[string]schema.Attribute {
	return BranchRepeaterFilterModel{}.GetSchema().Attributes
}

func (filter BranchRepeaterFilterModel) GetId() string {
	return filter.Id.ValueString()
}

func (filter BranchRepeaterFilterModel) GetPolicyId() string {
	return filter.PolicyId.ValueString()
}

func (filter BranchRepeaterFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	branchRepeaterFilterRequest := citrixorchestration.FilterRequest{}
	branchRepeaterFilterRequest.SetFilterType("BranchRepeater")
	branchRepeaterFilterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	branchRepeaterFilterRequest.SetIsEnabled(true)
	branchRepeaterFilterRequest.SetFilterData("")
	return branchRepeaterFilterRequest, nil
}

func (r BranchRepeaterFilterModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, filter citrixorchestration.FilterResponse) BranchRepeaterFilterModel {
	r.Id = types.StringValue(filter.GetFilterGuid())
	r.PolicyId = types.StringValue(filter.GetPolicyGuid())
	r.Allowed = types.BoolValue(filter.GetIsAllowed())

	return r
}
