// Copyright © 2024. Citrix Systems, Inc.
package qcs_image

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

type AwsWorkspacesImageDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	AccountId        types.String `tfsdk:"account_id"`
	AwsImageId       types.String `tfsdk:"aws_image_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	SessionSupport   types.String `tfsdk:"session_support"`
	OperatingSystem  types.String `tfsdk:"operating_system"`
	Tenancy          types.String `tfsdk:"tenancy"`
	IngestionProcess types.String `tfsdk:"ingestion_process"`
	State            types.String `tfsdk:"state"`
}

func (AwsWorkspacesImageDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS Workspaces Core --- Data Source of an AWS Workspaces image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the image.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"account_id": schema.StringAttribute{
				Description: "GUID identifier of the account.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the image.",
				Optional:    true,
			},
			"aws_image_id": schema.StringAttribute{
				Description: "Id of the image on AWS.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the image.",
				Computed:    true,
			},
			"session_support": schema.StringAttribute{
				Description: "The supported session type of the image. Possible values are `SingleSession` and `MultiSession`.",
				Computed:    true,
			},
			"operating_system": schema.StringAttribute{
				Description: "The type of operating system of the image. Possible values are `WINDOWS` and `LINUX`.",
				Computed:    true,
			},
			"ingestion_process": schema.StringAttribute{
				Description: "The type of ingestion process of the image. Possible values are `BYOL_REGULAR_BYOP` and `BYOL_GRAPHICS_G4DN_BYOP`.",
				Computed:    true,
			},
			"tenancy": schema.StringAttribute{
				Description: "The type of tenancy of the image.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The state of ingestion process of the image.",
				Computed:    true,
			},
		},
	}
}

func (AwsWorkspacesImageDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesImageDataSourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesImageDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, image *quickcreateservice.AwsEdcImage) AwsWorkspacesImageDataSourceModel {
	r.Id = types.StringValue(image.GetImageId())
	r.AccountId = types.StringValue(image.GetAccountId())
	r.AwsImageId = types.StringValue(image.GetAmazonImageId())
	r.Name = types.StringValue(image.GetName())
	r.Description = types.StringValue(image.GetDescription())
	r.SessionSupport = types.StringValue(util.SessionSupportEnumToString(image.GetSessionSupport()))
	r.OperatingSystem = types.StringValue(util.OperatingSystemTypeEnumToString(image.GetOperatingSystem()))
	r.Tenancy = types.StringValue(util.AwsEdcWorkspaceImageTenancyEnumToString(image.GetWorkspaceImageTenancy()))
	r.IngestionProcess = types.StringValue(util.AwsEdcWorkspaceImageIngestionProcessEnumToString(image.GetIngestionProcess()))
	r.State = types.StringValue(util.AwsEdcWorkspaceImageStateEnumToString(image.GetWorkspaceImageState()))

	return r
}
