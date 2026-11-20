// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// noResponseRequest mimics a generated GET request whose transport failed, so Execute returns no HTTP response.
type noResponseRequest struct {
	err error
}

func (r noResponseRequest) Execute() (*citrixorchestration.ZoneDetailResponseModel, *http.Response, error) {
	return nil, nil, r.err
}

func TestReadResource_NoResponse(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expectedText string
	}{
		{"zone", "Zone", "Error Reading Zone res-1"},
		{"resource location", "Resource Location", "Error Reading Resource Location res-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := test.NewTestDaaSClient()
			resp := &resource.ReadResponse{}
			request := noResponseRequest{err: errors.New("dial tcp: lookup api.cloud.com: no such host")}

			_, httpResp, err := ReadResource[*citrixorchestration.ZoneDetailResponseModel](request, test.DaaSTestContext(), client, resp, tt.resourceType, "res-1")

			require.Error(t, err)
			assert.Nil(t, httpResp)
			require.Equal(t, 1, resp.Diagnostics.ErrorsCount())
			assert.Equal(t, tt.expectedText, resp.Diagnostics.Errors()[0].Summary())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "no such host")
		})
	}
}

func TestProcessAsyncJobResponse_NoResponse(t *testing.T) {
	tests := []struct {
		name             string
		priorError       bool
		expectedErrors   int
		expectedFirstErr string
	}{
		{"adds an error when the caller has not", false, 1, "Error creating Admin xac"},
		{"keeps the caller's request error only", true, 1, "Error creating Admin: request failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := test.NewTestDaaSClient()
			var diagnostics diag.Diagnostics
			if tt.priorError {
				diagnostics.AddError("Error creating Admin: request failed", "dial tcp: connection refused")
			}

			err := ProcessAsyncJobResponse(test.DaaSTestContext(), client, nil, "Error creating Admin xac", &diagnostics, 5)

			var pollErr *JobPollError
			require.ErrorAs(t, err, &pollErr)
			require.Equal(t, tt.expectedErrors, diagnostics.ErrorsCount())
			assert.Equal(t, tt.expectedFirstErr, diagnostics.Errors()[0].Summary())
		})
	}
}

func TestGetAsyncJobResult_NoResponse(t *testing.T) {
	client, _ := test.NewTestDaaSClient()
	var diagnostics diag.Diagnostics

	result, err := GetAsyncJobResult[*citrixorchestration.ZoneDetailResponseModel](test.DaaSTestContext(), client, nil, "Error creating Zone z", &diagnostics, 5)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, diagnostics.ErrorsCount())
}

func TestProcessAsyncJobResponse_PollingErrorKeepsCause(t *testing.T) {
	client, mockClient := test.NewTestDaaSClient()
	defer mockClient.AssertExpectations(t)
	mockJobsAPI := citrixorchestration.GetMockJobsAPIsDAAS(mockClient.APIClient)
	mockJobsAPI.On("JobsGetJobExecute", mock.Anything).Return((*citrixorchestration.JobResponseModel)(nil), (*http.Response)(nil), context.Canceled).Once()
	var diagnostics diag.Diagnostics

	err := ProcessAsyncJobResponse(test.DaaSTestContext(), client, testutil.MockAsyncResponse("job-123", "txn-1"), "Error deleting Admin Scope: s", &diagnostics, 5)

	var pollErr *JobPollError
	require.ErrorAs(t, err, &pollErr)
	require.Equal(t, 1, diagnostics.ErrorsCount())
	detail := diagnostics.Errors()[0].Detail()
	assert.Contains(t, detail, "JobId: job-123")
	assert.Contains(t, detail, "context canceled")
}

func TestReadErrorHelpers_NilError(t *testing.T) {
	readers := map[string]func(error) string{
		"ReadClientError":               ReadClientError,
		"ReadGacError":                  ReadGacError,
		"ReadQcsClientError":            ReadQcsClientError,
		"ReadCatalogServiceClientError": ReadCatalogServiceClientError,
	}
	for name, read := range readers {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, noClientErrorDetails, read(nil))
			assert.Equal(t, "boom", read(errors.New("boom")))
		})
	}
}
