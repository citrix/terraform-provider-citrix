// Copyright © 2024. Citrix Systems, Inc.
package cvad_site

import (
	"context"
	"fmt"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (SiteSettingsModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data Source of the site settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the site.",
				Computed:    true,
			},
			"web_ui_policy_set_enabled": schema.BoolAttribute{
				Description: "Set this setting to `true` to show policy sets in the Policies node. With policy sets, you can group policies together for simplified, role-based access control.",
				Computed:    true,
			},
			"dns_resolution_enabled": schema.BoolAttribute{
				Description: "For Cloud environments, set this setting to `true` when enabling the Rendezvous protocol that lets HDX sessions bypass the Citrix Cloud Connector and connect directly and securely to the Citrix Gateway service. For On-Premises environments, Set this setting to `true` if you want to present DNS names instead of IP addresses in the ICA file.",
				Computed:    true,
			},
			"multiple_remote_pc_assignments": schema.BoolAttribute{
				Description: "Set this setting to `true` if you want to automatically assign multiple users to the next unassigned machine. Set it to `false` if you want to restrict the automatic assignment to a single user.",
				Computed:    true,
			},
			"trust_requests_sent_to_the_xml_service_port_enabled": schema.BoolAttribute{
				Description: "For Cloud customers, when set to `true`, the Cloud Connector (for cloud) or the Delivery Controller (for on-premises) trusts credentials sent from StoreFront. " +
					"\n\n~> **Please Note** This attribute should be set to `true` only when you have secured communications between the Cloud Connector (for cloud) or the Delivery Controller (for on-premises) and StoreFront using security keys, firewalls, or other mechanisms.",
				Computed: true,
			},
			"use_vertical_scaling_for_sessions_on_machines": schema.BoolAttribute{
				Description: "When set to `false`, sessions are distributed among the powered-on machines. For example, if you have two machines configured for 10 sessions each, the first machine handles five concurrent sessions and the second machine handles five. When set to `true`, sessions maximize powered-on machine capacity and save machine costs. For example, if you have two machines configured for 10 sessions each, the first machine handles the first 10 concurrent sessions. The second machine handles the eleventh session.",
				Computed:    true,
			},
			// On-Premises only settings
			"console_inactivity_timeout_minutes": schema.Int32Attribute{
				Description: "The inactivity duration in minutes after which administrators are automatically signed out of the Studio console. Minimum value is 10 and maximum value is 1440." +
					"\n\n~> **Please Note** This attribute is applicable only for On-Premises environments.",
				Computed: true,
			},
			"supported_authenticators": schema.StringAttribute{
				Description: fmt.Sprintf("The authentication methods for accessing the Studio. Available values are `%s` and `%s`", string(citrixorchestration.AUTHENTICATOR_BASIC), string(citrixorchestration.AUTHENTICATOR_ALL)) +
					fmt.Sprintf("\n\n~> **Please Note** This attribute is applicable only for On-Premises environments. When %s is specified, users authenticate to Studio using their domain credentials (user name and password). When %s is specified, users authenticate to Studio with their domain credentials (user name and password) or a with their Windows credentials, using Kerberos or NTLM.", string(citrixorchestration.AUTHENTICATOR_BASIC), string(citrixorchestration.AUTHENTICATOR_ALL)),
				Computed: true,
			},
			"allowed_cors_origins_for_iwa": schema.SetAttribute{
				Description: "Enable cross-origin access by adding the URL of the Web Studio server to the set." +
					"\n\n~> **Please Note** This attribute is applicable only for On-Premises environments. This attribute does not work if Web Studio is configured as a proxy for Delivery Controllers.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (SiteSettingsModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return SiteSettingsModel{}.GetDataSourceSchema().Attributes
}

func (r SiteSettingsModel) RefreshDataSourcePropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, siteSettingsResponse *citrixorchestration.SiteSettingsResponseModel, multipleRemotePcAssignments bool) SiteSettingsModel {
	r.SiteId = types.StringValue(client.ClientConfig.SiteId)
	r.WebUiPolicySetEnabled = types.BoolValue(siteSettingsResponse.GetWebUiPolicySetEnabled())
	r.DnsResolutionEnabled = types.BoolValue(siteSettingsResponse.GetDnsResolutionEnabled())
	r.MultipleRemotePCAssignments = types.BoolValue(multipleRemotePcAssignments)
	r.TrustRequestsSentToTheXmlServicePortEnabled = types.BoolValue(siteSettingsResponse.GetTrustRequestsSentToTheXmlServicePortEnabled())
	r.UseVerticalScalingForRdsLaunches = types.BoolValue(siteSettingsResponse.GetUseVerticalScalingForRdsLaunches())

	if client.AuthConfig.OnPremises {
		r.ConsoleInactivityTimeoutMinutes = types.Int32Value(siteSettingsResponse.GetConsoleInactivityTimeoutMinutes())
		r.SupportedAuthenticators = types.StringValue(string(siteSettingsResponse.GetSupportedAuthenticators()))
		r.AllowedCorsOriginsForIwa = util.StringArrayToStringSet(ctx, diagnostics, siteSettingsResponse.GetAllowedCorsOriginsForIwa())
	} else {
		r.ConsoleInactivityTimeoutMinutes = types.Int32Null()
		r.SupportedAuthenticators = types.StringNull()
		r.AllowedCorsOriginsForIwa = types.SetNull(types.StringType)
	}

	return r
}
