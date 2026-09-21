// Copyright © 2026. Citrix Systems, Inc.

package testresource

import (
	"context"
	"net/http"
)

// Test Resource type for organizing test cases
type ContinuationTokenTestResource struct{}

var _ = &ContinuationTokenTestResource{}

type MockItem struct {
	Id string
}

// Mock paginated collection response model, mirroring the generated *ResponseModelCollection types.
type MockResponseModelCollection struct{}

func (m *MockResponseModelCollection) GetItems() []MockItem         { return nil }
func (m *MockResponseModelCollection) GetContinuationToken() string { return "" }

// mockPaginableRequest mirrors a generated request builder for an endpoint that supports paging.
type mockPaginableRequest struct{}

func (r mockPaginableRequest) Limit(limit int32) mockPaginableRequest              { return r }
func (r mockPaginableRequest) ContinuationToken(token string) mockPaginableRequest { return r }

// mockNextTokenRequest mirrors a request builder that pages with NextToken (Global App Configuration).
type mockNextTokenRequest struct{}

func (r mockNextTokenRequest) NextToken(token string) mockNextTokenRequest { return r }

// mockNonPaginableRequest mirrors a request builder for an endpoint that returns every record at once
// and offers no way to page.
type mockNonPaginableRequest struct{}

func (r mockNonPaginableRequest) Fields(fields string) mockNonPaginableRequest { return r }

// mockDetailResponse is a non-paginated response model (no page-token accessor).
type mockDetailResponse struct{}

func (m *mockDetailResponse) GetId() string { return "" }

// ExecuteWithRetry mirrors the SDK execution helper: the request is the first argument.
func ExecuteWithRetry[T any](request any, client any) (T, *http.Response, error) {
	var zero T
	return zero, nil, nil
}

// FetchCollection stands in for a fetch mechanism other than ExecuteWithRetry that returns a collection.
func FetchCollection[T any](request any, client any) (T, *http.Response, error) {
	var zero T
	return zero, nil, nil
}

// GetAllPagesWithRetry mirrors the sanctioned wrapper that follows the page token internally.
func GetAllPagesWithRetry[T any](request any, client any) (T, *http.Response, error) {
	var zero T
	return zero, nil, nil
}

var mockClientValue = struct{}{}

func newPaginableRequest() mockPaginableRequest       { return mockPaginableRequest{} }
func newNextTokenRequest() mockNextTokenRequest       { return mockNextTokenRequest{} }
func newNonPaginableRequest() mockNonPaginableRequest { return mockNonPaginableRequest{} }

// Valid: paginable endpoint fetched through the sanctioned helper
func (r *ContinuationTokenTestResource) ValidUsesHelper(ctx context.Context) {
	request := newPaginableRequest()
	responseModel, _, _ := GetAllPagesWithRetry[*MockResponseModelCollection](request, mockClientValue)
	_ = responseModel.GetItems()
}

// Valid: endpoint that does not support paging may use ExecuteWithRetry directly
func (r *ContinuationTokenTestResource) ValidNonPaginableExecuteWithRetry(ctx context.Context) {
	request := newNonPaginableRequest().Fields("Id,Name")
	responseModel, _, _ := ExecuteWithRetry[*MockResponseModelCollection](request, mockClientValue)
	_ = responseModel.GetItems()
}

// Valid: transform helper that receives an already-fetched collection and only maps it; no fetch here
func (r *ContinuationTokenTestResource) ValidTransformReceivesCollection(ctx context.Context, collection *MockResponseModelCollection) {
	_ = collection.GetItems()
}

// violations:1
// Invalid: ExecuteWithRetry on a paginable endpoint must use GetAllPagesWithRetry
func (r *ContinuationTokenTestResource) InvalidExecuteWithRetryPaginable(ctx context.Context) {
	request := newPaginableRequest()
	responseModel, _, _ := ExecuteWithRetry[*MockResponseModelCollection](request, mockClientValue)
	_ = responseModel.GetItems()
}

// violations:1
// Invalid: a hand-written pagination loop is still flagged; GetAllPagesWithRetry is required
func (r *ContinuationTokenTestResource) InvalidHandWrittenLoop(ctx context.Context) {
	request := newPaginableRequest()
	continuationToken := ""
	for {
		request = request.ContinuationToken(continuationToken)
		responseModel, _, _ := ExecuteWithRetry[*MockResponseModelCollection](request, mockClientValue)
		if responseModel.GetContinuationToken() == "" {
			break
		}
		continuationToken = responseModel.GetContinuationToken()
	}
}

// Valid: paginable request whose response is not a paginated collection needs no helper
func (r *ContinuationTokenTestResource) ValidPaginableRequestNonCollectionResponse(ctx context.Context) {
	request := newPaginableRequest()
	detail, _, _ := ExecuteWithRetry[*mockDetailResponse](request, mockClientValue)
	_ = detail.GetId()
}

// violations:1
// Invalid: a fetch mechanism other than ExecuteWithRetry that returns a paginated collection
func (r *ContinuationTokenTestResource) InvalidCustomFetchMechanism(ctx context.Context) {
	request := newPaginableRequest()
	responseModel, _, _ := FetchCollection[*MockResponseModelCollection](request, mockClientValue)
	_ = responseModel.GetItems()
}

// violations:1
// Invalid: NextToken-paged endpoint using ExecuteWithRetry
func (r *ContinuationTokenTestResource) InvalidNextTokenExecuteWithRetry(ctx context.Context) {
	request := newNextTokenRequest()
	responseModel, _, _ := ExecuteWithRetry[*MockResponseModelCollection](request, mockClientValue)
	_ = responseModel.GetItems()
}

// violations:2
// Invalid: two separate paginable ExecuteWithRetry fetches
func (r *ContinuationTokenTestResource) InvalidMultiplePaginableFetches(ctx context.Context) {
	firstRequest := newPaginableRequest()
	first, _, _ := ExecuteWithRetry[*MockResponseModelCollection](firstRequest, mockClientValue)
	_ = first.GetItems()

	secondRequest := newPaginableRequest()
	second, _, _ := ExecuteWithRetry[*MockResponseModelCollection](secondRequest, mockClientValue)
	_ = second.GetItems()
}
