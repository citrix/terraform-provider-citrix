// Copyright © 2024. Citrix Systems, Inc.

package policy_set_resource

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (PolicySetV2Model) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Optional:    true,
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

func (PolicySetV2Model) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicySetV2Model{}.GetDataSourceSchema().Attributes
}
