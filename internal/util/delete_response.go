// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"net/http"

	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// IsNotFound reports whether the API answered 404 Not Found. A nil response (transport failure) is not a 404.
func IsNotFound(httpResp *http.Response) bool {
	return httpResp != nil && httpResp.StatusCode == http.StatusNotFound
}

// CheckDeleteResponse returns true when the delete was accepted and the caller should continue, for example to poll the async job.
// A 404 means the object is already gone: it returns false and adds nothing, so the framework removes the resource from state.
// Any other failure, including one with no response, adds an error and returns false.
func CheckDeleteResponse(diagnostics *diag.Diagnostics, httpResp *http.Response, err error, errTitle string) bool {
	if err == nil {
		return true
	}
	if IsNotFound(httpResp) {
		return false
	}
	diagnostics.AddError(
		errTitle,
		"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
			"\nError message: "+ReadClientError(err),
	)
	return false
}
