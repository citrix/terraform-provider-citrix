// Copyright © 2024. Citrix Systems, Inc.

package policy_setting

import (
	"context"
	"fmt"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policy_resource"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &policySettingDataSource{}
	_ datasource.DataSourceWithConfigure = &policySettingDataSource{}
)

func NewPolicySettingDataSource() datasource.DataSource {
	return &policySettingDataSource{}
}

type policySettingDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *policySettingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_setting"
}

func (d *policySettingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicySettingModel{}.GetDataSourceSchema()
}

func (d *policySettingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *policySettingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicySettingModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policySetting *citrixorchestration.SettingResponse
	var err error
	if !data.Id.IsNull() {
		// Get refreshed policy set from Orchestration
		policySetting, err = getPolicySetting(ctx, d.client, &resp.Diagnostics, data.Id.ValueString())
		if err != nil {
			return
		}
	} else if !data.Name.IsNull() && !data.PolicyId.IsNull() {
		policy, err := policy_resource.GetPolicy(ctx, d.client, &resp.Diagnostics, data.PolicyId.ValueString(), false, true)
		if err != nil {
			return
		}
		for _, remoteSetting := range policy.GetSettings() {
			if strings.EqualFold(data.Name.ValueString(), remoteSetting.GetSettingName()) {
				policySetting = &remoteSetting
				break
			}
		}
		if policySetting == nil {
			resp.Diagnostics.AddError(
				"Policy Setting not found",
				fmt.Sprintf("Policy Setting with name %s not found in Policy %s.", data.Name.ValueString(), data.PolicyId.ValueString()),
			)
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(policySetting)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
