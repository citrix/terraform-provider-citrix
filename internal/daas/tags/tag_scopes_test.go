// Copyright © 2026. Citrix Systems, Inc.

package tags

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/citrix-daas-rest-go/test"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// util.AnyScopeId is the built-in scope the DaaS service started assigning to tags on its
// own, which is the regression behind this ticket.
const customerScopeId = "0d2c0a81-b7ff-4bbb-a53a-6b8d0b1b6d16"

func TestFilterServiceAssignedScopes(t *testing.T) {
	testCases := []struct {
		name         string
		remote       []citrixorchestration.ScopeResponseModel
		configured   []string
		expectedKept []string
	}{
		{
			name:         "service assigned scope is dropped",
			remote:       []citrixorchestration.ScopeResponseModel{newScope(util.AnyScopeId, true)},
			expectedKept: []string{},
		},
		{
			name:         "customer scope is kept",
			remote:       []citrixorchestration.ScopeResponseModel{newScope(customerScopeId, false)},
			expectedKept: []string{customerScopeId},
		},
		{
			name:         "service assigned scope is dropped alongside a customer scope",
			remote:       []citrixorchestration.ScopeResponseModel{newScope(util.AnyScopeId, true), newScope(customerScopeId, false)},
			expectedKept: []string{customerScopeId},
		},
		{
			name:         "explicitly configured built-in scope is kept",
			remote:       []citrixorchestration.ScopeResponseModel{newScope(util.AnyScopeId, true)},
			configured:   []string{util.AnyScopeId},
			expectedKept: []string{util.AnyScopeId},
		},
		{
			name:         "configured scope matches regardless of casing",
			remote:       []citrixorchestration.ScopeResponseModel{newScope(strings.ToUpper(util.AnyScopeId), true)},
			configured:   []string{util.AnyScopeId},
			expectedKept: []string{strings.ToUpper(util.AnyScopeId)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			configuredScopeIds := map[string]bool{}
			for _, scopeId := range testCase.configured {
				configuredScopeIds[strings.ToLower(scopeId)] = true
			}

			kept := filterServiceAssignedScopes(testCase.remote, configuredScopeIds)

			assert.Equal(t, testCase.expectedKept, kept)
		})
	}
}

func TestApplyTagScopes(t *testing.T) {
	ctx := context.Background()
	nullScopes := types.SetNull(types.StringType)

	testCases := []struct {
		name           string
		planned        types.Set
		expectedScopes []string
	}{
		{
			name:           "configured scopes are sent",
			planned:        newScopeSet(t, ctx, customerScopeId),
			expectedScopes: []string{customerScopeId},
		},
		{
			// No scopes has to reach the service as an empty list rather than an absent
			// field, otherwise removing the last scope leaves it assigned to the tag.
			name:           "no scopes is sent as an empty list",
			planned:        nullScopes,
			expectedScopes: []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var body citrixorchestration.TagRequestModel
			var diagnostics diag.Diagnostics

			applyTagScopes(ctx, &diagnostics, &body, testCase.planned)

			assert.False(t, diagnostics.HasError())
			assert.Equal(t, testCase.expectedScopes, body.GetScopes())

			// A nil slice is dropped from the payload, so confirm the field is serialized
			// even when it is empty.
			serialized, err := body.ToMap()
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedScopes, serialized["Scopes"])
		})
	}
}

func TestRefreshTagScopes_NoScopes(t *testing.T) {
	tag := &citrixorchestration.TagDetailResponseModel{}
	tag.SetScopes([]citrixorchestration.ScopeResponseModel{})

	var diagnostics diag.Diagnostics
	scopes := refreshTagScopes(context.Background(), &diagnostics, tag, types.SetNull(types.StringType))

	assert.True(t, scopes.IsNull())
	assert.False(t, diagnostics.HasError())
}

// The regression behind this ticket: the service assigns a built-in scope the configuration
// never asked for, and it must not reach state.
func TestRefreshTagScopes_DropsServiceAssignedBuiltInScopes(t *testing.T) {
	tag := &citrixorchestration.TagDetailResponseModel{}
	tag.SetScopes([]citrixorchestration.ScopeResponseModel{
		newScope(util.AnyScopeId, true),
		newScope(customerScopeId, false),
	})

	ctx := context.Background()
	var diagnostics diag.Diagnostics
	scopes := refreshTagScopes(ctx, &diagnostics, tag, types.SetNull(types.StringType))

	assert.False(t, diagnostics.HasError())
	assert.Equal(t, []string{customerScopeId}, util.StringSetToStringArray(ctx, &diagnostics, scopes))
}

func TestCheckTagScopesSupport(t *testing.T) {
	testCases := []struct {
		name             string
		client           *citrixdaasclient.CitrixDaasClient
		expectedSupport  bool
		expectedErrorSub string
	}{
		{
			name:            "unconfigured provider has nothing to validate",
			client:          &citrixdaasclient.CitrixDaasClient{},
			expectedSupport: true,
		},
		{
			name:             "cloud orchestration version below the minimum is unsupported",
			client:           newTagScopesTestClient(false, util.ScopableTagsCloudOrchestrationApiVersion-1, supportedProductVersion, util.FeatureScopableTags),
			expectedSupport:  false,
			expectedErrorSub: "upgrade your DDC Orchestration Service version",
		},
		{
			name:             "on-premises product version below the minimum is unsupported",
			client:           newTagScopesTestClient(true, util.ScopableTagsOnPremOrchestrationApiVersion, unsupportedProductVersion, util.FeatureScopableTags),
			expectedSupport:  false,
			expectedErrorSub: "upgrade your DDC product version",
		},
		{
			name:             "feature not enabled is unsupported",
			client:           newTagScopesTestClient(false, util.ScopableTagsCloudOrchestrationApiVersion, supportedProductVersion),
			expectedSupport:  false,
			expectedErrorSub: "rolling out soon",
		},
		{
			name:            "supported version with the feature enabled",
			client:          newTagScopesTestClient(false, util.ScopableTagsCloudOrchestrationApiVersion, supportedProductVersion, util.FeatureScopableTags),
			expectedSupport: true,
		},
		{
			// The feature works the same way on-premises, so a supported DDC is supported
			// regardless of deployment.
			name:            "supported on-premises version with the feature enabled",
			client:          newTagScopesTestClient(true, util.ScopableTagsOnPremOrchestrationApiVersion, supportedProductVersion, util.FeatureScopableTags),
			expectedSupport: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var diagnostics diag.Diagnostics

			supported := checkTagScopesSupport(testCase.client, &diagnostics, "Error creating Tag: test-tag")

			assert.Equal(t, testCase.expectedSupport, supported)
			if testCase.expectedErrorSub == "" {
				assert.False(t, diagnostics.HasError())
				return
			}

			assert.True(t, diagnostics.HasError())
			assert.Contains(t, diagnostics.Errors()[0].Detail(), testCase.expectedErrorSub)
		})
	}
}

var supportedProductVersion = fmt.Sprintf("%d.%d", util.ScopableTagsProductMajorVersion, util.ScopableTagsProductMinorVersion)
var unsupportedProductVersion = fmt.Sprintf("%d.%d", util.ScopableTagsProductMajorVersion, util.ScopableTagsProductMinorVersion-1)

func newTagScopesTestClient(onPremises bool, orchestrationApiVersion int32, productVersion string, enabledFeatures ...string) *citrixdaasclient.CitrixDaasClient {
	client, _ := test.NewTestDaaSClient()
	client.AuthConfig = &citrixdaasclient.AuthenticationConfiguration{OnPremises: onPremises}
	client.ClientConfig.OrchestrationApiVersion = orchestrationApiVersion
	client.ClientConfig.ProductVersion = productVersion
	client.ClientConfig.EnabledFeatures = enabledFeatures
	return client
}

func newScopeSet(t *testing.T, ctx context.Context, scopeIds ...string) types.Set {
	var diagnostics diag.Diagnostics
	scopes := util.StringArrayToStringSet(ctx, &diagnostics, scopeIds)
	assert.False(t, diagnostics.HasError())
	return scopes
}

func newScope(scopeId string, isBuiltIn bool) citrixorchestration.ScopeResponseModel {
	scope := citrixorchestration.ScopeResponseModel{}
	scope.SetId(scopeId)
	scope.SetIsBuiltIn(isBuiltIn)
	return scope
}
