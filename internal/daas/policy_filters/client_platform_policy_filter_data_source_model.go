// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (ClientPlatformFilterModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: getPolicyFilterDataSourceDescription("Client Platform"),
		Attributes:  ClientPlatformFilterModel{}.GetDataSourceAttributes(),
	}
}

func (ClientPlatformFilterModel) GetDataSourceNestedAttributeObjectSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: ClientPlatformFilterModel{}.GetDataSourceAttributes(),
	}
}

func (ClientPlatformFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Id of the client platform policy filter.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
			},
		},
		"policy_id": schema.StringAttribute{
			Description: "Id of the policy to which the filter belongs.",
			Computed:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Indicate whether the filter is being enabled.",
			Computed:    true,
		},
		"allowed": schema.BoolAttribute{
			Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
			Computed:    true,
		},
		"platform": schema.StringAttribute{
			Description: "Name of the client platform to be filtered. Available values are `Windows`, `Linux`, `Mac`, `Ios`, `Android`, and `Html5`.",
			Computed:    true,
		},
	}
}
