// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"net/http"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	associatedMachineCatalogs := util.ObjectListToTypedArray[DeliveryGroupMachineCatalogModel](ctx, &resp.Diagnostics, plan.AssociatedMachineCatalogs)
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

	//Create Reboot Schedule after delivery group is created
	var editbody citrixorchestration.EditDeliveryGroupRequestModel
	editbody.SetRebootSchedules(body.GetRebootSchedules())
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
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	if plan.PolicySetId.ValueString() != "" {
		deliveryGroup.SetPolicySetGuid(plan.PolicySetId.ValueString())
	} else {
		deliveryGroup.SetPolicySetGuid(types.StringNull().ValueString())
	}

	if r.client.AuthConfig.OnPremises {
		// DDC 2402 LTSR has a bug where UPN is not returned for AD users. Call Identity API to fetch details for users
		deliveryGroup, deliveryGroupDesktops, _ = updateDeliveryGroupAndDesktopUsers(ctx, r.client, &resp.Diagnostics, deliveryGroup, deliveryGroupDesktops)
		// Do not return if there is an error. We need to set the resource in the state so that tf knows about the resource and marks it tainted (diagnostics already has the error)
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

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

	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

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

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, deliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

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

	// Get refreshed delivery group properties from Orchestration
	deliveryGroupId := plan.Id.ValueString()
	deliveryGroupName := plan.Name.ValueString()
	currentDeliveryGroup, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	editDeliveryGroupRequestBody, err := getRequestModelForDeliveryGroupUpdate(ctx, &resp.Diagnostics, r.client, plan, currentDeliveryGroup)
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
	deliveryGroupMachines, err := getDeliveryGroupMachines(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	//Get reboot schedule
	deliveryGroupRebootSchedule, err := getDeliveryGroupRebootSchedules(ctx, r.client, &resp.Diagnostics, deliveryGroupId)
	if err != nil {
		return
	}

	// Fetch updated delivery group from GetDeliveryGroup.
	updatedDeliveryGroup, err := getDeliveryGroup(ctx, r.client, &resp.Diagnostics, deliveryGroupId)

	if err != nil {
		return
	}

	if plan.PolicySetId.ValueString() != "" {
		updatedDeliveryGroup.SetPolicySetGuid(plan.PolicySetId.ValueString())
	} else {
		updatedDeliveryGroup.SetPolicySetGuid(types.StringNull().ValueString())
	}

	if r.client.AuthConfig.OnPremises {
		// DDC 2402 LTSR has a bug where UPN is not returned for AD users. Call Identity API to fetch details for users
		updatedDeliveryGroup, deliveryGroupDesktops, _ = updateDeliveryGroupAndDesktopUsers(ctx, r.client, &resp.Diagnostics, updatedDeliveryGroup, deliveryGroupDesktops)
		// Do not return if there is an error. We need to set the resource in the state so that tf knows about the resource and marks it tainted (diagnostics already has the error)
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedDeliveryGroup, deliveryGroupDesktops, deliveryGroupPowerTimeSchemes, deliveryGroupMachines, deliveryGroupRebootSchedule)

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

	if data.AssociatedMachineCatalogs.IsNull() || len(data.AssociatedMachineCatalogs.Elements()) < 1 {
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

	if plan.AssociatedMachineCatalogs.IsNull() {
		errorSummary := fmt.Sprintf("Error %s Delivery Group", operation)
		feature := "Delivery Groups without associated machine catalogs"
		isFeatureSupportedForCurrentDDC := util.CheckProductVersion(r.client, &resp.Diagnostics, 118, 7, 42, errorSummary, feature)

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

	associatedMachineCatalogs := util.ObjectListToTypedArray[DeliveryGroupMachineCatalogModel](ctx, &resp.Diagnostics, plan.AssociatedMachineCatalogs)
	associatedMachineCatalogProperties, err := validateAndReturnMachineCatalogSessionSupport(ctx, *r.client, &resp.Diagnostics, associatedMachineCatalogs, !create)
	if err != nil || associatedMachineCatalogProperties.SessionSupport == "" {
		return
	}

	isValid, errMsg := validatePowerManagementSettings(ctx, &resp.Diagnostics, plan, associatedMachineCatalogProperties.SessionSupport)

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
	}
}
