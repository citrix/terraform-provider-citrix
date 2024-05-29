// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getRequestModelForCreateMachineCatalog(plan MachineCatalogResourceModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, isOnPremises bool) (*citrixorchestration.CreateMachineCatalogRequestModel, error) {
	provisioningType, err := citrixorchestration.NewProvisioningTypeFromValue(plan.ProvisioningType.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported provisioning type.",
		)

		return nil, err
	}

	var machinesRequest []citrixorchestration.AddMachineToMachineCatalogRequestModel
	var body citrixorchestration.CreateMachineCatalogRequestModel

	isRemotePcCatalog := plan.IsRemotePc.ValueBool()

	// Generate API request body from plan
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetProvisioningType(*provisioningType) // Only support MCS and Manual. Block other types
	allocationType, err := citrixorchestration.NewAllocationTypeFromValue(plan.AllocationType.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported allocation type.",
		)
		return nil, err
	}
	body.SetAllocationType(*allocationType)
	sessionSupport, err := citrixorchestration.NewSessionSupportFromValue(plan.SessionSupport.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported session support.",
		)
		return nil, err
	}
	body.SetSessionSupport(*sessionSupport)
	persistChanges := citrixorchestration.PERSISTCHANGES_DISCARD
	if *sessionSupport == citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION && *allocationType == citrixorchestration.ALLOCATIONTYPE_STATIC {
		persistChanges = citrixorchestration.PERSISTCHANGES_ON_LOCAL
	}
	body.SetPersistUserChanges(persistChanges)
	body.SetZone(plan.Zone.ValueString())
	if !plan.VdaUpgradeType.IsNull() {
		body.SetVdaUpgradeType(citrixorchestration.VdaUpgradeType(plan.VdaUpgradeType.ValueString()))
	} else {
		body.SetVdaUpgradeType(citrixorchestration.VDAUPGRADETYPE_NOT_SET)
	}

	functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Error creating Machine Catalog",
			fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
		)
		return nil, err
	}
	body.SetMinimumFunctionalLevel(*functionalLevel)

	if *provisioningType == citrixorchestration.PROVISIONINGTYPE_MCS {
		provisioningScheme, err := getProvSchemeForMcsCatalog(plan, ctx, client, diagnostics, isOnPremises)
		if err != nil {
			return nil, err
		}
		body.SetProvisioningScheme(*provisioningScheme)
		return &body, nil
	}

	// Manual type catalogs
	machineType := citrixorchestration.MACHINETYPE_VIRTUAL
	if !plan.IsPowerManaged.ValueBool() {
		machineType = citrixorchestration.MACHINETYPE_PHYSICAL
	}

	body.SetMachineType(machineType)
	body.SetIsRemotePC(plan.IsRemotePc.ValueBool())

	if isRemotePcCatalog {
		remotePCEnrollmentScopes := getRemotePcEnrollmentScopes(ctx, diagnostics, plan, true)
		body.SetRemotePCEnrollmentScopes(remotePCEnrollmentScopes)
	} else {
		machinesRequest, err = getMachinesForManualCatalogs(ctx, diagnostics, client, util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, plan.MachineAccounts))
		if err != nil {
			diagnostics.AddError(
				"Error creating Machine Catalog",
				fmt.Sprintf("Failed to resolve machines, error: %s", err.Error()),
			)
			return nil, err
		}
		body.SetMachines(machinesRequest)
	}

	return &body, nil
}

func getRequestModelForUpdateMachineCatalog(plan MachineCatalogResourceModel, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, isOnPremises bool) (*citrixorchestration.UpdateMachineCatalogRequestModel, error) {
	// Generate API request body from plan
	var body citrixorchestration.UpdateMachineCatalogRequestModel
	body.SetName(plan.Name.ValueString())
	body.SetDescription(plan.Description.ValueString())
	body.SetZone(plan.Zone.ValueString())
	if !plan.VdaUpgradeType.IsNull() {
		body.SetVdaUpgradeType(citrixorchestration.VdaUpgradeType(plan.VdaUpgradeType.ValueString()))
	} else {
		body.SetVdaUpgradeType(citrixorchestration.VDAUPGRADETYPE_NOT_SET)
	}

	functionalLevel, err := citrixorchestration.NewFunctionalLevelFromValue(plan.MinimumFunctionalLevel.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog",
			fmt.Sprintf("Unsupported minimum functional level %s.", plan.MinimumFunctionalLevel.ValueString()),
		)
		return nil, err
	}
	body.SetMinimumFunctionalLevel(*functionalLevel)

	provisioningType, err := citrixorchestration.NewProvisioningTypeFromValue(plan.ProvisioningType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Machine Catalog",
			"Unsupported provisioning type.",
		)

		return nil, err
	}

	if *provisioningType == citrixorchestration.PROVISIONINGTYPE_MANUAL {

		if plan.IsRemotePc.ValueBool() {
			remotePCEnrollmentScopes := getRemotePcEnrollmentScopes(ctx, &resp.Diagnostics, plan, false)
			body.SetRemotePCEnrollmentScopes(remotePCEnrollmentScopes)
		}

		return &body, nil
	}

	provSchemeModel := util.ObjectValueToTypedObject[ProvisioningSchemeModel](ctx, &resp.Diagnostics, plan.ProvisioningScheme)
	if !checkIfProvSchemeIsCloudOnly(provSchemeModel, isOnPremises, &resp.Diagnostics, ctx) {
		return nil, fmt.Errorf("identity type %s is not supported in OnPremises environment. ", provSchemeModel.IdentityType.ValueString())
	}

	body, err = setProvSchemePropertiesForUpdateCatalog(provSchemeModel, body, ctx, client, &resp.Diagnostics)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func checkIfProvSchemeIsCloudOnly(provisoningScheme ProvisioningSchemeModel, isOnPremises bool, diagnostics *diag.Diagnostics, ctx context.Context) bool {
	if provisoningScheme.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_AZURE_AD) ||
		provisoningScheme.IdentityType.ValueString() == string(citrixorchestration.IDENTITYTYPE_WORKGROUP) {
		if isOnPremises {
			diagnostics.AddAttributeError(
				path.Root("identity_type"),
				"Unsupported Machine Catalog Configuration",
				fmt.Sprintf("Identity type %s is not supported in OnPremises environment. ", string(provisoningScheme.IdentityType.ValueString())),
			)

			return false
		}
	}
	return true
}

func generateBatchApiHeaders(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, provisioningSchemePlan ProvisioningSchemeModel, generateCredentialHeader bool) ([]citrixorchestration.NameValueStringPairModel, *http.Response, error) {
	headers := []citrixorchestration.NameValueStringPairModel{}

	cwsAuthToken, httpResp, err := client.SignIn()
	var token string
	if err != nil {
		return headers, httpResp, err
	}

	if cwsAuthToken != "" {
		token = strings.Split(cwsAuthToken, "=")[1]
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("Authorization")
		header.SetValue("Bearer " + token)
		headers = append(headers, header)
	}

	if generateCredentialHeader && !provisioningSchemePlan.MachineDomainIdentity.IsNull() {
		adminCredentialHeader := generateAdminCredentialHeader(util.ObjectValueToTypedObject[MachineDomainIdentityModel](ctx, diagnostics, provisioningSchemePlan.MachineDomainIdentity))
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("X-AdminCredential")
		header.SetValue(adminCredentialHeader)
		headers = append(headers, header)
	}

	return headers, httpResp, err
}

func readMachineCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, machineCatalogId string) (*citrixorchestration.MachineCatalogDetailResponseModel, *http.Response, error) {
	getMachineCatalogRequest := client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalog(ctx, machineCatalogId).Fields("Id,Name,Description,ProvisioningType,Zone,AllocationType,SessionSupport,TotalCount,HypervisorConnection,ProvisioningScheme,RemotePCEnrollmentScopes,IsPowerManaged,MinimumFunctionalLevel,IsRemotePC")
	catalog, httpResp, err := util.ReadResource[*citrixorchestration.MachineCatalogDetailResponseModel](getMachineCatalogRequest, ctx, client, resp, "Machine Catalog", machineCatalogId)

	return catalog, httpResp, err
}

func deleteMachinesFromCatalog(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.UpdateResponse, provisioningSchemePlan ProvisioningSchemeModel, machinesToDelete []citrixorchestration.MachineResponseModel, catalogNameOrId string, isMcsCatalog bool) error {
	batchApiHeaders, httpResp, err := generateBatchApiHeaders(ctx, &resp.Diagnostics, client, provisioningSchemePlan, false)
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogNameOrId,
			"TransactionId: "+txId+
				"\nCould not put machine(s) into maintenance mode before deleting them, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}

	for index, machineToDelete := range machinesToDelete {
		if machineToDelete.DeliveryGroup == nil {
			// if machine has no delivery group, there is no need to put it in maintenance mode
			continue
		}

		isMachineInMaintenanceMode := machineToDelete.GetInMaintenanceMode()

		if !isMachineInMaintenanceMode {
			// machine is not in maintenance mode. Put machine in maintenance mode first before deleting
			var updateMachineModel citrixorchestration.UpdateMachineRequestModel
			updateMachineModel.SetInMaintenanceMode(true)
			updateMachineStringBody, err := util.ConvertToString(updateMachineModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error removing Machine(s) from Machine Catalog "+catalogNameOrId,
					"An unexpected error occurred: "+err.Error(),
				)
				return err
			}
			relativeUrl := fmt.Sprintf("/Machines/%s?async=true", machineToDelete.GetId())

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(strconv.Itoa(index))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetBody(updateMachineStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}
	}

	if len(batchRequestItems) > 0 {
		// If there are any machines that need to be put in maintenance mode
		var batchRequestModel citrixorchestration.BatchRequestModel
		batchRequestModel.SetItems(batchRequestItems)
		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting machine(s) from Machine Catalog "+catalogNameOrId,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return err
		}

		if successfulJobs < len(batchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while putting machine(s) into maintenance mode before deleting them. %d of %d machines were put in the maintenance mode.", successfulJobs, len(batchRequestItems))
			err = fmt.Errorf(errMsg)
			resp.Diagnostics.AddError(
				"Error updating Machine Catalog "+catalogNameOrId,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return err
		}
	}

	batchApiHeaders, httpResp, err = generateBatchApiHeaders(ctx, &resp.Diagnostics, client, provisioningSchemePlan, isMcsCatalog)
	txId = citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogNameOrId,
			"TransactionId: "+txId+
				"\nCould not delete machine(s) from machine catalog, unexpected error: "+util.ReadClientError(err),
		)
		return err
	}

	deleteAccountOpion := "Leave"
	if isMcsCatalog {
		deleteAccountOpion = "Delete"
	}
	batchRequestItems = []citrixorchestration.BatchRequestItemModel{}
	for index, machineToDelete := range machinesToDelete {
		var batchRequestItem citrixorchestration.BatchRequestItemModel
		relativeUrl := fmt.Sprintf("/Machines/%s?deleteVm=%t&purgeDBOnly=false&deleteAccount=%s&async=true", machineToDelete.GetId(), isMcsCatalog, deleteAccountOpion)
		batchRequestItem.SetReference(strconv.Itoa(index))
		batchRequestItem.SetMethod(http.MethodDelete)
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	batchRequestModel := citrixorchestration.BatchRequestModel{}
	batchRequestModel.SetItems(batchRequestItems)
	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting machine(s) from Machine Catalog "+catalogNameOrId,
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
		return err
	}

	if successfulJobs < len(machinesToDelete) {
		errMsg := fmt.Sprintf("An error occurred while deleting machine(s) from Machine Catalog. %d of %d machines were deleted from the Machine Catalog.", successfulJobs, len(batchRequestItems))
		err = fmt.Errorf(errMsg)
		resp.Diagnostics.AddError(
			"Error updating Machine Catalog "+catalogNameOrId,
			"TransactionId: "+txId+
				"\n"+errMsg,
		)

		return err
	}

	return nil
}

func allocationTypeEnumToString(conn citrixorchestration.AllocationType) string {
	switch conn {
	case citrixorchestration.ALLOCATIONTYPE_UNKNOWN:
		return "Unknown"
	case citrixorchestration.ALLOCATIONTYPE_RANDOM:
		return "Random"
	case citrixorchestration.ALLOCATIONTYPE_STATIC:
		return "Static"
	default:
		return ""
	}
}

func (scope RemotePcOuModel) RefreshListItem(_ context.Context, _ *diag.Diagnostics, remote citrixorchestration.RemotePCEnrollmentScopeResponseModel) RemotePcOuModel {
	scope.OUName = types.StringValue(remote.GetOU())
	scope.IncludeSubFolders = types.BoolValue(remote.GetIncludeSubfolders())

	return scope
}
