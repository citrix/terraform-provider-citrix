// Copyright © 2026. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidateRebootSchedules pins the current behavior of validateRebootSchedules
// across its three rule clusters:
//
//	(a) frequency      - Weekly requires days_in_week; Monthly requires week_in_month AND day_in_month.
//	(b) natural-reboot - conflicts between natural_reboot_schedule, reboot_notification_to_users and
//	                     reboot_duration_minutes. EACH conflict branch returns from the whole function.
//	(c) nested         - when reboot_notification_to_users is set, notification_repeat_every_5_minutes
//	                     can only be true when notification_duration_minutes == 15.
//
// Each test case uses a SINGLE-element []DeliveryGroupRebootSchedule so an early-returning
// branch cannot be masked by a later schedule. Because several branches reuse the same error
// summary ("Missing Attribute Configuration"), expected-error cases assert on the error DETAIL
// substring rather than the summary alone.
func TestValidateRebootSchedules(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// notificationObject builds a non-null reboot_notification_to_users object.
	notificationObject := func(durationMinutes types.Int64, repeatEvery5 types.Bool) types.Object {
		var diags diag.Diagnostics
		obj := util.TypedObjectToObjectValue(ctx, &diags, DeliveryGroupRebootNotificationToUsers{
			NotificationDurationMinutes:     durationMinutes,
			NotificationRepeatEvery5Minutes: repeatEvery5,
			NotificationTitle:               types.StringValue("t"),
			NotificationMessage:             types.StringValue("m"),
		})
		if diags.HasError() {
			t.Fatalf("failed to construct notification object: %s", diags)
		}
		return obj
	}

	type testCase struct {
		schedule      DeliveryGroupRebootSchedule
		expectError   bool
		detailContain string // substring expected in at least one error Detail() when expectError is true
	}

	tests := map[string]testCase{
		// ---- Cluster (a): frequency (does NOT return; rest of schedule kept valid) ----
		"weekly missing days_in_week": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Weekly"),
				DaysInWeek:               types.Set{}, // zero-value set: Elements() len 0
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError:   true,
			detailContain: "Days in week must be specified",
		},
		"weekly with days_in_week ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Weekly"),
				DaysInWeek:               weekdaySet(ctx, t, "Monday"),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError: false,
		},
		"monthly missing week_in_month": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Monthly"),
				WeekInMonth:              types.StringNull(),
				DayInMonth:               types.StringValue("1"),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError:   true,
			detailContain: "Week in month must be specified",
		},
		"monthly missing day_in_month": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Monthly"),
				WeekInMonth:              types.StringValue("First"),
				DayInMonth:               types.StringNull(),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError:   true,
			detailContain: "Day in month must be specified",
		},
		"monthly with week_and_day ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Monthly"),
				WeekInMonth:              types.StringValue("First"),
				DayInMonth:               types.StringValue("1"),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError: false,
		},

		// ---- Cluster (b): natural-reboot conflicts (each branch returns) ----
		"natural true with notification set": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                              types.StringValue("Daily"),
				UseNaturalRebootSchedule:               types.BoolValue(true),
				DeliveryGroupRebootNotificationToUsers: notificationObject(types.Int64Value(15), types.BoolValue(true)),
			},
			expectError:   true,
			detailContain: "Reboot notification to users cannot be set when using natural reboot",
		},
		"natural true with reboot_duration set": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Daily"),
				UseNaturalRebootSchedule: types.BoolValue(true),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError:   true,
			detailContain: "Reboot duration minutes cannot be set when using natural reboot",
		},
		"natural false with reboot_duration null": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Daily"),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Null(),
			},
			expectError:   true,
			detailContain: "Reboot duration minutes must be specified when natural_reboot_schedule is set to false",
		},
		"natural false with reboot_duration set ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Daily"),
				UseNaturalRebootSchedule: types.BoolValue(false),
				RebootDurationMinutes:    types.Int64Value(30),
			},
			expectError: false,
		},
		"natural true with nothing set ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                types.StringValue("Daily"),
				UseNaturalRebootSchedule: types.BoolValue(true),
				RebootDurationMinutes:    types.Int64Null(),
			},
			expectError: false,
		},

		// ---- Cluster (c): nested notification (reached only after (b) passes) ----
		"notification repeat true duration not 15": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                              types.StringValue("Daily"),
				UseNaturalRebootSchedule:               types.BoolValue(false),
				RebootDurationMinutes:                  types.Int64Value(30),
				DeliveryGroupRebootNotificationToUsers: notificationObject(types.Int64Value(5), types.BoolValue(true)),
			},
			expectError:   true,
			detailContain: "NotificationRepeatEvery5Minutes can only be set to true when NotificationDurationMinutes is 15 minutes",
		},
		"notification repeat true duration 15 ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                              types.StringValue("Daily"),
				UseNaturalRebootSchedule:               types.BoolValue(false),
				RebootDurationMinutes:                  types.Int64Value(30),
				DeliveryGroupRebootNotificationToUsers: notificationObject(types.Int64Value(15), types.BoolValue(true)),
			},
			expectError: false,
		},
		"notification repeat null duration not 15 ok": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                              types.StringValue("Daily"),
				UseNaturalRebootSchedule:               types.BoolValue(false),
				RebootDurationMinutes:                  types.Int64Value(30),
				DeliveryGroupRebootNotificationToUsers: notificationObject(types.Int64Value(5), types.BoolNull()),
			},
			expectError: false,
		},

		// ---- Fully-valid happy path ----
		"fully valid weekly schedule": {
			schedule: DeliveryGroupRebootSchedule{
				Frequency:                              types.StringValue("Weekly"),
				DaysInWeek:                             weekdaySet(ctx, t, "Monday"),
				UseNaturalRebootSchedule:               types.BoolValue(false),
				RebootDurationMinutes:                  types.Int64Value(30),
				DeliveryGroupRebootNotificationToUsers: notificationObject(types.Int64Value(15), types.BoolValue(true)),
			},
			expectError: false,
		},
	}

	for name, test := range tests {
		t.Run(fmt.Sprintf("ValidateRebootSchedules - %s", name), func(t *testing.T) {
			t.Parallel()

			localCtx := context.Background()
			var diags diag.Diagnostics

			validateRebootSchedules(localCtx, &diags, []DeliveryGroupRebootSchedule{test.schedule})

			if test.expectError {
				if !diags.HasError() {
					t.Fatalf("expected error, got none")
				}
				found := false
				for _, err := range diags.Errors() {
					if strings.Contains(err.Detail(), test.detailContain) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected an error with detail containing %q, got: %s", test.detailContain, diags)
				}
			} else if diags.HasError() {
				t.Fatalf("expected no error, got: %s", diags)
			}
		})
	}
}

// weekdaySet builds a non-empty types.Set of weekday strings.
func weekdaySet(ctx context.Context, t *testing.T, days ...string) types.Set {
	t.Helper()
	var diags diag.Diagnostics
	set := util.StringArrayToStringSet(ctx, &diags, days)
	if diags.HasError() {
		t.Fatalf("failed to construct days_in_week set: %s", diags)
	}
	return set
}

// TestValidateRebootSchedulesNoSchedules pins the empty-input boundary: with no
// schedules the loop body never runs, so no diagnostics are produced.
func TestValidateRebootSchedulesNoSchedules(t *testing.T) {
	t.Parallel()
	var diags diag.Diagnostics
	validateRebootSchedules(context.Background(), &diags, []DeliveryGroupRebootSchedule{})
	if diags.HasError() {
		t.Fatalf("expected no error for empty schedule list, got: %s", diags)
	}
}

// TestValidateRebootSchedulesMultiple pins multi-element behavior for the non-returning
// frequency cluster: validation runs for EVERY schedule, so distinct errors accumulate.
//
// Note: the schedules must produce DIFFERENT diagnostics. terraform-plugin-framework's
// diag.Diagnostics.Append deduplicates byte-identical diagnostics (same severity, summary,
// detail and attribute path), so two identical Weekly-missing-days schedules would collapse
// to a single error. Using a Weekly-missing-days schedule plus a Monthly-missing-week
// schedule yields two distinct diagnostics (days_in_week + week_in_month) -> exactly two.
// Both keep the rest valid (natural_reboot_schedule=false, reboot_duration_minutes set) so
// cluster (b) does not early-return.
func TestValidateRebootSchedulesMultiple(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	weeklyMissingDays := DeliveryGroupRebootSchedule{
		Frequency:                types.StringValue("Weekly"),
		DaysInWeek:               types.Set{}, // zero-value set: Elements() len 0
		UseNaturalRebootSchedule: types.BoolValue(false),
		RebootDurationMinutes:    types.Int64Value(30),
	}
	monthlyMissingWeek := DeliveryGroupRebootSchedule{
		Frequency:                types.StringValue("Monthly"),
		WeekInMonth:              types.StringNull(),
		DayInMonth:               types.StringValue("1"),
		UseNaturalRebootSchedule: types.BoolValue(false),
		RebootDurationMinutes:    types.Int64Value(30),
	}
	var diags diag.Diagnostics
	validateRebootSchedules(ctx, &diags, []DeliveryGroupRebootSchedule{weeklyMissingDays, monthlyMissingWeek})
	if got := len(diags.Errors()); got != 2 {
		t.Fatalf("expected exactly 2 accumulated errors across 2 schedules, got %d: %s", got, diags)
	}
}
