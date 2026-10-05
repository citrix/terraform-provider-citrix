// Copyright © 2026. Citrix Systems, Inc.

package policy_priority

import (
	"slices"
	"sort"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
)

// policyIdsAndNamesByPriority returns the policy set's ids and names in priority order.
// Priority is optional and defaults to 0, so sort stably to keep ties in the API's response order.
func policyIdsAndNamesByPriority(policies *citrixorchestration.CollectionEnvelopeOfPolicyResponse) ([]string, []string) {
	ids := []string{}
	names := []string{}
	if policies == nil || policies.Items == nil {
		return ids, names
	}

	// Clone first: GetItems and Items share a backing array, so sorting would reorder the caller's envelope.
	items := slices.Clone(policies.Items)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].GetPriority() < items[j].GetPriority()
	})
	for _, policy := range items {
		ids = append(ids, policy.GetPolicyGuid())
		names = append(names, policy.GetPolicyName())
	}

	return ids, names
}

// dedupeFold drops case-insensitive repeats. UniqueValues only inspects known values, so ids
// taken from resource references reach here unchecked on first apply.
func dedupeFold(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		key := strings.ToLower(id)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, id)
	}
	return out
}

// resolveManagedPolicies narrows the set to the ids the configuration manages, keeping its own
// GUID spelling so casing cannot fail the plan-consistency check. orderFromRemote orders by the
// set (Read) instead of by managed (post-apply). adoptAll is a flag, not a nil check on managed,
// because a failed list conversion also yields nil and must fail closed.
func resolveManagedPolicies(managed []string, adoptAll bool, remoteIds, remoteNames []string, orderFromRemote bool) ([]string, []string) {
	if adoptAll {
		return remoteIds, remoteNames
	}

	nameByRemoteId := make(map[string]string, len(remoteIds))
	positionByRemoteId := make(map[string]int, len(remoteIds))
	for i, remoteId := range remoteIds {
		key := strings.ToLower(remoteId)
		positionByRemoteId[key] = i
		if i < len(remoteNames) {
			nameByRemoteId[key] = remoteNames[i]
		}
	}

	// Read drops ids no longer in the set, which is real drift. Post-apply echoes the plan
	// verbatim - neither deduped nor filtered - because policy_priority is Required and any
	// change in element count fails the framework's plan-consistency check.
	ids := []string{}
	for _, id := range managed {
		_, inRemote := positionByRemoteId[strings.ToLower(id)]
		if orderFromRemote && !inRemote {
			continue
		}
		ids = append(ids, id)
	}

	if orderFromRemote {
		sort.SliceStable(ids, func(i, j int) bool {
			return positionByRemoteId[strings.ToLower(ids[i])] < positionByRemoteId[strings.ToLower(ids[j])]
		})
	}

	// Look names up per id, not by index, so the two slices cannot drift out of alignment.
	names := []string{}
	for _, id := range ids {
		names = append(names, nameByRemoteId[strings.ToLower(id)])
	}

	return ids, names
}

// mergePolicyPriority builds the rank order: managed policies as configured, then the rest of the
// set in its existing order, so every listed policy outranks every unlisted one (XAC-77859).
func mergePolicyPriority(managed []string, remoteOrdered []string) []string {
	return util.RefreshList(dedupeFold(managed), remoteOrdered)
}
