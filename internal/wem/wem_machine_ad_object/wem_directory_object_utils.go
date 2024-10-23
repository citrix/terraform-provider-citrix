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

func readDirectoryObject(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, machineADObjectId string) (*citrixwemservice.MachineModel, error) {
	idInt64, err := strconv.ParseInt(machineADObjectId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid WEM Directory object ID: %v", err)
	}
	machineADObjectQueryRequest := client.WemClient.MachineADObjectDAAS.AdObjectQueryById(ctx, idInt64)
	machineADObjectQueryResponse, _, err := util.ReadResource[*citrixwemservice.MachineModel](machineADObjectQueryRequest, ctx, client, resp, "Catalog Directory Object", machineADObjectId)
	return machineADObjectQueryResponse, err
}

func getMachineADObjectBySid(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machineCatalogId string) (citrixwemservice.MachineModel, error) {
	machineADObjectQueryRequest := client.WemClient.MachineADObjectDAAS.AdObjectQuery(ctx)
	machineADObjectQueryRequest = machineADObjectQueryRequest.Sid(machineCatalogId)
	machineADObjectQueryResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixwemservice.AdObjectQuery200Response](machineADObjectQueryRequest, client)

	machineADObjectList := machineADObjectQueryResponse.GetItems()

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return citrixwemservice.MachineModel{}, err
	}

	if len(machineADObjectList) != 0 {
		return machineADObjectList[0], nil
	}
	return citrixwemservice.MachineModel{}, fmt.Errorf("WEM Directory object with SID " + machineCatalogId + " not found")
}

func getMachineADObjectById(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, machineADObjectId string) (*citrixwemservice.MachineModel, error) {
	idInt64, err := strconv.ParseInt(machineADObjectId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid WEM Directory object ID: %v", err)
	}
	machineADObjectQueryRequest := client.WemClient.MachineADObjectDAAS.AdObjectQueryById(ctx, idInt64)
	machineADObjectQueryResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixwemservice.MachineModel](machineADObjectQueryRequest, client)

	if err != nil {
		err = fmt.Errorf("TransactionId: " + citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp) + "\nError message: " + util.ReadClientError(err))
		return machineADObjectQueryResponse, err
	}

	if machineADObjectQueryResponse == nil {
		return machineADObjectQueryResponse, fmt.Errorf("wem directory object with ID " + machineADObjectId + " not found")
	}
	return machineADObjectQueryResponse, nil
}
