// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"errors"
	"fmt"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &deliveryGroupPolicyFilterDataSource{}
	_ datasource.DataSourceWithConfigure = &deliveryGroupPolicyFilterDataSource{}
)

func NewDeliveryGroupPolicyFilterDataSource() datasource.DataSource {
	return &deliveryGroupPolicyFilterDataSource{}
}

type deliveryGroupPolicyFilterDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *deliveryGroupPolicyFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_group_policy_filter"
}

func (d *deliveryGroupPolicyFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DeliveryGroupFilterModel{}.GetDataSourceSchema()
}

func (d *deliveryGroupPolicyFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *deliveryGroupPolicyFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data DeliveryGroupFilterModel
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
			// Check if this is a "policy filter not found" error
			if errors.Is(err, util.ErrPolicyFilterNotFound) {
				resp.Diagnostics.AddError(
					"Policy Filter not found",
					fmt.Sprintf("Policy Filter with ID %s was not found.", data.Id.ValueString()),
				)
			}
			return
		}
	}

	// Refresh values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, *policyFilter)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
