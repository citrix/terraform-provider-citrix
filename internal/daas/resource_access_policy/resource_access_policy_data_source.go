// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy

import (
	"context"
	"fmt"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ResourceAccessPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &ResourceAccessPolicyDataSource{}
)

func NewResourceAccessPolicyDataSource() datasource.DataSource {
	return &ResourceAccessPolicyDataSource{}
}

// ResourceAccessPolicyDataSource defines the data source implementation.
type ResourceAccessPolicyDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ResourceAccessPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_access_policy"
}

func (d *ResourceAccessPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ResourceAccessPolicyDataSourceModel{}.GetSchema()
}

func (d *ResourceAccessPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient) //nolint:forcetypeassert // framework guarantee
}

func (d *ResourceAccessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data ResourceAccessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyIdOrName := data.Id.ValueString()
	if data.Id.IsNull() {
		policyIdOrName = data.Name.ValueString()
	}
	errorSummary := "Error reading Resource Access Policy " + policyIdOrName

	policies, err := getResourceAccessPolicies(ctx, d.client, &resp.Diagnostics, errorSummary)
	if err != nil {
		return
	}

	policy := findPolicyById(policies, data.Id.ValueString())
	if data.Id.IsNull() {
		policy = findPolicyByName(policies, data.Name.ValueString())
	}
	if policy == nil {
		resp.Diagnostics.AddError(
			errorSummary,
			fmt.Sprintf("Resource Access Policy %s was not found.", policyIdOrName),
		)
		return
	}

	tagId, err := getTagIdOfPolicy(ctx, d.client, &resp.Diagnostics, policy, errorSummary)
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, policy, tagId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
