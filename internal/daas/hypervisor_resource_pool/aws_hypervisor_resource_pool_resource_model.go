// Copyright © 2023. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Vpc     types.String `tfsdk:"vpc"`
	Subnets types.List   `tfsdk:"subnets"` // List[string]
	/** AWS Resource Pool **/
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

func GetAwsHypervisorResourcePoolSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an AWS EC2 hypervisor resource pool.",
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
				Description: "Name of the virtual private cloud.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of subnets to allocate VDAs within the virtual private cloud.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"availability_zone": schema.StringAttribute{
				Description: "The name of the availability zone resource to use for provisioning operations in this resource pool.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

func (r AwsHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) AwsHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	virtualNetwork := resourcePool.GetVirtualPrivateCloud()
	r.Vpc = types.StringValue(virtualNetwork.GetName())
	availabilityZone := resourcePool.GetAvailabilityZone()
	r.AvailabilityZone = types.StringValue(strings.Split(availabilityZone.GetName(), " ")[0])
	var res []string
	for _, model := range resourcePool.GetNetworks() {
		name := model.GetName()
		res = append(res, strings.Split(name, " ")[0])
	}
	r.Subnets = util.RefreshListValues(ctx, diagnostics, r.Subnets, res)

	return r
}
