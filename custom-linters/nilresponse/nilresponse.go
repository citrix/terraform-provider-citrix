// Copyright © 2026. Citrix Systems, Inc.

package nilresponse

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = &analysis.Analyzer{
	Name: "nilresponse",
	Doc:  "Checks that an *http.Response returned alongside an error is only used when it cannot be nil",
	Run:  run,
}

// trustedPackages hold the generated clients and their retry wrappers. Tests add their fixture package.
var trustedPackages = []string{"github.com/citrix/citrix-daas-rest-go"}

// trustedNames are the calls in trustedPackages that return the server's response, which is non-nil whenever
// the error is nil: request Execute methods, generated XxxExecute service methods (matched by suffix), and the
// retry wrappers. Helpers that only pass an operation's response through (RetryOperationWithExponentialBackOff)
// are not trusted. Any other call returning (..., *http.Response, error) needs an explicit nil check.
var trustedNames = map[string]bool{
	"ExecuteWithRetry":            true,
	"ExecuteWithRetryContext":     true,
	"ExecuteWithRetryOnNotFound":  true,
	"GetAllPagesWithRetry":        true,
	"GetAllPagesWithRetryContext": true,
}

// exitCalls end the current function or test when they are the last statement of a branch.
var exitCalls = map[string]bool{
	"Fatal": true, "Fatalf": true, "Fatalln": true, "FailNow": true, "SkipNow": true, "Skip": true, "Skipf": true,
	"Panic": true, "Panicf": true, "Panicln": true, "Exit": true, "Goexit": true,
}

// assignment is a statement that sets an *http.Response variable from a call that also returns an error.
type assignment struct {
	end     token.Pos
	errVar  *types.Var
	trusted bool
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			checkFunc(pass, fn.Body)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, body *ast.BlockStmt) {
	assignments, errWrites := collectAssignments(pass, body)
	if len(assignments) == 0 {
		return
	}

	var stack []ast.Node
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)

		var ident *ast.Ident
		var use string
		switch e := n.(type) {
		case *ast.SelectorExpr:
			ident, _ = e.X.(*ast.Ident)
			if ident != nil {
				use = ident.Name + "." + e.Sel.Name
			}
		case *ast.StarExpr:
			ident, _ = e.X.(*ast.Ident)
			if ident != nil {
				use = "*" + ident.Name
			}
		}
		if ident == nil {
			return true
		}
		respVar, ok := pass.TypesInfo.Uses[ident].(*types.Var)
		if !ok {
			return true
		}
		assign, ok := latestAssignment(assignments[respVar], n.Pos())
		if !ok {
			return true
		}
		g := guard{pass: pass, respVar: respVar, assign: assign, errWrites: errWrites[assign.errVar]}
		if !g.isGuarded(stack) {
			pass.Reportf(n.Pos(), "%s uses a response that is nil when the call fails without one. Check err == nil first (trusted client calls only), guard with %s != nil, or use util.IsNotFound(%s)",
				use, ident.Name, ident.Name)
		}
		return true
	})
}

// collectAssignments records every `..., resp, ..., err := call()` (or `=`) where resp is an *http.Response,
// and the end position of every write to an error variable. Responses that do not come with an error
// (test helpers, literals, parameters) are not tracked.
func collectAssignments(pass *analysis.Pass, body *ast.BlockStmt) (map[*types.Var][]assignment, map[*types.Var][]token.Pos) {
	assignments := map[*types.Var][]assignment{}
	errWrites := map[*types.Var][]token.Pos{}
	errorType := types.Universe.Lookup("error").Type()

	ast.Inspect(body, func(n ast.Node) bool {
		stmt, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		var errVar *types.Var
		var responses []*types.Var
		for _, lhs := range stmt.Lhs {
			v := lhsVar(pass, lhs)
			if v == nil {
				continue
			}
			switch {
			case types.Identical(v.Type(), errorType):
				errVar = v
				errWrites[v] = append(errWrites[v], stmt.End())
			case isHTTPResponsePointer(v.Type()):
				responses = append(responses, v)
			}
		}

		if errVar == nil || len(stmt.Rhs) != 1 {
			return true
		}
		call, isCall := ast.Unparen(stmt.Rhs[0]).(*ast.CallExpr)
		if !isCall {
			return true
		}
		for _, v := range responses {
			assignments[v] = append(assignments[v], assignment{end: stmt.End(), errVar: errVar, trusted: isTrustedCall(pass, call)})
		}
		return true
	})
	return assignments, errWrites
}

// isTrustedCall reports whether the call returns a non-nil response whenever its error is nil: a net/http
// function such as Client.Do, or a generated client call or retry wrapper from trustedPackages. The callee is
// resolved through the type information, so generic calls such as ExecuteWithRetry[*T] are matched too.
func isTrustedCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok || fn.Pkg() == nil {
		return false
	}
	path := fn.Pkg().Path()
	if path == "net/http" {
		return true
	}
	if !inTrustedPackage(path) {
		return false
	}
	name := fn.Name()
	return trustedNames[name] || strings.HasSuffix(name, "Execute")
}

func inTrustedPackage(path string) bool {
	for _, p := range trustedPackages {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

func lhsVar(pass *analysis.Pass, expr ast.Expr) *types.Var {
	ident, ok := expr.(*ast.Ident)
	if !ok || ident.Name == "_" {
		return nil
	}
	obj := pass.TypesInfo.Defs[ident]
	if obj == nil {
		obj = pass.TypesInfo.Uses[ident]
	}
	v, _ := obj.(*types.Var)
	return v
}

func isHTTPResponsePointer(t types.Type) bool {
	ptr, ok := types.Unalias(t).(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := types.Unalias(ptr.Elem()).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "net/http" && obj.Name() == "Response"
}

// latestAssignment returns the last assignment that finished before the use. Using the end position means
// a use inside the right-hand side of a reassignment is matched to the previous assignment.
func latestAssignment(candidates []assignment, use token.Pos) (assignment, bool) {
	var latest assignment
	found := false
	for _, a := range candidates {
		if a.end <= use && (!found || a.end > latest.end) {
			latest = a
			found = true
		}
	}
	return latest, found
}

type guard struct {
	pass      *analysis.Pass
	respVar   *types.Var
	assign    assignment
	errWrites []token.Pos
}

// isGuarded walks from the use out to the function body and reports whether the response is known to be
// non-nil at that point.
func (g guard) isGuarded(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		child := stack[i+1]
		switch parent := stack[i].(type) {
		case *ast.BinaryExpr:
			if child != parent.Y {
				continue
			}
			if parent.Op == token.LAND && g.implies(parent.X, true) {
				return true
			}
			if parent.Op == token.LOR && g.implies(parent.X, false) {
				return true
			}
		case *ast.IfStmt:
			if child == parent.Body && g.implies(parent.Cond, true) {
				return true
			}
			if parent.Else != nil && child == parent.Else && g.implies(parent.Cond, false) {
				return true
			}
		case *ast.BlockStmt:
			if g.earlierExit(parent.List, child) {
				return true
			}
		case *ast.CaseClause:
			if g.earlierExit(parent.Body, child) {
				return true
			}
		case *ast.CommClause:
			if g.earlierExit(parent.Body, child) {
				return true
			}
		}
	}
	return false
}

// respGuardValid reports whether a nil check on the response at pos happens after its latest assignment.
func (g guard) respGuardValid(pos token.Pos) bool {
	return pos > g.assign.end
}

// errGuardValid reports whether an error check at pos proves the response non-nil: the call must be a
// trusted call, the check must come after it, and err must not have been overwritten in between.
func (g guard) errGuardValid(pos token.Pos) bool {
	if g.assign.errVar == nil || !g.assign.trusted || pos <= g.assign.end {
		return false
	}
	for _, w := range g.errWrites {
		if w > g.assign.end && w < pos {
			return false
		}
	}
	return true
}

// implies reports whether cond evaluating to `value` guarantees the response is non-nil.
// True branch: some && conjunct is `resp != nil` or `err == nil`.
// False branch: some || disjunct is `resp == nil` or `err != nil`.
func (g guard) implies(cond ast.Expr, value bool) bool {
	pos := cond.Pos()
	respOk, errOk := g.respGuardValid(pos), g.errGuardValid(pos)
	if value {
		for _, c := range flatten(cond, token.LAND) {
			if (respOk && g.compares(c, g.respVar, token.NEQ)) || (errOk && (g.compares(c, g.assign.errVar, token.EQL) || g.isDeleteCheck(c))) {
				return true
			}
		}
		return false
	}
	for _, d := range flatten(cond, token.LOR) {
		if (respOk && g.compares(d, g.respVar, token.EQL)) || (errOk && (g.compares(d, g.assign.errVar, token.NEQ) || g.isNegatedDeleteCheck(d))) {
			return true
		}
	}
	return false
}

// isDeleteCheck reports whether expr is util.CheckDeleteResponse(..., err, ...), which is true only when err is nil.
func (g guard) isDeleteCheck(expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn, ok := typeutil.Callee(g.pass.TypesInfo, call).(*types.Func)
	return ok && fn.Name() == "CheckDeleteResponse" && g.anyArgIs(call, g.assign.errVar)
}

// isNegatedDeleteCheck reports whether expr is !util.CheckDeleteResponse(..., err, ...), which is true when err is set.
func (g guard) isNegatedDeleteCheck(expr ast.Expr) bool {
	not, ok := ast.Unparen(expr).(*ast.UnaryExpr)
	return ok && not.Op == token.NOT && g.isDeleteCheck(not.X)
}

// earlierExit reports whether a statement before child leaves the block when the response is nil, for
// example `if err != nil { return }`, `if resp == nil { t.Fatal(...) }` or `require.NotNil(t, resp)`.
func (g guard) earlierExit(stmts []ast.Stmt, child ast.Node) bool {
	for _, stmt := range stmts {
		if stmt == child || stmt.Pos() >= child.Pos() {
			return false
		}
		if ifStmt, ok := stmt.(*ast.IfStmt); ok && terminates(ifStmt.Body) && g.implies(ifStmt.Cond, false) {
			return true
		}
		if g.isRequireGuard(stmt) {
			return true
		}
	}
	return false
}

// isRequireGuard reports whether stmt is a testify require call (`require.NotNil(t, resp)`,
// `require.NoError(t, err)`, `r.NoError(err)` from require.New, `s.Require().NoError(err)`) that stops the
// test before the response can be used as nil.
func (g guard) isRequireGuard(stmt ast.Stmt) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !g.isRequireCall(sel) {
		return false
	}
	pos := stmt.Pos()
	switch sel.Sel.Name {
	case "NotNil", "NotNilf":
		return g.respGuardValid(pos) && g.anyArgIs(call, g.respVar)
	case "NoError", "NoErrorf":
		return g.errGuardValid(pos) && g.anyArgIs(call, g.assign.errVar)
	}
	return false
}

// isRequireCall matches testify's require package and its Assertions type by package path.
func (g guard) isRequireCall(sel *ast.SelectorExpr) bool {
	fn, ok := g.pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	return ok && fn.Pkg() != nil && strings.HasSuffix(fn.Pkg().Path(), "testify/require")
}

func (g guard) anyArgIs(call *ast.CallExpr, v *types.Var) bool {
	for _, arg := range call.Args {
		if g.isVar(arg, v) {
			return true
		}
	}
	return false
}

// compares reports whether expr is `v <op> nil` or `nil <op> v`.
func (g guard) compares(expr ast.Expr, v *types.Var, op token.Token) bool {
	if v == nil {
		return false
	}
	bin, ok := ast.Unparen(expr).(*ast.BinaryExpr)
	if !ok || bin.Op != op {
		return false
	}
	return (g.isVar(bin.X, v) && g.isNil(bin.Y)) || (g.isVar(bin.Y, v) && g.isNil(bin.X))
}

func (g guard) isVar(expr ast.Expr, v *types.Var) bool {
	ident, ok := ast.Unparen(expr).(*ast.Ident)
	return ok && v != nil && g.pass.TypesInfo.Uses[ident] == v
}

func (g guard) isNil(expr ast.Expr) bool {
	ident, ok := ast.Unparen(expr).(*ast.Ident)
	if !ok {
		return false
	}
	_, isNil := g.pass.TypesInfo.Uses[ident].(*types.Nil)
	return isNil
}

func flatten(expr ast.Expr, op token.Token) []ast.Expr {
	if bin, ok := ast.Unparen(expr).(*ast.BinaryExpr); ok && bin.Op == op {
		return append(flatten(bin.X, op), flatten(bin.Y, op)...)
	}
	return []ast.Expr{expr}
}

// terminates reports whether a branch always leaves: it ends in return, break, continue, goto, panic, or a
// call such as t.Fatal, t.FailNow, log.Fatal or os.Exit.
func terminates(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	case *ast.ExprStmt:
		call, ok := last.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			return fun.Name == "panic"
		case *ast.SelectorExpr:
			return exitCalls[fun.Sel.Name]
		}
	}
	return false
}
