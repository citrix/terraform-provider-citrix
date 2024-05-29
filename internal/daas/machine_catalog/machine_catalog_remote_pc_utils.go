// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getRemotePcEnrollmentScopes(ctx context.Context, diagnostics *diag.Diagnostics, plan MachineCatalogResourceModel, includeMachines bool) []citrixorchestration.RemotePCEnrollmentScopeRequestModel {
	remotePCEnrollmentScopes := []citrixorchestration.RemotePCEnrollmentScopeRequestModel{}
	if !plan.RemotePcOus.IsNull() {
		remotePcOus := util.ObjectListToTypedArray[RemotePcOuModel](ctx, diagnostics, plan.RemotePcOus)
		for _, ou := range remotePcOus {
			var remotePCEnrollmentScope citrixorchestration.RemotePCEnrollmentScopeRequestModel
			remotePCEnrollmentScope.SetIncludeSubfolders(ou.IncludeSubFolders.ValueBool())
			remotePCEnrollmentScope.SetOU(ou.OUName.ValueString())
			remotePCEnrollmentScope.SetIsOrganizationalUnit(true)
			remotePCEnrollmentScopes = append(remotePCEnrollmentScopes, remotePCEnrollmentScope)
		}
	}

	if includeMachines && !plan.MachineAccounts.IsNull() {
		machineAccounts := util.ObjectListToTypedArray[MachineAccountsModel](ctx, diagnostics, plan.MachineAccounts)
		for _, machineAccount := range machineAccounts {
			machines := util.ObjectListToTypedArray[MachineCatalogMachineModel](ctx, diagnostics, machineAccount.Machines)
			for _, machine := range machines {
				var remotePCEnrollmentScope citrixorchestration.RemotePCEnrollmentScopeRequestModel
				remotePCEnrollmentScope.SetIncludeSubfolders(false)
				remotePCEnrollmentScope.SetOU(machine.MachineAccount.ValueString())
				remotePCEnrollmentScope.SetIsOrganizationalUnit(false)
				remotePCEnrollmentScopes = append(remotePCEnrollmentScopes, remotePCEnrollmentScope)
			}
		}
	}

	return remotePCEnrollmentScopes
}

func (r MachineCatalogResourceModel) updateCatalogWithRemotePcConfig(ctx context.Context, diagnostics *diag.Diagnostics, catalog *citrixorchestration.MachineCatalogDetailResponseModel) MachineCatalogResourceModel {
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL || !r.IsRemotePc.IsNull() {
		r.IsRemotePc = types.BoolValue(catalog.GetIsRemotePC())
	}
	rpcOUs := util.RefreshListValueProperties[RemotePcOuModel, citrixorchestration.RemotePCEnrollmentScopeResponseModel](ctx, diagnostics, r.RemotePcOus, "OUName", catalog.GetRemotePCEnrollmentScopes(), "OU", "RefreshListItem")
	r.RemotePcOus = rpcOUs
	return r
}
