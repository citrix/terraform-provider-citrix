// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type GcpHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Zone     types.String `tfsdk:"zone"`
	Scopes   types.Set    `tfsdk:"scopes"`   // Set[string]
	Metadata types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Tenants  types.Set    `tfsdk:"tenants"`  // Set[string]
	/** GCP Connection **/
	ServiceAccountId          types.String `tfsdk:"service_account_id"`
	ServiceAccountCredentials types.String `tfsdk:"service_account_credentials"`
}

func (GcpHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a GCP hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
				Required:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the hypervisor is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"service_account_id": schema.StringAttribute{
				Description: "The service account ID used to access the Google Cloud APIs.",
				Required:    true,
			},
			"service_account_credentials": schema.StringAttribute{
				Description: "The JSON-encoded service account credentials used to access the Google Cloud APIs.",
				Required:    true,
				Sensitive:   true,
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the hypervisor to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"metadata": util.GetMetadataListSchema("Hypervisor"),
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the hypervisor connection.",
				Computed:    true,
			},
		},
	}
}

func (GcpHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return GcpHypervisorResourceModel{}.GetSchema().Attributes
}

func (GcpHypervisorResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{
		"service_account_id": true,
	}
}

func (r GcpHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) GcpHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.ServiceAccountId = types.StringValue(hypervisor.GetServiceAccountId())
	scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), hypervisor.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, hypervisor.GetTenants())

	return r
}
