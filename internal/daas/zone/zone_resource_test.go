// Copyright © 2025. Citrix Systems, Inc.

package zone // whitebox testing in same package

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// newZone creates a ZoneDetailResponseModel with basic fields set.
func newZone(id, name, description string, metadata []citrixorchestration.NameValueStringPairModel) *citrixorchestration.ZoneDetailResponseModel {
	zone := &citrixorchestration.ZoneDetailResponseModel{}
	zone.SetId(id)
	zone.SetName(name)
	zone.SetDescription(description)
	if metadata != nil {
		zone.SetMetadata(metadata)
	} else {
		zone.SetMetadata([]citrixorchestration.NameValueStringPairModel{})
	}
	return zone
}

// ========== Tests ==========

func TestUpdateZoneAfterCreate(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name             string
		metadata         []util.NameValueStringPairModel
		expectedMetadata []citrixorchestration.NameValueStringPairModel
		setupMocks       func(*citrixorchestration.MockZonesAPIsDAAS, *citrixorchestration.MockJobsAPIsDAAS)
		errorContains    string
	}{
		{
			name: "With metadata",
			metadata: []util.NameValueStringPairModel{
				{Name: types.StringValue("key1"), Value: types.StringValue("value1")},
				{Name: types.StringValue("key2"), Value: types.StringValue("value2")},
			},
			expectedMetadata: []citrixorchestration.NameValueStringPairModel{
				testutil.NewMetadataPair("key1", "value1"),
				testutil.NewMetadataPair("key2", "value2"),
			},
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				jobId := "job-123"
				mockZonesAPI.On("ZonesEditZoneExecute", mock.Anything).Return(testutil.MockAsyncResponse(jobId, "txn-123"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob(jobId, citrixorchestration.JOBTYPE_EDIT_ZONE), testutil.MockSuccessResponse(), nil).Once()
				mockZonesAPI.On("ZonesGetZoneExecute", mock.Anything).Return(
					newZone("zone-123", "Test Zone", "Updated description", []citrixorchestration.NameValueStringPairModel{
						testutil.NewMetadataPair("key1", "value1"),
						testutil.NewMetadataPair("key2", "value2"),
					}),
					testutil.MockSuccessResponse(), nil).Once()
			},
		},
		{
			name:             "Without metadata",
			metadata:         nil,
			expectedMetadata: nil,
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				jobId := "job-123"
				mockZonesAPI.On("ZonesEditZoneExecute", mock.Anything).Return(testutil.MockAsyncResponse(jobId, "txn-123"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob(jobId, citrixorchestration.JOBTYPE_EDIT_ZONE), testutil.MockSuccessResponse(), nil).Once()
				mockZonesAPI.On("ZonesGetZoneExecute", mock.Anything).Return(
					newZone("zone-123", "Test Zone", "Updated description", nil),
					testutil.MockSuccessResponse(), nil).Once()
			},
		},
		{
			name: "Edit zone API fails",
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockZonesAPI.On("ZonesEditZoneExecute", mock.Anything).Return(testutil.MockErrorResponse(500, "txn-error"), assert.AnError).Once()
			},
			errorContains: "Error updating Zone",
		},
		{
			name: "Job fails during async operation",
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				jobId := "job-fail"
				mockZonesAPI.On("ZonesEditZoneExecute", mock.Anything).Return(testutil.MockAsyncResponse(jobId, "txn-job-fail"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(
					testutil.MockFailedJob(jobId, "Zone update failed due to internal error"),
					testutil.MockSuccessResponse(),
					nil,
				).Once()
			},
			errorContains: "Error updating zone",
		},
		{
			name: "Get zone fails after successful update",
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				jobId := "job-success"
				mockZonesAPI.On("ZonesEditZoneExecute", mock.Anything).Return(testutil.MockAsyncResponse(jobId, "txn-success"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob(jobId, citrixorchestration.JOBTYPE_EDIT_ZONE), testutil.MockSuccessResponse(), nil).Once()
				mockZonesAPI.On("ZonesGetZoneExecute", mock.Anything).Return(
					(*citrixorchestration.ZoneDetailResponseModel)(nil),
					testutil.MockErrorResponse(404, "txn-get-error"),
					assert.AnError,
				).Once()
			},
			errorContains: "Error reading Zone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)

			mockZonesAPI := citrixorchestration.GetMockZonesAPIsDAAS(mockClient.APIClient)
			mockJobsAPI := citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient)

			ctx := test.DaaSTestContext()
			zoneId := "zone-123"
			zoneName := "Test Zone"
			description := "Updated description"

			zoneModel := ZoneResourceModel{
				Id:          types.StringValue(zoneId),
				Name:        types.StringValue(zoneName),
				Description: types.StringValue(description),
			}

			if tt.metadata != nil {
				zoneModel.Metadata = util.TypedArrayToObjectList(ctx, nil, tt.metadata)
			} else {
				attributeMap, err := util.ResourceAttributeMapFromObject(util.NameValueStringPairModel{})
				require.NoError(t, err)
				zoneModel.Metadata = types.ListNull(types.ObjectType{AttrTypes: attributeMap})
			}

			initialZone := newZone(zoneId, zoneName, "", nil)

			tt.setupMocks(mockZonesAPI, mockJobsAPI)

			var diagnostics diag.Diagnostics
			result, err := zoneModel.updateZoneAfterCreate(ctx, client, &diagnostics, initialZone)

			if tt.errorContains != "" {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, diagnostics.HasError())
				assert.Contains(t, diagnostics.Errors()[0].Summary(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, diagnostics.HasError())
				assert.Equal(t, zoneId, result.GetId())
				assert.Equal(t, zoneName, result.GetName())
				assert.Equal(t, description, result.GetDescription())
				assert.Len(t, result.GetMetadata(), len(tt.expectedMetadata))
			}
		})
	}
}
