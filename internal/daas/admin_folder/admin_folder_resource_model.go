// Copyright © 2024. Citrix Systems, Inc.
package admin_folder

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
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
	Type       types.String `tfsdk:"type"`
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
				Description: "Name of the admin folder.",
				Required:    true,
			},
			"parent_path": schema.StringAttribute{
				Description: "Path of the parent admin folder.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the admin folder. Available values are `ContainsApplications`, `ContainsMachineCatalogs`, `ContainsDeliveryGroups`, and `ContainsApplicationGroups`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATIONS),
						string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_MACHINE_CATALOGS),
						string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_DELIVERY_GROUPS),
						string(citrixorchestration.ADMINFOLDEROBJECTIDENTIFIER_CONTAINS_APPLICATION_GROUPS),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
	r.Path = types.StringValue(adminFolder.GetPath())

	metadata := adminFolder.GetMetadata()
	if len(metadata) > 0 {
		typeInfo := metadata[0].Name.Get()
		r.Type = types.StringValue(*typeInfo)
	} else {
		diagnostics.AddError(
			"Unable to update admin folder type",
			"Could not parse type for the admin folder:"+adminFolder.GetName(),
		)
	}

	var parentPath = strings.TrimSuffix(adminFolder.GetPath(), adminFolder.GetName()+"\\")
	if parentPath != "" {
		r.ParentPath = types.StringValue(parentPath)
	} else {
		r.ParentPath = types.StringNull()
	}

	return r
}
