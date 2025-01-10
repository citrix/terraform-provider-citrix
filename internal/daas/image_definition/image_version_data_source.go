// Copyright © 2024. Citrix Systems, Inc.
package image_definition

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource = &ImageVersionDataSource{}
)

func NewImageVersionDataSource() datasource.DataSource {
	return &ImageVersionDataSource{}
}

type ImageVersionDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ImageVersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_version"
}

func (d *ImageVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ImageVersionModel{}.GetDataSourceSchema()
}

func (d *ImageVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ImageVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	util.CheckProductVersion(d.client, &resp.Diagnostics, 121, 118, 7, 41, "Error reading Image Version data source", "Image Version data source")

	var data ImageVersionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the data from the API
	useImageVersionId := true
	var imageVersionId string
	var imageVersionNumber int32

	imageDefinitionId := data.ImageDefinition.ValueString()

	if !data.Id.IsNull() {
		imageVersionId = data.Id.ValueString()
	}
	if !data.VersionNumber.IsNull() {
		useImageVersionId = false
		imageVersionNumber = data.VersionNumber.ValueInt32()
	}

	var imageVersion *citrixorchestration.ImageVersionResponseModel
	var err error
	if useImageVersionId {
		imageVersion, err = GetImageVersion(ctx, d.client, &resp.Diagnostics, imageDefinitionId, imageVersionId)
		if err != nil {
			return
		}
	} else {
		imageVersions, err := getImageVersions(ctx, &resp.Diagnostics, d.client, imageDefinitionId)
		if err != nil {
			return
		}
		for _, imageVersionInRemote := range imageVersions {
			if imageVersionInRemote.GetNumber() == imageVersionNumber {
				imageVersion = &imageVersionInRemote
				break
			}
		}
	}

	data = data.RefreshDataSourcePropertyValues(ctx, &resp.Diagnostics, imageVersion)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
