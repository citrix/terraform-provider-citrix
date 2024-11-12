// Copyright © 2024. Citrix Systems, Inc.
package qcs_connection

import (
	"regexp"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (AwsWorkspacesDirectoryConnectionModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Data Source of an AWS WorkSpaces directory connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the directory connection.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the directory connection.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"account": schema.StringAttribute{
				Description: "ID of the account the directory connection is associated with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"zone": schema.StringAttribute{
				Description: "ID of the zone the directory connection is associated with. Only one of `zone` and `resource_location` attributes can be specified.",
				Computed:    true,
			},
			"resource_location": schema.StringAttribute{
				Description: "ID of the resource location the directory connection is associated with. Only one of `resource_location` and `zone` attributes can be specified.",
				Computed:    true,
			},
			"directory": schema.StringAttribute{
				Description: "ID of the AWS directory the directory connection is associated with.",
				Computed:    true,
			},
			"subnets": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "IDs of the subnets the directory connection is associated with.",
				Computed:    true,
			},
			"tenancy": schema.StringAttribute{
				Description: "Tenancy of the directory connection. Possible values are `SHARED` and `DEDICATED`. Defaults to `DEDICATED`.",
				Computed:    true,
			},
			"user_enabled_as_local_administrator": schema.BoolAttribute{
				Description: "Enable users to be local administrators. Defaults to `false`.",
				Computed:    true,
			},
			"security_group": schema.StringAttribute{
				Description: "ID of the security group the directory connection is associated with.",
				Computed:    true,
			},
			"default_ou": schema.StringAttribute{
				Description: "Default OU for VDAs in the directory connection.",
				Computed:    true,
			},
		},
	}
}

func (AwsWorkspacesDirectoryConnectionModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDirectoryConnectionModel{}.GetDataSourceSchema().Attributes
}
