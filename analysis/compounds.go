package analysis

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Array is either a fixed or variable length array
type Array struct {
	Elem Type
	// -1 for slices
	Len int
}

type Map struct {
	Key  Type
	Elem Type
}

type CommentKind uint8

const (
	_ CommentKind = iota

	// CommentSQL is used to add SQL statements to the generated tables creation code
	CommentSQL

	// CommentQuery is used to generated custom SQL queries
	CommentQuery
)

// SpecialComment is a comment found on a struct
// declaration.
// See `CommentKind` for the supported comments.
type SpecialComment struct {
	Content string
	Kind    CommentKind
}

type StructField struct {
	Type  Type
	Field *types.Var        // returned by Struct.Field()
	Tag   reflect.StructTag // returned by Struct.Tag()
}

// JSONName returns the field name used by Go json package,
// that is, taking into account the json struct tag.
func (st StructField) JSONName() string {
	if name := st.Tag.Get("json"); name != "" {
		return name
	}
	return st.Field.Name()
}

// Exported returns `true` is the field is exported and should be
// included in the generated code.
// Ignored field are either :
//   - unexported
//   - with a 'json' tag '-'
//   - with a 'gomacro' tag 'ignore'
//
// Note that [IsSQLGuard] takes precedences over this function.
func (st StructField) Exported() bool {
	if name := st.Tag.Get("json"); name == "-" {
		return false
	}
	if name := st.Tag.Get("gomacro"); name == "ignore" {
		return false
	}
	return st.Field.Exported()
}

// IsSQLGuard returns true if the field has a "gomacro-sql-guard" tag,
// with its value
func (st StructField) IsSQLGuard() (value string, ok bool) {
	val := st.Tag.Get("gomacro-sql-guard")
	return val, val != ""
}

// IsOpaqueFor returns true if the field should be considered
// as dynamic when generating code for [target]
func (st StructField) IsOpaqueFor(target string) bool {
	if t := st.Tag.Get("gomacro-opaque"); t != "" {
		return strings.Contains(t, target)
	}
	return false
}

type Struct struct {
	Name *types.Named

	Fields     []StructField
	Comments   []SpecialComment
	Implements []*Union
}

func (cl *Struct) Type() types.Type { return cl.Name }

// setImplements set `Implements` with the union types this class implements,
// among the ones given.
func (cl *Struct) setImplements(unions unionsMap, accu map[types.Type]Type) {
	var out []*Union
	for unionName, v := range unions {
		unionType, isUnionAnalyzed := accu[unionName].(*Union)
		if !isUnionAnalyzed {
			continue // ignore the union if it was not required by the entry types
		}
		for _, member := range v {
			if cl.Name == member {
				out = append(out, unionType)
				break
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].name.String() < out[j].name.String() })

	cl.Implements = out
}

// findPackage recurses through the imports to find the package `obj` belongs to
// it panics if obj if not a user defined type
func (ps PkgSelector) findPackage(rootPackage *packages.Package, obj *types.TypeName) *packages.Package {
	var aux func(pa *packages.Package) *packages.Package
	aux = func(pa *packages.Package) *packages.Package {
		if obj.Pkg().Path() == pa.PkgPath {
			return pa
		}
		for _, importPkg := range pa.Imports {
			if ps.Ignore(importPkg) {
				continue
			}
			// recurse
			if out := aux(importPkg); out != nil {
				return out
			}
		}

		return nil
	}

	out := aux(rootPackage)
	if out == nil {
		panic(fmt.Sprintf("package %s not found", obj.Pkg()))
	}
	return out
}

func fetchStructComments(rootPackage *packages.Package, name *types.Named) (out []SpecialComment) {
	selector := NewPkgSelector(rootPackage)

	// ignore non user types
	if selector.ignorePath(name.Obj().Pkg().Path()) {
		return nil
	}

	pa := selector.findPackage(rootPackage, name.Obj())

	// make sure pa and name work with the same file set
	scope := pa.Types.Scope().Lookup(name.Obj().Name())
	if scope == nil {
		return nil
	}
	name = scope.Type().(*types.Named)

	pos := name.Obj().Pos()
	node := nodeAt(pa, pos-1) // move up by one char to get the line right before the struct
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Doc == nil {
		return nil
	}
	for _, line := range decl.Doc.List {
		if kind, content := isSpecialComment(line.Text); kind != 0 {
			out = append(out, SpecialComment{Kind: kind, Content: content})
		}
	}
	return out
}

var reComment = regexp.MustCompile(`^// gomacro:(\w+) (.+)`)

// isSpecialComment returns a non empty tag if the comment
// has a special form // gomacro:<tag> <content>
// it panics if <tag> is unknown
func isSpecialComment(comment string) (kind CommentKind, content string) {
	match := reComment.FindStringSubmatch(comment)
	if len(match) == 0 {
		return 0, ""
	}
	switch match[1] {
	case "SQL":
		return CommentSQL, match[2]
	case "QUERY":
		return CommentQuery, match[2]
	default:
		panic("unknown special comment " + match[1])
	}
}
