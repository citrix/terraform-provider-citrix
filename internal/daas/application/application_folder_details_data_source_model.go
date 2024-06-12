// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ApplicationFolderDetailsDataSourceModel struct {
	Path              types.String               `tfsdk:"path"`
	TotalApplications types.Int64                `tfsdk:"total_applications"`
	ApplicationsList  []ApplicationResourceModel `tfsdk:"applications_list"`
}

func (r ApplicationFolderDetailsDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, apps *citrixorchestration.ApplicationResponseModelCollection) ApplicationFolderDetailsDataSourceModel {

	var res []ApplicationResourceModel
	for _, app := range apps.GetItems() {
		res = append(res, ApplicationResourceModel{
			Name:                   types.StringValue(app.GetName()),
			PublishedName:          types.StringValue(app.GetPublishedName()),
			Description:            types.StringValue(app.GetDescription()),
			ApplicationFolderPath:  types.StringValue(*app.GetApplicationFolder().Name.Get()),
			InstalledAppProperties: r.getInstalledAppProperties(ctx, diagnostics, app),
			DeliveryGroups:         r.getDeliveryGroups(ctx, diagnostics, app),
		})
	}

	r.ApplicationsList = res
	r.TotalApplications = types.Int64Value(int64(*apps.TotalItems.Get()))
	return r
}

func (r ApplicationFolderDetailsDataSourceModel) getInstalledAppProperties(ctx context.Context, diagnostics *diag.Diagnostics, app citrixorchestration.ApplicationResponseModel) types.Object {
	var installedAppResponse = InstalledAppResponseModel{
		CommandLineArguments:  types.StringValue(app.GetInstalledAppProperties().CommandLineArguments),
		CommandLineExecutable: types.StringValue(app.GetInstalledAppProperties().CommandLineExecutable),
		WorkingDirectory:      types.StringValue(app.GetInstalledAppProperties().WorkingDirectory),
	}
	return util.TypedObjectToObjectValue(ctx, diagnostics, installedAppResponse)
}

func (r ApplicationFolderDetailsDataSourceModel) getDeliveryGroups(ctx context.Context, diagnostics *diag.Diagnostics, app citrixorchestration.ApplicationResponseModel) types.Set {
	return util.StringArrayToStringSet(ctx, diagnostics, app.AssociatedDeliveryGroupUuids)
}
