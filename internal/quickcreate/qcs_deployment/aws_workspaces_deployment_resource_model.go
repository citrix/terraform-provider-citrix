// Copyright © 2024. Citrix Systems, Inc.
package qcs_deployment

import (
	"context"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesScaleSettingsModel struct {
	SessionIdleTimeoutMinutes       types.Int64 `tfsdk:"disconnect_session_idle_timeout"`
	OffPeakDisconnectTimeoutMinutes types.Int64 `tfsdk:"shutdown_disconnect_timeout"`
	OffPeakLogOffTimeoutMinutes     types.Int64 `tfsdk:"shutdown_log_off_timeout"`
	OffPeakBufferSizePercentage     types.Int64 `tfsdk:"buffer_capacity_size_percentage"`
}

func (AwsWorkspacesScaleSettingsModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Manual power management configuration for the deployment.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"disconnect_session_idle_timeout": schema.Int64Attribute{
				Description: "Indicates timespan before disconnect sessions that are idle in minutes. Defaults to `15`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Default: int64default.StaticInt64(util.DefaultQcsAwsWorkspacesSessionIdleTimeoutMinutes),
			},
			"shutdown_disconnect_timeout": schema.Int64Attribute{
				Description: "Indicates timespan before shut down desktops with disconnected sessions in minutes. Defaults to `15`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Default: int64default.StaticInt64(util.DefaultQcsAwsWorkspacesOffPeakDisconnectTimeoutMinutes),
			},
			"shutdown_log_off_timeout": schema.Int64Attribute{
				Description: "Indicates timespan before shut down desktops after sign-out in minutes. Defaults to `5`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Default: int64default.StaticInt64(util.DefaultQcsAwsWorkspacesOffPeakLogOffTimeoutMinutes),
			},
			"buffer_capacity_size_percentage": schema.Int64Attribute{
				Description: "Indicates buffer capacity size. Defaults to `0`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Default: int64default.StaticInt64(util.DefaultQcsAwsWorkspacesOffPeakBufferSizePercent),
			},
		},
	}
}

func (AwsWorkspacesScaleSettingsModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesScaleSettingsModel{}.GetSchema().Attributes
}

// ensure AwsWorkspacesDeploymentWorkspaceModel implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixquickcreate.AwsEdcDeploymentMachine] = AwsWorkspacesDeploymentWorkspaceModel{}

type AwsWorkspacesDeploymentWorkspaceModel struct {
	Username        types.String `tfsdk:"username"`
	RootVolumeSize  types.Int64  `tfsdk:"root_volume_size"`
	UserVolumeSize  types.Int64  `tfsdk:"user_volume_size"`
	MaintenanceMode types.Bool   `tfsdk:"maintenance_mode"`
	WorkspaceId     types.String `tfsdk:"workspace_id"`
	MachineId       types.String `tfsdk:"machine_id"`
	MachineName     types.String `tfsdk:"machine_name"`
	BrokerMachineId types.String `tfsdk:"broker_machine_id"`
}

func (r AwsWorkspacesDeploymentWorkspaceModel) GetKey() string {
	return r.Username.ValueString()
}

func (AwsWorkspacesDeploymentWorkspaceModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username of the user to be assigned to this workspace. Required if `user_decoupled_workspaces` is set to `false`.",
				Optional:    true,
			},
			"root_volume_size": schema.Int64Attribute{
				Description: "Indicates the root volume size of the workspace.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Any(
						int64validator.OneOf(80),
						int64validator.Between(175, 2000),
					),
				},
			},
			"user_volume_size": schema.Int64Attribute{
				Description: "Indicates the user volume size of the workspace.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Any(
						int64validator.OneOf(10, 50),
						int64validator.Between(100, 2000),
					),
				},
			},
			"maintenance_mode": schema.BoolAttribute{
				Description: "Indicates if the workspace will be set to maintenance mode.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"workspace_id": schema.StringAttribute{
				Description: "Id of the AWS WorkSpaces machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"machine_id": schema.StringAttribute{
				Description: "Id of the machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"machine_name": schema.StringAttribute{
				Description: "Name of the machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"broker_machine_id": schema.StringAttribute{
				Description: "GUID identifier of the machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (AwsWorkspacesDeploymentWorkspaceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDeploymentWorkspaceModel{}.GetSchema().Attributes
}

func (workspace AwsWorkspacesDeploymentWorkspaceModel) RefreshListItem(ctx context.Context, diagnostics *diag.Diagnostics, desktop citrixquickcreate.AwsEdcDeploymentMachine) util.ResourceModelWithAttributes {
	workspace.Username = types.StringValue(desktop.GetUsername())
	workspace.RootVolumeSize = types.Int64Value(int64(desktop.GetRootVolumeSize()))
	workspace.UserVolumeSize = types.Int64Value(int64(desktop.GetUserVolumeSize()))
	workspace.MaintenanceMode = types.BoolValue(desktop.GetMaintenanceMode())
	workspace.WorkspaceId = types.StringValue(desktop.GetWorkspaceId())
	workspace.MachineId = types.StringValue(desktop.GetMachineId())
	workspace.MachineName = types.StringValue(desktop.GetMachineName())
	workspace.BrokerMachineId = types.StringValue(desktop.GetBrokerMachineId())

	return workspace
}

type AwsWorkspacesDeploymentResourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	AccountId               types.String `tfsdk:"account_id"`
	DirectoryId             types.String `tfsdk:"directory_connection_id"`
	ImageId                 types.String `tfsdk:"image_id"`
	Performance             types.String `tfsdk:"performance"`
	RootVolumeSize          types.Int64  `tfsdk:"root_volume_size"`
	UserVolumeSize          types.Int64  `tfsdk:"user_volume_size"`
	VolumesEncrypted        types.Bool   `tfsdk:"volumes_encrypted"`
	VolumesEncryptionKey    types.String `tfsdk:"volumes_encryption_key"`
	RunningMode             types.String `tfsdk:"running_mode"`
	ScaleSettings           types.Object `tfsdk:"scale_settings"` // AwsWorkspacesScaleSettingsModel
	UserDecoupledWorkspaces types.Bool   `tfsdk:"user_decoupled_workspaces"`
	Workspaces              types.List   `tfsdk:"workspaces"` // List[AwsWorkspacesDeploymentWorkspaceModel]
}

func (AwsWorkspacesDeploymentResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Manages an AWS WorkSpaces deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "GUID of the account.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"directory_connection_id": schema.StringAttribute{
				Description: "GUID of the directory.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"image_id": schema.StringAttribute{
				Description: "GUID of the image.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"performance": schema.StringAttribute{
				Description: "Performance of the workspace. Possible Values are `VALUE`, `STANDARD`, `PERFORMANCE`, `POWER`, `POWERPRO`, `GRAPHICS`, `GRAPHICSPRO`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_VALUE),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_STANDARD),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_PERFORMANCE),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_POWER),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_POWERPRO),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_GRAPHICS),
						string(citrixquickcreate.AWSEDCWORKSPACECOMPUTE_GRAPHICSPRO),
					),
				},
			},
			"root_volume_size": schema.Int64Attribute{
				Description: "Size of the root volume in GB. Possible Values are `80` or any integer between `175` and `2000` (inclusive).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Any(
						int64validator.OneOf(80),
						int64validator.Between(175, 2000),
					),
				},
			},
			"user_volume_size": schema.Int64Attribute{
				Description: "Size of the user volume in GB. Possible Values are `10`, `50`, or any integer between `100` and `2000` (inclusive).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Any(
						int64validator.OneOf(10, 50),
						int64validator.Between(100, 2000),
					),
				},
			},
			"volumes_encrypted": schema.BoolAttribute{
				Description: "Indicates if the volumes are encrypted.",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Bool{
					validators.AlsoRequiresOnBoolValues(
						[]bool{true},
						path.MatchRelative().AtParent().AtName("volumes_encryption_key"),
					),
				},
			},
			"volumes_encryption_key": schema.StringAttribute{
				Description: "AWS KMS key to be used for workspace encryption. Use `alias/aws/workspaces` for default AWS KMS workspace encryption key. ",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"running_mode": schema.StringAttribute{
				Description: "Running mode of the workspace. Possible Values are `ALWAYS_ON` and `MANUAL`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_ALWAYS_ON),
						string(citrixquickcreate.AWSEDCWORKSPACERUNNINGMODE_MANUAL),
					),
				},
			},
			"scale_settings": AwsWorkspacesScaleSettingsModel{}.GetSchema(),
			"user_decoupled_workspaces": schema.BoolAttribute{
				Description: "Indicates if the user decoupled workspaces are enabled. ",
				Required:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Bool{
					validators.AlsoRequiresOnBoolValues(
						[]bool{true},
						path.MatchRelative().AtParent().AtName("workspaces"),
					),
				},
			},
			"workspaces": schema.ListNestedAttribute{
				Description:  "Set of workspaces with assigned users.",
				Optional:     true,
				NestedObject: AwsWorkspacesDeploymentWorkspaceModel{}.GetSchema(),
			},
		},
	}
}

func (AwsWorkspacesDeploymentResourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDeploymentResourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesDeploymentResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deployment citrixquickcreate.AwsEdcDeployment) AwsWorkspacesDeploymentResourceModel {
	r.Id = types.StringValue(deployment.GetDeploymentId())
	r.Name = types.StringValue(deployment.GetDeploymentName())
	r.AccountId = types.StringValue(deployment.GetAccountId())
	r.DirectoryId = types.StringValue(deployment.GetConnectionId())
	r.ImageId = types.StringValue(deployment.GetImageId())
	r.Performance = types.StringValue(util.ComputeTypeEnumToString(deployment.GetComputeType()))
	r.RootVolumeSize = types.Int64Value(int64(deployment.GetRootVolumeSize()))
	r.UserVolumeSize = types.Int64Value(int64(deployment.GetUserVolumeSize()))
	r.VolumesEncrypted = types.BoolValue(deployment.GetVolumesEncrypted())
	if !r.VolumesEncryptionKey.IsNull() {
		r.VolumesEncryptionKey = types.StringValue(deployment.GetVolumesEncryptionKey())
	}
	r.RunningMode = types.StringValue(util.RunningModeEnumToString(deployment.GetRunningMode()))

	if !r.ScaleSettings.IsNull() && deployment.HasScaleSettings() {
		scaleSettingsResponse := deployment.GetScaleSettings()
		scaleSettings := AwsWorkspacesScaleSettingsModel{}

		scaleSettings.OffPeakBufferSizePercentage = types.Int64Value(int64(scaleSettingsResponse.GetOffPeakBufferSizePercent()))
		scaleSettings.OffPeakDisconnectTimeoutMinutes = types.Int64Value(int64(scaleSettingsResponse.GetOffPeakDisconnectTimeoutMinutes()))
		scaleSettings.OffPeakLogOffTimeoutMinutes = types.Int64Value(int64(scaleSettingsResponse.GetOffPeakLogOffTimeoutMinutes()))
		scaleSettings.SessionIdleTimeoutMinutes = types.Int64Value(int64(scaleSettingsResponse.GetSessionIdleTimeoutMinutes()))

		scaleSettingsObbject := util.TypedObjectToObjectValue(ctx, diagnostics, scaleSettings)
		r.ScaleSettings = scaleSettingsObbject
	} else {
		if attributes, err := util.ResourceAttributeMapFromObject(AwsWorkspacesScaleSettingsModel{}); err == nil {
			r.ScaleSettings = types.ObjectNull(attributes)
		} else {
			diagnostics.AddError("Error when creating null ScaleSettings", err.Error())
		}
	}

	r.UserDecoupledWorkspaces = types.BoolValue(deployment.GetUserDecoupledWorkspaces())

	if len(deployment.GetWorkspaces()) > 0 {
		if !deployment.GetUserDecoupledWorkspaces() {
			r.Workspaces = util.RefreshListValueProperties[AwsWorkspacesDeploymentWorkspaceModel, citrixquickcreate.AwsEdcDeploymentMachine](ctx, diagnostics, r.Workspaces, deployment.GetWorkspaces(), util.GetQcsAwsWorkspacesWithUsernameKey)
		} else {
			updatedWorkspaces := []AwsWorkspacesDeploymentWorkspaceModel{}
			for _, workspace := range deployment.GetWorkspaces() {
				updatedWorkspaces = append(updatedWorkspaces, AwsWorkspacesDeploymentWorkspaceModel{
					RootVolumeSize:  types.Int64Value(int64(workspace.GetRootVolumeSize())),
					UserVolumeSize:  types.Int64Value(int64(workspace.GetUserVolumeSize())),
					MaintenanceMode: types.BoolValue(workspace.GetMaintenanceMode()),
					WorkspaceId:     types.StringValue(workspace.GetWorkspaceId()),
					MachineId:       types.StringValue(workspace.GetMachineId()),
					MachineName:     types.StringValue(workspace.GetMachineName()),
					BrokerMachineId: types.StringValue(workspace.GetBrokerMachineId()),
				})
			}
			workspacesList := util.TypedArrayToObjectList(ctx, diagnostics, updatedWorkspaces)
			r.Workspaces = workspacesList
		}
	}

	return r
}
