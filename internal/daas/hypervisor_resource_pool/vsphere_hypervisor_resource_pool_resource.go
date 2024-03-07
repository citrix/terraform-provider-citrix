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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithConfigure   = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithImportState = &vsphereHypervisorResourcePoolResource{}
	_ resource.ResourceWithModifyPlan  = &vsphereHypervisorResourcePoolResource{}
)

// NewHypervisorResourcePoolResource is a helper function to simplify the provider implementation.
func NewVsphereHypervisorResourcePoolResource() resource.Resource {
	return &vsphereHypervisorResourcePoolResource{}
}

// hypervisorResource is the resource implementation.
type vsphereHypervisorResourcePoolResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (r *vsphereHypervisorResourcePoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if req.Plan.Raw.IsNull() {
		return
	}

	create := req.State.Raw.IsNull()

	var plan VsphereHypervisorResourcePoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !create {
		return
	}

	for _, storageModel := range plan.Storage {
		if storageModel.Superseded.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("storage"),
				"Incorrect attribute value",
				"Storage cannot be superseded when creating a new resource pool. Use only when updating the resource pool.",
			)
		}
	}

	for _, storageModel := range plan.TemporaryStorage {
		if storageModel.Superseded.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("temporary_storage"),
				"Incorrect attribute value",
				"Storage cannot be superseded when creating a new resource pool. Use only when updating the resource pool.",
			)
		}
	}
}

// Metadata returns the resource type name.
func (r *vsphereHypervisorResourcePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vsphere_hypervisor_resource_pool"
}

func (r *vsphereHypervisorResourcePoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vsphere hypervisor resource pool.",
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
			"cluster": schema.SingleNestedAttribute{
				Description: "Details of the cluster where resources reside and new resources will be created.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"datacenter": schema.StringAttribute{
						Description: "The name of the datacenter.",
						Required:    true,
					},
					"cluster_name": schema.StringAttribute{
						Description: "The name of the cluster.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.AtLeastOneOf(path.Expressions{
								path.MatchRelative().AtParent().AtName("host"),
							}...),
						},
					},
					"host": schema.StringAttribute{
						Description: "The IP address or FQDN of the host.",
						Optional:    true,
					},
				},
			},
			"networks": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of networks for allocating resources.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"storage": schema.ListNestedAttribute{
				Description:  "List of hypervisor storage to use for OS data.",
				Required:     true,
				NestedObject: getNestedAttributeObjectSchmeaForStorege(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"temporary_storage": schema.ListNestedAttribute{
				Description:  "List of hypervisor storage to use for temporary data.",
				Required:     true,
				NestedObject: getNestedAttributeObjectSchmeaForStorege(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"use_local_storage_caching": schema.BoolAttribute{
				Description: "Indicates whether intellicache is enabled to reduce load on the shared storage device. Will only be effective when shared storage is used. Default value is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *vsphereHypervisorResourcePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (r *vsphereHypervisorResourcePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan VsphereHypervisorResourcePoolResourceModel
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

	folderPath := hypervisor.GetXDPath()
	var relativePath string

	resource, err := util.GetSingleHypervisorResource(ctx, r.client, hypervisorId, folderPath, plan.Cluster.Datacenter.ValueString(), "datacenter", "", hypervisor)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Vsphere",
			fmt.Sprintf("Failed to resolve resource %s, error: %s", plan.Cluster.Datacenter.ValueString(), err.Error()),
		)
		return
	}
	folderPath = resource.GetXDPath()
	relativePath = resource.GetRelativePath()

	if !plan.Cluster.ClusterName.IsNull() {
		resource, err = util.GetSingleHypervisorResource(ctx, r.client, hypervisorId, folderPath, plan.Cluster.ClusterName.ValueString(), "cluster", "", hypervisor)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Vsphere",
				fmt.Sprintf("Failed to resolve resource %s, error: %s", plan.Cluster.ClusterName.ValueString(), err.Error()),
			)
			return
		}
		folderPath = resource.GetXDPath()
		relativePath = resource.GetRelativePath()
	}

	if !plan.Cluster.Host.IsNull() {
		resource, err = util.GetSingleHypervisorResource(ctx, r.client, hypervisorId, folderPath, plan.Cluster.Host.ValueString(), "computeresource", "", hypervisor)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating Hypervisor Resource Pool for Vsphere",
				fmt.Sprintf("Failed to resolve resource %s, error: %s", plan.Cluster.Host.ValueString(), err.Error()),
			)
			return
		}

		relativePath = resource.GetRelativePath()
	}

	var resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel
	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER {
		resp.Diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for Vsphere",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	resourcePoolDetails.SetName(plan.Name.ValueString())
	resourcePoolDetails.SetConnectionType(hypervisorConnectionType)
	resourcePoolDetails.SetRootPath(relativePath)
	storages, tempStorages := getStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, plan, true, false)
	networks := getNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, plan, true)
	if len(storages) == 0 || len(tempStorages) == 0 || len(networks) == 0 {
		// Error handled in helper function.
		return
	}
	resourcePoolDetails.SetStorage(storages)
	resourcePoolDetails.SetTemporaryStorage(tempStorages)
	resourcePoolDetails.SetNetworks(networks)

	resourcePoolDetails.SetUseLocalStorageCaching(plan.UseLocalStorageCaching.ValueBool())

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

func (r *vsphereHypervisorResourcePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state VsphereHypervisorResourcePoolResourceModel
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

func (r *vsphereHypervisorResourcePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan VsphereHypervisorResourcePoolResourceModel
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

	hypervisorConnectionType := hypervisor.GetConnectionType()
	if hypervisorConnectionType != citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER {
		resp.Diagnostics.AddError(
			"Error creating Resource Pool for Hypervisor",
			"Unsupported hypervisor connection type.",
		)
		return
	}

	var state VsphereHypervisorResourcePoolResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel
	editHypervisorResourcePool.SetName(plan.Name.ValueString())
	editHypervisorResourcePool.SetConnectionType(citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER)

	storagesToBeIncluded, tempStoragesToBeIncluded := getStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, plan, false, false)
	storagesToBeSuperseded, tempStoragesToBeSuperseded := getStorageList(ctx, r.client, &resp.Diagnostics, hypervisor, plan, false, true)
	networks := getNetworksList(ctx, r.client, &resp.Diagnostics, hypervisor, plan, false)
	if (storagesToBeIncluded == nil && storagesToBeSuperseded == nil) || (tempStoragesToBeIncluded == nil && tempStoragesToBeSuperseded == nil) || networks == nil {
		// Error handled in helper function.
		return
	}

	storageRequests := []citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
	for _, storage := range storagesToBeIncluded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		storageRequests = append(storageRequests, *request)
	}
	for _, storage := range storagesToBeSuperseded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		request.SetSuperseded(true)
		storageRequests = append(storageRequests, *request)
	}
	editHypervisorResourcePool.SetStorage(storageRequests)

	tempStorageRequests := []citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
	for _, storage := range tempStoragesToBeIncluded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		tempStorageRequests = append(tempStorageRequests, *request)
	}
	for _, storage := range tempStoragesToBeSuperseded {
		request := &citrixorchestration.HypervisorResourcePoolStorageRequestModel{}
		request.SetStoragePath(storage)
		request.SetSuperseded(true)
		tempStorageRequests = append(tempStorageRequests, *request)
	}
	editHypervisorResourcePool.SetTemporaryStorage(tempStorageRequests)

	editHypervisorResourcePool.SetNetworks(networks)

	editHypervisorResourcePool.SetUseLocalStorageCaching(plan.UseLocalStorageCaching.ValueBool())

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

func (r *vsphereHypervisorResourcePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *vsphereHypervisorResourcePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state VsphereHypervisorResourcePoolResourceModel
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

func getStorageList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, plan VsphereHypervisorResourcePoolResourceModel, isCreate bool, forSuperseded bool) ([]string, []string) {
	action := "updating"
	if isCreate {
		action = "creating"
	}

	storage := []basetypes.StringValue{}
	for _, storageModel := range plan.Storage {
		if storageModel.Superseded.ValueBool() == forSuperseded {
			storage = append(storage, storageModel.StorageName)
		}
	}
	storageNames := util.ConvertBaseStringArrayToPrimitiveStringArray(storage)
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	folderPath := fmt.Sprintf("%s\\%s.datacenter", hypervisor.GetXDPath(), plan.Cluster.Datacenter.ValueString())
	if !plan.Cluster.ClusterName.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.cluster", folderPath, plan.Cluster.ClusterName.ValueString())
	}

	if !plan.Cluster.Host.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.computeresource", folderPath, plan.Cluster.Host.ValueString())
	}
	storages, err := util.GetFilteredResourcePathList(ctx, client, hypervisorId, folderPath, "storage", storageNames, hypervisorConnectionType)

	if len(storage) > 0 && len(storages) == 0 {
		errDetail := "No storage found for the given storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Vsphere", action),
			errDetail,
		)
		return nil, nil
	}

	tempStorage := []basetypes.StringValue{}
	for _, storageModel := range plan.TemporaryStorage {
		if storageModel.Superseded.ValueBool() == forSuperseded {
			tempStorage = append(tempStorage, storageModel.StorageName)
		}
	}
	tempStorageNames := util.ConvertBaseStringArrayToPrimitiveStringArray(tempStorage)
	tempStorages, err := util.GetFilteredResourcePathList(ctx, client, hypervisorId, folderPath, "storage", tempStorageNames, hypervisorConnectionType)
	if len(tempStorage) > 0 && len(tempStorages) == 0 {
		errDetail := "No storage found for the given temporary storage names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Vsphere", action),
			errDetail,
		)
		return nil, nil
	}

	return storages, tempStorages
}

func getNestedAttributeObjectSchmeaForStorege() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"storage_name": schema.StringAttribute{
				Description: "The name of the storage.",
				Required:    true,
			},
			"superseded": schema.BoolAttribute{
				Description: "Indicates whether the storage has been superseded. Superseded storage may be used for existing virtual machines, but is not used when provisioning new virtual machines. Use only when updating the resource pool.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func getNetworksList(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diags *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, plan VsphereHypervisorResourcePoolResourceModel, isCreate bool) []string {
	hypervisorId := hypervisor.GetId()
	hypervisorConnectionType := hypervisor.GetConnectionType()
	action := "updating"
	if isCreate {
		action = "creating"
	}

	folderPath := fmt.Sprintf("%s\\%s.datacenter", hypervisor.GetXDPath(), plan.Cluster.Datacenter.ValueString())
	if !plan.Cluster.ClusterName.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.cluster", folderPath, plan.Cluster.ClusterName.ValueString())
	}

	if !plan.Cluster.Host.IsNull() {
		folderPath = fmt.Sprintf("%s\\%s.computeresource", folderPath, plan.Cluster.Host.ValueString())
	}

	networkNames := util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Networks)
	networks, err := util.GetFilteredResourcePathList(ctx, client, hypervisorId, folderPath, "network", networkNames, hypervisorConnectionType)
	if len(networks) == 0 {
		errDetail := "No network found for the given network names"
		if err != nil {
			errDetail = util.ReadClientError(err)
		}
		diags.AddError(
			fmt.Sprintf("Error %s Hypervisor Resource Pool for Vsphere", action),
			errDetail,
		)
		return nil
	}

	return networks
}
