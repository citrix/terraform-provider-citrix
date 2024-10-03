package wem_machine_ad_object

import (
	"context"
	"fmt"
	"strconv"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	citrixwemservice "github.com/citrix/citrix-daas-rest-go/devicemanagement"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func getMachineADObjectBySid(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machineCatalogId string) (citrixwemservice.MachineModel, error) {
	var resp *resource.ReadResponse
	machineADObjectQueryRequest := client.WemClient.MachineADObjectDAAS.AdObjectQuery(ctx)
	machineADObjectQueryRequest = machineADObjectQueryRequest.Sid(machineCatalogId)
	machineADObjectQueryResponse, httpResp, err := util.ReadResource[*citrixwemservice.AdObjectQuery200Response](machineADObjectQueryRequest, ctx, client, resp, "Sid", machineCatalogId)

	var machineADObject citrixwemservice.MachineModel
	machineADObjectList := machineADObjectQueryResponse.GetItems()

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return machineADObject, err
	}

	if len(machineADObjectList) != 0 {
		machineADObject = machineADObjectList[0]
	}

	if (machineADObject == citrixwemservice.MachineModel{}) {
		return machineADObject, fmt.Errorf("WEM Directory object with SID " + machineCatalogId + " not found")
	}
	return machineADObject, nil
}

func getMachineADObjectById(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machineADObjectId string) (*citrixwemservice.MachineModel, error) {
	var resp *resource.ReadResponse
	idInt64, err := strconv.ParseInt(machineADObjectId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid WEM Directory object ID: %v", err)
	}
	machineADObjectQueryRequest := client.WemClient.MachineADObjectDAAS.AdObjectQueryById(ctx, idInt64)
	machineADObjectQueryResponse, httpResp, err := util.ReadResource[*citrixwemservice.MachineModel](machineADObjectQueryRequest, ctx, client, resp, "Id", machineADObjectId)

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return machineADObjectQueryResponse, err
	}

	if machineADObjectQueryResponse == nil {
		return machineADObjectQueryResponse, fmt.Errorf("wem directory object with ID " + machineADObjectId + " not found")
	}
	return machineADObjectQueryResponse, nil
}
