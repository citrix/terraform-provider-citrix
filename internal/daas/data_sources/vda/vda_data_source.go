// Copyright © 2023. Citrix Systems, Inc.

package vda

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

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &VdaDataSource{}
)

func NewVdaDataSource() datasource.DataSource {
	return &VdaDataSource{}
}

// VdaDataSource defines the data source implementation.
type VdaDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *VdaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daas_vda"
}

func (d *VdaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data source for the list of VDAs that belong to either a machine catalog or a delivery group. Machine catalog and delivery group cannot be specified at the same time.",

		Attributes: map[string]schema.Attribute{
			"machine_catalog": schema.StringAttribute{
				MarkdownDescription: "The machine catalog which the VDAs are associated with.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("machine_catalog"), path.MatchRoot("delivery_group")), // Ensures that only one of either machine_catalog or delivery_group is provided. It will also cause a validation error if none are specified.,
				},
			},
			"delivery_group": schema.StringAttribute{
				MarkdownDescription: "The delivery group which the VDAs are associated with.",
				Optional:            true,
			},
			"vdas": schema.ListNestedAttribute{
				Description: "The VDAs associated with the specified machine catalog or delivery group.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"machine_name": schema.StringAttribute{
							Description: "Machine name of the VDA.",
							Computed:    true,
						},
						"hosted_machine_id": schema.StringAttribute{
							Description: "Machine ID within the hypervisor hosting unit.",
							Computed:    true,
						},
						"associated_machine_catalog": schema.StringAttribute{
							Description: "Machine catalog which the VDA is associated with.",
							Computed:    true,
						},
						"associated_delivery_group": schema.StringAttribute{
							Description: "Delivery group which the VDA is associated with.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *VdaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *VdaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VdaDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed machine catalog state from Orchestration
	machineCatalogId := data.MachineCatalog.ValueString()
	if machineCatalogId != "" {
		getMachineCatalogMachinesRequest := d.client.ApiClient.MachineCatalogsAPIsDAAS.MachineCatalogsGetMachineCatalogMachines(ctx, machineCatalogId)
		machineCatalogVdas, httpResp, err := citrixdaasclient.AddRequestData(getMachineCatalogMachinesRequest, d.client).Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing Machine Catalog VDAs"+machineCatalogId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}

		data = data.RefreshPropertyValues(machineCatalogVdas)
	}

	deliveryGroupId := data.DeliveryGroup.ValueString()
	if deliveryGroupId != "" {
		getDeliveryGroupMachinesRequest := d.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroupMachines(ctx, deliveryGroupId)
		deliveryGroupVdas, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupMachinesRequest, d.client).Execute()

		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing Delivery Group VDAs"+deliveryGroupId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}

		data = data.RefreshPropertyValues(deliveryGroupVdas)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
