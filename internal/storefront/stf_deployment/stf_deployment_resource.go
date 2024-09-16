// Copyright © 2024. Citrix Systems, Inc.
package stf_deployment

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &stfDeploymentResource{}
	_ resource.ResourceWithConfigure      = &stfDeploymentResource{}
	_ resource.ResourceWithImportState    = &stfDeploymentResource{}
	_ resource.ResourceWithValidateConfig = &stfDeploymentResource{}
)

// stfDeploymentResource is a helper function to simplify the provider implementation.
func NewSTFDeploymentResource() resource.Resource {
	return &stfDeploymentResource{}
}

// stfDeploymentResource is the resource implementation.
type stfDeploymentResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (*stfDeploymentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data STFDeploymentResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate Roaming Beacon Configuration: It can only exist if Roaming Gateway is present
	if !data.RoamingBeacon.IsNull() && data.RoamingGateway.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("roaming_beacon"),
			"Roaming Beacon Configuration Error",
			"Roaming Beacon can only be configured if Roaming Gateway is present",
		)
	}

	// Validate STA URLs: It can only exist if Roaming Gateway is present
	if !data.RoamingGateway.IsNull() {
		roamingGateways := util.ObjectListToTypedArray[RoamingGateway](ctx, &resp.Diagnostics, data.RoamingGateway)

		for _, roamingGateway := range roamingGateways {
			if !roamingGateway.SecureTicketAuthorityUrls.IsNull() {
				staUrlList := util.ObjectListToTypedArray[STFSecureTicketAuthority](ctx, &resp.Diagnostics, roamingGateway.SecureTicketAuthorityUrls)
				for _, staUrl := range staUrlList {
					if staUrl.StaValidationEnabled.ValueBool() {
						if staUrl.StaValidationSecret.IsNull() {
							resp.Diagnostics.AddAttributeError(
								path.Root("secure_ticket_authority_urls"),
								"Incorrect Attribute Configuration",
								"STA Validation Secret is required when STA Validation is enabled",
							)
						}
					} else if !staUrl.StaValidationSecret.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("secure_ticket_authority_urls"),
							"Incorrect Attribute Configuration",
							"STA Validation Secret should be empty when STA Validation is disabled",
						)
					}
				}
			}
		}
	}
	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Metadata returns the resource type name.
func (r *stfDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_deployment"
}

// Schema defines the schema for the resource.
func (r *stfDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = STFDeploymentResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *stfDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	// Retrieve values from plan
	var plan STFDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.CreateSTFDeploymentRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing site_id ",
			"Error message: "+err.Error(),
		)
		return
	}

	body.SetSiteId(siteIdInt)
	body.SetHostBaseUrl(plan.HostBaseUrl.ValueString())

	createDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentCreateSTFDeployment(ctx, body)

	// Create new STF Deployment
	DeploymentDetail, err := createDeploymentRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Deployment",
			"Error message: "+err.Error(),
		)
		return
	}

	//Create Roaming Gateway

	var gateway []citrixstorefront.STFRoamingGatewayResponseModel
	if !plan.RoamingGateway.IsNull() {
		setRoamingGateway(ctx, r.client, &resp.Diagnostics, gateway, plan)

		gateway, err = getRoamingGateway(ctx, r.client, &resp.Diagnostics, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching Roaming Gateway",
				"Error message: "+err.Error(),
			)
		}
	}

	// Map response body to schema and populate Computed attribute values
	var getRoamingBeaconInternalResponse *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel
	var getRoamingBeaconExternalResponse *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel
	if len(gateway) > 0 && !plan.RoamingBeacon.IsNull() {
		err = setRoamingBeacon(ctx, r.client, &resp.Diagnostics, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}

		getRoamingBeaconInternalResponse, err = getRoamingBeaconInternal(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching Internal Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}

		getRoamingBeaconExternalResponse, err = getRoamingBeaconExternal(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error fetching External Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &DeploymentDetail, gateway, getRoamingBeaconInternalResponse, getRoamingBeaconExternalResponse)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stfDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil {
		return
	}
	if deployment == nil {
		resp.Diagnostics.AddWarning(
			"StoreFront Deployment not found",
			"StoreFront Deployment was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
		)
		resp.State.RemoveResource(ctx)
	}

	gateway, err := getRoamingGateway(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Internal Roaming Beacon",
			"Error message: "+err.Error(),
		)
	}

	var getRoamingBeaconInternalResponse *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel
	var getRoamingBeaconExternalResponse *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel

	getRoamingBeaconInternalResponse, err = getRoamingBeaconInternal(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Internal Roaming Beacon",
			"Error message: "+err.Error(),
		)
	}

	getRoamingBeaconExternalResponse, err = getRoamingBeaconExternal(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching External Roaming Beacon",
			"Error message: "+err.Error(),
		)
	}
	state.RefreshPropertyValues(ctx, &resp.Diagnostics, deployment, gateway, getRoamingBeaconInternalResponse, getRoamingBeaconExternalResponse)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFDeploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed STFDeployment
	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	// Construct the update model
	var editSTFDeploymentBody citrixstorefront.SetSTFDeploymentRequestModel
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching state of StoreFront Authentication Service ",
			"Error message: "+err.Error(),
		)
	}
	editSTFDeploymentBody.SetSiteId(siteIdInt)
	editSTFDeploymentBody.SetHostBaseUrl(plan.HostBaseUrl.ValueString())

	// Update STFDeployment
	editDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentSetSTFDeployment(ctx, editSTFDeploymentBody)
	_, err = editDeploymentRequest.Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
	}

	// Fetch updated STFDeployment
	updatedSTFDeployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting updated StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
	}

	// Update Roaming Gateway
	var state STFDeploymentResourceModel
	req.State.Get(ctx, &state)
	existingGateways, err := getRoamingGateway(ctx, r.client, &resp.Diagnostics, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching existing Gateways",
			"Error message: "+err.Error(),
		)
	}
	setRoamingGateway(ctx, r.client, &resp.Diagnostics, existingGateways, plan)

	gateways, err := getRoamingGateway(ctx, r.client, &resp.Diagnostics, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting updated StoreFront Roaming Gateway ",
			"Error message: "+err.Error(),
		)
	}

	// Update resource state with updated property values
	var getRoamingBeaconInternalResponse *citrixstorefront.GetSTFRoamingInternalBeaconResponseModel
	var getRoamingBeaconExternalResponse *citrixstorefront.GetSTFRoamingExternalBeaconResponseModel
	if len(gateways) > 0 && !plan.RoamingBeacon.IsNull() {
		err = setRoamingBeacon(ctx, r.client, &resp.Diagnostics, plan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}

		getRoamingBeaconInternalResponse, err = getRoamingBeaconInternal(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Internal Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}

		getRoamingBeaconExternalResponse, err = getRoamingBeaconExternal(ctx, r.client, &resp.Diagnostics, plan)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating External Roaming Beacon",
				"Error message: "+err.Error(),
			)
		}
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updatedSTFDeployment, gateways, getRoamingBeaconInternalResponse, getRoamingBeaconExternalResponse)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFDeploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if STFDeployment exists
	deployment, err := GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	// Delete Roaming Gateway
	var roamingService citrixstorefront.STFRoamingServiceRequestModel
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting Site Id associated with STF Roaming Gateway ",
				"Error message: "+err.Error(),
			)
			return
		}
		roamingService.SetSiteId(siteIdInt)
	}

	var existingGateways citrixstorefront.GetSTFRoamingGatewayRequestModel

	if !state.RoamingGateway.IsNull() {
		roamingGateways := util.ObjectListToTypedArray[RoamingGateway](ctx, &resp.Diagnostics, state.RoamingGateway)
		for _, roamingGateway := range roamingGateways {
			existingGateways.SetName(roamingGateway.Name.ValueString())
			// Delete Roaming Gateway
			deleteRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayRemove(ctx, existingGateways, roamingService)
			err = deleteRoamingGatewayRequest.Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error deleting StoreFront Roaming Gateway ",
					"Error message: "+err.Error(),
				)
				return
			}
		}
	}

	// Delete Roaming Beacon
	if !state.RoamingBeacon.IsNull() {
		// Delete Roaming Beacon
		deleteRoamingBeaconRequest := r.client.StorefrontClient.RoamingSF.STFRoamingBeaconInternalRemove(ctx, roamingService)
		err = deleteRoamingBeaconRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting StoreFront Roaming Beacon ",
				"Error message: "+err.Error(),
			)
			return
		}
	}

	// Delete existing STF Deployment
	var body citrixstorefront.ClearSTFDeploymentRequestModel
	if state.SiteId.ValueString() != "" {
		siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting StoreFront Deployment ",
				"Error message: "+err.Error(),
			)
			return
		}
		body.SetSiteId(siteIdInt)
	}
	deleteDeploymentRequest := r.client.StorefrontClient.DeploymentSF.STFDeploymentClearSTFDeployment(ctx, body)
	_, err = deleteDeploymentRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Deployment ",
			"Error message: "+err.Error(),
		)
		return
	}
}

func (r *stfDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("site_id"), req, resp)
}

// Gets the STFDeployment and logs any errors
func GetSTFDeployment(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId *string) (*citrixstorefront.STFDeploymentDetailModel, error) {
	var body citrixstorefront.GetSTFDeploymentRequestModel
	if siteId != nil {
		siteIdInt, err := strconv.ParseInt(*siteId, 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of StoreFront Deployment ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}
	getSTFDeploymentRequest := client.StorefrontClient.DeploymentSF.STFDeploymentGetSTFDeployment(ctx, body)

	// Get refreshed STFDeployment properties from Orchestration
	STFDeployment, err := getSTFDeploymentRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			return nil, nil
		}
		return &STFDeployment, err
	}
	return &STFDeployment, nil
}

func getRoamingGateway(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFDeploymentResourceModel) ([]citrixstorefront.STFRoamingGatewayResponseModel, error) {
	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error getting Deployment SiteId ",
			"Error message: "+err.Error(),
		)
		return nil, err
	}

	roamingGateways := util.ObjectListToTypedArray[RoamingGateway](ctx, diagnostics, plan.RoamingGateway)
	getRoamingServiceBody.SetSiteId(siteIdInt)
	var gateways []citrixstorefront.STFRoamingGatewayResponseModel
	for _, roamingGateway := range roamingGateways {
		// Retrieve values from response
		var getRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel

		getRoamingGatewayBody.SetName(roamingGateway.Name.ValueString())

		// Get STF Roaming Gateway details
		getRoamingGatewayRequest := client.StorefrontClient.RoamingSF.STFRoamingGatewayGet(ctx, getRoamingGatewayBody, getRoamingServiceBody)
		remoteRoamingGateway, err := getRoamingGatewayRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error getting StoreFront Roaming Gateway details ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		gateways = append(gateways, remoteRoamingGateway)
	}
	return gateways, err
}

func getRoamingBeaconInternal(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFDeploymentResourceModel) (*citrixstorefront.GetSTFRoamingInternalBeaconResponseModel, error) {
	// Get the Internal IPs
	getRoamingIntBeaconRequest := client.StorefrontClient.RoamingSF.GetRoamingInternalBeacon(ctx)
	remoteRoamingInternalBeacon, err := getRoamingIntBeaconRequest.Execute()

	if err != nil {
		diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for Internal IP Addresses",
			"Error message: "+err.Error(),
		)
	}
	return &remoteRoamingInternalBeacon, nil
}

func getRoamingBeaconExternal(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFDeploymentResourceModel) (*citrixstorefront.GetSTFRoamingExternalBeaconResponseModel, error) {
	// Get the External IPs
	getRoamingExtBeaconRequest := client.StorefrontClient.RoamingSF.GetRoamingExternalBeacon(ctx)
	remoteRoamingExternalBeacon, err := getRoamingExtBeaconRequest.Execute()

	if err != nil {
		diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for External IP Addresses",
			"Error message: "+err.Error(),
		)
	}
	return &remoteRoamingExternalBeacon, nil
}

func setRoamingBeacon(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, plan STFDeploymentResourceModel) error {
	var roamingBeaconInternalBody citrixstorefront.SetSTFRoamingInternalBeaconRequestModel

	plannedRoamingBeacon := util.ObjectValueToTypedObject[RoamingBeacon](ctx, diagnostics, plan.RoamingBeacon)
	roamingBeaconInternalBody.SetInternal(plannedRoamingBeacon.Internal.ValueString())

	// Set STF Roaming Gateway
	if !plannedRoamingBeacon.External.IsNull() {
		var roamingBeaconExternalBody citrixstorefront.SetSTFRoamingExternalBeaconRequestModel
		roamingBeaconExternalBody.SetExternal(util.StringListToStringArray(ctx, diagnostics, plannedRoamingBeacon.External))

		roamingBeaconRequest := client.StorefrontClient.RoamingSF.SetRoamingExternalBeacon(ctx, roamingBeaconExternalBody, roamingBeaconInternalBody)
		err := roamingBeaconRequest.Execute()

		if err != nil {
			diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for External IP addresses and Internal IP Address",
				"Error message: "+err.Error(),
			)
			return err
		}
	} else {
		// Set STF Roaming Gateway
		roamingBeaconRequest := client.StorefrontClient.RoamingSF.SetRoamingInternalBeacon(ctx, roamingBeaconInternalBody)
		err := roamingBeaconRequest.Execute()
		if err != nil {
			diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for Internal IP adrress",
				"Error message: "+err.Error(),
			)
			return err
		}
	}
	return nil
}

func buildRoamingGatewayBody(roamingGatewayPlan RoamingGateway) citrixstorefront.AddSTFRoamingGatewayRequestModel {
	var addRoamingGatewayBody citrixstorefront.AddSTFRoamingGatewayRequestModel
	var diagnostics *diag.Diagnostics
	if !roamingGatewayPlan.Name.IsNull() {
		addRoamingGatewayBody.SetName(roamingGatewayPlan.Name.ValueString())
	}
	if !roamingGatewayPlan.LogonType.IsNull() {
		includedLogonType, err := models.NewLogonTypeFromValue(roamingGatewayPlan.LogonType.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Error updating Logon Type",
				fmt.Sprintf("Unsupported criteria type %s.", roamingGatewayPlan.LogonType.ValueString()),
			)
			return addRoamingGatewayBody
		}
		addRoamingGatewayBody.SetLogonType(*includedLogonType)
	}
	if !roamingGatewayPlan.SmartCardFallbackLogonType.IsNull() {
		includedSmartCardFallbackLogonType, err := citrixstorefront.NewLogonTypeFromValue(roamingGatewayPlan.SmartCardFallbackLogonType.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Error updating Smartcard Fallback Logon Type",
				fmt.Sprintf("Unsupported criteria type %s.", roamingGatewayPlan.SmartCardFallbackLogonType.ValueString()),
			)
			return addRoamingGatewayBody
		}
		addRoamingGatewayBody.SetSmartCardFallbackLogonType(*includedSmartCardFallbackLogonType)
	}
	if !roamingGatewayPlan.GatewayUrl.IsNull() {
		addRoamingGatewayBody.SetGatewayUrl(roamingGatewayPlan.GatewayUrl.ValueString())
	}
	if !roamingGatewayPlan.Version.IsNull() {
		addRoamingGatewayBody.SetVersion(roamingGatewayPlan.Version.ValueString())
	}
	if !roamingGatewayPlan.SubnetIPAddress.IsNull() {
		addRoamingGatewayBody.SetSubnetIPAddress(roamingGatewayPlan.SubnetIPAddress.ValueString())
	}
	if !roamingGatewayPlan.StasBypassDuration.IsNull() {
		addRoamingGatewayBody.SetStasBypassDuration(roamingGatewayPlan.StasBypassDuration.ValueString())
	}
	if !roamingGatewayPlan.GslbUrl.IsNull() {
		addRoamingGatewayBody.SetGslbUrl(roamingGatewayPlan.GslbUrl.ValueString())
	}
	if !roamingGatewayPlan.IsCloudGateway.IsNull() {
		addRoamingGatewayBody.SetIsCloudGateway(roamingGatewayPlan.IsCloudGateway.ValueBool())
	}
	if !roamingGatewayPlan.CallbackUrl.IsNull() {
		addRoamingGatewayBody.SetCallbackUrl(roamingGatewayPlan.CallbackUrl.ValueString())
	}
	if !roamingGatewayPlan.SessionReliability.IsNull() {
		addRoamingGatewayBody.SetSessionReliability(roamingGatewayPlan.SessionReliability.ValueBool())
	}
	if !roamingGatewayPlan.RequestTicketTwoSTAs.IsNull() {
		addRoamingGatewayBody.SetRequestTicketTwoSTAs(roamingGatewayPlan.RequestTicketTwoSTAs.ValueBool())
	}
	if !roamingGatewayPlan.StasUseLoadBalancing.IsNull() {
		addRoamingGatewayBody.SetStasUseLoadBalancing(roamingGatewayPlan.StasUseLoadBalancing.ValueBool())
	}
	return addRoamingGatewayBody

}

func setRoamingGateway(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, existingGateways []citrixstorefront.STFRoamingGatewayResponseModel, plan STFDeploymentResourceModel) error {
	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddError(
			"Error getting StoreFront Roaming Gateway SiteId",
			"Error message: "+err.Error(),
		)
		return err
	}

	// Create a map of existing gateway names
	existingGatewayNames := map[string]bool{}

	for _, existingGateway := range existingGateways {
		if existingGateway.Name.Get() == nil || *existingGateway.Name.Get() == "" {
			continue
		}
		existingGatewayNames[*existingGateway.Name.Get()] = true
	}

	// Create a map of gateway names from the plan
	planGatewayNames := map[string]bool{}
	for _, gateway := range util.ObjectListToTypedArray[RoamingGateway](ctx, diagnostics, plan.RoamingGateway) {
		planGatewayNames[gateway.Name.ValueString()] = true
	}

	// Delete Roaming Gateways that are in not in the plan but are in the existing gateways

	for _, existingGateway := range existingGateways {
		if existingGateway.Name.Get() == nil || *existingGateway.Name.Get() == "" {
			continue
		}
		if _, ok := planGatewayNames[*existingGateway.Name.Get()]; !ok {
			var deleteRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel
			deleteRoamingGatewayBody.SetName(*existingGateway.Name.Get())
			var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
			getRoamingServiceBody.SetSiteId(siteIdInt)
			deleteRoamingGatewayRequest := client.StorefrontClient.RoamingSF.STFRoamingGatewayRemove(ctx, deleteRoamingGatewayBody, getRoamingServiceBody)
			err := deleteRoamingGatewayRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error deleting Roaming Gateway",
					"Error message: "+err.Error(),
				)
				return err
			}
		}
	}

	//set Gateways for Roaming Gateway Request
	getRoamingServiceBody := citrixstorefront.STFRoamingServiceRequestModel{}
	getRoamingServiceBody.SetSiteId(siteIdInt)
	gateways := util.ObjectListToTypedArray[RoamingGateway](ctx, diagnostics, plan.RoamingGateway)

	// update the roaming gateways
	for _, gateway := range gateways {
		// create a list of STFSTAUrlModel
		stfStaUrls := []citrixstorefront.STFSTAUrlModel{}
		plannedStaUrls := util.ObjectListToTypedArray[STFSecureTicketAuthority](ctx, diagnostics, gateway.SecureTicketAuthorityUrls)

		for _, staUrl := range plannedStaUrls {
			staUrlModel := citrixstorefront.STFSTAUrlModel{}
			staUrlModel.SetStaUrl(staUrl.StaUrl.ValueString())
			staUrlModel.SetStaValidationEnabled(staUrl.StaValidationEnabled.ValueBool())
			staUrlModel.SetStaValidationSecret(staUrl.StaValidationSecret.ValueString())
			stfStaUrls = append(stfStaUrls, staUrlModel)
		}

		// if the gateway name is not in the existing gateways, create a new gateway
		if _, ok := existingGatewayNames[gateway.Name.ValueString()]; !ok {
			var gatewaySetBody = buildRoamingGatewayBody(gateway)
			var gatewayGetBody citrixstorefront.GetSTFRoamingGatewayRequestModel
			gatewayGetBody.SetName(gateway.Name.ValueString())

			createRoamingGatewayRequest := client.StorefrontClient.RoamingSF.STFRoamingGatewayAdd(ctx, gatewaySetBody, getRoamingServiceBody, stfStaUrls)
			_, err := createRoamingGatewayRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error creating Roaming Gateway",
					"Error message: "+err.Error(),
				)
				return err
			}
		} else {
			// if the gateway name is in the existing gateways, update the gateway
			var setRoamingGatewayBody citrixstorefront.SetSTFRoamingGatewayRequestModel

			setRoamingGatewayBody.SetName(gateway.Name.ValueString())
			if !gateway.LogonType.IsNull() {
				includedLogonType, err := models.NewLogonTypeFromValue(gateway.LogonType.ValueString())
				if err != nil {
					diagnostics.AddError(
						"Error updating Logon Type",
						fmt.Sprintf("Unsupported criteria type %s.", gateway.LogonType.ValueString()),
					)
					return err
				}
				setRoamingGatewayBody.SetLogonType(*includedLogonType)
			}

			if !gateway.SmartCardFallbackLogonType.IsNull() {
				includedSmartCardFallbackLogonType, err := citrixstorefront.NewLogonTypeFromValue(gateway.SmartCardFallbackLogonType.ValueString())
				if err != nil {
					diagnostics.AddError(
						"Error updating Smartcard Fallback Logon Type",
						fmt.Sprintf("Unsupported criteria type %s.", gateway.SmartCardFallbackLogonType.ValueString()),
					)
					return err
				}
				setRoamingGatewayBody.SetSmartCardFallbackLogonType(*includedSmartCardFallbackLogonType)
			}

			if !gateway.Version.IsNull() {
				setRoamingGatewayBody.SetVersion(gateway.Version.ValueString())
			}
			if !gateway.GatewayUrl.IsNull() {
				setRoamingGatewayBody.SetGatewayUrl(gateway.GatewayUrl.ValueString())
			}
			if !gateway.CallbackUrl.IsNull() {
				setRoamingGatewayBody.SetCallbackUrl(gateway.CallbackUrl.ValueString())
			}
			if !gateway.SessionReliability.IsNull() {
				setRoamingGatewayBody.SetSessionReliability(gateway.SessionReliability.ValueBool())
			}
			if !gateway.RequestTicketTwoSTAs.IsNull() {
				setRoamingGatewayBody.SetRequestTicketTwoSTAs(gateway.RequestTicketTwoSTAs.ValueBool())
			}

			setRoamingGatewayBody.SetSubnetIPAddress(gateway.SubnetIPAddress.ValueString())
			if !gateway.GslbUrl.IsNull() {
				setRoamingGatewayBody.SetGslbUrl(gateway.GslbUrl.ValueString())
			}
			if !gateway.IsCloudGateway.IsNull() {
				setRoamingGatewayBody.SetIsCloudGateway(gateway.IsCloudGateway.ValueBool())
			}

			setRoamingGatewayRequest := client.StorefrontClient.RoamingSF.STFRoamingGatewaySet(ctx, setRoamingGatewayBody, getRoamingServiceBody, stfStaUrls)
			err = setRoamingGatewayRequest.Execute()
			if err != nil {
				diagnostics.AddError(
					"Error updating Roaming Gateway",
					"Error message: "+err.Error(),
				)
				return err
			}
		}
	}
	return nil
}
