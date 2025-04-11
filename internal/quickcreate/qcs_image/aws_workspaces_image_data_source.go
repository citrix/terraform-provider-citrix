// Copyright © 2024. Citrix Systems, Inc.
package qcs_image

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ datasource.DataSource              = &AwsWorkspacesImageDataSource{}
	_ datasource.DataSourceWithConfigure = &AwsWorkspacesImageDataSource{}
)

func NewAwsWorkspacesImageDataSource() datasource.DataSource {
	return &AwsWorkspacesImageDataSource{}
}

// AwsWorkspacesImageDataSource is the data source implementation.
type AwsWorkspacesImageDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AwsWorkspacesImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickcreate_aws_workspaces_image"
}

func (d *AwsWorkspacesImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AwsWorkspacesImageModel{}.GetDataSourceSchema()
}

func (d *AwsWorkspacesImageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AwsWorkspacesImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AwsWorkspacesImageModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try getting the AWS WorkSpaces Image
	var image *citrixquickcreate.AwsEdcImage
	var err error
	if !data.Id.IsNull() {
		image, _, err = getAwsWorkspacesImageWithId(ctx, d.client, &resp.Diagnostics, data.AccountId.ValueString(), data.Id.ValueString(), false)
	} else {
		image, _, err = getAwsWorkspacesImageWithName(ctx, d.client, &resp.Diagnostics, data.AccountId.ValueString(), data.Name.ValueString())
	}

	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	data = data.RefreshPropertyValues(ctx, &resp.Diagnostics, false, image)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getAwsWorkspacesImageWithName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, accountId string, imageName string) (*citrixquickcreate.AwsEdcImage, *http.Response, error) {
	getImageRequest := client.QuickCreateClient.ImageQCS.GetImagesAsync(ctx, client.ClientConfig.CustomerId, accountId)
	images, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.Images](getImageRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error getting AWS WorkSpaces Image: "+imageName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadQcsClientError(err),
		)
		return nil, httpResp, err
	}

	for _, image := range images.GetItems() {
		if strings.EqualFold(image.GetName(), imageName) {
			return &image, httpResp, nil
		}
	}

	return nil, httpResp, fmt.Errorf("AWS WorkSpaces Image not found: %s", imageName)
}
