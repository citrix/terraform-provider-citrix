// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

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

	catalogNameExists := checkIfCatalogNameExists(ctx, r.client, plan.Name.ValueString())

	if catalogNameExists {
		// Validate machine catalog name uniqueness for create
		resp.Diagnostics.AddError(
			"Machine Catalog Name Already Exists",
			fmt.Sprintf("A Machine Catalog with the name '%s' already exists. Please choose a different name.", plan.Name.ValueString()),
		)
		return
	}

	if !plan.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		if !provSchemePlan.MachineADAccounts.IsNull() {
			machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, provSchemePlan.MachineADAccounts)
			if len(machineAccountsInPlan) > 0 {
				err := validateInUseMachineAccounts(ctx, &resp.Diagnostics, r.client, plan.Name.ValueString(), provSchemePlan, machineAccountsInPlan)
				if err != nil {
					return
				}
			}
		}
	}

	body, err := getRequestModelForCreateMachineCatalog(plan, ctx, r.client, &resp.Diagnostics, r.client.AuthConfig.OnPremises)
	if err != nil {
		return
	}

	createMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsCreateMachineCatalog(ctx)

	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)

	// Add domain credential header
	if (plan.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_MCS) || plan.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING)) && !provSchemeModel.MachineDomainIdentity.IsNull() {
		machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeModel.MachineDomainIdentity)
		if !machineDomainIdentityModel.ServiceAccount.IsNull() { // If service account is not provided, no need to create X-AdminCredential header since ServiceAccountId is being used
			header := generateAdminCredentialHeader(machineDomainIdentityModel)
			createMachineCatalogRequest = createMachineCatalogRequest.XAdminCredential(header)
		}
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

	timeoutConfigs := util.ObjectValueToTypedObject[MachineCatalogTimeout](ctx, &resp.Diagnostics, plan.Timeout)
	createTimeout := timeoutConfigs.Create.ValueInt32()
	if createTimeout == 0 {
		createTimeout = getMachineCatalogTimeoutConfigs().CreateDefault
	}
	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error creating Machine Catalog", &resp.Diagnostics, createTimeout)
	if errors.Is(err, &util.JobPollError{}) {
		return
	} // if the job failed continue processing

	machineCatalogPath := util.BuildResourcePathForGetRequest(plan.MachineCatalogFolderPath.ValueString(), plan.Name.ValueString())
	setMachineCatalogTags(ctx, &resp.Diagnostics, r.client, machineCatalogPath, plan.Tags)

	// Get the new catalog
	catalog, err := util.GetMachineCatalog(ctx, r.client, &resp.Diagnostics, machineCatalogPath, true)
	if err != nil {
		return
	}

	machines, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalog.GetId())
	if err != nil {
		return
	}

	hypervisorConnection := catalog.GetHypervisorConnection()
	hypervisorNameOrId := hypervisorConnection.GetId()
	// If hypervisor ID is not set, use the hypervisor name
	if hypervisorNameOrId == "" {
		hypervisorNameOrId = hypervisorConnection.GetName()
	}
	var connectionType citrixorchestration.HypervisorConnectionType
	var pluginId string
	if hypervisorNameOrId != "" {
		hypervisor, err := util.GetHypervisor(ctx, r.client, &resp.Diagnostics, hypervisorNameOrId)
		if err != nil {
			return
		}

		connectionType = hypervisor.GetConnectionType()
		pluginId = hypervisor.GetPluginId()
	}

	tags := getMachineCatalogTags(ctx, &resp.Diagnostics, r.client, catalog.GetId())

	var machineAdAccounts []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		machineAdAccounts = []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel{}
	} else {
		machineAdAccounts, err = getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, r.client, catalog.GetId())
		if err != nil {
			return
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, &connectionType, machines, pluginId, tags, machineAdAccounts)

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

	tags := getMachineCatalogTags(ctx, &resp.Diagnostics, r.client, catalog.GetId())

	var machineAdAccounts []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel
	if provScheme.GetId() == "" || catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		// if the provScheme doesn't exist or the provisioning type is manual, then there are no machine accounts to fetch
		machineAdAccounts = []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel{}
	} else {
		machineAdAccounts, err = getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, r.client, catalog.GetId())
		if err != nil {
			return
		}
	}

	// Overwrite items with refreshed state
	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, connectionType, machineCatalogMachines, pluginId, tags, machineAdAccounts)

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

	var machineAdAccounts []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		machineAdAccounts = []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel{}
	} else {
		machineAdAccounts, err = getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, r.client, catalog.GetId())
		if err != nil {
			return
		}
	}

	if !plan.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, provSchemePlan.MachineADAccounts)

		machineAccountsToBeAdded := []MachineADAccountModel{}
		for _, machineAccountInPlan := range machineAccountsInPlan {
			if !slices.ContainsFunc(machineAdAccounts, func(machineAccountInCatalog citrixorchestration.ProvisioningSchemeMachineAccountResponseModel) bool {
				return strings.EqualFold(machineAccountInCatalog.GetSamName(), machineAccountInPlan.ADAccountName.ValueString())
			}) {
				machineAccountsToBeAdded = append(machineAccountsToBeAdded, machineAccountInPlan)
			}
		}
		if len(machineAccountsToBeAdded) > 0 {
			err := validateInUseMachineAccounts(ctx, &resp.Diagnostics, r.client, plan.Name.ValueString(), provSchemePlan, machineAccountsToBeAdded)
			if err != nil {
				return
			}
		}
	}

	body, err := getRequestModelForUpdateMachineCatalog(plan, state, ctx, r.client, resp, r.client.AuthConfig.OnPremises)
	if err != nil {
		return
	}

	updateMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalog(ctx, catalogId)
	updateMachineCatalogRequest = updateMachineCatalogRequest.UpdateMachineCatalogRequestModel(*body)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateMachineCatalogRequest, r.client).Async(true).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	timeoutConfigs := util.ObjectValueToTypedObject[MachineCatalogTimeout](ctx, &resp.Diagnostics, plan.Timeout)
	updateTimeout := timeoutConfigs.Update.ValueInt32()
	if updateTimeout == 0 {
		updateTimeout = getMachineCatalogTimeoutConfigs().UpdateDefault
	}
	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error updating Machine Catalog "+catalogName, &resp.Diagnostics, updateTimeout)
	if err != nil {
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
		planProvSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		stateProvSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, state.ProvisioningScheme)

		// During update, the MachineAccountCreationRules in the terraform config stick to the updatePlan.
		// It uses the same StartsWith value for creating new machines which causes conflict with existing machines and throws error.
		// In order to avoid this conflict, we set the StartsWith value to null if the MachineAccountCreationRules in plan and state are equal.
		planRules := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, planProvSchemeModel.MachineAccountCreationRules)
		stateRules := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, stateProvSchemeModel.MachineAccountCreationRules)

		if stateRules.Equals(planRules) {
			planRules.StartsWith = types.StringNull()
			planProvSchemeModel.MachineAccountCreationRules = util.TypedObjectToObjectValue(ctx, &resp.Diagnostics, planRules)
		}

		machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, planProvSchemeModel.MachineADAccounts)
		err = updateCatalogImageAndMachineProfile(ctx, r.client, resp, catalog, plan, provisioningType, updateTimeout)

		if err != nil {
			return
		}

		if catalog.GetTotalCount() > int32(planProvSchemeModel.NumTotalMachines.ValueInt64()) {
			// delete machines from machine catalog
			err = deleteMachinesFromMcsPvsCatalog(ctx, r.client, resp, catalog, planProvSchemeModel, machineAccountsInPlan)
			if err != nil {
				return
			}
		}
		machineAccountsInCatalog, err := getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, r.client, catalog.GetId())
		if err != nil {
			return
		}
		machineAccountsInState := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, stateProvSchemeModel.MachineADAccounts)

		machineAccountsToAdd := []MachineADAccountModel{}
		machineAccountsToDelete := []MachineADAccountModel{}

		// Delete Machine Accounts that were present in the state but not in the plan
		for _, machineAccountInState := range machineAccountsInState {
			if machineAccountInState.State.ValueString() != string(citrixorchestration.PROVISIONINGSCHEMEMACHINEACCOUNTSTATE_IN_USE) && !slices.ContainsFunc(machineAccountsInPlan, func(machineAccountInPlan MachineADAccountModel) bool {
				return strings.EqualFold(machineAccountInPlan.ADAccountName.ValueString(), machineAccountInState.ADAccountName.ValueString())
			}) {
				machineAccountsToDelete = append(machineAccountsToDelete, machineAccountInState)
			}
		}

		machineAccountsToUpdate := []MachineADAccountModel{}
		for _, machineAccountInPlan := range machineAccountsInPlan {
			machineAccountIndex := slices.IndexFunc(machineAccountsInCatalog, func(machineAccountInCatalog citrixorchestration.ProvisioningSchemeMachineAccountResponseModel) bool {
				return strings.EqualFold(machineAccountInPlan.ADAccountName.ValueString(), machineAccountInCatalog.GetSamName())
			})
			if machineAccountIndex >= 0 {
				machineState := machineAccountsInCatalog[machineAccountIndex].GetState()
				if machineState != citrixorchestration.PROVISIONINGSCHEMEMACHINEACCOUNTSTATE_IN_USE &&
					machineState != citrixorchestration.PROVISIONINGSCHEMEMACHINEACCOUNTSTATE_AVAILABLE {
					machineAccountsToUpdate = append(machineAccountsToUpdate, machineAccountInPlan)
				}
			} else {
				machineAccountsToAdd = append(machineAccountsToAdd, machineAccountInPlan)
			}
		}

		// Remove Machine Accounts from Catalog here
		if len(machineAccountsToDelete) > 0 {
			batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, r.client, planProvSchemeModel, true)
			txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nCould not delete machine account(s) from Machine Catalog, unexpected error: "+util.ReadClientError(err),
				)
				return
			}

			batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
			for i, machineAccount := range machineAccountsToDelete {
				relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/MachineAccounts/%s", catalogId, machineAccount.ADAccountName.ValueString())
				var batchRequestItem citrixorchestration.BatchRequestItemModel
				batchRequestItem.SetMethod(http.MethodDelete)
				batchRequestItem.SetReference(strconv.Itoa(i))
				batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
				batchRequestItem.SetHeaders(batchApiHeaders)
				batchRequestItems = append(batchRequestItems, batchRequestItem)
			}

			var batchRequestModel citrixorchestration.BatchRequestModel
			batchRequestModel.SetItems(batchRequestItems)
			successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error deleting machine account(s) from Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}

			if successfulJobs < len(machineAccountsToDelete) {
				errMsg := fmt.Sprintf("An error occurred while deleting machine account(s) from the Machine Catalog. %d of %d machine account(s) were deleted from the Machine Catalog.", successfulJobs, len(machineAccountsToDelete))
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\n"+errMsg,
				)
				return
			}
		}

		if len(machineAccountsToUpdate) > 0 {
			batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, r.client, planProvSchemeModel, true)
			txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nCould not update machine account(s) in Machine Catalog, unexpected error: "+util.ReadClientError(err),
				)
				return
			}

			batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
			for i, machineAccount := range machineAccountsToUpdate {
				resetPassword := machineAccount.ResetPassword.ValueBool()
				updateAdMachineBody := citrixorchestration.UpdateMachineAccountRequestModel{}
				passwordFormat, err := citrixorchestration.NewIdentityPasswordFormatFromValue(machineAccount.PasswordFormat.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating Machine Catalog",
						fmt.Sprintf("Unsupported machine account password format type. Error: %s", err.Error()),
					)
					return
				}
				updateAdMachineBody.SetPasswordFormat(*passwordFormat)
				updateAdMachineBody.SetResetPassword(resetPassword)
				if !resetPassword {
					updateAdMachineBody.SetPassword(machineAccount.Password.ValueString())
				}

				updateAccountRequestString, err := util.ConvertToString(updateAdMachineBody)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating machine account(s) for Machine Catalog "+catalogName,
						"An unexpected error occurred: "+err.Error(),
					)
					return
				}

				relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/MachineAccounts/%s", catalogId, machineAccount.ADAccountName.ValueString())
				var batchRequestItem citrixorchestration.BatchRequestItemModel
				batchRequestItem.SetMethod(http.MethodPatch)
				batchRequestItem.SetReference(strconv.Itoa(i))
				batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
				batchRequestItem.SetHeaders(batchApiHeaders)
				batchRequestItem.SetBody(updateAccountRequestString)
				batchRequestItems = append(batchRequestItems, batchRequestItem)
			}

			var batchRequestModel citrixorchestration.BatchRequestModel
			batchRequestModel.SetItems(batchRequestItems)
			successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating machine account(s) from Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}

			if successfulJobs < len(machineAccountsToUpdate) {
				errMsg := fmt.Sprintf("An error occurred while updating machine account(s) for the Machine Catalog. %d of %d machine account(s) were updated in the Machine Catalog.", successfulJobs, len(machineAccountsToUpdate))
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\n"+errMsg,
				)
				return
			}
		}

		// Add new Machine Accounts to Catalog. Accounts to be add will be passed to the addMachinesToMcsPvsCatalog function
		if len(machineAccountsToAdd) > 0 {
			batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, r.client, planProvSchemeModel, true)
			txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nCould not add machine account(s) to Machine Catalog, unexpected error: "+util.ReadClientError(err),
				)
				return
			}

			batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
			relativeUrl := fmt.Sprintf("/MachineCatalogs/%s/MachineAccounts", catalogId)
			addMachineAccountRequests, err := constructAvailableMachineAccountsRequestModel(&resp.Diagnostics, machineAccountsToAdd, "updating")
			if err != nil {
				return
			}
			for i, addAccountRequest := range addMachineAccountRequests {
				addAccountRequestString, err := util.ConvertToString(addAccountRequest)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error adding machine account(s) to Machine Catalog "+catalogName,
						"An unexpected error occurred: "+err.Error(),
					)
					return
				}

				var batchRequestItem citrixorchestration.BatchRequestItemModel
				batchRequestItem.SetMethod(http.MethodPost)
				batchRequestItem.SetReference(strconv.Itoa(i))
				batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
				batchRequestItem.SetHeaders(batchApiHeaders)
				batchRequestItem.SetBody(addAccountRequestString)
				batchRequestItems = append(batchRequestItems, batchRequestItem)
			}

			var batchRequestModel citrixorchestration.BatchRequestModel
			batchRequestModel.SetItems(batchRequestItems)
			successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error adding machine account(s) to Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}

			if successfulJobs < len(machineAccountsToAdd) {
				errMsg := fmt.Sprintf("An error occurred while adding machine account(s) to the Machine Catalog. %d of %d machine account(s) were adding to the Machine Catalog.", successfulJobs, len(machineAccountsToAdd))
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+txId+
						"\n"+errMsg,
				)
				return
			}
		}

		if catalog.GetTotalCount() < int32(planProvSchemeModel.NumTotalMachines.ValueInt64()) {
			// add machines to machine catalog
			err = addMachinesToMcsPvsCatalog(ctx, r.client, resp, catalog, planProvSchemeModel)
			if err != nil {
				return
			}
		}
	}

	setMachineCatalogTags(ctx, &resp.Diagnostics, r.client, catalogId, plan.Tags)

	// Update Machine Catalog Provisioning Scheme Metadata
	if !plan.ProvisioningScheme.IsNull() && !state.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		provSchemeState := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, state.ProvisioningScheme)
		if !provSchemePlan.Metadata.Equal(provSchemeState.Metadata) {
			metadata := util.GetMetadataRequestModel(ctx, &resp.Diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, provSchemePlan.Metadata))
			updateProvSchemeReqBody := citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel{}
			updateProvSchemeReqBody.SetMetadata(metadata)

			updateProvSchemeMetadataReq := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsUpdateMachineCatalogProvisioningScheme(ctx, catalogId)
			updateProvSchemeMetadataReq = updateProvSchemeMetadataReq.UpdateMachineCatalogProvisioningSchemeRequestModel(updateProvSchemeReqBody)
			_, httpResp, err := citrixdaasclient.AddRequestData(updateProvSchemeMetadataReq, r.client).Async(true).Execute()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+catalogName,
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}

			err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error updating Machine Catalog "+catalogName, &resp.Diagnostics, 10)
			if err != nil {
				return
			}
		}
	}

	if !plan.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		if provSchemePlan.ApplyUpdatesToExistingMachines.ValueBool() {
			machines, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalog.GetId())
			if err != nil {
				return
			}

			for _, machine := range machines {
				machineSid := machine.GetSid()
				applyUpdateRequest := r.client.ApiClient.ProvisionedVirtualMachineAPIsDAAS.ProvisionedVirtualMachineApplyProvisionedVirtualMachineConfigurationUpdate(ctx, machineSid)
				httpResp, err := citrixdaasclient.AddRequestData(applyUpdateRequest, r.client).Execute()
				if err != nil {
					resp.Diagnostics.AddWarning(
						"Error applying update to machine "+machine.GetName(),
						"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
							"\nError message: "+util.ReadClientError(err),
					)
				}
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

	tags := getMachineCatalogTags(ctx, &resp.Diagnostics, r.client, catalogId)

	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		machineAdAccounts = []citrixorchestration.ProvisioningSchemeMachineAccountResponseModel{}
	} else {
		machineAdAccounts, err = getMachineCatalogMachineADAccounts(ctx, &resp.Diagnostics, r.client, catalog.GetId())
		if err != nil {
			return
		}
	}

	// Update resource state with updated items and timestamp
	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, r.client, catalog, &connectionType, machines, pluginId, tags, machineAdAccounts)

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

	// Set machines to maintenance mode before deletion
	machinesInCatalog, err := util.GetMachineCatalogMachines(ctx, r.client, &resp.Diagnostics, catalogId)
	if err != nil {
		return
	}
	machinesNeedToSetToMaintenanceMode := []citrixorchestration.MachineResponseModel{}
	for _, machine := range machinesInCatalog {
		if machine.DeliveryGroup == nil {
			// if machine has no delivery group, there is no need to put it in maintenance mode
			continue
		}
		isMachineInMaintenanceMode := machine.GetInMaintenanceMode()
		if !isMachineInMaintenanceMode {
			machinesNeedToSetToMaintenanceMode = append(machinesNeedToSetToMaintenanceMode, machine)
		}
	}
	if len(machinesNeedToSetToMaintenanceMode) > 0 {
		resp.Diagnostics.AddError(
			"Error deleting Machine Catalog "+catalog.GetName(),
			fmt.Sprintf("Machine catalog %s has %d machine(s) associated with delivery group(s) and need to be put in maintenance mode before deletion.", catalog.GetName(), len(machinesNeedToSetToMaintenanceMode)),
		)
		return
	}

	// Delete existing order
	catalogName := state.Name.ValueString()
	deleteMachineCatalogRequest := r.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsDeleteMachineCatalog(ctx, catalogId)
	deleteAccountOption := getMachineAccountDeleteOptionValue(state.DeleteMachineAccounts.ValueString())
	// Set default delete VM option to true for MCS and PVS Streaming
	deleteVmOption := true
	// Set delete VM option to false for manual provisioning type
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		deleteVmOption = false
	}

	// Override delete VM option with user specified value
	if !state.DeleteVirtualMachines.IsNull() {
		deleteVmOption = state.DeleteVirtualMachines.ValueBool()
	}

	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MCS || catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING {
		provScheme := catalog.GetProvisioningScheme()
		identityType := provScheme.GetIdentityType()

		if identityType == citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY || identityType == citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD {
			// If there's no provisioning scheme in state, there will not be any machines create by MCS.
			// Therefore we will just omit credential for removing machine accounts.
			if !state.ProvisioningScheme.IsNull() {
				// Add domain credential header
				provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, state.ProvisioningScheme)
				machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeModel.MachineDomainIdentity)
				if !machineDomainIdentityModel.ServiceAccount.IsNull() { // If service account is not provided, no need to create X-AdminCredential header since ServiceAccountId is being used
					header := generateAdminCredentialHeader(machineDomainIdentityModel)
					deleteMachineCatalogRequest = deleteMachineCatalogRequest.XAdminCredential(header)
				}
			}
		}
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

	timeoutConfigs := util.ObjectValueToTypedObject[MachineCatalogTimeout](ctx, &resp.Diagnostics, state.Timeout)
	deleteTimeout := timeoutConfigs.Delete.ValueInt32()
	if deleteTimeout == 0 {
		deleteTimeout = getMachineCatalogTimeoutConfigs().DeleteDefault
	}
	err = util.ProcessAsyncJobResponse(ctx, r.client, httpResp, "Error deleting Machine Catalog "+catalogName, &resp.Diagnostics, deleteTimeout)
	if errors.Is(err, &util.JobPollError{}) {
		return
	} // if the job failed continue processing
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

	if data.SessionSupport.ValueString() == string(citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION) && data.AllocationType.ValueString() == string(citrixorchestration.ALLOCATIONTYPE_RANDOM) && data.PersistUserChanges.ValueString() == string(citrixorchestration.PERSISTCHANGES_ON_LOCAL) {
		resp.Diagnostics.AddAttributeError(
			path.Root("persist_user_changes"),
			"Incorrect Attribute Configuration",
			"persist_user_changes cannot be set to OnLocal when session_support is set to Static and allocation_type is set to Random.",
		)
		return
	}

	if !data.Metadata.IsNull() {
		metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, data.Metadata)
		isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
		if !isValid {
			return
		}
	}

	provisioningTypeMcs := string(citrixorchestration.PROVISIONINGTYPE_MCS)
	provisioningTypeManual := string(citrixorchestration.PROVISIONINGTYPE_MANUAL)
	provisioningTypePvsStreaming := string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING)

	azureAD := string(citrixorchestration.IDENTITYTYPE_AZURE_AD)

	if data.ProvisioningType.ValueString() == provisioningTypeMcs {
		err := validateMcsPvsMachineCatalogDeleteOptions(&resp.Diagnostics, data)
		if err != nil {
			return
		}

		if !data.ProvisioningScheme.IsUnknown() && data.ProvisioningScheme.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Missing Attribute Configuration",
				fmt.Sprintf("Expected provisioning_scheme to be configured when value of provisioning_type is %s.", provisioningTypeMcs),
			)
		} else {
			// Validate Provisioning Scheme
			provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, data.ProvisioningScheme)

			if !provSchemeModel.Metadata.IsNull() {
				metadata := util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, &resp.Diagnostics, provSchemeModel.Metadata)
				isValid := util.ValidateMetadataConfig(ctx, &resp.Diagnostics, metadata)
				if !isValid {
					return
				}
			}

			if !provSchemeModel.AzureMachineConfig.IsNull() {
				azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AzureMachineConfig)
				// Validate Azure Machine Config
				// Validate Azure Master Image
				if (!azureMachineConfigModel.AzureMasterImage.IsUnknown() && azureMachineConfigModel.AzureMasterImage.IsNull()) &&
					(!azureMachineConfigModel.AzurePreparedImage.IsUnknown() && azureMachineConfigModel.AzurePreparedImage.IsNull()) {
					resp.Diagnostics.AddAttributeError(
						path.Root("azure_machine_config"),
						"Missing Attribute Configuration",
						fmt.Sprintf("Expected either `azure_master_image` or `prepared_image` to be configured when provisioning_type is %s.", provisioningTypeMcs),
					)
				}

				// Validate Azure PVS Configuration
				if !azureMachineConfigModel.AzurePvsConfiguration.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("azure_pvs_config"),
						"Incorrect Attribute Configuration",
						fmt.Sprintf("azure_pvs_config is not supported when provisioning_type is %s.", provisioningTypeMcs),
					)
				}

				if !azureMachineConfigModel.WritebackCache.IsNull() {
					// Validate Writeback Cache
					azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, &resp.Diagnostics, azureMachineConfigModel.WritebackCache)

					if !azureWbcModel.PersistWBC.IsUnknown() && azureWbcModel.PersistWBC.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("persist_wbc"),
							"Missing Attribute Configuration",
							fmt.Sprintf("persist_wbc for writeback_cache under azure_machine_config must be set when provisioning_type is %s.", provisioningTypeMcs),
						)
					}

					if !azureWbcModel.StorageCostSaving.IsUnknown() && azureWbcModel.StorageCostSaving.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("storage_cost_saving"),
							"Incorrect Attribute Configuration",
							fmt.Sprintf("storage_cost_saving for writeback_cache under azure_machine_config must be set when provisioning_type is %s.", provisioningTypeMcs),
						)
					}

					if !azureWbcModel.WriteBackCacheMemorySizeMB.IsUnknown() && azureWbcModel.WriteBackCacheMemorySizeMB.IsNull() {
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

					if !azureWbcModel.StorageCostSaving.IsUnknown() && !azureWbcModel.StorageCostSaving.IsNull() {
						if azureWbcModel.StorageCostSaving.ValueBool() &&
							(strings.EqualFold(azureWbcModel.WBCDiskStorageType.ValueString(), util.StandardLRS) ||
								strings.EqualFold(azureMachineConfigModel.StorageType.ValueString(), util.StandardLRS)) {
							resp.Diagnostics.AddAttributeError(
								path.Root("storage_cost_saving"),
								"Incorrect Attribute Configuration",
								fmt.Sprintf("storage_cost_saving cannot be set to `true` when storage_type is set to `%s` or wbc_disk_storage_type is set to `%s`.", util.StandardLRS, util.StandardLRS),
							)
						}
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

					// Exactly one of UseAzureComputeGallery or PreparedImage should be configured when the storage_type is set to AzureEphemeralOSDisk
					if !azureMachineConfigModel.UseAzureComputeGallery.IsUnknown() && azureMachineConfigModel.UseAzureComputeGallery.IsNull() &&
						!azureMachineConfigModel.AzurePreparedImage.IsUnknown() && azureMachineConfigModel.AzurePreparedImage.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("storage_type"),
							"Missing Attribute Configuration",
							fmt.Sprintf("exactly one of use_azure_compute_gallery or prepared_image should be set when storage_type is %s.", util.AzureEphemeralOSDisk),
						)
					}
				}

				if !azureMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, azureMachineConfigModel.ImageUpdateRebootOptions)
					rebootOptions.ValidateConfig(&resp.Diagnostics)
				}

				if !azureMachineConfigModel.SecondaryVmSizes.IsNull() && !azureMachineConfigModel.MachineProfile.IsUnknown() && azureMachineConfigModel.MachineProfile.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("secondary_vm_sizes"),
						"Incorrect Attribute Configuration",
						"secondary_vm_sizes cannot be configured when machine_profile is not set.",
					)
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

			if !provSchemeModel.AmazonWorkspacesCoreMachineConfig.IsNull() {
				amazonWorkspacesCoreMachineConfigModel := util.ObjectValueToTypedObject[AmazonWorkspacesCoreMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AmazonWorkspacesCoreMachineConfig)
				if !amazonWorkspacesCoreMachineConfigModel.ImageUpdateRebootOptions.IsNull() {
					// Validate Image Update Reboot Options
					rebootOptions := util.ObjectValueToTypedObject[ImageUpdateRebootOptionsModel](ctx, &resp.Diagnostics, amazonWorkspacesCoreMachineConfigModel.ImageUpdateRebootOptions)
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

			if !provSchemeModel.MachineDomainIdentity.IsNull() && provSchemeModel.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY) {
				machineDomainIdentityModel := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeModel.MachineDomainIdentity)
				if machineDomainIdentityModel.Domain.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("domain"),
						"Missing Attribute Configuration",
						"Expected domain to be configured when identity_type is Active Directory.",
					)
					return
				}
			}

			if !provSchemeModel.MachineAccountCreationRules.IsNull() {
				machineAccountCreationRulesModel := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, provSchemeModel.MachineAccountCreationRules)
				if !machineAccountCreationRulesModel.StartsWith.IsNull() {
					startsWith := machineAccountCreationRulesModel.StartsWith.ValueString()
					namingScheme := machineAccountCreationRulesModel.NamingScheme.ValueString()
					namingSchemeType := machineAccountCreationRulesModel.NamingSchemeType.ValueString()
					wildCardCount := strings.Count(namingScheme, "#")
					if len(startsWith) < wildCardCount {
						resp.Diagnostics.AddAttributeError(
							path.Root("starts_with"),
							"Incorrect Attribute Configuration",
							"Characters in starts_with must be equal to or greater than the number of wildcard (#) characters in naming_scheme.",
						)

						return
					}

					regexToMatch := ""
					errMsg := ""
					if strings.EqualFold(namingSchemeType, string(citrixorchestration.NAMINGSCHEMETYPE_ALPHABETIC)) {
						regexToMatch = util.UpperCaseRegex
						errMsg = "starts_with must contain only uppercase letters without any spaces."
					} else if strings.EqualFold(namingSchemeType, string(citrixorchestration.NAMINGSCHEMETYPE_NUMERIC)) {
						regexToMatch = util.NumbersRegex
						errMsg = "starts_with must contain only numbers without any spaces."
					}

					if match, _ := regexp.MatchString(regexToMatch, startsWith); !match {
						resp.Diagnostics.AddAttributeError(
							path.Root("starts_with"),
							"Incorrect Attribute Configuration",
							errMsg,
						)

						return
					}
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
		err := validateMcsPvsMachineCatalogDeleteOptions(&resp.Diagnostics, data)
		if err != nil {
			return
		}

		if !data.ProvisioningScheme.IsUnknown() && data.ProvisioningScheme.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("provisioning_scheme"),
				"Missing Attribute Configuration",
				fmt.Sprintf("Expected provisioning_scheme to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
			)
		} else {
			// Validate Provisioning Scheme
			provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, data.ProvisioningScheme)
			if !provSchemeModel.AzureMachineConfig.IsUnknown() && provSchemeModel.AzureMachineConfig.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("azure_machine_config"),
					"Missing Attribute Configuration",
					fmt.Sprintf("PVS Catalogs are currently only supported for Azure environment. Expected azure_machine_config to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			azureMachineConfigModel := util.ObjectValueToTypedObject[AzureMachineConfigModel](ctx, &resp.Diagnostics, provSchemeModel.AzureMachineConfig)

			if !azureMachineConfigModel.AzurePvsConfiguration.IsUnknown() && azureMachineConfigModel.AzurePvsConfiguration.IsNull() {
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

			if !azureMachineConfigModel.MachineProfile.IsUnknown() && azureMachineConfigModel.MachineProfile.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("machine_profile"),
					"Missing Attribute Configuration",
					fmt.Sprintf("Expected machine_profile to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			}

			if !azureMachineConfigModel.WritebackCache.IsUnknown() && azureMachineConfigModel.WritebackCache.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("writeback_cache"),
					"Missing Attribute Configuration",
					fmt.Sprintf("Expected writeback_cache to be configured when value of provisioning_type is %s.", provisioningTypePvsStreaming),
				)
			} else {

				azureWbcModel := util.ObjectValueToTypedObject[AzureWritebackCacheModel](ctx, &resp.Diagnostics, azureMachineConfigModel.WritebackCache)
				if !azureWbcModel.PersistWBC.IsUnknown() && (azureWbcModel.PersistWBC.IsNull() || !azureWbcModel.PersistWBC.ValueBool()) {
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

	} else if data.ProvisioningType.ValueString() == provisioningTypeManual && data.RemotePcPowerManagementHypervisor.IsNull() {
		// Manual provisioning type
		if !data.IsPowerManaged.IsUnknown() && data.IsPowerManaged.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_power_managed"),
				"Missing Attribute Configuration",
				fmt.Sprintf("expected is_power_managed to be defined when provisioning_type is %s.", provisioningTypeManual),
			)
		}

		if !data.IsRemotePc.IsUnknown() && data.IsRemotePc.IsNull() {
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
					if !machineAccount.Hypervisor.IsUnknown() && machineAccount.Hypervisor.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("machine_accounts"),
							"Missing Attribute Configuration",
							"Expected hypervisor to be configured when machines are power managed.",
						)
					}

					machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, &resp.Diagnostics, machineAccount.Machines)
					for _, machine := range machines {
						if !machine.MachineName.IsUnknown() && machine.MachineName.IsNull() {
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
	} else if data.ProvisioningType.ValueString() == provisioningTypeManual && !data.RemotePcPowerManagementHypervisor.IsNull() {
		if data.IsRemotePc.IsNull() || !data.IsRemotePc.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_remote_pc"),
				"Incorrect Attribute Configuration",
				"Wake on LAN hypervisor cannot be configured when is_remote_pc is not set to true.",
			)
		}

		if data.IsPowerManaged.IsNull() || !data.IsPowerManaged.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_power_managed"),
				"Incorrect Attribute Configuration",
				"Wake on LAN hypervisor cannot be configured when is_power_managed is not set to true.",
			)
		}

		if !data.MachineAccounts.IsNull() {
			machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, &resp.Diagnostics, data.MachineAccounts)
			for _, machineAccount := range machineAccounts {
				if !machineAccount.Hypervisor.IsUnknown() && !machineAccount.Hypervisor.IsNull() {
					resp.Diagnostics.AddAttributeError(
						path.Root("machine_accounts"),
						"Incorrect Attribute Configuration",
						"Hypervisor cannot be configured when machines are power managed by Remote PC Wake on LAN connection.",
					)
				}

				machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, &resp.Diagnostics, machineAccount.Machines)
				for _, machine := range machines {
					if !machine.MachineName.IsUnknown() && machine.MachineName.IsNull() {
						resp.Diagnostics.AddAttributeError(
							path.Root("machine_accounts"),
							"Missing Attribute Configuration",
							"Expected machine_name to be configured when machines are power managed.",
						)
					}
				}
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

	deleteVirtualMachines := data.DeleteVirtualMachines.ValueBool()
	if data.DeleteVirtualMachines.IsNull() && data.ProvisioningType.ValueString() != string(citrixorchestration.PROVISIONINGTYPE_MANUAL) {
		deleteVirtualMachines = true
	}

	if deleteVirtualMachines && data.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_MANUAL) {
		resp.Diagnostics.AddAttributeError(
			path.Root("delete_virtual_machines"),
			"Incorrect Attribute Configuration",
			fmt.Sprintf("`delete_virtual_machines` cannot be set to `true` when `provisioning_type` is set to `%s`.", provisioningTypeManual),
		)
		return
	}

	deleteMachineAccounts := data.DeleteMachineAccounts.ValueString()
	if data.DeleteMachineAccounts.IsNull() {
		deleteMachineAccounts = string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE)
	}
	if deleteMachineAccounts != string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE) && data.ProvisioningType.ValueString() == string(citrixorchestration.PROVISIONINGTYPE_MANUAL) {
		resp.Diagnostics.AddAttributeError(
			path.Root("delete_machine_accounts"),
			"Incorrect Attribute Configuration",
			fmt.Sprintf("`delete_machine_accounts` can only be set to `%s` when `provisioning_type` is set to `%s`.", string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE), provisioningTypeManual),
		)
		return
	}
}

func (r *machineCatalogResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if req.Plan.Raw.IsNull() {
		return
	}

	var plan MachineCatalogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DeleteVirtualMachines.IsUnknown() && !plan.DeleteVirtualMachines.IsNull() &&
		!plan.DeleteMachineAccounts.IsUnknown() && !plan.DeleteMachineAccounts.IsNull() {
		if !plan.DeleteVirtualMachines.ValueBool() && plan.DeleteMachineAccounts.ValueString() != string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE) {
			resp.Diagnostics.AddError(
				"Error validating `delete_machine_accounts`",
				fmt.Sprintf("`delete_machine_accounts` can only be set to `%s` when `delete_virtual_machines` is set to `false`.", string(citrixorchestration.MACHINEACCOUNTDELETEOPTION_NONE)),
			)
			return
		}
	}

	if !plan.ProvisioningScheme.IsUnknown() && !plan.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)

		if provSchemePlan.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_WORKGROUP) {
			util.CheckProductVersion(r.client, &resp.Diagnostics, 118, 118, 7, 43, "Unsupported Machine Catalog Configuration", "Identity type Workgroup")
			util.CheckFunctionalLevelValues(r.client, &resp.Diagnostics, plan.MinimumFunctionalLevel.String(), "Unsupported Machine Catalog Configuration", "Identity type Workgroup")
		}

		if !provSchemePlan.MachineADAccounts.IsUnknown() && !provSchemePlan.MachineADAccounts.IsNull() {
			machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, provSchemePlan.MachineADAccounts)

			if len(machineAccountsInPlan) > 0 {
				machineAccountsNamesInPlan := []string{}
				for _, machineAccountInPlan := range machineAccountsInPlan {
					if !machineAccountInPlan.ResetPassword.ValueBool() && machineAccountInPlan.Password.ValueString() == "" {
						resp.Diagnostics.AddError(
							"Error validating machine accounts for Machine Catalog "+plan.Name.ValueString(),
							"Password cannot be empty when `reset_password` is set to `false`.",
						)
						return
					}
					machineAccountsNamesInPlan = append(machineAccountsNamesInPlan, strings.TrimSuffix(machineAccountInPlan.ADAccountName.ValueString(), "$"))
				}
				_, _, err := getMachinesUsingIdentity(ctx, r.client, machineAccountsNamesInPlan)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error validating machine accounts for Machine Catalog "+plan.Name.ValueString(),
						"Error Message: "+err.Error(),
					)
					return
				}
			}

			if req.State.Raw.IsNull() {
				// Ensure machine account count is greater or equals to the number of machines in the catalog during create
				if len(machineAccountsInPlan) < int(provSchemePlan.NumTotalMachines.ValueInt64()) {
					resp.Diagnostics.AddError(
						"Error creating Machine Catalog",
						"Number of machine accounts must be greater than or equal to the number of machines during machine catalog creation.",
					)
				}
				return
			}
		}
	}

	if !plan.PersistUserChanges.IsUnknown() && plan.PersistUserChanges.IsNull() &&
		!plan.ProvisioningType.IsUnknown() && !plan.ProvisioningType.IsNull() &&
		!plan.SessionSupport.IsUnknown() && !plan.SessionSupport.IsNull() &&
		!plan.AllocationType.IsUnknown() && !plan.AllocationType.IsNull() {

		provisioningType := plan.ProvisioningType.ValueString()
		persistChanges := citrixorchestration.PERSISTCHANGES_DISCARD
		if provisioningType == string(citrixorchestration.PROVISIONINGTYPE_MANUAL) ||
			(provisioningType != string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING) &&
				plan.SessionSupport.ValueString() == string(citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION) &&
				plan.AllocationType.ValueString() == string(citrixorchestration.ALLOCATIONTYPE_STATIC)) {
			persistChanges = citrixorchestration.PERSISTCHANGES_ON_LOCAL
		}
		paths, diags := resp.Plan.PathMatches(ctx, path.MatchRoot("persist_user_changes"))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Plan.SetAttribute(ctx, paths[0], string(persistChanges))
	}

	catalogNameExists := checkIfCatalogNameExists(ctx, r.client, plan.Name.ValueString())

	if req.State.Raw.IsNull() {
		return
	}

	var state MachineCatalogResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if catalogNameExists && !strings.EqualFold(plan.Name.ValueString(), state.Name.ValueString()) {
		// Validate machine catalog name uniqueness for update if the name is changed
		resp.Diagnostics.AddError(
			"Machine Catalog Name Already Exists",
			fmt.Sprintf("A Machine Catalog with the name '%s' already exists. Please choose a different name.", plan.Name.ValueString()),
		)

		return
	}

	if !plan.ProvisioningScheme.IsNull() {
		provSchemePlan := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
		provSchemeState := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, state.ProvisioningScheme)
		machineAccountsInPlan := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, provSchemePlan.MachineADAccounts)

		if len(machineAccountsInPlan) < int(provSchemePlan.NumTotalMachines.ValueInt64()) && provSchemePlan.MachineAccountCreationRules.IsNull() {
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog "+state.Name.ValueString(),
				"`machine_account_creation_rules` must be specified when machine accounts associated with catalog is less than the desired number of machines.",
			)
			return
		}
		if !provSchemeState.MachineADAccounts.IsNull() {
			inUseMachineAccountsToBeDeleted := []string{}
			machineAccountsInState := util.ObjectListToTypedArray[MachineADAccountModel](ctx, &resp.Diagnostics, provSchemeState.MachineADAccounts)
			// Ensure machine accounts with state InUse is not deleted.
			for _, machineAccountInState := range machineAccountsInState {
				if machineAccountInState.State.ValueString() == string(citrixorchestration.PROVISIONINGSCHEMEMACHINEACCOUNTSTATE_IN_USE) &&
					!slices.ContainsFunc(machineAccountsInPlan, func(machineAccountInPlan MachineADAccountModel) bool {
						return strings.EqualFold(machineAccountInState.ADAccountName.ValueString(), machineAccountInPlan.ADAccountName.ValueString())
					}) {
					inUseMachineAccountsToBeDeleted = append(inUseMachineAccountsToBeDeleted, machineAccountInState.ADAccountName.ValueString())
				}
			}
			if len(inUseMachineAccountsToBeDeleted) > 0 {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+state.Name.ValueString(),
					fmt.Sprintf("Machine account(s) [ %s ] with state `InUse` cannot be deleted ", strings.Join(inUseMachineAccountsToBeDeleted, ", ")),
				)
				return
			}
		}

		if !provSchemePlan.MachineAccountCreationRules.IsNull() {
			machineAccountCreationRulesPlan := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, provSchemePlan.MachineAccountCreationRules)
			machineAccountCreationRulesState := util.ObjectValueToTypedObject[MachineAccountCreationRulesModel](ctx, &resp.Diagnostics, provSchemeState.MachineAccountCreationRules)
			machineDomainIdentityPlan := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemePlan.MachineDomainIdentity)
			machineDomainIdentityState := util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, &resp.Diagnostics, provSchemeState.MachineDomainIdentity)

			if machineDomainIdentityPlan.Ou != machineDomainIdentityState.Ou &&
				provSchemePlan.NumTotalMachines.ValueInt64() <= provSchemeState.NumTotalMachines.ValueInt64() {

				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+state.Name.ValueString(),
					"Machine Catalog OU can only be updated when adding machines.",
				)
				return
			}

			if machineAccountCreationRulesPlan != machineAccountCreationRulesState &&
				provSchemePlan.NumTotalMachines.ValueInt64() == provSchemeState.NumTotalMachines.ValueInt64() {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+state.Name.ValueString(),
					"machine_account_creation_rules can only be updated when adding machines.",
				)
				return
			}
		}
		// Validate metadata
		if !provSchemePlan.Metadata.IsUnknown() && !provSchemeState.Metadata.IsUnknown() {
			requireReplaceForMetadataChange := false
			if r.client.AuthConfig.OnPremises {
				majorVersion, minorVersion, err := util.GetProductMajorAndMinorVersion(r.client)
				if err != nil {
					requireReplaceForMetadataChange = true
				}
				if majorVersion < 7 || (majorVersion == 7 && minorVersion < 45) {
					requireReplaceForMetadataChange = true
				}
			} else {
				if r.client.ClientConfig.OrchestrationApiVersion < 123 {
					requireReplaceForMetadataChange = true
				}
			}
			if !provSchemePlan.Metadata.Equal(provSchemeState.Metadata) && requireReplaceForMetadataChange {
				resp.RequiresReplace = append(resp.RequiresReplace, path.Root("provisioning_scheme").AtName("metadata"))
			}
		}

		if provSchemePlan.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_AZURE_AD) {
			if provSchemePlan.MachineDomainIdentity.IsNull() && !provSchemeState.MachineDomainIdentity.IsNull() {
				resp.Diagnostics.AddError(
					"Error updating Machine Catalog "+state.Name.ValueString(),
					"Service Account cannot be removed when identity type is Azure AD.",
				)
				return
			}
		}
	}
}
