// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OktaIdentityProviderModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	OktaDomain       types.String `tfsdk:"okta_domain"`
	OktaClientId     types.String `tfsdk:"okta_client_id"`
	OktaClientSecret types.String `tfsdk:"okta_client_secret"`
	OktaApiToken     types.String `tfsdk:"okta_api_token"`
}

func (OktaIdentityProviderModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Manages a Citrix Cloud Okta Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Citrix Cloud Identity Provider instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Citrix Cloud Identity Provider instance.",
				Required:    true,
			},
			"okta_domain": schema.StringAttribute{
				Description: "Okta domain name for configuring Okta Identity Provider.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.OktaDomainRegex), "must end with `.okta.com`, `.okta-eu.com`, or `.oktapreview.com`"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"okta_client_id": schema.StringAttribute{
				Description: "ID of the Okta client for configuring Okta Identity Provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"okta_client_secret": schema.StringAttribute{
				Description: "Secret of the Okta client for configuring Okta Identity Provider.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"okta_api_token": schema.StringAttribute{
				Description: "Okta API token for configuring Okta Identity Provider.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (OktaIdentityProviderModel) GetAttributes() map[string]schema.Attribute {
	return OktaIdentityProviderModel{}.GetSchema().Attributes
}

func (r OktaIdentityProviderModel) RefreshPropertyValues(isResource bool, oktaIdp *citrixcws.IdpStatusModel) OktaIdentityProviderModel {

	// Overwrite Okta Identity Provider Resource with refreshed state
	r.Id = types.StringValue(oktaIdp.GetIdpInstanceId())
	r.Name = types.StringValue(oktaIdp.GetIdpNickname())
	if oktaIdp.GetClientId() != "" {
		r.OktaClientId = types.StringValue(oktaIdp.GetClientId())
	}

	additionalInfo := oktaIdp.GetAdditionalStatusInfo()
	if additionalInfo != nil {
		r.OktaDomain = types.StringValue(strings.ReplaceAll(additionalInfo["oktaDomain"], "https://", ""))
	} else if !isResource {
		r.OktaDomain = types.StringNull()
	}

	if !isResource {
		r.OktaClientId = types.StringNull()
		r.OktaClientSecret = types.StringNull()
		r.OktaApiToken = types.StringNull()
	}

	return r
}
