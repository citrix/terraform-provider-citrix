// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &PolicySetDataSource{}
	_ datasource.DataSourceWithConfigure = &PolicySetDataSource{}
)

func NewPolicySetDataSource() datasource.DataSource {
	return &PolicySetDataSource{}
}

// PolicySetDataSource defines the data source implementation.
type PolicySetDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *PolicySetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set"
}

func (d *PolicySetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PolicySetModel{}.GetDataSourceSchema()
}

func (d *PolicySetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *PolicySetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data PolicySetModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var policySet *citrixorchestration.PolicySetResponse
	var err error
	if !data.Id.IsNull() {
		// Get refreshed policy set from Orchestration
		policySet, err = GetPolicySet(ctx, d.client, &resp.Diagnostics, data.Id.ValueString())
		if err != nil {
			return
		}
	} else {
		policySets, err := GetPolicySets(ctx, d.client, &resp.Diagnostics)
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

	policies, err := GetPolicies(ctx, d.client, &resp.Diagnostics, policySet.GetPolicySetGuid())
	if err != nil {
		return
	}

	policySetScopes, err := util.FetchScopeIdsByNames(ctx, resp.Diagnostics, d.client, policySet.GetScopes())
	if err != nil {
		return
	}

	deliveryGroups, err := util.GetDeliveryGroups(ctx, d.client, &resp.Diagnostics, "Id,PolicySetGuid")
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, policySet, policies, policySetScopes, deliveryGroups)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
