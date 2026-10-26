// Copyright © 2026. Citrix Systems, Inc.

package machine_catalog // whitebox testing in same package

import (
	"context"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
)

// newGcpCatalog builds a GCP MCS catalog whose provisioning scheme carries the given master image.
// name is what Orchestration reports for the image: the snapshot name when a snapshot was used,
// otherwise the VM name.
func newGcpCatalog(name, xdPath string) citrixorchestration.MachineCatalogDetailResponseModel {
	masterImage := citrixorchestration.HypervisorResourceRefResponseModel{}
	masterImage.SetName(name)
	masterImage.SetXDPath(xdPath)

	provScheme := citrixorchestration.ProvisioningSchemeResponseModel{}
	provScheme.SetMasterImage(masterImage)

	catalog := citrixorchestration.MachineCatalogDetailResponseModel{}
	catalog.SetProvisioningScheme(provScheme)
	return catalog
}

func TestGcpMachineConfigRefreshMasterImage(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name             string
		imageName        string
		xdPath           string
		expectedImage    string
		expectedSnapshot string
	}{
		{
			name:             "snapshot image strips the .vm suffix from the VM segment",
			imageName:        "snap-2026-08-03-d-99-xd-gen-mvm",
			xdPath:           `XDHyp:\HostingUnits\vpc-net1\d-99-xd-gen-mvm.vm\snap-2026-08-03-d-99-xd-gen-mvm.snapshot`,
			expectedImage:    "d-99-xd-gen-mvm",
			expectedSnapshot: "snap-2026-08-03-d-99-xd-gen-mvm",
		},
		{
			name:          "VM image uses the reported image name",
			imageName:     "d-99-xd-gen-mvm",
			xdPath:        `XDHyp:\HostingUnits\vpc-net1\d-99-xd-gen-mvm.vm`,
			expectedImage: "d-99-xd-gen-mvm",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			diagnostics := diag.Diagnostics{}
			machineConfig := GcpMachineConfigModel{}

			machineConfig.RefreshProperties(ctx, &diagnostics, newGcpCatalog(testCase.imageName, testCase.xdPath))

			assert.False(t, diagnostics.HasError(), "unexpected diagnostics: %s", diagnostics)
			assert.Equal(t, testCase.expectedImage, machineConfig.MasterImage.ValueString())
			assert.Equal(t, testCase.expectedSnapshot, machineConfig.MachineSnapshot.ValueString())
		})
	}
}
