// Copyright © 2024. Citrix Systems, Inc.

package global_app_configuration

import (
	"context"
	"regexp"

	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GACDiscoveryResourceModel struct {
	Domain              types.String `tfsdk:"domain"`
	ServiceURLs         types.Set    `tfsdk:"service_urls"`
	AllowedWebStoreURLs types.Set    `tfsdk:"allowed_web_store_urls"` //
}

func (GACDiscoveryResourceModel) GetAttributes() map[string]schema.Attribute {
	return GACDiscoveryResourceModel{}.GetSchema().Attributes
}

func (GACDiscoveryResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Manages the Global App Configuration Discovery.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain name of the discovery record. The domain must not contain uppercase letters.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^A-Z]*$`), "The domain must not contain uppercase letters."),
				},
			},
			"service_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The list of store URLs that are returned for email-based discovery. Each URL must end with a port number like example.com:443.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(`:\d+$`), "Each URL must end with a port number."),
					),
				},
			},
			"allowed_web_store_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The list of custom Web URLs, URL needs to match the domain claimed (Optional).",
				Optional:    true,
			},
		},
	}
}

func (GACDiscoveryResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r GACDiscoveryResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, discoveryRecordModel globalappconfiguration.DiscoveryRecordModel) GACDiscoveryResourceModel {

	r.Domain = types.StringValue(discoveryRecordModel.Domain.GetName())

	var sURLs []string
	for _, serviceURL := range discoveryRecordModel.App.Workspace.ServiceURLs {
		sURLs = append(sURLs, *serviceURL.Url)
	}
	r.ServiceURLs = util.StringArrayToStringSet(ctx, diagnostics, sURLs)

	var aURLs []string
	for _, allowedWebStoreURL := range discoveryRecordModel.App.Workspace.AllowedWebStoreURLs {
		aURLs = append(aURLs, *allowedWebStoreURL.Url)
	}
	r.AllowedWebStoreURLs = util.StringArrayToStringSet(ctx, diagnostics, aURLs)

	return r
}
