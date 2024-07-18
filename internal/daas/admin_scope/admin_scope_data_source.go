// Copyright © 2024. Citrix Systems, Inc.

package admin_scope

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &AdminScopeDataSource{}
)

func NewAdminScopeDataSource() datasource.DataSource {
	return &AdminScopeDataSource{}
}

type AdminScopeDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AdminScopeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_scope"
}

func (d *AdminScopeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GetAdminScopeDataSourceSchema()
}

func (d *AdminScopeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AdminScopeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AdminScopeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var adminScopeNameOrId string

	if data.Id.ValueString() != "" {
		adminScopeNameOrId = data.Id.ValueString()
	}
	if data.Name.ValueString() != "" {
		adminScopeNameOrId = data.Name.ValueString()
	}

	getAdminScopeRequest := d.client.ApiClient.AdminAPIsDAAS.AdminGetAdminScope(ctx, adminScopeNameOrId)
	adminScope, httpResp, err := citrixdaasclient.AddRequestData(getAdminScopeRequest, d.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing AdminScope: "+adminScopeNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, adminScope)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
