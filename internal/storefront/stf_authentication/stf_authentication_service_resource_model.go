// Copyright © 2024. Citrix Systems, Inc.

package stf_authentication

import (
	"context"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"golang.org/x/exp/maps"

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
	SiteId            types.String `tfsdk:"site_id"`
	VirtualPath       types.String `tfsdk:"virtual_path"`
	FriendlyName      types.String `tfsdk:"friendly_name"`
	ClaimsFactoryName types.String `tfsdk:"claims_factory_name"`
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

	claimsFactoryNamesMapKeys := maps.Keys(claimsFactoryNamesMap)
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
		},
	}
}

func (STFAuthenticationServiceResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFAuthenticationServiceResourceModel{}.GetSchema().Attributes
}

func (STFAuthenticationServiceResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}
