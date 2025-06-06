// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getRemotePcEnrollmentScopes(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, plan MachineCatalogResourceModel, includeMachines bool) ([]citrixorchestration.RemotePCEnrollmentScopeRequestModel, error) {
	remotePCEnrollmentScopes := []citrixorchestration.RemotePCEnrollmentScopeRequestModel{}
	if !plan.RemotePcOus.IsNull() {
		remotePcOus := util.ObjectListToTypedArray[RemotePcOuModel](ctx, diagnostics, plan.RemotePcOus)
		for _, ou := range remotePcOus {

			getIdentityContainer := client.ApiClient.IdentityAPIsDAAS.IdentityGetContainer(ctx, ou.OUName.ValueString())
			identityContainer, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.IdentityContainerResponseModel](getIdentityContainer, client)

			if err != nil {
				diagnostics.AddError(
					"An error occurred while fetching OU "+ou.OUName.ValueString(),
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)

				return remotePCEnrollmentScopes, err
			}

			var remotePCEnrollmentScope citrixorchestration.RemotePCEnrollmentScopeRequestModel
			remotePCEnrollmentScope.SetIncludeSubfolders(ou.IncludeSubFolders.ValueBool())
			remotePCEnrollmentScope.SetOU(identityContainer.GetDistinguishedName())
			remotePCEnrollmentScope.SetIsOrganizationalUnit(true)
			remotePCEnrollmentScopes = append(remotePCEnrollmentScopes, remotePCEnrollmentScope)
		}
	}

	if includeMachines && !plan.MachineAccounts.IsNull() {
		machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, plan.MachineAccounts)
		for _, machineAccount := range machineAccounts {
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)

			// verify machine accounts using Identity API
			httpResp, err := verifyMachinesUsingIdentity(ctx, client, machines)
			if err != nil {
				diagnostics.AddError(
					"An error occurred while fetching machines using identity",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return remotePCEnrollmentScopes, err
			}

			for _, machine := range machines {
				var remotePCEnrollmentScope citrixorchestration.RemotePCEnrollmentScopeRequestModel
				remotePCEnrollmentScope.SetIncludeSubfolders(false)
				remotePCEnrollmentScope.SetOU(machine.MachineAccount.ValueString())
				remotePCEnrollmentScope.SetIsOrganizationalUnit(false)
				remotePCEnrollmentScopes = append(remotePCEnrollmentScopes, remotePCEnrollmentScope)
			}
		}
	}

	return remotePCEnrollmentScopes, nil
}

func (r MachineCatalogResourceModel) updateCatalogWithRemotePcConfig(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel) MachineCatalogResourceModel {
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		r.IsRemotePc = types.BoolValue(catalog.GetIsRemotePC())
	} else {
		r.IsRemotePc = types.BoolNull()
	}
	if r.IsRemotePc.ValueBool() && r.IsPowerManaged.ValueBool() {
		// If the catalog is Remote PC and power management is enabled, then its a Remote PC Wake-On-LAN catalog.
		hypervisor := catalog.GetHypervisorConnection()
		hypervisorName := hypervisor.GetName()
		hyp, err := util.GetHypervisor(ctx, client, diagnostics, hypervisorName)
		if err != nil {
			return r
		}
		r.RemotePcPowerManagementHypervisor = types.StringValue(hyp.GetId())
	}

	rpcOUs := util.RefreshListValueProperties[RemotePcOuModel, citrixorchestration.RemotePCEnrollmentScopeResponseModel](ctx, diagnostics, r.RemotePcOus, catalog.GetRemotePCEnrollmentScopes(), util.GetOrchestrationRemotePcOuKey)
	r.RemotePcOus = rpcOUs
	return r
}
