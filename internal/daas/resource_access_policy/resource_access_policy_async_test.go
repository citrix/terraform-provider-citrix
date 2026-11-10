// Copyright © 2026. Citrix Systems, Inc.

package resource_access_policy // whitebox testing in same package

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type setupAsyncMocks func(*citrixorchestration.MockDeliveryGroupsAPIsDAAS, *citrixorchestration.MockJobsAPIsDAAS)

func policyCollection(policies ...*citrixorchestration.ResourceAccessPolicyResponseModel) *citrixorchestration.ResourceAccessPolicyResponseModelCollection {
	items := []citrixorchestration.ResourceAccessPolicyResponseModel{}
	for _, policy := range policies {
		items = append(items, *policy)
	}
	return citrixorchestration.NewResourceAccessPolicyResponseModelCollection(items)
}

func TestCreateResourceAccessPolicy(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name          string
		setupMocks    setupAsyncMocks
		expectedId    int32
		errorContains string
	}{
		{
			name: "Finds the created policy by name once the job completes",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsCreateResourceAccessPolicyExecute", mock.Anything).Return((*citrixorchestration.ResourceAccessPolicyResponseModel)(nil), testutil.MockAsyncResponse("job-create", "txn-create"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob("job-create", citrixorchestration.JOBTYPE_CREATE_RESOURCE_ACCESS_POLICY), testutil.MockSuccessResponse(), nil).Once()
				mockDeliveryGroupsAPI.On("DeliveryGroupsGetResourceAccessPoliciesExecute", mock.Anything).Return(policyCollection(newPolicy(1, "other"), newPolicy(7, "Policy")), testutil.MockSuccessResponse(), nil).Once()
			},
			expectedId: 7,
		},
		{
			name: "Create request fails",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsCreateResourceAccessPolicyExecute", mock.Anything).Return((*citrixorchestration.ResourceAccessPolicyResponseModel)(nil), testutil.MockErrorResponse(400, "txn-error"), assert.AnError).Once()
			},
			errorContains: assert.AnError.Error(),
		},
		{
			name: "Job fails",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsCreateResourceAccessPolicyExecute", mock.Anything).Return((*citrixorchestration.ResourceAccessPolicyResponseModel)(nil), testutil.MockAsyncResponse("job-fail", "txn-fail"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockFailedJob("job-fail", "At least one of TagUid or ResourceType must be provided"), testutil.MockSuccessResponse(), nil).Once()
			},
			errorContains: "At least one of TagUid or ResourceType must be provided",
		},
		{
			name: "Created policy is not found",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsCreateResourceAccessPolicyExecute", mock.Anything).Return((*citrixorchestration.ResourceAccessPolicyResponseModel)(nil), testutil.MockAsyncResponse("job-create", "txn-create"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob("job-create", citrixorchestration.JOBTYPE_CREATE_RESOURCE_ACCESS_POLICY), testutil.MockSuccessResponse(), nil).Once()
				mockDeliveryGroupsAPI.On("DeliveryGroupsGetResourceAccessPoliciesExecute", mock.Anything).Return(policyCollection(), testutil.MockSuccessResponse(), nil).Once()
			},
			errorContains: "resource access policy policy was not found after it was created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)
			tt.setupMocks(citrixorchestration.GetMockDeliveryGroupsAPIsDAAS(mockClient.APIClient), citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient))

			var body citrixorchestration.ResourceAccessPolicyRequestModel
			body.SetName("policy")

			diagnostics := diag.Diagnostics{}
			policy, err := createResourceAccessPolicy(test.DaaSTestContext(), client, &diagnostics, body, 5, "summary")

			if tt.errorContains != "" {
				require.Error(t, err)
				require.True(t, diagnostics.HasError())
				assert.Contains(t, diagnostics.Errors()[0].Detail(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.False(t, diagnostics.HasError())
			assert.Equal(t, tt.expectedId, policy.GetId())
		})
	}
}

func TestUpdateResourceAccessPolicy(t *testing.T) {
	tests := []struct {
		name          string
		remote        *citrixorchestration.ResourceAccessPolicyResponseModel
		errorContains string
	}{
		{
			name:   "Finds the updated policy by id once the job completes",
			remote: newPolicy(7, "renamed"),
		},
		{
			name:          "Updated policy is not found",
			remote:        newPolicy(8, "other"),
			errorContains: "resource access policy 7 was not found after it was updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)

			mockDeliveryGroupsAPI := citrixorchestration.GetMockDeliveryGroupsAPIsDAAS(mockClient.APIClient)
			mockJobsAPI := citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient)
			mockDeliveryGroupsAPI.On("DeliveryGroupsPatchResourceAccessPolicyExecute", mock.Anything).Return((*citrixorchestration.ResourceAccessPolicyResponseModel)(nil), testutil.MockAsyncResponse("job-update", "txn-update"), nil).Once()
			mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob("job-update", citrixorchestration.JOBTYPE_UPDATE_RESOURCE_ACCESS_POLICY), testutil.MockSuccessResponse(), nil).Once()
			mockDeliveryGroupsAPI.On("DeliveryGroupsGetResourceAccessPoliciesExecute", mock.Anything).Return(policyCollection(tt.remote), testutil.MockSuccessResponse(), nil).Once()

			diagnostics := diag.Diagnostics{}
			policy, err := updateResourceAccessPolicy(test.DaaSTestContext(), client, &diagnostics, 7, citrixorchestration.ResourceAccessPolicyRequestModel{}, 5, "summary")

			if tt.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, diagnostics.Errors()[0].Detail(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "renamed", policy.GetName())
		})
	}
}

func TestDeleteResourceAccessPolicy(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    setupAsyncMocks
		errorContains string
	}{
		{
			name: "Waits for the job to complete",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsDeleteResourceAccessPolicyExecute", mock.Anything).Return(testutil.MockAsyncResponse("job-delete", "txn-delete"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockCompletedJob("job-delete", citrixorchestration.JOBTYPE_DELETE_RESOURCE_ACCESS_POLICY), testutil.MockSuccessResponse(), nil).Once()
			},
		},
		{
			name: "Already deleted",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsDeleteResourceAccessPolicyExecute", mock.Anything).Return(testutil.MockErrorResponse(404, "txn-404"), assert.AnError).Once()
			},
		},
		{
			name: "Job fails",
			setupMocks: func(mockDeliveryGroupsAPI *citrixorchestration.MockDeliveryGroupsAPIsDAAS, mockJobsAPI *citrixorchestration.MockJobsAPIsDAAS) {
				mockDeliveryGroupsAPI.On("DeliveryGroupsDeleteResourceAccessPolicyExecute", mock.Anything).Return(testutil.MockAsyncResponse("job-fail", "txn-fail"), nil).Once()
				mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return(testutil.MockFailedJob("job-fail", "delete failed"), testutil.MockSuccessResponse(), nil).Once()
			},
			errorContains: "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)
			tt.setupMocks(citrixorchestration.GetMockDeliveryGroupsAPIsDAAS(mockClient.APIClient), citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient))

			diagnostics := diag.Diagnostics{}
			err := deleteResourceAccessPolicy(test.DaaSTestContext(), client, &diagnostics, 7, 5, "summary")

			if tt.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, diagnostics.Errors()[0].Detail(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			assert.False(t, diagnostics.HasError())
		})
	}
}
