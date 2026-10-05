// Copyright © 2026. Citrix Systems, Inc.

package tags

import (
	"context"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const tagScopesFeature = "Assigning admin scopes to tags"

// checkTagScopesSupport reports whether admin scopes can be assigned to tags for the
// connected site, recording a diagnostic when they cannot. Support requires both a recent
// enough DDC and the scopable tags feature, which rolls out independently of the version.
func checkTagScopesSupport(client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, errorSummary string) bool {
	if client == nil || client.ClientConfig == nil || client.AuthConfig == nil {
		// The provider is not configured yet, so there is nothing to validate against.
		return true
	}

	if !util.CheckProductVersion(client, diagnostics,
		util.ScopableTagsCloudOrchestrationApiVersion, util.ScopableTagsOnPremOrchestrationApiVersion,
		util.ScopableTagsProductMajorVersion, util.ScopableTagsProductMinorVersion,
		errorSummary, tagScopesFeature) {
		return false
	}

	if !client.IsFeatureEnabled(util.FeatureScopableTags) {
		diagnostics.AddError(
			errorSummary,
			tagScopesFeature+" is not enabled for this site. Support for this feature is rolling out soon.",
		)
		return false
	}

	return true
}

// applyTagScopes sets the scopes on a tag request body. The attribute is optional without
// being computed, so a configuration without scopes means the tag should have none rather
// than that its scopes are unmanaged, and the request always states the full set.
//
// StringSetToStringArray returns nil for a null set and a nil slice is omitted from the
// request entirely, which would leave the existing scopes in place, so no scopes has to be
// sent as an explicit empty slice.
func applyTagScopes(ctx context.Context, diagnostics *diag.Diagnostics, body *citrixorchestration.TagRequestModel, plannedScopes types.Set) {
	scopes := []string{}
	if !plannedScopes.IsNull() {
		scopes = util.StringSetToStringArray(ctx, diagnostics, plannedScopes)
	}

	body.SetScopes(scopes)
}

// refreshTagScopes maps the tag's remote scopes into state, discarding scopes the service
// assigns automatically so that they do not surface as drift. Scopes the configuration
// listed explicitly are always kept, even if they are built-in.
//
// This filtering is applied regardless of whether the scopable tags feature is enabled:
// when it is not, a tag carries only service assigned scopes and the result is null.
func refreshTagScopes(ctx context.Context, diagnostics *diag.Diagnostics, tag *citrixorchestration.TagDetailResponseModel, configuredScopes types.Set) types.Set {
	configuredScopeIds := map[string]bool{}
	for _, scopeId := range util.StringSetToStringArray(ctx, diagnostics, configuredScopes) {
		configuredScopeIds[strings.ToLower(scopeId)] = true
	}

	keptScopeIds := filterServiceAssignedScopes(tag.GetScopes(), configuredScopeIds)
	if len(keptScopeIds) == 0 {
		return types.SetNull(types.StringType)
	}

	return util.StringArrayToStringSet(ctx, diagnostics, keptScopeIds)
}

// filterServiceAssignedScopes drops the built-in scopes the configuration did not ask for,
// preserving the order of the remaining ones.
func filterServiceAssignedScopes(scopes []citrixorchestration.ScopeResponseModel, configuredScopeIds map[string]bool) []string {
	keptScopeIds := []string{}
	for _, scope := range scopes {
		scopeId := scope.GetId()
		if scopeId == "" {
			continue
		}

		if configuredScopeIds[strings.ToLower(scopeId)] || !scope.GetIsBuiltIn() {
			keptScopeIds = append(keptScopeIds, scopeId)
		}
	}

	return keptScopeIds
}
