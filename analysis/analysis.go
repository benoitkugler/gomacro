// Package analysis defines the types structures used by every output format.
// It may be seen as an extension of `go/types`.
package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Type is the common interface for all the supported types.
// All implementation must be valid map keys, so that
// the can easily be mapped to output types.
type Type interface {
	// Name returns nil for universal types, such as
	// []bool, string, int, map[int]MyStruct, etc...
	// It also returns nil for time.Time, since we consider
	// it as a standard type.
	Name() *types.Named
}

// loadSource returns the `packages.Package` containing the given file.
func loadSource(sourceFile string) (*packages.Package, error) {
	_, err := os.Stat(sourceFile)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(sourceFile)
	cfg := &packages.Config{
		Dir: dir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, "file="+sourceFile)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("only one package expected, got %d", len(pkgs))
	}
	if pkgs[0].IllTyped || len(pkgs[0].Errors) != 0 {
		return nil, fmt.Errorf("go package %s contains error", pkgs[0].Name)
	}
	return pkgs[0], nil
}

// nodeAt returns the node at `obj`,
// or panics if not found
func nodeAt(pa *packages.Package, pos token.Pos) ast.Node {
	declFile := pa.Fset.File(pos).Pos(0)
	for _, file := range pa.Syntax { // select the right file
		if file.Pos() == declFile {
			out := nodeAtFile(pos, file)
			if out == nil {
				panic("node not found in *ast.File")
			}
			return out
		}
	}
	panic("missing source file in Package.Syntax")
}

// nodeAtFile returns the node at `pos` in `file`
func nodeAtFile(pos token.Pos, file *ast.File) (out ast.Node) {
	var candidates []ast.Node
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if n.Pos() <= pos && pos < n.End() {
			candidates = append(candidates, n)
		}
		return true
	})

	var (
		bestRange = math.MaxInt32
		bestNode  ast.Node
	)
	for _, c := range candidates {
		if c.Pos() == pos { // no ambiguity
			return c
		}

		ra := int(c.End() - c.Pos())
		if ra < bestRange {
			bestRange = ra
			bestNode = c
		}
	}
	return bestNode
}

// fetchEnumsAndUnions fetches the enums and unions of the given package and
// all its imports, restricted to the same "root" folder,
// which is <domain>/<org>/<root>.
// For example, if `pa` is github.com/gopher/lib/server,
// only the subpackages github.com/gopher/lib/... will be searched.
func fetchEnumsAndUnions(pa *packages.Package) (Enums, Unions) {
	outEnums := make(Enums)
	outUnions := make(Unions)

	chunks := strings.Split(pa.PkgPath, "/")
	var prefix string
	if len(chunks) >= 3 {
		prefix = strings.Join(chunks[:3], "/")
	}

	var accuFunc func(*packages.Package)
	accuFunc = func(p *packages.Package) {
		// handle the current package and merge the result into `out`
		for k, v := range fetchPkgEnums(p) {
			outEnums[k] = v
		}
		for k, v := range fetchPkgUnions(p) {
			outUnions[k] = v
		}

		// recurse if needed
		for _, imp := range p.Imports {
			ignore := prefix != "" && !strings.HasPrefix(imp.PkgPath, prefix)
			if !ignore {
				accuFunc(imp)
			}
		}
	}
	accuFunc(pa)

	return outEnums, outUnions
}

// Analysis is the result of analyzing one package.
type Analysis struct {
	// Types adds the additional analysis of this package,
	// and contains all the types needed by `Outline` and
	// their dependencies.
	Types map[types.Type]Type

	// Outline is the list of top-level types
	// defined in the analysis input file.
	Outline []*types.Named
}

// NewAnalysis calls `packages.Load` on the given `sourceFile`
// and then built the analysis.
func NewAnalysis(sourceFile string) (*Analysis, error) {
	pa, err := loadSource(sourceFile)
	if err != nil {
		return nil, err
	}

	sourceFileAbs, err := filepath.Abs(sourceFile)
	if err != nil {
		return nil, err
	}

	return newAnalysis(pa, sourceFileAbs), nil
}

func newAnalysis(pa *packages.Package, sourceFileAbs string) *Analysis {
	enums, unions := fetchEnumsAndUnions(pa)
	ctx := context{enums: enums, unions: unions, rootPackage: pa}

	out := &Analysis{Types: make(map[types.Type]Type)}
	// walk the top level type declarations
	scope := pa.Types.Scope()
	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		if pa.Fset.Position(object.Pos()).Filename != sourceFileAbs {
			// retrict to file declaration
			continue
		}

		typeName, isTypeName := object.(*types.TypeName)
		if !isTypeName || typeName.IsAlias() {
			// ignore non-type declarations
			continue
		}

		named := typeName.Type().(*types.Named)

		out.Outline = append(out.Outline, named)

		out.handleType(named, ctx)
	}

	return out
}

// context stores the parameters need by the analysis,
// which may vary between types
type context struct {
	rootPackage *packages.Package

	// enums and unions are used during analysis,
	// and may be used to handle enums and union types
	enums  Enums
	unions Unions

	extern *externMap // optional
}

type externMap struct {
	externalFiles map[string]string
	goPackage     string
}

// check for tag with the form gomacro-extern:"<pkg>:<mode1>:<targetFile1>:<mode2>:<targetFile2>"
func newExternMap(tag string) *externMap {
	de := reflect.StructTag(tag).Get("gomacro-extern")
	if de == "" {
		return nil
	}

	chunks := strings.Split(de, ":")
	goPackage := chunks[0]
	chunks = chunks[1:]
	if len(chunks)%2 != 0 {
		panic("invalid gomacro-extern tag " + de)
	}
	externalFiles := make(map[string]string)
	for i := 0; i < len(chunks)/2; i++ {
		externalFiles[chunks[2*i]] = chunks[2*i+1]
	}
	return &externMap{goPackage: goPackage, externalFiles: externalFiles}
}

func (an *Analysis) handleStructFields(typ *types.Struct, ctx context) []StructField {
	var out []StructField
	for i := 0; i < typ.NumFields(); i++ {
		field := typ.Field(i)
		tag := typ.Tag(i)

		// handle extern definition
		ctx.extern = newExternMap(tag)

		// recurse
		fieldType := an.handleType(field.Type(), ctx)

		// to simplify, we do not fully support embedded fields :
		// we only accept structs, and we merge the fields
		if field.Embedded() {
			if st, isStruct := fieldType.(*Struct); isStruct {
				log.Printf("gomacro: embedded struct field %s will be flattened", field.Name())
				out = append(out, st.Fields...)
				continue
			} else {
				log.Printf("gomacro: field %s: embedding will be ignored", field.Name())
			}
		}

		out = append(out, StructField{Type: fieldType, Field: field, Tag: tag})
	}
	return out
}

func (an *Analysis) createType(typ types.Type, ctx context) Type {
	if typ == nil {
		panic("nil types.Type")
	}

	// special case for time.Time
	if ti, isTime := newTime(typ); isTime {
		return ti
	}

	name, isNamed := typ.(*types.Named)
	if isNamed {
		if ctx.extern != nil { // check for external refs
			if name.Obj().Pkg().Name() == ctx.extern.goPackage {
				return &Extern{name: name, ExternalFiles: ctx.extern.externalFiles}
			}
		}

		// look for enums and unions
		if enum, isEnum := ctx.enums[name]; isEnum {
			return enum
		} else if union, isUnion := ctx.unions[name]; isUnion {
			return union
		}

		// otherwise, analyze the underlying type
	}

	switch underlying := typ.Underlying().(type) {
	case *types.Pointer:
		// we do not distinguish between pointer vs regular values,
		// simply resolve the indirection
		return an.handleType(underlying.Elem(), ctx)
	case *types.Basic:
		return &Basic{typ: typ}

	// to properly handle recursive types (for Array, Slice, Map, Struct), we first register
	// an incomplete type so that handleType() returns early

	case *types.Array:
		out := &Array{name: name, Len: int(underlying.Len())}
		an.Types[typ] = out
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Slice:
		out := &Array{name: name, Len: -1}
		an.Types[typ] = out
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Map:
		out := &Map{name: name}
		an.Types[typ] = out
		out.Key = an.handleType(underlying.Key(), ctx)   // recurse
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Struct:
		if !isNamed {
			panic("anonymous structs are not supported")
		}
		out := &Struct{
			name:     name,
			Comments: fetchStructComments(ctx.rootPackage, name),
		}
		an.Types[typ] = out
		out.Fields = an.handleStructFields(underlying, ctx) // recurse
		return out
	default:
		// unhandled type, should not happend on real case
		panic("unsupported type " + typ.String())
	}
}

// handleType analyzes the given type, registers the resulting `Type`
// and returns it.
// it is a no-op if `typ` as already been processed
func (an *Analysis) handleType(typ types.Type, ctx context) Type {
	if v, has := an.Types[typ]; has { // we have already seen this type
		return v
	}

	// resolve the type
	type_ := an.createType(typ, ctx)

	// register it
	an.Types[typ] = type_

	return type_
}
