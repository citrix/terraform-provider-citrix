// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AdminFolderDataSourceModel struct {
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

func (AdminFolderDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get details regarding a specific admin folder.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the admin folder.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("path")), // Ensures that only one of either Id or Path is provided. It will also cause a validation error if none are specified.
				},
			},
			"path": schema.StringAttribute{
				Description: "Path to the admin folder.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the admin folder.",
				Computed:    true,
			},
			"parent_path": schema.StringAttribute{
				Description: "Path of the parent admin folder.",
				Computed:    true,
			},
			"type": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Set of types of the admin folder.",
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

func (r AdminFolderDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminFolder *citrixorchestration.AdminFolderResponseModel) AdminFolderDataSourceModel {
	// Overwrite application folder with refreshed state
	r.Id = types.StringValue(adminFolder.GetId())
	r.Name = types.StringValue(adminFolder.GetName())

	r.Path = types.StringValue(adminFolder.GetPath())

	adminFolderTypes := []string{}
	adminFolderMetadata := adminFolder.GetMetadata()
	for _, metadata := range adminFolderMetadata {
		typeInfo := metadata.GetName()
		adminFolderTypes = append(adminFolderTypes, typeInfo)
	}
	adminFolderTypeSet := util.StringArrayToStringSet(ctx, diagnostics, adminFolderTypes)
	r.Type = adminFolderTypeSet

	var parentPath = strings.TrimSuffix(adminFolder.GetPath(), adminFolder.GetName()+"\\")
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
