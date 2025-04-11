// Copyright © 2024. Citrix Systems, Inc.

package policy_priority

import (
	"context"
	"regexp"
	"sort"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyPriorityModel struct {
	PolicySetId    types.String `tfsdk:"policy_set_id"`
	PolicySetName  types.String `tfsdk:"policy_set_name"`
	PolicyPriority types.List   `tfsdk:"policy_priority"` // List[String]
	PolicyNames    types.List   `tfsdk:"policy_names"`    // List[String]
}

func (PolicyPriorityModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages  the policy priorities within a policy set.",
		Attributes: map[string]schema.Attribute{
			"policy_set_id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_set_name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Computed:    true,
			},
			"policy_priority": schema.ListAttribute{
				Description: "Ordered list of policy IDs. \n\n-> **Note** The order of policy IDs in the list determines the priority of the policies.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					),
				},
			},
			"policy_names": schema.ListAttribute{
				Description: "Ordered list of policy names. \n\n-> **Note** The order of policy names in the list reflects the priority of the policies.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (PolicyPriorityModel) GetAttributes() map[string]schema.Attribute {
	return PolicyPriorityModel{}.GetSchema().Attributes
}

func (r PolicyPriorityModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, policySet *citrixorchestration.PolicySetResponse, policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse) PolicyPriorityModel {
	r.PolicySetId = types.StringValue(policySet.GetPolicySetGuid())
	r.PolicySetName = types.StringValue(policySet.GetName())

	policyIds := []string{}
	policyNames := []string{}
	if policies != nil && policies.Items != nil {
		policyItems := policies.Items
		sort.Slice(policyItems, func(i, j int) bool {
			return policyItems[i].GetPriority() < policyItems[j].GetPriority()
		})
		for _, policy := range policyItems {
			policyIds = append(policyIds, policy.GetPolicyGuid())
			policyNames = append(policyNames, policy.GetPolicyName())
		}

	} else {
		r.PolicyPriority = util.StringArrayToStringList(ctx, diagnostics, []string{})
		r.PolicyNames = util.StringArrayToStringList(ctx, diagnostics, []string{})
	}
	r.PolicyPriority = util.StringArrayToStringList(ctx, diagnostics, policyIds)
	r.PolicyNames = util.StringArrayToStringList(ctx, diagnostics, policyNames)

	return r
}
