// Copyright © 2024. Citrix Systems, Inc.

package storefront_server

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (StoreFrontServerResourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source of a StoreFront server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the StoreFront server.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the StoreFront server.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the StoreFront server.",
				Computed:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL for connecting to the StoreFront server.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates if the StoreFront server is enabled. Default is `true`.",
				Computed:    true,
			},
		},
	}
}
