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
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	// ExhaustiveTypeSwitch may be used as marker to make futur refactor easier
	ExhaustiveTypeSwitch = "exhaustive analysis.Type type switch"

	// ExhaustiveAnonymousTypeSwitch may be used as marker to make futur refactor easier
	ExhaustiveAnonymousTypeSwitch = "exhaustive analysis.AnonymousType type switch"

	// ExhaustiveBasicKindSwitch may be used as marker to make futur refactor easier
	ExhaustiveBasicKindSwitch = "exhaustive analysis.BasicKind switch"
)

// Type is the common interface for all the supported types.
// All implementation are pointers, so that
// they can easily be mapped to output types.
type Type interface {
	// Type returns the Go type corresponding to this tag.
	Type() types.Type
}

// LocalName returns the local name of the type.
// It will panic if `ty` is not named.
func LocalName(ty Type) string {
	return ty.Type().(*types.Named).Obj().Name()
}

func selectByFile(pkgs []*packages.Package, file string) *packages.Package {
	for _, pkg := range pkgs {
		for _, source := range pkg.GoFiles {
			if source == file {
				return pkg
			}
		}
	}
	return nil
}

func commonPrefix(paths []string) string {
	index := 0
	first := paths[0]
	for ; index < len(first); index++ {
		c := first[index]
		for _, other := range paths {
			if index >= len(other) || other[index] != c {
				// no more prefix
				return first[:index]
			}
		}
	}

	return first
}

// LoadSources returns for each source file, the `*packages.Package` containing it.
// Since it only calls `packages.Load` once, it is a faster alternative
// to repeated `LoadSource` calls.
func LoadSources(sourceFiles []string) ([]*packages.Package, error) {
	patterns := make([]string, len(sourceFiles))
	dirs := make([]string, len(sourceFiles))

	for i, sourceFile := range sourceFiles {
		_, err := os.Stat(sourceFile)
		if err != nil {
			return nil, err
		}
		patterns[i] = "file=" + sourceFile

		abs, err := filepath.Abs(sourceFile)
		if err != nil {
			return nil, err
		}
		dirs[i] = filepath.Dir(abs)
	}

	// compute the common directory
	dir := commonPrefix(dirs)
	if dir == "" {
		return nil, fmt.Errorf("invalid source directories: %v", dirs)
	}

	cfg := &packages.Config{
		Dir: dir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedTypesInfo,
	}
	// let packages handle the heavy lifting for us
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	nbErrors := packages.PrintErrors(pkgs)

	if nbErrors > 0 {
		return nil, fmt.Errorf("go packages have errors")
	}

	// match back the packages
	out := make([]*packages.Package, len(sourceFiles))
	for i, sourceFile := range sourceFiles {
		abs, err := filepath.Abs(sourceFile)
		if err != nil {
			return nil, err
		}

		selected := selectByFile(pkgs, abs)
		if selected == nil {
			return nil, fmt.Errorf("file %s not found in packages sources", abs)
		}
		out[i] = selected
	}

	return out, nil
}

// LoadSource returns the `packages.Package` containing the given file.
func LoadSource(sourceFile string) (*packages.Package, error) {
	pkgs, err := LoadSources([]string{sourceFile})
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("only one package expected, got %d", len(pkgs))
	}
	return pkgs[0], nil
}

// nodeAt returns the node at `obj`,
// or panics if not found
func nodeAt(pa *packages.Package, pos token.Pos) ast.Node {
	declFile := pa.Fset.File(pos)
	for _, file := range pa.Syntax { // select the right file
		tokenFile := pa.Fset.File(file.Pos())
		if tokenFile == declFile {
			out := nodeAtFile(pos, file)
			if out == nil {
				panic("node not found in *ast.File")
			}
			return out
		}
	}
	panic("missing source file in Package.Syntax " + pa.String())
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

// PkgSelector allows to retrict the packages
// import graph walk to the user written.
type PkgSelector struct {
	prefix string
}

func NewPkgSelector(root *packages.Package) PkgSelector {
	chunks := strings.Split(root.PkgPath, "/")
	var prefix string
	if len(chunks) >= 3 {
		prefix = strings.Join(chunks[:3], "/")
	}
	return PkgSelector{prefix: prefix}
}

// Ignore returns true if the given package should not
// be recursed on.
func (ps PkgSelector) Ignore(pa *packages.Package) bool {
	return ps.ignorePath(pa.PkgPath)
}

func (ps PkgSelector) ignorePath(path string) bool {
	return ps.prefix != "" && !strings.HasPrefix(path, ps.prefix)
}

// fetchEnumsAndUnions fetches the enums and unions of the given package and
// all its imports, restricted to the same "root" folder,
// which is <domain>/<org>/<root>.
// For example, if `pa` is github.com/gopher/lib/server,
// only the subpackages github.com/gopher/lib/... will be searched.
func fetchEnumsAndUnions(pa *packages.Package) (enumsMap, unionsMap) {
	outEnums := make(enumsMap)
	outUnions := make(unionsMap)

	selector := NewPkgSelector(pa)

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
			if selector.Ignore(imp) {
				continue
			}
			accuFunc(imp)
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

	// Root is the root package used to query type information.
	Root *packages.Package

	// Source is the list of top-level types
	// defined in the analysis input file.
	Source []types.Type
}

// NewAnalysisFromFile uses the given Package
// to build the analysis for the types defined in `sourceFile`.
func NewAnalysisFromFile(pa *packages.Package, sourceFile string) *Analysis {
	sourceFileAbs, err := filepath.Abs(sourceFile)
	if err != nil {
		panic(err)
	}

	// walk the top level type declarations
	var objs []*types.TypeName
	scope := pa.Types.Scope()
	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		if pa.Fset.Position(object.Pos()).Filename != sourceFileAbs {
			// retrict to file declaration
			continue
		}

		typeName, isTypeName := object.(*types.TypeName)
		if !isTypeName {
			// ignore non-type declarations
			continue
		}

		objs = append(objs, typeName)
	}

	// order according to source, so that for instance SQL constraints
	// are kept in correct order
	sort.Slice(objs, func(i, j int) bool { return objs[i].Pos() < objs[j].Pos() })

	nameds := make([]types.Type, len(objs))
	for i, obj := range objs {
		nameds[i] = obj.Type()
	}

	return NewAnalysisFromTypes(pa, nameds)
}

// NewAnalysisFromTypes build the analysis for the given `types`.
// `root` is the root package, required to query type information.
func NewAnalysisFromTypes(root *packages.Package, source []types.Type) *Analysis {
	out := &Analysis{Source: source, Root: root}
	out.populateTypes(root)
	return out
}

func (an *Analysis) populateTypes(pa *packages.Package) {
	enums, unions := fetchEnumsAndUnions(pa)

	ctx := context{enums: enums, unions: unions, rootPackage: pa}

	an.Types = make(map[types.Type]Type)
	for _, typ := range an.Source {
		an.handleType(typ, ctx)
	}

	for _, type_ := range an.Types {
		if st, isStruct := type_.(*Struct); isStruct {
			st.setImplements(ctx.unions, an.Types)
		}
	}
}

// context stores the parameters need by the analysis,
// which may vary between types
type context struct {
	rootPackage *packages.Package

	// enums and unions are used during analysis,
	// and may be used to handle enums and union types
	enums  enumsMap
	unions unionsMap

	extern *externMap // optional
}

type externMap struct {
	externalFiles map[string]string
	goPackage     string
}

// check for tag with the form gomacro-extern:"<pkg>:<mode1>:<targetFile1>:<mode2>:<targetFile2>"
func newExternMap(tag reflect.StructTag) *externMap {
	de := tag.Get("gomacro-extern")
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
		tag := reflect.StructTag(typ.Tag(i))

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

	// special case for time.Time, which require the name information
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

		// look for enums, unions and structs
		if enum, isEnum := ctx.enums[name]; isEnum {
			return enum
		} else if members, isUnion := ctx.unions[name]; isUnion {
			// add the member to the type to analyze
			out := &Union{name: name}
			for _, member := range members {
				out.Members = append(out.Members, an.handleType(member, ctx))
			}
			return out
		} else if st, isStruct := typ.Underlying().(*types.Struct); isStruct {
			out := &Struct{
				Name: name,
				// Implements are defered
			}
			an.Types[typ] = out                         // register before recursing
			out.Fields = an.handleStructFields(st, ctx) // recurse
			out.Comments = fetchStructComments(ctx.rootPackage, name)
			return out
		}

		// otherwise, analyze the underlying type
		under := an.handleType(typ.Underlying(), ctx).(AnonymousType)
		return &Named{name: name, Underlying: under}
	}

	switch underlying := typ.Underlying().(type) {
	case *types.Pointer:
		elem := an.handleType(underlying.Elem(), ctx) // recurse for the element
		return &Pointer{Elem: elem}
	case *types.Basic:
		return &Basic{B: underlying}

	// to properly handle recursive types (for Array, Slice, Map, Struct), we first register
	// an incomplete type so that handleType() returns early

	case *types.Array:
		out := &Array{Len: int(underlying.Len())}
		an.Types[typ] = out
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Slice:
		out := &Array{Len: -1}
		an.Types[typ] = out
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Map:
		out := &Map{}
		an.Types[typ] = out
		out.Key = an.handleType(underlying.Key(), ctx)   // recurse
		out.Elem = an.handleType(underlying.Elem(), ctx) // recurse
		return out
	case *types.Struct:
		panic("anonymous structs are not supported")
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
