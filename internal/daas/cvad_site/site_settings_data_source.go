// Copyright © 2024. Citrix Systems, Inc.
package cvad_site

import (
	"context"
	"fmt"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &SiteSettingsDataSource{}
)

func NewSiteSettingsDataSource() datasource.DataSource {
	return &SiteSettingsDataSource{}
}

type SiteSettingsDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *SiteSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_settings"
}

func (d *SiteSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SiteSettingsModel{}.GetDataSourceSchema()
}

func (d *SiteSettingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	// Remove Site ID from the URL
	configuredURL := d.client.ApiClient.GetConfig().Servers[0].URL
	updatedUrl := strings.ReplaceAll(configuredURL, fmt.Sprintf("/%s/%s", d.client.ClientConfig.CustomerId, d.client.ClientConfig.SiteId), "/"+d.client.ClientConfig.CustomerId)
	d.client.ApiClient.GetConfig().Servers[0].URL = updatedUrl
}

func (d *SiteSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data SiteSettingsModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh data.SiteSettings
	siteSettings, err := getSiteSettings(ctx, d.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	// Get multiple remote PC assignments settings from Orchestration
	multipleRemotePCAssignments, err := getMultipleRemotePCAssignmentsSetting(ctx, d.client, &resp.Diagnostics)
	if err != nil {
		return
	}
	data = data.RefreshDataSourcePropertyValues(ctx, &resp.Diagnostics, d.client, siteSettings, multipleRemotePCAssignments)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
