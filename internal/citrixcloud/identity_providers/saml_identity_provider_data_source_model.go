// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SamlAttributeNameMappingsDataSourceModel struct {
	UserDisplayName    types.String `tfsdk:"user_display_name"`
	UserGivenName      types.String `tfsdk:"user_given_name"`
	UserFamilyName     types.String `tfsdk:"user_family_name"`
	SecurityIdentifier types.String `tfsdk:"security_identifier"`
	UserPrincipalName  types.String `tfsdk:"user_principal_name"`
	Email              types.String `tfsdk:"email"`
	AdObjectIdentifier types.String `tfsdk:"ad_object_identifier"`
	AdForest           types.String `tfsdk:"ad_forest"`
	AdDomain           types.String `tfsdk:"ad_domain"`
}

func (SamlAttributeNameMappingsDataSourceModel) GetSchema() schema.SingleNestedAttribute {
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

func (SamlAttributeNameMappingsDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return SamlAttributeNameMappingsDataSourceModel{}.GetSchema().Attributes
}

type SamlIdentityProviderDataSourceModel struct {
	Id                              types.String `tfsdk:"id"`
	Name                            types.String `tfsdk:"name"`
	AuthDomainName                  types.String `tfsdk:"auth_domain_name"`
	EntityId                        types.String `tfsdk:"entity_id"`
	UseScopedEntityId               types.Bool   `tfsdk:"use_scoped_entity_id"`
	SignAuthRequest                 types.String `tfsdk:"sign_auth_request"`
	SingleSignOnServiceUrl          types.String `tfsdk:"single_sign_on_service_url"`
	SingleSignOnServiceBinding      types.String `tfsdk:"single_sign_on_service_binding"`
	SamlResponse                    types.String `tfsdk:"saml_response"`
	AuthenticationContext           types.String `tfsdk:"authentication_context"`
	AuthenticationContextComparison types.String `tfsdk:"authentication_context_comparison"`
	LogoutUrl                       types.String `tfsdk:"logout_url"`
	SignLogoutRequest               types.String `tfsdk:"sign_logout_request"`
	LogoutBinding                   types.String `tfsdk:"logout_binding"`
	AttributeNames                  types.Object `tfsdk:"attribute_names"` // SamlAttributeNameMappingsDataSourceModel
	CertCommonName                  types.String `tfsdk:"cert_common_name"`
	CertExpiration                  types.String `tfsdk:"cert_expiration"`
	ScopedEntityIdSuffix            types.String `tfsdk:"scoped_entity_id_suffix"`
}

func (SamlIdentityProviderDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Data Source of a SAML 2.0 Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the SAML 2.0 Identity Provider instance.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the SAML 2.0 Identity Provider instance.",
				Optional:    true,
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
			"attribute_names": SamlAttributeNameMappingsDataSourceModel{}.GetSchema(),
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

func (SamlIdentityProviderDataSourceModel) GetResourceAttributes() map[string]schema.Attribute {
	return SamlIdentityProviderDataSourceModel{}.GetSchema().Attributes
}

func (r SamlIdentityProviderDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, samlIdp *citrixcws.IdpStatusModel, samlConfig *citrixcws.SamlConfigModel) SamlIdentityProviderDataSourceModel {

	// Overwrite SAML 2.0 Identity Provider Data Source with refreshed state
	r.Id = types.StringValue(samlIdp.GetIdpInstanceId())
	r.Name = types.StringValue(samlIdp.GetIdpNickname())
	r.AuthDomainName = types.StringValue(samlIdp.GetAuthDomainName())

	r.EntityId = types.StringValue(samlConfig.GetSamlEntityId())

	useScopedEntityId := false
	if samlConfig.GetSamlSpEntityIdSuffix() != "" {
		useScopedEntityId = true
	}
	r.UseScopedEntityId = types.BoolValue(useScopedEntityId)

	if samlConfig == nil {
		return r
	}

	r.SignAuthRequest = types.StringValue(samlConfig.GetSamlSignAuthRequest())
	r.SingleSignOnServiceUrl = types.StringValue(samlConfig.GetSamlSingleSignOnServiceUrl())
	r.SingleSignOnServiceBinding = types.StringValue(samlConfig.GetSamlSingleSignOnServiceBinding())
	r.SamlResponse = types.StringValue(samlConfig.GetSamlResponse())
	r.AuthenticationContext = types.StringValue(samlConfig.GetSamlAuthenticationContext())
	r.AuthenticationContextComparison = types.StringValue(samlConfig.GetSamlAuthenticationContextComparison())
	r.LogoutUrl = types.StringValue(samlConfig.GetSamlLogoutUrl())
	if samlConfig.GetSamlLogoutUrl() != "" {
		r.SignLogoutRequest = types.StringValue(samlConfig.GetSamlSignLogoutRequest())
		r.LogoutBinding = types.StringValue(samlConfig.GetSamlLogoutRequestBinding())
	}

	// Refresh Attribute Name Mappings
	attributeNamesAttributesMap := map[string]attr.Type{
		"user_display_name":    types.StringType,
		"user_given_name":      types.StringType,
		"user_family_name":     types.StringType,
		"security_identifier":  types.StringType,
		"user_principal_name":  types.StringType,
		"email":                types.StringType,
		"ad_object_identifier": types.StringType,
		"ad_forest":            types.StringType,
		"ad_domain":            types.StringType,
	}
	updatedAttributeNames := SamlAttributeNameMappingsDataSourceModel{
		UserDisplayName:    types.StringValue(samlConfig.GetSamlAttributeNameForUserDisplayName()),
		UserGivenName:      types.StringValue(samlConfig.GetSamlAttributeNameForUserGivenName()),
		UserFamilyName:     types.StringValue(samlConfig.GetSamlAttributeNameForUserFamilyName()),
		SecurityIdentifier: types.StringValue(samlConfig.GetSamlAttributeNameForSid()),
		UserPrincipalName:  types.StringValue(samlConfig.GetSamlAttributeNameForUpn()),
		Email:              types.StringValue(samlConfig.GetSamlAttributeNameForEmail()),
		AdObjectIdentifier: types.StringValue(samlConfig.GetSamlAttributeNameForAdOid()),
		AdForest:           types.StringValue(samlConfig.GetSamlAttributeNameForAdForest()),
		AdDomain:           types.StringValue(samlConfig.GetSamlAttributeNameForAdDomain()),
	}
	attributeNames, diags := types.ObjectValueFrom(ctx, attributeNamesAttributesMap, updatedAttributeNames)
	if diags != nil {
		diagnostics.Append(diags...)
		attributeNames = types.ObjectUnknown(attributeNamesAttributesMap)
	}

	r.AttributeNames = attributeNames

	// Refresh Computed Values
	r.CertCommonName = types.StringValue(samlConfig.GetSamlCertCN())
	r.CertExpiration = types.StringValue(samlConfig.GetSamlCertExpiration())
	r.ScopedEntityIdSuffix = types.StringValue(samlConfig.GetSamlSpEntityIdSuffix())

	return r
}
