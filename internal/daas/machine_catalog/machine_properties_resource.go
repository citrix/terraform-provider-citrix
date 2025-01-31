// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &MachinePropertiesResource{}
	_ resource.ResourceWithConfigure      = &MachinePropertiesResource{}
	_ resource.ResourceWithImportState    = &MachinePropertiesResource{}
	_ resource.ResourceWithValidateConfig = &MachinePropertiesResource{}
	_ resource.ResourceWithModifyPlan     = &MachinePropertiesResource{}
)

// NewAdminFolderResource is a helper function to simplify the provider implementation.
func NewMachinePropertiesResource() resource.Resource {
	return &MachinePropertiesResource{}
}

// MachinePropertiesResource is the resource implementation.
type MachinePropertiesResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *MachinePropertiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_properties"
}

// Configure adds the provider configured client to the data source.
func (r *MachinePropertiesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the data source.
func (r *MachinePropertiesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = MachinePropertiesModel{}.GetSchema()
}

// Create implements resource.Resource.
func (r *MachinePropertiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan MachinePropertiesModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateMachineAndMachineCatalogExistence(ctx, r.client, &resp.Diagnostics, plan)

	machineName := strings.ReplaceAll(plan.Name.ValueString(), "\\", "|")
	err := setMachineTags(ctx, r.client, &resp.Diagnostics, machineName, plan.Tags, "managing tags for")
	if err != nil {
		return
	}

	// Get refreshed admin properties from Orchestration
	machineProperties, err := getMachineProperties(ctx, r.client, &resp.Diagnostics, machineName)
	if err != nil {
		return
	}

	machineTagIds, err := getMachineTagIds(ctx, r.client, &resp.Diagnostics, machineName)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, machineProperties, machineTagIds)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *MachinePropertiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state MachinePropertiesModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateMachineAndMachineCatalogExistence(ctx, r.client, &resp.Diagnostics, state)

	machineName := strings.ReplaceAll(state.Name.ValueString(), "\\", "|")
	// Get refreshed admin properties from Orchestration
	machineProperties, err := readMachineProperties(ctx, r.client, resp, machineName)
	if err != nil {
		return
	}

	machineTagIds, err := getMachineTagIds(ctx, r.client, &resp.Diagnostics, machineName)
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, machineProperties, machineTagIds)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *MachinePropertiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan MachinePropertiesModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateMachineAndMachineCatalogExistence(ctx, r.client, &resp.Diagnostics, plan)

	machineName := strings.ReplaceAll(plan.Name.ValueString(), "\\", "|")
	err := setMachineTags(ctx, r.client, &resp.Diagnostics, machineName, plan.Tags, "updating tags for")
	if err != nil {
		return
	}

	// Get refreshed admin properties from Orchestration
	machineProperties, err := getMachineProperties(ctx, r.client, &resp.Diagnostics, machineName)
	if err != nil {
		return
	}

	machineTagIds, err := getMachineTagIds(ctx, r.client, &resp.Diagnostics, machineName)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, machineProperties, machineTagIds)

	// Set refreshed state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *MachinePropertiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state MachinePropertiesModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	machineName := strings.ReplaceAll(state.Name.ValueString(), "\\", "|")
	getMachinePropertiesRequest := r.client.ApiClient.MachinesAPIsDAAS.MachinesGetMachine(ctx, machineName)
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineDetailResponseModel](getMachinePropertiesRequest, r.client)
	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error checking properties of machine "+state.Name.ValueString(),
			"\nTransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// If machine still exist, remove tags
	err = setMachineTags(ctx, r.client, &resp.Diagnostics, machineName, types.SetNull(types.StringType), "removing tags from")
	if err != nil {
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *MachinePropertiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *MachinePropertiesResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data MachinePropertiesModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *MachinePropertiesResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func readMachineProperties(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, machineNameOrId string) (*citrixorchestration.MachineDetailResponseModel, error) {
	getMachinePropertiesRequest := client.ApiClient.MachinesAPIsDAAS.MachinesGetMachine(ctx, machineNameOrId)
	machineProperties, _, err := util.ReadResource[*citrixorchestration.MachineDetailResponseModel](getMachinePropertiesRequest, ctx, client, resp, "Machine Properties", machineNameOrId)
	return machineProperties, err
}

func getMachineProperties(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineNameOrId string) (*citrixorchestration.MachineDetailResponseModel, error) {
	getMachinePropertiesRequest := client.ApiClient.MachinesAPIsDAAS.MachinesGetMachine(ctx, machineNameOrId)
	machineProperties, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineDetailResponseModel](getMachinePropertiesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading the properties of Machine "+strings.ReplaceAll(machineNameOrId, "|", "\\"),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	return machineProperties, err
}

func getMachineTagIds(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineNameOrId string) ([]string, error) {
	getTagsRequest := client.ApiClient.MachinesAPIsDAAS.MachinesGetMachineTags(ctx, machineNameOrId)
	machineTags, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.TagResponseModelCollection](getTagsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading the tags of Machine "+strings.ReplaceAll(machineNameOrId, "|", "\\"),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}
	tagIds := []string{}
	for _, tag := range machineTags.GetItems() {
		tagIds = append(tagIds, tag.GetId())
	}
	return tagIds, err
}

func setMachineTags(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machineNameOrId string, tags types.Set, tagOperation string) error {
	setMachineTagRequest := client.ApiClient.MachinesAPIsDAAS.MachinesSetMachineTags(ctx, machineNameOrId)
	tagRequestModel := util.ConstructTagsRequestModel(ctx, diagnostics, tags)
	setMachineTagRequest = setMachineTagRequest.TagsRequestModel(tagRequestModel)
	httpResp, err := citrixdaasclient.AddRequestData(setMachineTagRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error %s machine %s", tagOperation, strings.ReplaceAll(machineNameOrId, "|", "\\")),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	return nil
}

func validateMachineAndMachineCatalogExistence(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, machinePropertiesModel MachinePropertiesModel) error {
	machineName := strings.ReplaceAll(machinePropertiesModel.Name.ValueString(), "\\", "|")
	getMachinePropertiesRequest := client.ApiClient.MachinesAPIsDAAS.MachinesGetMachine(ctx, machineName)
	machineProperties, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.MachineDetailResponseModel](getMachinePropertiesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error reading the properties of Machine "+machinePropertiesModel.Name.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}
	machineCatalog := machineProperties.GetMachineCatalog()
	if !strings.EqualFold(machineCatalog.GetId(), machinePropertiesModel.MachineCatalogId.ValueString()) {
		err = fmt.Errorf("Machine catalog ID specified does not match the machine catalog ID of the machine")
		diagnostics.AddError(
			"Error reading the properties of Machine "+machinePropertiesModel.Name.ValueString(),
			err.Error(),
		)
		return err
	}
	return nil
}
