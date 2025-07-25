// Copyright © 2024. Citrix Systems, Inc.
package cma_catalog

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	catalogservice "github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CitrixManagedCatalogResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	CatalogType         types.String `tfsdk:"catalog_type"`
	Region              types.String `tfsdk:"region"`
	SubscriptionName    types.String `tfsdk:"subscription_name"`
	TemplateImageId     types.String `tfsdk:"template_image_id"`
	MachineSize         types.String `tfsdk:"machine_size"`
	StorageType         types.String `tfsdk:"storage_type"`
	UseManagedDisks     types.Bool   `tfsdk:"use_managed_disks"`
	NumberOfMachines    types.Int64  `tfsdk:"number_of_machines"`
	MaxUsersPerVm       types.Int64  `tfsdk:"max_users_per_vm"`
	MachineNamingScheme types.Object `tfsdk:"machine_naming_scheme"` // MachineNamingSchemeModel
	PowerSchedule       types.Object `tfsdk:"power_schedule"`        // PowerScheduleModel
}

func (CitrixManagedCatalogResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - Citrix Managed Azure --- Manages a Citrix Managed Catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the managed catalog.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the managed catalog.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"catalog_type": schema.StringAttribute{
				Description: "Denotes how the machines in the catalog are allocated to a user. Choose between `MultiSession`, `SingleSessionStatic` and `SingleSessionRandom`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixquickdeploy.AllowedAddCatalogTypeEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The Azure region to deploy the managed catalog.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subscription_name": schema.StringAttribute{
				Description: "The name of the Citrix Managed Azure subscription to deploy the managed catalog. Defaults to `Citrix Managed` if omitted.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Citrix Managed"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_image_id": schema.StringAttribute{
				Description: "The GUID identifier of the template image for creating the managed catalog.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be a valid GUID"),
				},
			},
			"machine_size": schema.StringAttribute{
				Description: "The Azure VM SKU to use for creating machines.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_type": schema.StringAttribute{
				Description: "Storage account type used for provisioned virtual machine disks on Azure. Storage types include: `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						util.StandardLRS,
						util.StandardSSDLRS,
						util.Premium_LRS,
					),
				},
			},
			"use_managed_disks": schema.BoolAttribute{
				Description: "Indicate whether to use Azure managed disks for the provisioned virtual machine. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"number_of_machines": schema.Int64Attribute{
				Description: "Number of VMs that will be provisioned for this catalog. Defaults to `1`.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_users_per_vm": schema.Int64Attribute{
				Description: "Maximum number of concurrent users that could launch session on the same machine. Only allowed to have more than 1 concurrent user when `catalog_type` is `MultiSession`. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"machine_naming_scheme": MachineNamingSchemeModel{}.GetSchema(),
			"power_schedule":        PowerScheduleModel{}.GetSchema(),
		},
	}
}

func (CitrixManagedCatalogResourceModel) GetAttributes() map[string]schema.Attribute {
	return CitrixManagedCatalogResourceModel{}.GetSchema().Attributes
}

func (CitrixManagedCatalogResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

type MachineNamingSchemeModel struct {
	NamingScheme     types.String `tfsdk:"naming_scheme"`
	NamingSchemeType types.String `tfsdk:"naming_scheme_type"`
}

func (MachineNamingSchemeModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Rules specifying how Active Directory machine accounts should be created when machines are provisioned." +
			"\n\n~> **Please Note** When importing a `citrix_quickdeploy_catalog` resource, `machine_naming_scheme` must be omitted in the terraform resource body. Explicitly setting it will result in replacing the quick deploy catalog.",
		Optional: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"naming_scheme": schema.StringAttribute{
				Description: "Defines the template name for AD accounts created in the identity pool.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(15),
				},
			},
			"naming_scheme_type": schema.StringAttribute{
				Description: "Type of naming scheme. This defines the format of the variable part of the AD account names that will be created. Choose between `Numeric` and `Alphabetic`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedAccountNamingSchemeTypeEnumValues),
				},
			},
		},
	}
}

func (MachineNamingSchemeModel) GetAttributes() map[string]schema.Attribute {
	return MachineNamingSchemeModel{}.GetSchema().Attributes
}

func (a MachineNamingSchemeModel) Equals(b MachineNamingSchemeModel) bool {
	return types.String.Equal(a.NamingScheme, b.NamingScheme) && types.String.Equal(a.NamingSchemeType, b.NamingSchemeType)
}

type PowerScheduleModel struct {
	PeakDisconnectedSessionTimeout    types.Int64  `tfsdk:"peak_disconnected_session_timeout"`
	OffPeakDisconnectedSessionTimeout types.Int64  `tfsdk:"off_peak_disconnected_session_timeout"`
	PeakExtendedDisconnectTimeout     types.Int64  `tfsdk:"peak_extended_disconnect_timeout"`
	OffPeakExtendedDisconnectTimeout  types.Int64  `tfsdk:"off_peak_extended_disconnect_timeout"`
	PeakBufferCapacity                types.Int64  `tfsdk:"peak_buffer_capacity"`
	OffPeakBufferCapacity             types.Int64  `tfsdk:"off_peak_buffer_capacity"`
	PeakMinInstances                  types.Int64  `tfsdk:"peak_min_instances"`
	OffPeakMinInstances               types.Int64  `tfsdk:"off_peak_min_instances"`
	PeakDisconnectedSessionAction     types.String `tfsdk:"peak_disconnected_session_action"`
	OffPeakDisconnectedSessionAction  types.String `tfsdk:"off_peak_disconnected_session_action"`
	PeakEndTime                       types.Int64  `tfsdk:"peak_end_time"`
	PeakStartTime                     types.Int64  `tfsdk:"peak_start_time"`
	PeakTimeZoneId                    types.String `tfsdk:"peak_time_zone_id"`
	Weekdays                          types.Set    `tfsdk:"weekdays"`
	PowerOffDelay                     types.Int64  `tfsdk:"power_off_delay"`
}

func (PowerScheduleModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The power management schedule for the Citrix Managed catalog.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"peak_disconnected_session_timeout": schema.Int64Attribute{
				Description: "The number of minutes before the configured action should be performed after a user session disconnects in peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"off_peak_disconnected_session_timeout": schema.Int64Attribute{
				Description: "The number of minutes before the configured action should be performed after a user session disconnectts outside peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"peak_extended_disconnect_timeout": schema.Int64Attribute{
				Description: "The number of minutes before the second configured action should be performed after a user session disconnects in peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"off_peak_extended_disconnect_timeout": schema.Int64Attribute{
				Description: "The number of minutes before the second configured action should be performed after a user session disconnects outside peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"peak_buffer_capacity": schema.Int64Attribute{
				Description: "The percentage of machines in the managed catalog that should be kept available in an idle state in peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"off_peak_buffer_capacity": schema.Int64Attribute{
				Description: "The percentage of machines in the delivery group that should be kept available in an idle state outside peak hours.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"peak_min_instances": schema.Int64Attribute{
				Description: "The minimum number of machines that should be powered on during peak hours. Defaults to `0`. Can only be set to more than `0` if `catalog_type` is `Dedicated`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"off_peak_min_instances": schema.Int64Attribute{
				Description: "The minimum number of machines that should be powered on during off peak hours. Defaults to `0`. Can only be set to more than `0` if `catalog_type` is `Dedicated`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"peak_disconnected_session_action": schema.StringAttribute{
				Description: "The action to be performed after a configurable period of a user session disconnecting in peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionChangeHostingActionEnumValues),
				},
			},
			"off_peak_disconnected_session_action": schema.StringAttribute{
				Description: "The action to be performed after a configurable period of a user session disconnecting outside peak hours. Choose between `Nothing`, `Suspend`, and `Shutdown`. Default is `Nothing`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.SESSIONCHANGEHOSTINGACTION_NOTHING)),
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionChangeHostingActionEnumValues),
				},
			},
			"peak_end_time": schema.Int64Attribute{
				Description: "The end time of peak hours (0-23).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(17),
				Validators: []validator.Int64{
					int64validator.Between(0, 23),
				},
			},
			"peak_start_time": schema.Int64Attribute{
				Description: "The start time of peak hours (0-23).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(9),
				Validators: []validator.Int64{
					int64validator.Between(0, 23),
				},
			},
			"peak_time_zone_id": schema.StringAttribute{
				Description: "The time zone for peak hours. Default is `Eastern Standard Time`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Eastern Standard Time"),
			},
			"weekdays": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The pattern of days of the week that the power time scheme covers.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"sunday",
							"monday",
							"tuesday",
							"wednesday",
							"thursday",
							"friday",
							"saturday",
						),
					),
				},
			},
			"power_off_delay": schema.Int64Attribute{
				Description: "Delay before machines are powered off, when scaling down. Specified in minutes. " +
					"\n\n~> **Please Note** Applies only to multi-session machines. " +
					"\n\n-> **Note** By default, the power-off delay is 30 minutes. You can set it in a range of 0 to 60 minutes. ",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
				Validators: []validator.Int64{
					int64validator.Between(0, 60),
				},
			},
		},
	}
}

func (PowerScheduleModel) GetAttributes() map[string]schema.Attribute {
	return PowerScheduleModel{}.GetSchema().Attributes
}

func (r CitrixManagedCatalogResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, catalog *citrixquickdeploy.CatalogOverview, capacity *citrixquickdeploy.CatalogCapacitySettingsModel, region *catalogservice.DeploymentRegionModel) CitrixManagedCatalogResourceModel {
	r.Id = types.StringValue(catalog.GetId())
	r.Name = types.StringValue(catalog.GetName())
	sessionSupport := catalog.GetSessionSupport()
	allocationType := catalog.GetAllocationType()
	switch sessionSupport {
	case citrixquickdeploy.SESSIONSUPPORT_MULTI_SESSION:
		r.CatalogType = types.StringValue(string(citrixquickdeploy.ADDCATALOGTYPE_MULTI_SESSION))
	case citrixquickdeploy.SESSIONSUPPORT_SINGLE_SESSION:
		switch allocationType {
		case citrixquickdeploy.CATALOGALLOCATIONTYPE_PERMANENT:
			r.CatalogType = types.StringValue(string(citrixquickdeploy.ADDCATALOGTYPE_SINGLE_SESSION_STATIC))
		case citrixquickdeploy.CATALOGALLOCATIONTYPE_RANDOM:
			r.CatalogType = types.StringValue(string(citrixquickdeploy.ADDCATALOGTYPE_SINGLE_SESSION_RANDOM))
		}
	}
	r.SubscriptionName = types.StringValue(catalog.GetSubscriptionName())
	if region == nil || r.shouldSetRegion(*region) {
		// Set region only if region is not set in state, or inconsistent with plan
		// If region is nil, it will not be consistent with plan
		r.Region = types.StringValue(catalog.GetRegion())
	}
	r.TemplateImageId = types.StringValue(catalog.GetImageId())
	computerWorker := capacity.GetComputeWorker()
	r.MachineSize = types.StringValue(computerWorker.GetInstanceTypeId())
	r.StorageType = types.StringValue(string(computerWorker.GetStorageType()))
	r.UseManagedDisks = types.BoolValue(computerWorker.GetUseManagedDisks())
	r.MaxUsersPerVm = types.Int64Value(int64(computerWorker.GetMaxUsersPerVM()))
	r.NumberOfMachines = types.Int64Value(int64(capacity.ScaleSettings.GetMaxInstances()))

	r = r.updatePlanWithPowerSchedule(ctx, diagnostics, capacity)

	return r
}

func (r CitrixManagedCatalogResourceModel) updatePlanWithPowerSchedule(ctx context.Context, diagnostics *diag.Diagnostics, capacity *citrixquickdeploy.CatalogCapacitySettingsModel) CitrixManagedCatalogResourceModel {
	scaleSettings := capacity.GetScaleSettings()
	powerSchedule := util.ObjectValueToTypedObject[PowerScheduleModel](ctx, diagnostics, r.PowerSchedule)

	powerSchedule.PeakDisconnectedSessionTimeout = types.Int64Value(int64(scaleSettings.GetPeakDisconnectedSessionTimeout()))
	powerSchedule.OffPeakDisconnectedSessionTimeout = types.Int64Value(int64(scaleSettings.GetOffPeakDisconnectedSessionTimeout()))
	powerSchedule.PeakExtendedDisconnectTimeout = types.Int64Value(int64(scaleSettings.GetPeakExtendedDisconnectTimeoutMinutes()))
	powerSchedule.OffPeakExtendedDisconnectTimeout = types.Int64Value(int64(scaleSettings.GetOffPeakExtendedDisconnectTimeoutMinutes()))
	powerSchedule.PeakBufferCapacity = types.Int64Value(int64(scaleSettings.GetBufferCapacity()))
	powerSchedule.OffPeakBufferCapacity = types.Int64Value(int64(scaleSettings.GetOffPeakBufferCapacity()))
	powerSchedule.PeakMinInstances = types.Int64Value(int64(scaleSettings.GetPeakMinInstances()))
	powerSchedule.OffPeakMinInstances = types.Int64Value(int64(scaleSettings.GetMinInstances()))
	powerSchedule.PeakDisconnectedSessionAction = types.StringValue(string(scaleSettings.GetPeakDisconnectedSessionAction()))
	powerSchedule.OffPeakDisconnectedSessionAction = types.StringValue(string(scaleSettings.GetOffPeakDisconnectedSessionAction()))
	powerSchedule.PeakStartTime = types.Int64Value(int64(scaleSettings.GetPeakStartTime()))
	powerSchedule.PeakEndTime = types.Int64Value(int64(scaleSettings.GetPeakEndTime()))
	powerSchedule.PeakTimeZoneId = types.StringValue(string(scaleSettings.GetPeakTimeZoneId()))
	powerSchedule.PowerOffDelay = types.Int64Value(int64(scaleSettings.GetPowerOffDelay()))

	weekdayString := scaleSettings.GetWeekdaysString()
	weekdays := []string{}
	if weekdayString != "" {
		weekdays = strings.Split(weekdayString, ",")
	}
	powerSchedule.Weekdays = util.StringArrayToStringSet(ctx, diagnostics, weekdays)

	r.PowerSchedule = util.TypedObjectToObjectValue(ctx, diagnostics, powerSchedule)
	return r
}

func (r CitrixManagedCatalogResourceModel) shouldSetRegion(region citrixquickdeploy.DeploymentRegionModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}

func getTimeSchemeDayValue(v string) citrixorchestration.TimeSchemeDays {
	timeSchemeDay, err := citrixorchestration.NewTimeSchemeDaysFromValue(v)
	if err != nil {
		return citrixorchestration.TIMESCHEMEDAYS_UNKNOWN
	}

	return *timeSchemeDay
}
