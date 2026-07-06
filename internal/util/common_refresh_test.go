// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Test fixtures (PLAN decision D4) ---------------------------------------

// fakeRefreshClient is the minimal "remote" client model used to drive the
// refresh helpers under test.
type fakeRefreshClient struct {
	key   string
	value string
}

// getFakeClientKey extracts the comparison key from a client item.
func getFakeClientKey(c fakeRefreshClient) string { return c.key }

// fakeRefreshModel is the minimal Terraform-side model implementing
// RefreshableListItemWithAttributes[fakeRefreshClient]. Methods use VALUE
// receivers so RefreshListItem returns the concrete type (required for the
// (tfType) type assertion inside refreshListProperties).
type fakeRefreshModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (m fakeRefreshModel) GetKey() string { return m.Key.ValueString() }

func (m fakeRefreshModel) RefreshListItem(_ context.Context, _ *diag.Diagnostics, c fakeRefreshClient) ResourceModelWithAttributes {
	m.Key = types.StringValue(c.key)
	m.Value = types.StringValue(c.value)
	return m
}

func (fakeRefreshModel) GetAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"key":   resourceSchema.StringAttribute{Optional: true},
		"value": resourceSchema.StringAttribute{Optional: true},
	}
}

// Compile-time assertion that the fake model satisfies the interface under test.
var _ RefreshableListItemWithAttributes[fakeRefreshClient] = fakeRefreshModel{}

// newFakeModel is a small constructor for readable test fixtures.
func newFakeModel(key, value string) fakeRefreshModel {
	return fakeRefreshModel{Key: types.StringValue(key), Value: types.StringValue(value)}
}

// --- refreshListProperties (typed core) ------------------------------------

func TestRefreshListProperties(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state  []fakeRefreshModel
		remote []fakeRefreshClient
		// expected holds the (key, value) pairs in expected order.
		expectedKeys   []string
		expectedValues []string
		// expectNil asserts the core returned a nil slice (empty-remote contract).
		expectNil bool
	}

	tests := map[string]testCase{
		"empty remote returns nil": {
			state:     []fakeRefreshModel{newFakeModel("a", "1")},
			remote:    []fakeRefreshClient{},
			expectNil: true,
		},
		"order preserved and value refreshed": {
			state: []fakeRefreshModel{newFakeModel("a", "1"), newFakeModel("b", "2")},
			remote: []fakeRefreshClient{
				{key: "a", value: "1-new"},
				{key: "b", value: "2-new"},
			},
			expectedKeys:   []string{"a", "b"},
			expectedValues: []string{"1-new", "2-new"},
		},
		"remote-only items appended at end in remote order": {
			state: []fakeRefreshModel{newFakeModel("a", "1")},
			remote: []fakeRefreshClient{
				{key: "a", value: "1"},
				{key: "c", value: "3"},
				{key: "b", value: "2"},
			},
			expectedKeys:   []string{"a", "c", "b"},
			expectedValues: []string{"1", "3", "2"},
		},
		"state item absent from remote is removed": {
			state: []fakeRefreshModel{newFakeModel("a", "1"), newFakeModel("b", "2")},
			remote: []fakeRefreshClient{
				{key: "a", value: "1"},
			},
			expectedKeys:   []string{"a"},
			expectedValues: []string{"1"},
		},
		"key match is case-sensitive": {
			// Discriminating scenario: a same-cased anchor "a" exists in both
			// state and remote, plus a differently-cased remote "A".
			//   - Case-SENSITIVE (current impl): "a" matches the state item and is
			//     refreshed in place to "1-lower"; "A" has no match and is appended
			//     -> 2 items [{a,1-lower},{A,1-upper}].
			//   - Case-INSENSITIVE (hypothetical): both "a" and "A" collapse onto
			//     the single state item -> 1 item [{A,1-upper}].
			// Asserting len 2 with both keys therefore genuinely pins case-sensitivity.
			state: []fakeRefreshModel{newFakeModel("a", "1")},
			remote: []fakeRefreshClient{
				{key: "a", value: "1-lower"},
				{key: "A", value: "1-upper"},
			},
			expectedKeys:   []string{"a", "A"},
			expectedValues: []string{"1-lower", "1-upper"},
		},
		"nil state appends all remote in remote order": {
			// Boundary: exercises the `state == nil` branch (state = []tfType{}); all
			// remote items take the append path and land in remote order.
			state:          nil,
			remote:         []fakeRefreshClient{{key: "a", value: "1"}, {key: "b", value: "2"}},
			expectedKeys:   []string{"a", "b"},
			expectedValues: []string{"1", "2"},
		},
		"empty (non-nil) state appends all remote": {
			state:          []fakeRefreshModel{},
			remote:         []fakeRefreshClient{{key: "a", value: "1"}},
			expectedKeys:   []string{"a"},
			expectedValues: []string{"1"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var diags diag.Diagnostics

			got := refreshListProperties[fakeRefreshModel, fakeRefreshClient](ctx, &diags, test.state, test.remote, getFakeClientKey)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", diags)
			}

			if test.expectNil {
				if got != nil {
					t.Fatalf("expected nil result for empty remote, got %#v", got)
				}
				return
			}

			if len(got) != len(test.expectedKeys) {
				t.Fatalf("expected %d items, got %d (%#v)", len(test.expectedKeys), len(got), got)
			}
			for i := range got {
				if k := got[i].Key.ValueString(); k != test.expectedKeys[i] {
					t.Errorf("item %d: expected key %q, got %q", i, test.expectedKeys[i], k)
				}
				if v := got[i].Value.ValueString(); v != test.expectedValues[i] {
					t.Errorf("item %d: expected value %q, got %q", i, test.expectedValues[i], v)
				}
			}
		})
	}
}

// --- RefreshListValueProperties (typed wrapper) ----------------------------

func TestRefreshListValueProperties(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state  []fakeRefreshModel
		remote []fakeRefreshClient
		// expectNull asserts the wrapper returned types.ListNull (empty-remote
		// contract: core returns nil -> TypedArrayToObjectList(nil) -> ListNull).
		expectNull     bool
		expectedKeys   []string
		expectedValues []string
	}

	tests := map[string]testCase{
		"empty remote yields null list": {
			state:      []fakeRefreshModel{newFakeModel("a", "1")},
			remote:     []fakeRefreshClient{},
			expectNull: true,
		},
		"non-empty yields non-null list with correct order": {
			state: []fakeRefreshModel{newFakeModel("a", "1"), newFakeModel("b", "2")},
			remote: []fakeRefreshClient{
				{key: "a", value: "1-new"},
				{key: "c", value: "3"},
			},
			expectedKeys:   []string{"a", "c"},
			expectedValues: []string{"1-new", "3"},
		},
		"null input list with remote yields non-null appended": {
			// Boundary: state nil -> TypedArrayToObjectList(nil) -> ListNull is fed as
			// the wrapper's input; remote items are appended and the result is non-null.
			state:          nil,
			remote:         []fakeRefreshClient{{key: "a", value: "1"}},
			expectedKeys:   []string{"a"},
			expectedValues: []string{"1"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var diags diag.Diagnostics

			// Build the state types.List from the typed fixtures.
			stateList := TypedArrayToObjectList[fakeRefreshModel](ctx, &diags, test.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics building state list: %s", diags)
			}

			got := RefreshListValueProperties[fakeRefreshModel, fakeRefreshClient](ctx, &diags, stateList, test.remote, getFakeClientKey)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics from RefreshListValueProperties: %s", diags)
			}

			if test.expectNull {
				if !got.IsNull() {
					t.Fatalf("expected null list for empty remote, got %#v", got)
				}
				return
			}

			if got.IsNull() {
				t.Fatalf("expected non-null list, got null")
			}

			refreshed := ObjectListToTypedArray[fakeRefreshModel](ctx, &diags, got)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics unwrapping result list: %s", diags)
			}
			if len(refreshed) != len(test.expectedKeys) {
				t.Fatalf("expected %d items, got %d (%#v)", len(test.expectedKeys), len(refreshed), refreshed)
			}
			for i := range refreshed {
				if k := refreshed[i].Key.ValueString(); k != test.expectedKeys[i] {
					t.Errorf("item %d: expected key %q, got %q", i, test.expectedKeys[i], k)
				}
				if v := refreshed[i].Value.ValueString(); v != test.expectedValues[i] {
					t.Errorf("item %d: expected value %q, got %q", i, test.expectedValues[i], v)
				}
			}
		})
	}
}

// --- RefreshList (string core) ---------------------------------------------

func TestRefreshList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state  []string
		remote []string
		// expected is the exact expected slice. For the empty-remote case this
		// is an empty (len 0) but NON-nil slice.
		expected []string
	}

	tests := map[string]testCase{
		"empty remote returns non-nil empty slice": {
			state:    []string{"a"},
			remote:   []string{},
			expected: []string{},
		},
		"order preserved": {
			state:    []string{"a", "b", "c"},
			remote:   []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		"new remote appended at end": {
			state:    []string{"a"},
			remote:   []string{"a", "c", "b"},
			expected: []string{"a", "c", "b"},
		},
		"state not in remote removed": {
			state:    []string{"a", "b"},
			remote:   []string{"a"},
			expected: []string{"a"},
		},
		"case-insensitive dedup preserves original case": {
			// state "ABC" matches remote "abc" case-insensitively: no duplicate
			// added, and the original state casing ("ABC") is preserved.
			state:    []string{"ABC"},
			remote:   []string{"abc"},
			expected: []string{"ABC"},
		},
		"empty state appends all remote": {
			state:    nil,
			remote:   []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		"both empty returns empty non-nil": {
			state:    []string{},
			remote:   []string{},
			expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := RefreshList(test.state, test.remote)

			// RefreshList always returns a non-nil slice (initialized to []string{}).
			if got == nil {
				t.Fatalf("expected non-nil slice, got nil")
			}
			if len(got) != len(test.expected) {
				t.Fatalf("expected %v (len %d), got %v (len %d)", test.expected, len(test.expected), got, len(got))
			}
			for i := range got {
				if got[i] != test.expected[i] {
					t.Errorf("item %d: expected %q, got %q", i, test.expected[i], got[i])
				}
			}
		})
	}
}

// --- RefreshListValues (string wrapper) ------------------------------------

func TestRefreshListValues(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state  []string
		remote []string
		// expectNull asserts the wrapper returned a null list. For the string
		// wrapper this is NEVER true on an empty remote (see asymmetry test).
		expectNull bool
		expected   []string
	}

	tests := map[string]testCase{
		"empty remote yields non-null empty list": {
			state:      []string{"a"},
			remote:     []string{},
			expectNull: false,
			expected:   []string{},
		},
		"values round-trip with order preserved": {
			state:      []string{"a"},
			remote:     []string{"a", "b"},
			expectNull: false,
			expected:   []string{"a", "b"},
		},
		"null input list with remote yields non-null": {
			// Boundary complement to the typed wrapper: even a null input list plus a
			// non-empty remote yields a NON-null list for the string wrapper.
			state:      nil,
			remote:     []string{"a"},
			expectNull: false,
			expected:   []string{"a"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var diags diag.Diagnostics

			stateList := StringArrayToStringList(ctx, &diags, test.state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics building state list: %s", diags)
			}

			got := RefreshListValues(ctx, &diags, stateList, test.remote)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics from RefreshListValues: %s", diags)
			}

			if got.IsNull() != test.expectNull {
				t.Fatalf("expected IsNull()==%v, got %v (%#v)", test.expectNull, got.IsNull(), got)
			}

			result := StringListToStringArray(ctx, &diags, got)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics unwrapping result list: %s", diags)
			}
			if len(result) != len(test.expected) {
				t.Fatalf("expected %v (len %d), got %v (len %d)", test.expected, len(test.expected), result, len(result))
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("item %d: expected %q, got %q", i, test.expected[i], result[i])
				}
			}
		})
	}
}

// --- Null-vs-empty asymmetry (PLAN 4e / spec D5) ---------------------------

// TestRefreshEmptyRemoteAsymmetry pins the INTENDED behavioral asymmetry between
// the typed and string refresh wrappers for the identical scenario (state holds
// exactly one item, remote is empty):
//   - The TYPED wrapper (RefreshListValueProperties) returns a NULL list, because
//     refreshListProperties returns nil and TypedArrayToObjectList(nil) -> ListNull.
//   - The STRING wrapper (RefreshListValues) returns a NON-null, empty list,
//     because RefreshList returns []string{} and StringArrayToStringList only
//     nulls on a nil input.
//
// This divergence is intentional (the CC3 bug-bash finding was dropped); this
// test exists to lock it in so an accidental "fix" to one side is caught.
func TestRefreshEmptyRemoteAsymmetry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var diags diag.Diagnostics

	// Typed side: one state item, empty remote -> null list.
	typedState := TypedArrayToObjectList[fakeRefreshModel](ctx, &diags, []fakeRefreshModel{newFakeModel("a", "1")})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building typed state list: %s", diags)
	}
	typedResult := RefreshListValueProperties[fakeRefreshModel, fakeRefreshClient](ctx, &diags, typedState, []fakeRefreshClient{}, getFakeClientKey)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics from RefreshListValueProperties: %s", diags)
	}
	if !typedResult.IsNull() {
		t.Errorf("typed wrapper: expected null list on empty remote, got non-null %#v", typedResult)
	}

	// String side: one state item, empty remote -> non-null empty list.
	stringState := StringArrayToStringList(ctx, &diags, []string{"a"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building string state list: %s", diags)
	}
	stringResult := RefreshListValues(ctx, &diags, stringState, []string{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics from RefreshListValues: %s", diags)
	}
	if stringResult.IsNull() {
		t.Errorf("string wrapper: expected non-null list on empty remote, got null")
	}
	if l := len(stringResult.Elements()); l != 0 {
		t.Errorf("string wrapper: expected empty list (len 0) on empty remote, got len %d", l)
	}
}
