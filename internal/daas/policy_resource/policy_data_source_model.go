// Copyright © 2024. Citrix Systems, Inc.

package policy_resource

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (PolicyModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of an instance of the Policy." +
			"\n\n~> **Please Note** Each policy can only associate with one policy set. The policy will be created in the default policy set if the policy is not referenced in any of the `policy_ids` of the `citrix_policy_set_v2` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"policy_set_id": schema.StringAttribute{
				Description: "Id of the policy set the policy belongs to.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("policy_set_id")),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the policy is being enabled.",
				Computed:    true,
			},
		},
	}
}

func (PolicyModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicyModel{}.GetDataSourceSchema().Attributes
}
