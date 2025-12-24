// Copyright © 2025. Citrix Systems, Inc.

package stf_authentication

import (
	"context"
	"maps"
	"slices"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"strconv"
)

// SFAuthenticationServiceResourceModel maps the resource schema data.

type STFAuthenticationServiceResourceModel struct {
	SiteId               types.String `tfsdk:"site_id"`
	VirtualPath          types.String `tfsdk:"virtual_path"`
	FriendlyName         types.String `tfsdk:"friendly_name"`
	ClaimsFactoryName    types.String `tfsdk:"claims_factory_name"`
	CitrixAGBasicOptions types.Object `tfsdk:"citrix_ag_basic_options"`
	AuthServiceProtocol  types.Object `tfsdk:"auth_service_protocol"`
}

func (r *STFAuthenticationServiceResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, authService *citrixstorefront.STFAuthenticationServiceResponseModel) {
	// Overwrite STFAuthenticationServiceResourceModel with refreshed state
	r.SiteId = types.StringValue(strconv.Itoa(*authService.SiteId.Get()))
	r.VirtualPath = types.StringValue(*authService.VirtualPath.Get())
	r.FriendlyName = types.StringValue(*authService.FriendlyName.Get())

	authSettings := authService.AuthenticationSettings
	claimsFactoryNamesMap := map[string]bool{}
	claimsFactoryNamesMap[*authSettings.IntegratedWindowsAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.CitrixAGBasicAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.ExplicitAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.HttpBasicAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.CertificateAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.CitrixFederationAuthentication.ClaimsFactoryName.Get()] = true
	claimsFactoryNamesMap[*authSettings.SamlForms.ClaimsFactoryName.Get()] = true

	claimsFactoryNamesMapKeys := slices.Collect(maps.Keys(claimsFactoryNamesMap))
	if len(claimsFactoryNamesMapKeys) != 1 {
		diagnostics.AddError(
			"Error refreshing STFAuthenticationService",
			"Claims factory names are not consistent across authentication settings.",
		)
		return
	}
	r.ClaimsFactoryName = types.StringValue(claimsFactoryNamesMapKeys[0])
}

func (STFAuthenticationServiceResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront Authentication Service.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site to configure the authentication service for. Defaults to `1`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_path": schema.StringAttribute{
				Description: "The IIS virtual path to use for the authentication service. Defaults to `/Citrix/Authentication`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/Citrix/Authentication"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name the authentication service should be known as. Defaults to `Authentication Service`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Authentication Service"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"claims_factory_name": schema.StringAttribute{
				Description: "The claims factory names to use for the StoreFront authentication services. Defaults to `standardClaimsFactory`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("standardClaimsFactory"),
			},
			"citrix_ag_basic_options": CitrixAGBasicOptions{}.GetSchema(),
			"auth_service_protocol":   AuthenticationServiceProtocol{}.GetSchema(),
		},
	}
}

func (STFAuthenticationServiceResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFAuthenticationServiceResourceModel{}.GetSchema().Attributes
}

func (STFAuthenticationServiceResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

type CitrixAGBasicOptions struct {
	CredentialValidationMode types.String `tfsdk:"credential_validation_mode"`
}

func (CitrixAGBasicOptions) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The Citrix AG Basic Authentication options.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"credential_validation_mode": schema.StringAttribute{
				Description: "The credential validation mode for Citrix AG Basic Authentication. Possible values are `Auto`, `Kerberos`, and `Password`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Auto", "Kerberos", "Password"),
				},
			},
		},
	}
}

func (CitrixAGBasicOptions) GetAttributes() map[string]schema.Attribute {
	return CitrixAGBasicOptions{}.GetSchema().Attributes
}

func (r *STFAuthenticationServiceResourceModel) RefreshCitrixAGBasicOptions(ctx context.Context, diagnostics *diag.Diagnostics, resp *citrixstorefront.STFCitrixAGBasicOptionsResponseModel) {
	refreshedAGBasicOptions := util.ObjectValueToTypedObject[CitrixAGBasicOptions](ctx, diagnostics, r.CitrixAGBasicOptions)

	switch resp.CredentialValidationMode {
	case 0:
		refreshedAGBasicOptions.CredentialValidationMode = types.StringValue("Password")
	case 1:
		refreshedAGBasicOptions.CredentialValidationMode = types.StringValue("Kerberos")
	case 2:
		refreshedAGBasicOptions.CredentialValidationMode = types.StringValue("Auto")
	}

	refreshedAGBasicOptionsObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedAGBasicOptions)
	r.CitrixAGBasicOptions = refreshedAGBasicOptionsObject
}

func (r *STFAuthenticationServiceResourceModel) RefreshAuthServiceProtocol(ctx context.Context, diagnostics *diag.Diagnostics, protocols []string) {
	refreshedProtocol := util.ObjectValueToTypedObject[AuthenticationServiceProtocol](ctx, diagnostics, r.AuthServiceProtocol)

	planProtocols := util.StringListToStringArray(ctx, diagnostics, refreshedProtocol.Name)

	remoteProtocolMap := make(map[string]bool)
	for _, p := range protocols {
		remoteProtocolMap[p] = true
	}

	resultProtocols := make([]string, 0)
	for _, p := range planProtocols {
		if remoteProtocolMap[p] {
			resultProtocols = append(resultProtocols, p)
			delete(remoteProtocolMap, p)
		}
	}

	refreshedProtocol.Name = util.RefreshListValues(ctx, diagnostics, refreshedProtocol.Name, resultProtocols)

	refreshedProtocolObject := util.TypedObjectToObjectValue(ctx, diagnostics, refreshedProtocol)
	r.AuthServiceProtocol = refreshedProtocolObject
}

type AuthenticationServiceProtocol struct {
	Name types.List `tfsdk:"name"`
}

// GetSchema returns the schema for AuthServiceProtocol.
func (AuthenticationServiceProtocol) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "StoreFront --- Authentication Service Protocol.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"name": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Names of the Authentication Service Protocol to be added.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (AuthenticationServiceProtocol) GetAttributes() map[string]schema.Attribute {
	return AuthenticationServiceProtocol{}.GetSchema().Attributes
}

func (r *AuthenticationServiceProtocol) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}
