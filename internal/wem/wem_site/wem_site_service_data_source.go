// Copyright © 2024. Citrix Systems, Inc.

package wem_site

import (
	"context"
	"strconv"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = &WemSiteDataSource{}
	_ datasource.DataSourceWithConfigure = &WemSiteDataSource{}
)

func NewWemSiteDataSource() datasource.DataSource {
	return &WemSiteDataSource{}
}

type WemSiteDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// WEM Configuration Set and WEM Site refer to the same object. These terms have been used interchangeably below.
func (d *WemSiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wem_configuration_set"
}

func (d *WemSiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GetWemSiteDataSourceSchema()
}

func (d *WemSiteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *WemSiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data WemSiteDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var wemSiteNameOrId string

	getWemSiteRequest := d.client.WemClient.SiteDAAS.SiteQuery(ctx)
	if !data.Id.IsNull() {
		wemSiteNameOrId = data.Id.ValueString()
		idInt64, err := strconv.ParseInt(data.Id.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error fetching WEM configuration set",
				"The WEM configuration set id format is incorrect. Please provide an existing WEM configuration set id and try again.",
			)
			return
		}
		getWemSiteRequest = getWemSiteRequest.Id(idInt64)
	} else {
		wemSiteNameOrId = data.Name.ValueString()
		getWemSiteRequest = getWemSiteRequest.Name(data.Name.ValueString())
	}
	getWemSiteResponse, httpResp, err := citrixdaasclient.AddRequestData(getWemSiteRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching WEM configuration set: "+wemSiteNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	wemSites := getWemSiteResponse.GetItems()

	if len(wemSites) == 0 {
		resp.Diagnostics.AddError(
			"Error fetching WEM configuration set",
			"Could not find WEM configuration set: "+wemSiteNameOrId+". Please provide the id or name of an existing WEM configuration set and try again.",
		)
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, &wemSites[0])

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
