// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	UseSystemProxyForHypervisorTrafficOnConnectors_CustomProperty = "UseSystemProxyForHypervisorTrafficOnConnectors"
)

// HypervisorResourceModel maps the resource schema data.
type AmazonWorkSpacesCoreHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id                                             types.String `tfsdk:"id"`
	Name                                           types.String `tfsdk:"name"`
	Zone                                           types.String `tfsdk:"zone"`
	Scopes                                         types.Set    `tfsdk:"scopes"`   // Set[string]
	Metadata                                       types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Tenants                                        types.Set    `tfsdk:"tenants"`  // Set[string]
	Region                                         types.String `tfsdk:"region"`
	ApiKey                                         types.String `tfsdk:"api_key"`
	SecretKey                                      types.String `tfsdk:"secret_key"`
	UseIamRole                                     types.Bool   `tfsdk:"use_iam_role"`
	UseSystemProxyForHypervisorTrafficOnConnectors types.Bool   `tfsdk:"use_system_proxy_for_hypervisor_traffic_on_connectors"`
}

func (AmazonWorkSpacesCoreHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an Amazon WorkSpaces Core hypervisor. Note that this feature is in Tech Preview.",
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
			"region": schema.StringAttribute{
				Description: "AWS region to connect to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"api_key": schema.StringAttribute{
				Description: "The API key used to authenticate with the AWS APIs.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(
						path.MatchRoot("secret_key"),
					),
					stringvalidator.NoneOfCaseInsensitive(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret),
				},
			},
			"secret_key": schema.StringAttribute{
				Description: "The secret key used to authenticate with the AWS APIs.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.NoneOfCaseInsensitive(util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret),
				},
			},
			"use_iam_role": schema.BoolAttribute{
				Description: "When set to `true`, the provider will use the IAM role configured on the Citrix Cloud Connector or Delivery Controller instead of the `api_key` and `secret_key` for authentication. Omit this attribute if you want to use `api_key` and `secret_key` for authentication. Default value is `false`. ",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
			"use_system_proxy_for_hypervisor_traffic_on_connectors": schema.BoolAttribute{
				Description: "When set to `true`, the hypervisor connection will be setup with the proxy configured during connector installation. Default value is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (AmazonWorkSpacesCoreHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return AmazonWorkSpacesCoreHypervisorResourceModel{}.GetSchema().Attributes
}

func (AmazonWorkSpacesCoreHypervisorResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{
		"api_key": true,
	}
}

func (r AmazonWorkSpacesCoreHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) AmazonWorkSpacesCoreHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.Region = types.StringValue(hypervisor.GetRegion())
	if strings.EqualFold(hypervisor.GetApiKey(), util.AmazonWorkSpacesCoreRoleBasedAuthKeyAndSecret) {
		r.UseIamRole = types.BoolValue(true)
	} else {
		r.UseIamRole = types.BoolValue(false)
		r.ApiKey = types.StringValue(hypervisor.GetApiKey())
	}
	scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), hypervisor.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.UseSystemProxyForHypervisorTrafficOnConnectors = types.BoolValue(false)

	customPropertiesString := hypervisor.GetCustomProperties()
	if customPropertiesString != "" {
		var customProperties []citrixorchestration.NameValueStringPairModel
		err := json.Unmarshal([]byte(customPropertiesString), &customProperties)
		if err == nil {
			for _, customProperty := range customProperties {
				if customProperty.GetName() == UseSystemProxyForHypervisorTrafficOnConnectors_CustomProperty {
					proxy, _ := strconv.ParseBool(customProperty.GetValue())
					r.UseSystemProxyForHypervisorTrafficOnConnectors = types.BoolValue(proxy)
				}
			}
		} else {
			diagnostics.AddWarning("Error reading AWS WorkSpaces Core Hypervisor custom properties", err.Error())
		}
	}

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, hypervisor.GetTenants())

	return r
}
