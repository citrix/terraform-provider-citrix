// Copyright © 2023. Citrix Systems, Inc.

package admin_scope

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source to get details regarding a specific Administrator scope.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Admin Scope.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Admin Scope.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Admin Scope.",
				Computed:            true,
			},
			"is_built_in": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the Admin Scope is built-in or not.",
				Computed:            true,
			},
			"is_all_scope": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the Admin Scope is all scope or not.",
				Computed:            true,
			},
			"is_tenant_scope": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the Admin Scope is tenant scope or not.",
				Computed:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "ID of the tenant to which the Admin Scope belongs.",
				Computed:            true,
			},
			"tenant_name": schema.StringAttribute{
				MarkdownDescription: "Name of the tenant to which the Admin Scope belongs.",
				Computed:            true,
			},
		},
	}
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

	data = data.RefreshPropertyValues(adminScope)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
