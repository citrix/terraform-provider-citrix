// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PvsDataSourceModel struct {
	FarmName  types.String `tfsdk:"pvs_farm_name"`
	SiteId    types.String `tfsdk:"pvs_site_id"`
	SiteName  types.String `tfsdk:"pvs_site_name"`
	StoreName types.String `tfsdk:"pvs_store_name"`
	VdiskId   types.String `tfsdk:"pvs_vdisk_id"`
	VdiskName types.String `tfsdk:"pvs_vdisk_name"`
}

func (r PvsDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, pvsSiteId string, pvsVdiskId string) PvsDataSourceModel {

	r.SiteId = types.StringValue(pvsSiteId)
	r.VdiskId = types.StringValue(pvsVdiskId)

	return r
}

func (PvsDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "PVS Configuration to create machine catalog using PVSStreaming.",
		Attributes: map[string]schema.Attribute{
			"pvs_farm_name": schema.StringAttribute{
				Description: "Name of the PVS farm.",
				Required:    true,
			},
			"pvs_site_id": schema.StringAttribute{
				Description: "Id of the PVS site.",
				Computed:    true,
			},
			"pvs_site_name": schema.StringAttribute{
				Description: "Name of the PVS site.",
				Required:    true,
			},
			"pvs_store_name": schema.StringAttribute{
				Description: "Name of the PVS store.",
				Required:    true,
			},
			"pvs_vdisk_id": schema.StringAttribute{
				Description: "Id of the PVS vDisk.",
				Computed:    true,
			},
			"pvs_vdisk_name": schema.StringAttribute{
				Description: "Name of the PVS vDisk.",
				Required:    true,
			},
		},
	}
}
