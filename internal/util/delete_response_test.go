// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResponse(statusCode int, transactionId string) *http.Response {
	header := http.Header{}
	header.Set("Citrix-TransactionId", transactionId)
	return &http.Response{StatusCode: statusCode, Header: header}
}

func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(nil))
	assert.True(t, IsNotFound(newResponse(http.StatusNotFound, "txn-404")))
	assert.False(t, IsNotFound(newResponse(http.StatusInternalServerError, "txn-500")))
}

func TestCheckDeleteResponse_NoError(t *testing.T) {
	tests := []struct {
		name          string
		httpResp      *http.Response
		err           error
		expectProceed bool
	}{
		{"accepted", newResponse(http.StatusAccepted, "txn-202"), nil, true},
		{"already deleted", newResponse(http.StatusNotFound, "txn-404"), errors.New("404 Not Found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diagnostics diag.Diagnostics

			proceed := CheckDeleteResponse(&diagnostics, tt.httpResp, tt.err, "Error deleting Zone zone-1")

			assert.Equal(t, tt.expectProceed, proceed)
			assert.False(t, diagnostics.HasError())
		})
	}
}

func TestCheckDeleteResponse_ReportsError(t *testing.T) {
	dnsErr := &url.Error{Op: "Delete", URL: "https://api.cloud.com/cvad/manage/Zones/zone-1", Err: errors.New("dial tcp: lookup api.cloud.com: no such host")}

	tests := []struct {
		name           string
		httpResp       *http.Response
		err            error
		expectedDetail string
	}{
		{"no response", nil, dnsErr, "no such host"},
		{"cancelled run", nil, context.Canceled, "context canceled"},
		{"server error", newResponse(http.StatusInternalServerError, "txn-500"), errors.New("500 Internal Server Error"), "TransactionId: txn-500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diagnostics diag.Diagnostics

			proceed := CheckDeleteResponse(&diagnostics, tt.httpResp, tt.err, "Error deleting Zone zone-1")

			assert.False(t, proceed)
			require.Equal(t, 1, diagnostics.ErrorsCount())
			assert.Equal(t, "Error deleting Zone zone-1", diagnostics.Errors()[0].Summary())
			assert.Contains(t, diagnostics.Errors()[0].Detail(), tt.expectedDetail)
		})
	}
}
