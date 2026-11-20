// Copyright © 2026. Citrix Systems, Inc.

package machine_catalog // whitebox testing in same package

import (
	"net/http"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type setupCatalogJobMocks func(*citrixorchestration.MockMachineCatalogsAPIsDAAS, *citrixorchestration.MockJobsAPIsDAAS)

func errorResponseWithTransactionId(statusCode int, transactionId string) *http.Response {
	header := http.Header{}
	header.Set("Citrix-TransactionId", transactionId)
	return &http.Response{StatusCode: statusCode, Header: header}
}

func newCatalogJobMocks(t *testing.T, mockClient *citrixorchestration.MockAPIClient) (*citrixorchestration.MockMachineCatalogsAPIsDAAS, *citrixorchestration.MockJobsAPIsDAAS) {
	mockCatalogsAPI := citrixorchestration.GetMockMachineCatalogsAPIsDAAS(mockClient.APIClient)
	mockJobsAPI := citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient)
	mockCatalogsAPI.Test(t)
	mockJobsAPI.Test(t)
	return mockCatalogsAPI, mockJobsAPI
}

func TestPublishCatalogImage(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks setupCatalogJobMocks
		wantErr    string
		wantDiag   string
	}{
		{
			name: "Request gets no response",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, _ *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogProvisioningSchemeExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), (*http.Response)(nil), assert.AnError).Once()
			},
			wantErr:  assert.AnError.Error(),
			wantDiag: "TransactionId: failed before request was sent\nError message: " + assert.AnError.Error(),
		},
		{
			name: "Request is rejected",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, _ *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogProvisioningSchemeExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), errorResponseWithTransactionId(http.StatusBadRequest, "txn-rejected"), assert.AnError).Once()
			},
			wantErr:  assert.AnError.Error(),
			wantDiag: "TransactionId: txn-rejected\nError message: " + assert.AnError.Error(),
		},
		{
			name: "Job cannot be polled",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogProvisioningSchemeExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), testutil.MockAsyncResponse("job-image", "txn-image"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), (*http.Response)(nil), assert.AnError).Once()
			},
			wantErr:  assert.AnError.Error(),
			wantDiag: "JobId: ",
		},
		{
			name: "Job fails, update continues",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogProvisioningSchemeExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), testutil.MockAsyncResponse("job-image", "txn-image"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockFailedJob("job-image", "Image preparation failed"), testutil.MockSuccessResponse(), nil).Once()
			},
			wantDiag: "Image preparation failed",
		},
		{
			name: "Job completes",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogProvisioningSchemeExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), testutil.MockAsyncResponse("job-image", "txn-image"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob("job-image", citrixorchestration.JOBTYPE_UPDATE_MACHINE_CATALOG_PROVISIONING_SCHEME), testutil.MockSuccessResponse(), nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)
			tt.setupMocks(newCatalogJobMocks(t, mockClient))

			diagnostics := diag.Diagnostics{}
			err := publishCatalogImage(test.DaaSTestContext(), client, &diagnostics, "catalog-id", "catalog", citrixorchestration.UpdateMachineCatalogProvisioningSchemeRequestModel{}, 5)

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
			if tt.wantDiag == "" {
				assert.False(t, diagnostics.HasError())
				return
			}
			require.Equal(t, 1, diagnostics.ErrorsCount())
			assert.Contains(t, diagnostics.Errors()[0].Detail(), tt.wantDiag)
		})
	}
}

func TestUpdateMemoryAndCpuCount(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks setupCatalogJobMocks
		wantErr    string
		wantDiag   string
	}{
		{
			name: "Job cannot be polled",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogExecute", mock.Anything).Return((*citrixorchestration.MachineCatalogDetailResponseModel)(nil), testutil.MockAsyncResponse("job-cpu", "txn-cpu"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), (*http.Response)(nil), assert.AnError).Once()
			},
			wantErr:  assert.AnError.Error(),
			wantDiag: "JobId: ",
		},
		{
			name: "Job fails, update continues",
			setupMocks: func(mockCatalogsAPI *citrixorchestration.MockMachineCatalogsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockCatalogsAPI.On("MachineCatalogsUpdateMachineCatalogExecute", mock.Anything).Return((*citrixorchestration.MachineCatalogDetailResponseModel)(nil), testutil.MockAsyncResponse("job-cpu", "txn-cpu"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockFailedJob("job-cpu", "Not enough memory"), testutil.MockSuccessResponse(), nil).Once()
			},
			wantDiag: "Not enough memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)
			tt.setupMocks(newCatalogJobMocks(t, mockClient))

			catalog := &citrixorchestration.MachineCatalogDetailResponseModel{}
			catalog.SetId("catalog-id")
			catalog.SetName("catalog")
			resp := &resource.UpdateResponse{}
			attributesMap, _ := util.ResourceAttributeMapFromObject(ProvisioningSchemeModel{})
			plan := MachineCatalogResourceModel{ProvisioningScheme: types.ObjectNull(attributesMap)}

			err := updateMemoryAndCpuCount(test.DaaSTestContext(), client, resp, catalog, plan, citrixorchestration.HYPERVISORCONNECTIONTYPE_OPEN_SHIFT)

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
			require.Equal(t, 1, resp.Diagnostics.ErrorsCount())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tt.wantDiag)
		})
	}
}
