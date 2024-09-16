// Copyright © 2024. Citrix Systems, Inc.

package qcs_account

import (
	"regexp"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesAccountResourceModel struct {
	AccountId             types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	AwsAccount            types.String `tfsdk:"aws_account"`
	AwsRegion             types.String `tfsdk:"aws_region"`
	AwsAccessKeyId        types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey    types.String `tfsdk:"aws_secret_access_key"`
	AwsRoleArn            types.String `tfsdk:"aws_role_arn"`
	AwsByolFeatureEnabled types.Bool   `tfsdk:"aws_byol_feature_enabled"`
}

func (AwsWorkspacesAccountResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Manages an AWS WorkSpaces account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the account.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the account.",
				Required:    true,
			},
			"aws_account": schema.StringAttribute{
				Description: "AWS account number associated with the account.",
				Computed:    true,
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region the account is associated with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsRegionRegex), "AWS Region can only contain alphanumeric characters and hyphens"),
				},
			},
			"aws_access_key_id": schema.StringAttribute{
				Description: "ID of the Access Key associated with the account.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsAccessKeyIdRegex), "AWS AccessKeyId can only contain letters, numbers and underscore"),
					stringvalidator.AlsoRequires(path.MatchRoot("aws_secret_access_key")),
				},
			},
			"aws_secret_access_key": schema.StringAttribute{
				Description: "Secret associated with the Access Key for the account.",
				Optional:    true,
				Sensitive:   true,
			},
			"aws_role_arn": schema.StringAttribute{
				Description: "ARN of the role to assume when making requests in this account.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(20),
					stringvalidator.LengthAtMost(2048),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsRoleArnRegex), "The Role ARN provided contains invalid characters or is in an incorrect format"),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("aws_access_key_id"),
					),
				},
			},
			"aws_byol_feature_enabled": schema.BoolAttribute{
				Description: "Indicates if the associated AWS EDC account has BYOL support enabled.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (AwsWorkspacesAccountResourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesAccountResourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesAccountResourceModel) RefreshPropertyValues(account *quickcreateservice.AwsEdcAccount) AwsWorkspacesAccountResourceModel {
	r.AccountId = types.StringValue(account.GetAccountId())
	r.Name = types.StringValue(account.GetName())
	r.AwsRegion = types.StringValue(account.GetAwsRegion())
	r.AwsByolFeatureEnabled = types.BoolValue(account.GetAwsByolFeatureEnabled())
	r.AwsAccount = types.StringValue(account.GetAwsAccount())

	return r
}
