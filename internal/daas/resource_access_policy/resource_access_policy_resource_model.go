// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceAccessPolicyResourceModel struct {
	Id                                  types.String `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	Description                         types.String `tfsdk:"description"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	TagId                               types.String `tfsdk:"tag_id"`
	ResourceType                        types.String `tfsdk:"resource_type"`
	EnableCriteriaForIncludeConnections types.Bool   `tfsdk:"enable_criteria_for_include_connections"`
	IncludeConnectionsCriteriaType      types.String `tfsdk:"include_connections_criteria_type"`
	IncludeCriteriaFilters              types.List   `tfsdk:"include_criteria_filters"` // List[util.SmartAccessTagModel]
	EnableCriteriaForExcludeConnections types.Bool   `tfsdk:"enable_criteria_for_exclude_connections"`
	ExcludeCriteriaFilters              types.List   `tfsdk:"exclude_criteria_filters"` // List[util.SmartAccessTagModel]
	Metadata                            types.List   `tfsdk:"metadata"`                 // List[util.NameValueStringPairModel]
}

// The update API ignores an omitted tag or resource type, so neither can be cleared in place.
func requiresReplaceWhenRemoved(attributeName string) planmodifier.String {
	description := fmt.Sprintf("Force replacement when `%s` is removed", attributeName)
	return stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		resp.RequiresReplace = !req.StateValue.IsNull() && req.ConfigValue.IsNull()
	}, description, description)
}

func (ResourceAccessPolicyResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages a site-wide resource access policy. A resource access policy restricts access to the published resources that carry a tag, " +
			"are of a resource type, or both, across all delivery groups, based on the SmartAccess tags of the user's connection." +
			"\n\nFor a user to gain access to a resource through a policy, their connection must match all of its enabled include criteria and none of its enabled exclude criteria.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the resource access policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource access policy. Must be unique within the site.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the resource access policy. Defaults to an empty string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the resource access policy is enabled. A disabled policy is not considered when evaluating resource access. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"tag_id": schema.StringAttribute{
				Description: "GUID identifier of the tag whose resources the policy applies to. At least one of `tag_id` and `resource_type` must be specified." +
					"\n\n~> **Please Note** Removing `tag_id` forces the policy to be replaced.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					requiresReplaceWhenRemoved("tag_id"),
				},
			},
			"resource_type": schema.StringAttribute{
				Description: "Type of resource the policy applies to. At least one of `tag_id` and `resource_type` must be specified. Choose between `Application`, `Desktop`, and `Content`." +
					"\n\n~> **Please Note** Removing `resource_type` forces the policy to be replaced.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_APPLICATION),
						string(citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_DESKTOP),
						string(citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_CONTENT),
					),
				},
				PlanModifiers: []planmodifier.String{
					requiresReplaceWhenRemoved("resource_type"),
				},
			},
			"enable_criteria_for_include_connections": schema.BoolAttribute{
				Description: "Whether to enable the criteria for include connections. When enabled, `include_connections_criteria_type` and `include_criteria_filters` must be specified. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"include_connections_criteria_type": schema.StringAttribute{
				Description: "Whether a connection must match all (`MatchAll`) or any (`MatchAny`) of the `include_criteria_filters` to be granted access. Can only be specified when `enable_criteria_for_include_connections` is `true`. Choose between `MatchAny` and `MatchAll`.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.FILTERMATCHTYPE_MATCH_ANY),
						string(citrixorchestration.FILTERMATCHTYPE_MATCH_ALL),
					),
				},
			},
			"include_criteria_filters": schema.ListNestedAttribute{
				Description:  "The SmartAccess filters a connection must match to be granted access. Can only be specified when `enable_criteria_for_include_connections` is `true`.",
				Optional:     true,
				NestedObject: util.SmartAccessTagModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"enable_criteria_for_exclude_connections": schema.BoolAttribute{
				Description: "Whether to enable the criteria for exclude connections. When enabled, `exclude_criteria_filters` must be specified. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"exclude_criteria_filters": schema.ListNestedAttribute{
				Description:  "The SmartAccess filters which deny access to a connection that matches any of them. Can only be specified when `enable_criteria_for_exclude_connections` is `true`.",
				Optional:     true,
				NestedObject: util.SmartAccessTagModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"metadata": util.GetMetadataListSchema("Resource Access Policy"),
		},
	}
}

func (ResourceAccessPolicyResourceModel) GetAttributes() map[string]schema.Attribute {
	return ResourceAccessPolicyResourceModel{}.GetSchema().Attributes
}

func (ResourceAccessPolicyResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

// Timeout in minutes for the create, update, and delete jobs
const resourceAccessPolicyJobTimeout int32 = 5

// ValidateConfig checks the constraints between attributes. Unknown values are skipped
// since they cannot be checked until apply.
func (r ResourceAccessPolicyResourceModel) ValidateConfig(ctx context.Context, diagnostics *diag.Diagnostics) {
	if r.TagId.IsNull() && r.ResourceType.IsNull() {
		diagnostics.AddAttributeError(
			path.Root("tag_id"),
			"Incorrect Attribute Configuration",
			"At least one of tag_id and resource_type must be specified.",
		)
	}

	if !r.EnableCriteriaForIncludeConnections.IsUnknown() {
		if r.EnableCriteriaForIncludeConnections.ValueBool() {
			requireAttribute(diagnostics, r.IncludeConnectionsCriteriaType.IsNull(), "include_connections_criteria_type", "enable_criteria_for_include_connections")
			requireAttribute(diagnostics, r.IncludeCriteriaFilters.IsNull(), "include_criteria_filters", "enable_criteria_for_include_connections")
		} else {
			forbidAttribute(diagnostics, !r.IncludeConnectionsCriteriaType.IsNull(), "include_connections_criteria_type", "enable_criteria_for_include_connections")
			forbidAttribute(diagnostics, !r.IncludeCriteriaFilters.IsNull(), "include_criteria_filters", "enable_criteria_for_include_connections")
		}
	}

	if !r.EnableCriteriaForExcludeConnections.IsUnknown() {
		if r.EnableCriteriaForExcludeConnections.ValueBool() {
			requireAttribute(diagnostics, r.ExcludeCriteriaFilters.IsNull(), "exclude_criteria_filters", "enable_criteria_for_exclude_connections")
		} else {
			forbidAttribute(diagnostics, !r.ExcludeCriteriaFilters.IsNull(), "exclude_criteria_filters", "enable_criteria_for_exclude_connections")
		}
	}

	validateCriteriaFilters(ctx, diagnostics, r.IncludeCriteriaFilters, "include_criteria_filters")
	validateCriteriaFilters(ctx, diagnostics, r.ExcludeCriteriaFilters, "exclude_criteria_filters")

	if !r.Metadata.IsNull() && !r.Metadata.IsUnknown() {
		util.ValidateMetadataConfig(ctx, diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata))
	}
}

func requireAttribute(diagnostics *diag.Diagnostics, isMissing bool, attributeName string, enableAttributeName string) {
	if isMissing {
		diagnostics.AddAttributeError(
			path.Root(attributeName),
			"Missing Attribute Configuration",
			fmt.Sprintf("%s must be specified when %s is set to true.", attributeName, enableAttributeName),
		)
	}
}

func forbidAttribute(diagnostics *diag.Diagnostics, isSet bool, attributeName string, enableAttributeName string) {
	if isSet {
		diagnostics.AddAttributeError(
			path.Root(attributeName),
			"Incorrect Attribute Configuration",
			fmt.Sprintf("%s cannot be specified when %s is set to false.", attributeName, enableAttributeName),
		)
	}
}

func validateCriteriaFilters(ctx context.Context, diagnostics *diag.Diagnostics, filters types.List, attributeName string) {
	if filters.IsNull() || filters.IsUnknown() {
		return
	}

	seen := []string{}
	for index, element := range filters.Elements() {
		// A filter whose object is unknown until apply would otherwise convert to a filter with null fields
		filterObject, ok := element.(types.Object)
		if !ok || filterObject.IsUnknown() {
			continue
		}

		filter := util.ObjectValueToTypedObject[util.SmartAccessTagModel](ctx, diagnostics, filterObject)
		if filter.FilterName.IsUnknown() || filter.FilterValue.IsUnknown() {
			continue
		}

		key := strings.ToLower(filter.FilterName.ValueString() + ":" + filter.FilterValue.ValueString())
		if slices.Contains(seen, key) {
			diagnostics.AddAttributeError(
				path.Root(attributeName).AtListIndex(index),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Filter name %s with filter value %s is specified more than once.", filter.FilterName.ValueString(), filter.FilterValue.ValueString()),
			)
			continue
		}
		seen = append(seen, key)
	}
}

// buildRequest creates the request body for both create and update. The update API replaces the
// filter lists and metadata whenever they are sent, so they are always sent, empty when removed.
// The tag and resource type are only sent when set, since removing them forces replacement.
func (r ResourceAccessPolicyResourceModel) buildRequest(ctx context.Context, diagnostics *diag.Diagnostics, tagUid *int32) citrixorchestration.ResourceAccessPolicyRequestModel {
	var body citrixorchestration.ResourceAccessPolicyRequestModel
	body.SetName(r.Name.ValueString())
	body.SetDescription(r.Description.ValueString())
	body.SetEnabled(r.Enabled.ValueBool())

	if tagUid != nil {
		body.SetTagUid(*tagUid)
	}

	if r.ResourceType.ValueString() != "" {
		body.SetResourceType(citrixorchestration.ResourceAccessPolicyResourceType(r.ResourceType.ValueString()))
	}

	body.SetIncludedSmartAccessFilterEnabled(r.EnableCriteriaForIncludeConnections.ValueBool())
	if r.IncludeConnectionsCriteriaType.ValueString() != "" {
		body.SetIncludedSmartAccessFilterType(citrixorchestration.FilterMatchType(r.IncludeConnectionsCriteriaType.ValueString()))
	}
	body.SetIncludedSmartAccessTags(util.GetSmartAccessTagsRequestModel(util.ObjectListToTypedArray[util.SmartAccessTagModel](ctx, diagnostics, r.IncludeCriteriaFilters)))

	body.SetExcludedSmartAccessFilterEnabled(r.EnableCriteriaForExcludeConnections.ValueBool())
	body.SetExcludedSmartAccessTags(util.GetSmartAccessTagsRequestModel(util.ObjectListToTypedArray[util.SmartAccessTagModel](ctx, diagnostics, r.ExcludeCriteriaFilters)))

	body.SetMetadata(util.GetMetadataRequestModel(ctx, diagnostics, util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata)))

	return body
}

// RefreshPropertyValues maps the remote policy onto the model. tagId is the GUID of the policy's
// tag, which the caller resolves since the API only reports the tag's numeric UID and name.
func (r ResourceAccessPolicyResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, policy *citrixorchestration.ResourceAccessPolicyResponseModel, tagId string) ResourceAccessPolicyResourceModel {
	r.Id = types.StringValue(strconv.Itoa(int(policy.GetId())))
	r.Name = types.StringValue(policy.GetName())
	r.Description = types.StringValue(policy.GetDescription())
	r.Enabled = types.BoolValue(policy.GetEnabled())

	if tagId == "" {
		r.TagId = types.StringNull()
	} else if !strings.EqualFold(r.TagId.ValueString(), tagId) {
		r.TagId = types.StringValue(tagId)
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
	r.IncludeCriteriaFilters = util.RefreshListValueProperties[util.SmartAccessTagModel](ctx, diagnostics, r.IncludeCriteriaFilters, policy.GetIncludedSmartAccessTags(), util.GetOrchestrationSmartAccessTagKey)

	r.EnableCriteriaForExcludeConnections = types.BoolValue(policy.GetExcludedSmartAccessFilterEnabled())
	r.ExcludeCriteriaFilters = util.RefreshListValueProperties[util.SmartAccessTagModel](ctx, diagnostics, r.ExcludeCriteriaFilters, policy.GetExcludedSmartAccessTags(), util.GetOrchestrationSmartAccessTagKey)

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), policy.GetMetadata())
	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	return r
}
