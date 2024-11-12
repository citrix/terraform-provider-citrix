// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (SamlAttributeNameMappings) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the attribute mappings for SAML 2.0 Identity Provider.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"user_display_name": schema.StringAttribute{
				Description: "The attribute name for the user's display name.",
				Computed:    true,
			},
			"user_given_name": schema.StringAttribute{
				Description: "The attribute name for the user's given name.",
				Computed:    true,
			},
			"user_family_name": schema.StringAttribute{
				Description: "The attribute name for the user's family name.",
				Computed:    true,
			},
			"security_identifier": schema.StringAttribute{
				Description: "The attribute name for the user's security identifier.",
				Computed:    true,
			},
			"user_principal_name": schema.StringAttribute{
				Description: "The attribute name for the user's principal name.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The attribute name for the user's email.",
				Computed:    true,
			},
			"ad_object_identifier": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory object identifier.",
				Computed:    true,
			},
			"ad_forest": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory forest.",
				Computed:    true,
			},
			"ad_domain": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory domain.",
				Computed:    true,
			},
		},
	}
}

func (SamlAttributeNameMappings) GetDataSourceAttributes() map[string]schema.Attribute {
	return SamlAttributeNameMappings{}.GetDataSourceSchema().Attributes
}

func (SamlIdentityProviderModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Data Source of a SAML 2.0 Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the SAML 2.0 Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the SAML 2.0 Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_domain_name": schema.StringAttribute{
				Description: "Auth Domain name of the SAML 2.0 Identity Provider instance.",
				Computed:    true,
			},
			"entity_id": schema.StringAttribute{
				Description: "The entity ID of the SAML 2.0 Identity Provider.",
				Computed:    true,
			},
			"use_scoped_entity_id": schema.BoolAttribute{
				Description: "Whether to use the Scoped Entity Id.",
				Computed:    true,
			},
			"sign_auth_request": schema.StringAttribute{
				Description: "Whether to sign the authentication request.",
				Computed:    true,
			},
			"single_sign_on_service_url": schema.StringAttribute{
				Description: "The URL of the single sign-on service.",
				Computed:    true,
			},
			"single_sign_on_service_binding": schema.StringAttribute{
				Description: "The binding of the single sign-on service.",
				Computed:    true,
			},
			"saml_response": schema.StringAttribute{
				Description: "The SAML response.",
				Computed:    true,
			},
			"cert_file_path": schema.StringAttribute{
				Description: "The file path of the certificate.",
				Computed:    true,
			},
			"authentication_context": schema.StringAttribute{
				Description: "The authentication context.",
				Computed:    true,
			},
			"authentication_context_comparison": schema.StringAttribute{
				Description: "The authentication context comparison type.",
				Computed:    true,
			},
			"logout_url": schema.StringAttribute{
				Description: "The URL of the logout service.",
				Computed:    true,
			},
			"sign_logout_request": schema.StringAttribute{
				Description: "Whether to sign the logout request.",
				Computed:    true,
			},
			"logout_binding": schema.StringAttribute{
				Description: "The binding of the logout service.",
				Computed:    true,
			},
			"attribute_names": SamlAttributeNameMappings{}.GetDataSourceSchema(),
			"cert_common_name": schema.StringAttribute{
				Description: "The common name of the SAML certificate.",
				Computed:    true,
			},
			"cert_expiration": schema.StringAttribute{
				Description: "The expiration date time of the SAML certificate.",
				Computed:    true,
			},
			"scoped_entity_id_suffix": schema.StringAttribute{
				Description: "The Scoped Entity Id Suffix for the SAML 2.0 Identity Provider.",
				Computed:    true,
			},
		},
	}
}

func (SamlIdentityProviderModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return SamlIdentityProviderModel{}.GetDataSourceSchema().Attributes
}
