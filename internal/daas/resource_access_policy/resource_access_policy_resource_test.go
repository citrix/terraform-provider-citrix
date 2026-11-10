// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy // whitebox testing in same package

import (
	"context"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

const testTagId = "c1b2d3e4-0000-4000-8000-000000000001"

func nullList[T util.ResourceModelWithAttributes](t *testing.T) types.List {
	var item T
	attributeMap, err := util.ResourceAttributeMapFromObject(item)
	require.NoError(t, err)
	return types.ListNull(types.ObjectType{AttrTypes: attributeMap})
}

func filterList(ctx context.Context, filters ...[2]string) types.List {
	models := []util.SmartAccessTagModel{}
	for _, filter := range filters {
		models = append(models, util.SmartAccessTagModel{
			FilterName:  types.StringValue(filter[0]),
			FilterValue: types.StringValue(filter[1]),
		})
	}
	return util.TypedArrayToObjectList(ctx, nil, models)
}

// newPolicyModel creates a model targeting a tag with both criteria disabled.
func newPolicyModel(t *testing.T) ResourceAccessPolicyResourceModel {
	return ResourceAccessPolicyResourceModel{
		Id:                                  types.StringNull(),
		Name:                                types.StringValue("policy"),
		Description:                         types.StringValue(""),
		Enabled:                             types.BoolValue(true),
		TagId:                               types.StringValue(testTagId),
		ResourceType:                        types.StringNull(),
		EnableCriteriaForIncludeConnections: types.BoolValue(false),
		IncludeConnectionsCriteriaType:      types.StringNull(),
		IncludeCriteriaFilters:              nullList[util.SmartAccessTagModel](t),
		EnableCriteriaForExcludeConnections: types.BoolValue(false),
		ExcludeCriteriaFilters:              nullList[util.SmartAccessTagModel](t),
		Metadata:                            nullList[util.NameValueStringPairModel](t),
	}
}

func newSmartAccessTag(farm, filter string) citrixorchestration.SmartAccessTagResponseModel {
	tag := citrixorchestration.SmartAccessTagResponseModel{}
	tag.SetFarm(farm)
	tag.SetFilter(filter)
	return tag
}

func newPolicy(id int32, name string) *citrixorchestration.ResourceAccessPolicyResponseModel {
	policy := &citrixorchestration.ResourceAccessPolicyResponseModel{}
	policy.SetId(id)
	policy.SetName(name)
	policy.SetDescription("description")
	policy.SetEnabled(true)
	return policy
}

func newTag(id, name string, uid int32) *citrixorchestration.TagDetailResponseModel {
	tag := &citrixorchestration.TagDetailResponseModel{}
	tag.SetId(id)
	tag.SetName(name)
	tag.SetUid(uid)
	return tag
}

// ========== Tests ==========

func TestResourceAccessPolicyValidateConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name          string
		modify        func(*ResourceAccessPolicyResourceModel)
		errorContains []string
	}{
		{
			name:   "Tag only",
			modify: func(m *ResourceAccessPolicyResourceModel) {},
		},
		{
			name: "Resource type only",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.TagId = types.StringNull()
				m.ResourceType = types.StringValue("Application")
			},
		},
		{
			name: "Unknown tag is not validated",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.TagId = types.StringUnknown()
			},
		},
		{
			name: "Neither tag nor resource type",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.TagId = types.StringNull()
			},
			errorContains: []string{"At least one of tag_id and resource_type must be specified."},
		},
		{
			name: "Include and exclude criteria enabled with filters",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.EnableCriteriaForIncludeConnections = types.BoolValue(true)
				m.IncludeConnectionsCriteriaType = types.StringValue("MatchAll")
				m.IncludeCriteriaFilters = filterList(ctx, [2]string{"farm", "filter"})
				m.EnableCriteriaForExcludeConnections = types.BoolValue(true)
				m.ExcludeCriteriaFilters = filterList(ctx, [2]string{"farm", "other"})
			},
		},
		{
			name: "Include criteria enabled without type or filters",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.EnableCriteriaForIncludeConnections = types.BoolValue(true)
			},
			errorContains: []string{
				"include_connections_criteria_type must be specified when enable_criteria_for_include_connections is set to true.",
				"include_criteria_filters must be specified when enable_criteria_for_include_connections is set to true.",
			},
		},
		{
			name: "Exclude criteria enabled without filters",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.EnableCriteriaForExcludeConnections = types.BoolValue(true)
			},
			errorContains: []string{"exclude_criteria_filters must be specified when enable_criteria_for_exclude_connections is set to true."},
		},
		{
			name: "Criteria specified while disabled",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.IncludeConnectionsCriteriaType = types.StringValue("MatchAny")
				m.IncludeCriteriaFilters = filterList(ctx, [2]string{"farm", "filter"})
				m.ExcludeCriteriaFilters = filterList(ctx, [2]string{"farm", "filter"})
			},
			errorContains: []string{
				"include_connections_criteria_type cannot be specified when enable_criteria_for_include_connections is set to false.",
				"include_criteria_filters cannot be specified when enable_criteria_for_include_connections is set to false.",
				"exclude_criteria_filters cannot be specified when enable_criteria_for_exclude_connections is set to false.",
			},
		},
		{
			name: "Duplicate filters ignoring case",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				m.EnableCriteriaForExcludeConnections = types.BoolValue(true)
				m.ExcludeCriteriaFilters = filterList(ctx, [2]string{"Farm", "Filter"}, [2]string{"farm", "filter"})
			},
			errorContains: []string{"Filter name farm with filter value filter is specified more than once."},
		},
		{
			name: "Unknown filters are not duplicates",
			modify: func(m *ResourceAccessPolicyResourceModel) {
				attributeMap, err := util.ResourceAttributeMapFromObject(util.SmartAccessTagModel{})
				require.NoError(t, err)
				objectType := types.ObjectType{AttrTypes: attributeMap}

				m.EnableCriteriaForExcludeConnections = types.BoolValue(true)
				m.ExcludeCriteriaFilters = types.ListValueMust(objectType, []attr.Value{
					types.ObjectUnknown(attributeMap),
					types.ObjectUnknown(attributeMap),
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newPolicyModel(t)
			tt.modify(&model)

			diagnostics := diag.Diagnostics{}
			model.ValidateConfig(ctx, &diagnostics)

			require.Len(t, diagnostics.Errors(), len(tt.errorContains), "unexpected errors: %v", diagnostics.Errors())
			for index, expected := range tt.errorContains {
				assert.Equal(t, expected, diagnostics.Errors()[index].Detail())
			}
		})
	}
}

func TestResourceAccessPolicyBuildRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("Removed criteria and metadata are sent empty", func(t *testing.T) {
		model := newPolicyModel(t)
		tagUid := int32(42)

		diagnostics := diag.Diagnostics{}
		body := model.buildRequest(ctx, &diagnostics, &tagUid)

		require.False(t, diagnostics.HasError())
		assert.Equal(t, "policy", body.GetName())
		assert.Empty(t, body.GetDescription())
		assert.True(t, body.GetEnabled())
		assert.Equal(t, int32(42), body.GetTagUid())
		assert.Nil(t, body.ResourceType)
		assert.False(t, body.GetIncludedSmartAccessFilterEnabled())
		assert.Nil(t, body.IncludedSmartAccessFilterType)
		assert.NotNil(t, body.IncludedSmartAccessTags)
		assert.Empty(t, body.IncludedSmartAccessTags)
		assert.False(t, body.GetExcludedSmartAccessFilterEnabled())
		assert.NotNil(t, body.ExcludedSmartAccessTags)
		assert.Empty(t, body.ExcludedSmartAccessTags)
		require.Len(t, body.Metadata, 1)
		assert.Equal(t, util.MetadataTerraformName, body.Metadata[0].GetName())
	})

	t.Run("All attributes", func(t *testing.T) {
		model := newPolicyModel(t)
		model.TagId = types.StringNull()
		model.ResourceType = types.StringValue("Desktop")
		model.EnableCriteriaForIncludeConnections = types.BoolValue(true)
		model.IncludeConnectionsCriteriaType = types.StringValue("MatchAll")
		model.IncludeCriteriaFilters = filterList(ctx, [2]string{"farm", "include"})
		model.EnableCriteriaForExcludeConnections = types.BoolValue(true)
		model.ExcludeCriteriaFilters = filterList(ctx, [2]string{"farm", "exclude"})
		model.Metadata = util.TypedArrayToObjectList(ctx, nil, []util.NameValueStringPairModel{
			{Name: types.StringValue("key"), Value: types.StringValue("value")},
		})

		diagnostics := diag.Diagnostics{}
		body := model.buildRequest(ctx, &diagnostics, nil)

		require.False(t, diagnostics.HasError())
		assert.False(t, body.TagUid.IsSet())
		assert.Equal(t, citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_DESKTOP, body.GetResourceType())
		assert.True(t, body.GetIncludedSmartAccessFilterEnabled())
		assert.Equal(t, citrixorchestration.FILTERMATCHTYPE_MATCH_ALL, body.GetIncludedSmartAccessFilterType())
		require.Len(t, body.IncludedSmartAccessTags, 1)
		assert.Equal(t, "farm", body.IncludedSmartAccessTags[0].GetFarm())
		assert.Equal(t, "include", body.IncludedSmartAccessTags[0].GetFilter())
		assert.True(t, body.GetExcludedSmartAccessFilterEnabled())
		require.Len(t, body.ExcludedSmartAccessTags, 1)
		assert.Equal(t, "exclude", body.ExcludedSmartAccessTags[0].GetFilter())
		assert.Len(t, body.Metadata, 2)
	})
}

func TestResourceAccessPolicyRefreshPropertyValues(t *testing.T) {
	ctx := context.Background()

	t.Run("Policy without tag or criteria", func(t *testing.T) {
		policy := newPolicy(7, "policy")
		policy.SetResourceType(citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_APPLICATION)
		policy.SetIncludedSmartAccessFilterType(citrixorchestration.FILTERMATCHTYPE_MATCH_ANY)
		policy.SetMetadata([]citrixorchestration.NameValueStringPairModel{testutil.NewMetadataPair(util.MetadataTerraformName, util.MetadataTerrafomValue)})

		diagnostics := diag.Diagnostics{}
		model := newPolicyModel(t).RefreshPropertyValues(ctx, &diagnostics, policy, "")

		require.False(t, diagnostics.HasError())
		assert.Equal(t, "7", model.Id.ValueString())
		assert.Equal(t, "description", model.Description.ValueString())
		assert.True(t, model.TagId.IsNull())
		assert.Equal(t, "Application", model.ResourceType.ValueString())
		assert.False(t, model.EnableCriteriaForIncludeConnections.ValueBool())
		assert.True(t, model.IncludeConnectionsCriteriaType.IsNull(), "the criteria type is only refreshed when the include criteria are enabled")
		assert.True(t, model.IncludeCriteriaFilters.IsNull())
		assert.True(t, model.ExcludeCriteriaFilters.IsNull())
		assert.True(t, model.Metadata.IsNull(), "metadata not in the configuration is not refreshed")
	})

	t.Run("Policy with tag and criteria", func(t *testing.T) {
		policy := newPolicy(7, "policy")
		policy.SetTagUid(42)
		policy.SetTagName("tag")
		policy.SetResourceType(citrixorchestration.RESOURCEACCESSPOLICYRESOURCETYPE_UNKNOWN)
		policy.SetIncludedSmartAccessFilterEnabled(true)
		policy.SetIncludedSmartAccessFilterType(citrixorchestration.FILTERMATCHTYPE_MATCH_ALL)
		policy.SetIncludedSmartAccessTags([]citrixorchestration.SmartAccessTagResponseModel{newSmartAccessTag("farm", "b"), newSmartAccessTag("farm", "a")})
		policy.SetExcludedSmartAccessFilterEnabled(true)
		policy.SetExcludedSmartAccessTags([]citrixorchestration.SmartAccessTagResponseModel{newSmartAccessTag("farm", "c")})

		model := newPolicyModel(t)
		model.TagId = types.StringValue("C1B2D3E4-0000-4000-8000-000000000001")
		model.IncludeCriteriaFilters = filterList(ctx, [2]string{"farm", "a"}, [2]string{"farm", "b"})

		diagnostics := diag.Diagnostics{}
		model = model.RefreshPropertyValues(ctx, &diagnostics, policy, testTagId)

		require.False(t, diagnostics.HasError())
		assert.Equal(t, "C1B2D3E4-0000-4000-8000-000000000001", model.TagId.ValueString(), "a tag id differing only in case is kept")
		assert.True(t, model.ResourceType.IsNull())
		assert.Equal(t, "MatchAll", model.IncludeConnectionsCriteriaType.ValueString())

		includeFilters := util.ObjectListToTypedArray[util.SmartAccessTagModel](ctx, &diagnostics, model.IncludeCriteriaFilters)
		require.Len(t, includeFilters, 2)
		assert.Equal(t, "a", includeFilters[0].FilterValue.ValueString(), "the configured order is kept")
		assert.Equal(t, "b", includeFilters[1].FilterValue.ValueString())

		excludeFilters := util.ObjectListToTypedArray[util.SmartAccessTagModel](ctx, &diagnostics, model.ExcludeCriteriaFilters)
		require.Len(t, excludeFilters, 1)
		assert.Equal(t, "c", excludeFilters[0].FilterValue.ValueString())
	})
}

func TestRequiresReplaceWhenRemoved(t *testing.T) {
	existingResource := tftypes.NewValue(tftypes.String, "resource")

	tests := []struct {
		name            string
		state           types.String
		config          types.String
		requiresReplace bool
	}{
		{name: "Removed", state: types.StringValue("Application"), config: types.StringNull(), requiresReplace: true},
		{name: "Changed", state: types.StringValue("Application"), config: types.StringValue("Desktop"), requiresReplace: false},
		{name: "Added", state: types.StringNull(), config: types.StringValue("Desktop"), requiresReplace: false},
		{name: "Unchanged", state: types.StringValue("Application"), config: types.StringValue("Application"), requiresReplace: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				State:       tfsdk.State{Raw: existingResource},
				Plan:        tfsdk.Plan{Raw: existingResource},
				StateValue:  tt.state,
				ConfigValue: tt.config,
				PlanValue:   tt.config,
			}
			resp := &planmodifier.StringResponse{PlanValue: tt.config}

			requiresReplaceWhenRemoved("resource_type").PlanModifyString(context.Background(), req, resp)

			assert.Equal(t, tt.requiresReplace, resp.RequiresReplace)
		})
	}
}

func TestGetTagUid(t *testing.T) {
	t.Run("No tag", func(t *testing.T) {
		client, mockClient := test.NewTestDaaSClient()
		defer mockClient.AssertExpectations(t)

		diagnostics := diag.Diagnostics{}
		tagUid, err := getTagUid(test.DaaSTestContext(), client, &diagnostics, types.StringNull(), "summary")

		require.NoError(t, err)
		assert.Nil(t, tagUid)
	})

	t.Run("Resolves the tag GUID to its UID", func(t *testing.T) {
		client, mockClient := test.NewTestDaaSClient()
		defer mockClient.AssertExpectations(t)

		mockTagsAPI := citrixorchestration.GetMockTagsAPIsDAAS(mockClient.APIClient)
		mockTagsAPI.On("TagsGetTagExecute", mock.Anything).Return(newTag(testTagId, "tag", 42), testutil.MockSuccessResponse(), nil).Once()

		diagnostics := diag.Diagnostics{}
		tagUid, err := getTagUid(test.DaaSTestContext(), client, &diagnostics, types.StringValue(testTagId), "summary")

		require.NoError(t, err)
		require.NotNil(t, tagUid)
		assert.Equal(t, int32(42), *tagUid)
	})

	t.Run("Tag not found", func(t *testing.T) {
		client, mockClient := test.NewTestDaaSClient()
		defer mockClient.AssertExpectations(t)

		mockTagsAPI := citrixorchestration.GetMockTagsAPIsDAAS(mockClient.APIClient)
		mockTagsAPI.On("TagsGetTagExecute", mock.Anything).Return((*citrixorchestration.TagDetailResponseModel)(nil), testutil.MockErrorResponse(404, "txn-404"), assert.AnError).Once()

		diagnostics := diag.Diagnostics{}
		_, err := getTagUid(test.DaaSTestContext(), client, &diagnostics, types.StringValue(testTagId), "summary")

		require.Error(t, err)
		require.True(t, diagnostics.HasError())
		assert.Contains(t, diagnostics.Errors()[0].Detail(), "Error reading tag "+testTagId)
	})
}

func TestGetTagIdOfPolicy(t *testing.T) {
	t.Run("Policy without tag", func(t *testing.T) {
		client, mockClient := test.NewTestDaaSClient()
		defer mockClient.AssertExpectations(t)

		diagnostics := diag.Diagnostics{}
		tagId, err := getTagIdOfPolicy(test.DaaSTestContext(), client, &diagnostics, newPolicy(7, "policy"), "summary")

		require.NoError(t, err)
		assert.Empty(t, tagId)
	})

	t.Run("Resolves the tag by name to its GUID", func(t *testing.T) {
		client, mockClient := test.NewTestDaaSClient()
		defer mockClient.AssertExpectations(t)

		mockTagsAPI := citrixorchestration.GetMockTagsAPIsDAAS(mockClient.APIClient)
		mockTagsAPI.On("TagsGetTagExecute", mock.MatchedBy(func(r citrixorchestration.ApiTagsGetTagRequest) bool {
			return citrixorchestration.MockApiTagsGetTagRequest{ApiTagsGetTagRequest: r}.GetNameOrId() == "tag"
		})).Return(newTag(testTagId, "tag", 42), testutil.MockSuccessResponse(), nil).Once()

		policy := newPolicy(7, "policy")
		policy.SetTagUid(42)
		policy.SetTagName("tag")

		diagnostics := diag.Diagnostics{}
		tagId, err := getTagIdOfPolicy(test.DaaSTestContext(), client, &diagnostics, policy, "summary")

		require.NoError(t, err)
		assert.Equal(t, testTagId, tagId)
	})
}

func TestFindPolicy(t *testing.T) {
	policies := []citrixorchestration.ResourceAccessPolicyResponseModel{*newPolicy(1, "First"), *newPolicy(2, "Second")}

	assert.Equal(t, "Second", findPolicyById(policies, "2").GetName())
	assert.Nil(t, findPolicyById(policies, "3"))
	assert.Equal(t, int32(1), findPolicyByName(policies, "first").GetId())
	assert.Nil(t, findPolicyByName(policies, "Third"))
}
