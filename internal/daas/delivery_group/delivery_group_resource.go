// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &deliveryGroupResource{}
	_ resource.ResourceWithConfigure      = &deliveryGroupResource{}
	_ resource.ResourceWithImportState    = &deliveryGroupResource{}
	_ resource.ResourceWithValidateConfig = &deliveryGroupResource{}
	_ resource.ResourceWithModifyPlan     = &deliveryGroupResource{}
)

// NewDeliveryGroupResource is a helper function to simplify the provider implementation.
func NewDeliveryGroupResource() resource.Resource {
	return &deliveryGroupResource{}
}

// deliveryGroupResource is the resource implementation.
type deliveryGroupResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *deliveryGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_group"
}

// Configure adds the provider configured client to the resource.
func (r *deliveryGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema defines the schema for the resource.
func (r *deliveryGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DeliveryGroupResourceModel{}.GetSchema()
}

func (r *deliveryGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get machine catalogs and verify all of them have the same session support
	associatedMachineCatalogs := util.ObjectSetToTypedArray[DeliveryGroupMachineCatalogModel](ctx, &resp.Diagnostics, plan.AssociatedMachineCatalogs)
	associatedMachineCatalogProperties, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, &resp.Diagnostics, associatedMachineCatalogs, true)

	if err != nil {
		return
	}

	if !plan.AutoscaleSettings.IsNull() && !associatedMachineCatalogProperties.IsPowerManaged {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group "+plan.Name.ValueString(),
			"Autoscale settings can only be configured if associated machine catalogs are power managed.",
		)
		return
	}

	if associatedMachineCatalogProperties.IsRemotePcCatalog && !plan.Desktops.IsNull() && len(plan.Desktops.Elements()) > 1 {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group "+plan.Name.ValueString(),
			"Only one assignment policy rule can be added to a Remote PC Delivery Group.",
		)
		return
	}

	if associatedMachineCatalogProperties.IsRemotePcCatalog && !plan.Desktops.IsNull() && len(plan.Desktops.Elements()) > 0 {
		desktops := util.ObjectListToTypedArray[DeliveryGroupDesktop](ctx, &resp.Diagnostics, plan.Desktops)
		if desktops[0].EnableSessionRoaming.ValueBool() {
			resp.Diagnostics.AddError(
				"Error creating Delivery Group "+plan.Name.ValueString(),
				"enable_session_roaming cannot be set to true for Remote PC Delivery Group.",
			)
			return
		}

		if !desktops[0].RestrictedAccessUsers.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Delivery Group "+plan.Name.ValueString(),
				"restricted_access_users needs to be set for Remote PC Delivery Group.",
			)
			return
		}
	}

	body, err := getRequestModelForDeliveryGroupCreate(ctx, &resp.Diagnostics, r.client, plan, associatedMachineCatalogProperties)
	if err != nil {
		return
	}

	createDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsCreateDeliveryGroup(ctx)
	createDeliveryGroupRequest = createDeliveryGroupRequest.CreateDeliveryGroupRequestModel(body)

	// Create new delivery group
	deliveryGroup, httpResp, err := citrixdaasclient.AddRequestData(createDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Delivery Group",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	deliveryGroupId := deliveryGroup.GetId()
	existingAdvancedAccessPolicies := deliveryGroup.GetAdvancedAccessPolicy()

	//Create Reboot Schedule after delivery group is created
	var editbody citrixorchestration.EditDeliveryGroupRequestModel
	// We need to set enabled in the edit request if it is false, as it is ignored in create request
	if !plan.Enabled.ValueBool() {
		editbody.SetEnabled(plan.Enabled.ValueBool())
	}
	editbody.SetRebootSchedules(body.GetRebootSchedules())
	advancedAccessPoliciesRequest := []citrixorchestration.AdvancedAccessPolicyRequestModel{}

	var allowedUsers citrixorchestration.AllowedUser
	if len(existingAdvancedAccessPolicies) > 0 {
		// This should always be true since there are default access policies associated with the dg.
		allowedUsers = existingAdvancedAccessPolicies[0].GetAllowedUsers()
	}

	if !plan.DefaultAccessPolicies.IsNull() {
		simpleAccessPolicy := body.GetSimpleAccessPolicy()
		defaultAccessPolicies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &resp.Diagnostics, plan.DefaultAccessPolicies)
		for _, defaultAccessPolicy := range defaultAccessPolicies {
			advancedAccessPolicyRequest, err := getAdvancedAccessPolicyRequestForDefaultPolicy(ctx, &resp.Diagnostics, defaultAccessPolicy, existingAdvancedAccessPolicies)
			if err != nil {
				return
			}
			advancedAccessPolicyRequest.SetIncludedUserFilterEnabled(simpleAccessPolicy.GetIncludedUserFilterEnabled())
			advancedAccessPolicyRequest.SetIncludedUsers(simpleAccessPolicy.GetIncludedUsers())
			advancedAccessPolicyRequest.SetExcludedUserFilterEnabled(simpleAccessPolicy.GetExcludedUserFilterEnabled())
			advancedAccessPolicyRequest.SetExcludedUsers(simpleAccessPolicy.GetExcludedUsers())
			advancedAccessPolicyRequest.SetAllowedUsers(allowedUsers)
			advancedAccessPoliciesRequest = append(advancedAccessPoliciesRequest, advancedAccessPolicyRequest)
		}

		editbody.SetAdvancedAccessPolicy(advancedAccessPoliciesRequest)
	}

	if !plan.CustomAccessPolicies.IsNull() {
		simpleAccessPolicy := body.GetSimpleAccessPolicy()
		if len(advancedAccessPoliciesRequest) == 0 { // if default policies are not defined by user, then use remote
			for _, existingAdvancedAccessPolicy := range existingAdvancedAccessPolicies {
				var advancedAccessPolicyRequest citrixorchestration.AdvancedAccessPolicyRequestModel
				advancedAccessPolicyRequest.SetId(existingAdvancedAccessPolicy.GetId())
				advancedAccessPolicyRequest.SetName(existingAdvancedAccessPolicy.GetName())
				advancedAccessPolicyRequest.SetIncludedUserFilterEnabled(simpleAccessPolicy.GetIncludedUserFilterEnabled())
				advancedAccessPolicyRequest.SetIncludedUsers(simpleAccessPolicy.GetIncludedUsers())
				advancedAccessPolicyRequest.SetExcludedUserFilterEnabled(simpleAccessPolicy.GetExcludedUserFilterEnabled())
				advancedAccessPolicyRequest.SetExcludedUsers(simpleAccessPolicy.GetExcludedUsers())
				advancedAccessPolicyRequest.SetAllowedUsers(existingAdvancedAccessPolicy.GetAllowedUsers())
				advancedAccessPoliciesRequest = append(advancedAccessPoliciesRequest, advancedAccessPolicyRequest)
			}
		}

		accessPolicies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &resp.Diagnostics, plan.CustomAccessPolicies)
		for _, accessPolicy := range accessPolicies {
			accessPolicyRequest, err := getAdvancedAccessPolicyRequest(ctx, &resp.Diagnostics, accessPolicy)
			if err != nil {
				return
			}
			accessPolicyRequest.SetIncludedUserFilterEnabled(simpleAccessPolicy.GetIncludedUserFilterEnabled())
			accessPolicyRequest.SetIncludedUsers(simpleAccessPolicy.GetIncludedUsers())
			accessPolicyRequest.SetExcludedUserFilterEnabled(simpleAccessPolicy.GetExcludedUserFilterEnabled())
			accessPolicyRequest.SetExcludedUsers(simpleAccessPolicy.GetExcludedUsers())
			accessPolicyRequest.SetAllowedUsers(allowedUsers)
			advancedAccessPoliciesRequest = append(advancedAccessPoliciesRequest, accessPolicyRequest)
		}

		if !plan.AppProtection.IsNull() {
			appProtection := util.ObjectValueToTypedObject[DeliveryGroupAppProtection](ctx, &resp.Diagnostics, plan.AppProtection)
			if !appProtection.ApplyContextually.IsNull() {
				appProtectionApplyContextually := util.ObjectListToTypedArray[DeliveryGroupAppProtectionApplyContextuallyModel](ctx, &resp.Diagnostics, appProtection.ApplyContextually)
				for _, applyContextually := range appProtectionApplyContextually {
					advancedAccessPoliciesRequest, err = setAppProtectionOnAdvancedAccessPolicies(&resp.Diagnostics, applyContextually, advancedAccessPoliciesRequest, deliveryGroup.GetName())
					if err != nil {
						return
					}
				}
			}
		}

		editbody.SetAdvancedAccessPolicy(advancedAccessPoliciesRequest)
	}

	updateDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsPatchDeliveryGroup(ctx, deliveryGroupId)
	updateDeliveryGroupRequest = updateDeliveryGroupRequest.EditDeliveryGroupRequestModel(editbody)
	httpResp, err = citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating reboot schedule for Delivery Group",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	setDeliveryGroupTags(ctx, &resp.Diagnostics, r.client, deliveryGroupId, plan.Tags)

	deliveryGroup, err = util.GetDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	// Get desktops
	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get power time schemes
	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get machines
	deliveryGroupMachines, err := util.GetDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	if r.client.AuthConfig.OnPremises {
		// DDC 2402 LTSR has a bug where UPN is not returned for AD users. Call Identity API to fetch details for users
		deliveryGroup, deliveryGroupDesktops, _ = updateDeliveryGroupAndDesktopUsers(ctx, r.client, &resp.Diagnostics, deliveryGroup, deliveryGroupDesktops)
		// Do not return if there is an error. We need to set the resource in the state so that tf knows about the resource and marks it tainted (diagnostics already has the error)
	}

	tags := getDeliveryGroupTags(ctx, &resp.Diagnostics, r.client, deliveryGroupId)

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule, tags)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deliveryGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state DeliveryGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deliveryGroupId := state.Id.ValueString()
	deliveryGroup, err := readDeliveryGroup(ctx, r.client, resp, deliveryGroupId)
	if err != nil {
		return
	}

	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	deliveryGroupMachines, err := util.GetDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	if deliveryGroup.GetPolicySetGuid() == util.DefaultSitePolicySetId {
		deliveryGroup.SetPolicySetGuid("")
	}

	if r.client.AuthConfig.OnPremises {
		// DDC 2402 LTSR has a bug where UPN is not returned for AD users. Call Identity API to fetch details for users and update dg and dg desktops
		deliveryGroup, deliveryGroupDesktops, err = updateDeliveryGroupAndDesktopUsers(ctx, r.client, &resp.Diagnostics, deliveryGroup, deliveryGroupDesktops)
		if err != nil {
			return
		}
	}

	tags := getDeliveryGroupTags(ctx, &resp.Diagnostics, r.client, deliveryGroupId)

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule, tags)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *deliveryGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get State to check the diff for metadata
	var state DeliveryGroupResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed delivery group properties from Orchestration
	deliveryGroupId := plan.Id.ValueString()
	deliveryGroupName := plan.Name.ValueString()
	currentDeliveryGroup, err := util.GetDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	editDeliveryGroupRequestBody, err := getRequestModelForDeliveryGroupUpdate(ctx, &resp.Diagnostics, r.client, plan, state, currentDeliveryGroup)
	if err != nil {
		return
	}

	updateDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsPatchDeliveryGroup(ctx, deliveryGroupId)
	updateDeliveryGroupRequest = updateDeliveryGroupRequest.EditDeliveryGroupRequestModel(editDeliveryGroupRequestBody)
	httpResp, err := citrixdaasclient.AddRequestData(updateDeliveryGroupRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Add or remove machines
	err = addRemoveMachinesFromDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId, plan)

	if err != nil {
		return
	}

	setDeliveryGroupTags(ctx, &resp.Diagnostics, r.client, deliveryGroupId, plan.Tags)

	// Get desktops
	deliveryGroupDesktops, err := getDeliveryGroupDesktops(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get power time schemes
	deliveryGroupPowerTimeSchemes, err := getDeliveryGroupPowerTimeSchemes(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	// Get machines
	deliveryGroupMachines, err := util.GetDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	// Fetch updated delivery group from GetDeliveryGroup.
	updatedDeliveryGroup, err := util.GetDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	if r.client.AuthConfig.OnPremises {
		// DDC 2402 LTSR has a bug where UPN is not returned for AD users. Call Identity API to fetch details for users
		updatedDeliveryGroup, deliveryGroupDesktops, _ = updateDeliveryGroupAndDesktopUsers(ctx, r.client, &resp.Diagnostics, updatedDeliveryGroup, deliveryGroupDesktops)
		// Do not return if there is an error. We need to set the resource in the state so that tf knows about the resource and marks it tainted (diagnostics already has the error)
	}

	tags := getDeliveryGroupTags(ctx, &resp.Diagnostics, r.client, deliveryGroupId)

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, updatedDeliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule, tags)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Deletes the resource and removes the Terraform state on success.
func (r *deliveryGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state DeliveryGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing delivery group
	deliveryGroupId := state.Id.ValueString()
	deliveryGroupName := state.Name.ValueString()
	deleteDeliveryGroupRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsDeleteDeliveryGroup(ctx, deliveryGroupId)
	httpResp, err := citrixdaasclient.AddRequestData(deleteDeliveryGroupRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Delivery Group "+deliveryGroupName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func (r *deliveryGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *deliveryGroupResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data DeliveryGroupResourceModel
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

	if !data.DefaultAccessPolicies.IsNull() {
		accesPolicies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &resp.Diagnostics, data.DefaultAccessPolicies)

		if len(accesPolicies) != 2 {
			resp.Diagnostics.AddAttributeError(
				path.Root("default_access_policies"),
				"Incorrect Attribute Configuration",
				"In-built policies cannot be added or removed, default_access_policies must have corresponding values for the two in-built policies.",
			)
			return
		}

		isGatewayConnectionPresent := false
		isNonGatewayConnectionPresent := false
		for index, accessPolicy := range accesPolicies {
			isValid := accessPolicy.ValidateConfig(ctx, &resp.Diagnostics, index)
			isValid = isValid && accessPolicy.ValidateConfigForDefaultPolicy(ctx, &resp.Diagnostics, index)
			if !isValid {
				return
			}
			if strings.EqualFold(accessPolicy.Name.ValueString(), util.CitrixGatewayConnections) {
				isGatewayConnectionPresent = true
			} else if strings.EqualFold(accessPolicy.Name.ValueString(), util.NonCitrixGatewayConnections) {
				isNonGatewayConnectionPresent = true
			}
		}

		if !isGatewayConnectionPresent || !isNonGatewayConnectionPresent {
			resp.Diagnostics.AddAttributeError(
				path.Root("default_access_policies"),
				"Incorrect Attribute Configuration",
				"In-built policies cannot be added or removed, default_access_policies must have corresponding values for the two in-built policies. "+
					"\nUse `Citrix Gateway connections` as the name for the default policy that is Via Access Gateway and `Non-Citrix Gateway connections` as the name for the default policy that is Not Via Access",
			)
		}
	}

	if !data.CustomAccessPolicies.IsNull() {
		accessPolicies := util.ObjectListToTypedArray[DeliveryGroupAccessPolicyModel](ctx, &resp.Diagnostics, data.CustomAccessPolicies)
		for index, accessPolicy := range accessPolicies {
			isValid := accessPolicy.ValidateConfig(ctx, &resp.Diagnostics, index)

			if !isValid {
				return
			}
		}
	}

	if !data.AppProtection.IsNull() {
		appPrtoection := util.ObjectValueToTypedObject[DeliveryGroupAppProtection](ctx, &resp.Diagnostics, data.AppProtection)
		isValid := appPrtoection.ValidateConfig(ctx, &resp.Diagnostics)

		if !isValid {
			return
		}
	}

	if !data.AssociatedMachineCatalogs.IsUnknown() &&
		(data.AssociatedMachineCatalogs.IsNull() || len(data.AssociatedMachineCatalogs.Elements()) < 1) {
		// if no machine catalogs are associated, sharing_kind and session_support must be specified

		errorSummary := "Incorrect Attribute Configuration"
		errorDetail := "session_support and sharing_kind must be specified if no machine catalogs are associated."

		if data.SessionSupport.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("session_support"),
				errorSummary,
				errorDetail,
			)

			return
		}

		if data.SharingKind.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("sharing_kind"),
				errorSummary,
				errorDetail,
			)

			return
		}
	}

	if !data.AutoscaleSettings.IsNull() {
		autoscale := util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](ctx, &resp.Diagnostics, data.AutoscaleSettings)
		validatePowerTimeSchemes(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[DeliveryGroupPowerTimeScheme](ctx, &resp.Diagnostics, autoscale.PowerTimeSchemes))

		if autoscale.LogOffWarningMessage.ValueString() != "" && autoscale.LogOffWarningTitle.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Error validating autoscale settings for Delivery Group "+data.Name.ValueString(),
				"`log_off_warning_title` cannot be empty string if `log_off_warning_message` is not empty string.",
			)
			return
		}

		if autoscale.AutoscaleLogOffReminderMessage.ValueString() != "" && autoscale.AutoscaleLogOffReminderTitle.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Error validating autoscale settings for Delivery Group "+data.Name.ValueString(),
				"`log_off_reminder_title` cannot be empty string if `log_off_reminder_message` is not empty string.",
			)
			return
		}
	}

	if !data.RebootSchedules.IsNull() {
		validateRebootSchedules(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[DeliveryGroupRebootSchedule](ctx, &resp.Diagnostics, data.RebootSchedules))
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

}

func (r *deliveryGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan DeliveryGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	operation := "updating"
	if create {
		operation = "creating"
	}

	if r.client.AuthConfig.OnPremises && !plan.DefaultDesktopIcon.IsNull() && plan.DefaultDesktopIcon.ValueString() != "1" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error %s Delivery Group", operation),
			"Customizing the `default_desktop_icon` is not supported for on-premises delivery group desktops.",
		)
		return
	}

	if plan.AssociatedMachineCatalogs.IsNull() {
		errorSummary := fmt.Sprintf("Error %s Delivery Group", operation)
		feature := "Delivery Groups without associated machine catalogs"
		isFeatureSupportedForCurrentDDC := util.CheckProductVersion(r.client, &resp.Diagnostics, 118, 118, 7, 42, errorSummary, feature)

		if !isFeatureSupportedForCurrentDDC {
			return
		}

		if !plan.AutoscaleSettings.IsNull() && !plan.AutoscaleSettings.IsUnknown() {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error %s Delivery Group", operation),
				"Autoscale settings can only be configured if associated machine catalogs are specified.",
			)
		}

		return
	}

	associatedMachineCatalogs := util.ObjectSetToTypedArray[DeliveryGroupMachineCatalogModel](ctx, &resp.Diagnostics, plan.AssociatedMachineCatalogs)
	associatedMachineCatalogProperties, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, &resp.Diagnostics, associatedMachineCatalogs, !create)
	if err != nil || associatedMachineCatalogProperties.SessionSupport == "" {
		return
	}

	// Validate Delivery Type
	if !plan.DeliveryType.IsNull() && !plan.DeliveryType.IsUnknown() && !plan.SessionSupport.IsUnknown() && !plan.SharingKind.IsUnknown() {
		deliveryType := plan.DeliveryType.ValueString()
		sharingKind := plan.SharingKind.ValueString()
		if (associatedMachineCatalogProperties.AllocationType == citrixorchestration.ALLOCATIONTYPE_STATIC || sharingKind == string(citrixorchestration.SHARINGKIND_PRIVATE)) &&
			deliveryType == string(citrixorchestration.DELIVERYKIND_DESKTOPS_AND_APPS) {
			resp.Diagnostics.AddAttributeError(
				path.Root("delivery_type"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("`delivery_type` can only be `%s` or `%s` when allocation type of the associated machine catalog is `%s` or `sharing_kind` is `%s`.", string(citrixorchestration.DELIVERYKIND_DESKTOPS_ONLY), string(citrixorchestration.DELIVERYKIND_APPS_ONLY), string(citrixorchestration.ALLOCATIONTYPE_STATIC), string(citrixorchestration.SHARINGKIND_PRIVATE)),
			)
			return
		}

		if deliveryType == string(citrixorchestration.DELIVERYKIND_APPS_ONLY) && !plan.Desktops.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("delivery_type"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("`delivery_type` cannot be `%s` when `desktops` is specified.", string(citrixorchestration.DELIVERYKIND_APPS_ONLY)),
			)
		}
	}

	isValid, errMsg := validatePowerManagementSettings(ctx, &resp.Diagnostics, plan, associatedMachineCatalogProperties.AllocationType, associatedMachineCatalogProperties.SessionSupport)

	if !isValid {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Error message: "+errMsg,
		)
		return
	}

	if !plan.AutoscaleSettings.IsNull() && !associatedMachineCatalogProperties.IsPowerManaged {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Autoscale settings can only be configured if associated machine catalogs are power managed.",
		)
		return
	}

	if associatedMachineCatalogProperties.IsRemotePcCatalog && plan.Desktops.IsNull() && len(plan.Desktops.Elements()) > 1 {
		resp.Diagnostics.AddError(
			"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
			"Only one assignment policy rule can be added to a Remote PC Delivery Group",
		)
		return
	}

	if associatedMachineCatalogProperties.IsRemotePcCatalog && plan.Desktops.IsNull() && len(plan.Desktops.Elements()) > 0 {
		desktops := util.ObjectListToTypedArray[DeliveryGroupDesktop](ctx, &resp.Diagnostics, plan.Desktops)
		if desktops[0].EnableSessionRoaming.ValueBool() {
			resp.Diagnostics.AddError(
				"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
				"enable_session_roaming cannot be set to true for Remote PC Delivery Group.",
			)
			return
		}

		if !desktops[0].RestrictedAccessUsers.IsNull() {
			resp.Diagnostics.AddError(
				"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
				"restricted_access_users needs to be set for Remote PC Delivery Group.",
			)
			return
		}
	}

	if (associatedMachineCatalogProperties.SessionSupport == citrixorchestration.SESSIONSUPPORT_MULTI_SESSION ||
		associatedMachineCatalogProperties.AllocationType == citrixorchestration.ALLOCATIONTYPE_STATIC ||
		!associatedMachineCatalogProperties.IsPowerManaged) &&
		!plan.MakeResourcesAvailableInLHC.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("make_resources_available_in_lhc"),
			"Incorrect Attribute Configuration",
			"make_resources_available_in_lhc can only be set for power managed Single Session OS Random (pooled) VDAs.",
		)
		return
	}

	if !plan.Desktops.IsNull() {
		sessionRoamingShouldBeSet := true
		if associatedMachineCatalogProperties.AllocationType == citrixorchestration.ALLOCATIONTYPE_STATIC {
			sessionRoamingShouldBeSet = false
		}

		if !plan.SharingKind.IsNull() {
			sharingKind := plan.SharingKind.ValueString()
			if sharingKind == string(citrixorchestration.SHARINGKIND_PRIVATE) {
				sessionRoamingShouldBeSet = false
			}
		}

		desktops := util.ObjectListToTypedArray[DeliveryGroupDesktop](ctx, &resp.Diagnostics, plan.Desktops)
		for _, desktop := range desktops {
			if desktop.EnableSessionRoaming.IsUnknown() {
				continue
			} else if !desktop.EnableSessionRoaming.IsNull() && !sessionRoamingShouldBeSet {
				resp.Diagnostics.AddError(
					"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
					"`enable_session_roaming` cannot be set when `sharing_kind` is `Private` or associated machine catalogs have `Static` allocation type.",
				)
				return
			} else if desktop.EnableSessionRoaming.IsNull() && sessionRoamingShouldBeSet {
				resp.Diagnostics.AddError(
					"Error "+operation+" Delivery Group "+plan.Name.ValueString(),
					"`enable_session_roaming` should be set when `sharing_kind` is `Shared` or associated machine catalogs have `Random` allocation type.",
				)
				return
			}
		}
	}
}
