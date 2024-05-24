// Copyright © 2023. Citrix Systems, Inc.
package stf_store

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfStoreServiceResource{}
	_ resource.ResourceWithConfigure   = &stfStoreServiceResource{}
	_ resource.ResourceWithImportState = &stfStoreServiceResource{}
)

// stfStoreServiceResource is a helper function to simplify the provider implementation.
func NewSTFStoreServiceResource() resource.Resource {
	return &stfStoreServiceResource{}
}

// stfStoreServiceResource is the resource implementation.
type stfStoreServiceResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfStoreServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_store_service"
}

// Schema defines the schema for the resource.
func (r *stfStoreServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Storefront StoreService.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the Storefront storeservice. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_path": schema.StringAttribute{
				Description: "The IIS VirtualPath at which the Store will be configured to be accessed by Receivers.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly name of the Store",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_service": schema.StringAttribute{
				Description: "The StoreFront Authentication Service to use for authenticating users.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"anonymous": schema.BoolAttribute{
				Description: "Whether the Store is anonymous. Anonymous Store not requiring authentication.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"load_balance": schema.BoolAttribute{
				Description: "Whether the Store is load balanced.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"farm_config": schema.SingleNestedAttribute{
				Description: "Farm configuration for the Store.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"farm_name": schema.StringAttribute{
						Description: "The name of the Farm.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"farm_type": schema.StringAttribute{
						Description: "The type of the Farm.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"servers": schema.ListAttribute{
						ElementType: types.StringType,
						Description: "The list of servers in the Farm.",
						Required:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *stfStoreServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stfStoreServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	var body citrixstorefront.CreateSTFStoreRequestModel

	siteIdInt, err := strconv.ParseInt(plan.SiteId.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Storefront StoreService ",
			"\nError message: "+err.Error(),
		)
		return
	}

	body.SetSiteId(siteIdInt)
	body.SetVirtualPath(plan.VirtualPath.String())
	body.SetFriendlyName(plan.FriendlyName.ValueString())

	if !plan.Anonymous.IsNull() && plan.Anonymous.ValueBool() {
		body.SetAnonymous(true)
	} else {
		body.SetAuthenticationService("(Get-STFAuthenticationService -VirtualPath " + plan.AuthenticationService.ValueString() + " ) ")
	}

	if !plan.LoadBalance.IsNull() {
		body.SetLoadBalance(plan.LoadBalance.ValueBool())
	}

	if plan.FarmConfig != nil {
		body.SetFarmName(plan.FarmConfig.FarmName.ValueString())
		body.SetFarmType(plan.FarmConfig.FarmType.ValueString())
		body.SetServers(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.FarmConfig.Servers))
	}

	createStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreCreateSTFStore(ctx, body)

	// Create new STF StoreService
	StoreServiceDetail, err := createStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Storefront StoreService",
			"TransactionId: ",
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.RefreshPropertyValues(&StoreServiceDetail)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *stfStoreServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFStoreServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	STFStoreService, err := getSTFStoreService(ctx, r.client, &resp.Diagnostics, state.SiteId.ValueStringPointer())
	if err != nil {
		return
	}
	state.RefreshPropertyValues(STFStoreService)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stfStoreServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFStoreServiceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed STFStoreService
	_, err := getSTFStoreService(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil {
		return
	}

	// Construct the update model
	var editSTFStoreServiceBody = &citrixstorefront.SetSTFStoreRequestModel{}
	editSTFStoreServiceBody.SetStoreService("(Get-STFStoreService -VirtualPath" + plan.VirtualPath.ValueString() + " )")

	// Update STFStoreService
	editStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreSetSTFStore(ctx, *editSTFStoreServiceBody)
	_, err = editStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Storefront StoreService ",
			"\nError message: "+err.Error(),
		)
	}

	// Fetch updated STFStoreService
	updatedSTFStoreService, err := getSTFStoreService(ctx, r.client, &resp.Diagnostics, plan.SiteId.ValueStringPointer())
	if err != nil {
		return
	}

	// Update resource state with updated property values
	plan.RefreshPropertyValues(updatedSTFStoreService)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stfStoreServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFStoreServiceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var body citrixstorefront.ClearSTFStoreRequestModel
	if state.SiteId.ValueString() != "" {
		body.SetStoreService("(Get-STFStoreService -VirtualPath " + state.VirtualPath.ValueString() + " -SiteId " + state.SiteId.ValueString() + " )")
	}

	// Delete existing STF StoreService
	deleteStoreServiceRequest := r.client.StorefrontClient.StoreSF.STFStoreClearSTFStore(ctx, body)
	_, err := deleteStoreServiceRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Storefront StoreService ",
			"\nError message: "+err.Error(),
		)
		return
	}
}

func (r *stfStoreServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `site_id,virtual_path`, got: %q", req.ID),
		)
		return
	}

	_, err := strconv.Atoi(idSegments[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Site ID in Import Identifier",
			fmt.Sprintf("Site ID should be an integer, got: %q", idSegments[0]),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_path"), idSegments[1])...)
}

// Gets the STFStoreService and logs any errors
func getSTFStoreService(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, siteId *string) (*citrixstorefront.STFStoreDetailModel, error) {
	var body citrixstorefront.GetSTFStoreRequestModel
	if siteId != nil {
		siteIdInt, err := strconv.ParseInt(*siteId, 10, 64)
		if err != nil {
			diagnostics.AddError(
				"Error fetching state of Storefront StoreService ",
				"Error message: "+err.Error(),
			)
			return nil, err
		}
		body.SetSiteId(siteIdInt)
	}
	getSTFStoreServiceRequest := client.StorefrontClient.StoreSF.STFStoreGetSTFStore(ctx, body)

	// Get refreshed STFStoreService properties from Orchestration
	STFStoreService, err := getSTFStoreServiceRequest.Execute()
	if err != nil {
		return &STFStoreService, err
	}
	return &STFStoreService, nil
}
