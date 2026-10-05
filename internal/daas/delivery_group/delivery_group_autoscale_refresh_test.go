// Copyright © 2026. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"testing"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newPoolSizeScheduleResponse builds a single remote pool size schedule entry.
func newPoolSizeScheduleResponse(timeRange string, poolSize int32) citrixorchestration.PoolSizeScheduleResponseModel {
	schedule := citrixorchestration.PoolSizeScheduleResponseModel{}
	schedule.SetTimeRange(timeRange)
	schedule.SetPoolSize(poolSize)
	return schedule
}

// newPowerTimeSchemeResponse builds a remote power time scheme.
func newPowerTimeSchemeResponse(displayName string, daysOfWeek []string, peakTimeRanges []string, poolUsingPercentage bool, schedules ...citrixorchestration.PoolSizeScheduleResponseModel) citrixorchestration.PowerTimeSchemeResponseModel {
	days := []citrixorchestration.TimeSchemeDays{}
	for _, day := range daysOfWeek {
		days = append(days, citrixorchestration.TimeSchemeDays(day))
	}

	scheme := citrixorchestration.PowerTimeSchemeResponseModel{}
	scheme.SetDisplayName(displayName)
	scheme.SetDaysOfWeek(days)
	scheme.SetPeakTimeRanges(peakTimeRanges)
	scheme.SetPoolUsingPercentage(poolUsingPercentage)
	scheme.SetPoolSizeSchedule(schedules)
	return scheme
}

// poolSizeScheduleList builds the Terraform list value for a scheme's pool_size_schedules.
func poolSizeScheduleList(ctx context.Context, t *testing.T, schedules ...PowerTimeSchemePoolSizeScheduleRequestModel) types.List {
	t.Helper()
	var diags diag.Diagnostics
	list := util.TypedArrayToObjectList(ctx, &diags, schedules)
	if diags.HasError() {
		t.Fatalf("failed to construct pool_size_schedules list: %s", diags)
	}
	return list
}

// poolSizeSchedule is a terser fixture constructor for a state-side schedule entry.
func poolSizeSchedule(timeRange string, poolSize int64) PowerTimeSchemePoolSizeScheduleRequestModel {
	return PowerTimeSchemePoolSizeScheduleRequestModel{
		TimeRange: types.StringValue(timeRange),
		PoolSize:  types.Int64Value(poolSize),
	}
}

// stringSetValues extracts a types.Set of strings for comparison.
func stringSetValues(ctx context.Context, t *testing.T, set types.Set) []string {
	t.Helper()
	var diags diag.Diagnostics
	values := util.StringSetToStringArray(ctx, &diags, set)
	if diags.HasError() {
		t.Fatalf("failed to read string set: %s", diags)
	}
	return values
}

// assertSameElements compares two string slices ignoring order, since days_of_week and
// peak_time_ranges are sets.
func assertSameElements(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: expected %v, got %v", label, want, got)
	}
	counts := map[string]int{}
	for _, v := range want {
		counts[v]++
	}
	for _, v := range got {
		counts[v]--
	}
	for value, remaining := range counts {
		if remaining != 0 {
			t.Fatalf("%s: expected %v, got %v (mismatch on %q)", label, want, got, value)
		}
	}
}

// TestDeliveryGroupPowerTimeSchemeRefreshListItem is the regression test for GitHub issue #374.
// Before the fix, a scheme already present in state had ONLY its pool_size_schedules refreshed,
// so remote changes to days_of_week, peak_time_ranges and pool_using_percentage were discarded
// and terraform plan reported no drift.
func TestDeliveryGroupPowerTimeSchemeRefreshListItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type testCase struct {
		state                   DeliveryGroupPowerTimeScheme
		remote                  citrixorchestration.PowerTimeSchemeResponseModel
		expectDaysOfWeek        []string
		expectPeakTimeRanges    []string
		expectPoolUsePercentage bool
		expectSchedules         []PowerTimeSchemePoolSizeScheduleRequestModel
		expectSchedulesNull     bool
	}

	baseState := func(ctx context.Context, t *testing.T) DeliveryGroupPowerTimeScheme {
		return DeliveryGroupPowerTimeScheme{
			DisplayName:         types.StringValue("Weekdays"),
			DaysOfWeek:          weekdaySet(ctx, t, "Monday"),
			PeakTimeRanges:      weekdaySet(ctx, t, "09:00-17:00"),
			PoolUsingPercentage: types.BoolValue(false),
			PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("09:00-17:00", 5)),
		}
	}

	tests := map[string]testCase{
		// The exact issue #374 scenario: days changed in Studio.
		"refreshes days_of_week from remote": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Tuesday", "Wednesday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5)),
			expectDaysOfWeek:        []string{"Tuesday", "Wednesday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5)},
		},
		"refreshes peak_time_ranges from remote": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"08:00-18:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"08:00-18:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5)},
		},
		"refreshes pool_using_percentage from remote": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, true, newPoolSizeScheduleResponse("09:00-17:00", 5)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: true,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5)},
		},
		// Second half of the bug: preserveOrderInPoolSizeSchedule left the exists-branch empty.
		"updates pool_size for an existing time_range": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 9)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 9)},
		},
		// The behavior the original bespoke code existed to provide: state order wins.
		"preserves pool_size_schedules order from state": {
			state: DeliveryGroupPowerTimeScheme{
				DisplayName:         types.StringValue("Weekdays"),
				DaysOfWeek:          weekdaySet(ctx, t, "Monday"),
				PeakTimeRanges:      weekdaySet(ctx, t, "09:00-17:00"),
				PoolUsingPercentage: types.BoolValue(false),
				PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("17:00-23:30", 2), poolSizeSchedule("09:00-17:00", 5)),
			},
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 6), newPoolSizeScheduleResponse("17:00-23:30", 3)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("17:00-23:30", 3), poolSizeSchedule("09:00-17:00", 6)},
		},
		"appends a new remote time_range": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5), newPoolSizeScheduleResponse("17:00-23:30", 2)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5), poolSizeSchedule("17:00-23:30", 2)},
		},
		"removes a time_range missing from remote": {
			state: DeliveryGroupPowerTimeScheme{
				DisplayName:         types.StringValue("Weekdays"),
				DaysOfWeek:          weekdaySet(ctx, t, "Monday"),
				PeakTimeRanges:      weekdaySet(ctx, t, "09:00-17:00"),
				PoolUsingPercentage: types.BoolValue(false),
				PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("09:00-17:00", 5), poolSizeSchedule("17:00-23:30", 2)),
			},
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5)},
		},
		// pool_size has an AtLeast(1) validator, so API-side zero-sized fill must never reach state.
		"drops pool size schedules with pool size 0": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5), newPoolSizeScheduleResponse("17:00-23:30", 0)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedules:         []PowerTimeSchemePoolSizeScheduleRequestModel{poolSizeSchedule("09:00-17:00", 5)},
		},
		"nulls pool_size_schedules when every remote entry is zero sized": {
			state:                   baseState(ctx, t),
			remote:                  newPowerTimeSchemeResponse("Weekdays", []string{"Monday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 0)),
			expectDaysOfWeek:        []string{"Monday"},
			expectPeakTimeRanges:    []string{"09:00-17:00"},
			expectPoolUsePercentage: false,
			expectSchedulesNull:     true,
		},
	}

	for name, tc := range tests {
		t.Run(fmt.Sprintf("DeliveryGroupPowerTimeScheme.RefreshListItem - %s", name), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			var diags diag.Diagnostics

			refreshed, ok := tc.state.RefreshListItem(ctx, &diags, tc.remote).(DeliveryGroupPowerTimeScheme)
			if !ok {
				t.Fatalf("RefreshListItem did not return a DeliveryGroupPowerTimeScheme")
			}
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}

			assertSameElements(t, "days_of_week", stringSetValues(ctx, t, refreshed.DaysOfWeek), tc.expectDaysOfWeek)
			assertSameElements(t, "peak_time_ranges", stringSetValues(ctx, t, refreshed.PeakTimeRanges), tc.expectPeakTimeRanges)

			if refreshed.PoolUsingPercentage.ValueBool() != tc.expectPoolUsePercentage {
				t.Fatalf("pool_using_percentage: expected %t, got %t", tc.expectPoolUsePercentage, refreshed.PoolUsingPercentage.ValueBool())
			}

			if tc.expectSchedulesNull {
				if !refreshed.PoolSizeSchedules.IsNull() {
					t.Fatalf("expected pool_size_schedules to be null, got %s", refreshed.PoolSizeSchedules)
				}
				return
			}

			gotSchedules := util.ObjectListToTypedArray[PowerTimeSchemePoolSizeScheduleRequestModel](ctx, &diags, refreshed.PoolSizeSchedules)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics reading pool_size_schedules: %s", diags)
			}
			if len(gotSchedules) != len(tc.expectSchedules) {
				t.Fatalf("pool_size_schedules: expected %d entries, got %d (%v)", len(tc.expectSchedules), len(gotSchedules), gotSchedules)
			}
			for i, want := range tc.expectSchedules {
				if gotSchedules[i].TimeRange.ValueString() != want.TimeRange.ValueString() {
					t.Fatalf("pool_size_schedules[%d].time_range: expected %q, got %q", i, want.TimeRange.ValueString(), gotSchedules[i].TimeRange.ValueString())
				}
				if gotSchedules[i].PoolSize.ValueInt64() != want.PoolSize.ValueInt64() {
					t.Fatalf("pool_size_schedules[%d].pool_size: expected %d, got %d", i, want.PoolSize.ValueInt64(), gotSchedules[i].PoolSize.ValueInt64())
				}
			}
		})
	}
}

// TestPowerTimeSchemePoolSizeScheduleRequestModelRefreshListItem pins that both attributes
// are taken from the remote model rather than carried over from state.
func TestPowerTimeSchemePoolSizeScheduleRequestModelRefreshListItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics

	state := poolSizeSchedule("09:00-17:00", 5)
	refreshed, ok := state.RefreshListItem(ctx, &diags, newPoolSizeScheduleResponse("09:00-17:00", 11)).(PowerTimeSchemePoolSizeScheduleRequestModel)
	if !ok {
		t.Fatalf("RefreshListItem did not return a PowerTimeSchemePoolSizeScheduleRequestModel")
	}
	if refreshed.TimeRange.ValueString() != "09:00-17:00" {
		t.Fatalf("time_range: expected %q, got %q", "09:00-17:00", refreshed.TimeRange.ValueString())
	}
	if refreshed.PoolSize.ValueInt64() != 11 {
		t.Fatalf("pool_size: expected 11, got %d", refreshed.PoolSize.ValueInt64())
	}
}

// TestRefreshPowerTimeSchemesPreservesOrder exercises the whole list refresh through the
// shared helper, which is how updatePlanWithAutoscaleSettings drives it.
func TestRefreshPowerTimeSchemesPreservesOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics

	weekdays := DeliveryGroupPowerTimeScheme{
		DisplayName:         types.StringValue("Weekdays"),
		DaysOfWeek:          weekdaySet(ctx, t, "Monday"),
		PeakTimeRanges:      weekdaySet(ctx, t, "09:00-17:00"),
		PoolUsingPercentage: types.BoolValue(false),
		PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("09:00-17:00", 5)),
	}
	weekend := DeliveryGroupPowerTimeScheme{
		DisplayName:         types.StringValue("Weekend"),
		DaysOfWeek:          weekdaySet(ctx, t, "Saturday"),
		PeakTimeRanges:      weekdaySet(ctx, t, "10:00-14:00"),
		PoolUsingPercentage: types.BoolValue(false),
		PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("10:00-14:00", 1)),
	}

	state := util.TypedArrayToObjectList(ctx, &diags, []DeliveryGroupPowerTimeScheme{weekdays, weekend})
	if diags.HasError() {
		t.Fatalf("failed to construct state list: %s", diags)
	}

	// Remote returns the schemes in the opposite order, and Weekdays has different days.
	remote := []citrixorchestration.PowerTimeSchemeResponseModel{
		newPowerTimeSchemeResponse("Weekend", []string{"Saturday"}, []string{"10:00-14:00"}, false, newPoolSizeScheduleResponse("10:00-14:00", 1)),
		newPowerTimeSchemeResponse("Weekdays", []string{"Tuesday", "Wednesday"}, []string{"09:00-17:00"}, false, newPoolSizeScheduleResponse("09:00-17:00", 5)),
	}

	refreshedList := util.RefreshListValueProperties[DeliveryGroupPowerTimeScheme, citrixorchestration.PowerTimeSchemeResponseModel](ctx, &diags, state, remote, util.GetOrchestrationPowerTimeSchemeKey)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	refreshed := util.ObjectListToTypedArray[DeliveryGroupPowerTimeScheme](ctx, &diags, refreshedList)
	if len(refreshed) != 2 {
		t.Fatalf("expected 2 power time schemes, got %d", len(refreshed))
	}
	if refreshed[0].DisplayName.ValueString() != "Weekdays" || refreshed[1].DisplayName.ValueString() != "Weekend" {
		t.Fatalf("expected state order [Weekdays Weekend], got [%s %s]", refreshed[0].DisplayName.ValueString(), refreshed[1].DisplayName.ValueString())
	}
	assertSameElements(t, "Weekdays days_of_week", stringSetValues(ctx, t, refreshed[0].DaysOfWeek), []string{"Tuesday", "Wednesday"})
}

// autoscaleStateWithRestrictTag builds a DeliveryGroupResourceModel whose autoscale_settings
// already carry a restrict autoscale tag and both min idle percentages.
func autoscaleStateWithRestrictTag(ctx context.Context, t *testing.T, restrictTag types.String) DeliveryGroupResourceModel {
	t.Helper()
	var diags diag.Diagnostics

	// power_time_schemes has to be a properly typed null list; a zero-value types.List carries a
	// dynamic pseudo type and fails object conversion.
	powerTimeSchemeAttributes, err := util.ResourceAttributeMapFromObject(DeliveryGroupPowerTimeScheme{})
	if err != nil {
		t.Fatalf("failed to build power_time_schemes attribute map: %s", err)
	}

	autoscale := util.TypedObjectToObjectValue(ctx, &diags, DeliveryGroupPowerManagementSettings{
		AutoscaleEnabled:     types.BoolValue(true),
		RestrictAutoscaleTag: restrictTag,
		RestrictAutoscaleMinIdleUntaggedPercentDuringPeak:    types.Int32Value(10),
		RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak: types.Int32Value(20),
		Timezone:         types.StringNull(),
		PowerTimeSchemes: types.ListNull(types.ObjectType{AttrTypes: powerTimeSchemeAttributes}),
	})
	if diags.HasError() {
		t.Fatalf("failed to construct autoscale_settings object: %s", diags)
	}

	return DeliveryGroupResourceModel{AutoscaleSettings: autoscale}
}

// refreshAutoscale runs updatePlanWithAutoscaleSettings and returns the refreshed settings.
// No mock client is needed: the only client use is getDeliveryGroupAllocationType, which
// iterates AssociatedMachineCatalogs; a zero-value set yields nil so the client is never touched.
func refreshAutoscale(ctx context.Context, t *testing.T, state DeliveryGroupResourceModel, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel) DeliveryGroupPowerManagementSettings {
	t.Helper()
	var diags diag.Diagnostics

	refreshed := state.updatePlanWithAutoscaleSettings(ctx, &diags, nil, deliveryGroup, &citrixorchestration.PowerTimeSchemeResponseModelCollection{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	return util.ObjectValueToTypedObject[DeliveryGroupPowerManagementSettings](ctx, &diags, refreshed.AutoscaleSettings)
}

const (
	restrictTagId   = "8a7b6c5d-4e3f-2a1b-0c9d-8e7f6a5b4c3d"
	restrictTagName = "MyTag"
)

// newDeliveryGroupWithRestrictTag builds a delivery group response carrying a restrict autoscale tag.
func newDeliveryGroupWithRestrictTag(peakPercent, offPeakPercent int32) *citrixorchestration.DeliveryGroupDetailResponseModel {
	tag := citrixorchestration.RefResponseModel{}
	tag.SetId(restrictTagId)
	tag.SetName(restrictTagName)

	deliveryGroup := &citrixorchestration.DeliveryGroupDetailResponseModel{}
	deliveryGroup.SetAutoScaleEnabled(true)
	deliveryGroup.SetRestrictAutoscaleTag(tag)
	deliveryGroup.SetRestrictAutoscaleMinIdleUntaggedPercentDuringPeak(peakPercent)
	deliveryGroup.SetRestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak(offPeakPercent)
	return deliveryGroup
}

// TestUpdatePlanWithAutoscaleSettingsRestrictAutoscaleTag covers the second refresh gap: the
// restrict autoscale tag and its min idle percentages were only ever refreshed inside an
// "if tag != nil" guard with no else branch, so removing the tag outside Terraform left stale
// values in state forever. It also pins the getRestrictToTagValue reuse so a configuration that
// specifies the tag by id does not get a perpetual diff against the remote tag name.
func TestUpdatePlanWithAutoscaleSettingsRestrictAutoscaleTag(t *testing.T) {
	t.Parallel()

	t.Run("updatePlanWithAutoscaleSettings - clears restrict_autoscale_tag when removed remotely", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		deliveryGroup := &citrixorchestration.DeliveryGroupDetailResponseModel{}
		deliveryGroup.SetAutoScaleEnabled(true)

		autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringValue(restrictTagName)), deliveryGroup)

		if !autoscale.RestrictAutoscaleTag.IsNull() {
			t.Fatalf("expected restrict_autoscale_tag to be null, got %q", autoscale.RestrictAutoscaleTag.ValueString())
		}
		if !autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.IsNull() {
			t.Fatalf("expected peak_restrict_min_idle_untagged_percent to be null, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.ValueInt32())
		}
		if !autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.IsNull() {
			t.Fatalf("expected off_peak_restrict_min_idle_untagged_percent to be null, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.ValueInt32())
		}
	})

	t.Run("updatePlanWithAutoscaleSettings - keeps configured tag id when remote reports the tag name", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringValue(restrictTagId)), newDeliveryGroupWithRestrictTag(10, 20))

		if autoscale.RestrictAutoscaleTag.ValueString() != restrictTagId {
			t.Fatalf("expected restrict_autoscale_tag to stay %q, got %q", restrictTagId, autoscale.RestrictAutoscaleTag.ValueString())
		}
	})

	t.Run("updatePlanWithAutoscaleSettings - uses tag name when state has no tag", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringNull()), newDeliveryGroupWithRestrictTag(10, 20))

		if autoscale.RestrictAutoscaleTag.ValueString() != restrictTagName {
			t.Fatalf("expected restrict_autoscale_tag %q, got %q", restrictTagName, autoscale.RestrictAutoscaleTag.ValueString())
		}
	})

	t.Run("updatePlanWithAutoscaleSettings - refreshes min idle percents", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringValue(restrictTagName)), newDeliveryGroupWithRestrictTag(35, 45))

		if autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.ValueInt32() != 35 {
			t.Fatalf("expected peak_restrict_min_idle_untagged_percent 35, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.ValueInt32())
		}
		if autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.ValueInt32() != 45 {
			t.Fatalf("expected off_peak_restrict_min_idle_untagged_percent 45, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.ValueInt32())
		}
	})

	t.Run("updatePlanWithAutoscaleSettings - nulls min idle percents when negative", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringValue(restrictTagName)), newDeliveryGroupWithRestrictTag(-1, -1))

		if !autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.IsNull() {
			t.Fatalf("expected peak_restrict_min_idle_untagged_percent to be null, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringPeak.ValueInt32())
		}
		if !autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.IsNull() {
			t.Fatalf("expected off_peak_restrict_min_idle_untagged_percent to be null, got %d", autoscale.RestrictAutoscaleMinIdleUntaggedPercentDuringOffPeak.ValueInt32())
		}
	})
}

// TestUpdatePlanWithAutoscaleSettingsLeavesTimezoneNull is a characterization test. timezone is
// Optional and NOT Computed, and updatePlanWithAutoscaleSettings also runs on Create and Update,
// so refreshing it when state holds null would fail the apply with "Provider produced inconsistent
// result after apply" for everyone who never configured it. Populating timezone on import needs a
// schema change, so this pins the current deliberate behavior.
func TestUpdatePlanWithAutoscaleSettingsLeavesTimezoneNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	deliveryGroup := &citrixorchestration.DeliveryGroupDetailResponseModel{}
	deliveryGroup.SetAutoScaleEnabled(true)
	deliveryGroup.SetTimeZone("GMT Standard Time")

	autoscale := refreshAutoscale(ctx, t, autoscaleStateWithRestrictTag(ctx, t, types.StringNull()), deliveryGroup)

	if !autoscale.Timezone.IsNull() {
		t.Fatalf("expected timezone to stay null, got %q", autoscale.Timezone.ValueString())
	}
}

// TestRefreshPowerTimeSchemesEmptyRemote pins the replacement of the deleted ListNull
// else-branch in updatePlanWithAutoscaleSettings.
func TestRefreshPowerTimeSchemesEmptyRemote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var diags diag.Diagnostics

	state := util.TypedArrayToObjectList(ctx, &diags, []DeliveryGroupPowerTimeScheme{{
		DisplayName:         types.StringValue("Weekdays"),
		DaysOfWeek:          weekdaySet(ctx, t, "Monday"),
		PeakTimeRanges:      weekdaySet(ctx, t, "09:00-17:00"),
		PoolUsingPercentage: types.BoolValue(false),
		PoolSizeSchedules:   poolSizeScheduleList(ctx, t, poolSizeSchedule("09:00-17:00", 5)),
	}})
	if diags.HasError() {
		t.Fatalf("failed to construct state list: %s", diags)
	}

	refreshedList := util.RefreshListValueProperties[DeliveryGroupPowerTimeScheme, citrixorchestration.PowerTimeSchemeResponseModel](ctx, &diags, state, []citrixorchestration.PowerTimeSchemeResponseModel{}, util.GetOrchestrationPowerTimeSchemeKey)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}
	if !refreshedList.IsNull() {
		t.Fatalf("expected a null list when remote has no power time schemes, got %s", refreshedList)
	}
}
