package httpapi

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

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

// look for a JWTMiddlewareForQuery middleware
func hasJWTMiddleware(callExpr *ast.CallExpr) bool {
	for _, node := range callExpr.Args[2:] {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}
		// restrict to <ct>.JWTMiddlewareForQuery
		selector, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if selector.Sel.Name == "JWTMiddlewareForQuery" {
			return true
		}
	}
	return false
}

// echoExtractor scans a file using the Echo framework, looking for method calls .GET .POST .PUT .DELETE
// inside all top level functions in `f`, and parameters bindings.
func (ex echoExtractor) extract(pkg *packages.Package, fi *ast.File) []Endpoint {
	comments := map[int]string{} // line to comment
	for _, cm := range fi.Comments {
		if len(cm.List) != 1 {
			continue
		}
		line := pkg.Fset.Position(cm.Pos()).Line
		comments[line] = strings.TrimSpace(cm.List[0].Text[2:])
	}

	var out []Endpoint

	ast.Inspect(fi, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		line := pkg.Fset.Position(callExpr.Pos()).Line
		comment := newSpecialComment(comments[line])

		if comment == ignore {
			log.Println("ignoring route at line", line)
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
		isUrlOnly := comment == urlOnly

		urlNode, handlerNode := callExpr.Args[0], callExpr.Args[1]
		path, err := resolveConstString(urlNode, pkg)
		if err != nil {
			panic("invalid endpoint URL :" + err.Error())
		}

		body, name, sourcePkg := parseEndpointFunc(handlerNode, pkg)
		contract := newContractFromEchoBody(sourcePkg, body, name)

		if hasJWTMiddleware(callExpr) {
			// add a token=<string> query parameters
			contract.InputQueryParams = append(contract.InputQueryParams, TypedParam{
				type_: types.Typ[types.String],
				Name:  "token",
			})
		}

		out = append(out, Endpoint{Url: path, Method: methodName, IsUrlOnly: isUrlOnly, Contract: contract})

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
					out.returnT = types.NewSlice(types.Typ[types.Byte])
					out.IsReturnBlob = true
				}
			} else if method.Sel.Name == "StreamJSON" && len(call.Args) == 2 {
				// utils.StreamJSON(c.Response(), it)
				iterType := resolveVarType(call.Args[1].(*ast.Ident), pkg)
				out.returnT = extractIter2Value(iterType)
				out.IsReturnStream = true
			}
		}
	}
}

// expect iter.Seq2[T, _] and return T
func extractIter2Value(ty types.Type) types.Type {
	fnType := ty.Underlying().(*types.Signature)
	if fnType.Params().Len() != 1 {
		panic("expected iter.Seq2")
	}
	tParams := ty.(*types.Named).TypeArgs()
	if tParams.Len() != 2 {
		panic("expected iter.Seq2")
	}
	return tParams.At(0)
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
		if queryParam, ty := parseCallWithString(rh, "QueryParam", pkg); queryParam != "" {
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: ty})
		}
		if queryParam, ty := parseCallWithString(rh, "QueryParamBool", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: ty})
		}
		if queryParam, ty := parseCallWithString(rh, "QueryParamInt", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: ty})
		}
		if queryParam, ty := parseCallWithString(rh, "QueryParamInt64", pkg); queryParam != "" { // special converter
			out.InputQueryParams = append(out.InputQueryParams, TypedParam{Name: queryParam, type_: ty})
		}
		if formValue, _ := parseCallWithString(rh, "FormValue", pkg); formValue != "" {
			out.InputForm.ValueNames = append(out.InputForm.ValueNames, formValue)
		}
		if formFile, _ := parseCallWithString(rh, "FormFile", pkg); formFile != "" {
			out.InputForm.File = formFile
		}
		if formField, ty := parseFormValueJSON(rh, pkg); formField != "" {
			out.InputForm.JSON = TypedParam{Name: formField, type_: ty}
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

func parseFormValueJSON(expr ast.Expr, pkg *packages.Package) (string, types.Type) {
	if call, ok := expr.(*ast.CallExpr); ok {
		function := call.Fun

		var name string
		switch caller := function.(type) {
		case *ast.SelectorExpr:
			name = caller.Sel.Name
		case *ast.Ident:
			name = caller.Name
		default:
			return "", nil
		}

		if name != "FormValueJSON" {
			return "", nil
		}

		// "c.<methodName>(<string>)"
		if len(call.Args) != 3 {
			panic("invalid argument length for FormValueJSON")
		}
		arg := call.Args[1]
		dst := call.Args[2]

		argS, err := resolveConstString(arg, pkg)
		if err != nil {
			panic(fmt.Sprintf("invalid first argument for FormValueJSON: %s", err))
		}

		outType := pkg.TypesInfo.TypeOf(dst)
		ptr, ok := outType.(*types.Pointer)
		if !ok {
			panic("expected pointer in FormValueJSON")
		}
		return argS, ptr.Elem()
	}
	return "", nil
}

func parseCallWithString(expr ast.Expr, methodName string, pkg *packages.Package) (string, types.Type) {
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
			return "", nil
		}

		if name != methodName {
			return "", nil
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

		outType := pkg.TypesInfo.TypeOf(expr)
		if tuple, ok := outType.(*types.Tuple); ok {
			outType = tuple.At(0).Type()
		}

		return argS, outType
	}
	return "", nil
}
