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
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
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
				Description: "Name of the policy setting name.",
				Required:    true,
			},
			"use_default": schema.BoolAttribute{
				Description: "Indicate whether using default value for the policy setting.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value of the policy setting.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("enabled"),
					),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether of the policy setting has enabled or allowed value.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Bool{
					boolvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("value")),
				},
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

type AccessControlFilterModel struct {
	Allowed    types.Bool `tfsdk:"allowed"`
	Enabled    types.Bool `tfsdk:"enabled"`
	Connection string     `json:"Connection"`
	Condition  string     `json:"Condition"`
	Gateway    string     `json:"Gateway"`
}

func (AccessControlFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"connection": schema.StringAttribute{
				Description: "Gateway connection for the policy filter.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"WithAccessGateway",
						"WithoutAccessGateway"}...),
				},
			},
			"condition": schema.StringAttribute{
				Description: "Gateway condition for the policy filter.",
				Required:    true,
			},
			"gateway": schema.StringAttribute{
				Description: "Gateway for the policy filter.",
				Required:    true,
			},
		},
	}
}

func (AccessControlFilterModel) GetAttributes() map[string]schema.Attribute {
	return AccessControlFilterModel{}.GetSchema().Attributes
}

type BranchRepeaterFilterModel struct {
	Allowed types.Bool `tfsdk:"allowed"`
	Enabled types.Bool `tfsdk:"enabled"`
}

func (BranchRepeaterFilterModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Set of policy filters.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
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

type ClientIPFilterModel struct {
	Allowed   types.Bool   `tfsdk:"allowed"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	IpAddress types.String `tfsdk:"ip_address"`
}

func (ClientIPFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "IP Address of the client to be filtered.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.IPv4Regex), "must be a valid IPv4 address without protocol (http:// or https://) and port number"),
				},
			},
		},
	}
}

func (ClientIPFilterModel) GetAttributes() map[string]schema.Attribute {
	return ClientIPFilterModel{}.GetSchema().Attributes
}

type ClientNameFilterModel struct {
	Allowed    types.Bool   `tfsdk:"allowed"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	ClientName types.String `tfsdk:"client_name"`
}

func (ClientNameFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"client_name": schema.StringAttribute{
				Description: "Name of the client to be filtered.",
				Required:    true,
			},
		},
	}
}

func (ClientNameFilterModel) GetAttributes() map[string]schema.Attribute {
	return ClientNameFilterModel{}.GetSchema().Attributes
}

type DeliveryGroupFilterModel struct {
	Allowed         types.Bool   `tfsdk:"allowed"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	DeliveryGroupId types.String `tfsdk:"delivery_group_id"`
}

func (DeliveryGroupFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"delivery_group_id": schema.StringAttribute{
				Description: "Id of the delivery group to be filtered.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
		},
	}
}

func (DeliveryGroupFilterModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupFilterModel{}.GetSchema().Attributes
}

type DeliveryGroupTypeFilterModel struct {
	Allowed           types.Bool   `tfsdk:"allowed"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	DeliveryGroupType types.String `tfsdk:"delivery_group_type"`
}

func (DeliveryGroupTypeFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"delivery_group_type": schema.StringAttribute{
				Description: "Type of the delivery groups to be filtered.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"Private",
						"PrivateApp",
						"Shared",
						"SharedApp"}...),
				},
			},
		},
	}
}

func (DeliveryGroupTypeFilterModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupTypeFilterModel{}.GetSchema().Attributes
}

type OuFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Ou      types.String `tfsdk:"ou"`
}

func (OuFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"ou": schema.StringAttribute{
				Description: "Organizational Unit to be filtered.",
				Required:    true,
			},
		},
	}
}

func (OuFilterModel) GetAttributes() map[string]schema.Attribute {
	return OuFilterModel{}.GetSchema().Attributes
}

type UserFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	UserSid types.String `tfsdk:"sid"`
}

func (UserFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"sid": schema.StringAttribute{
				Description: "SID of the user or user group to be filtered.",
				Required:    true,
			},
		},
	}
}

func (UserFilterModel) GetAttributes() map[string]schema.Attribute {
	return UserFilterModel{}.GetSchema().Attributes
}

type TagFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Tag     types.String `tfsdk:"tag"`
}

func (TagFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag to be filtered.",
				Required:    true,
			},
		},
	}
}

func (TagFilterModel) GetAttributes() map[string]schema.Attribute {
	return TagFilterModel{}.GetSchema().Attributes
}

type PolicyModel struct {
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	PolicySettings           types.Set    `tfsdk:"policy_settings"`             // Set[PolicySettingModel]
	AccessControlFilters     types.Set    `tfsdk:"access_control_filters"`      // Set[AccessControlFilterModel]
	BranchRepeaterFilter     types.Object `tfsdk:"branch_repeater_filter"`      // BranchRepeaterFilterModel
	ClientIPFilters          types.Set    `tfsdk:"client_ip_filters"`           // Set[ClientIPFilterModel]
	ClientNameFilters        types.Set    `tfsdk:"client_name_filters"`         // Set[ClientNameFilterModel]
	DeliveryGroupFilters     types.Set    `tfsdk:"delivery_group_filters"`      // Set[DeliveryGroupFilterModel]
	DeliveryGroupTypeFilters types.Set    `tfsdk:"delivery_group_type_filters"` // Set[DeliveryGroupTypeFilterModel]
	OuFilters                types.Set    `tfsdk:"ou_filters"`                  // Set[OuFilterModel]
	UserFilters              types.Set    `tfsdk:"user_filters"`                // Set[UserFilterModel]
	TagFilters               types.Set    `tfsdk:"tag_filters"`                 // Set[TagFilterModel]
}

func (PolicyModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
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

type PolicySetResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Scopes      types.Set    `tfsdk:"scopes"` // Set[String]
	IsAssigned  types.Bool   `tfsdk:"assigned"`
	Policies    types.List   `tfsdk:"policies"` // List[PolicyModel]
}

func (PolicySetResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority.", // TODO: Update this comment when policy set is available for cloud
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
				Description: "Type of the policy set. Type can be one of `SitePolicies`, `DeliveryGroupPolicies`, `SiteTemplates`, or `CustomTemplates`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("DeliveryGroupPolicies"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"SitePolicies",
						"DeliveryGroupPolicies",
						"SiteTemplates",
						"CustomTemplates"}...),
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
		},
	}
}

func (PolicySetResourceModel) GetAttributes() map[string]schema.Attribute {
	return PolicySetResourceModel{}.GetSchema().Attributes
}

func (r PolicySetResourceModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, policySet *citrixorchestration.PolicySetResponse, policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse, policySetScopes []string) PolicySetResourceModel {
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

	updatedPolicySetScopes := []string{}
	for _, scopeId := range policySetScopes {
		if !strings.EqualFold(scopeId, util.AllScopeId) && scopeId != "" {
			updatedPolicySetScopes = append(updatedPolicySetScopes, scopeId)
		}
	}

	r.Scopes = util.StringArrayToStringSet(ctx, diags, updatedPolicySetScopes)

	if policies != nil && policies.Items != nil {
		policyItems := policies.Items
		sort.Slice(policyItems, func(i, j int) bool {
			return policyItems[i].GetPriority() < policyItems[j].GetPriority()
		})
		refreshedPolicies := []PolicyModel{}
		for _, policy := range policyItems {
			policyModel := PolicyModel{
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

					refreshedPolicySettings = append(refreshedPolicySettings, policySetting)
				}
			}

			sort.Slice(refreshedPolicySettings, func(i, j int) bool {
				return refreshedPolicySettings[i].Name.ValueString() < refreshedPolicySettings[j].Name.ValueString()
			})

			policyModel.PolicySettings = util.TypedArrayToObjectSet(ctx, diags, refreshedPolicySettings)

			var accessControlFilters []AccessControlFilterModel

			attributes, _ := util.AttributeMapFromObject(BranchRepeaterFilterModel{})
			policyModel.BranchRepeaterFilter = types.ObjectNull(attributes)

			var clientIpFilters []ClientIPFilterModel
			var clientNameFilters []ClientNameFilterModel
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
					_ = json.Unmarshal([]byte(filter.GetFilterData()), &uuidFilterData)

					filterType := filter.GetFilterType()
					switch filterType {
					case "AccessControl":
						accessControlFilters = append(accessControlFilters, AccessControlFilterModel{
							Allowed:    types.BoolValue(filter.GetIsAllowed()),
							Enabled:    types.BoolValue(filter.GetIsEnabled()),
							Connection: gatewayFilterData.Connection,
							Condition:  gatewayFilterData.Condition,
							Gateway:    gatewayFilterData.Gateway,
						})
					case "BranchRepeater":
						policyModel.BranchRepeaterFilter = util.TypedObjectToObjectValue(ctx, diags, BranchRepeaterFilterModel{
							Allowed: types.BoolValue(filter.GetIsAllowed()),
							Enabled: types.BoolValue(filter.GetIsEnabled()),
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
			policyModel.AccessControlFilters = util.TypedArrayToObjectSet(ctx, diags, accessControlFilters)
			policyModel.ClientIPFilters = util.TypedArrayToObjectSet(ctx, diags, clientIpFilters)
			policyModel.ClientNameFilters = util.TypedArrayToObjectSet(ctx, diags, clientNameFilters)
			policyModel.DeliveryGroupFilters = util.TypedArrayToObjectSet(ctx, diags, desktopGroupFilters)
			policyModel.DeliveryGroupTypeFilters = util.TypedArrayToObjectSet(ctx, diags, desktopKindFilters)
			policyModel.TagFilters = util.TypedArrayToObjectSet(ctx, diags, desktopTagFilters)
			policyModel.OuFilters = util.TypedArrayToObjectSet(ctx, diags, ouFilters)
			policyModel.UserFilters = util.TypedArrayToObjectSet(ctx, diags, userFilters)

			refreshedPolicies = append(refreshedPolicies, policyModel)
		}
		updatedPolicies := util.TypedArrayToObjectList(ctx, diags, refreshedPolicies)
		r.Policies = updatedPolicies
	} else {
		attributesMap, err := util.AttributeMapFromObject(PolicyModel{})
		if err != nil {
			diags.AddError("Error converting schema to attribute map. Error: ", err.Error())
		}

		r.Policies = types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	}

	r.IsAssigned = types.BoolValue(policySet.GetIsAssigned())

	return r
}
