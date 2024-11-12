// Copyright © 2024. Citrix Systems, Inc.
package admin_role

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &AdminRoleDataSource{}
)

func NewAdminRoleDataSource() datasource.DataSource {
	return &AdminRoleDataSource{}
}

type AdminRoleDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AdminRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_role"
}

func (d *AdminRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AdminRoleModel{}.GetDataSourceSchema()
}

func (d *AdminRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AdminRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AdminRoleModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var adminRoleNameOrId string

	if data.Id.ValueString() != "" {
		adminRoleNameOrId = data.Id.ValueString()
	} else if data.Name.ValueString() != "" {
		adminRoleNameOrId = data.Name.ValueString()
	}

	adminRole, err := getAdminRole(ctx, d.client, &resp.Diagnostics, adminRoleNameOrId)

	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, adminRole)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
