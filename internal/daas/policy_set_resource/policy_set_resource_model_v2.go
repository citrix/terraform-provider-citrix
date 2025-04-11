// Copyright © 2024. Citrix Systems, Inc.

package policy_set_resource

import (
	"context"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicySetV2Model struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Scopes         types.Set    `tfsdk:"scopes"` // Set[String]
	IsAssigned     types.Bool   `tfsdk:"assigned"`
	DeliveryGroups types.Set    `tfsdk:"delivery_groups"` // Set[String]
}

func (PolicySetV2Model) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy set.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the policy set to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"assigned": schema.BoolAttribute{
				Description: "Indicate whether the policy set is being assigned to delivery groups.",
				Computed:    true,
			},
			"delivery_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the delivery groups for the policy set to apply on." +
					"\n\n~> **Please Note** If `delivery_groups` attribute is unset or configured as an empty set, the policy set will not be assigned to any delivery group. None of the policies in the policy set will be applied.",
				Optional: true,
				Computed: true,
				Default:  setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
		},
	}
}

func (PolicySetV2Model) GetAttributes() map[string]schema.Attribute {
	return PolicySetV2Model{}.GetSchema().Attributes
}

func (r PolicySetV2Model) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isDefaultPolicySet bool, policySet *citrixorchestration.PolicySetResponse, policySetScopes []string, deliveryGroups []citrixorchestration.DeliveryGroupResponseModel) PolicySetV2Model {
	// Set required values
	r.Id = types.StringValue(policySet.GetPolicySetGuid())
	r.Name = types.StringValue(policySet.GetName())

	// Set optional values
	r.Description = types.StringValue(policySet.GetDescription())

	updatedPolicySetScopes := []string{}
	for _, scopeId := range policySetScopes {
		if !strings.EqualFold(scopeId, util.AllScopeId) && scopeId != "" {
			updatedPolicySetScopes = append(updatedPolicySetScopes, scopeId)
		}
	}

	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, updatedPolicySetScopes)

	deliveryGroupIds := []string{}
	if !isDefaultPolicySet {
		for _, deliveryGroup := range deliveryGroups {
			if strings.EqualFold(policySet.GetPolicySetGuid(), deliveryGroup.GetPolicySetGuid()) {
				deliveryGroupIds = append(deliveryGroupIds, deliveryGroup.GetId())
			}
		}
	}
	r.DeliveryGroups = util.StringArrayToStringSet(ctx, diagnostics, deliveryGroupIds)

	r.IsAssigned = types.BoolValue(policySet.GetIsAssigned())

	return r
}
