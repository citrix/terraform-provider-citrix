// Copyright © 2026. Citrix Systems, Inc.

package panichandler_test

import (
	"testing"

	"github.com/citrix/terraform-provider-citrix/custom-linters/panichandler"
	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
)

func TestAllCases(t *testing.T) {
	testCases := testutil.ExtractTestCasesFromComments(t, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "panichandler_test_cases.go")
	testutil.RunAnalyzerTestCases(t, panichandler.Analyzer, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "panichandler_test_cases.go", testCases)
}
