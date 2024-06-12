// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GcpHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Region  types.String `tfsdk:"region"`
	Vpc     types.String `tfsdk:"vpc"`
	Subnets types.List   `tfsdk:"subnets"` // List[string]
	/** GCP Resource Pool **/
	ProjectName types.String `tfsdk:"project_name"`
	SharedVpc   types.Bool   `tfsdk:"shared_vpc"`
}

func GetGcpHypervisorResourcePoolSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a GCP hypervisor resource pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the resource pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource pool. Name should be unique across all hypervisors.",
				Required:    true,
			},
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor for which the resource pool needs to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"vpc": schema.StringAttribute{
				Description: "Name of the cloud virtual network.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Subnets to allocate VDAs within the virtual network.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Description: "Cloud Region where the virtual network sits in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = !req.ConfigValue.IsNull() && !req.StateValue.IsNull() &&
								(location.Normalize(req.ConfigValue.ValueString()) != location.Normalize(req.StateValue.ValueString()))
						},
						"Force replacement when region changes, unless changing between GCP region name (East US) and Id (eastus)",
						"Force replacement when region changes, unless changing between GCP region name (East US) and Id (eastus)",
					),
				},
			},
			"project_name": schema.StringAttribute{
				Description: "GCP Project name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"shared_vpc": schema.BoolAttribute{
				Description: "Indicate whether the GCP Virtual Private Cloud is a shared VPC.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

func (r GcpHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) GcpHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	region := resourcePool.GetRegion()
	if r.shouldSetRegion(region) {
		r.Region = types.StringValue(region.GetName())
	}
	project := resourcePool.GetProject()
	r.ProjectName = types.StringValue(project.GetName())
	vpc := resourcePool.GetVirtualPrivateCloud()
	r.Vpc = types.StringValue(vpc.GetName())
	var res []string
	for _, model := range resourcePool.GetNetworks() {
		res = append(res, model.GetName())
	}
	r.Subnets = util.RefreshListValues(ctx, diagnostics, r.Subnets, res)

	vpcType := vpc.GetObjectTypeName()
	if vpcType == "sharedvirtualprivatecloud" {
		r.SharedVpc = types.BoolValue(true)
	}

	return r
}

func (r GcpHypervisorResourcePoolResourceModel) shouldSetRegion(region citrixorchestration.HypervisorResourceRefResponseModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.IsNull() || r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}
