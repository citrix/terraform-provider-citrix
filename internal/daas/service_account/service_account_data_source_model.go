// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ServiceAccountDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	AccountId   types.String `tfsdk:"account_id"`
}

func (ServiceAccountDataSourceModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source to get details for a service account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the service account.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name of the service account.",
				Computed:    true,
			},
			"account_id": schema.StringAttribute{
				Description: "The account ID of the service account." +
					"\n\n -> **Note** For Active Directory, this is the username. Username should be in `domain\\username` format. For AzureAD, this is the application ID. The account ID must be in lowercase.",
				Required: true,
				Validators: []validator.String{
					// case sensitive
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "the account_id must be all lowercase"),
				},
			},
		},
	}
}

func (r ServiceAccountDataSourceModel) RefreshPropertyValues(ctx context.Context, serviceAccount citrixorchestration.ServiceAccountResponseModel) ServiceAccountDataSourceModel {
	r.Id = types.StringValue(serviceAccount.GetServiceAccountUid())
	r.DisplayName = types.StringValue(serviceAccount.GetDisplayName())
	r.AccountId = types.StringValue(strings.ToLower(serviceAccount.GetAccountId()))

	return r
}
