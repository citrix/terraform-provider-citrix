// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Create creates the resource and sets the initial Terraform state.
func CreateHypervisorResourcePool(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisor citrixorchestration.HypervisorDetailResponseModel, resourcePoolDetails citrixorchestration.CreateHypervisorResourcePoolRequestModel) (*citrixorchestration.HypervisorResourcePoolDetailResponseModel, error) {
	// Create new hypervisor resource pool
	createResourcePoolRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsCreateResourcePool(ctx, hypervisor.GetId())
	createResourcePoolRequest = createResourcePoolRequest.CreateHypervisorResourcePoolRequestModel(resourcePoolDetails).Async(true)
	_, httpResp, err := citrixdaasclient.AddRequestData(createResourcePoolRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error creating Resource Pool for Hypervisor "+hypervisor.GetName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error creating Resource Pool for Hypervisor "+hypervisor.GetName(), diagnostics, 10)
	if err != nil {
		return nil, err
	}

	// get updated resource pool
	resourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, hypervisor.GetId(), resourcePoolDetails.GetName())
	if err != nil {
		return resourcePool, err
	}

	return resourcePool, nil
}

// Update updates the resource and sets the updated Terraform state on success.
func UpdateHypervisorResourcePool(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId string, resourcePoolId string, editHypervisorResourcePool citrixorchestration.EditHypervisorResourcePoolRequestModel) (*citrixorchestration.HypervisorResourcePoolDetailResponseModel, error) {
	// Patch hypervisor
	patchResourcePoolRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsPatchHypervisorResourcePool(ctx, hypervisorId, resourcePoolId)
	patchResourcePoolRequest = patchResourcePoolRequest.EditHypervisorResourcePoolRequestModel(editHypervisorResourcePool).Async(true)
	httpResp, err := citrixdaasclient.AddRequestData(patchResourcePoolRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating Resource Pool "+resourcePoolId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating Resource Pool "+resourcePoolId, diagnostics, 5)
	if err != nil {
		return nil, err
	}

	// get updated resource pool
	resourcePool, err := util.GetHypervisorResourcePool(ctx, client, diagnostics, hypervisorId, resourcePoolId)
	if err != nil {
		return resourcePool, err
	}

	return resourcePool, nil
}

func ReadHypervisorResourcePool(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, hypervisorId, hypervisorResourcePoolId string) (*citrixorchestration.HypervisorResourcePoolDetailResponseModel, error) {
	getResourcePoolsRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePool(ctx, hypervisorId, hypervisorResourcePoolId)
	resourcePool, _, err := util.ReadResource[*citrixorchestration.HypervisorResourcePoolDetailResponseModel](getResourcePoolsRequest, ctx, client, resp, "Resource Pool", hypervisorResourcePoolId)
	return resourcePool, err
}
