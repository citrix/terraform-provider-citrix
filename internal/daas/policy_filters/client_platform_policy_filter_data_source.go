// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clientPlatformPolicyFilterDataSource{}
	_ datasource.DataSourceWithConfigure = &clientPlatformPolicyFilterDataSource{}
)

func NewClientPlatformPolicyFilterDataSource() datasource.DataSource {
	return &clientPlatformPolicyFilterDataSource{}
}

type clientPlatformPolicyFilterDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *clientPlatformPolicyFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_platform_policy_filter"
}

func (d *clientPlatformPolicyFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ClientPlatformFilterModel{}.GetDataSourceSchema()
}

func (d *clientPlatformPolicyFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *clientPlatformPolicyFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	isDdcVersionSupported := util.CheckProductVersion(d.client, &resp.Diagnostics, 124, 124, 7, 44, "Error reading Client Platform policy filter", "Client Platform policy filter data source")
	if !isDdcVersionSupported {
		return
	}

	var data ClientPlatformFilterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policyFilter *citrixorchestration.FilterResponse
	var err error
	if !data.Id.IsNull() {
		// Get refreshed policy set from Orchestration
		policyFilter, err = getPolicyFilter(ctx, d.client, &resp.Diagnostics, data.Id.ValueString())
		if err != nil {
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, *policyFilter)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
