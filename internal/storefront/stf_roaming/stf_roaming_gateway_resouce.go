// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_deployment"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &stfRoamingGatewayResource{}
	_ resource.ResourceWithConfigure      = &stfRoamingGatewayResource{}
	_ resource.ResourceWithImportState    = &stfRoamingGatewayResource{}
	_ resource.ResourceWithValidateConfig = &stfRoamingGatewayResource{}
)

// NewSTFRoamingGatewayResource is a helper function to simplify the provider implementation.
func NewSTFRoamingGatewayResource() resource.Resource {
	return &stfRoamingGatewayResource{}
}

// stfRoamingGatewayResource is the resource implementation.
type stfRoamingGatewayResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (*stfRoamingGatewayResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data STFRoamingGatewayResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.SecureTicketAuthorityUrls.IsNull() {
		staUrlList := util.ObjectListToTypedArray[STFSecureTicketAuthority](ctx, &resp.Diagnostics, data.SecureTicketAuthorityUrls)

		for _, staUrl := range staUrlList {
			if staUrl.StaValidationEnabled.ValueBool() {
				if staUrl.StaValidationSecret.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("sta_validation_secret"),
						"Incorrect Attribute Configuration",
						"STA Validation Secret is required when STA Validation is enabled",
					)
				}
			} else if !staUrl.StaValidationSecret.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("sta_validation_secret"),
					"Incorrect Attribute Configuration",
					"STA Validation Secret should be empty when STA Validation is disabled",
				)
			}
		}
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

}

// Metadata returns the resource type name.
func (r *stfRoamingGatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_roaming_gateway"
}

// Configure adds the provider configured client to the resource.
func (r *stfRoamingGatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (r *stfRoamingGatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = STFRoamingGatewayResourceModel{}.GetSchema()
}

// Create implements resource.Resource.
func (r *stfRoamingGatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFRoamingGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront WebReceiver ",
			"Error message: "+err.Error(),
		)
		return
	}

	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
	getRoamingServiceBody.SetSiteId(siteIdInt)

	var addRoamingGatewayBody citrixstorefront.AddSTFRoamingGatewayRequestModel
	addRoamingGatewayBody.SetName(plan.Name.ValueString())
	addRoamingGatewayBody.SetLogonType(plan.LogonType.ValueString())
	addRoamingGatewayBody.SetSmartCardFallbackLogonType(plan.SmartCardFallbackLogonType.ValueString())
	addRoamingGatewayBody.SetGatewayUrl(plan.GatewayUrl.ValueString())
	addRoamingGatewayBody.SetVersion(plan.Version.ValueString())
	addRoamingGatewayBody.SetSubnetIPAddress(plan.SubnetIPAddress.ValueString())
	addRoamingGatewayBody.SetStasBypassDuration(plan.StasBypassDuration.ValueString())
	addRoamingGatewayBody.SetGslbUrl(plan.GslbUrl.ValueString())
	addRoamingGatewayBody.SetIsCloudGateway(plan.IsCloudGateway.ValueBool())
	addRoamingGatewayBody.SetCallbackUrl(plan.CallbackUrl.ValueString())
	addRoamingGatewayBody.SetSessionReliability(plan.SessionReliability.ValueBool())
	addRoamingGatewayBody.SetRequestTicketTwoSTAs(plan.RequestTicketTwoSTAs.ValueBool())
	addRoamingGatewayBody.SetStasUseLoadBalancing(plan.StasUseLoadBalancing.ValueBool())

	stfStaUrls := []citrixstorefront.STFSTAUrlModel{}
	plannedStaUrls := util.ObjectListToTypedArray[STFSecureTicketAuthority](ctx, &resp.Diagnostics, plan.SecureTicketAuthorityUrls)
	for _, staUrl := range plannedStaUrls {
		staUrlModel := citrixstorefront.STFSTAUrlModel{}
		staUrlModel.SetStaUrl(staUrl.StaUrl.ValueString())
		staUrlModel.SetStaValidationEnabled(staUrl.StaValidationEnabled.ValueBool())
		staUrlModel.SetStaValidationSecret(staUrl.StaValidationSecret.ValueString())

		stfStaUrls = append(stfStaUrls, staUrlModel)
	}

	// Create new STF Roaming Gateway
	createRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayAdd(ctx, addRoamingGatewayBody, getRoamingServiceBody, stfStaUrls)
	_, err = createRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating StoreFront Roaming Gateway ",
			"Error message: "+err.Error(),
		)
		return
	}

	// Retrieve values from response
	var getRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel
	getRoamingGatewayBody.SetName(plan.Name.ValueString())

	// Get STF Roaming Gateway details
	getRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayGet(ctx, getRoamingGatewayBody, getRoamingServiceBody)
	remoteRoamingGateway, err := getRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing StoreFront Roaming Gateway details ",
			"Error message: "+err.Error(),
		)
		return
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingGateway)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *stfRoamingGatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFRoamingGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching StoreFront Roaming Gateway details",
			"Error message: "+err.Error(),
		)
		return
	}

	deployment, err := stf_deployment.GetSTFDeployment(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil || deployment == nil {
		return
	}

	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
	getRoamingServiceBody.SetSiteId(siteIdInt)

	// Retrieve values from response
	var getRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel
	getRoamingGatewayBody.SetName(state.Name.ValueString())

	// Get STF Roaming Gateway details
	getRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayGet(ctx, getRoamingGatewayBody, getRoamingServiceBody)
	remoteRoamingGateway, err := getRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching StoreFront Roaming Gateway details ",
			"Error message: "+err.Error(),
		)
		return
	}
	state.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingGateway)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *stfRoamingGatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFRoamingGatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Roaming Gateway configurations",
			"Error message: "+err.Error(),
		)
		return
	}

	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
	getRoamingServiceBody.SetSiteId(siteIdInt)

	// Retrieve values from response
	var getRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel
	getRoamingGatewayBody.SetName(plan.Name.ValueString())

	var setRoamingGatewayBody citrixstorefront.SetSTFRoamingGatewayRequestModel
	setRoamingGatewayBody.SetName(plan.Name.ValueString())
	setRoamingGatewayBody.SetLogonType(plan.LogonType.ValueString())
	setRoamingGatewayBody.SetSmartCardFallbackLogonType(plan.SmartCardFallbackLogonType.ValueString())
	setRoamingGatewayBody.SetVersion(plan.Version.ValueString())
	setRoamingGatewayBody.SetGatewayUrl(plan.GatewayUrl.ValueString())
	setRoamingGatewayBody.SetCallbackUrl(plan.CallbackUrl.ValueString())
	setRoamingGatewayBody.SetSessionReliability(plan.SessionReliability.ValueBool())
	setRoamingGatewayBody.SetRequestTicketTwoSTAs(plan.RequestTicketTwoSTAs.ValueBool())
	setRoamingGatewayBody.SetSubnetIPAddress(plan.SubnetIPAddress.ValueString())
	setRoamingGatewayBody.SetGslbUrl(plan.GslbUrl.ValueString())
	setRoamingGatewayBody.SetIsCloudGateway(plan.IsCloudGateway.ValueBool())

	stfStaUrls := []citrixstorefront.STFSTAUrlModel{}
	plannedStaUrls := util.ObjectListToTypedArray[STFSecureTicketAuthority](ctx, &resp.Diagnostics, plan.SecureTicketAuthorityUrls)
	for _, staUrl := range plannedStaUrls {
		staUrlModel := citrixstorefront.STFSTAUrlModel{}
		staUrlModel.SetStaUrl(staUrl.StaUrl.ValueString())
		staUrlModel.SetStaValidationEnabled(staUrl.StaValidationEnabled.ValueBool())
		staUrlModel.SetStaValidationSecret(staUrl.StaValidationSecret.ValueString())

		stfStaUrls = append(stfStaUrls, staUrlModel)
	}

	// Update STF Roaming Gateway
	updateRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewaySet(ctx, setRoamingGatewayBody, getRoamingServiceBody, stfStaUrls)
	err = updateRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Roaming Gateway configurations",
			"Error message: "+err.Error(),
		)
		return
	}

	// Get STF Roaming Gateway details
	getRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayGet(ctx, getRoamingGatewayBody, getRoamingServiceBody)
	remoteRoamingGateway, err := getRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating StoreFront Roaming Gateway details ",
			"Error message: "+err.Error(),
		)
		return
	}
	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingGateway)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *stfRoamingGatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFRoamingGatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var getRoamingGatewayBody citrixstorefront.GetSTFRoamingGatewayRequestModel
	getRoamingGatewayBody.SetName(state.Name.ValueString())

	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
	siteIdInt, err := strconv.ParseInt(state.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Roaming Gateway ",
			"Error message: "+err.Error(),
		)
		return
	}
	getRoamingServiceBody.SetSiteId(siteIdInt)

	// Delete existing STF Roaming Gateway
	deleteRoamingGatewayRequest := r.client.StorefrontClient.RoamingSF.STFRoamingGatewayRemove(ctx, getRoamingGatewayBody, getRoamingServiceBody)
	err = deleteRoamingGatewayRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Roaming Gateway ",
			"Error message: "+err.Error(),
		)
		return
	}
}

func (r *stfRoamingGatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `site_id,name`, got: %q", req.ID),
		)
		return
	}

	_, err := strconv.Atoi(idSegments[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Site ID in Import Identifier",
			fmt.Sprintf("Site ID should be an integer, got: %q", idSegments[0]),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idSegments[1])...)
}
