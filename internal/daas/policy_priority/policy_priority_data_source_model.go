// Copyright © 2024. Citrix Systems, Inc.

package policy_priority

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (PolicyPriorityModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages  the policy priorities within a policy set.",
		Attributes: map[string]schema.Attribute{
			"policy_set_id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("policy_set_name")),
				},
			},
			"policy_set_name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Optional:    true,
			},
			"policy_priority": schema.ListAttribute{
				Description: "Ordered list of policy IDs. \n\n-> **Note** The order of policy IDs in the list determines the priority of the policies.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"policy_names": schema.ListAttribute{
				Description: "Ordered list of policy names. \n\n-> **Note** The order of policy names in the list reflects the priority of the policies.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (PolicyPriorityModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicyPriorityModel{}.GetDataSourceSchema().Attributes
}
