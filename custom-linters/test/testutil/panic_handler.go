// Copyright © 2026. Citrix Systems, Inc.

package testutil

import "github.com/hashicorp/terraform-plugin-framework/diag"

// PanicHandler is a stub implementation for testing purposes.
// This function is used in testresource files to verify the custom linter behavior.
func PanicHandler(diag *diag.Diagnostics) {
	// Stub implementation for testing
}
