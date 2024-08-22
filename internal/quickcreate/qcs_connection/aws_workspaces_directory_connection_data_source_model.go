// Copyright © 2024. Citrix Systems, Inc.
package qcs_connection

import (
	"context"
	"regexp"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesDirectoryConnectionDataSourceModel struct {
	DirectoryConnectionId           types.String `tfsdk:"id"`
	AccountId                       types.String `tfsdk:"account"`
	Name                            types.String `tfsdk:"name"`
	ZoneId                          types.String `tfsdk:"zone"`
	ResourceLocationId              types.String `tfsdk:"resource_location"`
	Directory                       types.String `tfsdk:"directory"`
	Subnets                         types.Set    `tfsdk:"subnets"`
	Tenancy                         types.String `tfsdk:"tenancy"`
	UserEnabledAsLocalAdministrator types.Bool   `tfsdk:"user_enabled_as_local_administrator"`
	SecurityGroup                   types.String `tfsdk:"security_group"`
	DefaultOu                       types.String `tfsdk:"default_ou"`
}

func (AwsWorkspacesDirectoryConnectionDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Data Source of an AWS WorkSpaces directory connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the directory connection.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the directory connection.",
				Optional:    true,
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

func (AwsWorkspacesDirectoryConnectionDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDirectoryConnectionDataSourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesDirectoryConnectionDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, directory *quickcreateservice.AwsEdcDirectoryConnection) AwsWorkspacesDirectoryConnectionDataSourceModel {
	r.DirectoryConnectionId = types.StringValue(directory.GetConnectionId())
	r.Name = types.StringValue(directory.GetName())
	r.AccountId = types.StringValue(directory.GetAccountId())
	r.ZoneId = types.StringValue(directory.GetZoneId())
	r.ResourceLocationId = types.StringValue(directory.GetResourceLocationId())
	r.Directory = types.StringValue(directory.GetDirectoryId())
	r.Subnets = util.StringArrayToStringSet(ctx, diagnostics, []string{directory.GetSubnet1Id(), directory.GetSubnet2Id()})
	r.Tenancy = types.StringValue(string(directory.GetTenancy()))
	r.UserEnabledAsLocalAdministrator = types.BoolValue(directory.GetUserEnabledAsLocalAdministrator())
	r.SecurityGroup = types.StringValue(directory.GetSecurityGroupId())
	r.DefaultOu = types.StringValue(directory.GetDefaultOU())

	return r
}
