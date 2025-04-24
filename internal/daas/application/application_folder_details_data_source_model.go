// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApplicationFolderDetailsDataSourceModel struct {
	Path              types.String               `tfsdk:"path"`
	TotalApplications types.Int64                `tfsdk:"total_applications"`
	ApplicationsList  []ApplicationResourceModel `tfsdk:"applications_list"`
}

func (ApplicationFolderDetailsDataSourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source for retrieving details of applications belonging to a specific folder.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Description: "The path of the folder to get the applications from.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Application Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Application Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"total_applications": schema.Int64Attribute{
				Description: "The total number of applications in the folder.",
				Computed:    true,
			},
			"applications_list": schema.ListNestedAttribute{
				Description:  "The applications list associated with the specified folder.",
				Computed:     true,
				NestedObject: ApplicationResourceModel{}.GetDataSourceSchema(),
			},
		},
	}
}

func (ApplicationFolderDetailsDataSourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ApplicationFolderDetailsDataSourceModel{}.GetDataSourceSchema().Attributes
}

func (ApplicationResourceModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Computed:    true,
			},
			"published_name": schema.StringAttribute{
				Description: "The published name of the application.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the application.",
				Computed:    true,
			},
			"installed_app_properties": InstalledAppResponseModel{}.GetDataSourceSchema(),
			"application_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The application group IDs to which the application should be added.",
				Computed:    true,
			},
			"delivery_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The delivery groups which the application is associated with.",
				Computed:    true,
			},
			"delivery_groups_priority": schema.SetNestedAttribute{
				NestedObject: DeliveryGroupPriorityModel{}.GetDataSourceSchema(),
				Description:  "Set of delivery groups with their corresponding priority.",
				Computed:     true,
			},
			"application_folder_path": schema.StringAttribute{
				Description: "The path of the folder which the application belongs to",
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Description: "The Id of the icon to be associated with the application.",
				Computed:    true,
			},
			"limit_visibility_to_users": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "By default, the application is visible to all users within a delivery group. However, you can restrict its visibility to only certain users by specifying them in the `limit_visibility_to_users` list.",
				Computed:    true,
			},
			"application_category_path": schema.StringAttribute{
				Description: "The application category path allows users to organize and view applications under specific categories in Citrix Workspace App.",
				Computed:    true,
			},
			"metadata": schema.ListNestedAttribute{
				Description:  "Metadata for the Application.",
				Computed:     true,
				NestedObject: util.NameValueStringPairModel{}.GetDataSourceSchema(),
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the application.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates whether the application is enabled or disabled. Default is `true`.",
				Computed:    true,
			},
			"limit_to_one_instance_per_user": schema.BoolAttribute{
				Description: "Specifies if the use of the application should be limited to only one instance per user. Default is `false`.",
				Computed:    true,
			},
			"max_total_instances": schema.Int32Attribute{
				Description: "Control the use of this application by limiting the number of instances running at the same time. If set to 0, it allows unlimited use.",
				Computed:    true,
			},
			"shortcut_added_to_desktop": schema.BoolAttribute{
				Description: "Indicates whether a shortcut to the application is added to the desktop. Default is `false`.",
				Computed:    true,
			},
			"shortcut_added_to_start_menu": schema.BoolAttribute{
				Description: "Indicates whether a shortcut to the application is added to the start menu. Default is `false`.",
				Computed:    true,
			},
			"visible": schema.BoolAttribute{
				Description: "Specifies whether or not this application is visible to users. Note that it's possible for an application to be disabled and still visible. Default is `true`.",
				Computed:    true,
			},
			"browser_name": schema.StringAttribute{
				Description: "The browser name for the application. When omitted, the application name will be used as the browser name.",
				Computed:    true,
			},
		},
	}
}

func (ApplicationResourceModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ApplicationResourceModel{}.GetDataSourceSchema().Attributes
}

func (InstalledAppResponseModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The installed application properties of the application.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"command_line_arguments": schema.StringAttribute{
				Description: "The command-line arguments to use when launching the executable. Environment variables can be used.",
				Computed:    true,
			},
			"command_line_executable": schema.StringAttribute{
				Description: "The name of the executable file to launch. The full path need not be provided if it's already in the path. Environment variables can also be used.",
				Computed:    true,
			},
			"working_directory": schema.StringAttribute{
				Description: "The working directory which the executable is launched from. Environment variables can be used.",
				Computed:    true,
			},
		},
	}
}

func (InstalledAppResponseModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return InstalledAppResponseModel{}.GetDataSourceSchema().Attributes
}

func (r ApplicationFolderDetailsDataSourceModel) RefreshPropertyValues(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, apps *citrixorchestration.ApplicationResponseModelCollection) ApplicationFolderDetailsDataSourceModel {

	var res []ApplicationResourceModel
	for _, app := range apps.GetItems() {
		appDetail, err := getApplication(ctx, client, diagnostics, app.GetId())
		if err != nil {
			diagnostics.AddError(
				"Error getting application details for application: "+app.GetId(),
				err.Error(),
			)
			continue
		}
		adminFolder := appDetail.GetApplicationFolder()
		adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
		applicationPath := util.BuildResourcePathForGetRequest(adminFolderPath, app.GetName())
		appDeliveryGroups, err := getApplicationDeliveryGroups(ctx, client, diagnostics, applicationPath)
		if err != nil {
			continue
		}
		tags := getApplicationTags(ctx, diagnostics, client, applicationPath)
		appModel := ApplicationResourceModel{
			Id:                        types.StringValue(appDetail.GetId()),
			Name:                      types.StringValue(appDetail.GetName()),
			PublishedName:             types.StringValue(appDetail.GetPublishedName()),
			Description:               types.StringValue(appDetail.GetDescription()),
			Icon:                      types.StringValue(appDetail.GetIconId()),
			ApplicationCategoryPath:   types.StringValue(appDetail.GetClientFolder()),
			Enabled:                   types.BoolValue(appDetail.GetEnabled()),
			LimitToOneInstancePerUser: types.BoolValue(appDetail.GetMaxPerUserInstances() == 1),
			MaxTotalInstances:         types.Int32Value(appDetail.GetMaxTotalInstances()),
			ShortcutAddedToDesktop:    types.BoolValue(appDetail.GetShortcutAddedToDesktop()),
			ShortcutAddedToStartMenu:  types.BoolValue(appDetail.GetShortcutAddedToStartMenu()),
			Visible:                   types.BoolValue(appDetail.GetVisible()),
			BrowserName:               types.StringValue(appDetail.GetBrowserName()),
			InstalledAppProperties:    r.getInstalledAppProperties(ctx, diagnostics, app),
			ApplicationGroups:         util.StringArrayToStringList(ctx, diagnostics, appDetail.GetAssociatedApplicationGroupUuids()),
			ApplicationFolderPath:     types.StringValue(adminFolderPath),
		}

		includedUsers := appDetail.GetIncludedUsers()
		if appDetail.GetIncludedUserFilterEnabled() {
			userList := []string{}
			for _, user := range includedUsers {
				if user.GetSamName() != "" {
					userList = append(userList, user.GetSamName())
				} else if user.GetPrincipalName() != "" {
					userList = append(userList, user.GetPrincipalName())
				}
			}
			appModel.LimitVisibilityToUsers = util.StringArrayToStringSet(ctx, diagnostics, userList)
		} else {
			appModel.LimitVisibilityToUsers = types.SetNull(types.StringType)
		}

		if appDeliveryGroups != nil && len(appDeliveryGroups.GetItems()) > 0 {
			deliveryGroups := appDeliveryGroups.GetItems()
			deliveryGroupsWithPriority := []string{}
			deliveryGroupsPriority := []DeliveryGroupPriorityModel{}
			sort.Slice(deliveryGroups, func(i, j int) bool {
				return deliveryGroups[i].GetPriority() < deliveryGroups[j].GetPriority()
			})
			for _, deliveryGroup := range deliveryGroups {
				deliveryGroupsWithPriority = append(deliveryGroupsWithPriority, deliveryGroup.GetId())
				deliveryGroupsPriority = append(deliveryGroupsPriority, DeliveryGroupPriorityModel{
					Id:       types.StringValue(deliveryGroup.GetId()),
					Priority: types.Int32Value(deliveryGroup.GetPriority()),
				})
			}
			appModel.DeliveryGroups = util.StringArrayToStringList(ctx, diagnostics, deliveryGroupsWithPriority)
			appModel.DeliveryGroupsPriority = util.DataSourceTypedArrayToObjectSet(ctx, diagnostics, deliveryGroupsPriority)
		} else {
			appModel.DeliveryGroups = types.ListNull(types.StringType)
			attrMap, err := util.ResourceAttributeMapFromObject(DeliveryGroupPriorityModel{})
			if err != nil {
				diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
				continue
			}
			appModel.DeliveryGroupsPriority = types.SetNull(types.ObjectType{AttrTypes: attrMap})
		}

		metadataList := []util.NameValueStringPairModel{}
		for _, metadata := range appDetail.GetMetadata() {
			metadataList = append(metadataList, util.NameValueStringPairModel{
				Name:  types.StringValue(metadata.GetName()),
				Value: types.StringValue(metadata.GetValue()),
			})
		}
		appModel.Metadata = util.DataSourceTypedArrayToObjectList(ctx, diagnostics, metadataList)

		appModel.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

		res = append(res, appModel)
	}

	r.ApplicationsList = res
	r.TotalApplications = types.Int64Value(int64(*apps.TotalItems.Get()))
	return r
}

func (r ApplicationFolderDetailsDataSourceModel) getInstalledAppProperties(ctx context.Context, diagnostics *diag.Diagnostics, app citrixorchestration.ApplicationResponseModel) types.Object {
	var installedAppResponse = InstalledAppResponseModel{
		CommandLineArguments:  types.StringValue(app.GetInstalledAppProperties().CommandLineArguments),
		CommandLineExecutable: types.StringValue(app.GetInstalledAppProperties().CommandLineExecutable),
		WorkingDirectory:      types.StringValue(app.GetInstalledAppProperties().WorkingDirectory),
	}
	return util.TypedObjectToObjectValue(ctx, diagnostics, installedAppResponse)
}

func (r ApplicationFolderDetailsDataSourceModel) getDeliveryGroups(ctx context.Context, diagnostics *diag.Diagnostics, app citrixorchestration.ApplicationResponseModel) types.List {
	return util.StringArrayToStringList(ctx, diagnostics, app.AssociatedDeliveryGroupUuids)
}

func (DeliveryGroupPriorityModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Id of the delivery group.",
				Computed:    true,
			},
			"priority": schema.Int32Attribute{
				Description: "The priority of the delivery group. `0` means the highest priority.",
				Computed:    true,
			},
		},
	}
}

func (DeliveryGroupPriorityModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return DeliveryGroupPriorityModel{}.GetDataSourceSchema().Attributes
}
