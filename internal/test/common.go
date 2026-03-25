// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"fmt"
	"os"
	"testing"
)

var (
	citrix_provider_cloud = `
 provider "citrix" {
	cvad_config = {
		customer_id   = "%s"
		hostname = "%s"
		client_id     = "%s"
		client_secret = "%s"
	}
 }
`
	citrix_provider_on_prem = `
	provider "citrix" {
		cvad_config = {
			hostname = "%s"
			client_id     = "%s"
			client_secret = "%s"
			disable_ssl_verification = true
		}
	}
`
)

// Used to skip a test case if environment is cloud
func skipForCloud(isOnPremises bool) func() (bool, error) {
	return func() (bool, error) {
		if isOnPremises {
			return false, nil
		}

		return true, nil
	}
}

// Used to skip a test case if environment is cloud
func skipForOnPrem(isOnPremises bool) func() (bool, error) {
	return func() (bool, error) {
		if isOnPremises {
			return true, nil
		}

		return false, nil
	}
}

// Used to skip a test case if go test is running in GitHub Actions
func skipForGitHubAction(isGitHubAction bool) func() (bool, error) {
	return func() (bool, error) {
		if isGitHubAction {
			return true, nil
		}

		return false, nil
	}
}

func skipForCVADVersion(isPre2407AndOnPremises bool) func() (bool, error) {
	return func() (bool, error) {
		if isPre2407AndOnPremises {
			return true, nil
		}

		return false, nil
	}
}

func skipForPolicySet(isDDCVersionSupportedForPolicy bool) func() (bool, error) {
	return func() (bool, error) {
		if !isDDCVersionSupportedForPolicy {
			return true, nil
		}

		return false, nil
	}
}

// Used to aggregate arbitrary number of terraform resource blocks
func composeTestResourceTf(resources ...string) string {
	var result = ""
	for _, resource := range resources {
		result += resource
	}
	return result
}

func checkTestEnvironmentVariables(t *testing.T, envVarNames []string) {
	if testing.Short() {
		t.Skip("skipping acceptance test")
	}

	for _, v := range envVarNames {
		if os.Getenv(v) == "" {
			t.Fatalf("%s must be set for acceptance tests", v)
		}
	}
}

func BuildProvider(clientId string, clientSecret string, hostname string, customerId string, isOnPremises bool) string {
	if isOnPremises {
		return fmt.Sprintf(citrix_provider_on_prem, hostname, clientId, clientSecret)
	}
	return fmt.Sprintf(citrix_provider_cloud, customerId, hostname, clientId, clientSecret)
}
