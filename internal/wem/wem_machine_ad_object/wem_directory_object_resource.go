package wem_machine_ad_object

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var mutex = &sync.Mutex{}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &wemDirectoryResource{}
	_ resource.ResourceWithConfigure   = &wemDirectoryResource{}
	_ resource.ResourceWithImportState = &wemDirectoryResource{}
	_ resource.ResourceWithModifyPlan  = &wemDirectoryResource{}
)

// Define the wemMachineResource struct
type wemDirectoryResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// ModifyPlan implements resource.ResourceWithModifyPlan.
func (w *wemDirectoryResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if w.client != nil && (w.client.ApiClient == nil || w.client.WemClient == nil) {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if w.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Error managing WEM Directory Object", "Directory Objects are only supported for Cloud customers.")
	}
}

// ImportState implements resource.ResourceWithImportState.
func (w *wemDirectoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure implements resource.ResourceWithConfigure.
func (w *wemDirectoryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	w.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (w *wemDirectoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WemDirectoryResourceModel{}.GetSchema()
}

// Metadata implements resource.Resource.
func (w *wemDirectoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wem_directory_object"
}

// NewWemDirectoryResource creates a new instance of the WemDirectoryResource.
func NewWemDirectoryResource() resource.Resource {
	return &wemDirectoryResource{}
}

// Create implements resource.Resource.
func (w *wemDirectoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	mutex.Lock()
	defer mutex.Unlock()

	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan WemDirectoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Supporting only catalog as machine-level AD objects in WEM
	machineCatalogId := plan.CatalogId.ValueString()
	catalog, err := util.GetMachineCatalog(ctx, w.client, &diags, machineCatalogId, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading machine catalog",
			"Could not read machine catalog with ID "+machineCatalogId+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	machineCatalogName := catalog.GetName()

	// Generate API request body from plan
	var body citrixwemservice.MachineModel
	body.SetSiteId(plan.SiteId.ValueInt64())
	body.SetSid(machineCatalogId)
	body.SetName(machineCatalogName)
	body.SetType("Catalog")
	body.SetEnabled(plan.Enabled.ValueBool())
	body.SetPriority(1000) // reserved priority, currently not used in WEM API

	// Generate Create MAchine AD Object API request
	machineADObjectCreateRequest := w.client.WemClient.MachineADObjectDAAS.AdObjectCreate(ctx)
	machineADObjectCreateRequest = machineADObjectCreateRequest.Body(body)
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[any](machineADObjectCreateRequest, w.client)

	// In case of 400 Bad Request, add it to diagnostics and return
	if httpResp.StatusCode == http.StatusBadRequest && err.Error() == "400 Bad Request  (Duplicate property)" {
		resp.Diagnostics.AddError(
			"Failed to create directory object for catalog '"+machineCatalogName+"'",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: A Directory Object with the same Machine Catalog ID already exists.",
		)
		return
	}

	// In case of any other error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error binding "+machineCatalogName+" to WEM configuration set ID "+strconv.FormatInt(plan.SiteId.ValueInt64(), 10),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get newly created machine AD object from remote
	machineADObject, err := getMachineADObjectBySid(ctx, w.client, machineCatalogId)
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, &machineADObject)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (w *wemDirectoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	mutex.Lock()
	defer mutex.Unlock()

	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state WemDirectoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate Delete API request
	machineADObjectDeleteRequest := w.client.WemClient.MachineADObjectDAAS.AdObjectDelete(ctx, state.Id.ValueString())
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[any](machineADObjectDeleteRequest, w.client)

	// In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting WEM Directory Object "+state.Id.String(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// Read implements resource.Resource.
func (w *wemDirectoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	mutex.Lock()
	defer mutex.Unlock()

	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state WemDirectoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get WEM Directory object by ID
	machineADObject, err := readDirectoryObject(ctx, w.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, machineADObject)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (w *wemDirectoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	mutex.Lock()
	defer mutex.Unlock()

	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan WemDirectoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Supporting only catalog as machine-level AD objects in WEM
	machineCatalogId := plan.CatalogId.ValueString()
	catalog, err := util.GetMachineCatalog(ctx, w.client, &diags, machineCatalogId, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading machine catalog",
			"Could not read machine catalog with ID "+machineCatalogId+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	machineCatalogName := catalog.GetName()

	// Generate API request body from plan
	var body citrixwemservice.MachineModel
	idInt64, err := strconv.ParseInt(plan.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error converting ID to int64",
			"Could not convert ID "+plan.Id.ValueString()+" to int64: "+err.Error(),
		)
		return
	}
	body.SetId(idInt64)
	body.SetSiteId(plan.SiteId.ValueInt64())
	body.SetSid(machineCatalogId)
	body.SetName(machineCatalogName)
	body.SetType("Catalog")
	body.SetEnabled(plan.Enabled.ValueBool())
	body.SetPriority(1000) // reserved priority, currently not used in WEM API

	// Generate Update API request
	machineADObjectUpdateRequest := w.client.WemClient.MachineADObjectDAAS.AdObjectUpdate(ctx)
	machineADObjectUpdateRequest = machineADObjectUpdateRequest.Body(body)
	_, httpResp, err := citrixdaasclient.ExecuteWithRetry[any](machineADObjectUpdateRequest, w.client)

	// In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating WEM Directory Object with ID "+plan.Id.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Get Newly Updated Machine AD Object from remote by Id
	machineADObject, err := getMachineADObjectById(ctx, w.client, plan.Id.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, machineADObject)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
