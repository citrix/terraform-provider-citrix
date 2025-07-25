package autoscale_plugin_template

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &AutoscalePluginTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &AutoscalePluginTemplateDataSource{}
)

func NewAutoscalePluginTemplateDataSource() datasource.DataSource {
	return &AutoscalePluginTemplateDataSource{}
}

// AutoscalePluginTemplateDataSource defines the data source implementation.
type AutoscalePluginTemplateDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Configure implements datasource.DataSourceWithConfigure.
func (r *AutoscalePluginTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Metadata implements datasource.DataSource.
func (r *AutoscalePluginTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_autoscale_plugin_template"
}

// Read implements datasource.DataSource.
func (r *AutoscalePluginTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	isFeatureSupported := util.CheckProductVersion(r.client, &resp.Diagnostics, 125, 124, 7, 44, "Error reading Autoscale Plugin Template data resource", "Autoscale Plugin Template data resource")
	if !isFeatureSupported {
		return
	}

	var data AutoscalePluginTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	autoscalePluginTemplate, err := getAutoscalePluginTemplate(ctx, r.client, &resp.Diagnostics, data.Type.ValueString(), data.Name.ValueString())
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, autoscalePluginTemplate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Schema implements datasource.DataSource.
func (r *AutoscalePluginTemplateDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AutoscalePluginTemplateResourceModel{}.GetDataSourceSchema()
}
