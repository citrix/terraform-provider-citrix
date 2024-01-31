// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	resp.TypeName = req.ProviderTypeName + "_daas_gcp_hypervisor_resource_pool"
}

func (r *gcpHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description: "Name of the resource pool. Name should be unique across all hypervisors.",
				Required:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor for which the resource pool needs to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"vpc": schema.StringAttribute{
				Description: "Name of the cloud virtual network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of subnets to allocate VDAs within the virtual network.",
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Description: "Cloud Region where the virtual network sits in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !req.ConfigValue.IsNull() && !req.StateValue.IsNull() &&
								(location.Normalize(req.ConfigValue.ValueString()) != location.Normalize(req.StateValue.ValueString()))
						},
						"Force replacement when region changes, unless changing between GCP region name (East US) and Id (eastus)",
						"Force replacement when region changes, unless changing between GCP region name (East US) and Id (eastus)",
					),
				},
			},
			"project_name": schema.StringAttribute{
				Description: "GCP Project name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"shared_vpc": schema.BoolAttribute{
				Description: "Indicate whether the GCP Virtual Private Cloud is a shared VPC.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
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

	resourcePool, err := CreateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, *hypervisor, resourcePoolDetails)
	if err != nil {
		// Directly return. Error logs have been populated in common function
		return
	}

	plan = plan.RefreshPropertyValues(resourcePool)

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
	state = state.RefreshPropertyValues(resourcePool)

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

	updatedResourcePool, err := UpdateHypervisorResourcePool(ctx, r.client, &resp.Diagnostics, plan.Hypervisor.ValueString(), plan.Id.ValueString(), editHypervisorResourcePool)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(updatedResourcePool)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *gcpHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
