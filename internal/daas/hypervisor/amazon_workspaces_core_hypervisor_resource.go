// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &amazonWorkSpacesCoreHypervisorResource{}
	_ resource.ResourceWithConfigure      = &amazonWorkSpacesCoreHypervisorResource{}
	_ resource.ResourceWithImportState    = &amazonWorkSpacesCoreHypervisorResource{}
	_ resource.ResourceWithValidateConfig = &amazonWorkSpacesCoreHypervisorResource{}
	_ resource.ResourceWithModifyPlan     = &amazonWorkSpacesCoreHypervisorResource{}
)

// NewHypervisorResource is a helper function to simplify the provider implementation.
func NewAmazonWorkSpacesCoreHypervisorResource() resource.Resource {
	return &amazonWorkSpacesCoreHypervisorResource{}
}

// hypervisorResource is the resource implementation.
type amazonWorkSpacesCoreHypervisorResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *amazonWorkSpacesCoreHypervisorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_amazon_workspaces_core_hypervisor"
}

// Schema defines the schema for the resource.
func (r *amazonWorkSpacesCoreHypervisorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AmazonWorkSpacesCoreHypervisorResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *amazonWorkSpacesCoreHypervisorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *amazonWorkSpacesCoreHypervisorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AmazonWorkSpacesCoreHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* Generate ConnectionDetails API request body from plan */
	var connectionDetails citrixorchestration.HypervisorConnectionDetailRequestModel
	connectionDetails.SetName(plan.Name.ValueString())
	connectionDetails.SetZone(plan.Zone.ValueString())
	connectionDetails.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AMAZON_WORK_SPACES_CORE)
	if !plan.Scopes.IsNull() {
		connectionDetails.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}
	connectionDetails.SetRegion(plan.Region.ValueString())
	if plan.UseIamRole.ValueBool() {
		connectionDetails.SetApiKey(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret)
		connectionDetails.SetSecretKey(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret)
	} else {
		connectionDetails.SetApiKey(plan.ApiKey.ValueString())
		connectionDetails.SetSecretKey(plan.SecretKey.ValueString())
	}

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	connectionDetails.SetMetadata(metadata)

	customProperties := []citrixorchestration.NameValueStringPairModel{}
	UseSystemProxyForHypervisorTrafficOnConnectors := citrixorchestration.NameValueStringPairModel{}
	UseSystemProxyForHypervisorTrafficOnConnectors.SetName(UseSystemProxyForHypervisorTrafficOnConnectors_CustomProperty)
	UseSystemProxyForHypervisorTrafficOnConnectors.SetValue(strconv.FormatBool(plan.UseSystemProxyForHypervisorTrafficOnConnectors.ValueBool()))
	customProperties = append(customProperties, UseSystemProxyForHypervisorTrafficOnConnectors)

	customPropertyString, _ := json.Marshal(customProperties)
	connectionDetails.SetCustomProperties(string(customPropertyString))

	// Generate API request body from plan
	var body citrixorchestration.CreateHypervisorRequestModel
	body.SetConnectionDetails(connectionDetails)

	// Create new hypervisor
	hypervisor, err := CreateHypervisor(ctx, r.client, &resp.Diagnostics, body)
	if err != nil {
		// Directly return. Error logs have been populated in common function
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
func (r *amazonWorkSpacesCoreHypervisorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state AmazonWorkSpacesCoreHypervisorResourceModel
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

	if hypervisor.GetConnectionType() != citrixorchestration.HYPERVISORCONNECTIONTYPE_AMAZON_WORK_SPACES_CORE {
		resp.Diagnostics.AddError(
			"Error reading Hypervisor",
			"Hypervisor "+hypervisor.GetName()+" is not an Amazon WorkSpaces Core connection type hypervisor.",
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
func (r *amazonWorkSpacesCoreHypervisorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan AmazonWorkSpacesCoreHypervisorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state AmazonWorkSpacesCoreHypervisorResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construct the update model
	var editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel
	editHypervisorRequestBody.SetName(plan.Name.ValueString())
	editHypervisorRequestBody.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AMAZON_WORK_SPACES_CORE)
	if plan.UseIamRole.ValueBool() {
		editHypervisorRequestBody.SetApiKey(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret)
		editHypervisorRequestBody.SetSecretKey(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret)
	} else {
		editHypervisorRequestBody.SetApiKey(plan.ApiKey.ValueString())
		editHypervisorRequestBody.SetSecretKey(plan.SecretKey.ValueString())
	}
	if !plan.Scopes.IsNull() {
		editHypervisorRequestBody.SetScopes(util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.Scopes))
	}

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorRequestBody.SetMetadata(metadata)

	// Get refreshed hypervisor properties from Orchestration
	hypervisorId := plan.Id.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
	if err != nil {
		return
	}

	// Update custom properties
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

	useConnectorSystemProxy := false
	updatedCustomProperties := []*citrixorchestration.NameValueStringPairModel{}
	for _, customProperty := range customProperties {
		currentProperty := customProperty
		if currentProperty.GetName() == UseSystemProxyForHypervisorTrafficOnConnectors_CustomProperty {
			currentProperty.SetValue(strconv.FormatBool(plan.UseSystemProxyForHypervisorTrafficOnConnectors.ValueBool()))
			useConnectorSystemProxy = true
		}

		updatedCustomProperties = append(updatedCustomProperties, &currentProperty)
	}

	if !useConnectorSystemProxy {
		useSystemProxyForHypervisorTrafficOnConnectors := citrixorchestration.NameValueStringPairModel{}
		useSystemProxyForHypervisorTrafficOnConnectors.SetName(UseSystemProxyForHypervisorTrafficOnConnectors_CustomProperty)
		useSystemProxyForHypervisorTrafficOnConnectors.SetValue(strconv.FormatBool(plan.UseSystemProxyForHypervisorTrafficOnConnectors.ValueBool()))
		updatedCustomProperties = append(updatedCustomProperties, &useSystemProxyForHypervisorTrafficOnConnectors)
	}

	customPropertiesByte, _ := json.Marshal(updatedCustomProperties)
	editHypervisorRequestBody.SetCustomProperties(string(customPropertiesByte))

	// Patch hypervisor
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
func (r *amazonWorkSpacesCoreHypervisorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state AmazonWorkSpacesCoreHypervisorResourceModel
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

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Hypervisor "+hypervisorName, &resp.Diagnostics, 5)
	if err != nil {
		return
	}
}

func (r *amazonWorkSpacesCoreHypervisorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *amazonWorkSpacesCoreHypervisorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AmazonWorkSpacesCoreHypervisorResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	if !data.ApiKey.IsUnknown() || !data.UseIamRole.IsUnknown() || !data.SecretKey.IsUnknown() {
		if data.ApiKey.IsNull() && !data.UseIamRole.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_key"),
				"Invalid Attribute Configuration",
				"Either `use_iam_role` must be set to `true` or values must be provided for `api_key` and `secret_key`. Please update the configuration and try again.",
			)
		}
		if !data.UseIamRole.IsNull() && data.UseIamRole.ValueBool() {
			if !data.ApiKey.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("api_key"),
					"Invalid Attribute Configuration",
					"The `api_key` attribute cannot be set when `use_iam_role` is set to `true`. Please either remove the `api_key` attribute or set `use_iam_role` to `false`.",
				)
			}
			if !data.SecretKey.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("secret_key"),
					"Invalid Attribute Configuration",
					"The `secret_key` attribute cannot be set when `use_iam_role` is set to `true`. Please either remove the `secret_key` attribute or set `use_iam_role` to `false`.",
				)
			}
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *amazonWorkSpacesCoreHypervisorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Check if the DDC version supports Amazon WorkSpaces Core hypervisors
	isDdcVersionSupported := !r.client.AuthConfig.OnPremises && r.client.ClientConfig.OrchestrationApiVersion >= util.DDCVersion125
	if !isDdcVersionSupported {
		resp.Diagnostics.AddError(
			"Unsupported DDC Version",
			"The current DDC version does not support creating Amazon WorkSpaces Core hypervisors.",
		)
		return
	}
}
