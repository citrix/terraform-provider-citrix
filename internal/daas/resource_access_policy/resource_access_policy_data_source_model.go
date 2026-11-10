// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy

import (
	"context"
	"strconv"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceAccessPolicyDataSourceModel struct {
	Id                                  types.String `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	Description                         types.String `tfsdk:"description"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	TagId                               types.String `tfsdk:"tag_id"`
	TagName                             types.String `tfsdk:"tag_name"`
	ResourceType                        types.String `tfsdk:"resource_type"`
	EnableCriteriaForIncludeConnections types.Bool   `tfsdk:"enable_criteria_for_include_connections"`
	IncludeConnectionsCriteriaType      types.String `tfsdk:"include_connections_criteria_type"`
	IncludeCriteriaFilters              types.List   `tfsdk:"include_criteria_filters"` // List[util.SmartAccessTagModel]
	EnableCriteriaForExcludeConnections types.Bool   `tfsdk:"enable_criteria_for_exclude_connections"`
	ExcludeCriteriaFilters              types.List   `tfsdk:"exclude_criteria_filters"` // List[util.SmartAccessTagModel]
	Metadata                            types.List   `tfsdk:"metadata"`                 // List[util.NameValueStringPairModel]
}

func (ResourceAccessPolicyDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing site-wide resource access policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the resource access policy. Exactly one of `id` and `name` must be specified.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource access policy. Exactly one of `id` and `name` must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the resource access policy.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the resource access policy is enabled.",
				Computed:    true,
			},
			"tag_id": schema.StringAttribute{
				Description: "GUID identifier of the tag whose resources the policy applies to.",
				Computed:    true,
			},
			"tag_name": schema.StringAttribute{
				Description: "Name of the tag whose resources the policy applies to.",
				Computed:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Type of resource the policy applies to.",
				Computed:    true,
			},
			"enable_criteria_for_include_connections": schema.BoolAttribute{
				Description: "Whether the criteria for include connections are enabled.",
				Computed:    true,
			},
			"include_connections_criteria_type": schema.StringAttribute{
				Description: "Whether a connection must match all (`MatchAll`) or any (`MatchAny`) of the `include_criteria_filters` to be granted access.",
				Computed:    true,
			},
			"include_criteria_filters": schema.ListNestedAttribute{
				Description:  "The SmartAccess filters a connection must match to be granted access.",
				Computed:     true,
				NestedObject: util.SmartAccessTagModel{}.GetDataSourceSchema(),
			},
			"enable_criteria_for_exclude_connections": schema.BoolAttribute{
				Description: "Whether the criteria for exclude connections are enabled.",
				Computed:    true,
			},
			"exclude_criteria_filters": schema.ListNestedAttribute{
				Description:  "The SmartAccess filters which deny access to a connection that matches any of them.",
				Computed:     true,
				NestedObject: util.SmartAccessTagModel{}.GetDataSourceSchema(),
			},
			"metadata": schema.ListNestedAttribute{
				Description:  "Metadata of the resource access policy.",
				Computed:     true,
				NestedObject: util.NameValueStringPairModel{}.GetDataSourceSchema(),
			},
		},
	}
}

func (ResourceAccessPolicyDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return ResourceAccessPolicyDataSourceModel{}.GetSchema().Attributes
}

func (r ResourceAccessPolicyDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, policy *citrixorchestration.ResourceAccessPolicyResponseModel, tagId string) ResourceAccessPolicyDataSourceModel {
	r.Id = types.StringValue(strconv.Itoa(int(policy.GetId())))
	r.Name = types.StringValue(policy.GetName())
	r.Description = types.StringValue(policy.GetDescription())
	r.Enabled = types.BoolValue(policy.GetEnabled())

	if tagId == "" {
		r.TagId = types.StringNull()
		r.TagName = types.StringNull()
	} else {
		r.TagId = types.StringValue(tagId)
		r.TagName = types.StringValue(policy.GetTagName())
	}

	if policy.GetResourceType() == "" || policy.GetResourceType() == citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_UNKNOWN {
		r.ResourceType = types.StringNull()
	} else {
		r.ResourceType = types.StringValue(string(policy.GetResourceType()))
	}

	r.EnableCriteriaForIncludeConnections = types.BoolValue(policy.GetIncludedSmartAccessFilterEnabled())
	if policy.GetIncludedSmartAccessFilterEnabled() {
		r.IncludeConnectionsCriteriaType = types.StringValue(string(policy.GetIncludedSmartAccessFilterType()))
	} else {
		r.IncludeConnectionsCriteriaType = types.StringNull()
	}
	r.IncludeCriteriaFilters = util.DataSourceTypedArrayToObjectList[util.SmartAccessTagModel](ctx, diagnostics, util.ParseSmartAccessTagsToPluginModel(policy.GetIncludedSmartAccessTags()))

	r.EnableCriteriaForExcludeConnections = types.BoolValue(policy.GetExcludedSmartAccessFilterEnabled())
	r.ExcludeCriteriaFilters = util.DataSourceTypedArrayToObjectList[util.SmartAccessTagModel](ctx, diagnostics, util.ParseSmartAccessTagsToPluginModel(policy.GetExcludedSmartAccessTags()))

	r.Metadata = util.DataSourceTypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, util.ParseNameValueStringPairToPluginModel(policy.GetMetadata()))

	return r
}
