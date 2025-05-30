// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AdminFolderModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Type                   types.Set    `tfsdk:"type"` // Set[String]
	Path                   types.String `tfsdk:"path"`
	ParentPath             types.String `tfsdk:"parent_path"`
	TotalApplications      types.Int64  `tfsdk:"total_applications"`
	TotalMachineCatalogs   types.Int64  `tfsdk:"total_machine_catalogs"`
	TotalApplicationGroups types.Int64  `tfsdk:"total_application_groups"`
	TotalDeliveryGroups    types.Int64  `tfsdk:"total_delivery_groups"`
}

func (AdminFolderModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an admin folder.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the admin folder.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin folder. Admin Folder name should be unique within the same parent folder.",
				Required:    true,
			},
			"parent_path": schema.StringAttribute{
				Description: "Path of the parent admin folder. Please note that the parent path should not end with a `\\`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"type": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Set of types of the admin folder. Available values are `ContainsApplications`, `ContainsMachineCatalogs`, `ContainsDeliveryGroups`, and `ContainsApplicationGroups`.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.OneOf(
								string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATIONS),
								string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_MACHINE_CATALOGS),
								string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_DELIVERY_GROUPS),
								string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATION_GROUPS),
							),
						),
					),
					setvalidator.SizeAtLeast(1),
				},
			},
			"path": schema.StringAttribute{
				Description: "Path to the admin folder.",
				Computed:    true,
			},
			"total_applications": schema.Int64Attribute{
				Description: "Number of applications contained in the admin folder.",
				Computed:    true,
			},
			"total_machine_catalogs": schema.Int64Attribute{
				Description: "Number of machine catalogs contained in the admin folder.",
				Computed:    true,
			},
			"total_application_groups": schema.Int64Attribute{
				Description: "Number of application groups contained in the admin folder.",
				Computed:    true,
			},
			"total_delivery_groups": schema.Int64Attribute{
				Description: "Number of delivery groups contained in the admin folder.",
				Computed:    true,
			},
		},
	}
}

func (AdminFolderModel) GetAttributes() map[string]schema.Attribute {
	return AdminFolderModel{}.GetSchema().Attributes
}

func (AdminFolderModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r AdminFolderModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminFolder *citrixorchestration.AdminFolderResponseModel) AdminFolderModel {
	// Overwrite application folder with refreshed state
	r.Id = types.StringValue(adminFolder.GetId())
	r.Name = types.StringValue(adminFolder.GetName())

	// Set optional values
	r.Path = types.StringValue(strings.TrimSuffix(adminFolder.GetPath(), "\\"))

	adminFolderTypes := []string{}
	adminFolderMetadata := adminFolder.GetMetadata()
	for _, metadata := range adminFolderMetadata {
		typeInfo := metadata.GetName()
		adminFolderTypes = append(adminFolderTypes, typeInfo)
	}
	adminFolderTypeSet := util.StringArrayToStringSet(ctx, diagnostics, adminFolderTypes)
	r.Type = adminFolderTypeSet

	var parentPath = strings.TrimSuffix(adminFolder.GetPath(), adminFolder.GetName()+"\\")
	parentPath = strings.TrimSuffix(parentPath, "\\")
	if parentPath != "" {
		r.ParentPath = types.StringValue(parentPath)
	} else {
		r.ParentPath = types.StringNull()
	}

	r.TotalApplications = types.Int64Value(int64(adminFolder.GetTotalApplications()))
	r.TotalMachineCatalogs = types.Int64Value(int64(adminFolder.GetTotalMachineCatalogs()))
	r.TotalApplicationGroups = types.Int64Value(int64(adminFolder.GetTotalApplicationGroups()))
	r.TotalDeliveryGroups = types.Int64Value(int64(adminFolder.GetTotalDesktopGroups()))

	return r
}
