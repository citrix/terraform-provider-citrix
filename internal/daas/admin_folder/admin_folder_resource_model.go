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

type AdminFolderResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.Set    `tfsdk:"type"` // Set[String]
	Path       types.String `tfsdk:"path"`
	ParentPath types.String `tfsdk:"parent_path"`
}

func (AdminFolderResourceModel) GetSchema() schema.Schema {
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
		},
	}
}

func (AdminFolderResourceModel) GetAttributes() map[string]schema.Attribute {
	return AdminFolderResourceModel{}.GetSchema().Attributes
}

func (r AdminFolderResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminFolder *citrixorchestration.AdminFolderResponseModel) AdminFolderResourceModel {
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

	return r
}
