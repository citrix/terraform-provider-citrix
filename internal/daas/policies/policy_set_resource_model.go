// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicySettingModel struct {
	Name       types.String `tfsdk:"name"`
	UseDefault types.Bool   `tfsdk:"use_default"`
	Value      types.String `tfsdk:"value"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (PolicySettingModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the policy setting.",
				Required:    true,
			},
			"use_default": schema.BoolAttribute{
				Description: "Indicate whether using default value for the policy setting.",
				Required:    true,
				Validators: []validator.Bool{
					validators.AlsoRequiresOneOfOnBoolValues([]bool{false}, path.MatchRelative().AtParent().AtName("value"), path.MatchRelative().AtParent().AtName("enabled")),
					validators.ConflictsWithOnBoolValues([]bool{true}, path.MatchRelative().AtParent().AtName("value"), path.MatchRelative().AtParent().AtName("enabled")),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the policy setting.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether of the policy setting has enabled or allowed value.",
				Optional:    true,
			},
		},
	}
}

func (PolicySettingModel) GetAttributes() map[string]schema.Attribute {
	return PolicySettingModel{}.GetSchema().Attributes
}

type PolicyFilterUuidDataClientModel struct {
	Server string `json:"server,omitempty"`
	Uuid   string `json:"uuid,omitempty"`
}

type PolicyFilterGatewayDataClientModel struct {
	Connection string `json:"connection,omitempty"`
	Condition  string `json:"condition,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
}

type PolicyModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	PolicySettings           types.Set    `tfsdk:"policy_settings"`             // Set[PolicySettingModel]
	AccessControlFilters     types.Set    `tfsdk:"access_control_filters"`      // Set[AccessControlFilterModel]
	BranchRepeaterFilter     types.Object `tfsdk:"branch_repeater_filter"`      // BranchRepeaterFilterModel
	ClientIPFilters          types.Set    `tfsdk:"client_ip_filters"`           // Set[ClientIPFilterModel]
	ClientNameFilters        types.Set    `tfsdk:"client_name_filters"`         // Set[ClientNameFilterModel]
	ClientPlatformFilters    types.Set    `tfsdk:"client_platform_filters"`     // Set[ClientPlatformFilterModel]
	DeliveryGroupFilters     types.Set    `tfsdk:"delivery_group_filters"`      // Set[DeliveryGroupFilterModel]
	DeliveryGroupTypeFilters types.Set    `tfsdk:"delivery_group_type_filters"` // Set[DeliveryGroupTypeFilterModel]
	OuFilters                types.Set    `tfsdk:"ou_filters"`                  // Set[OuFilterModel]
	UserFilters              types.Set    `tfsdk:"user_filters"`                // Set[UserFilterModel]
	TagFilters               types.Set    `tfsdk:"tag_filters"`                 // Set[TagFilterModel]
}

func (PolicyModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"policy_settings": schema.SetNestedAttribute{
				Description:  "Set of policy settings.",
				Required:     true,
				NestedObject: PolicySettingModel{}.GetSchema(),
			},
			"access_control_filters": schema.SetNestedAttribute{
				Description:  "Set of Access control policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: AccessControlFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"branch_repeater_filter": BranchRepeaterFilterModel{}.GetSchema(),
			"client_ip_filters": schema.SetNestedAttribute{
				Description:  "Set of Client ip policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: ClientIPFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"client_name_filters": schema.SetNestedAttribute{
				Description:  "Set of Client name policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: ClientNameFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"client_platform_filters": schema.SetNestedAttribute{
				Description:  "Set of Client platform policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: ClientPlatformFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"delivery_group_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery group policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: DeliveryGroupFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"delivery_group_type_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery group type policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: DeliveryGroupTypeFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"ou_filters": schema.SetNestedAttribute{
				Description:  "Set of Organizational unit policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: OuFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"user_filters": schema.SetNestedAttribute{
				Description:  "Set of User policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: UserFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"tag_filters": schema.SetNestedAttribute{
				Description:  "Set of Tag policy filters.",
				Optional:     true,
				Computed:     true,
				NestedObject: TagFilterModel{}.GetSchema(),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (PolicyModel) GetAttributes() map[string]schema.Attribute {
	return PolicyModel{}.GetSchema().Attributes
}

type PolicySetModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Description    types.String `tfsdk:"description"`
	Scopes         types.Set    `tfsdk:"scopes"` // Set[String]
	IsAssigned     types.Bool   `tfsdk:"assigned"`
	Policies       types.List   `tfsdk:"policies"`        // List[PolicyModel]
	DeliveryGroups types.Set    `tfsdk:"delivery_groups"` // Set[String]
}

func (PolicySetModel) GetSchema() schema.Schema {
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
			"type": schema.StringAttribute{
				Description: "Type of the policy set. Type can only be set to `DeliveryGroupPolicies`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("DeliveryGroupPolicies"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"DeliveryGroupPolicies"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"policies": schema.ListNestedAttribute{
				Description:  "Ordered list of policies. \n\n-> **Note** The order of policies in the list determines the priority of the policies.",
				Required:     true,
				NestedObject: PolicyModel{}.GetSchema(),
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

func (PolicySetModel) GetAttributes() map[string]schema.Attribute {
	return PolicySetModel{}.GetSchema().Attributes
}

func (r PolicySetModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, isResource bool, policySet *citrixorchestration.PolicySetResponse, policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse, policySetScopes []string, deliveryGroups []citrixorchestration.DeliveryGroupResponseModel) PolicySetModel {
	// Set required values
	r.Id = types.StringValue(policySet.GetPolicySetGuid())
	r.Name = types.StringValue(policySet.GetName())
	r.Type = types.StringValue(string(policySet.GetPolicySetType()))

	// Set optional values
	r.Description = types.StringValue(policySet.GetDescription())

	updatedPolicySetScopes := []string{}
	for _, scopeId := range policySetScopes {
		if !strings.EqualFold(scopeId, util.AllScopeId) && scopeId != "" {
			updatedPolicySetScopes = append(updatedPolicySetScopes, scopeId)
		}
	}

	r.Scopes = util.StringArrayToStringSet(ctx, diags, updatedPolicySetScopes)

	deliveryGroupIds := []string{}
	for _, deliveryGroup := range deliveryGroups {
		if strings.EqualFold(policySet.GetPolicySetGuid(), deliveryGroup.GetPolicySetGuid()) {
			deliveryGroupIds = append(deliveryGroupIds, deliveryGroup.GetId())
		}
	}
	r.DeliveryGroups = util.StringArrayToStringSet(ctx, diags, deliveryGroupIds)

	if policies != nil && policies.Items != nil {
		policyItems := policies.Items
		sort.Slice(policyItems, func(i, j int) bool {
			return policyItems[i].GetPriority() < policyItems[j].GetPriority()
		})
		refreshedPolicies := []PolicyModel{}
		for _, policy := range policyItems {
			policyModel := PolicyModel{
				Id:          types.StringValue(policy.GetPolicyGuid()),
				Name:        types.StringValue(policy.GetPolicyName()),
				Description: types.StringValue(policy.GetDescription()),
				Enabled:     types.BoolValue(policy.GetIsEnabled()),
			}

			refreshedPolicySettings := []PolicySettingModel{}
			if policy.GetSettings() != nil && len(policy.GetSettings()) != 0 {
				for _, setting := range policy.GetSettings() {
					policySetting := PolicySettingModel{
						Name:       types.StringValue(setting.GetSettingName()),
						UseDefault: types.BoolValue(setting.GetUseDefault()),
					}
					if !setting.GetUseDefault() {
						settingValue := types.StringValue(setting.GetSettingValue())
						if strings.EqualFold(setting.GetSettingValue(), "true") ||
							setting.GetSettingValue() == "1" {
							policySetting.Enabled = types.BoolValue(true)
							policySetting.Value = types.StringNull()
						} else if strings.EqualFold(setting.GetSettingValue(), "false") ||
							setting.GetSettingValue() == "0" {
							policySetting.Enabled = types.BoolValue(false)
							policySetting.Value = types.StringNull()
						} else {
							policySetting.Enabled = types.BoolNull()
							policySetting.Value = settingValue
						}
					}

					refreshedPolicySettings = append(refreshedPolicySettings, policySetting)
				}
			}

			if isResource {
				policyModel.PolicySettings = util.TypedArrayToObjectSet(ctx, diags, refreshedPolicySettings)
			} else {
				policyModel.PolicySettings = util.DataSourceTypedArrayToObjectSet(ctx, diags, refreshedPolicySettings)
			}

			if isResource {
				attributes, _ := util.ResourceAttributeMapFromObject(BranchRepeaterFilterModel{})
				policyModel.BranchRepeaterFilter = types.ObjectNull(attributes)
			} else {
				attributes, _ := util.DataSourceAttributeMapFromObject(BranchRepeaterFilterModel{})
				policyModel.BranchRepeaterFilter = types.ObjectNull(attributes)
			}

			var accessControlFilters, branchRepeaterFilters, clientIpFilters, clientNameFilters, clientPlatformFilters, desktopGroupFilters, desktopKindFilters, desktopTagFilters, ouFilters, userFilters = ParsePolicyFilters(ctx, diags, policy)
			if isResource {
				policyModel.AccessControlFilters = util.TypedArrayToObjectSet(ctx, diags, accessControlFilters)
				if len(branchRepeaterFilters) > 0 {
					policyModel.BranchRepeaterFilter = util.TypedObjectToObjectValue(ctx, diags, branchRepeaterFilters[0])
				}
				policyModel.ClientIPFilters = util.TypedArrayToObjectSet(ctx, diags, clientIpFilters)
				policyModel.ClientNameFilters = util.TypedArrayToObjectSet(ctx, diags, clientNameFilters)
				policyModel.ClientPlatformFilters = util.TypedArrayToObjectSet(ctx, diags, clientPlatformFilters)
				policyModel.DeliveryGroupFilters = util.TypedArrayToObjectSet(ctx, diags, desktopGroupFilters)
				policyModel.DeliveryGroupTypeFilters = util.TypedArrayToObjectSet(ctx, diags, desktopKindFilters)
				policyModel.TagFilters = util.TypedArrayToObjectSet(ctx, diags, desktopTagFilters)
				policyModel.OuFilters = util.TypedArrayToObjectSet(ctx, diags, ouFilters)
				policyModel.UserFilters = util.TypedArrayToObjectSet(ctx, diags, userFilters)
			} else {
				policyModel.AccessControlFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, accessControlFilters)
				if len(branchRepeaterFilters) > 0 {
					policyModel.BranchRepeaterFilter = util.DataSourceTypedObjectToObjectValue(ctx, diags, branchRepeaterFilters[0])
				}
				policyModel.ClientIPFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, clientIpFilters)
				policyModel.ClientNameFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, clientNameFilters)
				policyModel.DeliveryGroupFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopGroupFilters)
				policyModel.DeliveryGroupTypeFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopKindFilters)
				policyModel.TagFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopTagFilters)
				policyModel.OuFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, ouFilters)
				policyModel.UserFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, userFilters)
			}

			refreshedPolicies = append(refreshedPolicies, policyModel)
		}
		var updatedPolicies types.List
		if isResource {
			updatedPolicies = util.TypedArrayToObjectList(ctx, diags, refreshedPolicies)
		} else {
			updatedPolicies = util.DataSourceTypedArrayToObjectList(ctx, diags, refreshedPolicies)
		}
		r.Policies = updatedPolicies
	} else {
		var attributesMap map[string]attr.Type
		var err error
		if isResource {
			attributesMap, err = util.ResourceAttributeMapFromObject(PolicyModel{})
			if err != nil {
				diags.AddError("Error converting schema to attribute map. Error: ", err.Error())
			}
		} else {
			attributesMap, err = util.DataSourceAttributeMapFromObject(PolicyModel{})
			if err != nil {
				diags.AddError("Error converting schema to attribute map. Error: ", err.Error())
			}
		}
		r.Policies = types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	}

	r.IsAssigned = types.BoolValue(policySet.GetIsAssigned())

	return r
}

func ParsePolicyFilters(ctx context.Context, diags *diag.Diagnostics, policy citrixorchestration.PolicyResponse) ([]AccessControlFilterModel, []BranchRepeaterFilterModel, []ClientIPFilterModel, []ClientNameFilterModel, []ClientPlatformFilterModel, []DeliveryGroupFilterModel, []DeliveryGroupTypeFilterModel, []TagFilterModel, []OuFilterModel, []UserFilterModel) {
	var accessControlFilters []AccessControlFilterModel
	var branchRepeaterFilters []BranchRepeaterFilterModel
	var clientIpFilters []ClientIPFilterModel
	var clientNameFilters []ClientNameFilterModel
	var clientPlatformFilters []ClientPlatformFilterModel
	var desktopGroupFilters []DeliveryGroupFilterModel
	var desktopKindFilters []DeliveryGroupTypeFilterModel
	var desktopTagFilters []TagFilterModel
	var ouFilters []OuFilterModel
	var userFilters []UserFilterModel
	if policy.GetFilters() != nil && len(policy.GetFilters()) != 0 {
		for _, filter := range policy.GetFilters() {

			var uuidFilterData PolicyFilterUuidDataClientModel
			_ = json.Unmarshal([]byte(filter.GetFilterData()), &uuidFilterData)

			var gatewayFilterData PolicyFilterGatewayDataClientModel
			_ = json.Unmarshal([]byte(filter.GetFilterData()), &gatewayFilterData)

			filterType := filter.GetFilterType()
			switch filterType {
			case "AccessControl":
				accessControlFilters = append(accessControlFilters, AccessControlFilterModel{
					Allowed:    types.BoolValue(filter.GetIsAllowed()),
					Enabled:    types.BoolValue(filter.GetIsEnabled()),
					Connection: types.StringValue(gatewayFilterData.Connection),
					Condition:  types.StringValue(gatewayFilterData.Condition),
					Gateway:    types.StringValue(gatewayFilterData.Gateway),
				})
			case "BranchRepeater":
				branchRepeaterFilters = append(branchRepeaterFilters, BranchRepeaterFilterModel{
					Allowed: types.BoolValue(filter.GetIsAllowed()),
				})
			case "ClientIP":
				clientIpFilters = append(clientIpFilters, ClientIPFilterModel{
					Allowed:   types.BoolValue(filter.GetIsAllowed()),
					Enabled:   types.BoolValue(filter.GetIsEnabled()),
					IpAddress: types.StringValue(filter.GetFilterData()),
				})
			case "ClientName":
				clientNameFilters = append(clientNameFilters, ClientNameFilterModel{
					Allowed:    types.BoolValue(filter.GetIsAllowed()),
					Enabled:    types.BoolValue(filter.GetIsEnabled()),
					ClientName: types.StringValue(filter.GetFilterData()),
				})
			case "ClientPlatform":
				clientPlatformFilters = append(clientPlatformFilters, ClientPlatformFilterModel{
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
					Enabled:  types.BoolValue(filter.GetIsEnabled()),
					Platform: types.StringValue(filter.GetFilterData()),
				})
			case "DesktopGroup":
				desktopGroupFilters = append(desktopGroupFilters, DeliveryGroupFilterModel{
					Allowed:         types.BoolValue(filter.GetIsAllowed()),
					Enabled:         types.BoolValue(filter.GetIsEnabled()),
					DeliveryGroupId: types.StringValue(uuidFilterData.Uuid),
				})
			case "DesktopKind":
				desktopKindFilters = append(desktopKindFilters, DeliveryGroupTypeFilterModel{
					Allowed:           types.BoolValue(filter.GetIsAllowed()),
					Enabled:           types.BoolValue(filter.GetIsEnabled()),
					DeliveryGroupType: types.StringValue(filter.GetFilterData()),
				})
			case "DesktopTag":
				desktopTagFilters = append(desktopTagFilters, TagFilterModel{
					Allowed: types.BoolValue(filter.GetIsAllowed()),
					Enabled: types.BoolValue(filter.GetIsEnabled()),
					Tag:     types.StringValue(uuidFilterData.Uuid),
				})
			case "OU":
				ouFilters = append(ouFilters, OuFilterModel{
					Allowed: types.BoolValue(filter.GetIsAllowed()),
					Enabled: types.BoolValue(filter.GetIsEnabled()),
					Ou:      types.StringValue(filter.GetFilterData()),
				})
			case "User":
				userFilters = append(userFilters, UserFilterModel{
					Allowed: types.BoolValue(filter.GetIsAllowed()),
					Enabled: types.BoolValue(filter.GetIsEnabled()),
					UserSid: types.StringValue(filter.GetFilterData()),
				})
			}
		}
	}
	return accessControlFilters, branchRepeaterFilters, clientIpFilters, clientNameFilters, clientPlatformFilters, desktopGroupFilters, desktopKindFilters, desktopTagFilters, ouFilters, userFilters
}
