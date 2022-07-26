package analysis

import (
	"go/ast"
	"go/constant"
	"go/types"
	"sort"
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

func (em EnumMember) int64() (int64, bool) {
	return constant.Int64Val(em.Const.Val())
}

type Enum struct {
	name *types.Named // by construction Underlying is *types.Basic

	// Members contains all the values, even the unexported one
	Members []EnumMember

	// IsIota is `true` if the enum exported values are consecutive positive integer starting at zero.
	IsIota bool
}

func (e *Enum) Type() types.Type { return e.name }

// IsInteger returns `true` is this enum is backed by
// integers (which may be negative and not contiguous)
func (e *Enum) IsInteger() bool {
	info := e.name.Underlying().(*types.Basic).Info()
	return info&types.IsInteger != 0
}

type sortBy struct {
	members []EnumMember
	values  []int64
}

func (a sortBy) Len() int { return len(a.members) }
func (a sortBy) Swap(i, j int) {
	a.members[i], a.members[j] = a.members[j], a.members[i]
	a.values[i], a.values[j] = a.values[j], a.values[i]
}
func (a sortBy) Less(i, j int) bool { return a.values[i] < a.values[j] }

// ensure the members are sorted
func (e *Enum) setIsIota() {
	if !e.IsInteger() {
		return
	}

	// check all the values
	values := make([]int64, len(e.Members))
	seen := make(map[int64]bool)
	var max int64 = -1
	for i, member := range e.Members {
		v, ok := member.int64()
		if !ok || v < 0 {
			return
		}
		values[i] = v
		if !member.Const.Exported() {
			continue // ignore non exported const
		}
		seen[v] = true
		if max < v {
			max = v
		}
	}
	if len(seen) != int(max+1) {
		return
	}

	sort.Sort(sortBy{
		members: e.Members,
		values:  values,
	})
	e.IsIota = true
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

	// once all values has been resolved, set IsIota
	for _, e := range out {
		e.setIsIota()
	}

	return out
}
