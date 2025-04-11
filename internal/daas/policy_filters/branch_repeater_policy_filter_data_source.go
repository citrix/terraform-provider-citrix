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
	_ datasource.DataSource              = &branchRepeaterPolicyFilterDataSource{}
	_ datasource.DataSourceWithConfigure = &branchRepeaterPolicyFilterDataSource{}
)

func NewBranchRepeaterPolicyFilterDataSource() datasource.DataSource {
	return &branchRepeaterPolicyFilterDataSource{}
}

type branchRepeaterPolicyFilterDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *branchRepeaterPolicyFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_repeater_policy_filter"
}

func (d *branchRepeaterPolicyFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = BranchRepeaterFilterModel{}.GetDataSourceSchema()
}

func (d *branchRepeaterPolicyFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *branchRepeaterPolicyFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data BranchRepeaterFilterModel
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
