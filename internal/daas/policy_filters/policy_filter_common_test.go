// Copyright © 2026. Citrix Systems, Inc.

package policy_filters

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
)

func newFilterResponse(t *testing.T, filterGuid, policyGuid, filterType, uuid string) citrixorchestration.FilterResponse {
	t.Helper()

	filter := citrixorchestration.FilterResponse{}
	filter.SetFilterGuid(filterGuid)
	filter.SetPolicyGuid(policyGuid)
	filter.SetFilterType(filterType)

	data, err := json.Marshal(util.PolicyFilterUuidDataClientModel{Uuid: uuid, Server: "example.xendesktop.net"})
	if err != nil {
		t.Fatalf("failed to marshal filter data: %s", err)
	}
	filter.SetFilterData(string(data))

	return filter
}

func TestFindConflictingDeliveryGroupFilter(t *testing.T) {
	t.Parallel()

	const policyId = "b8d0a1f6-0000-4000-8000-000000000001"
	const otherPolicyId = "b8d0a1f6-0000-4000-8000-000000000002"
	const deliveryGroupId = "c9e1b2a7-1111-4111-8111-111111111111"

	malformed := citrixorchestration.FilterResponse{}
	malformed.SetFilterGuid("malformed-filter")
	malformed.SetPolicyGuid(policyId)
	malformed.SetFilterType("DesktopGroup")
	malformed.SetFilterData("not json")

	tests := []struct {
		name             string
		filters          []citrixorchestration.FilterResponse
		expectConflict   bool
		expectFilterGuid string
	}{
		{
			name:           "no filters",
			filters:        nil,
			expectConflict: false,
		},
		{
			name: "same delivery group on same policy conflicts",
			filters: []citrixorchestration.FilterResponse{
				newFilterResponse(t, "existing-filter", policyId, "DesktopGroup", deliveryGroupId),
			},
			expectConflict:   true,
			expectFilterGuid: "existing-filter",
		},
		{
			name: "same delivery group with different GUID casing still conflicts",
			filters: []citrixorchestration.FilterResponse{
				newFilterResponse(t, "existing-filter", policyId, "DesktopGroup", "C9E1B2A7-1111-4111-8111-111111111111"),
			},
			expectConflict:   true,
			expectFilterGuid: "existing-filter",
		},
		{
			name: "same delivery group on a different policy does not conflict",
			filters: []citrixorchestration.FilterResponse{
				newFilterResponse(t, "other-policy-filter", otherPolicyId, "DesktopGroup", deliveryGroupId),
			},
			expectConflict: false,
		},
		{
			name: "different filter type does not conflict",
			filters: []citrixorchestration.FilterResponse{
				newFilterResponse(t, "tag-filter", policyId, "Tag", deliveryGroupId),
			},
			expectConflict: false,
		},
		{
			name: "different delivery group does not conflict",
			filters: []citrixorchestration.FilterResponse{
				newFilterResponse(t, "other-dg-filter", policyId, "DesktopGroup", "d0f2c3b8-2222-4222-8222-222222222222"),
			},
			expectConflict: false,
		},
		{
			name:           "malformed filter data is skipped",
			filters:        []citrixorchestration.FilterResponse{malformed},
			expectConflict: false,
		},
		{
			name: "malformed filter data does not mask a later conflict",
			filters: []citrixorchestration.FilterResponse{
				malformed,
				newFilterResponse(t, "existing-filter", policyId, "DesktopGroup", deliveryGroupId),
			},
			expectConflict:   true,
			expectFilterGuid: "existing-filter",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conflict, found := findConflictingDeliveryGroupFilter(test.filters, policyId, deliveryGroupId)

			if found != test.expectConflict {
				t.Fatalf("expected conflict=%t, got %t", test.expectConflict, found)
			}
			if found && conflict.GetFilterGuid() != test.expectFilterGuid {
				t.Errorf("expected conflicting filter %q, got %q", test.expectFilterGuid, conflict.GetFilterGuid())
			}
		})
	}
}

func TestBuildDuplicateDeliveryGroupFilterError(t *testing.T) {
	t.Parallel()

	message := buildDuplicateDeliveryGroupFilterError(
		"existing-filter-guid",
		"c9e1b2a7-1111-4111-8111-111111111111",
		"b8d0a1f6-0000-4000-8000-000000000001",
	)

	for _, want := range []string{
		"existing-filter-guid",
		"c9e1b2a7-1111-4111-8111-111111111111",
		"b8d0a1f6-0000-4000-8000-000000000001",
		"terraform import",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("expected error message to contain %q, got:\n%s", want, message)
		}
	}
}

func TestLockPolicySerialisesSamePolicy(t *testing.T) {
	t.Parallel()

	const policyId = "b8d0a1f6-0000-4000-8000-000000000010"
	const goroutines = 50

	// The start barrier and the sleep below are load-bearing: without them the critical section
	// is too short to interleave and this test passes even with no lock at all.
	created := ""
	winners := 0

	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutines)

	for range goroutines {
		go func() {
			defer waitGroup.Done()
			<-start
			defer lockPolicy(policyId)()

			if created == "" {
				time.Sleep(time.Millisecond)
				created = "filter"
				winners++
			}
		}()
	}

	close(start)
	waitGroup.Wait()

	if winners != 1 {
		t.Errorf("expected exactly 1 goroutine to pass the check, got %d - the lock is not serialising", winners)
	}
}

func TestLockPolicyDoesNotSerialiseDifferentPolicies(t *testing.T) {
	t.Parallel()

	bothHeld := make(chan struct{})
	done := make(chan struct{})

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	for _, policyId := range []string{"policy-a", "policy-b"} {
		go func(id string) {
			defer waitGroup.Done()
			defer lockPolicy(id)()
			bothHeld <- struct{}{}
			<-done
		}(policyId)
	}

	// Both goroutines must reach the signal while still holding their locks. Time out rather
	// than block, so a lock that is not keyed per policy fails here instead of hanging the
	// whole package.
	for i := range 2 {
		select {
		case <-bothHeld:
		case <-time.After(5 * time.Second):
			close(done)
			t.Fatalf("only %d of 2 policies could hold their lock at once - the lock is not keyed per policy", i)
		}
	}

	close(done)
	waitGroup.Wait()
}
