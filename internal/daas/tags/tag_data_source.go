// Copyright © 2024. Citrix Systems, Inc.
package tags

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &TagDataSource{}
)

func NewTagDataSource() datasource.DataSource {
	return &TagDataSource{}
}

type TagDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *TagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *TagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = TagDataSourceModel{}.GetSchema()
}

func (d *TagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *TagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	// Read Terraform configuration data into the model
	var data TagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var tagNameOrId string

	if data.Id.ValueString() != "" {
		tagNameOrId = data.Id.ValueString()
	}
	if data.Name.ValueString() != "" {
		tagNameOrId = data.Name.ValueString()
	}

	getTagRequest := d.client.ApiClient.TagsAPIsDAAS.TagsGetTag(ctx, tagNameOrId)
	tagDetailResponse, httpResp, err := citrixdaasclient.AddRequestData(getTagRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching details of tag: "+tagNameOrId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, tagDetailResponse)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
