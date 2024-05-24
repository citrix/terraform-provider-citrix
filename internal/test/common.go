// Copyright © 2023. Citrix Systems, Inc.

package test

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

// Used to aggregate arbitrary number of terraform resource blocks
func composeTestResourceTf(resources ...string) string {
	var result = ""
	for _, resource := range resources {
		result += resource
	}
	return result
}
