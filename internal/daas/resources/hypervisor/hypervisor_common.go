// Copyright © 2023. Citrix Systems, Inc.

package hypervisor

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Create creates the resource and sets the initial Terraform state.
func CreateHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, createHypervisorRequestBody citrixorchestration.CreateHypervisorRequestModel) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	// Create new hypervisor
	createHypervisorRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsCreateHypervisor(ctx)
	createHypervisorRequest = createHypervisorRequest.CreateHypervisorRequestModel(createHypervisorRequestBody).Async(true)

	_, httpResp, err := citrixdaasclient.AddRequestData(createHypervisorRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error creating Hypervisor",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error creating Hypervisor", diagnostics, 10)
	if err != nil {
		return nil, err
	}

	hypervisor, err := util.GetHypervisor(ctx, client, diagnostics, createHypervisorRequestBody.ConnectionDetails.GetName())
	if err != nil {
		return hypervisor, err
	}

	return hypervisor, nil
}

// Update updates the resource and sets the updated Terraform state on success.
func UpdateHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel, editHypervisorRequestBody citrixorchestration.EditHypervisorConnectionRequestModel) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	// Patch hypervisor
	patchHypervisorRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsPatchHypervisor(ctx, hypervisor.GetId())
	patchHypervisorRequest = patchHypervisorRequest.EditHypervisorConnectionRequestModel(editHypervisorRequestBody).Async(true)
	httpResp, err := citrixdaasclient.AddRequestData(patchHypervisorRequest, client).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error updating Hypervisor "+hypervisor.GetName(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	err = util.ProcessAsyncJobResponse(ctx, client, httpResp, "Error updating Hypervisor "+hypervisor.GetName(), diagnostics, 10)
	if err != nil {
		return nil, err
	}

	// Fetch updated hypervisor from GetHypervisor
	updatedHypervisor, err := util.GetHypervisor(ctx, client, diagnostics, hypervisor.GetId())
	if err != nil {
		return nil, err
	}

	return updatedHypervisor, err
}

func readHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, hypervisorId string) (*citrixorchestration.HypervisorDetailResponseModel, error) {
	getHypervisorReq := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorId)
	hypervisor, _, err := util.ReadResource[*citrixorchestration.HypervisorDetailResponseModel](getHypervisorReq, ctx, client, resp, "Hypervisor", hypervisorId)
	return hypervisor, err
}
