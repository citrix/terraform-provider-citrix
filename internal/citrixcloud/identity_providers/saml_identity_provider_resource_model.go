// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"context"
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SamlAttributeNameMappings struct {
	UserDisplayName    types.String `tfsdk:"user_display_name"`    // Optional, Computed, Defaults to `displayName`
	UserGivenName      types.String `tfsdk:"user_given_name"`      // Optional, Computed, Defaults to `givenName`
	UserFamilyName     types.String `tfsdk:"user_family_name"`     // Optional, Computed, Defaults to `familyName`
	SecurityIdentifier types.String `tfsdk:"security_identifier"`  // Requires one of `security_identifier` or `user_principal_name`, defaults to `cip_sid`
	UserPrincipalName  types.String `tfsdk:"user_principal_name"`  // Requires one of `security_identifier` or `user_principal_name`, defaults to `cip_upn`
	Email              types.String `tfsdk:"email"`                // Optional, Computed, Defaults to `cip_email`
	AdObjectIdentifier types.String `tfsdk:"ad_object_identifier"` // Optional, Computed, Defaults to `cip_oid`
	AdForest           types.String `tfsdk:"ad_forest"`            // Optional, Computed, Defaults to `cip_forest`
	AdDomain           types.String `tfsdk:"ad_domain"`            // Optional, Computed, Defaults to `cip_domain`
}

func (SamlAttributeNameMappings) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the attribute mappings for SAML 2.0 Identity Provider.",
		Required:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"user_display_name": schema.StringAttribute{
				Description: "The attribute name for the user's display name. Defaults to `displayName`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("displayName"),
			},
			"user_given_name": schema.StringAttribute{
				Description: "The attribute name for the user's given name. Defaults to `givenName`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("givenName"),
			},
			"user_family_name": schema.StringAttribute{
				Description: "The attribute name for the user's family name. Defaults to `familyName`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("familyName"),
			},
			"security_identifier": schema.StringAttribute{
				Description: "The attribute name for the user's security identifier. Defaults to `cip_sid`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_sid"),
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName("security_identifier"),
						path.MatchRelative().AtParent().AtName("user_principal_name"),
					),
				},
			},
			"user_principal_name": schema.StringAttribute{
				Description: "The attribute name for the user's principal name. Defaults to `cip_upn`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_upn"),
			},
			"email": schema.StringAttribute{
				Description: "The attribute name for the user's email. Defaults to `cip_email`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_email"),
			},
			"ad_object_identifier": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory object identifier. Defaults to `cip_oid`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_oid"),
			},
			"ad_forest": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory forest. Defaults to `cip_forest`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_forest"),
			},
			"ad_domain": schema.StringAttribute{
				Description: "The attribute name for the user's Active Directory domain. Defaults to `cip_domain`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("cip_domain"),
			},
		},
	}
}

func (SamlAttributeNameMappings) GetAttributes() map[string]schema.Attribute {
	return SamlAttributeNameMappings{}.GetSchema().Attributes
}

type SamlIdentityProviderModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AuthDomainName types.String `tfsdk:"auth_domain_name"`

	EntityId                        types.String `tfsdk:"entity_id"`                         // Required
	UseScopedEntityId               types.Bool   `tfsdk:"use_scoped_entity_id"`              // Optional, defaults to false
	SignAuthRequest                 types.String `tfsdk:"sign_auth_request"`                 // Required
	SingleSignOnServiceUrl          types.String `tfsdk:"single_sign_on_service_url"`        // Required
	SingleSignOnServiceBinding      types.String `tfsdk:"single_sign_on_service_binding"`    // Required
	SamlResponse                    types.String `tfsdk:"saml_response"`                     // Required
	CertFilePath                    types.String `tfsdk:"cert_file_path"`                    // Required
	AuthenticationContext           types.String `tfsdk:"authentication_context"`            // Required
	AuthenticationContextComparison types.String `tfsdk:"authentication_context_comparison"` // Required
	LogoutUrl                       types.String `tfsdk:"logout_url"`                        // Optional, defaults to empty string
	SignLogoutRequest               types.String `tfsdk:"sign_logout_request"`               // Required if LogoutUrl is provided
	LogoutBinding                   types.String `tfsdk:"logout_binding"`                    // Required if LogoutUrl is provided

	AttributeNames types.Object `tfsdk:"attribute_names"` // Required SamlAttributeNameMappings

	// Computed
	CertCommonName       types.String `tfsdk:"cert_common_name"`
	CertExpiration       types.String `tfsdk:"cert_expiration"`
	ScopedEntityIdSuffix types.String `tfsdk:"scoped_entity_id_suffix"`
}

func (SamlIdentityProviderModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Manages a SAML 2.0 Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the SAML 2.0 Identity Provider instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the SAML 2.0 Identity Provider instance.",
				Required:    true,
			},
			"auth_domain_name": schema.StringAttribute{
				Description: "Auth Domain name of the SAML 2.0 Identity Provider instance.",
				Required:    true,
			},
			"entity_id": schema.StringAttribute{
				Description: "The entity ID of the SAML 2.0 Identity Provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use_scoped_entity_id": schema.BoolAttribute{
				Description: "Whether to use the Scoped Entity Id.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"sign_auth_request": schema.StringAttribute{
				Description: "Whether to sign the authentication request. Valid values are `Yes` and `No`. Defaults to `Yes`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(citrixcws.SAMLSIGNREQUESTTYPE_YES)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLSIGNREQUESTTYPE_YES),
						string(citrixcws.SAMLSIGNREQUESTTYPE_NO),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"single_sign_on_service_url": schema.StringAttribute{
				Description: "The URL of the single sign-on service.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"single_sign_on_service_binding": schema.StringAttribute{
				Description: "The binding of the single sign-on service. Valid values are `HttpPost` and `HttpRedirect`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLREQUESTBINDINGTYPE_HTTP_POST),
						string(citrixcws.SAMLREQUESTBINDINGTYPE_HTTP_REDIRECT),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"saml_response": schema.StringAttribute{
				Description: "The SAML response. Valid values are `SignEitherResponseOrAssertion`, `MustSignResponse`, and `MustSignAssertion`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLRESPONSETYPE_SIGN_EITHER_RESPONSE_OR_ASSERTION),
						string(citrixcws.SAMLRESPONSETYPE_MUST_SIGN_RESPONSE),
						string(citrixcws.SAMLRESPONSETYPE_MUST_SIGN_ASSERTION),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert_file_path": schema.StringAttribute{
				Description: "The file path of the certificate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.SamlIdpCertRegex), "only PEM, CRT, and CER files are allowed."),
				},
			},
			"authentication_context": schema.StringAttribute{
				Description: "The authentication context. Valid values are `Unspecified`, `UserNameAndPassword`, `X509Cert`, `IntegratedWinAuth`, `Kerberos`, `PasswordProtectedTransport` and `TLSClient`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLAUTHCONTEXTTYPE_UNSPECIFIED),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_USER_NAME_AND_PASSWORD),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_X509_CERT),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_INTEGRATED_WIN_AUTH),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_KERBEROS),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_PASSWORD_PROTECTED_TRANSPORT),
						string(citrixcws.SAMLAUTHCONTEXTTYPE_TLS_CLIENT),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_context_comparison": schema.StringAttribute{
				Description: "The authentication context comparison type. Valid values are `Exact` and `Minimum`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLAUTHCONTEXTCOMPARISONTYPE_EXACT),
						string(citrixcws.SAMLAUTHCONTEXTCOMPARISONTYPE_MINIMUM),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"logout_url": schema.StringAttribute{
				Description: "The URL of the logout service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sign_logout_request": schema.StringAttribute{
				Description: "Whether to sign the logout request. Valid values are `Yes` and `No`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLSIGNREQUESTTYPE_YES),
						string(citrixcws.SAMLSIGNREQUESTTYPE_NO),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"logout_binding": schema.StringAttribute{
				Description: "The binding of the logout service. Valid values are `HttpPost` and `HttpRedirect`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixcws.SAMLREQUESTBINDINGTYPE_HTTP_POST),
						string(citrixcws.SAMLREQUESTBINDINGTYPE_HTTP_REDIRECT),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attribute_names": SamlAttributeNameMappings{}.GetSchema(),
			"cert_common_name": schema.StringAttribute{
				Description: "The common name of the SAML certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cert_expiration": schema.StringAttribute{
				Description: "The expiration date time of the SAML certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scoped_entity_id_suffix": schema.StringAttribute{
				Description: "The Scoped Entity Id Suffix for the SAML 2.0 Identity Provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (SamlIdentityProviderModel) GetAttributes() map[string]schema.Attribute {
	return SamlIdentityProviderModel{}.GetSchema().Attributes
}

func (r SamlIdentityProviderModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, samlIdp *citrixcws.IdpStatusModel, samlConfig *citrixcws.SamlConfigModel) SamlIdentityProviderModel {

	// Overwrite SAML 2.0 Identity Provider Resource with refreshed state
	r.Id = types.StringValue(samlIdp.GetIdpInstanceId())
	r.Name = types.StringValue(samlIdp.GetIdpNickname())
	if samlConfig == nil {
		return r
	}

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
	if !isResource {
		r.CertFilePath = types.StringNull()
	}
	r.AuthenticationContext = types.StringValue(samlConfig.GetSamlAuthenticationContext())
	r.AuthenticationContextComparison = types.StringValue(samlConfig.GetSamlAuthenticationContextComparison())
	r.LogoutUrl = types.StringValue(samlConfig.GetSamlLogoutUrl())
	if samlConfig.GetSamlLogoutUrl() != "" {
		r.SignLogoutRequest = types.StringValue(samlConfig.GetSamlSignLogoutRequest())
		r.LogoutBinding = types.StringValue(samlConfig.GetSamlLogoutRequestBinding())
	}

	// Refresh Attribute Name Mappings
	var samlAttributeNameMappings SamlAttributeNameMappings
	samlAttributeNameMappings.SecurityIdentifier = types.StringValue(samlConfig.GetSamlAttributeNameForSid())
	samlAttributeNameMappings.UserPrincipalName = types.StringValue(samlConfig.GetSamlAttributeNameForUpn())
	samlAttributeNameMappings.Email = types.StringValue(samlConfig.GetSamlAttributeNameForEmail())
	samlAttributeNameMappings.AdObjectIdentifier = types.StringValue(samlConfig.GetSamlAttributeNameForAdOid())
	samlAttributeNameMappings.AdForest = types.StringValue(samlConfig.GetSamlAttributeNameForAdForest())
	samlAttributeNameMappings.AdDomain = types.StringValue(samlConfig.GetSamlAttributeNameForAdDomain())
	samlAttributeNameMappings.UserDisplayName = types.StringValue(samlConfig.GetSamlAttributeNameForUserDisplayName())
	samlAttributeNameMappings.UserGivenName = types.StringValue(samlConfig.GetSamlAttributeNameForUserGivenName())
	samlAttributeNameMappings.UserFamilyName = types.StringValue(samlConfig.GetSamlAttributeNameForUserFamilyName())

	var samlAttributeNameMappingsObject types.Object
	if isResource {
		samlAttributeNameMappingsObject = util.TypedObjectToObjectValue(ctx, diagnostics, samlAttributeNameMappings)
	} else {
		samlAttributeNameMappingsObject = util.DataSourceTypedObjectToObjectValue(ctx, diagnostics, samlAttributeNameMappings)
	}
	r.AttributeNames = samlAttributeNameMappingsObject

	// Refresh Computed Values
	r.CertCommonName = types.StringValue(samlConfig.GetSamlCertCN())
	r.CertExpiration = types.StringValue(samlConfig.GetSamlCertExpiration())
	r.ScopedEntityIdSuffix = types.StringValue(samlConfig.GetSamlSpEntityIdSuffix())

	return r
}
