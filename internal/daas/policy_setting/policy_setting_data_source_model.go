// Copyright © 2024. Citrix Systems, Inc.

package policy_setting

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (PolicySettingModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of a policy setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Id of the policy to which the setting belongs.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("policy_id")),
				},
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
