// Copyright © 2024. Citrix Systems, Inc.
package image_definition

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &ImageDefinitionDataSource{}
	_ datasource.DataSourceWithConfigure = &ImageDefinitionDataSource{}
)

func NewImageDefinitionDataSource() datasource.DataSource {
	return &ImageDefinitionDataSource{}
}

type ImageDefinitionDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ImageDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_definition"
}

func (d *ImageDefinitionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ImageDefinitionModel{}.GetDataSourceSchema()
}

func (d *ImageDefinitionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ImageDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	util.CheckProductVersion(d.client, &resp.Diagnostics, 121, 118, 7, 41, "Error reading Image Definition data source", "Image Definition data source")

	var data ImageDefinitionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	var imageDefinitionNameOrId string
	if data.Id.ValueString() != "" {
		imageDefinitionNameOrId = data.Id.ValueString()
	}
	if data.Name.ValueString() != "" {
		imageDefinitionNameOrId = data.Name.ValueString()
	}

	imageDefinition, err := GetImageDefinition(ctx, d.client, &resp.Diagnostics, imageDefinitionNameOrId)
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, imageDefinition)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
