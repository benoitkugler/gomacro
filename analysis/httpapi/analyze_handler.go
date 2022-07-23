package httpapi

// import (
// 	"go/ast"
// 	"go/token"
// 	"go/types"

// 	"github.com/benoitkugler/structgen/api/gents"
// 	tstypes "github.com/benoitkugler/structgen/ts-types"
// )

// // Look for Bind(), QueryParam() and JSON() method calls
// // Some custom parsing method are also supported :
// // 	- .BindNoId -> expects a type without id field
// //	- .QueryParamBool(c, ...) -> convert string to boolean
// //	- .QueryParamInt64(, ...) -> convert string to int64
// // pkg is the package of the method
// func analyzeHandler(body []ast.Stmt, pkg *types.Package) gents.Contrat {
// 	var out gents.Contrat
// 	for _, stmt := range body {
// 		switch stmt := stmt.(type) {
// 		case *ast.ReturnStmt:
// 			if len(stmt.Results) != 1 { // should not happend : the method return error
// 				continue
// 			}
// 			if call, ok := stmt.Results[0].(*ast.CallExpr); ok {
// 				if method, ok := call.Fun.(*ast.SelectorExpr); ok {
// 					if method.Sel.Name == "JSON" || method.Sel.Name == "JSONPretty" {
// 						if len(call.Args) >= 2 {
// 							output := call.Args[1] // c.JSON(200, output)
// 							switch output := output.(type) {
// 							case *ast.Ident:
// 								out.Return = resolveLocalType(output, pkg)
// 							case *ast.CompositeLit:
// 								out.Return = parseCompositeLit(output, pkg)
// 							}
// 						}
// 					}
// 				}
// 			}

// 		case *ast.AssignStmt:
// 			parseAssignments(stmt.Rhs, pkg, &out)
// 		case *ast.IfStmt:
// 			if assign, ok := stmt.Init.(*ast.AssignStmt); ok {
// 				parseAssignments(assign.Rhs, pkg, &out)
// 			}
// 		}
// 	}
// 	return out
// }

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

// func resolveLocalType(arg *ast.Ident, pkg *types.Package) types.Type {
// 	localScope := pkg.Scope().Innermost(arg.Pos())
// 	obj := localScope.Lookup(arg.Name)
// 	for obj == nil {
// 		localScope = localScope.Parent()
// 		obj = localScope.Lookup(arg.Name)
// 	}
// 	return obj.Type()
// }

// func parseCompositeLit(lit *ast.CompositeLit, pkg *types.Package) types.Type {
// 	switch type_ := lit.Type.(type) {
// 	case *ast.Ident:
// 		return resolveLocalType(type_, pkg)
// 	}
// 	return nil
// }
