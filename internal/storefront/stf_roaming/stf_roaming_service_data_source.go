// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"
	"strconv"

	citrixstorefrontModels "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &STFRoamingServiceDataSource{}
)

func NewSTFRoamingServiceDataSource() datasource.DataSource {
	return &STFRoamingServiceDataSource{}
}

type STFRoamingServiceDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata implements datasource.DataSource.
func (*STFRoamingServiceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_roaming_service"
}

// Schema implements datasource.DataSource.
func (*STFRoamingServiceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GetSTFRoamingServiceDataSourceSchema()
}

func (d *STFRoamingServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Read implements datasource.DataSource.
func (d *STFRoamingServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data STFRoamingServiceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var getBody citrixstorefrontModels.STFRoamingServiceRequestModel
	siteIdInt, err := strconv.ParseInt(data.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing site_id of StoreFront Roaming Service ",
			"\nError message: "+err.Error(),
		)
		return
	}
	getBody.SetSiteId(siteIdInt)
	getRoamingServiceRequest := d.client.StorefrontClient.RoamingSF.STFRoamingServiceGet(ctx, getBody)
	roamingService, err := getRoamingServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get StoreFront Roaming Service details",
			"Error message: "+err.Error())
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, &roamingService)
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
