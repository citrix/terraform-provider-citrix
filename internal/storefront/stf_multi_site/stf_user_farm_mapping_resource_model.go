// Copyright © 2024. Citrix Systems, Inc.
package stf_multi_site

import (
	"context"
	"regexp"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensure UserFarmMappingGroup implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixstorefront.STFGroupMemberResponseModel] = UserFarmMappingGroup{}

type UserFarmMappingGroup struct {
	GroupName  types.String `tfsdk:"group_name"`
	AccountSid types.String `tfsdk:"account_sid"`
}

func (UserFarmMappingGroup) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"group_name": schema.StringAttribute{
				Description: "A display only group name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.NoneOfCaseInsensitive("everyone"),
				},
			},
			"account_sid": schema.StringAttribute{
				Description: "Sid of the account.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.ActiveDirectorySidRegex), "must be in Active Directory SID format"),
				},
			},
		},
	}
}

func (UserFarmMappingGroup) GetAttributes() map[string]schema.Attribute {
	return UserFarmMappingGroup{}.GetSchema().Attributes
}

func (r UserFarmMappingGroup) GetKey() string {
	return r.GroupName.ValueString()
}

func (r UserFarmMappingGroup) RefreshListItem(_ context.Context, _ *diag.Diagnostics, item citrixstorefront.STFGroupMemberResponseModel) util.ResourceModelWithAttributes {
	// Implement the logic to refresh the list item based on the item
	groupName := types.StringValue(*item.GroupName.Get())
	r.GroupName = groupName

	accountSid := types.StringValue(*item.AccountSid.Get())
	r.AccountSid = accountSid

	return r
}

// ensure EquivalentFarmSet implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixstorefront.STFFarmSetResponseModel] = EquivalentFarmSet{}

type EquivalentFarmSet struct {
	Name                 types.String `tfsdk:"name"`
	AggregationGroupName types.String `tfsdk:"aggregation_group_name"`
	PrimaryFarms         types.List   `tfsdk:"primary_farms"`     // List[string]
	BackupFarms          types.List   `tfsdk:"backup_farms"`      // List[string]
	LoadBalanceMode      types.String `tfsdk:"load_balance_mode"` // Failover or LoadBalanced
	FarmsAreIdentical    types.Bool   `tfsdk:"farms_are_identical"`
}

func (EquivalentFarmSet) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique Name used to identify the EquivalentFarmSet.",
				Required:    true,
			},
			"aggregation_group_name": schema.StringAttribute{
				Description: "The AggregationGroupName used to de-duplicate applications and desktops that are available on multiple EquivalentFarmSets. Where multiple EquivalentFarmSets are defined the AggregationGroup will prevent the user seeing the application multiple times if it exists in both places.",
				Required:    true,
			},
			"primary_farms": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The PrimaryFarms. The farm names should match those defined in the Store service.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"backup_farms": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "The BackupFarms. The farm names should match those defined in the Store Service.",
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"load_balance_mode": schema.StringAttribute{
				Description: "The load balance mode, either `Failover` or `LoadBalanced`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Failover",
						"LoadBalanced",
					),
				},
			},
			"farms_are_identical": schema.BoolAttribute{
				Description: "Whether the PrimaryFarms in the EquivalentFarmSet all publish identical resources. Set to true if all resources are identical on all primary farms. Set to false if the deployment has some unique resources per farm. Default to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (EquivalentFarmSet) GetAttributes() map[string]schema.Attribute {
	return EquivalentFarmSet{}.GetSchema().Attributes
}

func (r EquivalentFarmSet) GetKey() string {
	return r.Name.ValueString()
}

func (r EquivalentFarmSet) RefreshListItem(ctx context.Context, diagnostics *diag.Diagnostics, item citrixstorefront.STFFarmSetResponseModel) util.ResourceModelWithAttributes {
	// Implement the logic to refresh the list item based on the item
	name := types.StringValue(*item.Name.Get())
	r.Name = name

	aggregationGroupName := types.StringValue(*item.AggregationGroupName.Get())
	r.AggregationGroupName = aggregationGroupName

	primaryFarms := util.RefreshListValues(ctx, diagnostics, r.PrimaryFarms, item.PrimaryFarms)
	r.PrimaryFarms = primaryFarms

	backupFarms := util.RefreshListValues(ctx, diagnostics, r.BackupFarms, item.BackupFarms)
	r.BackupFarms = backupFarms

	LoadBalanceMode := types.StringValue(*item.LoadBalanceMode.Get())
	r.LoadBalanceMode = LoadBalanceMode

	farmsAreIdentical := types.BoolValue(*item.FarmsAreIdentical.Get())
	r.FarmsAreIdentical = farmsAreIdentical

	return r
}

type STFUserFarmMappingResourceModel struct {
	VirtualPath        types.String `tfsdk:"store_virtual_path"`
	Name               types.String `tfsdk:"name"`
	GroupMembers       types.List   `tfsdk:"group_members"`        // List of UserFarmMappingGroup
	EquivalentFarmSets types.List   `tfsdk:"equivalent_farm_sets"` // List of EquivalentFarmSets
}

// Map response body to schema and populate Computed attribute values
func (r *STFUserFarmMappingResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, result citrixstorefront.STFUserFarmMappingResponseModel) {
	// Implement the logic to refresh the property values based on the result
	r.Name = types.StringValue(*result.Name.Get())
	r.VirtualPath = types.StringValue(*result.VirtualPath.Get())

	if result.GroupMembers != nil && len(result.GroupMembers) > 0 && strings.ToLower(*result.GroupMembers[0].GroupName.Get()) != "everyone" && strings.ToLower(*result.GroupMembers[0].AccountSid.Get()) != "everyone" {
		updatedGroupMembers := util.RefreshListValueProperties[UserFarmMappingGroup, citrixstorefront.STFGroupMemberResponseModel](ctx, diagnostics, r.GroupMembers, result.GroupMembers, util.GetSTFGroupMemberKey)
		r.GroupMembers = updatedGroupMembers
	} else {
		if attributeMap, err := util.ResourceAttributeMapFromObject(UserFarmMappingGroup{}); err == nil {
			r.GroupMembers = types.ListNull(types.ObjectType{AttrTypes: attributeMap})
		} else {
			diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		}
	}

	updatedEquivalentFarmSets := util.TypedArrayToObjectList[EquivalentFarmSet](ctx, diagnostics, []EquivalentFarmSet{})
	if result.FarmSets != nil && len(result.FarmSets) > 0 {
		updatedEquivalentFarmSets = util.RefreshListValueProperties[EquivalentFarmSet, citrixstorefront.STFFarmSetResponseModel](ctx, diagnostics, r.EquivalentFarmSets, result.FarmSets, util.GetSTFFarmSetKey)
	}
	r.EquivalentFarmSets = updatedEquivalentFarmSets
}

func (STFUserFarmMappingResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "StoreFront --- StoreFront User Farm Mapping Resource",
		Attributes: map[string]schema.Attribute{
			"store_virtual_path": schema.StringAttribute{
				Description: "The IIS VirtualPath at which the Store is configured to be accessed by Receivers.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "The unique name used to identify the UserFarmMapping.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_members": schema.ListNestedAttribute{
				Description: "The Windows groups to which the UserFarmMapping will apply. Not specifying this field will assign all users to the UserFarmMapping.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: UserFarmMappingGroup{}.GetSchema(),
			},
			"equivalent_farm_sets": schema.ListNestedAttribute{
				Description: "Configurations of the EquivalentFarmSets that will be assigned to the UserFarmMapping.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: EquivalentFarmSet{}.GetSchema(),
			},
		},
	}
}

func (STFUserFarmMappingResourceModel) GetAttributes() map[string]schema.Attribute {
	return STFUserFarmMappingResourceModel{}.GetSchema().Attributes
}
