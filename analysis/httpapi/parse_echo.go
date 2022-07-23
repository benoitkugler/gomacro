package httpapi

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// implements a parser for the echo framework

func isHttpMethod(name string) bool {
	switch name {
	case "GET", "PUT", "POST", "DELETE":
		return true
	default:
		return false
	}
}

// Parse looks for method calls .GET .POST .PUT .DELETE
// inside all top level functions in `f`.
func Parse(pkg *packages.Package, fi *ast.File) []Endpoint {
	var out []Endpoint

	ast.Inspect(fi, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// restrict to <echo>.{GET} methods
		selector, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		methodName := selector.Sel.Name
		if !isHttpMethod(methodName) || len(callExpr.Args) < 2 {
			// we are looking for .<METHOD>(url, handler)
			return true
		}

		path := parseEndpointURL(callExpr.Args[0], pkg)
		body, name, sourcePkg := parseEndpointFunc(callExpr.Args[1], pkg)
		contract := newContractFromEchoBody(sourcePkg, body, name)
		out = append(out, Endpoint{Url: path, Method: methodName, Contract: contract})

		return false
	})

	return out
}

// Look for Bind(), QueryParam() and JSON() method calls
// Some custom parsing method are also supported :
// 	- .BindNoId -> expects a type without id field
//	- .QueryParamBool(c, ...) -> convert string to boolean
//	- .QueryParamInt64(, ...) -> convert string to int64
// pkg is the package where the function is defined
func newContractFromEchoBody(pkg *types.Info, body *ast.BlockStmt, contractName string) Contract {
	out := Contract{Name: contractName}

	// func analyzeHandler(body []ast.Stmt, pkg *types.Package) gents.Contrat {
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// we detect return value type with c.JSON calls,
		// and input values through assignments
		switch stmt := n.(type) {
		case *ast.ReturnStmt:
			if len(stmt.Results) != 1 { // should not happend : the method return error
				panic("expected at least one result in return statement")
			}
			if call, ok := stmt.Results[0].(*ast.CallExpr); ok {
				if method, ok := call.Fun.(*ast.SelectorExpr); ok {
					if method.Sel.Name == "JSON" || method.Sel.Name == "JSONPretty" {
						if len(call.Args) >= 2 {
							output := call.Args[1] // c.JSON(200, output)
							switch output := output.(type) {
							case *ast.Ident:
								out.returnT = resolveVarType(output, pkg)
							case *ast.CompositeLit:
								out.returnT = parseCompositeLit(output, pkg)
							default:
								panic("unsupported return value")
							}
						}
					}
				}
			}
			return false

			// case *ast.AssignStmt: // TODO:
			// 	parseAssignments(stmt.Rhs, pkg, &out)
		}
		return true
	})

	return out
}

func resolveVarType(arg *ast.Ident, pkg *types.Info) types.Type {
	return resolveIdentifier(arg, pkg).Type()
}

func parseCompositeLit(lit *ast.CompositeLit, pkg *types.Info) types.Type {
	return pkg.Types[lit.Type].Type
}

// func parseAssignments(rhs []ast.Expr, pkg *types.Package, out *gents.Contrat) {
// 	for _, rh := range rhs {
// 		if typeIn := parseBindCall(rh, pkg); typeIn.Type != nil {
// 			out.Input = typeIn
// 		}
// 		if queryParam := parseCallWithString(rh, "QueryParam"); queryParam != "" {
// 			out.QueryParams = append(out.QueryParams, gents.TypedParam{Name: queryParam, Type: tstypes.TsString})
// 		}
// 		if queryParam := parseCallWithString(rh, "QueryParamBool"); queryParam != "" { // special converter
// 			out.QueryParams = append(out.QueryParams, gents.TypedParam{Name: queryParam, Type: tstypes.TsBoolean})
// 		}
// 		if queryParam := parseCallWithString(rh, "QueryParamInt64"); queryParam != "" { // special converter
// 			out.QueryParams = append(out.QueryParams, gents.TypedParam{Name: queryParam, Type: tstypes.TsNumber})
// 		}
// 		if formValue := parseCallWithString(rh, "FormValue"); formValue != "" {
// 			out.Form.Values = append(out.Form.Values, formValue)
// 		}
// 		if formFile := parseCallWithString(rh, "FormFile"); formFile != "" {
// 			out.Form.File = formFile
// 		}
// 	}
// }

// // TODO: support New<T> types
// func resolveBindTarget(arg ast.Expr, pkg *types.Package) types.Type {
// 	switch arg := arg.(type) {
// 	case *ast.Ident: // c.Bind(pointer)
// 		return resolveLocalType(arg, pkg)
// 	case *ast.UnaryExpr: // c.Bind(&value)
// 		if ident, ok := arg.X.(*ast.Ident); arg.Op == token.AND && ok {
// 			return resolveLocalType(ident, pkg)
// 		}
// 	}
// 	return nil
// }

// func parseBindCall(expr ast.Expr, pkg *types.Package) gents.TypeNoId {
// 	if call, ok := expr.(*ast.CallExpr); ok {
// 		switch caller := call.Fun.(type) {
// 		case *ast.SelectorExpr:
// 			if caller.Sel.Name == "Bind" && len(call.Args) == 1 { // "c.Bind(in)"
// 				typ := resolveBindTarget(call.Args[0], pkg)
// 				return gents.TypeNoId{Type: typ}
// 			}
// 		case *ast.Ident:
// 			if caller.Name == "BindNoId" && len(call.Args) == 2 { // BindNoId(c, in)
// 				typ := resolveBindTarget(call.Args[1], pkg)
// 				return gents.TypeNoId{Type: typ, NoId: true}
// 			}
// 		}
// 	}
// 	return gents.TypeNoId{}
// }

// func parseCallWithString(expr ast.Expr, methodName string) string {
// 	if call, ok := expr.(*ast.CallExpr); ok {
// 		var name string
// 		switch caller := call.Fun.(type) {
// 		case *ast.SelectorExpr:
// 			name = caller.Sel.Name
// 		case *ast.Ident:
// 			name = caller.Name
// 		default:
// 			return ""
// 		}

// 		if name != methodName {
// 			return ""
// 		}

// 		var arg ast.Expr
// 		if len(call.Args) == 1 { // "c.<methodName>(<string>)"
// 			arg = call.Args[0]
// 		} else if len(call.Args) == 2 { // "ct.<methodName>(c, <string>)" or <functionName>(c, <string>)
// 			arg = call.Args[1]
// 		}

// 		if lit, ok := arg.(*ast.BasicLit); ok {
// 			return stringLitteral(lit)
// 		}
// 	}
// 	return ""
// }
