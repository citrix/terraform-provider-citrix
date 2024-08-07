// Copyright © 2024. Citrix Systems, Inc.
package stf_roaming

import (
	"context"
	"strconv"

	citrixstorefrontModels "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type STFRoamingServiceDataSourceModel struct {
	SiteId            types.String `tfsdk:"site_id"`
	Name              types.String `tfsdk:"name"`
	FriendlyName      types.String `tfsdk:"friendly_name"`
	VirtualPath       types.String `tfsdk:"virtual_path"`
	FeatureInstanceId types.String `tfsdk:"feature_instance_id"`
	ConfigurationFile types.String `tfsdk:"configuration_file"`
	TenantId          types.String `tfsdk:"tenant_id"`
	PhysicalPath      types.String `tfsdk:"physical_path"`
}

func (r STFRoamingServiceDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, roamingService *citrixstorefrontModels.STFRoamingServiceResponseModel) STFRoamingServiceDataSourceModel {
	if roamingService.SiteId.IsSet() {
		r.SiteId = types.StringValue(strconv.Itoa(*roamingService.SiteId.Get()))
	}
	if roamingService.Name.IsSet() {
		r.Name = types.StringValue(*roamingService.Name.Get())
	}
	if roamingService.FriendlyName.IsSet() {
		r.FriendlyName = types.StringValue(*roamingService.FriendlyName.Get())
	}
	if roamingService.VirtualPath.IsSet() {
		r.VirtualPath = types.StringValue(*roamingService.VirtualPath.Get())
	}
	if roamingService.FeatureInstanceId.IsSet() {
		r.FeatureInstanceId = types.StringValue(*roamingService.FeatureInstanceId.Get())
	}
	if roamingService.ConfigurationFile.IsSet() {
		r.ConfigurationFile = types.StringValue(*roamingService.ConfigurationFile.Get())
	}
	if roamingService.TenantId.IsSet() {
		r.TenantId = types.StringValue(*roamingService.TenantId.Get())
	}
	if roamingService.PhysicalPath.IsSet() {
		r.PhysicalPath = types.StringValue(*roamingService.PhysicalPath.Get())
	}

	return r
}

func GetSTFRoamingServiceDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StoreFront --- Data source to get details regarding a specific StoreFront Roaming Service instance.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				MarkdownDescription: "ID of the site where the StoreFront Roaming Service instance is created.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the StoreFront Roaming Service instance.",
				Computed:            true,
			},
			"friendly_name": schema.StringAttribute{
				MarkdownDescription: "Friendly name of the StoreFront Roaming Service instance.",
				Computed:            true,
			},
			"virtual_path": schema.StringAttribute{
				MarkdownDescription: "Virtual path of the StoreFront Roaming Service instance.",
				Computed:            true,
			},
			"feature_instance_id": schema.StringAttribute{
				MarkdownDescription: "ID of the StoreFront Roaming Service feature instance.",
				Computed:            true,
			},
			"configuration_file": schema.StringAttribute{
				MarkdownDescription: "Path of the configuration file of the StoreFront Roaming Service instance.",
				Computed:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "ID of the tenant to which the StoreFront Roaming Service instance belongs.",
				Computed:            true,
			},
			"physical_path": schema.StringAttribute{
				MarkdownDescription: "Physical path of the StoreFront Roaming Service instance.",
				Computed:            true,
			},
		},
	}
}
