// Copyright © 2026. Citrix Systems, Inc.

package machine_catalog // whitebox testing in same package

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		configuredImage  string
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
		{
			name:             "snapshots folder image keeps the configured master image",
			configuredImage:  "d-99-xd-gen-mvm",
			imageName:        "snap-2026-08-03-d-99-xd-gen-mvm",
			xdPath:           `XDHyp:\HostingUnits\vpc-net1\snapshots.folder\snap-2026-08-03-d-99-xd-gen-mvm.snapshot`,
			expectedImage:    "d-99-xd-gen-mvm",
			expectedSnapshot: "snap-2026-08-03-d-99-xd-gen-mvm",
		},
		{
			name:             "snapshots folder image on import leaves the master image unset",
			imageName:        "snap-2026-08-03-d-99-xd-gen-mvm",
			xdPath:           `XDHyp:\HostingUnits\vpc-net1\snapshots.folder\snap-2026-08-03-d-99-xd-gen-mvm.snapshot`,
			expectedSnapshot: "snap-2026-08-03-d-99-xd-gen-mvm",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			diagnostics := diag.Diagnostics{}
			machineConfig := GcpMachineConfigModel{}
			if testCase.configuredImage != "" {
				machineConfig.MasterImage = types.StringValue(testCase.configuredImage)
			}

			machineConfig.RefreshProperties(ctx, &diagnostics, newGcpCatalog(testCase.imageName, testCase.xdPath))

			assert.False(t, diagnostics.HasError(), "unexpected diagnostics: %s", diagnostics)
			assert.Equal(t, testCase.expectedImage, machineConfig.MasterImage.ValueString())
			assert.Equal(t, testCase.expectedSnapshot, machineConfig.MachineSnapshot.ValueString())
		})
	}
}

func newGcpMachineConfig(ctx context.Context, masterImage, snapshot, storageType string) types.Object {
	attributesMap, _ := util.ResourceAttributeMapFromObject(GcpMachineConfigModel{})
	machineConfig := util.ObjectValueToTypedObject[GcpMachineConfigModel](ctx, nil, types.ObjectNull(attributesMap))
	if masterImage != "" {
		machineConfig.MasterImage = types.StringValue(masterImage)
	}
	if snapshot != "" {
		machineConfig.MachineSnapshot = types.StringValue(snapshot)
	}
	machineConfig.StorageType = types.StringValue(storageType)
	return util.TypedObjectToObjectValue(ctx, &diag.Diagnostics{}, machineConfig)
}

func TestGcpMachineConfigRequiresReplace(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name                    string
		stateImage              string
		stateSnapshot           string
		planImage               string
		planStorageType         string
		expectedRequiresReplace bool
	}{
		{
			name:            "imported snapshots folder catalog adopts the configured master image",
			stateSnapshot:   "snap-1",
			planImage:       "vm-1",
			planStorageType: "pd-ssd",
		},
		{
			name:            "catalog imported by an earlier version with the folder as master image adopts the configured master image",
			stateImage:      "snapshots.folder",
			stateSnapshot:   "snap-1",
			planImage:       "vm-1",
			planStorageType: "pd-ssd",
		},
		{
			name:                    "imported snapshots folder catalog with another change is replaced",
			stateSnapshot:           "snap-1",
			planImage:               "vm-1",
			planStorageType:         "pd-standard",
			expectedRequiresReplace: true,
		},
		{
			name:                    "changing a known master image is replaced",
			stateImage:              "vm-1",
			stateSnapshot:           "snap-1",
			planImage:               "vm-2",
			planStorageType:         "pd-ssd",
			expectedRequiresReplace: true,
		},
		{
			name:                    "master image without a snapshot is replaced",
			planImage:               "vm-1",
			planStorageType:         "pd-ssd",
			expectedRequiresReplace: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			req := planmodifier.ObjectRequest{
				StateValue: newGcpMachineConfig(ctx, testCase.stateImage, testCase.stateSnapshot, "pd-ssd"),
				PlanValue:  newGcpMachineConfig(ctx, testCase.planImage, testCase.stateSnapshot, testCase.planStorageType),
			}
			resp := &objectplanmodifier.RequiresReplaceIfFuncResponse{}

			gcpMachineConfigRequiresReplace(ctx, req, resp)

			assert.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %s", resp.Diagnostics)
			assert.Equal(t, testCase.expectedRequiresReplace, resp.RequiresReplace)
		})
	}
}

func mockResourcePoolLookup(mockHypervisorsAPI *citrixorchestration.MockHypervisorsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS, folderPath string, job *citrixorchestration.JobResponseModel, results string) {
	mockHypervisorsAPI.On("HypervisorsGetHypervisorResourcePoolResourcesExecute", mock.MatchedBy(func(r citrixorchestration.ApiHypervisorsGetHypervisorResourcePoolResourcesRequest) bool {
		path := citrixorchestration.MockApiHypervisorsGetHypervisorResourcePoolResourcesRequest{ApiHypervisorsGetHypervisorResourcePoolResourcesRequest: r}.GetPath()
		return path != nil && *path == folderPath
	})).Return((*citrixorchestration.HypervisorResourceResponseModel)(nil), testutil.MockAsyncResponse(job.GetId(), "txn-"+job.GetId()), nil).Once()
	mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(job, testutil.MockSuccessResponse(), nil).Once()
	if job.GetStatus() == citrixorchestration.JOBSTATUS_COMPLETE {
		mockJobsAPI.On("JobsGetJobResultsExecute", mock.Anything).Return(results, testutil.MockSuccessResponse(), nil).Once()
	}
}

func TestGetGcpSnapshotPath(t *testing.T) {
	const vmFolderPath = `XDHyp:\HostingUnits\vpc-net1\vm-1.vm\snap-1.snapshot`
	const snapshotsFolderPath = `XDHyp:\HostingUnits\vpc-net1\snapshots.folder\snap-1.snapshot`
	snapshotResult := func(xdPath string) string {
		result, err := json.Marshal(map[string][]map[string]string{"Children": {{"Name": "snap-1", "XDPath": xdPath}}})
		require.NoError(t, err)
		return string(result)
	}
	missingVmFolder := testutil.MockFailedJob("job-vm", `Cannot find path 'XDHyp:\HostingUnits\vpc-net1\vm-1.vm' because it does not exist.`)

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name          string
		imageVm       string
		setupMocks    func(*citrixorchestration.MockHypervisorsAPIsDAAS, *citrixorchestration.MockJobsAPIsDAAS)
		expectedPath  string
		errorContains string
	}{
		{
			name: "snapshot without a source VM only searches the snapshots folder",
			setupMocks: func(mockHypervisorsAPI *citrixorchestration.MockHypervisorsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, gcpSnapshotsFolder, testutil.MockCompletedJob("job-folder", citrixorchestration.JOBTYPE_UNKNOWN), snapshotResult(snapshotsFolderPath))
			},
			expectedPath: snapshotsFolderPath,
		},
		{
			name:    "snapshot under the source VM",
			imageVm: "vm-1",
			setupMocks: func(mockHypervisorsAPI *citrixorchestration.MockHypervisorsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, "vm-1.vm", testutil.MockCompletedJob("job-vm", citrixorchestration.JOBTYPE_UNKNOWN), snapshotResult(vmFolderPath))
			},
			expectedPath: vmFolderPath,
		},
		{
			name:    "snapshot under the snapshots folder",
			imageVm: "vm-1",
			setupMocks: func(mockHypervisorsAPI *citrixorchestration.MockHypervisorsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, "vm-1.vm", missingVmFolder, "")
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, "vm-1.vm", missingVmFolder, "")
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, gcpSnapshotsFolder, testutil.MockCompletedJob("job-folder", citrixorchestration.JOBTYPE_UNKNOWN), snapshotResult(snapshotsFolderPath))
			},
			expectedPath: snapshotsFolderPath,
		},
		{
			name:    "snapshot in neither location",
			imageVm: "vm-1",
			setupMocks: func(mockHypervisorsAPI *citrixorchestration.MockHypervisorsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, "vm-1.vm", missingVmFolder, "")
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, "vm-1.vm", missingVmFolder, "")
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, gcpSnapshotsFolder, testutil.MockCompletedJob("job-folder", citrixorchestration.JOBTYPE_UNKNOWN), `{"Children":[]}`)
				mockResourcePoolLookup(mockHypervisorsAPI, mockJobsAPI, gcpSnapshotsFolder, testutil.MockCompletedJob("job-folder", citrixorchestration.JOBTYPE_UNKNOWN), `{"Children":[]}`)
			},
			errorContains: "could not find resource",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)
			testCase.setupMocks(citrixorchestration.GetMockHypervisorsAPIsDAAS(mockClient.APIClient), citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient))
			diagnostics := diag.Diagnostics{}

			imagePath, _, err := getGcpSnapshotPath(test.DaaSTestContext(), client, &diagnostics, "gcp-hyp", "vpc-net1", testCase.imageVm, "snap-1")

			if testCase.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.errorContains)
				return
			}
			require.NoError(t, err)
			assert.False(t, diagnostics.HasError(), "unexpected diagnostics: %s", diagnostics)
			assert.Equal(t, testCase.expectedPath, imagePath)
		})
	}
}

func TestImageUpdateRebootOptionsValidateConfig(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name                  string
		rebootDuration        types.Int64
		warningDuration       types.Int64
		warningRepeatInterval types.Int64
		expectedError         string
	}{
		{
			name:                  "unknown reboot_duration skips the reboot_duration checks",
			rebootDuration:        types.Int64Unknown(),
			warningDuration:       types.Int64Value(-1),
			warningRepeatInterval: types.Int64Null(),
		},
		{
			name:                  "unknown warning_duration skips the repeat interval check",
			rebootDuration:        types.Int64Value(60),
			warningDuration:       types.Int64Unknown(),
			warningRepeatInterval: types.Int64Value(5),
		},
		{
			name:                  "unknown warning_repeat_interval skips the repeat interval check",
			rebootDuration:        types.Int64Value(-1),
			warningDuration:       types.Int64Value(-1),
			warningRepeatInterval: types.Int64Unknown(),
		},
		{
			name:                  "unknown warning_repeat_interval still checks a known warning_duration",
			rebootDuration:        types.Int64Value(60),
			warningDuration:       types.Int64Value(-1),
			warningRepeatInterval: types.Int64Unknown(),
			expectedError:         "warning_duration cannot be set to -1 when reboot_duration is not set to -1.",
		},
		{
			name:                  "unknown reboot_duration still checks a known warning_repeat_interval",
			rebootDuration:        types.Int64Unknown(),
			warningDuration:       types.Int64Value(10),
			warningRepeatInterval: types.Int64Value(10),
			expectedError:         "warning_repeat_interval must be shorter than warning_duration.",
		},
		{
			name:                  "warning_duration -1 without reboot_duration -1 is rejected",
			rebootDuration:        types.Int64Value(60),
			warningDuration:       types.Int64Value(-1),
			warningRepeatInterval: types.Int64Null(),
			expectedError:         "warning_duration cannot be set to -1 when reboot_duration is not set to -1.",
		},
		{
			name:                  "positive warning_duration with reboot_duration -1 is rejected",
			rebootDuration:        types.Int64Value(-1),
			warningDuration:       types.Int64Value(30),
			warningRepeatInterval: types.Int64Null(),
			expectedError:         "warning_duration can only be set to -1 or 0 when reboot_duration is set to -1.",
		},
		{
			name:                  "warning_repeat_interval not shorter than warning_duration is rejected",
			rebootDuration:        types.Int64Value(60),
			warningDuration:       types.Int64Value(10),
			warningRepeatInterval: types.Int64Value(10),
			expectedError:         "warning_repeat_interval must be shorter than warning_duration.",
		},
		{
			name:                  "known valid values pass",
			rebootDuration:        types.Int64Value(60),
			warningDuration:       types.Int64Value(30),
			warningRepeatInterval: types.Int64Value(10),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := diag.Diagnostics{}
			rebootOptions := ImageUpdateRebootOptionsModel{
				RebootDuration:        testCase.rebootDuration,
				WarningDuration:       testCase.warningDuration,
				WarningMessage:        types.StringValue("Rebooting in %m% minutes."),
				WarningRepeatInterval: testCase.warningRepeatInterval,
			}

			rebootOptions.ValidateConfig(&diagnostics)

			if testCase.expectedError == "" {
				assert.False(t, diagnostics.HasError(), "unexpected diagnostics: %s", diagnostics)
				return
			}
			require.Equal(t, 1, diagnostics.ErrorsCount(), "diagnostics: %s", diagnostics)
			assert.Equal(t, testCase.expectedError, diagnostics.Errors()[0].Detail())
		})
	}
}
