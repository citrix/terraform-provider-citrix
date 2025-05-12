// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"encoding/json"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyFiltersModel struct {
	PolicyId                 types.String `tfsdk:"policy_id"`
	AccessControlFilters     types.Set    `tfsdk:"access_control_filters"`      // Set[AccessControlFilterModel]
	BranchRepeaterFilter     types.Set    `tfsdk:"branch_repeater_filter"`      // Set[BranchRepeaterFilterModel]
	ClientIPFilters          types.Set    `tfsdk:"client_ip_filters"`           // Set[ClientIPFilterModel]
	ClientNameFilters        types.Set    `tfsdk:"client_name_filters"`         // Set[ClientNameFilterModel]
	ClientPlatformFilters    types.Set    `tfsdk:"client_platform_filters"`     // Set[ClientPlatformFilterModel]
	DeliveryGroupFilters     types.Set    `tfsdk:"delivery_group_filters"`      // Set[DeliveryGroupFilterModel]
	DeliveryGroupTypeFilters types.Set    `tfsdk:"delivery_group_type_filters"` // Set[DeliveryGroupTypeFilterModel]
	OuFilters                types.Set    `tfsdk:"ou_filters"`                  // Set[OuFilterModel]
	UserFilters              types.Set    `tfsdk:"user_filters"`                // Set[UserFilterModel]
	TagFilters               types.Set    `tfsdk:"tag_filters"`                 // Set[TagFilterModel]
}

func (PolicyFiltersModel) GetDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Data source for fetching all the filters within a given policy.",
		Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				Description: "Id of the policy to look up filters.",
				Required:    true,
			},
			"access_control_filters": schema.SetNestedAttribute{
				Description:  "Set of Access Control policy filters.",
				Computed:     true,
				NestedObject: AccessControlFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"branch_repeater_filter": schema.SetNestedAttribute{
				Description:  "Set of Branch Repeater policy filters.",
				Computed:     true,
				NestedObject: BranchRepeaterFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"client_ip_filters": schema.SetNestedAttribute{
				Description:  "Set of Client IP policy filters.",
				Computed:     true,
				NestedObject: ClientIPFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"client_name_filters": schema.SetNestedAttribute{
				Description:  "Set of Client Name policy filters.",
				Computed:     true,
				NestedObject: ClientNameFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"client_platform_filters": schema.SetNestedAttribute{
				Description:  "Set of Client Platform policy filters.",
				Computed:     true,
				NestedObject: ClientPlatformFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"delivery_group_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery Group policy filters.",
				Computed:     true,
				NestedObject: DeliveryGroupFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"delivery_group_type_filters": schema.SetNestedAttribute{
				Description:  "Set of Delivery Group Type policy filters.",
				Computed:     true,
				NestedObject: DeliveryGroupTypeFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"ou_filters": schema.SetNestedAttribute{
				Description:  "Set of Organizational Unit policy filters.",
				Computed:     true,
				NestedObject: OuFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"user_filters": schema.SetNestedAttribute{
				Description:  "Set of User policy filters.",
				Computed:     true,
				NestedObject: UserFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
			"tag_filters": schema.SetNestedAttribute{
				Description:  "Set of Tag policy filters.",
				Computed:     true,
				NestedObject: TagFilterModel{}.GetDataSourceNestedAttributeObjectSchema(),
			},
		},
	}
}

func (PolicyFiltersModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return PolicyFiltersModel{}.GetDataSourceSchema().Attributes
}

func (d PolicyFiltersModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, policy citrixorchestration.PolicyResponse) PolicyFiltersModel {
	d.PolicyId = types.StringValue(policy.GetPolicyGuid())
	var accessControlFilters, branchRepeaterFilters, clientIpFilters, clientNameFilters, clientPlatformFilters, desktopGroupFilters, desktopKindFilters, desktopTagFilters, ouFilters, userFilters = ParsePolicyFilters(ctx, diags, policy)
	d.AccessControlFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, accessControlFilters)
	d.BranchRepeaterFilter = util.DataSourceTypedArrayToObjectSet(ctx, diags, branchRepeaterFilters)
	d.ClientIPFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, clientIpFilters)
	d.ClientNameFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, clientNameFilters)
	d.ClientPlatformFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, clientPlatformFilters)
	d.DeliveryGroupFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopGroupFilters)
	d.DeliveryGroupTypeFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopKindFilters)
	d.TagFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, desktopTagFilters)
	d.OuFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, ouFilters)
	d.UserFilters = util.DataSourceTypedArrayToObjectSet(ctx, diags, userFilters)
	return d
}

func ParsePolicyFilters(ctx context.Context, diags *diag.Diagnostics, policy citrixorchestration.PolicyResponse) ([]AccessControlFilterModel, []BranchRepeaterFilterModel, []ClientIPFilterModel, []ClientNameFilterModel, []ClientPlatformFilterModel, []DeliveryGroupFilterModel, []DeliveryGroupTypeFilterModel, []TagFilterModel, []OuFilterModel, []UserFilterModel) {
	var accessControlFilters []AccessControlFilterModel
	var branchRepeaterFilters []BranchRepeaterFilterModel
	var clientIpFilters []ClientIPFilterModel
	var clientNameFilters []ClientNameFilterModel
	var clientPlatformFilters []ClientPlatformFilterModel
	var desktopGroupFilters []DeliveryGroupFilterModel
	var desktopKindFilters []DeliveryGroupTypeFilterModel
	var desktopTagFilters []TagFilterModel
	var ouFilters []OuFilterModel
	var userFilters []UserFilterModel
	if policy.GetFilters() != nil && len(policy.GetFilters()) != 0 {
		for _, filter := range policy.GetFilters() {

			var uuidFilterData util.PolicyFilterUuidDataClientModel
			_ = json.Unmarshal([]byte(filter.GetFilterData()), &uuidFilterData)

			var gatewayFilterData util.PolicyFilterGatewayDataClientModel
			_ = json.Unmarshal([]byte(filter.GetFilterData()), &gatewayFilterData)

			filterType := filter.GetFilterType()
			switch filterType {
			case "AccessControl":
				accessControlFilters = append(accessControlFilters, AccessControlFilterModel{
					PolicyId:   types.StringValue(filter.GetPolicyGuid()),
					Allowed:    types.BoolValue(filter.GetIsAllowed()),
					Enabled:    types.BoolValue(filter.GetIsEnabled()),
					Connection: types.StringValue(gatewayFilterData.Connection),
					Condition:  types.StringValue(gatewayFilterData.Condition),
					Gateway:    types.StringValue(gatewayFilterData.Gateway),
				})
			case "BranchRepeater":
				branchRepeaterFilters = append(branchRepeaterFilters, BranchRepeaterFilterModel{
					PolicyId: types.StringValue(filter.GetPolicyGuid()),
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
				})
			case "ClientIP":
				clientIpFilters = append(clientIpFilters, ClientIPFilterModel{
					PolicyId:  types.StringValue(filter.GetPolicyGuid()),
					Allowed:   types.BoolValue(filter.GetIsAllowed()),
					Enabled:   types.BoolValue(filter.GetIsEnabled()),
					IpAddress: types.StringValue(filter.GetFilterData()),
				})
			case "ClientName":
				clientNameFilters = append(clientNameFilters, ClientNameFilterModel{
					PolicyId:   types.StringValue(filter.GetPolicyGuid()),
					Allowed:    types.BoolValue(filter.GetIsAllowed()),
					Enabled:    types.BoolValue(filter.GetIsEnabled()),
					ClientName: types.StringValue(filter.GetFilterData()),
				})
			case "ClientPlatform":
				clientPlatformFilters = append(clientPlatformFilters, ClientPlatformFilterModel{
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
					Enabled:  types.BoolValue(filter.GetIsEnabled()),
					Platform: types.StringValue(filter.GetFilterData()),
				})
			case "DesktopGroup":
				desktopGroupFilters = append(desktopGroupFilters, DeliveryGroupFilterModel{
					PolicyId:        types.StringValue(filter.GetPolicyGuid()),
					Allowed:         types.BoolValue(filter.GetIsAllowed()),
					Enabled:         types.BoolValue(filter.GetIsEnabled()),
					DeliveryGroupId: types.StringValue(uuidFilterData.Uuid),
				})
			case "DesktopKind":
				desktopKindFilters = append(desktopKindFilters, DeliveryGroupTypeFilterModel{
					PolicyId:          types.StringValue(filter.GetPolicyGuid()),
					Allowed:           types.BoolValue(filter.GetIsAllowed()),
					Enabled:           types.BoolValue(filter.GetIsEnabled()),
					DeliveryGroupType: types.StringValue(filter.GetFilterData()),
				})
			case "DesktopTag":
				desktopTagFilters = append(desktopTagFilters, TagFilterModel{
					PolicyId: types.StringValue(filter.GetPolicyGuid()),
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
					Enabled:  types.BoolValue(filter.GetIsEnabled()),
					Tag:      types.StringValue(uuidFilterData.Uuid),
				})
			case "OU":
				ouFilters = append(ouFilters, OuFilterModel{
					PolicyId: types.StringValue(filter.GetPolicyGuid()),
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
					Enabled:  types.BoolValue(filter.GetIsEnabled()),
					Ou:       types.StringValue(filter.GetFilterData()),
				})
			case "User":
				userFilters = append(userFilters, UserFilterModel{
					PolicyId: types.StringValue(filter.GetPolicyGuid()),
					Allowed:  types.BoolValue(filter.GetIsAllowed()),
					Enabled:  types.BoolValue(filter.GetIsEnabled()),
					UserSid:  types.StringValue(filter.GetFilterData()),
				})
			}
		}
	}
	return accessControlFilters, branchRepeaterFilters, clientIpFilters, clientNameFilters, clientPlatformFilters, desktopGroupFilters, desktopKindFilters, desktopTagFilters, ouFilters, userFilters
}
