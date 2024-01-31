// Copyright © 2023. Citrix Systems, Inc.

package application_folder_details

import (
	"context"
	"strings"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSourceSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource defines the data source implementation.
type ApplicationDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_application_folder_details"
}

// Schema defines the data source schema.
func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for retrieving details of applications belonging to a specific folder.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "The path of the folder to get the applications from.",
				Required:            true,
			},
			"total_applications": schema.Int64Attribute{
				MarkdownDescription: "The total number of applications in the folder.",
				Computed:            true,
			},
			"applications_list": schema.ListNestedAttribute{
				Description: "The applications list associated with the specified folder.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the application.",
							Computed:            true,
						},
						"published_name": schema.StringAttribute{
							MarkdownDescription: "The published name of the application.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the application.",
							Computed:            true,
						},
						"installed_app_properties": schema.SingleNestedAttribute{
							MarkdownDescription: "The installed application properties of the application.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"command_line_arguments": schema.StringAttribute{
									MarkdownDescription: "The command-line arguments to use when launching the executable. Environment variables can be used.",
									Computed:            true,
								},
								"command_line_executable": schema.StringAttribute{
									MarkdownDescription: "The name of the executable file to launch. The full path need not be provided if it's already in the path. Environment variables can also be used.",
									Computed:            true,
								},
								"working_directory": schema.StringAttribute{
									MarkdownDescription: "The working directory which the executable is launched from. Environment variables can be used.",
									Computed:            true,
								},
							},
						},
						"delivery_groups": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The delivery groups which the application is associated with.",
							Computed:            true,
						},
						"application_folder_path": schema.StringAttribute{
							MarkdownDescription: "The path of the folder which the application belongs to",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationFolderDetailsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the list of applications using the path
	path := data.Path.ValueString()
	if path != "" {
		applicationFolderPath := strings.ReplaceAll(path, "\\", "|")
		getApplicationsRequest := d.client.ApiClient.AdminFoldersAPIsDAAS.AdminFoldersGetAdminFolderApplications(ctx, applicationFolderPath)
		apps, httpResp, err := citrixdaasclient.AddRequestData(getApplicationsRequest, d.client).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting Applications from folder "+path,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
			return // Stop processing
		}
		data = data.RefreshPropertyValues(apps)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
