// Copyright © 2026. Citrix Systems, Inc.

package unknowncheck_test

import (
	"testing"

	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
	"github.com/citrix/terraform-provider-citrix/custom-linters/unknowncheck"
)

func TestAllCases(t *testing.T) {
	testCases := testutil.ExtractTestCasesFromComments(t, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "unknowncheck_test_cases.go")
	testutil.RunAnalyzerTestCases(t, unknowncheck.Analyzer, "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource", "unknowncheck_test_cases.go", testCases)
}
