// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &gcpHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure   = &gcpHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState = &gcpHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewGcpHypervisorResourcePoolResource() resource.Resource {
	return &gcpHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type gcpHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *gcpHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_hypervisor_resource_pool"
}

func (r *gcpHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetGcpHypervisorResourcePoolSchema()
}

// Configure adds the provider configured client to the resource.
func (r *gcpHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *gcpHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GcpHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)

	if err != nil {
		return
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM {
		resp.Diagnostics.AddError(
			"Error creating GCP Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)

	if plan.Region.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for GCP",
			"Cloud Region is missing.",
		)
		return
	}
	regionPath := fmt.Sprintf("%s.project/%s.region", plan.ProjectName.ValueString(), plan.Region.ValueString())
	resourcePoolDetails.SetRegion(regionPath)
	vnetPath := fmt.Sprintf("%s/%s.virtualprivatecloud", regionPath, plan.Vpc.ValueString())
	if plan.SharedVpc.ValueBool() {
		// Support shared VPC if specified as true
		vnetPath = fmt.Sprintf("%s/%s.sharedvirtualprivatecloud", regionPath, plan.Vpc.ValueString())
	}
	resourcePoolDetails.SetVirtualPrivateCloud(vnetPath)
	//Checking the subnet
	if plan.Subnets.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for GCP",
			"Subnet is missing.",
		)
		return
	}
	planSubnet := util.StringListToStringArray(ctx, &diags, plan.Subnets)
	subnets, err := util.GetFilteredResourcePathList(ctx, r.client, hypervisorId, vnetPath, util.NetworkResourceType, planSubnet, hypervisorConnectionType, hypervisor.GetPluginId())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for GCP",
			"Error message: "+util.ReadClientError(err),
		)
		return
	}

	if len(plan.Subnets.Elements()) != len(subnets) {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for GCP",
			"Subnet contains invalid value.",
		)
		return
	}
	resourcePoolDetails.SetNetworks(subnets)

	resourcePool, err := CreateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, *hypervisor, resourcePoolDetails)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &diags, resourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *gcpHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state GcpHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get hypervisor properties from Orchestration
	hypervisorId := state.Hypervisor.ValueString()

	// Get the resource pool
	resourcePool, err := ReadHypervisorResourcePool(ctx, r.client, resp, hypervisorId, state.Id.ValueString())
	if err != nil {
		return
	}

	// Override with refreshed state
	state = state.RefreshPropertyValues(ctx, &diags, resourcePool)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *gcpHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GcpHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state GcpHypervisorResourcePoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM)

	planSubnet := util.StringListToStringArray(ctx, &diags, plan.Subnets)
	regionPath := fmt.Sprintf("%s.project/%s.region", plan.ProjectName.ValueString(), plan.Region.ValueString())
	vnetPath := fmt.Sprintf("%s/%s.virtualprivatecloud", regionPath, plan.Vpc.ValueString())
	if plan.SharedVpc.ValueBool() {
		// Support shared VPC if specified as true
		vnetPath = fmt.Sprintf("%s/%s.sharedvirtualprivatecloud", regionPath, plan.Vpc.ValueString())
	}
	subnets, err := util.GetFilteredResourcePathList(ctx, r.client, plan.Hypervisor.ValueString(), vnetPath, util.NetworkResourceType, planSubnet, citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM, "")
	if err != nil {
		return
	}
	editHypervisorResourcePool.SetNetworks(subnets)

	updatedResourcePool, err := UpdateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.Hypervisor.ValueString(), plan.Id.ValueString(), editHypervisorResourcePool)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &diags, updatedResourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *gcpHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: hypervisorId,hypervisorResourcePoolId. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hypervisor"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)

}

func (r *gcpHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state GcpHypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource pool
	hypervisorId := state.Hypervisor.ValueString()
	deleteHypervisorResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisorResourcePool(ctx, hypervisorId, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorResourcePoolRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Resource Pool for Hypervisor "+hypervisorId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}
