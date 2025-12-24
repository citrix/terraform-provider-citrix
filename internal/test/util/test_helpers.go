// Copyright © 2025. Citrix Systems, Inc.

package util

import (
	"net/http"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
)

// ========== Test Helpers ==========

// NewMetadataPair creates a NameValueStringPairModel with the given key and value.
func NewMetadataPair(name, value string) citrixorchestration.NameValueStringPairModel {
	pair := citrixorchestration.NameValueStringPairModel{}
	pair.SetName(name)
	pair.SetValue(value)
	return pair
}

// MockAsyncResponse creates a mock HTTP response for async operations with job ID and transaction ID.
func MockAsyncResponse(jobId, transactionId string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusAccepted,
		Header: http.Header{
			"Location":             []string{"https://api.cloud.com/jobs/" + jobId},
			"Citrix-TransactionId": []string{transactionId},
		},
	}
}

// MockErrorResponse creates a mock HTTP error response with status code and transaction ID.
func MockErrorResponse(statusCode int, transactionId string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Citrix-TransactionId": []string{transactionId},
		},
	}
}

// mockSuccessResponse creates a mock HTTP success response.
func MockSuccessResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK}
}

// MockCompletedJob creates a job response model with completed status.
func MockCompletedJob(jobId string, jobType citrixorchestration.JobType) *citrixorchestration.JobResponseModel {
	return &citrixorchestration.JobResponseModel{
		Id:     jobId,
		Status: citrixorchestration.JOBSTATUS_COMPLETE,
		Type:   jobType,
	}
}

// MockFailedJob creates a job response model with failed status and error message.
func MockFailedJob(jobId, errorMessage string) *citrixorchestration.JobResponseModel {
	job := &citrixorchestration.JobResponseModel{
		Id:     jobId,
		Status: citrixorchestration.JOBSTATUS_FAILED,
	}
	job.ErrorString.Set(&errorMessage)
	return job
}
