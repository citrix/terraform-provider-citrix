// Copyright © 2026. Citrix Systems, Inc.

package unknowncheck

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/citrix/terraform-provider-citrix/custom-linters/util"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "unknowncheck",
	Doc:  "Checks that IsUnknown() is called before IsNull() in ValidateConfig and ModifyPlan functions to properly handle Terraform properties",
	Run:  run,
}

// Terraform functions that require IsUnknown checks before IsNull
var targetFunctions = map[string]bool{
	"ValidateConfig": true,
	"ModifyPlan":     true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Only check ValidateConfig and ModifyPlan functions
			if !isTargetFunction(funcDecl) {
				return true
			}

			// Track properties that have been checked with IsUnknown
			checkedUnknown := make(map[string]bool)

			// Inspect function body for IsNull calls
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				// Look for if statements
				ifStmt, ok := n.(*ast.IfStmt)
				if !ok {
					return true
				}

				// Check the condition
				checkCondition(pass, funcDecl, ifStmt.Cond, checkedUnknown)

				return true
			})

			return true
		})
	}
	return nil, nil
}

func isTargetFunction(funcDecl *ast.FuncDecl) bool {
	return util.IsTerraformMethod(funcDecl, targetFunctions)
}

func checkCondition(pass *analysis.Pass, funcDecl *ast.FuncDecl, expr ast.Expr, checkedUnknown map[string]bool) {
	checkConditionInternal(pass, funcDecl, expr, checkedUnknown, false)
}

func checkConditionInternal(pass *analysis.Pass, funcDecl *ast.FuncDecl, expr ast.Expr, checkedUnknown map[string]bool, negated bool) {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			// When we see a NOT, flip the negation flag
			checkConditionInternal(pass, funcDecl, e.X, checkedUnknown, !negated)
		}
	case *ast.BinaryExpr:
		// Handle && and || operators
		switch e.Op { //nolint:exhaustive // Only LAND and LOR operators are relevant for IsUnknown/IsNull analysis
		case token.LAND:
			// For &&, check left side first as it's evaluated first
			// If left side checks IsUnknown, mark it
			markUnknownChecks(e.X, checkedUnknown)
			// Also mark IsUnknown checks from right side for AND conditions
			markUnknownChecks(e.Y, checkedUnknown)
			checkConditionInternal(pass, funcDecl, e.X, checkedUnknown, negated)
			checkConditionInternal(pass, funcDecl, e.Y, checkedUnknown, negated)
		case token.LOR:
			// For ||, both sides should be checked independently
			leftChecked := make(map[string]bool)
			rightChecked := make(map[string]bool)
			for k, v := range checkedUnknown {
				leftChecked[k] = v
				rightChecked[k] = v
			}
			checkConditionInternal(pass, funcDecl, e.X, leftChecked, negated)
			checkConditionInternal(pass, funcDecl, e.Y, rightChecked, negated)
		default:
			// Other binary operators don't affect our analysis
		}
	case *ast.CallExpr:
		// Check if this is an IsNull() call
		// Only flag it if it's NOT negated (i.e., we're checking if field.IsNull() is true)
		// We don't need to check IsUnknown before !field.IsNull() since that checks if field has a value
		if isIsNullCall(e) && !negated {
			propertyName := getPropertyName(e)
			if propertyName != "" && !checkedUnknown[propertyName] {
				pass.Reportf(e.Pos(), "IsNull() called on %s without checking IsUnknown() first in %s. Properties must be checked with IsUnknown() before IsNull() to handle Terraform variables correctly", propertyName, funcDecl.Name.Name)
			}
		}
		// If this is an IsUnknown() call, mark the property as checked
		if isIsUnknownCall(e) {
			propertyName := getPropertyName(e)
			if propertyName != "" {
				checkedUnknown[propertyName] = true
			}
		}
	}
}

func markUnknownChecks(expr ast.Expr, checkedUnknown map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if isIsUnknownCall(call) {
				propertyName := getPropertyName(call)
				if propertyName != "" {
					checkedUnknown[propertyName] = true
				}
			}
		}
		return true
	})
}

func isIsNullCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "IsNull"
}

func isIsUnknownCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "IsUnknown"
}

func getPropertyName(call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	// Build the full property path (e.g., "data.Id" or "plan.Name")
	path := buildExprPath(sel.X)

	// Skip Raw fields - they don't need IsUnknown checks
	// Raw represents the entire request object and is used to check if a plan/state/config exists
	if isRawField(path) {
		// Return empty string to skip this check
		return ""
	}

	return path
}

func isRawField(path string) bool {
	// Check if the path ends with ".Raw" (e.g., "req.Plan.Raw", "req.State.Raw", "req.Config.Raw")
	// These are special fields representing the entire request/plan/state/config object
	// and don't need IsUnknown() checks before IsNull()
	return strings.HasSuffix(path, ".Raw")
}

func buildExprPath(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		base := buildExprPath(e.X)
		if base != "" {
			return base + "." + e.Sel.Name
		}
		return e.Sel.Name
	default:
		return ""
	}
}
