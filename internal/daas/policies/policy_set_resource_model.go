// Copyright © 2023. Citrix Systems, Inc.

package policies

import (
	"sort"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicySettingModel struct {
	Name       types.String `tfsdk:"name"`
	UseDefault types.Bool   `tfsdk:"use_default"`
	Value      types.String `tfsdk:"value"`
}

type PolicyFilterModel struct {
	Type      types.String `tfsdk:"type"`
	Data      types.String `tfsdk:"data"`
	IsAllowed types.Bool   `tfsdk:"is_allowed"`
	IsEnabled types.Bool   `tfsdk:"is_enabled"`
}

type PolicyModel struct {
	Name           types.String         `tfsdk:"name"`
	Description    types.String         `tfsdk:"description"`
	IsEnabled      types.Bool           `tfsdk:"is_enabled"`
	PolicySettings []PolicySettingModel `tfsdk:"policy_settings"`
	PolicyFilters  []PolicyFilterModel  `tfsdk:"policy_filters"`
}

type PolicySetResourceModel struct {
	Id          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Type        types.String   `tfsdk:"type"`
	Description types.String   `tfsdk:"description"`
	Scopes      []types.String `tfsdk:"scopes"`
	IsAssigned  types.Bool     `tfsdk:"is_assigned"`
	Policies    []PolicyModel  `tfsdk:"policies"`
}

func (r PolicySetResourceModel) RefreshPropertyValues(policySet *citrixorchestration.PolicySetResponse, policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse) PolicySetResourceModel {
	// Set required values
	r.Id = types.StringValue(policySet.GetPolicySetGuid())
	r.Name = types.StringValue(policySet.GetName())
	r.Type = types.StringValue(string(policySet.GetPolicySetType()))

	// Set optional values
	if policySet.GetDescription() != "" {
		r.Description = types.StringValue(policySet.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	if policySet.GetScopes() != nil {
		r.Scopes = util.ConvertPrimitiveStringArrayToBaseStringArray(policySet.GetScopes())
	} else {
		r.Scopes = nil
	}

	if policies != nil && policies.Items != nil {
		policyItems := policies.Items
		sort.Slice(policyItems, func(i, j int) bool {
			return policyItems[i].GetPriority() < policyItems[j].GetPriority()
		})
		r.Policies = []PolicyModel{}
		for _, policy := range policyItems {
			policyModel := PolicyModel{
				Name:        types.StringValue(policy.GetPolicyName()),
				Description: types.StringValue(policy.GetDescription()),
				IsEnabled:   types.BoolValue(policy.GetIsEnabled()),
			}

			policyModel.PolicySettings = []PolicySettingModel{}
			if policy.GetSettings() != nil && len(policy.GetSettings()) != 0 {
				for _, setting := range policy.GetSettings() {
					policyModel.PolicySettings = append(policyModel.PolicySettings, PolicySettingModel{
						Name:       types.StringValue(setting.GetSettingName()),
						UseDefault: types.BoolValue(setting.GetUseDefault()),
						Value:      types.StringValue(setting.GetSettingValue()),
					})
				}
			}

			policyModel.PolicyFilters = []PolicyFilterModel{}
			if policy.GetFilters() != nil && len(policy.GetFilters()) != 0 {
				for _, filter := range policy.GetFilters() {
					policyModel.PolicyFilters = append(policyModel.PolicyFilters, PolicyFilterModel{
						Type:      types.StringValue(filter.GetFilterType()),
						IsAllowed: types.BoolValue(filter.GetIsAllowed()),
						IsEnabled: types.BoolValue(filter.GetIsEnabled()),
						Data:      types.StringValue(filter.GetFilterData()),
					})
				}
			}

			r.Policies = append(r.Policies, policyModel)
		}
	}

	r.IsAssigned = types.BoolValue(policySet.GetIsAssigned())

	return r
}
