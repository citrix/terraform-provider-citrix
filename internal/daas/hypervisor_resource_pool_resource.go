package daas

import (
	"context"
	"fmt"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &hypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure   = &hypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState = &hypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewHypervisorResourcePoolResource() resource.Resource {
	return &hypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type hypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *hypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_hypervisor_resource_pool"
}

func (r *hypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a hypervisor resource pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the resource pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource pool.",
				Required:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor for which the resource pool needs to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hypervisor_connection_type": schema.StringAttribute{
				Description: "Connection Type of the hypervisor (AzureRM, AWS, GCP).",
				Computed:    true,
			},
			"virtual_network_resource_group": schema.StringAttribute{
				Description: "The name of the resource group where the vnet resides. Required when connection type is Azure.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_network": schema.StringAttribute{
				Description: "Name of the cloud virtual network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of subets to allocate VDAs within the virtual network. Required when connection type is Azure or GCP.",
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Cloud Region where the virtual network sits in. Required when connection type is Azure or GCP.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Description: "The name of the availability zone resource to use for provisioning operations in this resource pool. Required when connection type is AWS.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"project_name": schema.StringAttribute{
				Description: "GCP Project name. Required when connection type is GCP.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *hypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *hypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.HypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hypervisorId := plan.Hypervisor.ValueString()
	hypervisor, err := GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)

	if err != nil {
		return
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	hypervisorConnectionType := hypervisor.GetConnectionType()

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)

	switch hypervisorConnectionType {
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AWS:
		if plan.AvailabilityZone.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for AWS",
				"Availability Zone is missing.",
			)
			return
		}
		virtualNetworkPath := fmt.Sprintf("%s.virtualprivatecloud", plan.VirtualNetwork.ValueString())
		resourcePoolDetails.SetVirtualPrivateCloud(virtualNetworkPath)
		availabilityZonePath := fmt.Sprintf("%s/%s.availabilityzone", virtualNetworkPath, plan.AvailabilityZone.ValueString())
		resourcePoolDetails.SetAvailabilityZone(availabilityZonePath)
		planSubnet := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Subnets)
		subnets, err := util.GetFilteredResourcePathList(ctx, r.client, hypervisorId, availabilityZonePath, "Network", planSubnet, hypervisorConnectionType)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for AWS",
				"Error message: "+util.ReadClientError(err),
			)
			return
		}

		if len(plan.Subnets) != len(subnets) {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for AWS",
				"Subnet contains invalid value.",
			)
			return
		}

		resourcePoolDetails.SetNetworks(subnets)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_AZURE_RM:
		if plan.Region.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Azure",
				"Cloud Region is missing.",
			)
			return
		}
		regionPath := util.GetSingleResourcePath(ctx, r.client, hypervisorId, "", plan.Region.ValueString(), "", "")
		resourcePoolDetails.SetRegion(regionPath)
		vnetPath := util.GetSingleResourcePath(ctx, r.client, hypervisorId, fmt.Sprintf("%s/virtualprivatecloud.folder", regionPath), plan.VirtualNetwork.ValueString(), "VirtualPrivateCloud", plan.VirtualNetworkResourceGroup.ValueString())
		resourcePoolDetails.SetVirtualNetwork(vnetPath)
		//Checking the subnet
		if len(plan.Subnets) == 0 {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Azure",
				"Subnet is missing.",
			)
			return
		}
		planSubnet := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Subnets)
		subnets, err := util.GetFilteredResourcePathList(ctx, r.client, hypervisorId, fmt.Sprintf("%s/virtualprivatecloud.folder/%s", regionPath, vnetPath), "Network", planSubnet, hypervisorConnectionType)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Azure",
				"Error message: "+util.ReadClientError(err),
			)
			return
		}

		if len(plan.Subnets) != len(subnets) {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Azure",
				"Subnet contains invalid value.",
			)
			return
		}
		resourcePoolDetails.SetSubnets(subnets)
	case citrixorchestration.HYPERVISORCONNECTIONTYPE_GOOGLE_CLOUD_PLATFORM:
		if plan.Region.IsNull() {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for GCP",
				"Cloud Region is missing.",
			)
			return
		}
		regionPath := fmt.Sprintf("%s.project/%s.region", plan.ProjectName.ValueString(), plan.Region.ValueString())
		resourcePoolDetails.SetRegion(regionPath)
		vnetPath := fmt.Sprintf("%s/%s.virtualprivatecloud", regionPath, plan.VirtualNetwork.ValueString())
		resourcePoolDetails.SetVirtualPrivateCloud(vnetPath)
		//Checking the subnet
		if len(plan.Subnets) == 0 {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for GCP",
				"Subnet is missing.",
			)
			return
		}
		planSubnet := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Subnets)
		subnets, err := util.GetFilteredResourcePathList(ctx, r.client, hypervisorId, vnetPath, "Network", planSubnet, hypervisorConnectionType)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for GCP",
				"Error message: "+util.ReadClientError(err),
			)
			return
		}

		if len(plan.Subnets) != len(subnets) {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for GCP",
				"Subnet contains invalid value.",
			)
			return
		}
		resourcePoolDetails.SetNetworks(subnets)
	default:
		resp.Diagnostics.AddError(
			"Error creating Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	createResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsCreateResourcePool(ctx, hypervisorId)
	createResourcePoolRequest = createResourcePoolRequest.CreateHypervisorResourcePoolRequestModel(resourcePoolDetails)
	resourcePool, httpResp, err := citrixdaasclient.AddRequestData(createResourcePoolRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource Pool for Hypervisor "+hypervisorId,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	plan = plan.RefreshPropertyValues(*resourcePool, hypervisorConnectionType)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *hypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.HypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get hypervisor properties from Orchestration
	hypervisorId := state.Hypervisor.ValueString()

	// Get the resource pool
	resourcePool, err := GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, hypervisorId, state.Id.ValueString())
	if err != nil {
		return
	}

	// Override with refreshed state
	state = state.RefreshPropertyValues(*resourcePool, resourcePool.GetConnectionType())

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *hypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.HypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.HypervisorResourcePoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	connectionType, err := citrixorchestration.NewHypervisorConnectionTypeFromValue(state.HypervisorConnectionType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor "+plan.Hypervisor.ValueString(),
			"Unsupported hypervisor connection type.",
		)
		return
	}
	editHypervisorResourcePool.SetConnectionType(*connectionType)
	patchResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsPatchHypervisorResourcePool(ctx, plan.Hypervisor.ValueString(), plan.Id.ValueString())
	patchResourcePoolRequest = patchResourcePoolRequest.EditHypervisorResourcePoolRequestModel(editHypervisorResourcePool).Async(true)
	httpResp, err := citrixdaasclient.AddRequestData(patchResourcePoolRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor "+plan.Hypervisor.ValueString(),
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	jobId := util.GetJobIdFromHttpResponse(*httpResp)
	jobResponseModel, err := r.client.WaitForJob(ctx, jobId, 5)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor "+plan.Hypervisor.ValueString(),
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorDetail := "TransactionId: " + util.GetTransactionIdFromHttpResponse(httpResp) +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		resp.Diagnostics.AddError(
			"Error updating Resource Pool for Hypervisor "+plan.Hypervisor.ValueString(),
			errorDetail,
		)
	}

	// get updated resource pool
	hypervisorId := plan.Hypervisor.ValueString()
	resourcePool, err := GetHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, hypervisorId, plan.Id.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(*resourcePool, resourcePool.GetConnectionType())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *hypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *hypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.HypervisorResourcePoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete resource pool
	hypervisorId := state.Hypervisor.ValueString()
	deleteHypervisorResourcePoolRequest := r.client.ApiClient.HypervisorsAPIsDAAS.HypervisorsDeleteHypervisorResourcePool(ctx, hypervisorId, state.Id.ValueString())
	httpResp, err := citrixdaasclient.AddRequestData(deleteHypervisorResourcePoolRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Resource Pool for Hypervisor "+hypervisorId,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

func GetHypervisorResourcePool(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, hypervisorResourcePoolId string) (*citrixorchestration.HypervisorResourcePoolDetailResponseModel, error) {
	getResourcePoolsRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePool(ctx, hypervisorId, hypervisorResourcePoolId)
	resourcePool, httpResp, err := citrixdaasclient.AddRequestData(getResourcePoolsRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error reading ResourcePool for Hypervisor "+hypervisorId,
			"TransactionId: "+util.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return resourcePool, err
}
