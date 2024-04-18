// Copyright © 2023. Citrix Systems, Inc.

package policies

import (
	"encoding/json"
	"sort"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicySettingModel struct {
	Name       types.String `tfsdk:"name"`
	UseDefault types.Bool   `tfsdk:"use_default"`
	Value      types.String `tfsdk:"value"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

type PolicyFilterModel struct {
	Type      types.String          `tfsdk:"type"`
	Data      PolicyFilterDataModel `tfsdk:"data"`
	IsAllowed types.Bool            `tfsdk:"allowed"`
	IsEnabled types.Bool            `tfsdk:"enabled"`
}

type PolicyFilterDataModel struct {
	Server     types.String `tfsdk:"server"`
	Uuid       types.String `tfsdk:"uuid"`
	Connection types.String `tfsdk:"connection"`
	Condition  types.String `tfsdk:"condition"`
	Gateway    types.String `tfsdk:"gateway"`
	Value      types.String `tfsdk:"value"`
}

type PolicyFilterDataClientModel struct {
	Server     string `json:"server,omitempty"`
	Uuid       string `json:"uuid,omitempty"`
	Connection string `json:"Connection,omitempty"`
	Condition  string `json:"Condition,omitempty"`
	Gateway    string `json:"Gateway,omitempty"`
}

type PolicyModel struct {
	Name           types.String         `tfsdk:"name"`
	Description    types.String         `tfsdk:"description"`
	IsEnabled      types.Bool           `tfsdk:"enabled"`
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

	if policySet.GetScopes() != nil && len(policySet.GetScopes()) != 0 {
		r.Scopes = util.ConvertPrimitiveStringArrayToBaseStringArray(policySet.GetScopes())

		scopeContainsAll := false

		for _, scope := range r.Scopes {
			if strings.EqualFold(scope.ValueString(), "all") {
				scopeContainsAll = true
				break
			}
		}

		if !scopeContainsAll {
			r.Scopes = append(r.Scopes, types.StringValue("All"))
		}

		sort.Slice(r.Scopes, func(i, j int) bool {
			return r.Scopes[i].ValueString() < r.Scopes[j].ValueString()
		})
	} else {
		r.Scopes = []types.String{types.StringValue("All")}
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
					policySetting := PolicySettingModel{
						Name:       types.StringValue(setting.GetSettingName()),
						UseDefault: types.BoolValue(setting.GetUseDefault()),
					}

					settingValue := types.StringValue(setting.GetSettingValue())
					if strings.EqualFold(setting.GetSettingValue(), "true") ||
						setting.GetSettingValue() == "1" {
						policySetting.Enabled = types.BoolValue(true)
					} else if strings.EqualFold(setting.GetSettingValue(), "false") ||
						setting.GetSettingValue() == "0" {
						policySetting.Enabled = types.BoolValue(false)
					} else {
						policySetting.Value = settingValue
					}

					policyModel.PolicySettings = append(policyModel.PolicySettings, policySetting)
				}
			}

			sort.Slice(policyModel.PolicySettings, func(i, j int) bool {
				return policyModel.PolicySettings[i].Name.ValueString() < policyModel.PolicySettings[j].Name.ValueString()
			})

			policyModel.PolicyFilters = []PolicyFilterModel{}
			if policy.GetFilters() != nil && len(policy.GetFilters()) != 0 {
				for _, filter := range policy.GetFilters() {
					var filterData PolicyFilterDataClientModel
					filterDataModel := PolicyFilterDataModel{}
					err := json.Unmarshal([]byte(filter.GetFilterData()), &filterData)
					if err != nil {
						filterDataModel.Value = types.StringValue(filter.GetFilterData())
					} else {
						if filterData.Server != "" {
							filterDataModel.Server = types.StringValue(filterData.Server)
						}
						if filterData.Uuid != "" {
							filterDataModel.Uuid = types.StringValue(filterData.Uuid)
						}
						if filterData.Connection != "" {
							filterDataModel.Connection = types.StringValue(filterData.Connection)
						}
						if filterData.Condition != "" {
							filterDataModel.Condition = types.StringValue(filterData.Condition)
						}
						if filterData.Gateway != "" {
							filterDataModel.Gateway = types.StringValue(filterData.Gateway)
						}
					}

					policyModel.PolicyFilters = append(policyModel.PolicyFilters, PolicyFilterModel{
						Type:      types.StringValue(filter.GetFilterType()),
						IsAllowed: types.BoolValue(filter.GetIsAllowed()),
						IsEnabled: types.BoolValue(filter.GetIsEnabled()),
						Data:      filterDataModel,
					})
				}
			}

			sort.Slice(policyModel.PolicyFilters, func(i, j int) bool {
				return policyModel.PolicyFilters[i].Type.ValueString() < policyModel.PolicyFilters[j].Type.ValueString()
			})

			r.Policies = append(r.Policies, policyModel)
		}
	}

	r.IsAssigned = types.BoolValue(policySet.GetIsAssigned())

	return r
}
