// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (PolicySettingModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the policy setting.",
				Computed:    true,
			},
			"use_default": schema.BoolAttribute{
				Description: "Indicate whether using default value for the policy setting.",
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value of the policy setting.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether of the policy setting has enabled or allowed value.",
				Computed:    true,
			},
		},
	}
}

func (PolicySettingModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicySettingModel{}.GetDataSourceSchema().Attributes
}

func (PolicyModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the policy is being enabled.",
				Computed:    true,
			},
			"policy_settings": schema.SetNestedAttribute{
				Description:  "Set of policy settings.",
				Computed:     true,
				NestedObject: PolicySettingModel{}.GetDataSourceSchema(),
			},
			"access_control_filters": schema.SetNestedAttribute{
				Description:  "Set of Access control policy filters.",
				Computed:     true,
				NestedObject: AccessControlFilterModel{}.GetDataSourceSchema(),
			},
			"branch_repeater_filter": BranchRepeaterFilterModel{}.GetDataSourceSchema(),
			"client_ip_filters": schema.SetNestedAttribute{
				Description:  "Set of Client ip policy filters.",
				Computed:     true,
				NestedObject: ClientIPFilterModel{}.GetDataSourceSchema(),
			},
			"client_name_filters": schema.SetNestedAttribute{
				Description:  "Set of Client name policy filters.",
				Computed:     true,
				NestedObject: ClientNameFilterModel{}.GetDataSourceSchema(),
			},
			"delivery_group_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery group policy filters.",
				Computed:     true,
				NestedObject: DeliveryGroupFilterModel{}.GetDataSourceSchema(),
			},
			"delivery_group_type_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery group type policy filters.",
				Computed:     true,
				NestedObject: DeliveryGroupTypeFilterModel{}.GetDataSourceSchema(),
			},
			"ou_filters": schema.SetNestedAttribute{
				Description:  "Set of Organizational unit policy filters.",
				Computed:     true,
				NestedObject: OuFilterModel{}.GetDataSourceSchema(),
			},
			"user_filters": schema.SetNestedAttribute{
				Description:  "Set of User policy filters.",
				Computed:     true,
				NestedObject: UserFilterModel{}.GetDataSourceSchema(),
			},
			"tag_filters": schema.SetNestedAttribute{
				Description:  "Set of Tag policy filters.",
				Computed:     true,
				NestedObject: TagFilterModel{}.GetDataSourceSchema(),
			},
		},
	}
}

func (PolicyModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicyModel{}.GetDataSourceSchema().Attributes
}

func (PolicySetModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of a policy set and the policies within it. The order of the policies in this data source reflects the policy priority.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the policy set.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy set.",
				Computed:    true,
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the policy set to be a part of.",
				Computed:    true,
			},
			"policies": schema.ListNestedAttribute{
				Description:  "Ordered list of policies. \n\n-> **Note** The order of policies in the list determines the priority of the policies.",
				Computed:     true,
				NestedObject: PolicyModel{}.GetDataSourceSchema(),
			},
			"assigned": schema.BoolAttribute{
				Description: "Indicate whether the policy set is being assigned to delivery groups.",
				Computed:    true,
			},
			"delivery_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the delivery groups for the policy set to apply on." +
					"\n\n~> **Please Note** If `delivery_groups` attribute is unset or configured as an empty set, the policy set will not be assigned to any delivery group. None of the policies in the policy set will be applied.",
				Computed: true,
			},
		},
	}
}

func (PolicySetModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicySetModel{}.GetDataSourceSchema().Attributes
}
