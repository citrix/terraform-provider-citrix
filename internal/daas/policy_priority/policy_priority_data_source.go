// Copyright © 2024. Citrix Systems, Inc.

package policy_priority

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
	_ datasource.DataSource              = &policyPriorityDataSource{}
	_ datasource.DataSourceWithConfigure = &policyPriorityDataSource{}
)

func NewPolicyPriorityDataSource() datasource.DataSource {
	return &policyPriorityDataSource{}
}

type policyPriorityDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *policyPriorityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_priority"
}

func (d *policyPriorityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicyPriorityModel{}.GetDataSourceSchema()
}

func (d *policyPriorityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *policyPriorityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicyPriorityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policySet *citrixorchestration.PolicySetResponse
	var policiesInPolicySet *citrixorchestration.CollectionEnvelopeOfPolicyResponse
	var err error
	if !data.PolicySetId.IsNull() {
		// Get refreshed policy set from Orchestration
		policySet, err = policies.GetPolicySet(ctx, d.client, &resp.Diagnostics, data.PolicySetId.ValueString())
		if err != nil {
			return
		}

		policiesInPolicySet, err = policies.GetPolicies(ctx, d.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
		if err != nil {
			return
		}
	} else if !data.PolicySetName.IsNull() {
		policySets, err := policies.GetPolicySets(ctx, d.client, &resp.Diagnostics)
		if err != nil {
			return
		}
		for _, remotePolicySet := range policySets {
			if strings.EqualFold(data.PolicySetName.ValueString(), remotePolicySet.GetName()) {
				policySet = &remotePolicySet
				break
			}
		}
		if policySet == nil {
			resp.Diagnostics.AddError(
				"Policy Set not found",
				fmt.Sprintf("Policy Set with name %s not found.", data.PolicySetName.ValueString()),
			)
			return
		}
		policiesInPolicySet, err = policies.GetPolicies(ctx, d.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
		if err != nil {
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, policySet, policiesInPolicySet)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
