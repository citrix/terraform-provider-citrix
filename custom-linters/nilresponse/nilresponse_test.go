// Copyright © 2026. Citrix Systems, Inc.

package nilresponse

import (
	"testing"

	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
)

const fixturePackage = "github.com/citrix/terraform-provider-citrix/custom-linters/test/testresource"

func TestAllCases(t *testing.T) {
	// The fixture mocks the generated clients and retry wrappers, so trust its package like citrix-daas-rest-go.
	saved := trustedPackages
	trustedPackages = append(append([]string{}, saved...), fixturePackage)
	t.Cleanup(func() { trustedPackages = saved })

	testCases := testutil.ExtractTestCasesFromComments(t, fixturePackage, "nilresponse_test_cases.go")
	testutil.RunAnalyzerTestCases(t, Analyzer, fixturePackage, "nilresponse_test_cases.go", testCases)
}
