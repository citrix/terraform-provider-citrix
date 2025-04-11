// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policy_resource"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &policyFiltersDataSource{}
	_ datasource.DataSourceWithConfigure = &policyFiltersDataSource{}
)

func NewPolicyFiltersDataSource() datasource.DataSource {
	return &policyFiltersDataSource{}
}

type policyFiltersDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *policyFiltersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_filters"
}

func (d *policyFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicyFiltersModel{}.GetDataSourceSchema()
}

func (d *policyFiltersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *policyFiltersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicyFiltersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *citrixorchestration.PolicyResponse
	var err error
	if !data.PolicyId.IsNull() {
		// Get refreshed policy set from Orchestration
		policy, err = policy_resource.GetPolicy(ctx, d.client, &resp.Diagnostics, data.PolicyId.ValueString(), true, false)
		if err != nil {
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, *policy)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
