// Copyright © 2024. Citrix Systems, Inc.
package qcs_image

import (
	"context"
	"regexp"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesImageModel struct {
	Id                 types.String `tfsdk:"id"`
	AccountId          types.String `tfsdk:"account_id"`
	AwsImageId         types.String `tfsdk:"aws_image_id"`
	AwsImportedImageId types.String `tfsdk:"aws_imported_image_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	SessionSupport     types.String `tfsdk:"session_support"`
	OperatingSystem    types.String `tfsdk:"operating_system"`
	Tenancy            types.String `tfsdk:"tenancy"`
	IngestionProcess   types.String `tfsdk:"ingestion_process"`
	State              types.String `tfsdk:"state"`
}

func (AwsWorkspacesImageModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Manages an AWS WorkSpaces image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "GUID identifier of the account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_image_id": schema.StringAttribute{
				Description: "Id of the image to be imported in AWS.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsAmiAndWsiRegex), "must be in AMI (`ami-{ImageId}`) or WSI (`wsi-{ImageId}`) format"),
				},
			},
			"aws_imported_image_id": schema.StringAttribute{
				Description: "The Id of the image imported in AWS WorkSpaces.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the image.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the image.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"session_support": schema.StringAttribute{
				Description: "The supported session type of the image. Possible values are `SingleSession` and `MultiSession`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(quickcreateservice.SESSIONSUPPORT_SINGLE_SESSION),
						string(quickcreateservice.SESSIONSUPPORT_MULTI_SESSION),
					),
				},
			},
			"operating_system": schema.StringAttribute{
				Description: "The type of operating system of the image. Possible values are `WINDOWS` and `LINUX`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(quickcreateservice.OPERATINGSYSTEMTYPE_WINDOWS),
						string(quickcreateservice.OPERATINGSYSTEMTYPE_LINUX),
					),
				},
			},
			"ingestion_process": schema.StringAttribute{
				Description: "The type of ingestion process of the image. Possible values are `BYOL_REGULAR_BYOP` and `BYOL_GRAPHICS_G4DN_BYOP`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(quickcreateservice.AWSEDCWORKSPACEIMAGEINGESTIONPROCESS_BYOL_REGULAR_BYOP),
						string(quickcreateservice.AWSEDCWORKSPACEIMAGEINGESTIONPROCESS_BYOL_GRAPHICS_G4_DN_BYOP),
					),
				},
			},
			"tenancy": schema.StringAttribute{
				Description: "The type of tenancy of the image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Description: "The state of ingestion process of the image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (AwsWorkspacesImageModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesImageModel{}.GetSchema().Attributes
}

func (AwsWorkspacesImageModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r AwsWorkspacesImageModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, image *quickcreateservice.AwsEdcImage) AwsWorkspacesImageModel {
	r.Id = types.StringValue(image.GetImageId())
	r.AccountId = types.StringValue(image.GetAccountId())
	if !isResource {
		r.AwsImageId = types.StringNull()
	}
	r.AwsImportedImageId = types.StringValue(image.GetAmazonImageId())
	r.Name = types.StringValue(image.GetName())
	r.Description = types.StringValue(image.GetDescription())
	r.SessionSupport = types.StringValue(util.QcsSessionSupportEnumToString(image.GetSessionSupport()))
	r.OperatingSystem = types.StringValue(util.OperatingSystemTypeEnumToString(image.GetOperatingSystem()))
	r.Tenancy = types.StringValue(util.AwsEdcWorkspaceImageTenancyEnumToString(image.GetWorkspaceImageTenancy()))
	r.IngestionProcess = types.StringValue(util.AwsEdcWorkspaceImageIngestionProcessEnumToString(image.GetIngestionProcess()))
	r.State = types.StringValue(util.AwsEdcWorkspaceImageStateEnumToString(image.GetWorkspaceImageState()))

	return r
}
