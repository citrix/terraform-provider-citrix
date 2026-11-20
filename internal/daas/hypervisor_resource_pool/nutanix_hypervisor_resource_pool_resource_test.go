// Copyright © 2026. Citrix Systems, Inc.

package hypervisor_resource_pool

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestIsNutanixPrismElement(t *testing.T) {
	tests := []struct {
		name           string
		connectionType citrixorchestration.HypervisorConnectionType
		pluginId       string
		expected       bool
	}{
		{"Prism Element", citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM, util.NUTANIX_PLUGIN_ID, true},
		{"other Custom plugin", citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM, "SomeOtherFactory", false},
		{"Prism Central (Citrix plugin)", citrixorchestration.HYPERVISORCONNECTIONTYPE_NUTANIX_AHV_PC, "AcropolisHypervisorPCFactory", false},
		{"Prism Central (Nutanix plugin)", citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM, "AcropolisPCFactory", false},
		{"Nutanix Xi", citrixorchestration.HYPERVISORCONNECTIONTYPE_CUSTOM, "AcropolisXIFactory", false},
		{"vCenter", citrixorchestration.HYPERVISORCONNECTIONTYPE_V_CENTER, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isNutanixPrismElement(tt.connectionType, tt.pluginId))
		})
	}
}
