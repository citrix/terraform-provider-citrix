// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OktaIdentityProviderDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	OktaDomain types.String `tfsdk:"okta_domain"`
}

func (OktaIdentityProviderDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Data source of a Citrix Cloud Okta Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Citrix Cloud Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Citrix Cloud Identity Provider instance.",
				Optional:    true,
			},
			"okta_domain": schema.StringAttribute{
				Description: "Okta domain name for configuring Okta Identity Provider.",
				Computed:    true,
			},
		},
	}
}

func (OktaIdentityProviderDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return OktaIdentityProviderDataSourceModel{}.GetSchema().Attributes
}

func (r OktaIdentityProviderDataSourceModel) RefreshPropertyValues(oktaIdp *citrixcws.IdpStatusModel) OktaIdentityProviderDataSourceModel {

	// Overwrite Okta Identity Provider Data Source with refreshed state
	r.Id = types.StringValue(oktaIdp.GetIdpInstanceId())
	r.Name = types.StringValue(oktaIdp.GetIdpNickname())

	additionalInfo := oktaIdp.GetAdditionalStatusInfo()
	if additionalInfo != nil {
		r.OktaDomain = types.StringValue(strings.ReplaceAll(additionalInfo["oktaDomain"], "https://", ""))
	}

	return r
}
