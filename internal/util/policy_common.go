// Copyright © 2024. Citrix Systems, Inc.

package util

type PolicyFilterUuidDataClientModel struct {
	Server string `json:"server,omitempty"`
	Uuid   string `json:"uuid,omitempty"`
}

type PolicyFilterGatewayDataClientModel struct {
	Connection string `json:"Connection,omitempty"`
	Condition  string `json:"Condition,omitempty"`
	Gateway    string `json:"Gateway,omitempty"`
}
