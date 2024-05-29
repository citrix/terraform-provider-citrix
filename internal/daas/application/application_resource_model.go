// Copyright © 2023. Citrix Systems, Inc.

package application

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (InstalledAppResponseModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The install application properties.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"command_line_arguments": schema.StringAttribute{
				Description: "The command-line arguments to use when launching the executable.",
				Optional:    true,
				Validators: []validator.String{
					validator.String(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
			"command_line_executable": schema.StringAttribute{
				Description: "The path of the executable file to launch.",
				Required:    true,
			},
			"working_directory": schema.StringAttribute{
				Description: "The working directory which the executable is launched from.",
				Optional:    true,
				Validators: []validator.String{
					validator.String(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
		},
	}
}

func (InstalledAppResponseModel) GetAttributes() map[string]schema.Attribute {
	return InstalledAppResponseModel{}.GetSchema().Attributes
}

// ApplicationResourceModel maps the resource schema data.
type ApplicationResourceModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	PublishedName          types.String `tfsdk:"published_name"`
	Description            types.String `tfsdk:"description"`
	InstalledAppProperties types.Object `tfsdk:"installed_app_properties"` // InstalledAppResponseModel
	DeliveryGroups         types.List   `tfsdk:"delivery_groups"`          //List[string]
	ApplicationFolderPath  types.String `tfsdk:"application_folder_path"`
}

// Schema defines the schema for the data source.
func GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Resource for creating and managing applications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the application.",
				Required:    true,
			},
			"published_name": schema.StringAttribute{
				Description: "A display name for the application that is shown to users.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the application.",
				Optional:    true,
			},
			"installed_app_properties": InstalledAppResponseModel{}.GetSchema(),
			"delivery_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The delivery group id's to which the application should be added.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"application_folder_path": schema.StringAttribute{
				Description: "The application folder path in which the application should be created.",
				Optional:    true,
			},
		},
	}
}

func (r ApplicationResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {
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

	r = r.updatePlanWithInsalledAppProperties(ctx, diagnostics, application)
	r = r.updatePlanWithDeliveryGroups(ctx, diagnostics, application)

	return r
}

func (r ApplicationResourceModel) updatePlanWithInsalledAppProperties(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {

	var installedAppProperties = InstalledAppResponseModel{}

	installedAppProperties.CommandLineExecutable = types.StringValue(application.InstalledAppProperties.GetCommandLineExecutable())

	// Set optional values
	if application.InstalledAppProperties.GetWorkingDirectory() != "" {
		installedAppProperties.WorkingDirectory = types.StringValue(application.InstalledAppProperties.GetWorkingDirectory())
	}

	if application.InstalledAppProperties.GetCommandLineArguments() != "" {
		installedAppProperties.CommandLineArguments = types.StringValue(application.InstalledAppProperties.GetCommandLineArguments())
	}

	r.InstalledAppProperties = util.TypedObjectToObjectValue(ctx, diagnostics, installedAppProperties)
	return r
}

func (r ApplicationResourceModel) updatePlanWithDeliveryGroups(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationDetailResponseModel) ApplicationResourceModel {

	// Add the server delivery group ids to a map
	serverDeliveryGroupIdsMap := map[string]bool{}
	for _, id := range application.GetAssociatedDeliveryGroupUuids() {
		serverDeliveryGroupIdsMap[id] = false
	}

	// Create a result list of delivery group ids
	resultDeliveryGroupIds := []string{}

	for _, existingDeliveryGroupId := range util.StringListToStringArray(ctx, diagnostics, r.DeliveryGroups) {
		_, exists := serverDeliveryGroupIdsMap[existingDeliveryGroupId]
		if exists {
			// Add the existing delivery group ids which matches with server data to the result list
			resultDeliveryGroupIds = append(resultDeliveryGroupIds, existingDeliveryGroupId)
			// Mark the server delivery group ids as visited
			serverDeliveryGroupIdsMap[existingDeliveryGroupId] = true // Mark as visited
		}
	}

	for serverDeliveryGroupId, visited := range serverDeliveryGroupIdsMap {
		// Add only unvisited delivery groups ids
		if !visited {
			resultDeliveryGroupIds = append(resultDeliveryGroupIds, serverDeliveryGroupId)
		}
	}

	r.DeliveryGroups = util.StringArrayToStringList(ctx, diagnostics, resultDeliveryGroupIds)
	return r
}
