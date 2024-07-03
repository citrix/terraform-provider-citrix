// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"fmt"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &PvsDataSource{}
)

func NewPvsDataSource() datasource.DataSource {
	return &PvsDataSource{}
}

type PvsDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *PvsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pvs"
}

func (d *PvsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PvsDataSourceModel{}.GetSchema()
}

func (d *PvsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *PvsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data PvsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pvsFarmName := data.FarmName.ValueString()
	pvsSiteName := data.SiteName.ValueString()
	pvsStoreName := data.StoreName.ValueString()
	pvsVdiskName := data.VdiskName.ValueString()

	// Get the farm id and the pvs site id from the pvs farm and site names
	getPvsStreamingSitesRequest := d.client.ApiClient.PvsStreamingAPIsDAAS.PvsStreamingGetPvsStreamingSites(ctx)
	pvsStreamingSitesResponse, httpResp, err := citrixdaasclient.AddRequestData(getPvsStreamingSitesRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching PVS sites",
			"TransactionId:"+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Loop through items to find out the block for which the farm name and site names match
	var pvsSiteId string
	var pvsFarmId string
	for _, site := range pvsStreamingSitesResponse.Items {
		if strings.EqualFold(site.GetFarmName(), pvsFarmName) && strings.EqualFold(site.GetSiteName(), pvsSiteName) {
			pvsSiteId = site.GetSiteId()
			pvsFarmId = site.GetFarmId()
		}
	}

	if pvsSiteId == "" || pvsFarmId == "" {
		err = fmt.Errorf("one or more values for the fields pvs_farm_name and pvs_site_name are incorrect. Please check the values and try again")
		resp.Diagnostics.AddError(
			"Error reading PVS farm and site details",
			err.Error(),
		)

		return
	}

	// Use the farm id to fetch the store details
	getPvsStreamingStoresRequest := d.client.ApiClient.PvsStreamingAPIsDAAS.PvsStreamingGetPvsStreamingStores(ctx, pvsFarmId)
	pvsStreamingStoresResponse, httpResp, err := citrixdaasclient.AddRequestData(getPvsStreamingStoresRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PVS stores",
			"TransactionId:"+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)

		return
	}

	// Loop through items to find out the block for which the store name matches
	var pvsStoreId string
	for _, store := range pvsStreamingStoresResponse.Items {
		if strings.EqualFold(store.GetStoreName(), pvsStoreName) {
			pvsStoreId = store.GetStoreId()
		}
	}

	if pvsStoreId == "" {
		err = fmt.Errorf("could not find a store with name: %s. Please check the value of pvs_store_name field and try again", pvsStoreName)
		resp.Diagnostics.AddError(
			"Error fetching PVS store details",
			err.Error(),
		)

		return
	}

	// Use the farmID, pvs site id and the pvs store id from the previous call to fetch the disklocatorId which is the vdisk id to be used for the pvs deployment
	getPvsStreamingVdisksRequest := d.client.ApiClient.PvsStreamingAPIsDAAS.PvsStreamingGetPvsStreamingVDisks(ctx)
	getPvsStreamingVdisksRequest = getPvsStreamingVdisksRequest.FarmId(pvsFarmId).PvsSiteId(pvsSiteId).StoreId(pvsStoreId)
	pvsStreamingVdisksResponse, httpResp, err := citrixdaasclient.AddRequestData(getPvsStreamingVdisksRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PVS vdisks",
			"TransactionId:"+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)

		return
	}

	// Loop through items to find out the block for which the vdisk name matches
	var pvsVdiskId string
	for _, vdisk := range pvsStreamingVdisksResponse.Items {
		if strings.EqualFold(vdisk.GetDiskLocatorName(), pvsVdiskName) {
			pvsVdiskId = vdisk.GetDiskLocatorId()
		}
	}

	if pvsVdiskId == "" {
		err = fmt.Errorf("could not find a vdisk with name: %s. Please check the value of pvs_vdisk_name field and try again", pvsVdiskName)
		resp.Diagnostics.AddError(
			"Error fetching PVS vdisk details",
			err.Error(),
		)

		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, pvsSiteId, pvsVdiskId)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
