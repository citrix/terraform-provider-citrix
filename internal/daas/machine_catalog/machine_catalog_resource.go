// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

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
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &machineCatalogResource{}
	_ resource.ResourceWithConfigure      = &machineCatalogResource{}
	_ resource.ResourceWithImportState    = &machineCatalogResource{}
	_ resource.ResourceWithValidateConfig = &machineCatalogResource{}
)

// NewMachineCatalogResource is a helper function to simplify the provider implementation.
func NewMachineCatalogResource() resource.Resource {
	return &machineCatalogResource{}
}

// machineCatalogResource is the resource implementation.
type machineCatalogResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *machineCatalogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_catalog"
}

// Schema defines the schema for the resource.
func (r *machineCatalogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchemaForMachineCatalogResource()
}

// Configure adds the provider configured client to the resource.
func (r *machineCatalogResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *machineCatalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan MachineCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var connectionType citrixorchestration.HypervisorConnectionType

	body, err := getRequestModelForCreateMachineCatalog(plan, ctx, r.client, &resp.Diagnostics, &connectionType, r.client.AuthConfig.OnPremises)

	if err != nil {
		return
	}

	createMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsCreateMachineCatalog(ctx)

	// Add domain credential header
	if plan.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_MCS) && plan.ProvisioningScheme.MachineDomainIdentity != nil {
		header := generateAdminCredentialHeader(plan)
		createMachineCatalogRequest = createMachineCatalogRequest.XAdminCredential(header)
	}

	// Add request body
	createMachineCatalogRequest = createMachineCatalogRequest.CreateMachineCatalogRequestModel(*body)

	// Make request async
	createMachineCatalogRequest = createMachineCatalogRequest.Async(true)

	// Create new machine catalog
	_, httpResp, err := citrixdaasclient.AddRequestData(createMachineCatalogRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error creating Machine Catalog", &resp.Diagnostics, 120, false)
	if err != nil {
		return
	}

	// Get the new catalog
	catalog, err := util.GetMachineCatalog(ctx, r.client, &resp.Diagnostics, plan.Name.ValueString(), true)

	if err != nil {
		return
	}

	machines, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalog.GetId())

	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, r.client, catalog, &connectionType, machines)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *machineCatalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state MachineCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalog state from Orchestration
	catalogId := state.Id.ValueString()

	catalog, _, err := readMachineCatalog(ctx, r.client, resp, catalogId)
	if err != nil {
		return
	}

	machineCatalogMachines, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalogId)
	if err != nil {
		return
	}

	// Resolve resource path for service offering and master image
	provScheme := catalog.GetProvisioningScheme()
	resourcePool := provScheme.GetResourcePool()
	hypervisor := resourcePool.GetHypervisor()
	hypervisorName := hypervisor.GetName()

	var connectionType *citrixorchestration.HypervisorConnectionType

	if hypervisorName != "" {
		hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorName)
		if err != nil {
			return
		}
		connectionType = hypervisor.GetConnectionType().Ptr()
	}
	// Overwrite items with refreshed state
	state = state.RefreshPropertyValues(ctx, r.client, catalog, connectionType, machineCatalogMachines)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *machineCatalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan MachineCatalogResourceModel
	var state MachineCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalogs from Orchestration
	catalogId := plan.Id.ValueString()
	catalogName := plan.Name.ValueString()
	catalog, err := util.GetMachineCatalog(ctx, r.client, &resp.Diagnostics, catalogId, true)

	if err != nil {
		return
	}

	var connectionType citrixorchestration.HypervisorConnectionType

	body, err := getRequestModelForUpdateMachineCatalog(plan, state, catalog, ctx, r.client, resp, &connectionType, r.client.AuthConfig.OnPremises)
	if err != nil {
		return
	}

	updateMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalog(ctx, catalogId)
	updateMachineCatalogRequest = updateMachineCatalogRequest.UpdateMachineCatalogRequestModel(*body)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateMachineCatalogRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	provisioningType, err := citrixorchestration.NewProvisioningTypeFromValue(plan.ProvisioningType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported provisioning type.",
		)

		return
	}

	if *provisioningType == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		// For manual, compare state and plan to find machines to add and delete
		addMachinesList, deleteMachinesMap := createAddAndRemoveMachinesListForManualCatalogs(state, plan)

		addMachinesToManualCatalog(ctx, r.client, resp, addMachinesList, catalogId)
		deleteMachinesFromManualCatalog(ctx, r.client, resp, deleteMachinesMap, catalogId, catalog.GetIsPowerManaged())
	} else {
		err = updateCatalogImage(ctx, r.client, resp, catalog, plan)

		if err != nil {
			return
		}

		if catalog.GetTotalCount() > int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) {
			// delete machines from machine catalog
			err = deleteMachinesFromMcsCatalog(ctx, r.client, resp, catalog, plan)
			if err != nil {
				return
			}
		}

		if catalog.GetTotalCount() < int32(plan.ProvisioningScheme.NumTotalMachines.ValueInt64()) {
			// add machines to machine catalog
			err = addMachinesToMcsCatalog(ctx, r.client, resp, catalog, plan)
			if err != nil {
				return
			}
		}
	}

	// Fetch updated machine catalog from GetMachineCatalog.
	catalog, err = util.GetMachineCatalog(ctx, r.client, &resp.Diagnostics, catalogId, true)
	if err != nil {
		return
	}

	machines, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalog.GetId())
	if err != nil {
		return
	}

	// Update resource state with updated items and timestamp
	plan = plan.RefreshPropertyValues(ctx, r.client, catalog, &connectionType, machines)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *machineCatalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state MachineCatalogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := state.Id.ValueString()

	catalog, httpResp, err := readMachineCatalog(ctx, r.client, nil, catalogId)

	if err != nil {
		if httpResp.StatusCode == http.StatusNotFound {
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Machine Catalog "+catalogId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)

		return
	}

	// Delete existing order
	catalogName := state.Name.ValueString()
	deleteMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsDeleteMachineCatalog(ctx, catalogId)
	deleteAccountOption := citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE
	deleteVmOption := false
	if catalog.ProvisioningType == citrixorchestration.PROVISIONINGTYPE_MCS {
		// If there's no provisioning scheme in state, there will not be any machines create by MCS.
		// Therefore we will just omit credential for removing machine accounts.
		if catalog.ProvisioningScheme != nil {
			// Add domain credential header
			header := generateAdminCredentialHeader(state)
			deleteMachineCatalogRequest = deleteMachineCatalogRequest.XAdminCredential(header)
		}

		deleteAccountOption = citrixorchestration.MACHINEACCOUNTDELETEOPTION_DELETE
		deleteVmOption = true
	}

	deleteMachineCatalogRequest = deleteMachineCatalogRequest.DeleteVm(deleteVmOption).DeleteAccount(deleteAccountOption).Async(true)
	httpResp, err = citrixdaasclient.AddRequestData(deleteMachineCatalogRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error deleting Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Machine Catalog "+catalogName, &resp.Diagnostics, 60, false)
	if err != nil {
		return
	}
}

func (r *machineCatalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *machineCatalogResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data MachineCatalogResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	provisioningTypeMcs := string(citrixorchestration.PROVISIONINGTYPE_MCS)
	provisioningTypeManual := string(citrixorchestration.PROVISIONINGTYPE_MANUAL)

	if data.ProvisioningType.ValueString() == provisioningTypeMcs {
		if data.ProvisioningScheme == nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Missing Attribute Configuration",
				fmt.Sprintf("Expected provisioning_scheme to be configured when value of provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		if data.ProvisioningScheme != nil && data.ProvisioningScheme.AzureMachineConfig != nil && data.ProvisioningScheme.AzureMachineConfig.WritebackCache != nil {
			wbc := data.ProvisioningScheme.AzureMachineConfig.WritebackCache
			if !wbc.PersistOsDisk.ValueBool() && wbc.PersistVm.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					path.Root("persist_vm"),
					"Incorrect Attribute Configuration",
					"persist_os_disk must be enabled to enable persist_vm.",
				)
			}

			if !wbc.PersistWBC.ValueBool() && wbc.StorageCostSaving.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					path.Root("storage_cost_saving"),
					"Incorrect Attribute Configuration",
					"persist_wbc must be enabled to enable storage_cost_saving.",
				)
			}
		}

		if data.MachineAccounts != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("machine_accounts"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("machine_accounts cannot be configured when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		if data.IsRemotePc.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_remote_pc"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Remote PC access catalog cannot be created when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		if !data.IsPowerManaged.IsNull() && !data.IsPowerManaged.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_power_managed"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Machines have to be power managed when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		data.IsPowerManaged = types.BoolValue(true) // set power managed to true for MCS catalog
	} else {
		// Manual provisioning type
		if data.IsPowerManaged.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_power_managed"),
				"Missing Attribute Configuration",
				fmt.Sprintf("expected is_power_managed to be defined when provisioning_type is %s.", provisioningTypeManual),
			)
		}

		if data.IsRemotePc.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_remote_pc"),
				"Missing Attribute Configuration",
				fmt.Sprintf(" expected is_remote_pc to be defined when provisioning_type is %s.", provisioningTypeManual),
			)
		}

		if data.ProvisioningScheme != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("provisioning_scheme cannot be configured when provisioning_type is not %s.", provisioningTypeMcs),
			)
		}

		if data.IsPowerManaged.ValueBool() {
			if data.MachineAccounts != nil {
				for _, machineAccount := range data.MachineAccounts {
					if machineAccount.Hypervisor.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("machine_accounts"),
							"Missing Attribute Configuration",
							"Expected hypervisor to be configured when machines are power managed.",
						)
					}

					for _, machine := range machineAccount.Machines {
						if machine.MachineName.IsNull() {
							resp.Diagnostics.AddAttributeError(
								path.Root("machine_accounts"),
								"Missing Attribute Configuration",
								"Expected machine_name to be configured when machines are power managed.",
							)
						}
					}
				}
			}

			if data.IsRemotePc.ValueBool() {
				resp.Diagnostics.AddAttributeError(
					path.Root("is_remote_pc"),
					"Incorrect Attribute Configuration",
					"Remote PC Access catalog cannot be power managed.",
				)
			}
		}
	}

	if data.IsRemotePc.ValueBool() {
		sessionSupport, err := citrixorchestration.NewSessionSupportFromValue(data.SessionSupport.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("session_support"),
				"Incorrect Attribute Configuration",
				"Unsupported session support.",
			)
			return
		}
		if sessionSupport != nil && *sessionSupport != citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION {
			resp.Diagnostics.AddAttributeError(
				path.Root("session_support"),
				"Incorrect Attribute Configuration",
				"Only Single Session is supported for Remote PC Access catalog.",
			)
		}
	}
}
