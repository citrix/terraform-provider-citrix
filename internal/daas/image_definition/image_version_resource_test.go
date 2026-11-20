// Copyright © 2026. Citrix Systems, Inc.

package image_definition // whitebox testing in same package

import (
	"context"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
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
		NutanixPCImageSpecs:            nullObject[NutanixPCImageSpecsModel](t),
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

func newNutanixPCImageSpecs(t *testing.T, masterImageVersion string, memoryMB int32) types.Object {
	t.Helper()
	var diagnostics diag.Diagnostics
	specs := util.TypedObjectToObjectValue(context.Background(), &diagnostics, NutanixPCImageSpecsModel{
		Cluster:               types.StringValue("cluster-1"),
		MasterImage:           types.StringValue("win11"),
		MasterImageVersion:    types.StringValue(masterImageVersion),
		MachineProfile:        types.StringValue("win11"),
		MachineProfileVersion: types.StringValue("v1"),
		CpuCount:              types.Int32Value(2),
		CoresPerCpuCount:      types.Int32Value(2),
		MemoryMB:              types.Int32Value(memoryMB),
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
			name: "Nutanix Prism Central master image version changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v1", 8192)
				plan.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v2", 8192)
				return state, plan
			},
			expected: true,
		},
		{
			name: "Nutanix Prism Central memory changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v1", 8192)
				plan.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v1", 16384)
				return state, plan
			},
			expected: true,
		},
		{
			name: "Nutanix Prism Central unchanged, description changed",
			modify: func(state ImageVersionModel, plan ImageVersionModel) (ImageVersionModel, ImageVersionModel) {
				state.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v1", 8192)
				plan.NutanixPCImageSpecs = newNutanixPCImageSpecs(t, "v1", 8192)
				plan.Description = types.StringValue("new description")
				return state, plan
			},
			expected: false,
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

func TestNutanixPCImageSpecsRefreshFromImageScheme(t *testing.T) {
	masterImage := citrixorchestration.HypervisorResourceRefResponseModel{}
	masterImage.SetXDPath(`XDHyp:\HostingUnits\pc-pool\Templates.folder\win11.template\v3.templateversion`)

	machineProfile := citrixorchestration.HypervisorResourceRefResponseModel{}
	machineProfile.SetXDPath(`XDHyp:\HostingUnits\pc-pool\Templates.folder\win11-profile.template\p1.templateversion`)

	clusterId := citrixorchestration.NameValueStringPairModel{}
	clusterId.SetName(util.NutanixPCClusterIdKey)
	clusterId.SetValue("0005-uuid")
	cpuCores := citrixorchestration.NameValueStringPairModel{}
	cpuCores.SetName(util.NutanixPCCpuCoresKey)
	cpuCores.SetValue("4")

	imageScheme := citrixorchestration.ImageSchemeResponseModel{}
	imageScheme.SetMachineProfile(machineProfile)
	imageScheme.SetCpuCount(2)
	imageScheme.SetMemoryMB(8192)
	imageScheme.SetCustomProperties([]citrixorchestration.NameValueStringPairModel{clusterId, cpuCores})

	specs := NutanixPCImageSpecsModel{}
	specs.refreshFromImageScheme(masterImage, imageScheme)

	assert.Equal(t, "win11", specs.MasterImage.ValueString())
	assert.Equal(t, "v3", specs.MasterImageVersion.ValueString())
	assert.Equal(t, "win11-profile", specs.MachineProfile.ValueString())
	assert.Equal(t, "p1", specs.MachineProfileVersion.ValueString())
	assert.Equal(t, int32(2), specs.CpuCount.ValueInt32())
	assert.Equal(t, int32(4), specs.CoresPerCpuCount.ValueInt32())
	assert.Equal(t, int32(8192), specs.MemoryMB.ValueInt32())
}

func TestNutanixPCImageVersionRefreshKeepsExistingCluster(t *testing.T) {
	client, mockClient := test.NewTestDaaSClient()
	defer mockClient.AssertExpectations(t)
	ctx := test.DaaSTestContext()

	masterImage := citrixorchestration.HypervisorResourceRefResponseModel{}
	masterImage.SetXDPath(`XDHyp:\HostingUnits\pc-pool\Templates.folder\win11.template\v3.templateversion`)

	clusterId := citrixorchestration.NameValueStringPairModel{}
	clusterId.SetName(util.NutanixPCClusterIdKey)
	clusterId.SetValue("0005-uuid")
	imageScheme := citrixorchestration.ImageSchemeResponseModel{}
	imageScheme.SetCustomProperties([]citrixorchestration.NameValueStringPairModel{clusterId})

	imageContext := citrixorchestration.ImageVersionSpecContextResponseModel{}
	imageContext.SetPluginFactoryName(util.NUTANIX_PC_PLUGIN_ID)
	imageContext.SetImageScheme(imageScheme)

	spec := citrixorchestration.ImageVersionSpecResponseModel{}
	spec.SetContext(imageContext)
	spec.SetMasterImage(masterImage)

	imageVersion := citrixorchestration.ImageVersionResponseModel{}
	imageVersion.SetImageVersionSpecs([]citrixorchestration.ImageVersionSpecResponseModel{spec})

	var diagnostics diag.Diagnostics
	state := newImageVersionModel(t)
	state.NutanixPCImageSpecs = util.TypedObjectToObjectValue(ctx, &diagnostics, NutanixPCImageSpecsModel{Cluster: types.StringValue("cluster-1")})

	// The mock client has no expectations, so resolving the cluster ID through the API again fails the test.
	refreshed := state.RefreshPropertyValues(ctx, &diagnostics, client, &imageVersion)

	require.False(t, diagnostics.HasError())
	specs := util.ObjectValueToTypedObject[NutanixPCImageSpecsModel](ctx, &diagnostics, refreshed.NutanixPCImageSpecs)
	assert.Equal(t, "cluster-1", specs.Cluster.ValueString())
	assert.Equal(t, "win11", specs.MasterImage.ValueString())
}
