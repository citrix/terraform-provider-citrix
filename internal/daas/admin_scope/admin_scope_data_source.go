// Copyright © 2025. Citrix Systems, Inc.

package admin_scope

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ datasource.DataSource              = &AdminScopeDataSource{}
	_ datasource.DataSourceWithConfigure = &AdminScopeDataSource{}
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
	resp.Schema = AdminScopeModel{}.GetDataSourceSchema()
}

func (d *AdminScopeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient) //nolint:forcetypeassert // framework guarantee
}

func (d *AdminScopeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AdminScopeModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var adminScopeNameOrId string

	if !data.Id.IsNull() {
		adminScopeNameOrId = data.Id.ValueString()
	} else {
		adminScopeNameOrId = data.Name.ValueString()
	}

	getAdminScopeRequest := d.client.ApiClient.AdminAPIsDAAS.AdminGetAdminScope(ctx, adminScopeNameOrId)
	adminScope, httpResp, err := citrixdaasclient.AddRequestData(getAdminScopeRequest, d.client).Execute()

	if err != nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		// Check for Tenant ID
		adminScope, httpResp, err = getTenantScopeByTenantNameOrId(ctx, d.client, &resp.Diagnostics, adminScopeNameOrId)
	}

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

func getTenantScopeByTenantNameOrId(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, tenantNameOrId string) (*citrixorchestration.ScopeResponseModel, *http.Response, error) {
	adminScopes, httpResp, err := util.FetchScopes(ctx, client, diagnostics)
	if err != nil {
		return nil, httpResp, err
	}
	for _, scope := range adminScopes {
		if strings.EqualFold(scope.GetTenantId(), tenantNameOrId) || strings.EqualFold(scope.GetTenantName(), tenantNameOrId) {
			return &scope, httpResp, nil
		}
	}
	return nil, httpResp, fmt.Errorf("no scope matched for scope name: %s", tenantNameOrId)
}
