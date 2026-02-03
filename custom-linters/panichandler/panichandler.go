// Copyright © 2026. Citrix Systems, Inc.

package panichandler

import (
	"go/ast"

	"github.com/citrix/terraform-provider-citrix/custom-linters/util"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "panichandler",
	Doc:  "Checks that Terraform provider SDK interface functions start with defer util.PanicHandler(&resp.Diagnostics)",
	Run:  run,
}

// Terraform resource and datasource interface methods that require panic handlers
var requiredFunctions = map[string]bool{
	"Create":         true,
	"Read":           true,
	"Update":         true,
	"Delete":         true,
	"ModifyPlan":     true,
	"ValidateConfig": true,
	// Explicitly excluded methods
	"ImportState": false,
	"Metadata":    false,
	"Configure":   false,
	"Schema":      false,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !util.IsTerraformMethod(funcDecl, requiredFunctions) {
				return true
			}

			if !hasPanicHandlerDefer(funcDecl) {
				pass.Reportf(funcDecl.Pos(), "function %s must start with defer util.PanicHandler(...)", funcDecl.Name.Name)
			}

			return true
		})
	}
	return nil, nil
}

func hasPanicHandlerDefer(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
		return false
	}

	firstStmt := funcDecl.Body.List[0]
	deferStmt, ok := firstStmt.(*ast.DeferStmt)
	if !ok {
		return false
	}

	callExpr, ok := deferStmt.Call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	utilIdent, ok := callExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Accept both util.PanicHandler (production code) and testutil.PanicHandler (test code)
	if (utilIdent.Name != "util" && utilIdent.Name != "testutil") || callExpr.Sel.Name != "PanicHandler" {
		return false
	}

	return true
}
