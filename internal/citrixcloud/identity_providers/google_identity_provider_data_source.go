// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &GoogleIdentityProviderDataSource{}
)

func NewGoogleIdentityProviderDataSource() datasource.DataSource {
	return &GoogleIdentityProviderDataSource{}
}

type GoogleIdentityProviderDataSource struct {
	client  *citrixdaasclient.CitrixDaasClient
	idpType string
}

func (d *GoogleIdentityProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_google_identity_provider"
}

func (d *GoogleIdentityProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = GoogleIdentityProviderDataSourceModel{}.GetSchema()
}

func (d *GoogleIdentityProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
	d.idpType = string(citrixcws.CWSIDENTITYPROVIDERTYPE_GOOGLE)
}

func (d *GoogleIdentityProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data GoogleIdentityProviderDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var idpStatus *citrixcws.IdpStatusModel
	var err error
	if data.Id.ValueString() != "" {
		idpStatus, err = getIdentityProviderById(ctx, d.client, &resp.Diagnostics, d.idpType, data.Id.ValueString())
	} else if data.Name.ValueString() != "" {
		idpStatus, err = getIdentityProviderByName(ctx, d.client, &resp.Diagnostics, d.idpType, data.Name.ValueString())
	}

	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(idpStatus)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
