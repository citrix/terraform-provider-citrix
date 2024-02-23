// Copyright © 2023. Citrix Systems, Inc.

package application

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApplicationFolderDetailsDataSourceModel struct {
	Path              types.String               `tfsdk:"path"`
	TotalApplications types.Int64                `tfsdk:"total_applications"`
	ApplicationsList  []ApplicationResourceModel `tfsdk:"applications_list"`
}

func (r ApplicationFolderDetailsDataSourceModel) RefreshPropertyValues(apps *citrixorchestration.ApplicationResponseModelCollection) ApplicationFolderDetailsDataSourceModel {

	var res []ApplicationResourceModel
	for _, app := range apps.GetItems() {
		res = append(res, ApplicationResourceModel{
			Name:                   types.StringValue(app.GetName()),
			PublishedName:          types.StringValue(app.GetPublishedName()),
			Description:            types.StringValue(app.GetDescription()),
			ApplicationFolderPath:  types.StringValue(*app.GetApplicationFolder().Name.Get()),
			InstalledAppProperties: r.getInstalledAppProperties(app), // Fix: Change the type to *InstalledAppResponseModel
			DeliveryGroups:         r.getDeliveryGroups(app),
		})
	}

	r.ApplicationsList = res
	r.TotalApplications = types.Int64Value(int64(*apps.TotalItems.Get()))
	return r
}

func (r ApplicationFolderDetailsDataSourceModel) getInstalledAppProperties(app citrixorchestration.ApplicationResponseModel) *InstalledAppResponseModel {
	return &InstalledAppResponseModel{
		CommandLineArguments:  types.StringValue(app.GetInstalledAppProperties().CommandLineArguments),
		CommandLineExecutable: types.StringValue(app.GetInstalledAppProperties().CommandLineExecutable),
		WorkingDirectory:      types.StringValue(app.GetInstalledAppProperties().WorkingDirectory),
	}
}

func (r ApplicationFolderDetailsDataSourceModel) getDeliveryGroups(app citrixorchestration.ApplicationResponseModel) []types.String {
	var res []types.String
	for _, dg := range app.AssociatedDeliveryGroupUuids {
		res = append(res, types.StringValue(dg))
	}
	return res
}
