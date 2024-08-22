// Copyright © 2024. Citrix Systems, Inc.

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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VsphereHypervisorClusterModel struct {
	Datacenter  types.String `tfsdk:"datacenter"`
	ClusterName types.String `tfsdk:"cluster_name"`
	Host        types.String `tfsdk:"host"`
}

func (VsphereHypervisorClusterModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Details of the cluster where resources reside and new resources will be created.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"datacenter": schema.StringAttribute{
				Description: "The name of the datacenter.",
				Required:    true,
			},
			"cluster_name": schema.StringAttribute{
				Description: "The name of the cluster.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRelative().AtParent().AtName("host"),
					}...),
				},
			},
			"host": schema.StringAttribute{
				Description: "The IP address or FQDN of the host.",
				Optional:    true,
			},
		},
	}
}

func (VsphereHypervisorClusterModel) GetAttributes() map[string]schema.Attribute {
	return VsphereHypervisorClusterModel{}.GetSchema().Attributes
}

type VsphereHypervisorResourcePoolResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Hypervisor types.String `tfsdk:"hypervisor"`
	/**** Resource Pool Details ****/
	Cluster                types.Object `tfsdk:"cluster"`           //VsphereHypervisorClusterModel
	Networks               types.List   `tfsdk:"networks"`          // List[string]
	Storage                types.List   `tfsdk:"storage"`           // List[HypervisorStorageModel]
	TemporaryStorage       types.List   `tfsdk:"temporary_storage"` // List[HypervisorStorageModel]
	UseLocalStorageCaching types.Bool   `tfsdk:"use_local_storage_caching"`
}

func (VsphereHypervisorResourcePoolResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a VMware vSphere hypervisor resource pool.",
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
			"cluster": VsphereHypervisorClusterModel{}.GetSchema(),
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
			"use_local_storage_caching": schema.BoolAttribute{
				Description: "Indicates whether intellicache is enabled to reduce load on the shared storage device. Will only be effective when shared storage is used. Default value is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (VsphereHypervisorResourcePoolResourceModel) GetAttributes() map[string]schema.Attribute {
	return VsphereHypervisorResourcePoolResourceModel{}.GetSchema().Attributes
}

func (r VsphereHypervisorResourcePoolResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) VsphereHypervisorResourcePoolResourceModel {

	r.Id = types.StringValue(resourcePool.GetId())
	r.Name = types.StringValue(resourcePool.GetName())

	hypervisorConnection := resourcePool.GetHypervisorConnection()
	r.Hypervisor = types.StringValue(hypervisorConnection.GetId())

	r.UseLocalStorageCaching = types.BoolValue(resourcePool.GetUseLocalStorageCaching())

	remoteNetwork := []string{}
	for _, network := range resourcePool.GetNetworks() {
		remoteNetwork = append(remoteNetwork, network.GetName())
	}
	r.Networks = util.RefreshListValues(ctx, diagnostics, r.Networks, remoteNetwork)
	r.Storage = util.RefreshListValueProperties[HypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](ctx, diagnostics, r.Storage, resourcePool.GetStorage(), util.GetOrchestrationHypervisorStorageKey)
	r.TemporaryStorage = util.RefreshListValueProperties[HypervisorStorageModel, citrixorchestration.HypervisorStorageResourceResponseModel](ctx, diagnostics, r.TemporaryStorage, resourcePool.GetTemporaryStorage(), util.GetOrchestrationHypervisorStorageKey)
	r = r.updatePlanWithCluster(ctx, diagnostics, resourcePool)
	return r
}

func (r VsphereHypervisorResourcePoolResourceModel) updatePlanWithCluster(ctx context.Context, diagnostics *diag.Diagnostics, resourcePool *citrixorchestration.HypervisorResourcePoolDetailResponseModel) VsphereHypervisorResourcePoolResourceModel {
	relativePath := resourcePool.RootPath.GetRelativePath()
	var vsphereHypervisorClusterModel = util.ObjectValueToTypedObject[VsphereHypervisorClusterModel](ctx, diagnostics, r.Cluster)
	if relativePath != "" {
		resourceSegments := strings.Split(relativePath, "/")
		for _, segment := range resourceSegments {
			if strings.Contains(segment, ".datacenter") {
				datacenter := strings.TrimSuffix(segment, ".datacenter")
				if !strings.EqualFold(vsphereHypervisorClusterModel.Datacenter.ValueString(), datacenter) {
					vsphereHypervisorClusterModel.Datacenter = types.StringValue(datacenter)
				}
			} else if strings.Contains(segment, ".cluster") {
				clusterName := strings.TrimSuffix(segment, ".cluster")
				if !strings.EqualFold(vsphereHypervisorClusterModel.ClusterName.ValueString(), clusterName) {
					vsphereHypervisorClusterModel.ClusterName = types.StringValue(clusterName)
				}
			} else if strings.Contains(segment, ".computeresource") {
				host := strings.TrimSuffix(segment, ".computeresource")
				if !strings.EqualFold(vsphereHypervisorClusterModel.Host.ValueString(), host) {
					vsphereHypervisorClusterModel.Host = types.StringValue(host)
				}
			}
		}
	}
	r.Cluster = util.TypedObjectToObjectValue(ctx, diagnostics, vsphereHypervisorClusterModel)
	return r
}
