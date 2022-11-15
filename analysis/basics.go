package analysis

import (
	"go/types"
	"strings"
)

// AnonymousType are the types which may be
// seen without associated names.
// Contrary to the Go language, this package
// does not support anonymous structs, enums and unions.
type AnonymousType interface {
	Type

	isAnonymous()
}

func (*Basic) isAnonymous() {}
func (*Time) isAnonymous()  {}
func (*Array) isAnonymous() {}
func (*Map) isAnonymous()   {}

func (b *Basic) Type() types.Type { return b.B }

func (*Time) Type() types.Type {
	return types.NewNamed(types.NewTypeName(0, nil, "Time", nil), &types.Struct{}, nil)
}

func (ar *Array) Type() types.Type {
	if ar.Len >= 0 {
		return types.NewArray(ar.Elem.Type(), int64(ar.Len))
	}
	return types.NewSlice(ar.Elem.Type())
}

func (ma *Map) Type() types.Type { return types.NewMap(ma.Key.Type(), ma.Elem.Type()) }

// Named is a named type pointing
// to an `AnonymousType`. Structs, enums,
// and unions are NOT `Named`
type Named struct {
	name       *types.Named
	Underlying AnonymousType
}

func (na *Named) Type() types.Type { return na.name }

var (
	String = &Basic{B: types.Typ[types.String]}
	Bool   = &Basic{B: types.Typ[types.Bool]}
	Int    = &Basic{B: types.Typ[types.Int]}
	Float  = &Basic{B: types.Typ[types.Float64]}
)

// BasicKind is a simplified information of the kind
// of a basic type, typically shared by Dart, JSON and TypeScript generator
type BasicKind uint8

const (
	BKString BasicKind = iota
	BKInt
	BKFloat
	BKBool
)

func NewBasicKind(info types.BasicInfo) (BasicKind, bool) {
	if info&types.IsBoolean != 0 {
		return BKBool, true
	} else if info&types.IsInteger != 0 {
		return BKInt, true
	} else if info&types.IsFloat != 0 {
		return BKFloat, true
	} else if info&types.IsString != 0 {
		return BKString, true
	} else {
		return 0, false
	}
}

// Basic represents all simple types.
// Enums are special cased, meaning they are not of type `Basic`.
type Basic struct {
	B *types.Basic
}

func (b *Basic) Kind() BasicKind {
	info := b.B.Underlying().(*types.Basic).Info()
	out, ok := NewBasicKind(info)
	if !ok {
		panic("unsupported basic kind")
	}
	return out
}

// Time is a special case for *time.Time,
// which is not handled like a regular "pointer to named struct"
type Time struct {
	// IsDate is true if only year/month/day are actually
	// of interest for this type.
	// The following heuristic is used to compute it :
	//	- a type containing Date in its name (case insensitive match) will have IsDate = true
	IsDate bool
}

var (
	timeT = &Time{}
	dateT = &Time{IsDate: true}
)

// newTime returns `true` is the underlying type
// of `typ` is time.Time
func newTime(typ types.Type) (Type, bool) {
	// since we can't acces the underlying name of named types,
	// we check against this string, to detect time.Time
	const timeString = "struct{wall uint64; ext int64; loc *time.Location}"

	if typ.Underlying().String() != timeString {
		return nil, false
	}
	name, isNamed := typ.(*types.Named)
	isDate := isNamed && strings.Contains(strings.ToLower(name.Obj().Name()), "date")
	isCustomNamed := name.Obj().Pkg().Path() != "time"
	out := timeT
	if isDate {
		out = dateT
	}
	if isCustomNamed {
		return &Named{name: name, Underlying: out}, true
	}
	return out, true
}

// Extern is a special type used when
// no declaration should be written for a type,
// but the type should be imported from an other file.
type Extern struct {
	origin *types.Named

	// Optional
	Underlying AnonymousType

	// the files where to find the definition,
	// depending on the generation mode
	ExternalFiles map[string]string
}

func (e *Extern) Type() types.Type { return e.origin }

// Pointer is a pointer to a type
type Pointer struct {
	Elem Type
}

func (p *Pointer) Type() types.Type {
	return types.NewPointer(p.Elem.Type())
}
