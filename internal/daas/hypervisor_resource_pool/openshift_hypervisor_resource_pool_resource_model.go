// Copyright © 2024. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"context"
	"regexp"

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

type OpenShiftHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Namespace        types.String `tfsdk:"namespace"`
	Networks         types.List   `tfsdk:"networks"`          // List[string]
	Storage          types.List   `tfsdk:"storage"`           // List[HypervisorStorageModel]
	TemporaryStorage types.List   `tfsdk:"temporary_storage"` // List[HypervisorStorageModel]
}

func (OpenShiftHypervisorResourcePoolResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a OpenShift hypervisor resource pool.",
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
			"namespace": schema.StringAttribute{
				Description: "Namespace for the connection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"networks": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Networks for allocating resources.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"storage": schema.ListNestedAttribute{
				Description:  "Storage resources to use for OS data.",
				Required:     true,
				NestedObject: HypervisorStorageModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"temporary_storage": schema.ListNestedAttribute{
				Description:  "Storage resources to use for temporary data.",
				Required:     true,
				NestedObject: HypervisorStorageModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (OpenShiftHypervisorResourcePoolResourceModel) GetAttributes() map[string]schema.Attribute {
	return OpenShiftHypervisorResourcePoolResourceModel{}.GetSchema().Attributes
}

func (r OpenShiftHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) OpenShiftHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	rootPath := resourcePool.GetRootPath()
	r.Namespace = types.StringValue(rootPath.GetName())
	remoteNetwork := []string{}
	for _, network := range resourcePool.GetNetworks() {
		remoteNetwork = append(remoteNetwork, network.GetName())
	}
	r.Networks = util.RefreshListValues(ctx, diagnostics, r.Networks, remoteNetwork)
	r.Storage = util.RefreshListValueProperties[HypervisorStorageModel](ctx, diagnostics, r.Storage, resourcePool.GetStorage(), util.GetOrchestrationHypervisorStorageKey)
	r.TemporaryStorage = util.RefreshListValueProperties[HypervisorStorageModel](ctx, diagnostics, r.TemporaryStorage, resourcePool.GetTemporaryStorage(), util.GetOrchestrationHypervisorStorageKey)

	return r
}
