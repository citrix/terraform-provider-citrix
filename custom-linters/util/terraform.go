// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"go/ast"
	"strings"
)

// IsTerraformMethod checks if a function declaration is a Terraform SDK interface method
// by verifying it has a receiver (is a method), matches one of the known method names,
// and has the correct Terraform SDK signature (3 parameters with req and resp parameters).
// This is a common check across custom linters as Terraform SDK methods are always
// methods on a struct with a specific signature, not standalone functions or helper methods.
func IsTerraformMethod(funcDecl *ast.FuncDecl, methodNames map[string]bool) bool {
	// Must be a method with a receiver
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	// Check if function name is in the list of target methods
	if !methodNames[funcDecl.Name.Name] {
		return false
	}

	// Terraform SDK interface methods have exactly 3 parameters: (ctx, req, resp)
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 3 {
		return false
	}

	// Check that the second parameter is a Request type and third is a Response pointer type
	// Terraform methods have signatures like:
	// - func (r *T) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
	// - func (d *T) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse)
	params := funcDecl.Type.Params.List

	// Check second parameter (req) type name contains "Request"
	reqTypeName := getTypeName(params[1].Type)
	if !strings.Contains(reqTypeName, "Request") {
		return false
	}

	// Check third parameter (resp) is a pointer type containing "Response"
	respTypeName := getTypeName(params[2].Type)
	return strings.Contains(respTypeName, "Response")
}

// getTypeName extracts the type name from an ast.Expr, handling both simple and pointer types
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		// Pointer type like *resource.CreateResponse
		return getTypeName(t.X)
	case *ast.SelectorExpr:
		// Qualified type like resource.CreateRequest
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	}
	return ""
}
