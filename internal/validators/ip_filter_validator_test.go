// Copyright © 2026. Citrix Systems, Inc.

package validators

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIPFilterValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		ipAddress   types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown String": {
			ipAddress: types.StringUnknown(),
		},
		"null String": {
			ipAddress: types.StringNull(),
		},
		"valid ip *.*.*.*": {
			ipAddress: types.StringValue("*.*.*.*"),
		},
		"valid ip 12.0.0.*": {
			ipAddress: types.StringValue("12.0.0.*"),
		},
		"valid ip 12.0.*.*": {
			ipAddress: types.StringValue("12.0.*.*"),
		},
		"valid ip 12.*.*.*": {
			ipAddress: types.StringValue("12.*.*.*"),
		},
		"valid ip 12.0.0.0": {
			ipAddress: types.StringValue("12.0.0.0"),
		},
		"valid ip 12.0.0.1-12.0.0.70": {
			ipAddress: types.StringValue("12.0.0.1-12.0.0.70"),
		},
		"valid ip 12.0.0.1 mask 24": {
			ipAddress: types.StringValue("12.0.0.1/24"),
		},
		"valid ip 2001:0db8:3c4d:0015:0:0:abcd:ef12": {
			ipAddress: types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12"),
		},
		"valid ip 2001:0db8:3c4d:0015:0:0::": {
			ipAddress: types.StringValue("2001:0db8:3c4d:0015:0:0::"),
		},
		"valid ip ::3c4d:0015:0:0:abcd:ef12": {
			ipAddress: types.StringValue("::3c4d:0015:0:0:abcd:ef12"),
		},
		"valid ip 2001:0db8:3c4d:0015:0:0:abcd:ef12 mask 39": {
			ipAddress: types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12/39"),
		},
		"valid ip 2001:0db8:3c4d:0015:0:0:abcd:ef12 mask 40": {
			ipAddress: types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12/40"),
		},
		"valid ip 2001:db8::/64": {
			ipAddress: types.StringValue("2001:db8::/64"),
		},
		"valid ip 2001:db8::/128": {
			ipAddress: types.StringValue("2001:db8::/128"),
		},
		"valid ip 2001:db8::1-2001:db8::2": {
			ipAddress: types.StringValue("2001:db8::1-2001:db8::2"),
		},
		"invalid ip *.1.*.*": {
			ipAddress:   types.StringValue("*.1.*.*"),
			expectError: true,
		},
		"invalid ip 12.0.*.0": {
			ipAddress:   types.StringValue("12.0.*.0"),
			expectError: true,
		},
		"invalid ip 12.*.0.0": {
			ipAddress:   types.StringValue("12.*.0.0"),
			expectError: true,
		},
		"invalid ip 12.*.*.1": {
			ipAddress:   types.StringValue("12.*.*.1"),
			expectError: true,
		},
		"invalid ip 12.*.1.1": {
			ipAddress:   types.StringValue("12.*.1.1"),
			expectError: true,
		},
		"invalid ip *.0.0.0": {
			ipAddress:   types.StringValue("*.0.0.0"),
			expectError: true,
		},
		"invalid ip 12.0.0.* mask 16": {
			ipAddress:   types.StringValue("12.0.0.*/16"),
			expectError: true,
		},
		"invalid ip 12.0.*.": {
			ipAddress:   types.StringValue("12.0.*."),
			expectError: true,
		},
		"invalid ip 12.0.0.256": {
			ipAddress:   types.StringValue("12.0.0.256"),
			expectError: true,
		},
		"invalid ip 12.0.0.1 mask 24-12.0.0.70": {
			ipAddress:   types.StringValue("12.0.0.1/24-12.0.0.70"),
			expectError: true,
		},
		"invalid ip 12.0.0.1-12.0.0.70 mask 24": {
			ipAddress:   types.StringValue("12.0.0.1-12.0.0.70/24"),
			expectError: true,
		},
		"invalid ip 12.0.0.1 mask 24-12.0.0.70 mask 24": {
			ipAddress:   types.StringValue("12.0.0.1/24-12.0.0.70/24"),
			expectError: true,
		},
		"invalid ip 12.0.0.70-12.0.0.70": {
			ipAddress:   types.StringValue("12.0.0.70-12.0.0.70"),
			expectError: true,
		},
		"invalid ip 12.0.0.71-12.0.0.70": {
			ipAddress:   types.StringValue("12.0.0.71-12.0.0.70"),
			expectError: true,
		},
		"invalid ip 12.0.0.1 mask 40": {
			ipAddress:   types.StringValue("12.0.0.1/40"),
			expectError: true,
		},
		"invalid ip 2001:0db8:3c4d:0015:0:0:abcd:ef12 mask asd": {
			ipAddress:   types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12/asd"),
			expectError: true,
		},
		"invalid ip 2001:0db8:3c4d:0015:0:0:": {
			ipAddress:   types.StringValue("2001:0db8:3c4d:0015:0:0:"),
			expectError: true,
		},
		"invalid ip 2001:0db8:3c4d:0015:0:0:::": {
			ipAddress:   types.StringValue("2001:0db8:3c4d:0015:0:0:::"),
			expectError: true,
		},
		"invalid ip :3c4d:0015:0:0:abcd:ef12": {
			ipAddress:   types.StringValue(":3c4d:0015:0:0:abcd:ef12"),
			expectError: true,
		},
		"invalid ip :::3c4d:0015:0:0:abcd:ef12": {
			ipAddress:   types.StringValue(":::3c4d:0015:0:0:abcd:ef12"),
			expectError: true,
		},
		"invalid ip 2001:0db8:3c4d:0015:0:0:abcd:ef12 mask null": {
			ipAddress:   types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12/"),
			expectError: true,
		},
		"invalid ip empty mask": {
			ipAddress:   types.StringValue("2001:0db8:3c4d:0015:0:0:abcd:ef12/ "),
			expectError: true,
		},
		"invalid ip ipv6 with asterisk": {
			ipAddress:   types.StringValue("2001:0db8:*:0015:0:0:abcd:ef12"),
			expectError: true,
		},
		"invalid ip ipv6 with asterisk and mask": {
			ipAddress:   types.StringValue("2001:0db8:*:0015:0:0:abcd:ef12/24"),
			expectError: true,
		},
		"invalid ip empty string": {
			ipAddress:   types.StringValue(""),
			expectError: true,
		},
		"invalid ip 2001:db8::2-2001:db8::1": {
			ipAddress:   types.StringValue("2001:db8::2-2001:db8::1"),
			expectError: true,
		},
		"invalid ip 2001:db8::1-2001:db8::1": {
			ipAddress:   types.StringValue("2001:db8::1-2001:db8::1"),
			expectError: true,
		},
		"invalid ip 12.0.0.1-2001:db8::1": {
			ipAddress:   types.StringValue("12.0.0.1-2001:db8::1"),
			expectError: true,
		},
		// --- XAC-71946 edge-case hardening: boundaries, error paths, family-agnostic
		// ordering, and the former 32-bit-overflow path removed with convertIpAddressToNumber ---
		"invalid ip multiple hyphens 12.0.0.1-12.0.0.2-12.0.0.3": {
			ipAddress:   types.StringValue("12.0.0.1-12.0.0.2-12.0.0.3"),
			expectError: true,
		},
		"invalid ip empty range start": {
			ipAddress:   types.StringValue("-12.0.0.1"),
			expectError: true,
		},
		"invalid ip empty range end": {
			ipAddress:   types.StringValue("12.0.0.1-"),
			expectError: true,
		},
		"invalid ip cidr-cidr range 12.0.0.0/24-12.0.0.128/25": {
			ipAddress:   types.StringValue("12.0.0.0/24-12.0.0.128/25"),
			expectError: true,
		},
		"invalid ip ipv6 cidr in range 2001:db8::/64-2001:db8::5": {
			ipAddress:   types.StringValue("2001:db8::/64-2001:db8::5"),
			expectError: true,
		},
		"valid ip 0.0.0.0 mask 0": {
			ipAddress: types.StringValue("0.0.0.0/0"),
		},
		"valid ip :: mask 0": {
			ipAddress: types.StringValue("::/0"),
		},
		"invalid ip 2001:db8:: mask 129": {
			ipAddress:   types.StringValue("2001:db8::/129"),
			expectError: true,
		},
		"invalid ip 12.0.0.1 mask 33": {
			ipAddress:   types.StringValue("12.0.0.1/33"),
			expectError: true,
		},
		"invalid ip equal ipv6 different notation": {
			ipAddress:   types.StringValue("2001:db8::1-2001:0db8:0000:0000:0000:0000:0000:0001"),
			expectError: true,
		},
		"valid ip cross-octet ascending 12.0.0.255-12.0.1.0": {
			ipAddress: types.StringValue("12.0.0.255-12.0.1.0"),
		},
		"valid ip high range 240.0.0.1-250.0.0.1": {
			ipAddress: types.StringValue("240.0.0.1-250.0.0.1"),
		},
		"invalid ip high range reversed 250.0.0.1-240.0.0.1": {
			ipAddress:   types.StringValue("250.0.0.1-240.0.0.1"),
			expectError: true,
		},
		"invalid ip leading space": {
			ipAddress:   types.StringValue(" 12.0.0.1"),
			expectError: true,
		},
		"invalid ip ipv6 with zone fe80::1%eth0": {
			ipAddress:   types.StringValue("fe80::1%eth0"),
			expectError: true,
		},
		"valid ip ipv4-mapped ipv6 ::ffff:12.0.0.1": {
			ipAddress: types.StringValue("::ffff:12.0.0.1"),
		},
	}

	for name, test := range tests {
		t.Run(fmt.Sprintf("ValidateIPFilter - %s", name), func(t *testing.T) {
			t.Parallel()
			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.ipAddress,
			}
			response := validator.StringResponse{}
			IPFilterValidator{}.ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
