// Copyright © 2026. Citrix Systems, Inc.

package testutil

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

// TestCase represents a single test case with expected violation count
type TestCase struct {
	FunctionName       string
	ExpectedViolations int
}

// RunAnalyzerTestCases runs an analyzer and validates violations per test case
func RunAnalyzerTestCases(t *testing.T, analyzer *analysis.Analyzer, pkgPath, filename string, testCases []TestCase) {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  "..",
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		t.Fatalf("failed to load package: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		t.Fatalf("package has errors: %v", pkg.Errors)
	}

	// Filter to only the specified file
	var filteredFiles []*ast.File
	for i, f := range pkg.Syntax {
		if filepath.Base(pkg.CompiledGoFiles[i]) == filename {
			filteredFiles = append(filteredFiles, f)
		}
	}

	if len(filteredFiles) == 0 {
		t.Fatalf("file %s not found in package", filename)
	}

	// Track violations by function
	violationsByFunc := make(map[string]int)

	// Run the analyzer
	pass := &analysis.Pass{
		Analyzer:   analyzer,
		Fset:       pkg.Fset,
		Files:      filteredFiles,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
	}

	pass.Report = func(d analysis.Diagnostic) {
		funcName := getFunctionName(pass.Fset, filteredFiles[0], d.Pos)
		violationsByFunc[funcName]++
		t.Logf("Violation in %s: %s", funcName, d.Message)
	}

	_, err = analyzer.Run(pass)
	if err != nil {
		t.Fatalf("analyzer failed: %v", err)
	}

	// Validate each test case
	for _, tc := range testCases {
		actual := violationsByFunc[tc.FunctionName]
		expected := tc.ExpectedViolations

		if actual != expected {
			t.Errorf("Function %s: expected %d violation(s), got %d", tc.FunctionName, expected, actual)
		}
	}

	// Check for unexpected violations in functions not listed in test cases
	expectedFuncs := make(map[string]bool)
	for _, tc := range testCases {
		expectedFuncs[tc.FunctionName] = true
	}

	for funcName, count := range violationsByFunc {
		if !expectedFuncs[funcName] {
			t.Errorf("Unexpected violations in function %s: %d violation(s)", funcName, count)
		}
	}
}

// getFunctionName finds the function name containing the given position
func getFunctionName(_ *token.FileSet, file *ast.File, pos token.Pos) string {
	var funcName string

	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Pos() <= pos && pos <= fn.End() {
				// Format as Receiver.Method or just Function
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					recvType := getReceiverType(fn.Recv.List[0].Type)
					funcName = fmt.Sprintf("%s.%s", recvType, fn.Name.Name)
				} else {
					funcName = fn.Name.Name
				}
				return false
			}
		}
		return true
	})

	return funcName
}

// getReceiverType extracts the receiver type name
func getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "Unknown"
}

// ExtractTestCasesFromComments parses test case expectations from comments in the test file.
// Looks for comments in the format: "violations:N" where N is the expected number of violations.
// Example:
//
//	// violations:0  <- expects no violations (valid code)
//	// violations:3  <- expects 3 violations (invalid code)
func ExtractTestCasesFromComments(t *testing.T, pkgPath, filename string) []TestCase {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  "..",
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		t.Fatalf("failed to load package: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		t.Fatalf("package has errors: %v", pkg.Errors)
	}

	var testCases []TestCase

	for i, f := range pkg.Syntax {
		if filepath.Base(pkg.CompiledGoFiles[i]) != filename {
			continue
		}

		// Extract test cases from comments
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Recv == nil || len(fn.Recv.List) == 0 {
					continue
				}

				recvType := getReceiverType(fn.Recv.List[0].Type)
				fullName := fmt.Sprintf("%s.%s", recvType, fn.Name.Name)

				// Check for test expectation in comments
				found := false
				if fn.Doc != nil {
					for _, comment := range fn.Doc.List {
						text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

						if strings.HasPrefix(text, "violations:") {
							var count int
							_, err := fmt.Sscanf(text, "violations:%d", &count)
							if err == nil {
								testCases = append(testCases, TestCase{
									FunctionName:       fullName,
									ExpectedViolations: count,
								})
								found = true
								break
							}
						}
					}
				}

				// Default to 0 violations if no annotation found
				if !found {
					testCases = append(testCases, TestCase{
						FunctionName:       fullName,
						ExpectedViolations: 0,
					})
				}
			}
		}
	}

	return testCases
}
