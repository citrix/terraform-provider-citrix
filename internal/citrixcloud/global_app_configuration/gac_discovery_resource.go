// Copyright © 2024. Citrix Systems, Inc.

package global_app_configuration

import (
	"context"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	globalappconfiguration "github.com/citrix/citrix-daas-rest-go/globalappconfiguration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &gacDiscoveryResource{}
	_ resource.ResourceWithConfigure      = &gacDiscoveryResource{}
	_ resource.ResourceWithImportState    = &gacDiscoveryResource{}
	_ resource.ResourceWithModifyPlan     = &gacDiscoveryResource{}
	_ resource.ResourceWithValidateConfig = &gacDiscoveryResource{}
)

// NewGACDiscoveryResource is a helper function to simplify the provider implementation.
func NewGacDiscoveryResource() resource.Resource {
	return &gacDiscoveryResource{}
}

// GACDiscoveryResource is the resource implementation.
type gacDiscoveryResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *gacDiscoveryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gac_discovery"
}

// Schema defines the schema for the resource.
func (r *gacDiscoveryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GACDiscoveryResourceModel{}.GetSchema()
}

// Configure adds the provider configured client to the resource.
func (r *gacDiscoveryResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *gacDiscoveryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GACDiscoveryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var app globalappconfiguration.Apps
	var workspace globalappconfiguration.Workspace
	var domain globalappconfiguration.Domain
	domain.SetName(plan.Domain.ValueString())

	var aURLs []globalappconfiguration.AdminDomainURL
	var sURLs []globalappconfiguration.AdminDomainURL

	// Set the allowed web store urls
	urls := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.AllowedWebStoreURLs)
	for _, url := range urls {
		var aURL globalappconfiguration.AdminDomainURL
		aURL.SetUrl(url)
		aURLs = append(aURLs, aURL)
	}

	// Set the service urls
	urls = util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.ServiceURLs)
	for _, url := range urls {
		var sURL globalappconfiguration.AdminDomainURL
		sURL.SetUrl(url)
		sURLs = append(sURLs, sURL)
	}

	workspace.SetAllowedWebStoreURLs(aURLs)
	workspace.SetServiceURLs(sURLs)
	app.SetWorkspace(workspace)

	var body globalappconfiguration.DiscoveryRecordModel

	body.SetApp(app)
	body.SetDomain(domain)

	// Call the API
	createDiscoveryRequest := r.client.GacClient.DiscoveryDAAS.CreateDiscovery(ctx, util.GacAppName)
	createDiscoveryRequest = createDiscoveryRequest.DiscoveryRecordModel(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(createDiscoveryRequest, r.client).Execute()

	//In case of error, add it to diagnostics and return
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GAC Discovery for Domain: "+plan.Domain.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

	//Try to get the new discovery configuration from remote
	discoveryConfiguration, err := getDiscoveryConfiguration(ctx, r.client, &resp.Diagnostics, plan.Domain.ValueString())
	if err != nil {
		return
	}

	// Map response body to schema and populate computed attribute values
	if len(discoveryConfiguration.GetItems()) == 0 {
		resp.Diagnostics.AddError("Error fetching discovery configuration for domain: "+plan.Domain.ValueString(), "No discovery configuration found for domain: "+plan.Domain.ValueString())
		return
	} else {
		plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, discoveryConfiguration.GetItems()[0])
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *gacDiscoveryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state GACDiscoveryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get Service Url discovery from remote
	discoveryConfiguration, err := readDiscoveryConfiguration(ctx, r.client, resp, state.Domain.ValueString())
	if err != nil {
		return
	}

	state = state.RefreshPropertyValues(ctx, &resp.Diagnostics, discoveryConfiguration.GetItems()[0])

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *gacDiscoveryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var plan GACDiscoveryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var app globalappconfiguration.Apps
	var workspace globalappconfiguration.Workspace
	var domain globalappconfiguration.Domain
	domain.SetName(plan.Domain.ValueString())

	var aURLs []globalappconfiguration.AdminDomainURL
	var sURLs []globalappconfiguration.AdminDomainURL

	// Set the allowed web store urls
	urls := util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.AllowedWebStoreURLs)
	for _, url := range urls {
		var aURL globalappconfiguration.AdminDomainURL
		aURL.SetUrl(url)
		aURLs = append(aURLs, aURL)
	}

	// Set the service urls
	urls = util.StringSetToStringArray(ctx, &resp.Diagnostics, plan.ServiceURLs)
	for _, url := range urls {
		var sURL globalappconfiguration.AdminDomainURL
		sURL.SetUrl(url)
		sURLs = append(sURLs, sURL)
	}

	workspace.SetAllowedWebStoreURLs(aURLs)
	workspace.SetServiceURLs(sURLs)
	app.SetWorkspace(workspace)

	var body globalappconfiguration.DiscoveryRecordModel

	body.SetApp(app)
	body.SetDomain(domain)

	// Call the API
	updateDiscoveryRequest := r.client.GacClient.DiscoveryDAAS.UpdateDiscovery(ctx, util.GacAppName, plan.Domain.ValueString())
	updateDiscoveryRequest = updateDiscoveryRequest.DiscoveryRecordModel(body)
	_, httpResp, err := citrixdaasclient.AddRequestData(updateDiscoveryRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating discovery configuration: "+plan.Domain.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

	// Try to get Service Url discovery from remote
	updateddiscoveryConfiguration, err := getDiscoveryConfiguration(ctx, r.client, &resp.Diagnostics, plan.Domain.ValueString())
	if err != nil {
		return
	}

	plan = plan.RefreshPropertyValues(ctx, &resp.Diagnostics, updateddiscoveryConfiguration.GetItems()[0])

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *gacDiscoveryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state GACDiscoveryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Delete discovery configuration for the domain
	deleteDiscoveryRequest := r.client.GacClient.DiscoveryDAAS.DeleteDiscovery(ctx, util.GacAppName, state.Domain.ValueString())
	_, httpResp, err := citrixdaasclient.AddRequestData(deleteDiscoveryRequest, r.client).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting gac discovery configuration for domain: "+state.Domain.ValueString(),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return
	}

}

func (r *gacDiscoveryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}

func getDiscoveryConfiguration(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, domain string) (*globalappconfiguration.GetAllDiscoveryResponse, error) {
	getDiscoveryRequest := client.GacClient.DiscoveryDAAS.RetrieveDiscovery(ctx, util.GacAppName, domain)
	getDiscoveryResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*globalappconfiguration.GetAllDiscoveryResponse](getDiscoveryRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error fetching discovery configuration for doamin: "+domain,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadGacError(err),
		)
		return nil, err
	}

	return getDiscoveryResponse, nil
}

func readDiscoveryConfiguration(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, domain string) (*globalappconfiguration.GetAllDiscoveryResponse, error) {
	getDiscoveryRequest := client.GacClient.DiscoveryDAAS.RetrieveDiscovery(ctx, util.GacAppName, domain)
	getDiscoveryResponse, _, err := util.ReadResource[*globalappconfiguration.GetAllDiscoveryResponse](getDiscoveryRequest, ctx, client, resp, "Discovery Configuration", domain)
	return getDiscoveryResponse, err
}

func (r *gacDiscoveryResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	if r.client != nil && r.client.GacClient == nil {
		resp.Diagnostics.AddError(util.ProviderInitializationErrorMsg, util.MissingProviderClientIdAndSecretErrorMsg)
		return
	}
}

func (r *gacDiscoveryResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var data GACDiscoveryResourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaType, configValuesForSchema := util.GetConfigValuesForSchema(ctx, &resp.Diagnostics, &data)
	tflog.Debug(ctx, "Validate Config - "+schemaType, configValuesForSchema)
}
