// Copyright © 2023. Citrix Systems, Inc.

package test

// Used to skip a test case if environment is cloud
func getSkipFunc(isOnPremises bool) func() (bool, error) {
	return func() (bool, error) {
		if isOnPremises {
			return false, nil
		}

		return true, nil
	}
}
