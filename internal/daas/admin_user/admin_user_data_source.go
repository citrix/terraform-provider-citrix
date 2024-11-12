// Copyright © 2024. Citrix Systems, Inc.

package admin_user

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &AdminUserDataSource{}
)

func NewAdminUserDataSource() datasource.DataSource {
	return &AdminUserDataSource{}
}

type AdminUserDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AdminUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_user"
}

func (d *AdminUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AdminUserResourceModel{}.GetDataSourceSchema()
}

func (d *AdminUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AdminUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	if !d.client.AuthConfig.OnPremises {
		resp.Diagnostics.AddError("Environment Not Supported", "This terraform resource is only supported for on-premise deployments")
	}

	var data AdminUserResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	adminUser, err := getAdminUser(ctx, d.client, &resp.Diagnostics, data.Id.ValueString())
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, adminUser)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
