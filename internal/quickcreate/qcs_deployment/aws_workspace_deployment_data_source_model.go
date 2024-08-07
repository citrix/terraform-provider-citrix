// Copyright © 2024. Citrix Systems, Inc.
package qcs_deployment

import (
	"context"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesDeploymentDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AccountId   types.String `tfsdk:"account_id"`
	DirectoryId types.String `tfsdk:"directory_connection_id"`
	ImageId     types.String `tfsdk:"image_id"`
}

func (AwsWorkspacesDeploymentDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS Workspaces Core --- Data source to get details of an AWS Workspaces deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the deployment.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("account_id")),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "GUID of the account.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.AlsoRequires(path.MatchRoot("name")),
				},
			},
			"directory_connection_id": schema.StringAttribute{
				Description: "GUID of the directory connection.",
				Computed:    true,
			},
			"image_id": schema.StringAttribute{
				Description: "GUID of the image.",
				Computed:    true,
			},
		},
	}
}

func (AwsWorkspacesDeploymentDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDeploymentDataSourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesDeploymentDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deployment citrixquickcreate.AwsEdcDeployment) AwsWorkspacesDeploymentDataSourceModel {
	r.Id = types.StringValue(deployment.GetDeploymentId())
	r.Name = types.StringValue(deployment.GetDeploymentName())
	r.AccountId = types.StringValue(deployment.GetAccountId())
	r.DirectoryId = types.StringValue(deployment.GetConnectionId())
	r.ImageId = types.StringValue(deployment.GetImageId())

	return r
}
