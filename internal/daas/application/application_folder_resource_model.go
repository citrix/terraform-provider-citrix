// Copyright © 2023. Citrix Systems, Inc.

package application

import (
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApplicationFolderResourceModel maps the resource schema data.
type ApplicationFolderResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Path       types.String `tfsdk:"path"`
	ParentPath types.String `tfsdk:"parent_path"`
}

func (r ApplicationFolderResourceModel) RefreshPropertyValues(application *citrixorchestration.AdminFolderResponseModel) ApplicationFolderResourceModel {
	// Overwrite application folder with refreshed state
	r.Id = types.StringValue(application.GetId())
	r.Name = types.StringValue(application.GetName())

	// Set optional values
	if application.GetPath() != "" {
		r.Path = types.StringValue(application.GetPath())
	} else {
		r.Path = types.StringNull()
	}

	var parent_path = strings.TrimSuffix(application.GetPath(), application.GetName()+"\\")
	if parent_path != "" {
		r.ParentPath = types.StringValue(parent_path)
	} else {
		r.ParentPath = types.StringNull()
	}

	return r
}
