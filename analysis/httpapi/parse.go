package httpapi

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/benoitkugler/gomacro/analysis"
	"golang.org/x/tools/go/packages"
)

// selectFileByPath returns the ast of `filePathAbs`
func selectFileByPath(pkg *packages.Package, filePathAbs string) *ast.File {
	for _, file := range pkg.Syntax {
		if pkg.Fset.File(file.Package).Name() == filePathAbs {
			return file
		}
	}
	panic(fmt.Sprintf("file %s not found", filePathAbs))
}

// selectFileByPos returns the ast of the file containing `pos`
func selectFileByPos(pkg *packages.Package, pos token.Pos) *ast.File {
	for _, file := range pkg.Syntax {
		if file.Pos() <= pos && pos <= file.End() {
			return file
		}
	}
	// recurse
	for _, imported := range pkg.Imports {
		if fi := selectFileByPos(imported, pos); fi != nil {
			return fi
		}
	}
	return nil
}

func selectPackage(rootPackage *packages.Package, target *types.Package) *packages.Package {
	if rootPackage.Types.Path() == target.Path() {
		return rootPackage
	}
	// recurse
	for _, imported := range rootPackage.Imports {
		if pa := selectPackage(imported, target); pa != nil {
			return pa
		}
	}
	return nil
}

// expects an expression evaluable to a constant string URL
func parseEndpointURL(arg ast.Expr, pkg *packages.Package) string {
	exprString := types.ExprString(arg)
	tv, err := types.Eval(pkg.Fset, pkg.Types, arg.Pos(), exprString)
	if err != nil {
		panic(fmt.Sprintf("can't resolved URL at %s: %s", pkg.Fset.Position(arg.Pos()), exprString))
	}
	if tv.Value.Kind() != constant.String {
		panic(fmt.Sprintf("expecting string for URL at %s, got %s", pkg.Fset.Position(arg.Pos()), exprString))
	}
	return constant.StringVal(tv.Value)
}

func resolveIdentifier(x *ast.Ident, pkg *types.Info) types.Object {
	if u := pkg.Uses[x]; u != nil {
		return u
	}
	return pkg.Defs[x]
}

func selectMethod(named *types.Named, methodName string) *types.Func {
	for i := 0; i < named.NumMethods(); i++ {
		fn := named.Method(i)
		if methodName == fn.Name() {
			return fn
		}
	}
	panic(fmt.Sprintf("method %s not found on type %s", methodName, named))
}

func resolveFunc(pkg *packages.Package, fn *types.Func) (body *ast.BlockStmt, name string, sourcePkg *types.Info) {
	pos := fn.Pos()
	fi := selectFileByPos(pkg, pos) // pos is at the begining of the name
	if fi == nil {
		panic(fmt.Sprintf("file %s not found", pkg.Fset.Position(pos).Filename))
	}

	ast.Inspect(fi, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		// pos is at the begining of the name
		if n.Pos() <= pos && pos < n.End() {
			if decl, isDecl := n.(*ast.FuncDecl); isDecl {
				body = decl.Body
				return false
			}
		}

		return true
	})
	if body == nil {
		panic(fmt.Sprintf("can't find body for %s", fn))
	}
	return body, fn.Name(), selectPackage(pkg, fn.Pkg()).TypesInfo
}

func parseEndpointFunc(arg ast.Expr, pkg *packages.Package) (body *ast.BlockStmt, name string, sourcePkg *types.Info) {
	if method, ok := arg.(*ast.SelectorExpr); ok { // <X>.<Sel>
		if ident, ok := method.X.(*ast.Ident); ok {
			xObj := resolveIdentifier(ident, pkg.TypesInfo)
			switch xObj := xObj.(type) {
			case *types.PkgName: // top level function
				fn := xObj.Imported().Scope().Lookup(method.Sel.Name)
				return resolveFunc(pkg, fn.(*types.Func))
			case *types.Var: // method
				ty := xObj.Type().(*types.Named)
				return resolveFunc(pkg, selectMethod(ty, method.Sel.Name))
			default:
				panic(fmt.Sprintf("unsupported identifier %s", xObj))
			}
		}
	} else if ident, ok := arg.(*ast.Ident); ok {
		obj := resolveIdentifier(ident, pkg.TypesInfo)
		if fn, isFn := obj.(*types.Func); isFn {
			return resolveFunc(pkg, fn)
		} else {
			panic(fmt.Sprintf("unsupported identifier %s for %s", obj, ident))
		}
	} else if fnLitt, ok := arg.(*ast.FuncLit); ok {
		return fnLitt.Body, fmt.Sprintf("Anonymous%d", arg.Pos()), pkg.TypesInfo
	}

	panic(fmt.Sprintf("unsupported handler function %s", arg))
}

// resolveTypes list the required types, perform their analysis,
// and update `endpoints`
func resolveTypes(rootPkg *packages.Package, endpoints []Endpoint) {
	// collect the required types
	var required []types.Type
	for _, endpoint := range endpoints {
		required = append(required, endpoint.Contract.inputT, endpoint.Contract.returnT)
		for _, param := range endpoint.Contract.QueryParams {
			required = append(required, param.type_)
		}
	}

	// performs the analysis
	an := analysis.NewAnalysisFromTypes(rootPkg, required)

	// update back the endpoints
	for i := range endpoints {
		ct := &endpoints[i].Contract
		ct.Input = an.Types[ct.inputT]
		ct.Return = an.Types[ct.returnT]
		for j := range ct.QueryParams {
			param := &ct.QueryParams[j]
			param.Type = an.Types[param.type_]
		}
	}
}
