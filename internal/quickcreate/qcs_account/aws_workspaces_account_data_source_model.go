// Copyright © 2024. Citrix Systems, Inc.

package qcs_account

import (
	"regexp"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesAccountDataSourceModel struct {
	AccountId             types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	AwsAccount            types.String `tfsdk:"aws_account"`
	AwsRegion             types.String `tfsdk:"aws_region"`
	AwsByolFeatureEnabled types.Bool   `tfsdk:"aws_byol_feature_enabled"`
}

func (AwsWorkspacesAccountDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Data source to get details of an AWS WorkSpaces account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the account.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the account.",
				Optional:    true,
			},
			"aws_account": schema.StringAttribute{
				Description: "AWS account number associated with the account.",
				Computed:    true,
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region the account is associated with.",
				Computed:    true,
			},
			"aws_byol_feature_enabled": schema.BoolAttribute{
				Description: "Indicates if the associated AWS EDC account has BYOL support enabled.",
				Computed:    true,
			},
		},
	}
}

func (AwsWorkspacesAccountDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesAccountDataSourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesAccountDataSourceModel) RefreshPropertyValues(account *quickcreateservice.AwsEdcAccount) AwsWorkspacesAccountDataSourceModel {
	r.AccountId = types.StringValue(account.GetAccountId())
	r.Name = types.StringValue(account.GetName())
	r.AwsRegion = types.StringValue(account.GetAwsRegion())
	r.AwsByolFeatureEnabled = types.BoolValue(account.GetAwsByolFeatureEnabled())
	r.AwsAccount = types.StringValue(account.GetAwsAccount())

	return r
}
