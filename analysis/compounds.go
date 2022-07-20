package analysis

import (
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

type ClassField struct {
	Type Type
	Name string // the name of the field in the Go struct
}

type Class struct {
	name *types.Named

	Fields   []ClassField
	Comments []SpecialComment
}

func (cl *Class) Name() *types.Named { return cl.name }

// Implements return the union types this class implements,
// among the ones given.
func (cl *Class) Implements(unions Unions) []*Union {
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

func fetchStructComments(pa *packages.Package, obj *types.Named) (out []SpecialComment) {
	pos := obj.Obj().Pos()
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
