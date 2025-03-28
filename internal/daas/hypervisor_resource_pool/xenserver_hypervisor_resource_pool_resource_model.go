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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type XenserverHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	Metadata   types.List   `tfsdk:"metadata"` //List[NameValueStringPairModel]
	VmTagging  types.Bool   `tfsdk:"vm_tagging"`
	/**** Resource Pool Details ****/
	Networks               types.List `tfsdk:"networks"`          //List[string]
	Storage                types.List `tfsdk:"storage"`           //List[HypervisorStorageModel]
	TemporaryStorage       types.List `tfsdk:"temporary_storage"` //List[HypervisorStorageModel]
	UseLocalStorageCaching types.Bool `tfsdk:"use_local_storage_caching"`
}

func (XenserverHypervisorResourcePoolResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an XenServer hypervisor resource pool.",
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
			"networks": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Networks for allocating resources.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"storage": schema.ListNestedAttribute{
				Description:  "List of hypervisor storage to use for OS data.",
				Required:     true,
				NestedObject: HypervisorStorageModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"temporary_storage": schema.ListNestedAttribute{
				Description:  "List of hypervisor storage to use for temporary data.",
				Required:     true,
				NestedObject: HypervisorStorageModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"use_local_storage_caching": schema.BoolAttribute{
				Description: "Indicates whether intellicache is enabled to reduce load on the shared storage device. Will only be effective when shared storage is used. Default value is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"metadata": util.GetMetadataListSchema("Hypervisor Resource Pool"),
			"vm_tagging": schema.BoolAttribute{
				Description: "Indicates whether VMs created by provisioning operations should be tagged. Default is `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (XenserverHypervisorResourcePoolResourceModel) GetAttributes() map[string]schema.Attribute {
	return XenserverHypervisorResourcePoolResourceModel{}.GetSchema().Attributes
}

func (r XenserverHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) XenserverHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())
	r.VmTagging = types.BoolValue(resourcePool.GetVMTaggingEnabled())

	r.UseLocalStorageCaching = types.BoolValue(resourcePool.GetUseLocalStorageCaching())

	remoteNetwork := []string{}
	for _, network := range resourcePool.GetNetworks() {
		remoteNetwork = append(remoteNetwork, network.GetName())
	}
	r.Networks = util.RefreshListValues(ctx, diagnostics, r.Networks, remoteNetwork)
	r.Storage = util.RefreshListValueProperties[HypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](ctx, diagnostics, r.Storage, resourcePool.GetStorage(), util.GetOrchestrationHypervisorStorageKey)
	r.TemporaryStorage = util.RefreshListValueProperties[HypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](ctx, diagnostics, r.TemporaryStorage, resourcePool.GetTemporaryStorage(), util.GetOrchestrationHypervisorStorageKey)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), resourcePool.GetMetadata())
	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	return r
}
