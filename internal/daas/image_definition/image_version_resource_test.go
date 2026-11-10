// Copyright © 2026. Citrix Systems, Inc.

package image_definition // whitebox testing in same package

import (
	"context"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

func nullObject[T util.ResourceModelWithAttributes](t *testing.T) types.Object {
	t.Helper()
	var model T
	attributeMap, err := util.ResourceAttributeMapFromObject(model)
	require.NoError(t, err)
	return types.ObjectNull(attributeMap)
}

// newImageVersionModel creates an ImageVersionModel with all image specs and network mapping unset.
func newImageVersionModel(t *testing.T) ImageVersionModel {
	t.Helper()
	var diagnostics diag.Diagnostics
	model := ImageVersionModel{
		AzureImageSpecs:                nullObject[util.AzureImageSpecsModel](t),
		AwsEc2ImageSpecs:               nullObject[AwsEc2ImageSpecsModel](t),
		VsphereImageSpecs:              nullObject[VsphereImageSpecsModel](t),
		AmazonWorkspacesCoreImageSpecs: nullObject[AmazonWorkspacesCoreImageSpecsModel](t),
		NetworkMapping:                 util.TypedArrayToObjectList[util.NetworkMappingModel](context.Background(), &diagnostics, nil),
		Description:                    types.StringValue(""),
	}
	require.False(t, diagnostics.HasError())
	return model
}

func newVsphereImageSpecs(t *testing.T, masterImageVm string, imageSnapshot string, cpuCount int32, diskSize types.Int32) types.Object {
	t.Helper()
	var diagnostics diag.Diagnostics
	specs := util.TypedObjectToObjectValue(context.Background(), &diagnostics, VsphereImageSpecsModel{
		MasterImageVm:  types.StringValue(masterImageVm),
		ImageSnapshot:  types.StringValue(imageSnapshot),
		CpuCount:       types.Int32Value(cpuCount),
		MemoryMB:       types.Int32Value(4096),
		MachineProfile: types.StringNull(),
		DiskSize:       diskSize,
	})
	require.False(t, diagnostics.HasError())
	return specs
}

func newAmazonWorkspacesCoreImageSpecs(t *testing.T, imageAmi string) types.Object {
	t.Helper()
	var diagnostics diag.Diagnostics
	specs := util.TypedObjectToObjectValue(context.Background(), &diagnostics, AmazonWorkspacesCoreImageSpecsModel{
		ServiceOffering: types.StringValue("w2.large"),
		MasterImage:     types.StringValue("golden-image"),
		ImageAmi:        types.StringValue(imageAmi),
		MachineProfile:  nullObject[util.AmazonWorkspacesCoreMachineProfileModel](t),
	})
	require.False(t, diagnostics.HasError())
	return specs
}

func newNetworkMapping(t *testing.T, network string) types.List {
	t.Helper()
	var diagnostics diag.Diagnostics
	networkMapping := util.TypedArrayToObjectList(context.Background(), &diagnostics, []util.NetworkMappingModel{
		{NetworkDevice: types.StringValue("0"), Network: types.StringValue(network)},
	})
	require.False(t, diagnostics.HasError())
	return networkMapping
}

// ========== Tests ==========

func TestImageVersionSpecsChanged(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name     string
		modify   func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel)
		expected bool
	}{
		{
			name: "Only description changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.AmazonWorkspacesCoreImageSpecs = newAmazonWorkspacesCoreImageSpecs(t, "ami-0123456789abcdef0")
				plan.AmazonWorkspacesCoreImageSpecs = newAmazonWorkspacesCoreImageSpecs(t, "ami-0123456789abcdef0")
				plan.Description = types.StringValue("new description")
				return state, plan
			},
			expected: false,
		},
		{
			name: "Amazon WorkSpaces Core AMI changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.AmazonWorkspacesCoreImageSpecs = newAmazonWorkspacesCoreImageSpecs(t, "ami-0123456789abcdef0")
				plan.AmazonWorkspacesCoreImageSpecs = newAmazonWorkspacesCoreImageSpecs(t, "ami-0fedcba9876543210")
				return state, plan
			},
			expected: true,
		},
		{
			name: "vSphere cpu count changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.VsphereImageSpecs = newVsphereImageSpecs(t, "golden-vm", "snapshot-1", 2, types.Int32Value(64))
				plan.VsphereImageSpecs = newVsphereImageSpecs(t, "golden-vm", "snapshot-1", 4, types.Int32Value(64))
				return state, plan
			},
			expected: true,
		},
		{
			name: "Network mapping changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.NetworkMapping = newNetworkMapping(t, "network-1")
				plan.NetworkMapping = newNetworkMapping(t, "network-2")
				return state, plan
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, plan := tt.modify(newImageVersionModel(t), newImageVersionModel(t))

			assert.Equal(t, tt.expected, imageVersionSpecsChanged(plan, state))
		})
	}
}

func TestVsphereMasterImageChanged(t *testing.T) {
	stateSpecs := newVsphereImageSpecs(t, "golden-vm", "snapshot-1", 2, types.Int32Value(64))

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name       string
		planSpecs  types.Object
		stateSpecs types.Object
		expected   bool
	}{
		{
			name:       "Snapshot changed",
			planSpecs:  newVsphereImageSpecs(t, "golden-vm", "snapshot-1/snapshot-2", 2, types.Int32Value(64)),
			stateSpecs: stateSpecs,
			expected:   true,
		},
		{
			name:       "Master image VM changed",
			planSpecs:  newVsphereImageSpecs(t, "golden-vm-2", "snapshot-1", 2, types.Int32Value(64)),
			stateSpecs: stateSpecs,
			expected:   true,
		},
		{
			name:       "Only cpu count changed",
			planSpecs:  newVsphereImageSpecs(t, "golden-vm", "snapshot-1", 4, types.Int32Value(64)),
			stateSpecs: stateSpecs,
			expected:   false,
		},
		{
			name:       "vSphere specs not in state",
			planSpecs:  newVsphereImageSpecs(t, "golden-vm", "snapshot-1", 2, types.Int32Unknown()),
			stateSpecs: nullObject[VsphereImageSpecsModel](t),
			expected:   false,
		},
		{
			name:       "vSphere specs unknown in plan",
			planSpecs:  types.ObjectUnknown(stateSpecs.AttributeTypes(context.Background())),
			stateSpecs: stateSpecs,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := newImageVersionModel(t)
			plan.VsphereImageSpecs = tt.planSpecs
			state := newImageVersionModel(t)
			state.VsphereImageSpecs = tt.stateSpecs
			var diagnostics diag.Diagnostics

			changed := vsphereMasterImageChanged(context.Background(), &diagnostics, plan, state)

			assert.False(t, diagnostics.HasError())
			assert.Equal(t, tt.expected, changed)
		})
	}
}
