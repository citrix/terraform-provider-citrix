package site_backup_schedule

import (
	"context"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &SiteBackupScheduleResource{}
	_ resource.ResourceWithConfigure      = &SiteBackupScheduleResource{}
	_ resource.ResourceWithImportState    = &SiteBackupScheduleResource{}
	_ resource.ResourceWithValidateConfig = &SiteBackupScheduleResource{}
	_ resource.ResourceWithModifyPlan     = &SiteBackupScheduleResource{}
)

func NewSiteBackupScheduleResource() resource.Resource {
	return &SiteBackupScheduleResource{}
}

type SiteBackupScheduleResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (r *SiteBackupScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_backup_schedule"
}

func (r *SiteBackupScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *SiteBackupScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SiteBackupScheduleModel{}.GetSchema()
}

func (r *SiteBackupScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan SiteBackupScheduleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Site Backup Schedule resources cannot have the same name. Check for naming conflicts.
	namingConflictExists, err := checkNamingConflict(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
	if err != nil || namingConflictExists {
		return
	}

	var body citrixorchestration.BackupRestoreScheduleRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetEnabled(plan.Enabled.ValueBool())
	body.SetDescription(plan.Description.ValueString())
	body.SetTimeZoneId(plan.TimeZone.ValueString())
	body.SetFrequency(citrixorchestration.BackupRestoreScheduleFrequency(plan.Frequency.ValueString()))
	body.SetStartDate(plan.StartDate.ValueString())
	body.SetStartTime(plan.StartTime.ValueString())
	body.SetFrequencyFactor(plan.FrequencyFactor.ValueInt32())

	backupScheduleModelRequest := r.client.ApiClient.BackupRestoreAPIsDAAS.BackupRestoreCreateBackupSchedule(ctx)
	backupScheduleModelRequest = backupScheduleModelRequest.BackupRestoreScheduleRequestModel(body)

	// Create BackupRestoreSchedule
	backupScheduleResponseModel, httpResp, err := citrixdaasclient.AddRequestData(backupScheduleModelRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Backup Schedule",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get the backup restore schedule from remote
	id := backupScheduleResponseModel.GetUid()
	backupSchedule, err := getBackupSchedule(ctx, r.client, &resp.Diagnostics, id)
	if err != nil {
		return
	}

	// Refresh the plan with newly assigned values and set Terraform state
	plan = plan.RefreshPropertyValues(backupSchedule)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteBackupScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state SiteBackupScheduleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the backup restore schedule from remote
	uid := convertStringToInt32(state.Id.ValueString())
	backupSchedule, err := getBackupSchedule(ctx, r.client, &resp.Diagnostics, uid)
	if err != nil {
		return
	}

	// Set refreshed state
	state = state.RefreshPropertyValues(backupSchedule)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteBackupScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state SiteBackupScheduleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve updated values from plan
	var plan SiteBackupScheduleModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This means that the name of the backup schedule is being updated.
	// Check for naming conflicts.
	if state.Name != plan.Name {
		namingConflictExists, err := checkNamingConflict(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString())
		if err != nil || namingConflictExists {
			return
		}
	}

	var body citrixorchestration.BackupRestoreScheduleRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetEnabled(plan.Enabled.ValueBool())
	body.SetDescription(plan.Description.ValueString())
	body.SetTimeZoneId(plan.TimeZone.ValueString())
	body.SetFrequency(citrixorchestration.BackupRestoreScheduleFrequency(plan.Frequency.ValueString()))
	body.SetStartDate(plan.StartDate.ValueString())
	body.SetStartTime(plan.StartTime.ValueString())
	body.SetFrequencyFactor(plan.FrequencyFactor.ValueInt32())

	// Get String ID and convert it into int32
	uid := convertStringToInt32(state.Id.ValueString())

	modifyBackupScheduleRequest := r.client.ApiClient.BackupRestoreAPIsDAAS.BackupRestoreModifyBackupSchedule(ctx, uid)
	modifyBackupScheduleRequest = modifyBackupScheduleRequest.BackupRestoreScheduleRequestModel(body)

	// Modify BackupRestoreSchedule
	modifyBackupScheduleResponse, httpResp, err := citrixdaasclient.AddRequestData(modifyBackupScheduleRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Backup Schedule",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get the updated backup restore schedule from remote
	id := modifyBackupScheduleResponse.GetUid()
	modifyBackupSchedule, err := getBackupSchedule(ctx, r.client, &resp.Diagnostics, id)
	if err != nil {
		return
	}

	// Refresh the plan with the updated values and set Terraform state
	plan = plan.RefreshPropertyValues(modifyBackupSchedule)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SiteBackupScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state SiteBackupScheduleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the backup restore schedule
	uid := convertStringToInt32(state.Id.ValueString())
	deleteBackupScheduleRequest := r.client.ApiClient.BackupRestoreAPIsDAAS.BackupRestoreDeleteBackupSchedule(ctx, uid)
	httpResp, err := citrixdaasclient.AddRequestData(deleteBackupScheduleRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Backup Schedule with Id: "+string(uid),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *SiteBackupScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SiteBackupScheduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data SiteBackupScheduleModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *SiteBackupScheduleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func getBackupSchedule(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, id int32) (*citrixorchestration.BackupRestoreScheduleModel, error) {
	getBackupScheduleRequest := client.ApiClient.BackupRestoreAPIsDAAS.BackupRestoreGetBackupSchedule(ctx, id)
	backupRestoreScheduleModel, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.BackupRestoreScheduleModel](getBackupScheduleRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error getting Backup Schedule",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return backupRestoreScheduleModel, nil
}

// The method converts the string Id into int32 value for the API call.
// If the conversion fails, it returns -1 as a default value.
func convertStringToInt32(value string) int32 {
	intVal, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return -1
	}
	return int32(intVal)
}

func getAllBackupSchedules(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.BackupRestoreScheduleModel, error) {
	getAllBackupSchedulesRequest := client.ApiClient.BackupRestoreAPIsDAAS.BackupRestoreGetBackupSchedules(ctx)
	backupRestoreScheduleModelList, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.BackupRestoreScheduleModelCollection](getAllBackupSchedulesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error getting all Backup Schedules",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	return backupRestoreScheduleModelList.GetItems(), nil
}

// This function checks if a backup schedule with the same name already exists in the list of backup schedules.
func checkNamingConflict(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, name string) (bool, error) {
	// Get all the backup schedules from remote to check for naming conflicts.
	backupSchedulesList, err := getAllBackupSchedules(ctx, client, diagnostics)
	if err != nil {
		return false, err
	}

	for _, backupScheduleItem := range backupSchedulesList {
		if strings.EqualFold(backupScheduleItem.GetName(), name) {
			diagnostics.AddError("Naming conflict in Backup Schedule",
				"The backup schedule name "+name+" already exists. Please choose a different name.",
			)
			return true, nil
		}
	}
	return false, nil
}
