// Copyright © 2026. Citrix Systems, Inc.

package executewithretry

import (
	"testing"

	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
)

func TestAllCases(t *testing.T) {
	testCases := testutil.ExtractTestCasesFromComments(t, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "executewithretry_test_cases.go")
	testutil.RunAnalyzerTestCases(t, Analyzer, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "executewithretry_test_cases.go", testCases)
}
