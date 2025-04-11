// Copyright © 2024. Citrix Systems, Inc.

package policy_set_resource

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &policySetV2DataSource{}
	_ datasource.DataSourceWithConfigure = &policySetV2DataSource{}
)

func NewPolicySetV2DataSource() datasource.DataSource {
	return &policySetV2DataSource{}
}

type policySetV2DataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *policySetV2DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set_v2"
}

func (d *policySetV2DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicySetV2Model{}.GetDataSourceSchema()
}

func (d *policySetV2DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *policySetV2DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicySetV2Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policySet *citrixorchestration.PolicySetResponse
	var err error
	if !data.Id.IsNull() {
		// Get refreshed policy set from Orchestration
		policySet, err = policies.GetPolicySet(ctx, d.client, &resp.Diagnostics, data.Id.ValueString())
		if err != nil {
			return
		}
	} else {
		policySets, err := policies.GetPolicySets(ctx, d.client, &resp.Diagnostics)
		if err != nil {
			return
		}
		for _, remotePolicySet := range policySets {
			if strings.EqualFold(data.Name.ValueString(), remotePolicySet.GetName()) {
				policySet = &remotePolicySet
				break
			}
		}
		if policySet == nil {
			resp.Diagnostics.AddError("Policy Set not found", "Policy Set with name "+data.Name.ValueString()+" not found.")
			return
		}
	}

	// Refresh values
	policySet, policySetScopes, associatedDeliveryGroups, err := getPolicySetDetailsForRefreshState(ctx, &resp.Diagnostics, d.client, policySet.GetPolicySetGuid())
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, policySet, policySetScopes, associatedDeliveryGroups)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
