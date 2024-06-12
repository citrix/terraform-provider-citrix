// Copyright © 2024. Citrix Systems, Inc.

package stf_store

import (
	"context"
	"strconv"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"

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

type FarmConfig struct {
	FarmName types.String `tfsdk:"farm_name"`
	FarmType types.String `tfsdk:"farm_type"`
	Servers  types.List   `tfsdk:"servers"` // []types.String
}

func (FarmConfig) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
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
	}
}

func (FarmConfig) GetAttributes() map[string]schema.Attribute {
	return FarmConfig{}.GetSchema().Attributes
}

// SFStoreServiceResourceModel maps the resource schema data.
type STFStoreServiceResourceModel struct {
	VirtualPath           types.String `tfsdk:"virtual_path"`
	SiteId                types.String `tfsdk:"site_id"`
	FriendlyName          types.String `tfsdk:"friendly_name"`
	AuthenticationService types.String `tfsdk:"authentication_service"`
	Anonymous             types.Bool   `tfsdk:"anonymous"`
	LoadBalance           types.Bool   `tfsdk:"load_balance"`
	FarmConfig            types.Object `tfsdk:"farm_config"` // FarmConfig
}

func (*stfStoreServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "StoreFront StoreService.",
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "The IIS site id of the StoreFront storeservice. Defaults to 1.",
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
			"farm_config": FarmConfig{}.GetSchema(),
		},
	}
}

func (r *STFStoreServiceResourceModel) RefreshPropertyValues(storeservice *citrixstorefront.STFStoreDetailModel) {
	// Overwrite SFStoreServiceResourceModel with refreshed state
	r.VirtualPath = types.StringValue(*storeservice.VirtualPath.Get())
	r.SiteId = types.StringValue(strconv.Itoa(*storeservice.SiteId.Get()))
	r.FriendlyName = types.StringValue(*storeservice.FriendlyName.Get())
}
