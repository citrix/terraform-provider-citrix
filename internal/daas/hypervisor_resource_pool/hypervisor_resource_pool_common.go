// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"fmt"
	"time"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure HypervisorStorageModel implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixorchestration.HypervisorStorageResourceResponseModel] = HypervisorStorageModel{}

type HypervisorStorageModel struct {
	StorageName types.String `tfsdk:"storage_name"`
	Superseded  types.Bool   `tfsdk:"superseded"`
}

func (h HypervisorStorageModel) GetKey() string {
	return h.StorageName.ValueString()
}

func (v HypervisorStorageModel) RefreshListItem(_ context.Context, _ *diag.Diagnostics, remote citrixorchestration.HypervisorStorageResourceResponseModel) util.ResourceModelWithAttributes {
	v.StorageName = types.StringValue(remote.GetName())
	v.Superseded = types.BoolValue(remote.GetSuperseded())
	return v
}

func (HypervisorStorageModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"storage_name": schema.StringAttribute{
				Description: "The name of the storage.",
				Required:    true,
			},
			"superseded": schema.BoolAttribute{
				Description: "Indicates whether the storage has been superseded. Superseded storage may be used for existing virtual machines, but is not used when provisioning new virtual machines. Use only when updating the resource pool.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (HypervisorStorageModel) GetAttributes() map[string]schema.Attribute {
	return HypervisorStorageModel{}.GetSchema().Attributes
}

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

func getHypervisorResourcePoolSubnets(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, folderPath string, subnets []string, connectionType citrixorchestration.HypervisorConnectionType) ([]string, error) {
	remoteSubnets, err := util.GetFilteredResourcePathListWithNoCacheRetry(ctx, client, diagnostics, hypervisorId, folderPath, util.NetworkResourceType, subnets, connectionType, "")
	if err != nil {
		diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for "+string(connectionType),
			"Error message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	if len(remoteSubnets) != len(subnets) {
		diagnostics.AddError(
			"Error creating Hypervisor Resource Pool for "+string(connectionType),
			"Subnet contains invalid value.",
		)
		return nil, fmt.Errorf("subnet contains invalid value")
	}

	return remoteSubnets, nil
}

func waitForProvImagesPendingDelete(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorId, resourcePoolId string, timeoutMinutes int32) error {
	// Poll interval in seconds
	pollInterval := 60 * time.Second
	deadline := time.Now().Add(time.Duration(timeoutMinutes) * time.Minute)

	// Accumulate last seen blocking details to include in the final error if we time out
	var lastHosts []string
	var lastImages map[string]int32
	var lastCatalogs map[string]int32
	var lastTasks int32
	var lastProvisioningSchemes int32
	var totalImpactedCatalogs int32

	for time.Now().Before(deadline) {
		poolDeleteRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisorResourcePoolDeletePreview(ctx, hypervisorId, resourcePoolId)
		poolDeleteResponseModel, httpResp, err := citrixdaasclient.AddRequestData(poolDeleteRequest, client).Execute()
		if err != nil {
			diagnostics.AddError(
				"Error fetching resource pool delete preview",
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return err
		}

		// Determine whether any items would block deletion. If none, proceed.
		if poolDeleteResponseModel != nil {
			if poolDeleteResponseModel.GetCanBeDeleted() {
				return nil
			} else {
				if poolDeleteResponseModel.GetTotalImpactedCatalogCount() > 0 {
					totalImpactedCatalogs = poolDeleteResponseModel.GetTotalImpactedCatalogCount()
					lastCatalogs = poolDeleteResponseModel.GetImpactedCatalogs()
				}

				if poolDeleteResponseModel.GetTaskReferences() > 0 {
					lastTasks = poolDeleteResponseModel.GetTaskReferences()
				}

				if len(poolDeleteResponseModel.GetHostsToBeDeleted()) > 0 {
					lastHosts = poolDeleteResponseModel.GetHostsToBeDeleted()
				}

				if len(poolDeleteResponseModel.GetImpactedImageDefinitions()) > 0 {
					lastImages = poolDeleteResponseModel.GetImpactedImageDefinitions()
				}

				if poolDeleteResponseModel.GetProvisioningSchemeReferences() > 0 {
					lastProvisioningSchemes = poolDeleteResponseModel.GetProvisioningSchemeReferences()
				}
			}
		}

		time.Sleep(pollInterval)
	}

	// Timed out - include last seen blocking objects in a multi-line diagnostics message
	lines := []string{
		fmt.Sprintf("Machine Catalogs, image definitions, and related resources pending deletion remain for resource pool %s after %d minutes; check for in-progress tasks, impacted catalogs, image definitions, or hosts to be deleted and resolve them before retrying", resourcePoolId, timeoutMinutes),
	}

	if totalImpactedCatalogs > 0 {
		lines = append(lines, fmt.Sprintf("Total impacted catalog count: %d", totalImpactedCatalogs))
		lines = append(lines, "Impacted catalogs:")
		for catalogName, imgCount := range lastCatalogs {
			lines = append(lines, fmt.Sprintf("  - %s (provisioned images pending delete: %d)", catalogName, imgCount))
		}
	}
	if len(lastImages) > 0 {
		lines = append(lines, "Impacted image definitions:")
		for imgName, imgCount := range lastImages {
			lines = append(lines, fmt.Sprintf("  - %s (provisioned images pending delete: %d)", imgName, imgCount))
		}
	}

	if lastTasks > 0 {
		lines = append(lines, fmt.Sprintf("Task references: %d", lastTasks))
	}
	if lastProvisioningSchemes > 0 {
		lines = append(lines, fmt.Sprintf("Provisioning scheme references: %d", lastProvisioningSchemes))
	}
	if len(lastHosts) > 0 {
		lines = append(lines, fmt.Sprintf("Hosts to be deleted: %v", lastHosts))
	}

	details := ""
	for i, l := range lines {
		if i == 0 {
			details = l
		} else {
			details += "\n" + l
		}
	}

	diagnostics.AddError(
		"Timeout waiting for Machine Catalogs, image definitions, and related resources to be deleted",
		details,
	)
	return fmt.Errorf("timed out after %d minutes waiting for Machine Catalogs, image definitions, and related resources to be deleted from resource pool %s", timeoutMinutes, resourcePoolId)

}
