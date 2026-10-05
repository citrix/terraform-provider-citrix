// Copyright © 2026. Citrix Systems, Inc.

package policy_priority

import (
	"context"
	"regexp"

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
				Description: "Ordered list of policy IDs. Each policy may be listed only once. \n\n-> **Note** The order of policy IDs in the list determines the priority of the policies." +
					"\n\n-> **Note** The list does not have to name every policy in the policy set. Policies in the set that are not listed keep their relative order and are given lower priority than every listed policy." +
					"\n\n~> **Warning** A policy added to the policy set outside Terraform, for example in Studio, is moved below the listed policies the next time this resource is applied. Such a policy does not itself produce a plan, so if it was given a higher priority than the listed policies, `terraform plan` reports no changes while it continues to outrank them. The order is only corrected once some other change causes this resource to be applied.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					),
					// Only catches literal repeats; unknown ids are deduped by dedupeFold.
					listvalidator.UniqueValues(),
				},
			},
			"policy_names": schema.ListAttribute{
				Description: "Ordered list of policy names. \n\n-> **Note** The order of policy names in the list reflects the priority of the policies." +
					"\n\n-> **Note** Only the policies named in `policy_priority` appear here. Policies in the policy set that are not listed are omitted, except immediately after `terraform import`, which adopts every policy in the set.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (PolicyPriorityModel) GetAttributes() map[string]schema.Attribute {
	return PolicyPriorityModel{}.GetSchema().Attributes
}

func (PolicyPriorityModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

// RefreshPropertyValues maps the policy set onto the model. Read sets orderFromRemote to surface drift.
func (r PolicyPriorityModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, policySet *citrixorchestration.PolicySetResponse, policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse, orderFromRemote bool) PolicyPriorityModel {
	r.PolicySetId = types.StringValue(policySet.GetPolicySetGuid())
	r.PolicySetName = types.StringValue(policySet.GetName())

	policyIds, policyNames := policyIdsAndNamesByPriority(policies)

	// Narrow to the managed subset so an unlisted policy cannot manufacture a diff; import adopts all.
	adoptAll := r.PolicyPriority.IsNull() || r.PolicyPriority.IsUnknown()
	var managed []string
	if !adoptAll {
		managed = util.StringListToStringArray(ctx, diagnostics, r.PolicyPriority)
	}
	policyIds, policyNames = resolveManagedPolicies(managed, adoptAll, policyIds, policyNames, orderFromRemote)

	r.PolicyPriority = util.StringArrayToStringList(ctx, diagnostics, policyIds)
	r.PolicyNames = util.StringArrayToStringList(ctx, diagnostics, policyNames)

	return r
}
