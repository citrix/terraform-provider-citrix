// Copyright © 2026. Citrix Systems, Inc.

package testresource

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test Resource type for organizing test cases
type NilResponseTestResource struct{}

var _ = &NilResponseTestResource{}

// nilResponseRequest mimics a generated GET request: on a transport failure Execute returns a nil response
// and an error, and on success a non-nil response.
type nilResponseRequest struct{}

func (nilResponseRequest) Execute() (any, *http.Response, error) {
	return nil, nil, nil
}

func (r nilResponseRequest) After(header http.Header) nilResponseRequest {
	return r
}

// nilResponseDeleteRequest mimics a generated DELETE request, which returns only the response and an error.
type nilResponseDeleteRequest struct{}

func (nilResponseDeleteRequest) Execute() (*http.Response, error) {
	return nil, nil
}

// nilResponseSignIn mimics a hand-written helper that returns a nil response with a nil error on success.
func nilResponseSignIn() (string, *http.Response, error) {
	return "", nil, nil
}

// nilResponseOther mimics an unrelated call that also sets err.
func nilResponseOther() (any, error) {
	return nil, nil
}

// nilResponseMock mimics a test helper that always builds a response and returns no error.
func nilResponseMock() *http.Response {
	return &http.Response{}
}

// CheckDeleteResponse mimics util.CheckDeleteResponse, which returns true only when err is nil.
func CheckDeleteResponse(diagnostics any, httpResp *http.Response, err error, title string) bool {
	return err == nil
}

// Valid: the error path returns before the response is read
func (r *NilResponseTestResource) ValidErrCheckReturns() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// Valid: retry wrappers are trusted client calls too
func (r *NilResponseTestResource) ValidRetryHelper() {
	_, resp, err := citrixdaasclient.ExecuteWithRetry(nilResponseRequest{}, nil)
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// Valid: a generic retry wrapper call is resolved through its instantiation
func (r *NilResponseTestResource) ValidGenericRetryHelper() {
	_, resp, err := ExecuteWithRetry[*http.Request](nilResponseRequest{}, nil)
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// Valid: net/http's Client.Do returns a non-nil response whenever err is nil
func (r *NilResponseTestResource) ValidStdlibDo(req *http.Request) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Header.Get("Location")
}

// Valid: nil check on the left of &&
func (r *NilResponseTestResource) ValidNilGuardAnd() {
	httpResp, err := nilResponseDeleteRequest{}.Execute()
	if err != nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
}

// Valid: read inside a nil-checked block, as in the fixed ReadResource
func (r *NilResponseTestResource) ValidNilGuardBlock() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		if resp != nil {
			_ = resp.Body
			_ = resp.StatusCode
		}
	}
}

// Valid: err == nil on the left of &&
func (r *NilResponseTestResource) ValidErrNilAnd() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err == nil && resp.StatusCode == http.StatusOK {
		return
	}
}

// Valid: else branch of an err != nil check
func (r *NilResponseTestResource) ValidElseOfErrCheck() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		log.Print(err)
	} else {
		_ = resp.Header
	}
}

// Valid: early return when the response is nil
func (r *NilResponseTestResource) ValidNilReturn() {
	_, resp, err := nilResponseRequest{}.Execute()
	if resp == nil {
		log.Print(err)
		return
	}
	_ = resp.StatusCode
}

// Valid: resp == nil short-circuits the || before the read
func (r *NilResponseTestResource) ValidOrShortCircuit() {
	_, resp, err := nilResponseRequest{}.Execute()
	if resp == nil || resp.StatusCode != http.StatusOK {
		log.Print(err)
	}
}

// Valid: deferred Body.Close after the error check
func (r *NilResponseTestResource) ValidDeferAfterErrCheck() {
	resp, err := nilResponseDeleteRequest{}.Execute()
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// Valid: the response is used only after the shared delete helper confirms the call succeeded
func (r *NilResponseTestResource) ValidAfterCheckDeleteResponse() {
	httpResp, err := nilResponseDeleteRequest{}.Execute()
	if !CheckDeleteResponse(nil, httpResp, err, "Error deleting") {
		return
	}
	_ = *httpResp
}

// Valid: the response is used inside the success branch of the shared delete helper
func (r *NilResponseTestResource) ValidInsideCheckDeleteResponse() {
	httpResp, err := nilResponseDeleteRequest{}.Execute()
	if CheckDeleteResponse(nil, httpResp, err, "Error deleting") {
		_ = httpResp.Header
	}
}

// Valid: a response from a call without an error is not checked
func (r *NilResponseTestResource) ValidNoErrorReturned() {
	resp := nilResponseMock()
	_ = resp.StatusCode
}

// Valid: the reassigned response is checked before it is read
func (r *NilResponseTestResource) ValidReassignedChecked() {
	_, resp, err := nilResponseRequest{}.Execute()
	log.Print(resp, err)
	_, resp, err = nilResponseRequest{}.Execute()
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// Valid: a read in the right-hand side of a reassignment uses the previous, checked response
func (r *NilResponseTestResource) ValidReadInReassignRHS() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		return
	}
	_, resp, err = nilResponseRequest{}.After(resp.Header).Execute()
	log.Print(resp, err)
}

// Valid: the assignment is in the if's init statement and the branch returns
func (r *NilResponseTestResource) ValidIfInitAssignment() {
	var resp *http.Response
	var err error
	if _, resp, err = (nilResponseRequest{}).Execute(); err != nil {
		return
	}
	_ = resp.StatusCode
}

// Valid: t.Fatalf ends the test on the error path
func (r *NilResponseTestResource) ValidTFatal(t *testing.T) {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = resp.StatusCode
}

// Valid: require.NotNil stops the test before a nil response is read
func (r *NilResponseTestResource) ValidRequireNotNil(t *testing.T) {
	_, resp, err := nilResponseRequest{}.Execute()
	log.Print(err)
	require.NotNil(t, resp)
	_ = resp.StatusCode
}

// Valid: require.NoError stops the test unless the trusted call succeeded
func (r *NilResponseTestResource) ValidRequireNoError(t *testing.T) {
	_, resp, err := nilResponseRequest{}.Execute()
	require.NoError(t, err)
	_ = resp.StatusCode
}

// Valid: require.New(t) assertions stop the test the same way
func (r *NilResponseTestResource) ValidRequireAssertions(t *testing.T) {
	assertions := require.New(t)
	_, resp, err := nilResponseRequest{}.Execute()
	assertions.NoError(err)
	_ = resp.StatusCode
}

// Valid: a non-trusted helper's response is read only after a nil check
func (r *NilResponseTestResource) ValidNonTrustedNilCheck() {
	_, resp, err := nilResponseSignIn()
	if err != nil {
		return
	}
	if resp != nil {
		_ = resp.StatusCode
	}
}

// violations:1
// Invalid: the old Delete pattern, read on the error path without a nil check
func (r *NilResponseTestResource) InvalidOldDeletePattern() {
	httpResp, err := nilResponseDeleteRequest{}.Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		return
	}
}

// violations:1
// Invalid: read inside the error branch
func (r *NilResponseTestResource) InvalidReadInErrBranch() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return
		}
	}
}

// violations:1
// Invalid: read inside the error branch of a generic retry wrapper call
func (r *NilResponseTestResource) InvalidGenericReadInErrBranch() {
	_, resp, err := ExecuteWithRetry[*http.Request](nilResponseRequest{}, nil)
	if err != nil {
		log.Print(resp.Status)
	}
}

// violations:1
// Invalid: read right after the call with no check at all
func (r *NilResponseTestResource) InvalidNoCheck() {
	_, resp, err := nilResponseRequest{}.Execute()
	if resp.StatusCode == http.StatusForbidden {
		log.Print(err)
	}
}

// violations:1
// Invalid: the error branch logs but does not return, so execution reaches the read
func (r *NilResponseTestResource) InvalidErrBranchFallsThrough() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		log.Print(err)
	}
	_ = resp.StatusCode
}

// violations:2
// Invalid: two reads on the error path
func (r *NilResponseTestResource) InvalidTwoReads() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		_ = resp.StatusCode
		_ = resp.Body
	}
}

// violations:1
// Invalid: deferred Body.Close before the error is checked
func (r *NilResponseTestResource) InvalidDeferBeforeCheck() {
	resp, err := nilResponseDeleteRequest{}.Execute()
	log.Print(err)
	defer resp.Body.Close()
}

// violations:1
// Invalid: require.NotNil on a different response does not guard this one
func (r *NilResponseTestResource) InvalidRequireOtherResponse(t *testing.T) {
	_, resp, err := nilResponseRequest{}.Execute()
	_, other, otherErr := nilResponseRequest{}.Execute()
	log.Print(err, otherErr)
	require.NotNil(t, other)
	_ = resp.StatusCode
}

// violations:1
// Invalid: the guard sits between two assignments, so it does not cover the second response
func (r *NilResponseTestResource) InvalidGuardBeforeReassign() {
	_, resp, err := nilResponseRequest{}.Execute()
	if err != nil {
		return
	}
	log.Print(resp)
	_, resp, err = nilResponseRequest{}.Execute()
	log.Print(err)
	_ = resp.StatusCode
}

// violations:1
// Invalid: err was overwritten by another call before it was checked
func (r *NilResponseTestResource) InvalidErrOverwritten() {
	_, resp, err := nilResponseRequest{}.Execute()
	log.Print(err)
	_, err = nilResponseOther()
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// violations:1
// Invalid: the outer err == nil was evaluated before the response was reassigned
func (r *NilResponseTestResource) InvalidOuterErrNilBeforeReassign() {
	_, resp, err := nilResponseRequest{}.Execute()
	log.Print(resp)
	if err == nil {
		_, resp, err = nilResponseRequest{}.Execute()
		log.Print(err)
		_ = resp.StatusCode
	}
}

// violations:1
// Invalid: an error check does not prove a non-trusted helper's response is non-nil
func (r *NilResponseTestResource) InvalidNonTrustedErrGuard() {
	_, resp, err := nilResponseSignIn()
	if err != nil {
		return
	}
	_ = resp.StatusCode
}

// violations:1
// Invalid: require.NoError does not prove a non-trusted helper's response is non-nil
func (r *NilResponseTestResource) InvalidRequireNoErrorNonTrusted(t *testing.T) {
	_, resp, err := nilResponseSignIn()
	require.NoError(t, err)
	_ = resp.StatusCode
}

// violations:2
// Invalid: dereferencing the response and calling its methods both panic on nil
func (r *NilResponseTestResource) InvalidDerefAndMethod() {
	resp, err := nilResponseDeleteRequest{}.Execute()
	log.Print(err)
	_ = *resp
	_ = resp.Cookies()
}
