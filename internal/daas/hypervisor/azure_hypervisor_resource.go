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
	EnableAzureADDeviceManagement_CustomProperty          = "AzureAdDeviceManagement"
	ProxyHypervisorTrafficThroughConnector_CustomProperty = "ProxyHypervisorTrafficThroughConnector"
	AuthenticationMode_CustomProperty                     = "AuthenticationMode"
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

	if plan.SubscriptionId.IsNull() || plan.ActiveDirectoryId.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor for AzureRM",
			"SubscriptionId/ActiveDirectoryId is missing.",
		)
		return
	}
	if !plan.ApplicationId.IsNull() {
		connectionDetails.SetApplicationId(plan.ApplicationId.ValueString())
	}
	if !plan.ApplicationSecret.IsNull() {
		connectionDetails.SetApplicationSecret(plan.ApplicationSecret.ValueString())
	}
	metadata := getMetadataForAzureRmHypervisor(plan)
	additionalMetadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	metadata = append(metadata, additionalMetadata...)
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

	enableProxyHypervisorTraffic := citrixorchestration.NameValueStringPairModel{}
	enableProxyHypervisorTraffic.SetName(ProxyHypervisorTrafficThroughConnector_CustomProperty)
	enableProxyHypervisorTraffic.SetValue(strconv.FormatBool(plan.ProxyHypervisorTrafficThroughConnector.ValueBool()))
	customProperties = append(customProperties, enableProxyHypervisorTraffic)

	enableAuthenticationMode := citrixorchestration.NameValueStringPairModel{}
	enableAuthenticationMode.SetName(AuthenticationMode_CustomProperty)
	enableAuthenticationMode.SetValue(plan.AuthenticationMode.ValueString())
	customProperties = append(customProperties, enableAuthenticationMode)

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

	// Get current state
	var state AzureHypervisorResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM)

	editHypervisorRequestBody.SetApplicationId(plan.ApplicationId.ValueString())

	if !plan.ApplicationSecret.IsNull() {
		editHypervisorRequestBody.SetApplicationSecret(plan.ApplicationSecret.ValueString())
	}

	metadata := getMetadataForAzureRmHypervisor(plan)
	additionalMetadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	metadata = append(metadata, additionalMetadata...)
	editHypervisorRequestBody.SetMetadata(metadata)
	if !plan.Scopes.IsNull() {
		editHypervisorRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
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

	isEnableAADeviceManagement := false
	isEnableProxyHyp := false
	isEnableAuthMode := false

	updatedCustomProperties := []*citrixorchestration.NameValueStringPairModel{}
	for _, customProperty := range customProperties {
		currentProperty := customProperty
		if currentProperty.GetName() == EnableAzureADDeviceManagement_CustomProperty {
			currentProperty.SetValue(strconv.FormatBool(plan.EnableAzureADDeviceManagement.ValueBool()))
			isEnableAADeviceManagement = true
		}
		if currentProperty.GetName() == ProxyHypervisorTrafficThroughConnector_CustomProperty {
			currentProperty.SetValue(strconv.FormatBool(plan.ProxyHypervisorTrafficThroughConnector.ValueBool()))
			isEnableProxyHyp = true
		}
		if currentProperty.GetName() == AuthenticationMode_CustomProperty {
			currentProperty.SetValue(plan.AuthenticationMode.ValueString())
			isEnableAuthMode = true
		}
		updatedCustomProperties = append(updatedCustomProperties, &currentProperty)
	}

	if !isEnableAADeviceManagement {
		enableAADDeviceManagementProperty := citrixorchestration.NameValueStringPairModel{}
		enableAADDeviceManagementProperty.SetName(EnableAzureADDeviceManagement_CustomProperty)
		enableAADDeviceManagementProperty.SetValue(strconv.FormatBool(plan.EnableAzureADDeviceManagement.ValueBool()))
		updatedCustomProperties = append(updatedCustomProperties, &enableAADDeviceManagementProperty)
	}

	if !isEnableProxyHyp {
		enableProxyHypervisorTraffic := citrixorchestration.NameValueStringPairModel{}
		enableProxyHypervisorTraffic.SetName(ProxyHypervisorTrafficThroughConnector_CustomProperty)
		enableProxyHypervisorTraffic.SetValue(strconv.FormatBool(plan.ProxyHypervisorTrafficThroughConnector.ValueBool()))
		updatedCustomProperties = append(updatedCustomProperties, &enableProxyHypervisorTraffic)
	}

	if !isEnableAuthMode {
		enableAuthenticationMode := citrixorchestration.NameValueStringPairModel{}
		enableAuthenticationMode.SetName(AuthenticationMode_CustomProperty)
		enableAuthenticationMode.SetValue(plan.AuthenticationMode.ValueString())
		updatedCustomProperties = append(updatedCustomProperties, &enableAuthenticationMode)
	}

	customPropertiesByte, _ := json.Marshal(updatedCustomProperties)
	editHypervisorRequestBody.SetCustomProperties(string(customPropertiesByte))

	// Fetch updated hypervisor from GetHypervisor
	updatedHypervisor, err := UpdateHypervisor(ctx, r.client, &resp.Diagnostics, editHypervisorRequestBody, plan.Id.ValueString(), plan.Name.ValueString())
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
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorRequest, r.client).Async(true).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Hypervisor "+hypervisorName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Hypervisor "+hypervisorName, &resp.Diagnostics, 5, true)
	if err != nil {
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
	secretExpirationDateMetada.SetName(util.MetadataHypervisorSecretExpirationDateName)
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

	if !data.AuthenticationMode.IsNull() && (data.AuthenticationMode.ValueString() == util.UserAssignedManagedIdentity || data.AuthenticationMode.ValueString() == util.SystemAssignedManagedIdentity) {
		if !data.ProxyHypervisorTrafficThroughConnector.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("proxy_hypervisor_traffic_through_connector"),
				"proxy_hypervisor_traffic_through_connector Configuration Error.",
				"proxy_hypervisor_traffic_through_connector should be set to true if the authentication_mode is set to either UserAssignedManagedIdentity or SystemAssignedManagedIdentity.",
			)
		}
		if data.EnableAzureADDeviceManagement.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("enable_azure_ad_device_management"),
				"enable_azure_ad_device_management Configuration Error.",
				"enable_azure_ad_device_management should either not be set or should be set to false if the authentication_mode is set to either UserAssignedManagedIdentity or SystemAssignedManagedIdentity.",
			)
		}
		if !data.ApplicationSecret.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("application_secret"),
				"application_secret Configuration Error.",
				"application_secret should not be set if the authentication_mode is set to either UserAssignedManagedIdentity or SystemAssignedManagedIdentity.",
			)
		}
	} else {
		if data.ApplicationSecret.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("application_secret"),
				"application_secret Configuration Error",
				"application_secret should be set if the authentication_mode is either not set or set to AppClientSecret.",
			)
		}
	}

	if !data.AuthenticationMode.IsNull() && data.AuthenticationMode.ValueString() == util.SystemAssignedManagedIdentity {
		if !data.ApplicationId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("application_id"),
				"application_id Configuration Error.",
				"application_id should not be set if the authentication_mode is set to SystemAssignedManagedIdentity.",
			)
		}
	} else {
		if data.ApplicationId.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("application_id"),
				"application_id Configuration Error.",
				"application_id should be set if the authentication_mode is either not set or is set to one of AppClientSecret or UserAssignedManagedIdentity.",
			)
		}
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *azureHypervisorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
