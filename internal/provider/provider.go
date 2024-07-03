// Copyright © 2024. Citrix Systems, Inc.

package provider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_role"
	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_user"
	"github.com/citrix/terraform-provider-citrix/internal/daas/application"
	"github.com/citrix/terraform-provider-citrix/internal/daas/gac_settings"
	"github.com/citrix/terraform-provider-citrix/internal/daas/resource_locations"
	"github.com/citrix/terraform-provider-citrix/internal/daas/storefront_server"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_authentication"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_deployment"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_multi_site"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_roaming"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_store"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_webreceiver"

	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_scope"
	"github.com/citrix/terraform-provider-citrix/internal/daas/delivery_group"
	"github.com/citrix/terraform-provider-citrix/internal/daas/hypervisor"
	"github.com/citrix/terraform-provider-citrix/internal/daas/hypervisor_resource_pool"
	"github.com/citrix/terraform-provider-citrix/internal/daas/machine_catalog"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/daas/zone"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/google/uuid"
	"golang.org/x/mod/semver"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &citrixProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &citrixProvider{
			version: version,
		}
	}
}

// citrixProvider is the provider implementation.
type citrixProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// citrixProviderModel maps provider schema data to a Go type.
type citrixProviderModel struct {
	Hostname               types.String      `tfsdk:"hostname"`
	Environment            types.String      `tfsdk:"environment"`
	CustomerId             types.String      `tfsdk:"customer_id"`
	ClientId               types.String      `tfsdk:"client_id"`
	ClientSecret           types.String      `tfsdk:"client_secret"`
	DisableSslVerification types.Bool        `tfsdk:"disable_ssl_verification"`
	StoreFrontRemoteHost   *storefrontConfig `tfsdk:"storefront_remote_host"`
}

type storefrontConfig struct {
	ComputerName    types.String `tfsdk:"computer_name"`
	ADadminUserName types.String `tfsdk:"ad_admin_username"`
	AdAdminPassword types.String `tfsdk:"ad_admin_password"`
}

// Metadata returns the provider type name.
func (p *citrixProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "citrix"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *citrixProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage and deploy Citrix resources easily using the Citrix Terraform provider. The provider currently supports both Citrix Virtual Apps & Desktops(CVAD) and Citrix Desktop as a Service (DaaS) solutions. You can automate creation of site setup including host connections, machine catalogs and delivery groups etc for both CVAD and Citrix DaaS. You can deploy resources in Citrix supported hypervisors and public clouds. Currently, we support deployments in Nutanix, VMware vSphere, XenServer, Microsoft Azure, AWS EC2 and Google Cloud Compute. Additionally, you can also use Manual provisioning or RemotePC to add workloads. The provider is developed and maintained by Citrix. Please note that this provider is still in **Tech Preview**.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "Host name / base URL of Citrix DaaS service. " + "<br />" +
					"For Citrix on-premises customers (Required): Use this to specify Delivery Controller hostname. " + "<br />" +
					"For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname." + "<br />" +
					"Can be set via Environment Variable **CITRIX_HOSTNAME**.",
				Optional: true,
			},
			"environment": schema.StringAttribute{
				Description: "Citrix Cloud environment of the customer. Only applicable for Citrix Cloud customers. Available options: `Production`, `Staging`, `Japan`, `JapanStaging`, `Gov`, `GovStaging`. " + "<br />" +
					"Can be set via Environment Variable **CITRIX_ENVIRONMENT**.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Production",
						"Staging",
						"Japan",
						"JapanStaging",
						"Gov",
						"GovStaging",
					),
				},
			},
			"customer_id": schema.StringAttribute{
				Description: "Citrix Cloud customer ID. Only applicable for Citrix Cloud customers." + "<br />" +
					"Can be set via Environment Variable **CITRIX_CUSTOMER_ID**.",
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Description: "Client Id for Citrix DaaS service authentication. " + "<br />" +
					"For Citrix On-Premises customers: Use this to specify a DDC administrator username. " + "<br />" +
					"For Citrix Cloud customers: Use this to specify Cloud API Key Client Id." + "<br />" +
					"Can be set via Environment Variable **CITRIX_CLIENT_ID**.",
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client Secret for Citrix DaaS service authentication. " + "<br />" +
					"For Citrix on-premises customers: Use this to specify a DDC administrator password. " + "<br />" +
					"For Citrix Cloud customers: Use this to specify Cloud API Key Client Secret." + "<br />" +
					"Can be set via Environment Variable **CITRIX_CLIENT_SECRET**.",
				Optional:  true,
				Sensitive: true,
			},
			"disable_ssl_verification": schema.BoolAttribute{
				Description: "Disable SSL verification against the target DDC. " + "<br />" +
					"Only applicable to on-premises customers. Citrix Cloud customers should omit this option. Set to true to skip SSL verification only when the target DDC does not have a valid SSL certificate issued by a trusted CA. " + "<br />" +
					"When set to true, please make sure that your provider config is set for a known DDC hostname. " + "<br />" +
					"[It is recommended to configure a valid certificate for the target DDC](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/secure-web-studio-deployment) " + "<br />" +
					"Can be set via Environment Variable **CITRIX_DISABLE_SSL_VERIFICATION**.",
				Optional: true,
			},
			"storefront_remote_host": schema.SingleNestedAttribute{
				Description: "StoreFront Remote Host for Citrix DaaS service. " + "<br />" +
					"Only applicable for Citrix on-premises StoreFront. Use this to specify StoreFront Remote Host. " + "<br />",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"computer_name": schema.StringAttribute{
						Required: true,
						Description: "StoreFront server computer Name " + "<br />" +
							"Only applicable for Citrix on-premises customers. Use this to specify StoreFront server computer name " + "<br />" +
							"Can be set via Environment Variable **SF_COMPUTER_NAME**.",
					},
					"ad_admin_username": schema.StringAttribute{
						Description: "Active Directory Admin Username to connect to storefront server " + "<br />" +
							"Only applicable for Citrix on-premises customers. Use this to specify AD admin username " + "<br />" +
							"Can be set via Environment Variable **SF_AD_ADMIN_USERNAME**.",
						Required: true,
					},
					"ad_admin_password": schema.StringAttribute{
						Description: "Active Directory Admin Password to connect to storefront server " + "<br />" +
							"Only applicable for Citrix on-premises customers. Use this to specify AD admin password" + "<br />" +
							"Can be set via Environment Variable **SF_AD_ADMIN_PASSWORD**.",
						Required: true,
					},
				},
			},
		},
	}
}

func middlewareAuthFunc(authClient *citrixclient.CitrixDaasClient, r *http.Request) {
	// Auth
	if authClient != nil && r.Header.Get("Authorization") == "" {
		token, _, err := authClient.SignIn()
		if err != nil {
			tflog.Error(r.Context(), "Could not sign into Citrix DaaS, error: "+err.Error())
		}
		r.Header["Authorization"] = []string{token}
	}

	// TransactionId
	transactionId := r.Header.Get("Citrix-TransactionId")
	if transactionId == "" {
		transactionId = uuid.NewString()
		r.Header.Add("Citrix-TransactionId", transactionId)
	}

	// Log the request
	tflog.Info(r.Context(), "Orchestration API request", map[string]interface{}{
		"url":           r.URL.String(),
		"method":        r.Method,
		"transactionId": transactionId,
	})
}

type registryResponse struct {
	Version  string   `json:"version"`
	Versions []string `json:"versions"`
}

func getVersionFromTerraformRegistry() (string, error) {
	httpResp, err := http.Get("https://registry.terraform.io/v1/providers/citrix/citrix")
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != 200 {
		return "", err
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	registryResp := registryResponse{}
	err = json.Unmarshal(body, &registryResp)
	if err != nil {
		return "", err
	}

	// find the last stable version
	// the versions are returned in order from oldest to newest so reverse it first
	slices.Reverse(registryResp.Versions)
	for _, ver := range registryResp.Versions {
		if semver.Prerelease("v"+ver) == "" {
			return ver, nil
		}
	}

	// if no stable version found just return the latest version
	return registryResp.Version, nil
}

// best effort version check, if anything goes wrong just bail out
func (p *citrixProvider) versionCheck(resp *provider.ConfigureResponse) {
	if !semver.IsValid("v" + p.version) {
		return
	}

	var registryVersion string
	updateVersionFile := false

	versionCheckFilePath := filepath.Join(os.TempDir(), "citrix_provider_version_check.txt")
	info, err := os.Stat(versionCheckFilePath)
	if err == nil && info != nil && time.Now().Before(info.ModTime().Add(time.Hour*time.Duration(24))) {
		// use cached version for version check
		txt, err := os.ReadFile(versionCheckFilePath)
		if err != nil {
			return
		}
		registryVersion = string(txt)
	} else {
		registryVersion, err = getVersionFromTerraformRegistry()
		if err != nil {
			return
		}
		updateVersionFile = true
	}

	if semver.Compare("v"+registryVersion, "v"+p.version) > 0 {
		resp.Diagnostics.AddWarning(
			"New version of the citrix/citrix provider is available",
			fmt.Sprintf("Please update the provider version in terraform configuration to >=%s and then run `terraform init --upgrade` to get the latest version.", registryVersion))
	}

	if updateVersionFile {
		err = os.WriteFile(versionCheckFilePath, []byte(registryVersion), 0660)
		if err != nil {
			return
		}
	}
}

// Configure prepares a Citrixdaas API client for data sources and resources.
func (p *citrixProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Citrix Cloud client")
	defer util.PanicHandler(&resp.Diagnostics)

	p.versionCheck(resp)

	// Retrieve provider data from configuration
	var config citrixProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Hostname.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Unknown Citrix DaaS API Hostname",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix DaaS API hostname. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CITRIX_HOSTNAME environment variable.",
		)
	}

	if config.Environment.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("environment"),
			"Unknown Citrix Cloud Environment",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix Cloud environment. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CITRIX_ENVIRONMENT environment variable.",
		)
	}

	if config.CustomerId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("customer_id"),
			"Unknown Citrix Cloud Customer ID",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix Cloud Customer ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CITRIX_CUSTOMER_ID environment variable.",
		)
	}

	if config.ClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Citrix API Client Id",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix API ClientId. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CITRIX_CLIENT_ID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Citrix API Client Secret",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix API ClientSecret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the CITRIX_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	hostname := os.Getenv("CITRIX_HOSTNAME")
	environment := os.Getenv("CITRIX_ENVIRONMENT")
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	clientId := os.Getenv("CITRIX_CLIENT_ID")
	clientSecret := os.Getenv("CITRIX_CLIENT_SECRET")
	disableSslVerification := strings.EqualFold(os.Getenv("CITRIX_DISABLE_SSL_VERIFICATION"), "true")
	storefront_computer_name := os.Getenv("SF_COMPUTER_NAME")
	storefront_ad_admin_username := os.Getenv("SF_AD_ADMIN_USERNAME")
	storefront_ad_admin_password := os.Getenv("SF_AD_ADMIN_PASSWORD")

	if !config.Hostname.IsNull() {
		hostname = config.Hostname.ValueString()
	}

	if !config.Environment.IsNull() {
		environment = config.Environment.ValueString()
	}

	if !config.CustomerId.IsNull() {
		customerId = config.CustomerId.ValueString()
	}

	if !config.ClientId.IsNull() {
		clientId = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	if !config.DisableSslVerification.IsNull() {
		disableSslVerification = config.DisableSslVerification.ValueBool()
	}

	if config.StoreFrontRemoteHost != nil {
		if !config.StoreFrontRemoteHost.ComputerName.IsNull() {
			storefront_computer_name = config.StoreFrontRemoteHost.ComputerName.ValueString()
		}
		if !config.StoreFrontRemoteHost.ADadminUserName.IsNull() {
			storefront_ad_admin_username = config.StoreFrontRemoteHost.ADadminUserName.ValueString()
		}
		if !config.StoreFrontRemoteHost.AdAdminPassword.IsNull() {
			storefront_ad_admin_password = config.StoreFrontRemoteHost.AdAdminPassword.ValueString()
		}
	}

	if environment == "" {
		environment = "Production" // default to production
	}

	if customerId == "" {
		customerId = "CitrixOnPremises"
	}

	onPremises := false
	if customerId == "CitrixOnPremises" {
		onPremises = true
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	// On-Premises customer must specify hostname with DDC hostname / IP address
	if onPremises && hostname == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Missing Citrix DaaS API Host",
			"The provider cannot create the Citrix API client as there is a missing or empty value for the Citrix DaaS API hostname for on-premises customers. "+
				"Set the host value in the configuration. Ensure the value is not empty. ",
		)
	}

	if !onPremises && disableSslVerification {
		resp.Diagnostics.AddAttributeError(
			path.Root("disable_ssl_verification"),
			"Cannot disable SSL verification for Citrix Cloud customer",
			"The provider cannot disable SSL verification in the Citrix API client against a Citrix Cloud customer as all Citrix Cloud requests has to go through secured TLS / SSL connection. "+
				"Omit disable_ssl_verification or set to false for Citrix Cloud customer. ",
		)
	}

	if clientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Citrix API key ClientId",
			"The provider cannot create the Citrix API client as there is a missing or empty value for the Citrix API ClientId. "+
				"Set the clientId value in the configuration. Ensure the value is not empty. ",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Citrix API key ClientSecret",
			"The provider cannot create the Citrix API client as there is a missing or empty value for the Citrix API ClientSecret. "+
				"Set the clientSecret value in the configuration. Ensure the value is not empty. ",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Indicate whether Citrix Cloud requests should go through API Gateway
	apiGateway := true
	ccUrl := ""
	if !onPremises {
		if environment == "Production" {
			ccUrl = "api.cloud.com"
		} else if environment == "Staging" {
			ccUrl = "api.cloudburrito.com"
		} else if environment == "Japan" {
			ccUrl = "api.citrixcloud.jp"
		} else if environment == "JapanStaging" {
			ccUrl = "api.citrixcloudstaging.jp"
		} else if environment == "Gov" {
			ccUrl = fmt.Sprintf("registry.citrixworkspacesapi.us/%s", customerId)
		} else if environment == "GovStaging" {
			ccUrl = fmt.Sprintf("registry.ctxwsstgapi.us/%s", customerId)
		}
		if hostname == "" {
			if environment == "Gov" {
				hostname = fmt.Sprintf("%s.xendesktop.us", customerId)
				apiGateway = false
			} else if environment == "GovStaging" {
				hostname = fmt.Sprintf("%s.xdstaging.us", customerId)
				apiGateway = false
			} else {
				hostname = ccUrl
			}
		} else if !strings.HasPrefix(hostname, "api.") {
			// When a cloud customer sets explicit hostname to the cloud DDC, bypass API Gateway
			apiGateway = false
		}
	}

	authUrl := ""
	isGov := false
	if onPremises {
		authUrl = fmt.Sprintf("https://%s/citrix/orchestration/api/tokens", hostname)
	} else {
		if environment == "Production" {
			authUrl = fmt.Sprintf("https://api.cloud.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Staging" {
			authUrl = fmt.Sprintf("https://api.cloudburrito.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Japan" {
			authUrl = fmt.Sprintf("https://api.citrixcloud.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "JapanStaging" {
			authUrl = fmt.Sprintf("https://api.citrixcloudstaging.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Gov" {
			authUrl = fmt.Sprintf("https://trust.citrixworkspacesapi.us/%s/tokens/clients", customerId)
			isGov = true
		} else if environment == "GovStaging" {
			authUrl = fmt.Sprintf("https://trust.ctxwsstgapi.us/%s/tokens/clients", customerId)
			isGov = true
		} else {
			authUrl = fmt.Sprintf("https://%s/cctrustoauth2/%s/tokens/clients", hostname, customerId)
		}
	}

	ctx = tflog.SetField(ctx, "citrix_hostname", hostname)
	if !onPremises {
		ctx = tflog.SetField(ctx, "citrix_customer_id", customerId)
	}
	ctx = tflog.SetField(ctx, "citrix_client_id", clientId)
	ctx = tflog.SetField(ctx, "citrix_client_secret", clientSecret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "citrix_client_secret")
	ctx = tflog.SetField(ctx, "citrix_on_premises", onPremises)
	if !onPremises {
		// customerId is considered sensitive information for Citrix Cloud customers
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "customerId", "customer_id", "citrix_customer_id")
		ctx = tflog.MaskAllFieldValuesStrings(ctx, customerId)
	}

	tflog.Debug(ctx, "Creating Citrix API client")

	userAgent := "citrix-terraform-provider/" + p.version + " (https://github.com/citrix/terraform-provider-citrix)"
	// Create a new Citrix API client using the configuration values
	client, httpResp, err := citrixclient.NewCitrixDaasClient(ctx, authUrl, ccUrl, hostname, customerId, clientId, clientSecret, onPremises, apiGateway, isGov, disableSslVerification, &userAgent, middlewareAuthFunc)
	if err != nil {
		if httpResp != nil {
			if httpResp.StatusCode == 401 {
				resp.Diagnostics.AddError(
					"Invalid credential in provider config",
					"Make sure client_id and client_secret is correct in provider config. ",
				)
			} else if httpResp.StatusCode >= 500 {
				if onPremises {
					resp.Diagnostics.AddError(
						"Citrix DaaS service unavailable",
						"Please check if you can access Web Studio. \n\n"+
							"Please ensure that Citrix Orchestration Service on the target DDC(s) are running reachable from this Machine.",
					)
				} else {
					resp.Diagnostics.AddError(
						"Citrix DaaS service unavailable",
						"The DDC(s) for the customer cannot be reached. Please check if you can access DaaS UI.",
					)
				}
			} else {
				resp.Diagnostics.AddError(
					"Unable to Create Citrix API Client",
					"An unexpected error occurred when creating the Citrix API client. \n\n"+
						"Error: "+err.Error(),
				)
			}
		} else {
			// Case 1: DDC off
			urlErr := new(url.Error)
			opErr := new(net.OpError)
			syscallErr := new(os.SyscallError)
			if errors.As(err, &urlErr) && errors.As(urlErr.Err, &opErr) && errors.As(opErr.Err, &syscallErr) && syscallErr.Err == syscall.Errno(10060) {
				resp.Diagnostics.AddError(
					"DDC(s) cannot be reached",
					"Ensure that the DDC(s) are running. Make sure this machine has proper network routing to reach the DDC(s) and is not blocked by any firewall rules.",
				)

				return
			}

			// Case 2: Invalid certificate
			cryptoErr := new(tls.CertificateVerificationError)
			errors.As(urlErr.Err, &cryptoErr)
			if len(cryptoErr.UnverifiedCertificates) > 0 {
				resp.Diagnostics.AddError(
					"DDC(s) does not have a valid SSL certificate issued by a trusted Certificate Authority",
					"If you are running against on-premises DDC(s) that does not have an SSL certificate issue by a trusted CA, consider setting \"disable_ssl_verification\" to \"true\" in provider config.",
				)

				return
			}

			// Case 3: Malformed hostname
			if urlErr != nil && opErr.Err == nil {
				resp.Diagnostics.AddError(
					"Invalid DDC(s) hostname",
					"Please revise the hostname in provider config and make sure it is a valid hostname or IP address.",
				)

				return
			}

			// Case 4: Catch all other errors
			resp.Diagnostics.AddError(
				"Unable to Create Citrix API Client",
				"An unexpected error occurred when creating the Citrix API client. \n\n"+
					"Error: "+err.Error(),
			)
		}

		return
	}

	// Set StoreFront Client
	client = citrixclient.NewStoreFrontClient(ctx, storefront_computer_name, storefront_ad_admin_username, storefront_ad_admin_password, client)

	// Make the Citrix API client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured Citrix API client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *citrixProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		zone.NewZoneDataSource,
		hypervisor.NewHypervisorDataSource,
		hypervisor_resource_pool.NewHypervisorResourcePoolDataSource,
		machine_catalog.NewMachineCatalogDataSource,
		delivery_group.NewDeliveryGroupDataSource,
		vda.NewVdaDataSource,
		application.NewApplicationDataSourceSource,
		admin_scope.NewAdminScopeDataSource,
		machine_catalog.NewPvsDataSource,
		// StoreFront DataSources
		stf_roaming.NewSTFRoamingServiceDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *citrixProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		zone.NewZoneResource,
		hypervisor.NewAzureHypervisorResource,
		hypervisor.NewAwsHypervisorResource,
		hypervisor.NewGcpHypervisorResource,
		hypervisor.NewVsphereHypervisorResource,
		hypervisor.NewXenserverHypervisorResource,
		hypervisor.NewNutanixHypervisorResource,
		hypervisor.NewSCVMMHypervisorResource,
		hypervisor_resource_pool.NewAzureHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewAwsHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewGcpHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewXenserverHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewVsphereHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewNutanixHypervisorResourcePoolResource,
		hypervisor_resource_pool.NewSCVMMHypervisorResourcePoolResource,
		machine_catalog.NewMachineCatalogResource,
		delivery_group.NewDeliveryGroupResource,
		storefront_server.NewStoreFrontServerResource,
		application.NewApplicationResource,
		application.NewApplicationFolderResource,
		application.NewApplicationGroupResource,
		application.NewApplicationIconResource,
		admin_scope.NewAdminScopeResource,
		admin_role.NewAdminRoleResource,
		policies.NewPolicySetResource,
		admin_user.NewAdminUserResource,
		gac_settings.NewAGacSettingsResource,
		resource_locations.NewResourceLocationResource,
		// StoreFront Resources
		stf_deployment.NewSTFDeploymentResource,
		stf_authentication.NewSTFAuthenticationServiceResource,
		stf_store.NewSTFStoreServiceResource,
		stf_store.NewSTFStoreFarmResource,
		stf_webreceiver.NewSTFWebReceiverResource,
		stf_multi_site.NewSTFUserFarmMappingResource,
		stf_roaming.NewSTFRoamingGatewayResource,
		// Add resource here
	}
}
