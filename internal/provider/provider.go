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
	cc_admin_user "github.com/citrix/terraform-provider-citrix/internal/citrixcloud/admin_user"
	"github.com/citrix/terraform-provider-citrix/internal/citrixcloud/gac_settings"
	cc_identity_providers "github.com/citrix/terraform-provider-citrix/internal/citrixcloud/identity_providers"
	"github.com/citrix/terraform-provider-citrix/internal/citrixcloud/resource_locations"
	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_role"
	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_user"
	"github.com/citrix/terraform-provider-citrix/internal/daas/application"
	"github.com/citrix/terraform-provider-citrix/internal/daas/bearer_token"
	"github.com/citrix/terraform-provider-citrix/internal/daas/cvad_site"
	"github.com/citrix/terraform-provider-citrix/internal/daas/desktop_icon"
	"github.com/citrix/terraform-provider-citrix/internal/daas/storefront_server"
	"github.com/citrix/terraform-provider-citrix/internal/daas/tags"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/citrix/terraform-provider-citrix/internal/middleware"
	"github.com/citrix/terraform-provider-citrix/internal/quickcreate/qcs_account"
	"github.com/citrix/terraform-provider-citrix/internal/quickcreate/qcs_connection"
	"github.com/citrix/terraform-provider-citrix/internal/quickcreate/qcs_deployment"
	"github.com/citrix/terraform-provider-citrix/internal/quickcreate/qcs_image"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_authentication"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_deployment"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_multi_site"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_roaming"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_store"
	"github.com/citrix/terraform-provider-citrix/internal/storefront/stf_webreceiver"

	"github.com/citrix/terraform-provider-citrix/internal/wem/wem_machine_ad_object"
	"github.com/citrix/terraform-provider-citrix/internal/wem/wem_site"

	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_folder"
	"github.com/citrix/terraform-provider-citrix/internal/daas/admin_scope"
	"github.com/citrix/terraform-provider-citrix/internal/daas/delivery_group"
	"github.com/citrix/terraform-provider-citrix/internal/daas/hypervisor"
	"github.com/citrix/terraform-provider-citrix/internal/daas/hypervisor_resource_pool"
	"github.com/citrix/terraform-provider-citrix/internal/daas/machine_catalog"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/daas/zone"
	"github.com/citrix/terraform-provider-citrix/internal/util"

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
	CvadConfig           *cvadConfig       `tfsdk:"cvad_config"`
	StoreFrontRemoteHost *storefrontConfig `tfsdk:"storefront_remote_host"`
}

type cvadConfig struct {
	Hostname               types.String `tfsdk:"hostname"`
	Environment            types.String `tfsdk:"environment"`
	CustomerId             types.String `tfsdk:"customer_id"`
	ClientId               types.String `tfsdk:"client_id"`
	ClientSecret           types.String `tfsdk:"client_secret"`
	DisableSslVerification types.Bool   `tfsdk:"disable_ssl_verification"`
	DisableDaaSClient      types.Bool   `tfsdk:"disable_daas_client"`
	WemRegion              types.String `tfsdk:"wem_region"`
}

type storefrontConfig struct {
	ComputerName           types.String `tfsdk:"computer_name"`
	ADadminUserName        types.String `tfsdk:"ad_admin_username"`
	AdAdminPassword        types.String `tfsdk:"ad_admin_password"`
	DisableSslVerification types.Bool   `tfsdk:"disable_ssl_verification"`
}

// Metadata returns the provider type name.
func (p *citrixProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "citrix"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *citrixProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage and deploy Citrix resources easily using the Citrix Terraform provider. The provider currently supports both Citrix Virtual Apps & Desktops(CVAD) and Citrix Desktop as a Service (DaaS) solutions. You can automate creation of site setup including host connections, machine catalogs and delivery groups etc for both CVAD and Citrix DaaS. You can deploy resources in Citrix supported hypervisors and public clouds. Currently, we support deployments in Nutanix, VMware vSphere, XenServer, Microsoft Azure, AWS EC2 and Google Cloud Compute. Additionally, you can also use Manual provisioning or RemotePC to add workloads. The provider is developed and maintained by Citrix.",
		Attributes: map[string]schema.Attribute{
			"cvad_config": schema.SingleNestedAttribute{
				Description: "Configuration for CVAD service.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Description: "Host name / base URL of Citrix DaaS service. " +
							"\nFor Citrix on-premises customers: Use this to specify Delivery Controller hostname. " +
							"\nFor Citrix Cloud customers: Use this to force override the Citrix DaaS service hostname." +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_HOSTNAME**." +
							"\n\n~> **Please Note** This parameter is required for on-premises customers to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"environment": schema.StringAttribute{
						Description: "Citrix Cloud environment of the customer. Available options: `Production`, `Staging`, `Japan`, `JapanStaging`, `Gov`, `GovStaging`. " +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_ENVIRONMENT**." +
							"\n\n~> **Please Note** Only applicable for Citrix Cloud customers.",
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
						Description: "The Citrix Cloud customer ID." +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_CUSTOMER_ID**." +
							"\n\n~> **Please Note** This parameter is required for Citrix Cloud customers to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"client_id": schema.StringAttribute{
						Description: "Client Id for Citrix DaaS service authentication. " +
							"\nFor Citrix On-Premises customers: Use this to specify a DDC administrator username. " +
							"\nFor Citrix Cloud customers: Use this to specify Cloud API Key Client Id." +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_CLIENT_ID**." +
							"\n\n~> **Please Note** This parameter is required to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"client_secret": schema.StringAttribute{
						Description: "Client Secret for Citrix DaaS service authentication. " +
							"\nFor Citrix on-premises customers: Use this to specify a DDC administrator password. " +
							"\nFor Citrix Cloud customers: Use this to specify Cloud API Key Client Secret." +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_CLIENT_SECRET**." +
							"\n\n~> **Please Note** This parameter is required to be specified in the provider configuration or via environment variable.",
						Optional:  true,
						Sensitive: true,
					},
					"disable_ssl_verification": schema.BoolAttribute{
						Description: "Disable SSL verification against the target DDC. " +
							"\nSet to true to skip SSL verification only when the target DDC does not have a valid SSL certificate issued by a trusted CA. " +
							"\nWhen set to true, please make sure that your provider config is set for a known DDC hostname. " +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_DISABLE_SSL_VERIFICATION**." +
							"\n\n~> **Please Note** [It is recommended to configure a valid certificate for the target DDC](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/secure-web-studio-deployment) ",
						Optional: true,
					},
					"disable_daas_client": schema.BoolAttribute{
						Description: "Disable Citrix DaaS client setup. " +
							"\nSet to true to skip Citrix DaaS client setup. " +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_DISABLE_DAAS_CLIENT**.",
						Optional: true,
					},
					"wem_region": schema.StringAttribute{
						Description: "WEM Hosting Region of the Citrix Cloud customer. Available values are `US`, `EU`, and `APS`." +
							"\n\n-> **Note** Can be set via Environment Variable **CITRIX_WEM_REGION**." +
							"\n\n~> **Please Note** Only applicable for Citrix Workspace Environment Management (WEM) Cloud customers.",
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"US",
								"EU",
								"APS",
							),
						},
					},
				},
			},
			"storefront_remote_host": schema.SingleNestedAttribute{
				Description: "StoreFront Remote Host for Citrix DaaS service. " + "<br />" +
					"Only applicable for Citrix on-premises StoreFront. Use this to specify StoreFront Remote Host. " + "<br />",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"computer_name": schema.StringAttribute{
						Description: "StoreFront server computer Name " + "<br />" +
							"Use this to specify StoreFront server computer name " + "<br />" +
							"Can be set via Environment Variable **SF_COMPUTER_NAME**." + "<br />" +
							"This parameter is **required** to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"ad_admin_username": schema.StringAttribute{
						Description: "Active Directory Admin Username to connect to storefront server " + "<br />" +
							"Use this to specify AD admin username " + "<br />" +
							"Can be set via Environment Variable **SF_AD_ADMIN_USERNAME**." + "<br />" +
							"This parameter is **required** to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"ad_admin_password": schema.StringAttribute{
						Description: "Active Directory Admin Password to connect to storefront server " + "<br />" +
							"Use this to specify AD admin password" + "<br />" +
							"Can be set via Environment Variable **SF_AD_ADMIN_PASSWORD**." + "<br />" +
							"This parameter is **required** to be specified in the provider configuration or via environment variable.",
						Optional: true,
					},
					"disable_ssl_verification": schema.BoolAttribute{
						Description: "Disable SSL verification against the target storefront server. " + "<br />" +
							"Only applicable to customers connecting to storefront server remotely. Customers should omit this option when running storefront provider locally. Set to true to skip SSL verification only when the target DDC does not have a valid SSL certificate issued by a trusted CA. " + "<br />" +
							"When set to true, please make sure that your provider storefront_remote_host is set for a known storefront hostname. " + "<br />" +
							"Can be set via Environment Variable **SF_DISABLE_SSL_VERIFICATION**.",
						Optional: true,
					},
				},
			},
		},
	}
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
// **Important Note**: Provider client initialization logic is also implemented for sweeper in sweeper_test.go
// Please make sure to update the sweeper client initialization in sweeper_test.go if any changes are made here.
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

	client := &citrixclient.CitrixDaasClient{}

	storeFrontClientInitialized := false
	daasClientInitialized := false

	// Initialize storefront client
	storefront_computer_name := os.Getenv("SF_COMPUTER_NAME")
	storefront_ad_admin_username := os.Getenv("SF_AD_ADMIN_USERNAME")
	storefront_ad_admin_password := os.Getenv("SF_AD_ADMIN_PASSWORD")
	storefront_disable_ssl_verification := strings.EqualFold(os.Getenv("SF_DISABLE_SSL_VERIFICATION"), "true")

	if storefrontConfig := config.StoreFrontRemoteHost; storefrontConfig != nil || (storefront_computer_name != "" && storefront_ad_admin_username != "" && storefront_ad_admin_password != "") {
		if storefrontConfig != nil {
			if !storefrontConfig.ComputerName.IsNull() {
				storefront_computer_name = storefrontConfig.ComputerName.ValueString()
			}
			if !storefrontConfig.ADadminUserName.IsNull() {
				storefront_ad_admin_username = storefrontConfig.ADadminUserName.ValueString()
			}
			if !storefrontConfig.AdAdminPassword.IsNull() {
				storefront_ad_admin_password = storefrontConfig.AdAdminPassword.ValueString()
			}
			if !storefrontConfig.DisableSslVerification.IsNull() {
				storefront_disable_ssl_verification = storefrontConfig.DisableSslVerification.ValueBool()
			}
		}

		validateAndInitializeStorefrontClient(ctx, resp, client, storefront_computer_name, storefront_ad_admin_username, storefront_ad_admin_password, storefront_disable_ssl_verification)
		if resp.Diagnostics.HasError() {
			return
		}
		storeFrontClientInitialized = true
	}

	// Initialize cvad client
	clientId := os.Getenv("CITRIX_CLIENT_ID")
	clientSecret := os.Getenv("CITRIX_CLIENT_SECRET")
	hostname := os.Getenv("CITRIX_HOSTNAME")
	environment := os.Getenv("CITRIX_ENVIRONMENT")
	customerId := os.Getenv("CITRIX_CUSTOMER_ID")
	disableSslVerification := strings.EqualFold(os.Getenv("CITRIX_DISABLE_SSL_VERIFICATION"), "true")
	disableDaasClient := strings.EqualFold(os.Getenv("CITRIX_DISABLE_DAAS_CLIENT"), "true")
	wemRegion := os.Getenv("CITRIX_WEM_REGION")
	wemHostName := os.Getenv("CITRIX_WEM_HOSTNAME")
	quick_create_host_name := os.Getenv("CITRIX_QUICK_CREATE_HOST_NAME")

	if cvadConfig := config.CvadConfig; cvadConfig != nil || (clientId != "" && clientSecret != "") {
		if cvadConfig != nil {
			if !cvadConfig.ClientId.IsNull() {
				clientId = cvadConfig.ClientId.ValueString()
			}

			if !cvadConfig.ClientSecret.IsNull() {
				clientSecret = cvadConfig.ClientSecret.ValueString()
			}

			if !cvadConfig.Hostname.IsNull() {
				hostname = cvadConfig.Hostname.ValueString()
			}

			if !cvadConfig.Environment.IsNull() {
				environment = cvadConfig.Environment.ValueString()
			}

			if !cvadConfig.CustomerId.IsNull() {
				customerId = cvadConfig.CustomerId.ValueString()
			}

			if !cvadConfig.DisableSslVerification.IsNull() {
				disableSslVerification = cvadConfig.DisableSslVerification.ValueBool()
			}

			if !cvadConfig.DisableDaaSClient.IsNull() {
				disableDaasClient = cvadConfig.DisableDaaSClient.ValueBool()
			}
			if !cvadConfig.WemRegion.IsNull() {
				wemRegion = cvadConfig.WemRegion.ValueString()
			}
		}

		validateAndInitializeDaaSClient(ctx, resp, client, clientId, clientSecret, hostname, environment, wemHostName, wemRegion, customerId, quick_create_host_name, p.version, disableSslVerification, disableDaasClient)
		if resp.Diagnostics.HasError() {
			return
		}
		daasClientInitialized = true
	}

	if !storeFrontClientInitialized && !daasClientInitialized {
		resp.Diagnostics.AddError(
			"Invalid Provider Configuration",
			"At least one of `cvad_config` and `storefront_remote_host` attributes must be specified.",
		)
		return
	}

	// Make the Citrix API client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Citrix API client", map[string]any{"success": true})
}

func validateAndInitializeStorefrontClient(ctx context.Context, resp *provider.ConfigureResponse, client *citrixclient.CitrixDaasClient, storefront_computer_name, storefront_ad_admin_username, storefront_ad_admin_password string, storefront_disable_ssl_verification bool) {
	if storefront_computer_name == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("storefront_remote_host").AtName("computer_name"),
			"Unknown StoreFront Computer Name",
			"The provider cannot create the Citrix StoreFront client as there is an unknown configuration value for the StoreFront Computer Name. "+
				"Either set the value in the provider configuration, or use the SF_COMPUTER_NAME environment variable.",
		)
		return
	}
	if storefront_ad_admin_username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("storefront_remote_host").AtName("ad_admin_username"),
			"Unknown StoreFront AD Admin Username",
			"The provider cannot create the Citrix StoreFront client as there is an unknown configuration value for the StoreFront AD Admin Username. "+
				"Either set the value in the provider configuration, or use the SF_AD_ADMIN_USERNAME environment variable.",
		)
		return
	}
	if storefront_ad_admin_password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("storefront_remote_host").AtName("ad_admin_password"),
			"Unknown StoreFront AD Admin Password",
			"The provider cannot create the Citrix StoreFront client as there is an unknown configuration value for the StoreFront AD Admin Password. "+
				"Either set the value in the provider configuration, or use the SF_AD_ADMIN_PASSWORD environment variable.",
		)
		return
	}
	client.InitializeStoreFrontClient(ctx, storefront_computer_name, storefront_ad_admin_username, storefront_ad_admin_password, storefront_disable_ssl_verification)
}

func validateAndInitializeDaaSClient(ctx context.Context, resp *provider.ConfigureResponse, client *citrixclient.CitrixDaasClient, clientId, clientSecret, hostname, environment, wemHostName, wemRegion, customerId, quick_create_host_name, version string, disableSslVerification bool, disableDaasClient bool) {
	if clientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cvad_config").AtName("client_id"),
			"Unknown Citrix API Client Id",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix API ClientId. "+
				"Either set the value in the provider configuration, or use the CITRIX_CLIENT_ID environment variable.",
		)
		return
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cvad_config").AtName("client_secret"),
			"Unknown Citrix API Client Secret",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix API ClientSecret. "+
				"Either set the value in the provider configuration, or use the CITRIX_CLIENT_SECRET environment variable.",
		)
		return
	}

	if customerId == "" && hostname == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cvad_config").AtName("customer_id"),
			"Citrix API Customer Id or Hostname is Required",
			"The provider cannot create the Citrix API client as there is an unknown configuration value for the Citrix API CustomerId and Hostname. "+
				"Either set the value in the provider configuration, or using the CITRIX_CUSTOMER_ID or CITRIX_HOSTNAME environment variables.",
		)
		return
	}

	if environment == "" {
		environment = "Production" // default to production
	}

	onPremises := false
	if customerId == "" {
		customerId = "CitrixOnPremises"
		onPremises = true
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	// On-Premises customer must specify hostname with DDC hostname / IP address
	if onPremises && hostname == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cvad_config").AtName("hostname"),
			"Missing Citrix DaaS API Hostname",
			"The provider cannot create the Citrix API client as there is a missing or empty value for the Citrix DaaS API hostname for on-premises customers. "+
				"Set the host value in the configuration. Ensure the value is not empty. ",
		)
	}

	if !onPremises && disableSslVerification {
		resp.Diagnostics.AddAttributeError(
			path.Root("cvad_config").AtName("disable_ssl_verification"),
			"Cannot disable SSL verification for Citrix Cloud customer",
			"The provider cannot disable SSL verification in the Citrix API client against a Citrix Cloud customer as all Citrix Cloud requests has to go through secured TLS / SSL connection. "+
				"Omit disable_ssl_verification or set to false for Citrix Cloud customer. ",
		)
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

	quickCreateHostname := ""
	if quick_create_host_name != "" {
		// If customer specified a quick create host name, use it
		quickCreateHostname = quick_create_host_name
	} else {
		if environment == "Production" {
			quickCreateHostname = "api.cloud.com/quickcreateservice"
		} else if environment == "Staging" {
			quickCreateHostname = "api.cloudburrito.com/quickcreateservice"
		} else if environment == "Japan" {
			quickCreateHostname = "api.citrixcloud.jp/quickcreateservice"
		} else if environment == "JapanStaging" {
			quickCreateHostname = "api.citrixcloudstaging.jp/quickcreateservice"
		} else if environment == "Gov" {
			quickCreateHostname = "quickcreate.apps.cloud.us"
		} else if environment == "GovStaging" {
			quickCreateHostname = "quickcreate.apps.cloudstaging.us"
		}
	}

	cwsHostName := ""
	if environment == "Production" {
		cwsHostName = "cws.citrixworkspacesapi.net"
	} else if environment == "Staging" {
		cwsHostName = "cws.ctxwsstgapi.net"
	} else if environment == "Japan" {
		cwsHostName = "cws.citrixworkspacesapi.jp"
	} else if environment == "JapanStaging" {
		cwsHostName = "cws.citrixstagingapi.jp"
	} else if environment == "Gov" {
		cwsHostName = "cws.citrixworkspacesapi.us"
	} else if environment == "GovStaging" {
		cwsHostName = "cws.ctxwsstgapi.us"
	}

	if wemHostName == "" {
		if environment == "Production" {
			if wemRegion == "EU" {
				wemHostName = "eu-api.wem.cloud.com"
			} else if wemRegion == "APS" {
				wemHostName = "aps-api.wem.cloud.com"
			} else {
				wemHostName = "api.wem.cloud.com"
			}
		} else if environment == "Japan" {
			wemHostName = "jp-api.wem.citrixcloud.jp"
		} else if environment == "Staging" {
			wemHostName = "api.wem.cloudburrito.com"
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

	userAgent := "citrix-terraform-provider/" + version + " (https://github.com/citrix/terraform-provider-citrix)"

	// Setup the Citrix API Client
	token, httpResp, err := client.SetupCitrixClientsContext(ctx, authUrl, ccUrl, hostname, customerId, clientId, clientSecret, onPremises, apiGateway, isGov, disableSslVerification, &userAgent, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 401 {
			resp.Diagnostics.AddError(
				"Invalid credential in provider config",
				"Make sure client_id and client_secret are correct in provider config.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Create Citrix API Client",
				"An unexpected error occurred when creating the Citrix API client.\n\n"+
					"Error: "+err.Error(),
			)
		}
		return
	}

	// Initialize the Cloud Clients if not on-premises
	if !onPremises {
		client.InitializeCitrixCloudClients(ctx, ccUrl, hostname, middleware.MiddlewareAuthFunc, middleware.MiddlewareAuthWithCustomerIdHeaderFunc)
	}

	if !disableDaasClient {
		// Setup the DAAS Client
		httpResp, err = client.InitializeCitrixDaasClient(ctx, customerId, token, onPremises, apiGateway, disableSslVerification, &userAgent)
		if err != nil {
			if httpResp != nil {
				if httpResp.StatusCode >= 500 {
					if onPremises {
						resp.Diagnostics.AddError(
							"Citrix DaaS service unavailable",
							"Please check if you can access Web Studio. \n\n"+
								"Please ensure that Citrix Orchestration Service on the target DDC(s) are running reachable from this Machine.",
						)
					} else {
						resp.Diagnostics.AddError(
							"Citrix DaaS service unavailable",
							"The DDC(s) for the customer cannot be reached. Please check if you can access DaaS UI.\n\n"+
								"Note: If you are running resources that do not require DaaS/CVAD entitlement, you can set `disable_daas_client` to `true` in the provider configuration to skip the DaaS client setup.",
						)
					}
				} else {
					resp.Diagnostics.AddError(
						"Unable to Create Citrix API Client",
						"An unexpected error occurred when creating the Citrix API client.\n\n"+
							"Error: "+err.Error(),
					)
				}

			} else {
				handleNetworkError(err, resp)
			}
		}
	}

	// Set Quick Create Client
	if quickCreateHostname != "" {
		client.InitializeQuickCreateClient(ctx, quickCreateHostname, middleware.MiddlewareAuthFunc)
	}
	// Set CWS Client
	if cwsHostName != "" {
		client.InitializeCwsClient(ctx, cwsHostName, middleware.MiddlewareAuthFunc)
	}
	// Set WEM Client
	if wemHostName != "" {
		client.InitializeWemClient(ctx, wemHostName, middleware.MiddlewareAuthFunc)
	}
}

func handleNetworkError(err error, resp *provider.ConfigureResponse) {
	urlErr := new(url.Error)
	opErr := new(net.OpError)
	syscallErr := new(os.SyscallError)

	// Case 1: DDC off
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
	return
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
		admin_folder.NewAdminFolderDataSource,
		admin_scope.NewAdminScopeDataSource,
		machine_catalog.NewPvsDataSource,
		bearer_token.NewBearerTokenDataSource,
		cvad_site.NewSiteDataSource,
		tags.NewTagDataSource,
		// StoreFront DataSources
		stf_roaming.NewSTFRoamingServiceDataSource,
		// QuickCreate DataSources
		qcs_image.NewAwsWorkspacesImageDataSource,
		qcs_account.NewAccountDataSource,
		qcs_account.NewAwsWorkspacesCloudFormationDataSource,
		qcs_connection.NewAwsWorkspacesDirectoryConnectionDataSource,
		qcs_deployment.NewAwsWorkspacesDeploymentDataSource,
		// CC Identity Provider Resources
		cc_identity_providers.NewOktaIdentityProviderDataSource,
		cc_identity_providers.NewGoogleIdentityProviderDataSource,
		cc_identity_providers.NewSamlIdentityProviderDataSource,
		// CC Resource Locations
		resource_locations.NewResourceLocationsDataSource,
		// WEM
		wem_site.NewWemSiteDataSource,
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
		application.NewApplicationGroupResource,
		application.NewApplicationIconResource,
		desktop_icon.NewDesktopIconResource,
		admin_folder.NewAdminFolderResource,
		admin_role.NewAdminRoleResource,
		admin_scope.NewAdminScopeResource,
		policies.NewPolicySetResource,
		admin_user.NewAdminUserResource,
		gac_settings.NewGacSettingsResource,
		resource_locations.NewResourceLocationResource,
		cc_admin_user.NewCCAdminUserResource,
		tags.NewTagResource,
		// StoreFront Resources
		stf_deployment.NewSTFDeploymentResource,
		stf_authentication.NewSTFAuthenticationServiceResource,
		stf_store.NewSTFStoreServiceResource,
		stf_store.NewXenappDefaultStoreResource,
		stf_webreceiver.NewSTFWebReceiverResource,
		stf_multi_site.NewSTFUserFarmMappingResource,
		// QuickCreate Resources
		qcs_account.NewAwsWorkspacesAccountResource,
		qcs_image.NewAwsEdcImageResource,
		qcs_connection.NewAwsWorkspacesDirectoryConnectionResource,
		qcs_deployment.NewAwsWorkspacesDeploymentResource,
		// CC Identity Provider Resources
		cc_identity_providers.NewGoogleIdentityProviderResource,
		cc_identity_providers.NewOktaIdentityProviderResource,
		cc_identity_providers.NewSamlIdentityProviderResource,
		// Wem Resources
		wem_site.NewWemSiteServiceResource,
		wem_machine_ad_object.NewWemDirectoryResource,
		// Add resource here
	}
}
