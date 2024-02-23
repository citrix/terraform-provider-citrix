// Copyright © 2023. Citrix Systems, Inc.

package application

import (
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InstalledAppResponseModel Response object for installed application properties.
type InstalledAppResponseModel struct {
	// The command-line arguments to use when launching the executable. Environment variables can be used.
	CommandLineArguments types.String `tfsdk:"command_line_arguments"`
	// The name of the executable file to launch. The full path need not be provided if it's already in the path. Environment variables can also be used.
	CommandLineExecutable types.String `tfsdk:"command_line_executable"`
	// The working directory which the executable is launched from. Environment variables can be used.
	WorkingDirectory types.String `tfsdk:"working_directory"`
}

// ApplicationResourceModel maps the resource schema data.
type ApplicationResourceModel struct {
	Id                     types.String               `tfsdk:"id"`
	Name                   types.String               `tfsdk:"name"`
	PublishedName          types.String               `tfsdk:"published_name"`
	Description            types.String               `tfsdk:"description"`
	InstalledAppProperties *InstalledAppResponseModel `tfsdk:"installed_app_properties"`
	DeliveryGroups         []types.String             `tfsdk:"delivery_groups"`
	ApplicationFolderPath  types.String               `tfsdk:"application_folder_path"`
}

func (r ApplicationResourceModel) RefreshPropertyValues(application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {
	// Overwrite application with refreshed state
	r.Id = types.StringValue(application.GetId())
	r.Name = types.StringValue(application.GetName())
	r.PublishedName = types.StringValue(application.GetPublishedName())

	// Set optional values
	if application.GetDescription() != "" {
		r.Description = types.StringValue(application.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	if *application.GetApplicationFolder().Name.Get() != "" {
		r.ApplicationFolderPath = types.StringValue(*application.GetApplicationFolder().Name.Get())
	} else {
		r.ApplicationFolderPath = types.StringNull()
	}

	r = r.updatePlanWithInsalledAppProperties(application)
	r = r.updatePlanWithDeliveryGroups(application)

	return r
}

func (r ApplicationResourceModel) updatePlanWithInsalledAppProperties(application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {

	r.InstalledAppProperties = &InstalledAppResponseModel{}

	r.InstalledAppProperties.CommandLineExecutable = types.StringValue(application.InstalledAppProperties.GetCommandLineExecutable())

	// Set optional values
	if application.InstalledAppProperties.GetWorkingDirectory() != "" {
		r.InstalledAppProperties.WorkingDirectory = types.StringValue(application.InstalledAppProperties.GetWorkingDirectory())
	}

	if application.InstalledAppProperties.GetCommandLineArguments() != "" {
		r.InstalledAppProperties.CommandLineArguments = types.StringValue(application.InstalledAppProperties.GetCommandLineArguments())
	}

	return r
}

func (r ApplicationResourceModel) updatePlanWithDeliveryGroups(application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {

	// Add the server delivery group ids to a map
	serverDeliveryGroupIdsMap := map[string]bool{}
	for _, id := range application.GetAssociatedDeliveryGroupUuids() {
		serverDeliveryGroupIdsMap[id] = false
	}

	// Create a result list of delivery group ids
	resultDeliveryGroupIds := []types.String{}

	for _, existingDeliveryGroupId := range r.DeliveryGroups {
		_, exists := serverDeliveryGroupIdsMap[existingDeliveryGroupId.ValueString()]
		if exists {
			// Add the existing delivery group ids which matches with server data to the result list
			resultDeliveryGroupIds = append(resultDeliveryGroupIds, existingDeliveryGroupId)
			// Mark the server delivery group ids as visited
			serverDeliveryGroupIdsMap[existingDeliveryGroupId.ValueString()] = true // Mark as visited
		}
	}

	for serverDeliveryGroupId, visited := range serverDeliveryGroupIdsMap {
		// Add only unvisited delivery groups ids
		if !visited {
			resultDeliveryGroupIds = append(resultDeliveryGroupIds, types.StringValue(serverDeliveryGroupId))
		}
	}

	r.DeliveryGroups = resultDeliveryGroupIds
	return r
}
