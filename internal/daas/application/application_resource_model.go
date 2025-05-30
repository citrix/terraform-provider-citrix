// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"regexp"
	"sort"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeliveryGroupPriorityModel struct {
	Id       types.String `tfsdk:"id"`
	Priority types.Int32  `tfsdk:"priority"`
}

func (DeliveryGroupPriorityModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The Id of the delivery group.",
				Required:    true,
				Validators: []validator.String{
					validator.String(
						stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					),
				},
			},
			"priority": schema.Int32Attribute{
				Description: "The priority of the delivery group. `0` means the highest priority.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
		},
	}
}

func (DeliveryGroupPriorityModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupPriorityModel{}.GetSchema().Attributes
}

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
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	PublishedName             types.String `tfsdk:"published_name"`
	Description               types.String `tfsdk:"description"`
	InstalledAppProperties    types.Object `tfsdk:"installed_app_properties"` // InstalledAppResponseModel
	ApplicationGroups         types.List   `tfsdk:"application_groups"`       // List[string]
	DeliveryGroups            types.List   `tfsdk:"delivery_groups"`          // List[string]
	DeliveryGroupsPriority    types.Set    `tfsdk:"delivery_groups_priority"` // List[DeliveryGroupPriorityModel]
	ApplicationFolderPath     types.String `tfsdk:"application_folder_path"`
	Icon                      types.String `tfsdk:"icon"`
	LimitVisibilityToUsers    types.Set    `tfsdk:"limit_visibility_to_users"` // Set[string]
	ApplicationCategoryPath   types.String `tfsdk:"application_category_path"`
	Metadata                  types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Tags                      types.Set    `tfsdk:"tags"`     // Set[string]
	Enabled                   types.Bool   `tfsdk:"enabled"`
	MaxTotalInstances         types.Int32  `tfsdk:"max_total_instances"`
	ShortcutAddedToDesktop    types.Bool   `tfsdk:"shortcut_added_to_desktop"`
	ShortcutAddedToStartMenu  types.Bool   `tfsdk:"shortcut_added_to_start_menu"`
	LimitToOneInstancePerUser types.Bool   `tfsdk:"limit_to_one_instance_per_user"`
	Visible                   types.Bool   `tfsdk:"visible"`
	BrowserName               types.String `tfsdk:"browser_name"`
	CpuPriorityLevel          types.String `tfsdk:"cpu_priority_level"`
	HomeZoneMode              types.String `tfsdk:"home_zone_mode"`
	HomeZone                  types.String `tfsdk:"home_zone"`
}

// Schema defines the schema for the data source.
func (ApplicationResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Resource for creating and managing applications.",
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
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"installed_app_properties": InstalledAppResponseModel{}.GetSchema(),
			"application_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The application group IDs to which the application should be added.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"delivery_groups": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The delivery group IDs to which the application should be added." +
					"\n\n-> **Note** The order of delivery group in the `delivery_groups` list determines the priority of the delivery group. Alternatively, you can use the `delivery_groups_priority` attribute to selectively set the priority of delivery groups.",
				Optional: true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
					listvalidator.AtLeastOneOf(
						path.MatchRoot("application_groups"),
						path.MatchRoot("delivery_groups"),
						path.MatchRoot("delivery_groups_priority"),
					),
					listvalidator.ConflictsWith(
						path.MatchRoot("delivery_groups_priority"),
					),
				},
			},
			"delivery_groups_priority": schema.SetNestedAttribute{
				NestedObject: DeliveryGroupPriorityModel{}.GetSchema(),
				Description:  "Set of delivery groups with their corresponding priority.",
				Optional:     true,
			},
			"application_folder_path": schema.StringAttribute{
				Description: "The application folder path in which the application should be created.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"icon": schema.StringAttribute{
				Description: "The Id of the icon to be associated with the application.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("0"),
			},
			"limit_visibility_to_users": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "By default, the application is visible to all users within a delivery group. However, you can restrict its visibility to only certain users by specifying them in the `limit_visibility_to_users` list." +
					"\n\n-> **Note** Users must be in `DOMAIN\\UserOrGroupName` or `user@domain.com` format",
				Optional: true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `Domain\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
				},
			},
			"application_category_path": schema.StringAttribute{
				Description: "The application category path allows users to organize and view applications under specific categories in Citrix Workspace App.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					validator.String(
						stringvalidator.RegexMatches(regexp.MustCompile(util.AppCategoryPathRegex), "the category path must be in the format of `Category1\\Category2` and cannot contain any of the following characters: \"<>|*?:/"),
					),
				},
			},
			"metadata": util.GetMetadataListSchema("Application"),
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the application.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates whether the application is enabled or disabled. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"limit_to_one_instance_per_user": schema.BoolAttribute{
				Description: "Specifies if the use of the application should be limited to only one instance per user. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"max_total_instances": schema.Int32Attribute{
				Description: "Control the use of this application by limiting the number of instances running at the same time. If set to 0, it allows unlimited use.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(0),
			},
			"shortcut_added_to_desktop": schema.BoolAttribute{
				Description: "Indicates whether a shortcut to the application is added to the desktop. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"shortcut_added_to_start_menu": schema.BoolAttribute{
				Description: "Indicates whether a shortcut to the application is added to the start menu. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"visible": schema.BoolAttribute{
				Description: "Specifies whether or not this application is visible to users. Note that it’s possible for an application to be disabled and still visible. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"browser_name": schema.StringAttribute{
				Description: "The browser name for the application. When omitted, the application name will be used as the browser name.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cpu_priority_level": schema.StringAttribute{
				Description: "Specifies the CPU priority level for the application. Valid values are: `Low`, `BelowNormal`, `Normal`, `AboveNormal`, and `High`. Default is `Normal`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Normal"),
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedCpuPriorityLevelEnumValues),
				},
			},
			"home_zone_mode": schema.StringAttribute{
				Description: "Defines the home zone mode for the application. Allowed values are: Prefer, Ignore, Only, User.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.HOMEZONEMODE_USER)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.HOMEZONEMODE_PREFER),
						string(citrixorchestration.HOMEZONEMODE_IGNORE),
						string(citrixorchestration.HOMEZONEMODE_ONLY),
						string(citrixorchestration.HOMEZONEMODE_USER),
					),
				},
			},
			"home_zone": schema.StringAttribute{
				Description: "Specifies the home zone for the application. This can be set using the zone ID.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("00000000-0000-0000-0000-000000000000"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
		},
	}
}

func (ApplicationResourceModel) GetAttributes() map[string]schema.Attribute {
	return ApplicationResourceModel{}.GetSchema().Attributes
}

func (ApplicationResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{
		"limit_visibility_to_users": true,
	}
}

func (r ApplicationResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationDetailResponseModel, applicationGroups *citrixorchestration.ApplicationGroupResponseModelCollection, applicationDeliveryGroups *citrixorchestration.ApplicationDeliveryGroupResponseModelCollection, tags []string) ApplicationResourceModel {
	// Overwrite application with refreshed state
	r.Id = types.StringValue(application.GetId())
	r.Name = types.StringValue(application.GetName())
	r.PublishedName = types.StringValue(application.GetPublishedName())
	r.Description = types.StringValue(application.GetDescription())
	r.Icon = types.StringValue(application.GetIconId())
	r.ApplicationCategoryPath = types.StringValue(application.GetClientFolder())
	r.Enabled = types.BoolValue(application.GetEnabled())
	r.LimitToOneInstancePerUser = types.BoolValue(application.GetMaxPerUserInstances() == 1)
	r.MaxTotalInstances = types.Int32Value(application.GetMaxTotalInstances())
	r.ShortcutAddedToDesktop = types.BoolValue(application.GetShortcutAddedToDesktop())
	r.ShortcutAddedToStartMenu = types.BoolValue(application.GetShortcutAddedToStartMenu())
	r.Visible = types.BoolValue(application.GetVisible())
	r.BrowserName = types.StringValue(application.GetBrowserName())
	r.CpuPriorityLevel = types.StringValue(string(application.GetCpuPriorityLevel()))
	r.HomeZoneMode = types.StringValue(string(application.GetHomeZoneMode()))
	r.HomeZone = types.StringValue(application.HomeZone.GetId())

	// Set optional values
	adminFolder := application.GetApplicationFolder()
	adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
	if adminFolderPath != "" {
		r.ApplicationFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.ApplicationFolderPath = types.StringNull()
	}

	includedUsers := application.GetIncludedUsers()
	if application.GetIncludedUserFilterEnabled() {
		r.LimitVisibilityToUsers = util.RefreshUsersList(ctx, diagnostics, r.LimitVisibilityToUsers, includedUsers)
	} else {
		r.LimitVisibilityToUsers = types.SetNull(types.StringType)
	}

	appGroups := []string{}
	for _, appGroup := range applicationGroups.GetItems() {
		appGroups = append(appGroups, appGroup.GetId())
	}
	r.ApplicationGroups = util.RefreshListValues(ctx, diagnostics, r.ApplicationGroups, appGroups)

	if applicationDeliveryGroups != nil && len(applicationDeliveryGroups.GetItems()) > 0 {
		deliveryGroups := applicationDeliveryGroups.GetItems()
		if !r.DeliveryGroups.IsNull() {
			deliveryGroupsWithPriority := []string{}
			sort.Slice(deliveryGroups, func(i, j int) bool {
				return deliveryGroups[i].GetPriority() < deliveryGroups[j].GetPriority()
			})
			for _, deliveryGroup := range deliveryGroups {
				deliveryGroupsWithPriority = append(deliveryGroupsWithPriority, deliveryGroup.GetId())
			}
			r.DeliveryGroups = util.StringArrayToStringList(ctx, diagnostics, deliveryGroupsWithPriority)
		} else {
			// During create and update, if DeliveryGroups is nil, DeliveryGroupsPriority will be set.
			// During import, both DeliveryGroups and DeliveryGroupsPriority are not set.
			// Set DeliveryGroupsPriority in both cases.
			deliveryGroupsPriority := []DeliveryGroupPriorityModel{}
			for _, deliveryGroup := range deliveryGroups {
				deliveryGroupPriority := DeliveryGroupPriorityModel{
					Id:       types.StringValue(deliveryGroup.GetId()),
					Priority: types.Int32Value(deliveryGroup.GetPriority()),
				}
				deliveryGroupsPriority = append(deliveryGroupsPriority, deliveryGroupPriority)
			}
			r.DeliveryGroupsPriority = util.TypedArrayToObjectSet(ctx, diagnostics, deliveryGroupsPriority)
		}
	} else {
		r.DeliveryGroups = types.ListNull(types.StringType)
	}

	r.InstalledAppProperties = r.updatePlanWithInstalledAppProperties(ctx, diagnostics, application)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), application.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

	return r
}

func (r ApplicationResourceModel) updatePlanWithInstalledAppProperties(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationDetailResponseModel) types.Object {

	var installedAppProperties = InstalledAppResponseModel{}

	installedAppProperties.CommandLineExecutable = types.StringValue(application.InstalledAppProperties.GetCommandLineExecutable())

	// Set optional values
	if application.InstalledAppProperties.GetWorkingDirectory() != "" {
		installedAppProperties.WorkingDirectory = types.StringValue(application.InstalledAppProperties.GetWorkingDirectory())
	}

	if application.InstalledAppProperties.GetCommandLineArguments() != "" {
		installedAppProperties.CommandLineArguments = types.StringValue(application.InstalledAppProperties.GetCommandLineArguments())
	}

	return util.TypedObjectToObjectValue(ctx, diagnostics, installedAppProperties)
}
