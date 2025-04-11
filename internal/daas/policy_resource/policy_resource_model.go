// Copyright © 2024. Citrix Systems, Inc.

package policy_resource

import (
	"context"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyModel struct {
	Id          types.String `tfsdk:"id"`
	PolicySetId types.String `tfsdk:"policy_set_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func (PolicyModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an instance of the Policy Setting." +
			"\n\n~> **Please Note** Each policy can only associate with one policy set. The policy will be created in the default policy set if the policy is not referenced in any of the `policy_ids` of the `citrix_policy_set_v2` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "Id of the policy set the policy belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the policy is being enabled.",
				Required:    true,
			},
		},
	}
}

func (PolicyModel) GetAttributes() map[string]schema.Attribute {
	return PolicyModel{}.GetSchema().Attributes
}

func (r PolicyModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, policy *citrixorchestration.PolicyResponse) PolicyModel {
	r.Id = types.StringValue(policy.GetPolicyGuid())
	r.PolicySetId = types.StringValue(policy.GetPolicySetGuid())
	r.Name = types.StringValue(policy.GetPolicyName())
	r.Description = types.StringValue(policy.GetDescription())
	r.Enabled = types.BoolValue(policy.GetIsEnabled())

	return r
}
