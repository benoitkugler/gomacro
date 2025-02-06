package httpapi

import (
	"fmt"
	"go/ast"
	"go/token"
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

// echoExtractor scans a file using the Echo framework, looking for method calls .GET .POST .PUT .DELETE
// inside all top level functions in `f`, and parameters bindings.
func echoExtractor(pkg *packages.Package, fi *ast.File) []Endpoint {
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

		path, err := resolveConstString(callExpr.Args[0], pkg)
		if err != nil {
			panic("invalid endpoint URL :" + err.Error())
		}

		body, name, sourcePkg := parseEndpointFunc(callExpr.Args[1], pkg)
		contract := newContractFromEchoBody(sourcePkg, body, name)
		out = append(out, Endpoint{Url: path, Method: methodName, Contract: contract})

		return false
	})

	return out
}

// Look for Bind(), QueryParam(), FormValue(), FromFile() and JSON() method calls
// Some custom parsing method are also supported :
//   - .QueryParamBool(c, ...) -> convert string to boolean
//   - .QueryParamInt64(, ...) -> convert string to int64
//
// pkg is the package where the function is defined
func newContractFromEchoBody(pkg *packages.Package, body *ast.BlockStmt, contractName string) Contract {
	out := Contract{Name: contractName}

	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// we detect return value type with c.JSON calls,
		// and input values through assignments
		switch stmt := n.(type) {
		case *ast.ReturnStmt:
			parseReturnStmt(stmt, pkg.TypesInfo, &out)
			return false
		case *ast.AssignStmt:
			parseAssignments(stmt.Rhs, pkg, &out)
			return false
		}
		return true
	})

	return out
}

func parseReturnStmt(stmt *ast.ReturnStmt, pkg *types.Info, out *Contract) {
	if len(stmt.Results) != 1 { // should not happend : the method return error
		panic("expected at least one result in return statement")
	}
	if call, ok := stmt.Results[0].(*ast.CallExpr); ok {
		if method, ok := call.Fun.(*ast.SelectorExpr); ok {
			if method.Sel.Name == "JSON" || method.Sel.Name == "JSONPretty" {
				if len(call.Args) >= 2 { // c.JSON(200, output)
					output := call.Args[1]
					switch output := output.(type) {
					case *ast.Ident:
						out.returnT = resolveVarType(output, pkg)
					case *ast.CompositeLit:
						out.returnT = parseCompositeLit(output, pkg)
					default:
						panic("unsupported return value")
					}
				}
			} else if method.Sel.Name == "Blob" {
				if len(call.Args) >= 3 { // c.Blob(200, name, bytes)
					output := call.Args[2]
					switch output := output.(type) {
					case *ast.Ident:
						out.returnT = resolveVarType(output, pkg)
						out.IsReturnBlob = true
					default:
						panic("unsupported return value in Blob")
					}
				}
			}
		}
	}
}

func resolveVarType(arg *ast.Ident, pkg *types.Info) types.Type {
	return resolveIdentifier(arg, pkg).Type()
}

func parseCompositeLit(lit *ast.CompositeLit, pkg *types.Info) types.Type {
	return pkg.Types[lit.Type].Type
}

func parseAssignments(rhs []ast.Expr, pkg *packages.Package, out *Contract) {
	for _, rh := range rhs {
		if typeIn := tryParseBindCall(rh, pkg.TypesInfo); typeIn != nil {
			out.inputT = typeIn
		}
		if queryParam := parseCallWithString(rh, "QueryParam", pkg); queryParam != "" {
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: types.Typ[types.String]})
		}
		if queryParam := parseCallWithString(rh, "QueryParamBool", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: types.Typ[types.Bool]})
		}
		if queryParam := parseCallWithString(rh, "QueryParamInt", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: types.Typ[types.Int]})
		}
		if queryParam := parseCallWithString(rh, "QueryParamInt64", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: types.Typ[types.Int]})
		}
		if formValue := parseCallWithString(rh, "FormValue", pkg); formValue != "" {
			out.InputForm.ValueNames = append(out.InputForm.ValueNames, formValue)
		}
		if formFile := parseCallWithString(rh, "FormFile", pkg); formFile != "" {
			out.InputForm.File = formFile
		}
	}
}

func resolveBindTarget(arg ast.Expr, pkg *types.Info) types.Type {
	switch arg := arg.(type) {
	case *ast.Ident: // c.Bind(pointer)
		return resolveVarType(arg, pkg)
	case *ast.UnaryExpr: // c.Bind(&value)
		if ident, ok := arg.X.(*ast.Ident); arg.Op == token.AND && ok {
			return resolveVarType(ident, pkg)
		}
	}
	panic("unsupported Bind() expression")
}

// parse a c.Bind(&params) call
func tryParseBindCall(expr ast.Expr, pkg *types.Info) types.Type {
	if call, ok := expr.(*ast.CallExpr); ok {
		switch caller := call.Fun.(type) {
		case *ast.SelectorExpr:
			if caller.Sel.Name == "Bind" && len(call.Args) == 1 { // "c.Bind(&params)"
				return resolveBindTarget(call.Args[0], pkg)
			}
		}
	}
	return nil
}

func parseCallWithString(expr ast.Expr, methodName string, pkg *packages.Package) string {
	if call, ok := expr.(*ast.CallExpr); ok {
		function := call.Fun
		// generic support
		if index, isGeneric := function.(*ast.IndexExpr); isGeneric {
			function = index.X
		}

		var name string
		switch caller := function.(type) {
		case *ast.SelectorExpr:
			name = caller.Sel.Name
		case *ast.Ident:
			name = caller.Name
		default:
			return ""
		}

		if name != methodName {
			return ""
		}

		var arg ast.Expr
		if len(call.Args) == 1 { // "c.<methodName>(<string>)"
			arg = call.Args[0]
		} else if len(call.Args) == 2 { // "ct.<methodName>(c, <string>)" or <functionName>(c, <string>)
			arg = call.Args[1]
		}

		argS, err := resolveConstString(arg, pkg)
		if err != nil {
			panic(fmt.Sprintf("invalid %s argument: %s", methodName, err))
		}
		return argS
	}
	return ""
}
