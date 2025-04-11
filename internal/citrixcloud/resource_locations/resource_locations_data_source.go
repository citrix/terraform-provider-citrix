// Copyright © 2024. Citrix Systems, Inc.

package resource_locations

import (
	"context"
	"strings"

	ccresourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	resourcelocations "github.com/citrix/citrix-daas-rest-go/ccresourcelocations"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ResourceLocationsDataSource{}
	_ datasource.DataSourceWithConfigure = &ResourceLocationsDataSource{}
)

func NewResourceLocationsDataSource() datasource.DataSource {
	return &ResourceLocationsDataSource{}
}

// ResourceLocationsDataSource defines the data source implementation.
type ResourceLocationsDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ResourceLocationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_resource_location"
}

func (d *ResourceLocationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ResourceLocationModel{}.GetDataSourceSchema()
}

func (d *ResourceLocationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ResourceLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data ResourceLocationModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get resource location
	getResourceLocationRequest := d.client.ResourceLocationsClient.LocationsDAAS.LocationsGetAll(ctx)
	resourceLocation, httpResp, err := citrixdaasclient.ExecuteWithRetry[*resourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationsResultsModel](getResourceLocationRequest, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get resource location."+err.Error(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	var matchedResourceLocation *ccresourcelocations.CitrixCloudServicesRegistryApiModelsLocationsResourceLocationModel
	// Refresh data with the latest state
	for _, rl := range resourceLocation.Items {
		if strings.EqualFold(rl.GetName(), data.Name.ValueString()) {
			matchedResourceLocation = &rl
			break
		}
	}

	if matchedResourceLocation == nil {
		resp.Diagnostics.AddError(
			"Error reading Resource Location",
			"Resource location with name "+data.Name.ValueString()+" was not found",
		)
		return
	}

	data = data.RefreshPropertyValues(matchedResourceLocation)
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
