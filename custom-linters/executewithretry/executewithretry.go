// Copyright © 2026. Citrix Systems, Inc.

package executewithretry

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "executewithretry",
	Doc:  "Checks that API GET calls use ExecuteWithRetry wrapper instead of calling Execute() directly for proper resilience",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if this is ExecuteWithRetry being called incorrectly with StoreFront
			if isExecuteWithRetryCall(callExpr) {
				if isStoreFrontOperation(callExpr, pass) {
					pass.Reportf(callExpr.Pos(), "StoreFront operations should not use ExecuteWithRetry - StoreFront Execute() methods return (result, error) without *http.Response")
					return true
				}
				// ExecuteWithRetry used correctly with non-StoreFront operations
				return true
			}

			// Check if this is a .Execute() call
			if !isExecuteCall(callExpr) {
				return true
			}

			// Check if there's a GET request anywhere in the call chain leading to Execute()
			if hasGetRequestInChain(callExpr, pass) {
				// GET operations must use ExecuteWithRetry, not direct Execute()
				pass.Reportf(callExpr.Pos(), "GET operation should use ExecuteWithRetry wrapper instead of calling Execute() directly for proper resilience")
				return true
			}

			// Non-GET operations can use direct Execute() - no violation
			return true
		})
	}
	return nil, nil
}

// isExecuteCall checks if the call expression is calling .Execute()
func isExecuteCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "Execute"
}

// hasGetRequestInChain recursively searches for a GET request type anywhere in the expression
// tree leading to the Execute() call. This handles any chaining pattern like:
// - request.Execute()
// - AddRequestData(request, ...).Execute()
// - AddRequestData(request, ...).Async(true).Execute()
// - futureWrapper(request, ...).SomeMethod().Execute()
func hasGetRequestInChain(call *ast.CallExpr, pass *analysis.Pass) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Recursively check the receiver/base expression
	return containsGetRequest(sel.X, pass)
}

// containsGetRequest recursively searches an expression tree for GET request types
func containsGetRequest(expr ast.Expr, pass *analysis.Pass) bool {
	if expr == nil {
		return false
	}

	// Check the type of this expression
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType != nil {
		typeName := exprType.String()
		if isGetRequestType(typeName) {
			return true
		}
	}

	// Recursively check sub-expressions based on expression type
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Check function arguments (e.g., AddRequestData(request, client))
		for _, arg := range e.Args {
			if containsGetRequest(arg, pass) {
				return true
			}
		}
		// Check the function being called (for chained calls)
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			if containsGetRequest(sel.X, pass) {
				return true
			}
		}
	case *ast.SelectorExpr:
		// Check the base of the selector (e.g., obj.Method -> check obj)
		return containsGetRequest(e.X, pass)
	case *ast.Ident:
		// Terminal case - check if this identifier is a GET request
		identType := pass.TypesInfo.TypeOf(e)
		if identType != nil && isGetRequestType(identType.String()) {
			return true
		}
	}

	return false
}

// isExecuteWithRetryCall checks if the call expression is calling ExecuteWithRetry
func isExecuteWithRetryCall(call *ast.CallExpr) bool {
	// ExecuteWithRetry is a generic function, so it could be called as:
	// citrixdaasclient.ExecuteWithRetry[Type](...)
	// or citrixclient.ExecuteWithRetry[Type](...)

	// Check for index expression (generic function call)
	indexExpr, ok := call.Fun.(*ast.IndexExpr)
	if !ok {
		indexListExpr, ok := call.Fun.(*ast.IndexListExpr)
		if !ok {
			return false
		}
		// Check the function being indexed
		sel, ok := indexListExpr.X.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		return sel.Sel.Name == "ExecuteWithRetry"
	}

	// Check the function being indexed
	sel, ok := indexExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "ExecuteWithRetry"
}

// isStoreFrontOperation checks if the ExecuteWithRetry call is being used with a StoreFront request
func isStoreFrontOperation(call *ast.CallExpr, pass *analysis.Pass) bool {
	// ExecuteWithRetry takes the request as its first argument
	if len(call.Args) < 1 {
		return false
	}

	// Get the type of the first argument (the request)
	requestArg := call.Args[0]
	requestType := pass.TypesInfo.TypeOf(requestArg)
	if requestType == nil {
		return false
	}

	typeName := requestType.String()
	// StoreFront request types are in the citrixstorefront package or contain "STF" in their name
	return strings.Contains(typeName, "citrixstorefront") || strings.Contains(typeName, "STF")
}

// isGetRequestType checks if the request type name indicates a GET/read operation
// based on the autogenerated naming convention: Api<Service><Method><Function>Request
// NOTE: StoreFront operations are excluded because they should not use ExecuteWithRetry
func isGetRequestType(typeName string) bool {
	// Exclude StoreFront operations - they should use direct Execute()
	// Check for citrixstorefront package (models and apis), or STF prefix in type names
	lowerTypeName := strings.ToLower(typeName)
	if strings.Contains(lowerTypeName, "citrixstorefront") ||
		strings.Contains(lowerTypeName, "/apis.") ||
		strings.Contains(typeName, "STF") {
		return false
	}

	// Must be a Request type
	if !strings.Contains(typeName, "Request") {
		return false
	}

	// CVAD/DaaS autogenerated patterns - check for GET method keywords in the type name
	// Pattern: Api<Service><Method>Request where Method contains Get/Fetch/List/Query/Read/Retrieve
	getMethodKeywords := []string{
		"Get",
		"Fetch",
		"List",
		"Query",
		"Read",
		"Retrieve",
	}

	// Check if the type name starts with or contains Api, and contains a GET method keyword
	if strings.Contains(typeName, "Api") {
		for _, keyword := range getMethodKeywords {
			if strings.Contains(typeName, keyword) {
				return true
			}
		}
	}

	return false
}
