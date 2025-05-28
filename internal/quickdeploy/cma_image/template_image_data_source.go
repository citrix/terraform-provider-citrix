// Copyright © 2024. Citrix Systems, Inc.
package cma_image

import (
	"context"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &CitrixManagedAzureImageDataSource{}
	_ datasource.DataSourceWithConfigure = &CitrixManagedAzureImageDataSource{}
)

func NewCitrixManagedAzureImageDataSource() datasource.DataSource {
	return &CitrixManagedAzureImageDataSource{}
}

// CitrixManagedAzureImageDataSource is the data source implementation.
type CitrixManagedAzureImageDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *CitrixManagedAzureImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickdeploy_template_image"
}

func (d *CitrixManagedAzureImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CitrixManagedAzureImageDataSourceModel{}.GetDataSourceSchema()
}

func (d *CitrixManagedAzureImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *CitrixManagedAzureImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CitrixManagedAzureImageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get image Id from Name
	getTemplateImagesRequest := d.client.QuickDeployClient.MasterImageCMD.GetImages(ctx, d.client.ClientConfig.CustomerId, d.client.ClientConfig.SiteId)
	images, _, err := citrixdaasclient.AddRequestData(getTemplateImagesRequest, d.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Error getting Citrix Managed Azure Template Images", err.Error())
		return
	}
	id := ""
	for _, image := range images.GetItems() {
		if strings.EqualFold(image.Name, data.Name.ValueString()) {
			id = image.Id
			break
		}
	}
	if id == "" {
		resp.Diagnostics.AddError("Error getting Citrix Managed Azure Template Image", "Image with name "+data.Name.ValueString()+" not found")
		return
	}

	// Try getting the Citirx Managed Azure Template Image
	image, _, err := getTemplateImageWithId(ctx, d.client, &resp.Diagnostics, id, true)
	if err != nil {
		// Remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, image)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
