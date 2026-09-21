// Copyright © 2026. Citrix Systems, Inc.

package continuationtoken

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "continuationtoken",
	Doc:  "Checks that paginated GET requests use GetAllPagesWithRetry so results are not truncated at the API page size",
	Run:  run,
}

// requestTokenSetters are the request-builder methods that prove an endpoint actually supports paging.
// Requiring one keeps the check precise: many collection response models expose GetContinuationToken
// structurally even though their endpoint returns every record in a single response and offers no way to page.
var requestTokenSetters = []string{"ContinuationToken", "NextToken"}

// responseTokenGetters are the accessors that identify a value as a paginated collection response.
var responseTokenGetters = []string{"GetContinuationToken", "GetNextToken"}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// GetAllPagesWithRetry is the sanctioned wrapper that follows the page token internally.
			if isGetAllPagesWithRetry(call) {
				return true
			}
			// Any other fetch that is handed a paginable request and returns a paginated collection only
			// retrieves the first page. Requiring both the request token setter and the response token
			// accessor keeps this precise while catching fetch mechanisms other than ExecuteWithRetry.
			if fetchesPaginableRequest(pass, call) && returnsPaginatedCollection(pass, call) {
				pass.Reportf(call.Pos(), "paginated GET must use citrixdaasclient.GetAllPagesWithRetry so every page is retrieved, not just the first")
			}
			return true
		})
	}
	return nil, nil
}

// isGetAllPagesWithRetry reports whether the call is GetAllPagesWithRetry, unwrapping the generic
// instantiation GetAllPagesWithRetry[*T].
func isGetAllPagesWithRetry(call *ast.CallExpr) bool {
	fun := call.Fun
	switch f := fun.(type) {
	case *ast.IndexExpr:
		fun = f.X
	case *ast.IndexListExpr:
		fun = f.X
	}

	switch f := fun.(type) {
	case *ast.SelectorExpr:
		return f.Sel.Name == "GetAllPagesWithRetry"
	case *ast.Ident:
		return f.Name == "GetAllPagesWithRetry"
	}
	return false
}

// returnsPaginatedCollection reports whether any result of the call is a paginated collection type,
// identified by one of the generated page-token accessors on the response model.
func returnsPaginatedCollection(pass *analysis.Pass, call *ast.CallExpr) bool {
	resultType := pass.TypesInfo.TypeOf(call)
	if resultType == nil {
		return false
	}

	if tuple, ok := resultType.(*types.Tuple); ok {
		for i := range tuple.Len() {
			if isPaginatedCollection(tuple.At(i).Type()) {
				return true
			}
		}
		return false
	}
	return isPaginatedCollection(resultType)
}

// isPaginatedCollection reports whether values of type t expose a page-token accessor.
func isPaginatedCollection(t types.Type) bool {
	for _, getter := range responseTokenGetters {
		if hasMethod(t, getter) {
			return true
		}
	}
	return false
}

// fetchesPaginableRequest reports whether the request passed to the execution helper belongs to an endpoint
// that supports paging, identified by a token setter on the request builder type.
func fetchesPaginableRequest(pass *analysis.Pass, call *ast.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}

	requestType := pass.TypesInfo.TypeOf(call.Args[0])
	for _, setter := range requestTokenSetters {
		if hasMethod(requestType, setter) {
			return true
		}
	}
	return false
}

// hasMethod reports whether type t (or its pointer) has an accessible method with the given name.
func hasMethod(t types.Type, name string) bool {
	if t == nil {
		return false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Request builders use value receivers, but check the pointer method set too for safety.
	for _, candidate := range []types.Type{t, types.NewPointer(t)} {
		methodSet := types.NewMethodSet(candidate)
		for i := range methodSet.Len() {
			if methodSet.At(i).Obj().Name() == name {
				return true
			}
		}
	}

	return false
}
