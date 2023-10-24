package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	citrixdaas "github.com/citrix/terraform-provider-citrix/internal/daas"
	"github.com/google/uuid"

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
	Hostname               types.String `tfsdk:"hostname"`
	Environment            types.String `tfsdk:"environment"`
	CustomerId             types.String `tfsdk:"customer_id"`
	ClientId               types.String `tfsdk:"client_id"`
	ClientSecret           types.String `tfsdk:"client_secret"`
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
		Description: "Interact with Citrix Cloud / Citrix On-Premise Service.",
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Description: "Host name / base URL of Citrix DaaS service. " + "<br />" +
					"For Citrix On-Premise customers (Required): Use this to specify Delivery Controller hostname. " + "<br />" +
					"For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname." + "<br />" +
					"Can be set via Environment Variable **CITRIX_HOSTNAME**.",
				Optional: true,
			},
			"environment": schema.StringAttribute{
				Description: "Citrix Cloud environment of the customer. Only applicable for Citrix Cloud customers. Available options: `Production`, `Staging`, `Japan`, `JapanStaging`. " + "<br />" +
					"Can be set via Environment Variable **CITRIX_ENVIRONMENT**.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Production",
						"Staging",
						"Japan",
						"JapanStaging",
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
					"For Citrix On-Premise customers: Use this to specify Doamin Admin Username. " + "<br />" +
					"For Citrix Cloud customers: Use this to specify Cloud API Key Client Id." + "<br />" +
					"Can be set via Environment Variable **CITRIX_CLIENT_ID**.",
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client Secret for Citrix DaaS service authentication. " + "<br />" +
					"For Citrix On-Premise customers: Use this to specify Doamin Admin Password. " + "<br />" +
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
		},
	}
}

func getClientInterceptor(ctx context.Context) citrixclient.MiddlewareAuthFunction {
	return func(authClient *citrixclient.CitrixDaasClient, r *http.Request) {
		// Auth
		if authClient != nil && r.Header.Get("Authorization") == "" {
			token, err := authClient.SignIn()
			if err != nil {
				tflog.Error(ctx, "Could not sign into Citrix DaaS, error: "+err.Error())
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
		tflog.Info(ctx, "Orchestration API request", map[string]interface{}{
			"url":           r.URL.String(),
			"method":        r.Method,
			"transactionId": transactionId,
		})
	}
}

// Configure prepares a Citrixdaas API client for data sources and resources.
func (p *citrixProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Citrix Cloud client")

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

	if customerId == "" {
		customerId = "CitrixOnPremises"
	}

	onPremise := false
	if customerId == "CitrixOnPremises" {
		onPremise = true
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	// On-premise customer must specify hostname with DDC hostname / IP address
	if onPremise && hostname == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("hostname"),
			"Missing Citrix DaaS API Host",
			"The provider cannot create the Citrix API client as there is a missing or empty value for the Citrix DaaS API hostname for on-premise customers. "+
				"Set the host value in the configuration. Ensure the value is not empty. ",
		)
	}

	if !onPremise && disableSslVerification {
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

	if !onPremise && hostname == "" {
		if environment == "Production" {
			hostname = "api.cloud.com"
		} else if environment == "Staging" {
			hostname = "api.cloudburrito.com"
		} else if environment == "Japan" {
			hostname = "api.citrixcloud.jp"
		} else if environment == "JapanStaging" {
			hostname = "api.citrixcloudstaging.jp"
		}
	}

	authUrl := ""
	if onPremise {
		authUrl = fmt.Sprintf("https://%s/citrix/orchestration/api/techpreview/tokens", hostname)
	} else {
		if environment == "Production" {
			authUrl = fmt.Sprintf("https://api.cloud.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Staging" {
			authUrl = fmt.Sprintf("https://api.cloudburrito.com/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "Japan" {
			authUrl = fmt.Sprintf("https://api.citrixcloud.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else if environment == "JapanStaging" {
			authUrl = fmt.Sprintf("https://api.citrixcloudstaging.jp/cctrustoauth2/%s/tokens/clients", customerId)
		} else {
			authUrl = fmt.Sprintf("https://%s/cctrustoauth2/%s/tokens/clients", hostname, customerId)
		}
	}

	ctx = tflog.SetField(ctx, "citrix_hostname", hostname)
	if !onPremise {
		ctx = tflog.SetField(ctx, "citrix_customer_id", customerId)
	}
	ctx = tflog.SetField(ctx, "citrix_client_id", clientId)
	ctx = tflog.SetField(ctx, "citrix_client_secret", clientSecret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "citrix_client_secret")
	ctx = tflog.SetField(ctx, "citrix_on_premise", onPremise)

	tflog.Debug(ctx, "Creating Citrix API client")

	userAgent := "citrix-terraform-provider/" + p.version + " (https://github.com/citrix/terraform-provider-citrix)"

	// Create a new Citrix API client using the configuration values
	client, err := citrixclient.NewCitrixDaasClient(authUrl, hostname, customerId, clientId, clientSecret, onPremise, disableSslVerification, &userAgent, getClientInterceptor(ctx))
	if err != nil {
		errMsg := err.Error()
		resp.Diagnostics.AddError(
			"Unable to Create Citrix API Client",
			"An unexpected error occurred when creating the Citrix API client. \n\n"+
				"Error: "+errMsg+"\n\n"+
				"Ensure that the DDC(s) are running and that the client Id and secret are valid.\n\n"+
				"If you are running against on-premise DDC(s) that does not have an SSL certificate issue by a trusted CA, consider setting \"disable_ssl_verification\" to \"true\" in provider config.\n\n"+
				"If the error persists, please contact the provider developers.",
		)
		return
	}

	// Make the Citrix API client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured Citrix API client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *citrixProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *citrixProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		citrixdaas.NewMachineCatalogResource,
		citrixdaas.NewZoneResource,
		citrixdaas.NewHypervisorResource,
		citrixdaas.NewDeliveryGroupResource,
		citrixdaas.NewHypervisorResourcePoolResource,
	}
}
