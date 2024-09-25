// Copyright © 2024. Citrix Systems, Inc.
package cc_identity_providers

import (
	"regexp"

	"github.com/citrix/citrix-daas-rest-go/citrixcws"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GoogleIdentityProviderResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	AuthDomainName   types.String `tfsdk:"auth_domain_name"`
	ClientEmail      types.String `tfsdk:"client_email"`
	PrivateKey       types.String `tfsdk:"private_key"`
	ImpersonatedUser types.String `tfsdk:"impersonated_user"`
	GoogleCustomerId types.String `tfsdk:"google_customer_id"`
	GoogleDomain     types.String `tfsdk:"google_domain"`
}

func (GoogleIdentityProviderResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Citrix Cloud --- Manages a Citrix Cloud Google Cloud Identity Provider instance. Note that this feature is in Tech Preview.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Citrix Cloud Google Cloud Identity Provider instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Citrix Cloud Google Cloud Identity Provider instance.",
				Required:    true,
			},
			"auth_domain_name": schema.StringAttribute{
				Description: "User authentication domain name for Google Cloud Identity Provider.",
				Required:    true,
			},
			"client_email": schema.StringAttribute{
				Description: "Email of the Google client for configuring Google Cloud Identity Provider.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.EmailRegex), "must be email format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "Private key of the Google Cloud Identity Provider.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"impersonated_user": schema.StringAttribute{
				Description: "Impersonated user for configuring Google Cloud Identity Provider.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (GoogleIdentityProviderResourceModel) GetAttributes() map[string]schema.Attribute {
	return GoogleIdentityProviderResourceModel{}.GetSchema().Attributes
}

func (r GoogleIdentityProviderResourceModel) RefreshIdAndNameValues(googleIdp *citrixcws.IdpStatusModel) GoogleIdentityProviderResourceModel {
	r.Id = types.StringValue(googleIdp.GetIdpInstanceId())
	r.Name = types.StringValue(googleIdp.GetIdpNickname())
	return r
}

func (r GoogleIdentityProviderResourceModel) RefreshPropertyValues(googleIdp *citrixcws.IdpStatusModel) GoogleIdentityProviderResourceModel {

	// Overwrite resource location with refreshed state
	r = r.RefreshIdAndNameValues(googleIdp)

	r.AuthDomainName = types.StringValue(googleIdp.GetAuthDomainName())

	additionalInfo := googleIdp.GetAdditionalStatusInfo()
	if additionalInfo != nil {
		r.GoogleCustomerId = types.StringValue(additionalInfo["googleCustomerId"])
		r.GoogleDomain = types.StringValue(additionalInfo["googleDomain"])
	} else {
		r.GoogleCustomerId = types.StringNull()
		r.GoogleDomain = types.StringNull()
	}

	return r
}
