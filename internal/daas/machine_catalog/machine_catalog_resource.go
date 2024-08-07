// Copyright © 2024. Citrix Systems, Inc.

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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &machineCatalogResource{}
	_ resource.ResourceWithConfigure      = &machineCatalogResource{}
	_ resource.ResourceWithImportState    = &machineCatalogResource{}
	_ resource.ResourceWithValidateConfig = &machineCatalogResource{}
	_ resource.ResourceWithModifyPlan     = &machineCatalogResource{}
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
	resp.Schema = MachineCatalogResourceModel{}.GetSchema()
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

	body, err := getRequestModelForCreateMachineCatalog(plan, ctx, r.client, &resp.Diagnostics, r.client.AuthConfig.OnPremises)
	if err != nil {
		return
	}

	createMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsCreateMachineCatalog(ctx)

	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)

	// Add domain credential header
	if (plan.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_MCS) || plan.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING)) && !provSchemeModel.MachineDomainIdentity.IsNull() {
		header := generateAdminCredentialHeader(util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeModel.MachineDomainIdentity))
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

	hypervisorConnection := catalog.GetHypervisorConnection()
	hypervisorId := hypervisorConnection.GetId()
	var connectionType citrixorchestration.HypervisorConnectionType
	var pluginId string
	if hypervisorId != "" {
		hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
		if err != nil {
			return
		}

		connectionType = hypervisor.GetConnectionType()
		pluginId = hypervisor.GetPluginId()
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, &connectionType, machines, pluginId)

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
	var pluginId string
	if hypervisorName != "" {
		hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorName)
		if err != nil {
			return
		}
		connectionType = hypervisor.GetConnectionType().Ptr()
		pluginId = hypervisor.GetPluginId()
	}
	// Overwrite items with refreshed state
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, connectionType, machineCatalogMachines, pluginId)

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

	body, err := getRequestModelForUpdateMachineCatalog(plan, ctx, r.client, resp, r.client.AuthConfig.OnPremises)
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
		addMachinesList, deleteMachinesMap := createAddAndRemoveMachinesListForManualCatalogs(ctx, &resp.Diagnostics, state, plan)

		addMachinesToManualCatalog(ctx, &resp.Diagnostics, r.client, resp, addMachinesList, catalogId)
		deleteMachinesFromManualCatalog(ctx, r.client, resp, deleteMachinesMap, catalogId)
	} else {
		provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		err = updateCatalogImageAndMachineProfile(ctx, r.client, resp, catalog, plan, provisioningType)

		if err != nil {
			return
		}

		if catalog.GetTotalCount() > int32(provSchemeModel.NumTotalMachines.ValueInt64()) {
			// delete machines from machine catalog
			err = deleteMachinesFromMcsPvsCatalog(ctx, r.client, resp, catalog, provSchemeModel)
			if err != nil {
				return
			}
		}

		if catalog.GetTotalCount() < int32(provSchemeModel.NumTotalMachines.ValueInt64()) {
			// add machines to machine catalog
			err = addMachinesToMcsPvsCatalog(ctx, r.client, resp, catalog, provSchemeModel)
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

	hypervisorConnection := catalog.GetHypervisorConnection()
	hypervisorId := hypervisorConnection.GetId()
	var connectionType citrixorchestration.HypervisorConnectionType
	var pluginId string
	if hypervisorId != "" {
		hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorId)
		if err != nil {
			return
		}

		connectionType = hypervisor.GetConnectionType()
		pluginId = hypervisor.GetPluginId()
	}

	// Update resource state with updated items and timestamp
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, &connectionType, machines, pluginId)

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
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MCS || catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING {
		provScheme := catalog.GetProvisioningScheme()
		identityType := provScheme.GetIdentityType()

		if identityType == citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY || identityType == citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD {
			// If there's no provisioning scheme in state, there will not be any machines create by MCS.
			// Therefore we will just omit credential for removing machine accounts.
			if !state.ProvisioningScheme.IsNull() {
				// Add domain credential header
				provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, state.ProvisioningScheme)
				header := generateAdminCredentialHeader(util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeModel.MachineDomainIdentity))
				deleteMachineCatalogRequest = deleteMachineCatalogRequest.XAdminCredential(header)
			}

			deleteAccountOption = citrixorchestration.MACHINEACCOUNTDELETEOPTION_DELETE
		}

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
	defer util.PanicHandler(&resp.Diagnostics)

	var data MachineCatalogResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)

	sessionSupportMultiSession := string(citrixorchestration.SESSIONSUPPORT_MULTI_SESSION)
	allocationTypeStatic := string(citrixorchestration.ALLOCATIONTYPE_STATIC)
	if data.SessionSupport.ValueString() == sessionSupportMultiSession && data.AllocationType.ValueString() == allocationTypeStatic {
		resp.Diagnostics.AddAttributeError(
			path.Root("allocation_type"),
			"Incorrect Attribute Configuration",
			"Static allocation type is not supported by MultiSession session type machine catalog.",
		)
	}

	provisioningTypeMcs := string(citrixorchestration.PROVISIONINGTYPE_MCS)
	provisioningTypeManual := string(citrixorchestration.PROVISIONINGTYPE_MANUAL)
	provisioningTypePvsStreaming := string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING)

	azureAD := string(citrixorchestration.IDENTITYTYPE_AZURE_AD)

	if data.ProvisioningType.ValueString() == provisioningTypeMcs {
		if data.ProvisioningScheme.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Missing Attribute Configuration",
				fmt.Sprintf("Expected provisioning_scheme to be configured when value of provisioning_type is %s.", provisioningTypeMcs),
			)
		} else {
			// Validate Provisioning Scheme
			provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, data.ProvisioningScheme)
			if !provSchemeModel.AzureMachineConfig.IsNull() {
				azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AzureMachineConfig)
				// Validate Azure Machine Config
				if !azureMachineConfigModel.WritebackCache.IsNull() {
					// Validate Writeback Cache
					azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, &resp.Diagnostics, azureMachineConfigModel.WritebackCache)

					if azureWbcModel.PersistWBC.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("persist_wbc"),
							"Missing Attribute Configuration",
							fmt.Sprintf("persist_wbc for writeback_cache under azure_machine_config must be set when provisioning_type is %s.", provisioningTypeMcs),
						)
					}

					if azureWbcModel.StorageCostSaving.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("storage_cost_saving"),
							"Incorrect Attribute Configuration",
							fmt.Sprintf("storage_cost_saving for writeback_cache under azure_machine_config must be set when provisioning_type is %s.", provisioningTypeMcs),
						)
					}

					if azureWbcModel.WriteBackCacheMemorySizeMB.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("writeback_cache_memory_size_mb"),
							"Incorrect Attribute Configuration",
							fmt.Sprintf("writeback_cache_memory_size_mb for writeback_cache under azure_machine_config must be set when provisioning_type is %s.", provisioningTypeMcs),
						)
					}

					if !azureWbcModel.PersistOsDisk.ValueBool() && azureWbcModel.PersistVm.ValueBool() {
						resp.Diagnostics.AddAttributeError(
							path.Root("persist_vm"),
							"Incorrect Attribute Configuration",
							"persist_os_disk for writeback_cache under azure_machine_config must be enabled to enable persist_vm.",
						)
					}

					if !azureWbcModel.PersistWBC.ValueBool() && azureWbcModel.StorageCostSaving.ValueBool() {
						resp.Diagnostics.AddAttributeError(
							path.Root("storage_cost_saving"),
							"Incorrect Attribute Configuration",
							"persist_wbc for writeback_cache under azure_machine_config must be enabled to enable storage_cost_saving.",
						)
					}
				}

				// Validate Azure Intune Enrollment
				if provSchemeModel.IdentityType.ValueString() != azureAD &&
					!azureMachineConfigModel.EnrollInIntune.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("enroll_in_intune"),
						"Incorrect Attribute Configuration",
						"enroll_in_intune can only be configured when identity_type is Azure AD.",
					)
				}
				if data.AllocationType.ValueString() != allocationTypeStatic &&
					azureMachineConfigModel.EnrollInIntune.ValueBool() {
					resp.Diagnostics.AddAttributeError(
						path.Root("enroll_in_intune"),
						"Incorrect Attribute Configuration",
						fmt.Sprintf("Azure Intune auto enrollment is only supported when `allocation_type` is %s.", allocationTypeStatic),
					)
				}

				if !azureMachineConfigModel.AzurePvsConfiguration.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("azure_pvs_config"),
						"Incorrect Attribute Configuration",
						fmt.Sprintf("azure_pvs_config is not supported when provisioning_type is %s.", provisioningTypeMcs),
					)
				}

				if azureMachineConfigModel.StorageType.ValueString() == util.AzureEphemeralOSDisk {
					// Validate Azure Ephemeral OS Disk
					if !azureMachineConfigModel.UseManagedDisks.ValueBool() {
						resp.Diagnostics.AddAttributeError(
							path.Root("use_managed_disks"),
							"Incorrect Attribute Configuration",
							fmt.Sprintf("use_managed_disks must be set to true when storage_type is %s.", util.AzureEphemeralOSDisk),
						)
					}

					if azureMachineConfigModel.UseAzureComputeGallery.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("use_azure_compute_gallery"),
							"Missing Attribute Configuration",
							fmt.Sprintf("use_azure_compute_gallery must be set when storage_type is %s.", util.AzureEphemeralOSDisk),
						)
					}
				}

				if !azureMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, azureMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.AwsMachineConfig.IsNull() {
				awsMachineConfigModel := util.ObjectValueToTypedObject[AwsMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AwsMachineConfig)
				if !awsMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, awsMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.GcpMachineConfig.IsNull() {
				gcpMachineConfigModel := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.GcpMachineConfig)
				if !gcpMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, gcpMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.VsphereMachineConfig.IsNull() {
				vSphereMachineConfigModel := util.ObjectValueToTypedObject[VsphereMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.VsphereMachineConfig)
				if !vSphereMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, vSphereMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.XenserverMachineConfig.IsNull() {
				xenserverMachineConfigModel := util.ObjectValueToTypedObject[XenserverMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.XenserverMachineConfig)
				if !xenserverMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, xenserverMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.NutanixMachineConfig.IsNull() {
				nutanixMachineConfigModel := util.ObjectValueToTypedObject[NutanixMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.NutanixMachineConfig)
				if !nutanixMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, nutanixMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}

			if !provSchemeModel.SCVMMMachineConfigModel.IsNull() {
				scvmmMachineConfigModel := util.ObjectValueToTypedObject[SCVMMMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.SCVMMMachineConfigModel)
				if !scvmmMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, scvmmMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}
			}
		}

		if !data.MachineAccounts.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("machine_accounts"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("machine_accounts cannot be configured when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		if !data.IsRemotePc.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_remote_pc"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("is_remote_pc cannot be configured when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		if !data.IsPowerManaged.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_power_managed"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("is_power_managed cannot be configured when provisioning_type is %s.", provisioningTypeMcs),
			)
		}

		data.IsPowerManaged = types.BoolValue(true) // set power managed to true for MCS catalog
	} else if data.ProvisioningType.ValueString() == provisioningTypePvsStreaming {
		// Add checks for PVSStreaming catalogs
		if data.ProvisioningScheme.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Missing Attribute Configuration",
				fmt.Sprintf("Expected provisioning_scheme to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
			)
		} else {
			// Validate Provisioning Scheme
			provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, data.ProvisioningScheme)
			if provSchemeModel.AzureMachineConfig.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("azure_machine_config"),
					"Missing Attribute Configuration",
					fmt.Sprintf("PVS Catalogs are currently only supported for Azure environment. Expected azure_machine_config to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AzureMachineConfig)

			if azureMachineConfigModel.AzurePvsConfiguration.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("azure_pvs_config"),
					"Missing Attribute Configuration",
					fmt.Sprintf("Expected azure_pvs_config to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if !azureMachineConfigModel.AzureMasterImage.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("azure_master_image"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("azure_master_image cannot be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if azureMachineConfigModel.MasterImageNote.ValueString() != "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("master_image_note"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("master_image_note cannot be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if azureMachineConfigModel.StorageType.ValueString() == util.AzureEphemeralOSDisk {
				resp.Diagnostics.AddAttributeError(
					path.Root("storage_type"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("Storage type cannot be set to Azure_Ephemeral_OS_Disk when provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if !azureMachineConfigModel.UseAzureComputeGallery.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("use_azure_compute_gallery"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("use_azure_compute_gallery cannot be configured when provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if !azureMachineConfigModel.EnrollInIntune.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("enroll_in_intune"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("enroll_in_intune cannot be configured when provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if !azureMachineConfigModel.DiskEncryptionSet.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("disk_encryption_set"),
					"Incorrect Attribute Configuration",
					fmt.Sprintf("disk_encryption_set cannot be configured when provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if azureMachineConfigModel.MachineProfile.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("machine_profile"),
					"Missing Attribute Configuration",
					fmt.Sprintf("Expected machine_profile to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if azureMachineConfigModel.WritebackCache.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("writeback_cache"),
					"Missing Attribute Configuration",
					fmt.Sprintf("Expected writeback_cache to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			} else {

				azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, &resp.Diagnostics, azureMachineConfigModel.WritebackCache)
				if azureWbcModel.PersistWBC.IsNull() || !azureWbcModel.PersistWBC.ValueBool() {
					resp.Diagnostics.AddAttributeError(
						path.Root("persist_wbc"),
						"Incorrect Attribute Configuration",
						fmt.Sprintf("persist_wbc for writeback_cache under azure_machine_config needs to be set to true for provisioning type %s.", provisioningTypePvsStreaming),
					)
				}

				if !azureWbcModel.StorageCostSaving.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("storage_cost_saving"),
						"Incorrect Attribute Configuration",
						fmt.Sprintf("storage_cost_saving for writeback_cache under azure_machine_config cannot be configured when provisioning_type is %s.", provisioningTypePvsStreaming),
					)
				}
			}

		}

		if !data.MachineAccounts.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("machine_accounts"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("machine_accounts cannot be configured when provisioning_type is %s.", provisioningTypePvsStreaming),
			)
		}

		if data.IsRemotePc.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_remote_pc"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Remote PC access catalog cannot be created when provisioning_type is %s.", provisioningTypePvsStreaming),
			)
		}

	} else if data.ProvisioningType.ValueString() == provisioningTypeManual {
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

		if !data.ProvisioningScheme.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("provisioning_scheme cannot be configured when provisioning_type is not %s or %s.", provisioningTypeMcs, provisioningTypePvsStreaming),
			)
		}

		if data.IsPowerManaged.ValueBool() {
			if !data.MachineAccounts.IsNull() {
				machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, &resp.Diagnostics, data.MachineAccounts)
				for _, machineAccount := range machineAccounts {
					if machineAccount.Hypervisor.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("machine_accounts"),
							"Missing Attribute Configuration",
							"Expected hypervisor to be configured when machines are power managed.",
						)
					}

					machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, &resp.Diagnostics, machineAccount.Machines)
					for _, machine := range machines {
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

	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, data.ProvisioningScheme)
	if !data.ProvisioningScheme.IsNull() && !provSchemeModel.CustomProperties.IsNull() {
		customProperties := util.ObjectListToTypedArray[CustomPropertyModel](ctx, &resp.Diagnostics, provSchemeModel.CustomProperties)
		for _, customProperty := range customProperties {
			propertyName := customProperty.Name.ValueString()
			if val, ok := MappedCustomProperties[propertyName]; ok {
				resp.Diagnostics.AddAttributeError(
					path.Root("custom_properties"),
					"Duplicated Custom Property",
					fmt.Sprintf("Use Terraform field \"%s\" for custom property \"%s\".", val, propertyName),
				)
			}
		}
	}
}

func (r *machineCatalogResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}
