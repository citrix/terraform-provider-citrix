// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"github.com/citrix/citrix-daas-rest-go/citrixcws"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GoogleIdentityProviderDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	AuthDomainName   types.String `tfsdk:"auth_domain_name"`
	GoogleCustomerId types.String `tfsdk:"google_customer_id"`
	GoogleDomain     types.String `tfsdk:"google_domain"`
}

func (GoogleIdentityProviderDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Data Source of a Citrix Cloud Google Cloud Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Citrix Cloud Google Cloud Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Citrix Cloud Google Cloud Identity Provider instance.",
				Optional:    true,
			},
			"auth_domain_name": schema.StringAttribute{
				Description: "User authentication domain name for Google Cloud Identity Provider.",
				Computed:    true,
			},
			"google_customer_id": schema.StringAttribute{
				Description: "Customer ID of the configured  Google Cloud Identity Provider.",
				Computed:    true,
			},
			"google_domain": schema.StringAttribute{
				Description: "Domain of the configured Google Cloud Identity Provider.",
				Computed:    true,
			},
		},
	}
}

func (GoogleIdentityProviderDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return GoogleIdentityProviderDataSourceModel{}.GetSchema().Attributes
}

func (r GoogleIdentityProviderDataSourceModel) RefreshPropertyValues(googleIdp *citrixcws.IdpStatusModel) GoogleIdentityProviderDataSourceModel {

	// Overwrite resource location with refreshed state
	r.Id = types.StringValue(googleIdp.GetIdpInstanceId())
	r.Name = types.StringValue(googleIdp.GetIdpNickname())
	r.AuthDomainName = types.StringValue(googleIdp.GetAuthDomainName())

	additionalInfo := googleIdp.GetAdditionalStatusInfo()
	if additionalInfo != nil {
		r.GoogleCustomerId = types.StringValue(additionalInfo["googleCustomerId"])
		r.GoogleDomain = types.StringValue(additionalInfo["googleDomain"])
	}

	return r
}
