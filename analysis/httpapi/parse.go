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
func selectFileByPos(rootPackage *packages.Package, pos token.Pos) *ast.File {
	sel := analysis.NewPkgSelector(rootPackage)

	var aux func(pa *packages.Package) *ast.File
	aux = func(pa *packages.Package) *ast.File {
		for _, file := range pa.Syntax {
			if file.Pos() <= pos && pos <= file.End() {
				return file
			}
		}
		// recurse
		for _, imported := range pa.Imports {
			if sel.Ignore(imported) {
				continue
			}

			if fi := aux(imported); fi != nil {
				return fi
			}
		}
		return nil
	}

	fi := aux(rootPackage)
	if fi == nil {
		panic(fmt.Sprintf("file %s not found", rootPackage.Fset.Position(pos).Filename))
	}
	return fi
}

func selectPackage(rootPackage *packages.Package, target *types.Package) *packages.Package {
	sel := analysis.NewPkgSelector(rootPackage)

	var aux func(pa *packages.Package) *packages.Package
	aux = func(pa *packages.Package) *packages.Package {
		if pa.Types.Path() == target.Path() {
			return pa
		}
		// recurse
		for _, imported := range pa.Imports {
			if sel.Ignore(imported) {
				continue
			}

			if out := aux(imported); out != nil {
				return out
			}
		}
		return nil
	}

	out := aux(rootPackage)
	if out == nil {
		panic(fmt.Sprintf("can't find package %s", target))
	}
	return out
}

// expects an expression evaluable to a constant string
func resolveConstString(arg ast.Expr, pkg *packages.Package) (string, error) {
	exprString := types.ExprString(arg)
	tv, err := types.Eval(pkg.Fset, pkg.Types, arg.Pos(), exprString)
	if err != nil {
		return "", fmt.Errorf("can't resolve string at %s: %s", pkg.Fset.Position(arg.Pos()), exprString)
	}
	if tv.Value.Kind() != constant.String {
		return "", fmt.Errorf("expecting string at %s, got %s", pkg.Fset.Position(arg.Pos()), exprString)
	}
	return constant.StringVal(tv.Value), nil
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

func resolveFunc(rootPackage *packages.Package, fn *types.Func) (body *ast.BlockStmt, name string, sourcePkg *packages.Package) {
	pos := fn.Pos()
	fi := selectFileByPos(rootPackage, pos) // pos is at the begining of the name

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
	return body, fn.Name(), selectPackage(rootPackage, fn.Pkg())
}

func parseEndpointFunc(arg ast.Expr, pkg *packages.Package) (body *ast.BlockStmt, name string, sourcePkg *packages.Package) {
	if method, ok := arg.(*ast.SelectorExpr); ok { // <X>.<Sel>
		if ident, ok := method.X.(*ast.Ident); ok {
			xObj := resolveIdentifier(ident, pkg.TypesInfo)
			switch xObj := xObj.(type) {
			case *types.PkgName: // top level function
				fn := xObj.Imported().Scope().Lookup(method.Sel.Name)
				return resolveFunc(pkg, fn.(*types.Func))
			case *types.Var: // method
				// check for pointers
				var ty *types.Named
				if ptr, isPointer := xObj.Type().(*types.Pointer); isPointer {
					ty = ptr.Elem().(*types.Named)
				} else {
					ty = xObj.Type().(*types.Named)
				}
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
		return fnLitt.Body, fmt.Sprintf("Anonymous%d", arg.Pos()), pkg
	}

	panic(fmt.Sprintf("unsupported handler function %s", arg))
}

// resolveTypes list the required types, perform their analysis,
// and update `endpoints`
func resolveTypes(rootPkg *packages.Package, endpoints []Endpoint) {
	// collect the required types
	var required []types.Type
	for _, endpoint := range endpoints {
		// TODO: should we enfore something about input and output ?
		if ty := endpoint.Contract.inputT; ty != nil {
			required = append(required, ty)
		}
		if ty := endpoint.Contract.returnT; ty != nil {
			required = append(required, ty)
		}
		for _, param := range endpoint.Contract.InputQueryParams {
			if param.type_ == nil {
				panic(fmt.Sprint(param))
			}
			required = append(required, param.type_)
		}
	}

	// performs the analysis
	an := analysis.NewAnalysisFromTypes(rootPkg, required)

	// update back the endpoints
	for i := range endpoints {
		ct := &endpoints[i].Contract
		ct.InputBody = an.Types[ct.inputT]
		ct.Return = an.Types[ct.returnT]
		for j := range ct.InputQueryParams {
			ct.InputQueryParams[j].resolveType(an)
		}
	}
}

// parse applies the given parser to extract the endpoints, and resolves the types found.
func parse(pkg *packages.Package, absFilePath string, parser func(pkg *packages.Package, fi *ast.File) []Endpoint) []Endpoint {
	fi := selectFileByPath(pkg, absFilePath)
	apis := parser(pkg, fi)
	resolveTypes(pkg, apis)
	return apis
}

// ParseEcho scans a file using the Echo framework.
func ParseEcho(pkg *packages.Package, absFilePath string) []Endpoint {
	return parse(pkg, absFilePath, echoExtractor)
}
