// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (OktaIdentityProviderModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Data source of a Citrix Cloud Okta Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Citrix Cloud Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Citrix Cloud Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"okta_domain": schema.StringAttribute{
				Description: "Okta domain name for configuring Okta Identity Provider.",
				Computed:    true,
			},
			"okta_client_id": schema.StringAttribute{
				Description: "ID of the Okta client for configuring Okta Identity Provider.",
				Computed:    true,
			},
			"okta_client_secret": schema.StringAttribute{
				Description: "Secret of the Okta client for configuring Okta Identity Provider.",
				Computed:    true,
			},
			"okta_api_token": schema.StringAttribute{
				Description: "Okta API token for configuring Okta Identity Provider.",
				Computed:    true,
			},
		},
	}
}

func (OktaIdentityProviderModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return OktaIdentityProviderModel{}.GetDataSourceSchema().Attributes
}
