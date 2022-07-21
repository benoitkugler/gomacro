package analysis

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

const IgnoreDeclComment = "gomacro:no-enum"

// enumsMap is the set of named types which are
// considered as enums.
type enumsMap map[*types.Named]*Enum

// EnumMember decribe the value of one item in an enumeration type.
type EnumMember struct {
	Const *types.Const
	// Comment is an optional comment associated with
	// the enum value.
	// If provided, it is used as label.
	Comment string
}

type Enum struct {
	name *types.Named // by construction Underlying is *types.Basic

	// Members contains only the exported members
	Members []EnumMember
}

func (t *Enum) Name() *types.Named { return t.name }

// func (t *Enum) Underlying() types.Type { return t.name.Underlying() }

// IsInteger returns `true` is this enum is backed by
// positive integer values (as opposed for instance to string values).
func (e *Enum) IsInteger() bool {
	info := e.name.Underlying().(*types.Basic).Info()
	if info&types.IsInteger == 0 {
		return false
	}

	if info&types.IsUnsigned != 0 { // fast path
		return true
	}
	// check all the values
	for _, member := range e.Members {
		v, ok := constant.Int64Val(member.Const.Val())
		if !ok || v < 0 {
			return false
		}
	}
	return true
}

// fetchConstComment retrieve the comment, not exposed in go/types
func fetchConstComment(pa *packages.Package, obj *types.Const) string {
	node := nodeAt(pa, obj.Pos())
	spec := node.(*ast.ValueSpec)
	if spec.Comment == nil {
		return ""
	}
	return strings.TrimSpace(spec.Comment.Text())
}

// fetchPkgEnums walks through all the constants defined by the given package
// to extract enums types (and their members)
// unexported values are ignored
func fetchPkgEnums(pa *packages.Package) enumsMap {
	// we detect values and then build the types from it
	out := make(enumsMap)
	scope := pa.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		decl, isConst := obj.(*types.Const)
		if !isConst {
			continue
		}
		if !decl.Exported() {
			continue
		}
		named, isNamed := decl.Type().(*types.Named)
		if !isNamed {
			continue
		}
		// per the spec, only basic types may be constant

		comment := fetchConstComment(pa, decl)
		if strings.Contains(comment, IgnoreDeclComment) { // this value does not implies an enum
			continue
		}

		enum := out[named]
		if enum == nil {
			enum = &Enum{name: named}
			out[named] = enum
		}
		enum.Members = append(enum.Members, EnumMember{
			Const:   decl,
			Comment: comment,
		})
	}
	return out
}
