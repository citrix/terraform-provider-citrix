// Copyright © 2026. Citrix Systems, Inc.

package cma_image // whitebox testing in same package

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateTrustedLaunchConfig(t *testing.T) {
	const uri = "https://example.blob.core.windows.net/vhds/guest.vmgs"

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name          string
		gen           types.String
		vtpm          types.Bool
		secureBoot    types.Bool
		guestDisk     types.String
		errorContains []string
	}{
		// XAC-75526: an Unknown value must never be read as an explicit false. Every rule whose
		// operands are not all known has to defer instead of firing backwards.
		{"all unknown defers everything", types.StringUnknown(), types.BoolUnknown(), types.BoolUnknown(), types.StringUnknown(), nil},
		{"unknown secure boot with guest disk set", types.StringValue("V2"), types.BoolUnknown(), types.BoolUnknown(), types.StringValue(uri), nil},
		{"unknown vtpm inside secure boot block", types.StringValue("V2"), types.BoolUnknown(), types.BoolValue(true), types.StringValue(uri), nil},
		{"unknown generation defers V2 rules", types.StringUnknown(), types.BoolValue(true), types.BoolValue(false), types.StringNull(), nil},
		{"unknown guest disk with secure boot on", types.StringValue("V2"), types.BoolValue(true), types.BoolValue(true), types.StringUnknown(), nil},

		// Known-value rules must still fire. Create re-runs this helper against the resolved
		// plan, so these are also the shapes it sees at apply time.
		{"valid trusted launch", types.StringValue("V2"), types.BoolValue(true), types.BoolValue(true), types.StringValue(uri), nil},
		{"secure boot without guest disk", types.StringValue("V2"), types.BoolValue(true), types.BoolValue(true), types.StringNull(),
			[]string{"Guest Disk URI must be specified when Secure Boot is enabled"}},
		{"guest disk without secure boot", types.StringValue("V2"), types.BoolValue(false), types.BoolValue(false), types.StringValue(uri),
			[]string{"Guest Disk URI is only applicable when Secure Boot is enabled"}},
		{"secure boot without vtpm", types.StringValue("V2"), types.BoolValue(false), types.BoolValue(true), types.StringValue(uri),
			[]string{"vTPM must be enabled when Secure Boot is enabled"}},
		{"vtpm on gen1", types.StringValue("V1"), types.BoolValue(true), types.BoolValue(false), types.StringNull(),
			[]string{"vTPM is only supported for V2 generation images"}},
		{"secure boot on gen1", types.StringValue("V1"), types.BoolValue(true), types.BoolValue(true), types.StringValue(uri),
			[]string{"vTPM is only supported for V2 generation images", "Secure Boot is only supported for V2 generation images"}},

		// A null bool resolves to false through booldefault.StaticBool(false), so treating it as
		// an explicit false is correct.
		{"null secure boot with guest disk set", types.StringValue("V2"), types.BoolNull(), types.BoolNull(), types.StringValue(uri),
			[]string{"Guest Disk URI is only applicable when Secure Boot is enabled"}},
		{"null vtpm with secure boot on", types.StringValue("V2"), types.BoolNull(), types.BoolValue(true), types.StringValue(uri),
			[]string{"vTPM must be enabled when Secure Boot is enabled"}},

		// An empty guest_disk_uri counts as unset. It is rejected by the attribute's
		// LengthAtLeast(1) validator, so this helper must not add a contradictory error.
		{"empty guest disk with secure boot off", types.StringValue("V2"), types.BoolValue(false), types.BoolValue(false), types.StringValue(""), nil},

		// machine_generation is Required, but the framework's missing-attribute check runs after
		// ValidateConfig, so a null value still reaches this helper. It must degrade to "not V1".
		{"null generation does not fire V1 rules", types.StringNull(), types.BoolValue(true), types.BoolValue(true), types.StringValue(uri), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := CitrixManagedAzureImageResourceModel{
				MachineGeneration: tt.gen,
				VtpmEnabled:       tt.vtpm,
				SecureBootEnabled: tt.secureBoot,
				GuestDiskUri:      tt.guestDisk,
			}

			var diagnostics diag.Diagnostics
			validateTrustedLaunchConfig(&model, &diagnostics)

			if len(tt.errorContains) == 0 {
				assert.False(t, diagnostics.HasError(), "unexpected diagnostics: %v", diagnostics.Errors())
				return
			}

			assert.True(t, diagnostics.HasError())
			assert.Len(t, diagnostics.Errors(), len(tt.errorContains))
			var details []string
			for _, d := range diagnostics.Errors() {
				details = append(details, d.Detail())
			}
			for _, want := range tt.errorContains {
				assert.Contains(t, details, want)
			}
		})
	}
}
