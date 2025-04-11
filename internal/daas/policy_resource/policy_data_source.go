// Copyright © 2024. Citrix Systems, Inc.

package policy_resource

import (
	"context"
	"fmt"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &policyDataSource{}
	_ datasource.DataSourceWithConfigure = &policyDataSource{}
)

func NewPolicyDataSource() datasource.DataSource {
	return &policyDataSource{}
}

type policyDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *policyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *policyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicyModel{}.GetDataSourceSchema()
}

func (d *policyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *policyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *citrixorchestration.PolicyResponse
	var err error
	if !data.Id.IsNull() {
		// Get refreshed policy set from Orchestration
		policy, err = GetPolicy(ctx, d.client, &resp.Diagnostics, data.Id.ValueString(), false, false)
		if err != nil {
			return
		}
	} else if !data.PolicySetId.IsNull() && !data.Name.IsNull() {
		policyResources, err := policies.GetPolicySet(ctx, d.client, &resp.Diagnostics, data.PolicySetId.ValueString())
		if err != nil {
			return
		}
		for _, policyResource := range policyResources.GetPolicies() {
			if strings.EqualFold(data.Name.ValueString(), policyResource.GetPolicyName()) {
				policy = &policyResource
				break
			}
		}
		if policy == nil {
			resp.Diagnostics.AddError(
				"Policy not found",
				fmt.Sprintf("Policy with name %s in Policy Set %s not found.", data.Name.ValueString(), data.PolicySetId.ValueString()),
			)
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, policy)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
