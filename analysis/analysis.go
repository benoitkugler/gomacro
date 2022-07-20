// Package analysis defines the types structures used by every output format.
// It may be seen as an extension of `go/types`.
package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Type is the common interface for all the supported types.
// All implementation must be valid map keys, so that
// the can easily be mapped to output types.
type Type interface {
	// Name returns nil for universal types, such as
	// []bool, string, int, map[int]MyStruct, etc...
	Name() *types.Named

	// // Underlying returns the associated underlying Go type, which is
	// // thus never *types.Named
	// // For named types, it will always be Name().Underlying(), but
	// // not for anonymous ones.
	// Underlying() types.Type
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
	Types map[*types.Named]Type

	// Outline is the list of top-level types
	// defined in the analysis input file.
	Outline []*types.Named

	// enums and unions are used during analysis,
	// and may be used to handle enums and union types
	// not that, once `Types` is built, these fields are
	// somewhat redundant, since included in `Types`
	enums  Enums
	unions Unions
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

	out := &Analysis{Types: make(map[*types.Named]Type), unions: unions, enums: enums}
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

		out.handleType(named)
	}

	return out
}

// handleType analyzes the given type, registers the resulting `Type`
// and returns it.
// it is a no-op of `named` as already been processed
func (an *Analysis) handleType(named *types.Named) Type {
	if v, has := an.Types[named]; has { // we have already seen this type
		return v
	}

	// TODO: actually create the type
	var type_ Type

	// register it
	an.Types[named] = type_
	return type_
}
