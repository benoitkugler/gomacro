package analysis

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/packages"
)

// Array is either a fixed or variable length array
type Array struct {
	name *types.Named // optional
	Elem Type
	// -1 for slices
	Len int
}

func (ar *Array) Name() *types.Named { return ar.name }

type Map struct {
	name *types.Named // optional
	Key  Type
	Elem Type
}

func (ma *Map) Name() *types.Named { return ma.name }

type CommentKind uint8

const (
	_ CommentKind = iota

	// CommentSQL is used to add SQL statements to the generated tables creation code
	CommentSQL
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
	Field *types.Var // returned by Struct.Field()
	Tag   string     // returned by Struct.Tag()
}

type Struct struct {
	name *types.Named

	Fields   []StructField
	Comments []SpecialComment
}

func (cl *Struct) Name() *types.Named { return cl.name }

// Implements return the union types this class implements,
// among the ones given.
func (cl *Struct) Implements(unions Unions) []*Union {
	cl.name.Obj().Pos()
	var out []*Union
	for _, v := range unions {
		for _, member := range v.Members {
			if cl.name == member {
				out = append(out, v)
			}
		}
	}
	return out
}

// findPackage recurses through the imports to find the package `obj` belongs to
func findPackage(rootPackage *packages.Package, obj types.Object) *packages.Package {
	if obj.Pkg().Path() == rootPackage.PkgPath {
		return rootPackage
	}
	for _, importPkg := range rootPackage.Imports {
		if pa := findPackage(importPkg, obj); pa != nil {
			return pa
		}
	}
	return nil
}

func fetchStructComments(rootPackage *packages.Package, name *types.Named) (out []SpecialComment) {
	pa := findPackage(rootPackage, name.Obj())
	if pa == nil {
		panic(fmt.Sprintf("package %s not found", name.Obj().Pkg()))
	}
	// make sure pa and name work with the same file set
	name = pa.Types.Scope().Lookup(name.Obj().Name()).Type().(*types.Named)

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
	default:
		panic("unknown special comment " + match[1])
	}
}
