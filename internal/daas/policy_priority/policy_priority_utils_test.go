// Copyright © 2026. Citrix Systems, Inc.

package policy_priority

import "testing"

const (
	testPolicy1   = "b8d0a1f6-0000-4000-8000-000000000001"
	testPolicy2   = "b8d0a1f6-0000-4000-8000-000000000002"
	testUnlisted  = "b8d0a1f6-0000-4000-8000-000000000003"
	testUnlisted2 = "b8d0a1f6-0000-4000-8000-000000000004"
)

func TestResolveManagedPolicies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		managed         []string
		adoptAll        bool
		remoteIds       []string
		remoteNames     []string
		orderFromRemote bool
		expectIds       []string
		expectNames     []string
	}{
		{
			name:            "unlisted remote policy is excluded",
			managed:         []string{testPolicy1, testPolicy2},
			remoteIds:       []string{testPolicy1, testPolicy2, testUnlisted},
			remoteNames:     []string{"one", "two", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{testPolicy1, testPolicy2},
			expectNames:     []string{"one", "two"},
		},
		{
			name:            "Read surfaces a remote reorder of managed policies as drift",
			managed:         []string{testPolicy1, testPolicy2},
			remoteIds:       []string{testPolicy2, testPolicy1, testUnlisted},
			remoteNames:     []string{"two", "one", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{testPolicy2, testPolicy1},
			expectNames:     []string{"two", "one"},
		},
		{
			name:            "post-apply keeps the plan order even if remote disagrees",
			managed:         []string{testPolicy1, testPolicy2},
			remoteIds:       []string{testPolicy2, testPolicy1, testUnlisted},
			remoteNames:     []string{"two", "one", "unlisted"},
			orderFromRemote: false,
			expectIds:       []string{testPolicy1, testPolicy2},
			expectNames:     []string{"one", "two"},
		},
		{
			name:            "configuration GUID casing is preserved, not the API's",
			managed:         []string{"B8D0A1F6-0000-4000-8000-000000000001"},
			remoteIds:       []string{testPolicy1, testUnlisted},
			remoteNames:     []string{"one", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{"B8D0A1F6-0000-4000-8000-000000000001"},
			expectNames:     []string{"one"},
		},
		{
			// A repeat the plan committed to must be echoed back, or the framework rejects the
			// apply on a changed element count. Only the rank body dedupes.
			name:            "a repeated managed id is preserved in state",
			managed:         []string{testPolicy1, testPolicy1, testPolicy2},
			remoteIds:       []string{testPolicy1, testPolicy2, testUnlisted},
			remoteNames:     []string{"one", "two", "unlisted"},
			orderFromRemote: false,
			expectIds:       []string{testPolicy1, testPolicy1, testPolicy2},
			expectNames:     []string{"one", "one", "two"},
		},
		{
			name:            "managed policy deleted out of band drops out",
			managed:         []string{testPolicy1, testPolicy2},
			remoteIds:       []string{testPolicy1},
			remoteNames:     []string{"one"},
			orderFromRemote: true,
			expectIds:       []string{testPolicy1},
			expectNames:     []string{"one"},
		},
		{
			name:            "adoptAll takes the whole set (import)",
			managed:         nil,
			adoptAll:        true,
			remoteIds:       []string{testPolicy1, testUnlisted},
			remoteNames:     []string{"one", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{testPolicy1, testUnlisted},
			expectNames:     []string{"one", "unlisted"},
		},
		{
			// nil without adoptAll means the list failed to convert: fail closed, never widen.
			name:            "nil managed list without adoptAll manages nothing",
			managed:         nil,
			adoptAll:        false,
			remoteIds:       []string{testPolicy1, testUnlisted},
			remoteNames:     []string{"one", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{},
			expectNames:     []string{},
		},
		{
			name:            "empty managed list resolves to nothing managed",
			managed:         []string{},
			remoteIds:       []string{testPolicy1, testUnlisted},
			remoteNames:     []string{"one", "unlisted"},
			orderFromRemote: true,
			expectIds:       []string{},
			expectNames:     []string{},
		},
		{
			name:            "empty policy set resolves to nothing managed",
			managed:         []string{testPolicy1},
			remoteIds:       []string{},
			remoteNames:     []string{},
			orderFromRemote: true,
			expectIds:       []string{},
			expectNames:     []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotIds, gotNames := resolveManagedPolicies(tc.managed, tc.adoptAll, tc.remoteIds, tc.remoteNames, tc.orderFromRemote)

			if !equalStrings(gotIds, tc.expectIds) {
				t.Errorf("ids: got %v, want %v", gotIds, tc.expectIds)
			}
			if !equalStrings(gotNames, tc.expectNames) {
				t.Errorf("names: got %v, want %v", gotNames, tc.expectNames)
			}
		})
	}
}

func TestMergePolicyPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		managed []string
		remote  []string
		expect  []string
	}{
		{
			name:    "unlisted policy is appended after the managed ones",
			managed: []string{testPolicy2, testPolicy1},
			remote:  []string{testPolicy1, testPolicy2, testUnlisted},
			expect:  []string{testPolicy2, testPolicy1, testUnlisted},
		},
		{
			name:    "several unlisted policies keep their relative remote order",
			managed: []string{testPolicy1},
			remote:  []string{testUnlisted2, testPolicy1, testUnlisted},
			expect:  []string{testPolicy1, testUnlisted2, testUnlisted},
		},
		{
			name:    "nothing unlisted leaves the managed order untouched",
			managed: []string{testPolicy2, testPolicy1},
			remote:  []string{testPolicy1, testPolicy2},
			expect:  []string{testPolicy2, testPolicy1},
		},
		{
			name:    "empty managed list ranks the remote order as-is",
			managed: []string{},
			remote:  []string{testPolicy1, testUnlisted},
			expect:  []string{testPolicy1, testUnlisted},
		},
		{
			name:    "managed id matching remote case-insensitively is not duplicated",
			managed: []string{"B8D0A1F6-0000-4000-8000-000000000001"},
			remote:  []string{testPolicy1, testUnlisted},
			expect:  []string{"B8D0A1F6-0000-4000-8000-000000000001", testUnlisted},
		},
		{
			// UniqueValues cannot see ids that are unknown at plan time, so a repeat can reach
			// here and must not be sent to the rank API twice.
			name:    "a repeated managed id is sent once",
			managed: []string{testPolicy1, testPolicy1, testPolicy2},
			remote:  []string{testPolicy1, testPolicy2, testUnlisted},
			expect:  []string{testPolicy1, testPolicy2, testUnlisted},
		},
		{
			name:    "a repeat differing only in case is sent once",
			managed: []string{testPolicy1, "B8D0A1F6-0000-4000-8000-000000000001"},
			remote:  []string{testPolicy1, testUnlisted},
			expect:  []string{testPolicy1, testUnlisted},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mergePolicyPriority(tc.managed, tc.remote)

			if !equalStrings(got, tc.expect) {
				t.Errorf("got %v, want %v", got, tc.expect)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
