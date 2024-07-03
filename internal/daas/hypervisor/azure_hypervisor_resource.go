// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	EnableAzureADDeviceManagement_CustomProperty = "AzureAdDeviceManagement"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &azureHypervisorResource{}
	_ resource.ResourceWithConfigure      = &azureHypervisorResource{}
	_ resource.ResourceWithImportState    = &azureHypervisorResource{}
	_ resource.ResourceWithValidateConfig = &azureHypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewAzureHypervisorResource() resource.Resource {
	return &azureHypervisorResource{}
}

// hypervisorResource is the resource implementation.
type azureHypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *azureHypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_hypervisor"
}

// Schema defines the schema for the resource.
func (r *azureHypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AzureHypervisorResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *azureHypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *azureHypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AzureHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM)

	if plan.ApplicationId.IsNull() || plan.ApplicationSecret.IsNull() || plan.SubscriptionId.IsNull() || plan.ActiveDirectoryId.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor for AzureRM",
			"ApplicationId/ApplicationSecret/SubscriptionId/ActiveDirectoryId is missing.",
		)
		return
	}
	connectionDetails.SetApplicationId(plan.ApplicationId.ValueString())
	connectionDetails.SetApplicationSecret(plan.ApplicationSecret.ValueString())
	metadata := getMetadataForAzureRmHypervisor(plan)
	connectionDetails.SetMetadata(metadata)
	connectionDetails.SetSubscriptionId(plan.SubscriptionId.ValueString())
	connectionDetails.SetActiveDirectoryId(plan.ActiveDirectoryId.ValueString())
	if !plan.Scopes.IsNull() {
		connectionDetails.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}
	// Set custom properties for enabling AzureAD Device Management
	customProperties := []citrixorchestration.NameValueStringPairModel{}
	enableAADDeviceManagementProperty := citrixorchestration.NameValueStringPairModel{}
	enableAADDeviceManagementProperty.SetName(EnableAzureADDeviceManagement_CustomProperty)
	enableAADDeviceManagementProperty.SetValue(strconv.FormatBool(plan.EnableAzureADDeviceManagement.ValueBool()))
	customProperties = append(customProperties, enableAADDeviceManagementProperty)
	customPropertyString, _ := json.Marshal(customProperties)
	connectionDetails.SetCustomProperties(string(customPropertyString))

	// Generate API request body from plan
	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	hypervisor, err := CreateHypervisor(ctx, r.client, &resp.Diagnostics, body)
	if err != nil {
		// Directly return. Error logs have been populated in common function.
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, hypervisor)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *azureHypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AzureHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := state.Id.ValueString()
	hypervisor, err := readHypervisor(ctx, r.client, resp, hypervisorId)
	if err != nil {
		return
	}

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not an Azure connection type hypervisor.",
		)
		return
	}

	// Overwrite hypervisor with refreshed state
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, hypervisor)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *azureHypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AzureHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}
	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not an Azure connection type hypervisor.",
		)
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM)
	editHypervisorRequestBody.SetApplicationId(plan.ApplicationId.ValueString())
	editHypervisorRequestBody.SetApplicationSecret(plan.ApplicationSecret.ValueString())
	metadata := getMetadataForAzureRmHypervisor(plan)
	editHypervisorRequestBody.SetMetadata(metadata)
	if !plan.Scopes.IsNull() {
		editHypervisorRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	// Modify custom properties
	customPropertiesString := hypervisor.GetCustomProperties()
	var customProperties []citrixorchestration.NameValueStringPairModel
	err = json.Unmarshal([]byte(customPropertiesString), &customProperties)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" failed to be retrieved from remote.",
		)
		return
	}

	updatedCustomProperties := []*citrixorchestration.NameValueStringPairModel{}
	for _, customProperty := range customProperties {
		currentProperty := customProperty
		if currentProperty.GetName() == EnableAzureADDeviceManagement_CustomProperty {
			currentProperty.SetValue(strconv.FormatBool(plan.EnableAzureADDeviceManagement.ValueBool()))
		}
		updatedCustomProperties = append(updatedCustomProperties, &currentProperty)
	}

	customPropertiesByte, _ := json.Marshal(updatedCustomProperties)
	editHypervisorRequestBody.SetCustomProperties(string(customPropertiesByte))

	// Fetch updated hypervisor from GetHypervisor
	updatedHypervisor, err := UpdateHypervisor(ctx, r.client, &resp.Diagnostics, hypervisor, editHypervisorRequestBody)
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedHypervisor)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *azureHypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AzureHypervisorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing hypervisor
	hypervisorId := state.Id.ValueString()
	hypervisorName := state.Name.ValueString()
	deleteHypervisorRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisor(ctx, hypervisorId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *azureHypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getMetadataForAzureRmHypervisor(plan AzureHypervisorResourceModel) []citrixorchestration.NameValueStringPairModel {
	secretExpirationDate := "2099-12-31 23:59:59"
	if !plan.ApplicationSecretExpirationDate.IsNull() {
		secretExpirationDate = plan.ApplicationSecretExpirationDate.ValueString()
		secretExpirationDate = secretExpirationDate + " 23:59:59"
	}

	parsedTime, _ := time.Parse(time.DateTime, secretExpirationDate)
	secretExpirationDateInUnix := parsedTime.UnixMilli()
	secretExpirationDateMetada := citrixorchestration.NameValueStringPairModel{}
	secretExpirationDateMetada.SetName("Citrix_Orchestration_Hypervisor_Secret_Expiration_Date")
	secretExpirationDateMetada.SetValue(strconv.Itoa(int(secretExpirationDateInUnix)))
	metadata := []citrixorchestration.NameValueStringPairModel{
		secretExpirationDateMetada,
	}

	return metadata
}

func (r *azureHypervisorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AzureHypervisorResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
