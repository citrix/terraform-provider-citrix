// Copyright © 2024. Citrix Systems, Inc.

package util

import "errors"

// Custom error types for policy-related "not found" scenarios
// These allow for type-safe error checking instead of string comparisons

var (
	// ErrPolicyNotFound is returned when a policy is not found
	ErrPolicyNotFound = errors.New("policy not found")
	
	// ErrPolicyFilterNotFound is returned when a policy filter is not found
	ErrPolicyFilterNotFound = errors.New("policy filter not found")
	
	// ErrPolicySettingNotFound is returned when a policy setting is not found
	ErrPolicySettingNotFound = errors.New("policy setting not found")
	
	// ErrPolicySetNotFound is returned when a policy set is not found
	ErrPolicySetNotFound = errors.New("policy set not found")
)

type PolicyFilterUuidDataClientModel struct {
	Server string `json:"server,omitempty"`
	Uuid   string `json:"uuid,omitempty"`
}

type PolicyFilterGatewayDataClientModel struct {
	Connection string `json:"Connection,omitempty"`
	Condition  string `json:"Condition,omitempty"`
	Gateway    string `json:"Gateway,omitempty"`
}

const (
	POLICYSETTING_GO_VALUETYPE_STATE        = "State"
	POLICYSETTING_GO_VALUETYPE_STATEALLOWED = "StateAllowed"
)
