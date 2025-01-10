// Copyright © 2024. Citrix Systems, Inc.
package admin_role

import (
	"context"
	"sync"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ datasource.DataSource = &AdminPermissionsDataSource{}
)

func NewAdminPermissionsDataSource() datasource.DataSource {
	return &AdminPermissionsDataSource{}
}

type AdminPermissionsDataSource struct {
	client *citrixdaasclient.CitrixDaasClient
}

func (d *AdminPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_permissions"
}

func (d *AdminPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AdminPermissionsDataSourceModel{}.GetDataSourceSchema()
}

func (d *AdminPermissionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

func (d *AdminPermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if d.client != nil && d.client.ApiClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}

	var data AdminPermissionsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	permissions, err := getAdminPermissions(ctx, d.client, &resp.Diagnostics, true)
	if err != nil {
		return
	}

	data = data.RefreshPropertyValues(permissions)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Cache the admin predefined permissions as they do not change
var adminPermissionsCache []citrixorchestration.PredefinedPermissionResponseModel
var cacheLoad sync.Once
var cacheLoadError error

func getAdminPermissions(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, filterCloudRestrictedPermissions bool) ([]citrixorchestration.PredefinedPermissionResponseModel, error) {
	cacheLoad.Do(func() {
		getPermissionsRequest := client.ApiClient.AdminAPIsDAAS.AdminGetPredefinedPermissions(ctx)
		permissions := []citrixorchestration.PredefinedPermissionResponseModel{}
		continuationToken := ""
		for {
			getPermissionsRequest = getPermissionsRequest.ContinuationToken(continuationToken)
			resp, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.PredefinedPermissionResponseModelCollection](getPermissionsRequest, client)
			if err != nil {
				diagnostics.AddError(
					"Error reading predefined admin permissions",
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				cacheLoadError = err
				return
			}

			permissions = append(permissions, resp.GetItems()...)
			if resp.GetContinuationToken() == "" {
				adminPermissionsCache = append(adminPermissionsCache, permissions...)
				return
			}
			continuationToken = resp.GetContinuationToken()
		}
	})

	if cacheLoadError != nil {
		return nil, cacheLoadError
	}

	if filterCloudRestrictedPermissions {
		filteredPermissions := []citrixorchestration.PredefinedPermissionResponseModel{}
		for _, permission := range adminPermissionsCache {
			if _, ok := util.RestrictedPermissionsInCloud[permission.GetId()]; !ok {
				filteredPermissions = append(filteredPermissions, permission)
			}
		}
		return filteredPermissions, nil
	} else {
		return adminPermissionsCache, nil
	}
}
