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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &azureHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure      = &azureHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState    = &azureHypervisorResourcePoolResource{}
	_ resource.ResourceWithValidateConfig = &azureHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan     = &azureHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewAzureHypervisorResourcePoolResource() resource.Resource {
	return &azureHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type azureHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *azureHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_hypervisor_resource_pool"
}

func (r *azureHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AzureHypervisorResourcePoolResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *azureHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *azureHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan AzureHypervisorResourcePoolResourceModel
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
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM {
		resp.Diagnostics.AddError(
			"Error creating Azure Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)

	if plan.Region.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Azure",
			"Cloud Region is missing.",
		)
		return
	}
	region, httpResp, err := util.GetSingleHypervisorResource(ctx, r.client, &resp.Diagnostics, hypervisorId, "", plan.Region.ValueString(), "Region", "", hypervisor)
	regionPath := region.GetRelativePath()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Azure",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\nFailed to resolve region %s, error: %s", plan.Region.ValueString(), err.Error()),
		)
		return
	}
	resourcePoolDetails.SetRegion(regionPath)
	vnet, httpResp, err := util.GetSingleHypervisorResource(ctx, r.client, &resp.Diagnostics, hypervisorId, fmt.Sprintf("%s/virtualprivatecloud.folder", regionPath), plan.VirtualNetwork.ValueString(), util.VirtualPrivateCloudResourceType, plan.VirtualNetworkResourceGroup.ValueString(), hypervisor)
	vnetPath := vnet.GetRelativePath()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Azure",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\nFailed to resolve virtual network %s in region %s, error: %s", plan.VirtualNetwork.ValueString(), plan.Region.ValueString(), err.Error()),
		)
		return
	}
	resourcePoolDetails.SetVirtualNetwork(vnetPath)
	//Checking the subnet
	if plan.Subnets.IsNull() {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Azure",
			"Subnet is missing.",
		)
		return
	}
	planSubnet := util.StringListToStringArray(ctx, &diags, plan.Subnets)
	availabilityZonePath := fmt.Sprintf("%s/virtualprivatecloud.folder/%s", regionPath, vnetPath)
	subnets, err := getHypervisorResourcePoolSubnets(ctx, r.client, &resp.Diagnostics, hypervisorId, availabilityZonePath, planSubnet, hypervisorConnectionType)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}
	resourcePoolDetails.SetSubnets(subnets)

	metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	resourcePoolDetails.SetMetadata(metadata)

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

func (r *azureHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state AzureHypervisorResourcePoolResourceModel
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

func (r *azureHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan AzureHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AzureHypervisorResourcePoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM)

	planSubnet := util.StringListToStringArray(ctx, &diags, plan.Subnets)
	regionPath := fmt.Sprintf("%s.region", plan.Region.ValueString())
	resourceGroupPath := fmt.Sprintf("%s.resourcegroup", plan.VirtualNetworkResourceGroup.ValueString())
	vnetPath := fmt.Sprintf("%s.virtualprivatecloud", plan.VirtualNetwork.ValueString())
	availabilityZonePath := fmt.Sprintf("%s/virtualprivatecloud.folder/%s/%s", regionPath, resourceGroupPath, vnetPath)
	subnets, err := getHypervisorResourcePoolSubnets(ctx, r.client, &resp.Diagnostics, plan.Hypervisor.ValueString(), availabilityZonePath, planSubnet, citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}
	editHypervisorResourcePool.SetSubnets(subnets)

	metadata := util.GetUpdatedMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, state.Metadata), util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, plan.Metadata))
	editHypervisorResourcePool.SetMetadata(metadata)

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

func (r *azureHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *azureHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state AzureHypervisorResourcePoolResourceModel
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

func (r *azureHypervisorResourcePoolResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data AzureHypervisorResourcePoolResourceModel
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

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}

func (r *azureHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
