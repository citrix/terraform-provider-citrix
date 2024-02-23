// Copyright © 2023. Citrix Systems, Inc.

package machine_catalog

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getRemotePcEnrollmentScopes(plan MachineCatalogResourceModel, includeMachines bool) []citrixorchestration.RemotePCEnrollmentScopeRequestModel {
	remotePCEnrollmentScopes := []citrixorchestration.RemotePCEnrollmentScopeRequestModel{}
	if plan.RemotePcOus != nil {
		for _, ou := range plan.RemotePcOus {
			var remotePCEnrollmentScope citrixorchestration.RemotePCEnrollmentScopeRequestModel
			remotePCEnrollmentScope.SetIncludeSubfolders(ou.IncludeSubFolders.ValueBool())
			remotePCEnrollmentScope.SetOU(ou.OUName.ValueString())
			remotePCEnrollmentScope.SetIsOrganizationalUnit(true)
			remotePCEnrollmentScopes = append(remotePCEnrollmentScopes, remotePCEnrollmentScope)
		}
	}

	if includeMachines && plan.MachineAccounts != nil {
		for _, machineAccount := range plan.MachineAccounts {
			for _, machine := range machineAccount.Machines {
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

func (r MachineCatalogResourceModel) updateCatalogWithRemotePcConfig(catalog *citrixorchestration.MachineCatalogDetailResponseModel) MachineCatalogResourceModel {
	if catalog.GetProvisioningType() == citrixorchestration.PROVISIONINGTYPE_MANUAL || !r.IsRemotePc.IsNull() {
		r.IsRemotePc = types.BoolValue(catalog.GetIsRemotePC())
	}
	rpcOUs := util.RefreshListProperties[RemotePcOuModel, citrixorchestration.RemotePCEnrollmentScopeResponseModel](r.RemotePcOus, "OUName", catalog.GetRemotePCEnrollmentScopes(), "OU", "RefreshListItem")
	r.RemotePcOus = rpcOUs
	return r
}
