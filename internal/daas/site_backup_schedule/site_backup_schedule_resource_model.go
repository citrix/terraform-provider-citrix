package site_backup_schedule

import (
	"regexp"
	"strconv"
	"time"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SiteBackupScheduleModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Description     types.String `tfsdk:"description"`
	TimeZone        types.String `tfsdk:"timezone"`
	Frequency       types.String `tfsdk:"frequency"`
	StartDate       types.String `tfsdk:"start_date"`
	StartTime       types.String `tfsdk:"start_time"`
	FrequencyFactor types.Int32  `tfsdk:"frequency_factor"`
}

func (SiteBackupScheduleModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Backup schedule configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the backup schedule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the backup schedule.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the backup schedule is enabled. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				Description: "Description of the backup schedule. Defaults to an empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"timezone": schema.StringAttribute{
				Description: "Time zone associated with the backup schedule. Defaults to `GMT Standard Time`. Please refer to the `Timezone` column in the following [table](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/default-time-zones?view=windows-11#time-zones) for allowed values.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("GMT Standard Time"),
				Validators: []validator.String{
					stringvalidator.OneOf(util.AllowedTimeZoneValues...),
				},
			},
			"frequency": schema.StringAttribute{
				Description: "Frequency of the backup schedule. Defaults to `Daily`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixorchestration.BACKUPRESTORESCHEDULEFREQUENCY_DAILY)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Daily",
						"Weekly",
						"Monthly",
					),
				},
			},
			"start_date": schema.StringAttribute{
				Description: "Start date for the backup schedule. Defaults to the current date in UTC. " +
					"\n\n-> **Note** The date format should be YYYY-MM-DD.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(time.Now().UTC().Format(time.DateOnly)),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.DateRegex), "Date must be in the format YYYY-MM-DD"),
				},
			},
			"start_time": schema.StringAttribute{
				Description: "Start time for the backup schedule. " +
					"\n\n-> **Note** The time should be in HH:MM:SS format.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.TimeWithSecondsRegex), "Time must be in HH:MM:SS format and should have a valid value."),
				},
			},
			"frequency_factor": schema.Int32Attribute{
				Description: "Frequency factor for the backup schedule. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(1),
			},
		},
	}
}

func (SiteBackupScheduleModel) GetAttributes() map[string]schema.Attribute {
	return SiteBackupScheduleModel{}.GetSchema().Attributes
}

func (r SiteBackupScheduleModel) RefreshPropertyValues(backupSchedule *citrixorchestration.BackupRestoreScheduleModel) SiteBackupScheduleModel {
	r.Id = types.StringValue(strconv.Itoa(int(backupSchedule.GetUid())))
	r.Name = types.StringValue(backupSchedule.GetName())
	r.Enabled = types.BoolValue(backupSchedule.GetEnabled())
	r.Description = types.StringValue(backupSchedule.GetDescription())
	r.TimeZone = types.StringValue(backupSchedule.GetTimeZoneId())
	r.Frequency = types.StringValue(string(backupSchedule.GetFrequency()))
	r.StartDate = types.StringValue(backupSchedule.GetStartDate())
	r.StartTime = types.StringValue(backupSchedule.GetStartTime())
	r.FrequencyFactor = types.Int32Value(backupSchedule.GetFrequencyFactor())

	return r
}
