// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"
	"fmt"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &stfRoamingBeaconResource{}
	_ resource.ResourceWithConfigure      = &stfRoamingBeaconResource{}
	_ resource.ResourceWithImportState    = &stfRoamingBeaconResource{}
	_ resource.ResourceWithValidateConfig = &stfRoamingBeaconResource{}
)

// NewSTFRoamingGatewayResource is a helper function to simplify the provider implementation.
func NewSTFRoamingBeaconResource() resource.Resource {
	return &stfRoamingBeaconResource{}
}

// stfRoamingBeaconResource is the resource implementation.
type stfRoamingBeaconResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ValidateConfig implements resource.ResourceWithValidateConfig.
func (*stfRoamingBeaconResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data STFRoamingBeaconResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

// Metadata returns the resource type name.
func (r *stfRoamingBeaconResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_roaming_beacon"
}

// Configure adds the provider configured client to the resource.
func (r *stfRoamingBeaconResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (*stfRoamingBeaconResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = STFRoamingBeaconResourceModel{}.GetSchema()
}

// Create is the implementation of the Create method in the ResourceWithImportState interface.
func (r *stfRoamingBeaconResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFRoamingBeaconResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error

	var roamingBeaconInternalBody citrixstorefront.SetSTFRoamingInternalBeaconRequestModel
	roamingBeaconInternalBody.SetInternal(plan.Internal.ValueString())

	var remoteRoamingInternalBeacon citrixstorefront.GetSTFRoamingInternalBeaconResponseModel
	var remoteRoamingExternalBeacon citrixstorefront.GetSTFRoamingExternalBeaconResponseModel

	// Set STF Roaming Gateway
	if !plan.External.IsNull() {
		var roamingBeaconExternalBody citrixstorefront.SetSTFRoamingExternalBeaconRequestModel
		roamingBeaconExternalBody.SetExternal(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.External))

		roamingBeaconRequest := r.client.StorefrontClient.RoamingSF.SetRoamingExternalBeacon(ctx, roamingBeaconExternalBody, roamingBeaconInternalBody)
		err = roamingBeaconRequest.Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for External IP addresses",
				"Error message: "+err.Error(),
			)
			return
		}

		// Get the External IPs
		getRoamingExtBeaconRequest := r.client.StorefrontClient.RoamingSF.GetRoamingExternalBeacon(ctx)
		remoteRoamingExternalBeacon, err = getRoamingExtBeaconRequest.Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting StoreFront Roaming Beacon for External IP Addresses",
				"Error message: "+err.Error(),
			)
			return
		}
	} else {
		// Set STF Roaming Gateway
		roamingBeaconRequest := r.client.StorefrontClient.RoamingSF.SetRoamingInternalBeacon(ctx, roamingBeaconInternalBody)
		err = roamingBeaconRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for Internal IP adrress",
				"Error message: "+err.Error(),
			)
			return
		}

	}

	// Get STF Roaming Internal Beacon details
	getRoamingBeaconRequest := r.client.StorefrontClient.RoamingSF.GetRoamingInternalBeacon(ctx)
	remoteRoamingInternalBeacon, err = getRoamingBeaconRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for Internal IP Address",
			"Error message: "+err.Error(),
		)
		return
	}

	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingInternalBeacon, &remoteRoamingExternalBeacon)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read implements resource.Resource.
func (r *stfRoamingBeaconResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	// Get current state
	var state STFRoamingBeaconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get STF Roaming Gateway details
	getRoamingBeaconInternalRequest := r.client.StorefrontClient.RoamingSF.GetRoamingInternalBeacon(ctx)
	remoteRoamingInternalBeacon, err := getRoamingBeaconInternalRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for Internal IP Address",
			"Error message: "+err.Error(),
		)
		return
	}

	// Get STF Roaming Gateway details
	getRoamingBeaconExternalRequest := r.client.StorefrontClient.RoamingSF.GetRoamingExternalBeacon(ctx)
	remoteRoamingExternalBeacon, err := getRoamingBeaconExternalRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for External IP Addresses",
			"Error message: "+err.Error(),
		)
		return
	}
	state.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingInternalBeacon, &remoteRoamingExternalBeacon)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *stfRoamingBeaconResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFRoamingBeaconResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error

	var roamingBeaconInternalBody citrixstorefront.SetSTFRoamingInternalBeaconRequestModel
	roamingBeaconInternalBody.SetInternal(plan.Internal.ValueString())

	var remoteRoamingInternalBeacon citrixstorefront.GetSTFRoamingInternalBeaconResponseModel
	var remoteRoamingExternalBeacon citrixstorefront.GetSTFRoamingExternalBeaconResponseModel

	// Set STF Roaming Gateway
	if !plan.External.IsNull() {
		var roamingBeaconExternalBody citrixstorefront.SetSTFRoamingExternalBeaconRequestModel
		roamingBeaconExternalBody.SetExternal(util.StringListToStringArray(ctx, &resp.Diagnostics, plan.External))

		roamingBeaconRequest := r.client.StorefrontClient.RoamingSF.SetRoamingExternalBeacon(ctx, roamingBeaconExternalBody, roamingBeaconInternalBody)
		err = roamingBeaconRequest.Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for External IP addresses",
				"Error message: "+err.Error(),
			)
			return
		}

		// Get the External IPs
		getRoamingExtBeaconRequest := r.client.StorefrontClient.RoamingSF.GetRoamingExternalBeacon(ctx)
		remoteRoamingExternalBeacon, err = getRoamingExtBeaconRequest.Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting StoreFront Roaming Beacon for External IP Addresses",
				"Error message: "+err.Error(),
			)
			return
		}
	} else {
		// Set STF Roaming Gateway
		roamingBeaconRequest := r.client.StorefrontClient.RoamingSF.SetRoamingInternalBeacon(ctx, roamingBeaconInternalBody)
		err = roamingBeaconRequest.Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting StoreFront Roaming Beacon for Internal IP adrress",
				"Error message: "+err.Error(),
			)
			return
		}

	}

	// Get STF Roaming Internal Beacon details
	getRoamingBeaconRequest := r.client.StorefrontClient.RoamingSF.GetRoamingInternalBeacon(ctx)
	remoteRoamingInternalBeacon, err = getRoamingBeaconRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting StoreFront Roaming Beacon for Internal IP Address",
			"Error message: "+err.Error(),
		)
		return
	}

	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &remoteRoamingInternalBeacon, &remoteRoamingExternalBeacon)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *stfRoamingBeaconResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFRoamingBeaconResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var getRoamingServiceBody citrixstorefront.STFRoamingServiceRequestModel
	getRoamingServiceBody.SetSiteId(state.SiteId.ValueInt64())
	// Delete existing STF Roaming Gateway
	deleteRoamingBeaconRequest := r.client.StorefrontClient.RoamingSF.STFRoamingBeaconInternalRemove(ctx, getRoamingServiceBody)
	err := deleteRoamingBeaconRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting StoreFront Roaming Beacon",
			"Error message: "+err.Error(),
		)
		return
	}

}

func (r *stfRoamingBeaconResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)
	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `site_id,internal_ip`, got: %q", req.ID),
		)
		return
	}
	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("internal_ip"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idSegments[0])...)
}
