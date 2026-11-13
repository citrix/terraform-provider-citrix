// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/stretchr/testify/assert"
)

func newCustomProperties(props map[string]string) []citrixorchestration.NameValueStringPairModel {
	result := []citrixorchestration.NameValueStringPairModel{}
	for name, value := range props {
		property := citrixorchestration.NameValueStringPairModel{}
		property.SetName(name)
		property.SetValue(value)
		result = append(result, property)
	}
	return result
}

func TestParseNutanixPCTemplateVersionPath(t *testing.T) {
	tests := []struct {
		name             string
		xdPath           string
		expectedTemplate string
		expectedVersion  string
		expectedOk       bool
	}{
		{"template version", `XDHyp:\HostingUnits\pc-pool\Templates.folder\win11-static.template\v3.templateversion`, "win11-static", "v3", true},
		{"version name with dots", `XDHyp:\HostingUnits\pc-pool\Templates.folder\win11.static.template\2026.11.templateversion`, "win11.static", "2026.11", true},
		{"plain template", `XDHyp:\HostingUnits\pc-pool\Templates.folder\win11-static.template`, "", "", false},
		{"vm", `XDHyp:\HostingUnits\pc-pool\VirtualMachines.folder\gold.vm`, "", "", false},
		{"empty", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, version, ok := ParseNutanixPCTemplateVersionPath(tt.xdPath)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedTemplate, template)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestGetNutanixPCCoresPerCpu(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]string
		expected int64
		ok       bool
	}{
		{"set", map[string]string{"ClusterId": "abc", "CPUCores": "2"}, 2, true},
		{"missing", map[string]string{"ClusterId": "abc"}, 0, false},
		{"not a number", map[string]string{"CPUCores": "two"}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cores, ok := GetNutanixPCCoresPerCpu(newCustomProperties(tt.props))
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, cores)
		})
	}
}

func TestGetNutanixPCClusterId(t *testing.T) {
	assert.Equal(t, "0005-uuid", GetNutanixPCClusterId(newCustomProperties(map[string]string{"ClusterId": "0005-uuid", "CPUCores": "2"})))
	assert.Empty(t, GetNutanixPCClusterId(newCustomProperties(map[string]string{"CPUCores": "2"})))
}

func TestNutanixPCCustomProperties(t *testing.T) {
	props := NutanixPCCustomProperties("0005-uuid", 4)
	assert.Equal(t, "0005-uuid", GetNutanixPCClusterId(props))
	cores, ok := GetNutanixPCCoresPerCpu(props)
	assert.True(t, ok)
	assert.Equal(t, int64(4), cores)
	assert.Len(t, props, 2)
}
